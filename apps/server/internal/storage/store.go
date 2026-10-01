package storage

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"carteira_investimentos/server/internal/domain"

	"github.com/redis/go-redis/v9"
)

// SnapshotStore é o contrato universal de armazenamento de cotações e análises
type SnapshotStore interface {
	Salvar(ativo *domain.Ativo) error
	SalvarLote(ativos []*domain.Ativo) error
	Obter(ticker string) (*domain.Ativo, error)
	ListarTodos() []*domain.Ativo
}

// ----------------------------------------------------------------------
// 1. Adaptador In-Memory (Para Desenvolvimento Local no Windows)
// ----------------------------------------------------------------------

type InMemoryStore struct {
	mu     sync.RWMutex
	ativos map[string]*domain.Ativo
}

func NovoInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		ativos: make(map[string]*domain.Ativo),
	}
}

func (s *InMemoryStore) Salvar(ativo *domain.Ativo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ativos[ativo.Ticker] = ativo
	return nil
}

func (s *InMemoryStore) SalvarLote(ativos []*domain.Ativo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range ativos {
		s.ativos[a.Ticker] = a
	}
	return nil
}

func (s *InMemoryStore) Obter(ticker string) (*domain.Ativo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ativo, existe := s.ativos[ticker]
	if !existe {
		return nil, errors.New("ativo não encontrado no snapshot")
	}
	return ativo, nil
}

func (s *InMemoryStore) ListarTodos() []*domain.Ativo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var lista []*domain.Ativo
	for _, a := range s.ativos {
		lista = append(lista, a)
	}
	return lista
}

// ----------------------------------------------------------------------
// 2. Adaptador Redis (Para Produção na VPS - redis_central)
// ----------------------------------------------------------------------

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NovoRedisStore(endereco string, senha string, db int, ttl time.Duration) *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     endereco,
		Password: senha,
		DB:       db,
	})
	return &RedisStore{
		client: rdb,
		ttl:    ttl,
	}
}

func (r *RedisStore) Salvar(ativo *domain.Ativo) error {
	ctx := context.Background()
	dados, err := json.Marshal(ativo)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "b3:snapshot:"+ativo.Ticker, dados, r.ttl).Err()
}

func (r *RedisStore) SalvarLote(ativos []*domain.Ativo) error {
	for _, a := range ativos {
		if err := r.Salvar(a); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisStore) Obter(ticker string) (*domain.Ativo, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, "b3:snapshot:"+ticker).Result()
	if err != nil {
		return nil, err
	}
	var a domain.Ativo
	if err := json.Unmarshal([]byte(val), &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *RedisStore) ListarTodos() []*domain.Ativo {
	ctx := context.Background()
	keys, err := r.client.Keys(ctx, "b3:snapshot:*").Result()
	if err != nil {
		return nil
	}
	var lista []*domain.Ativo
	for _, k := range keys {
		val, err := r.client.Get(ctx, k).Result()
		if err == nil {
			var a domain.Ativo
			if json.Unmarshal([]byte(val), &a) == nil {
				lista = append(lista, &a)
			}
		}
	}
	return lista
}
