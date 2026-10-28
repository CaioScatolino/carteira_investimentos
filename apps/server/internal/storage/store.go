package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// CacheStore é o contrato genérico para armazenamento temporário chave-valor (Redis/In-Memory)
type CacheStore interface {
	ObterCache(ctx context.Context, chave string) ([]byte, error)
	SalvarCache(ctx context.Context, chave string, dados []byte, ttl time.Duration) error
}

// ----------------------------------------------------------------------
// 1. Adaptador In-Memory (Para Desenvolvimento Local no Windows)
// ----------------------------------------------------------------------

type inMemoryCacheItem struct {
	dados    []byte
	expiraEm time.Time
}

type InMemoryStore struct {
	mu     sync.RWMutex
	ativos map[string]*domain.Ativo
	cache  map[string]inMemoryCacheItem
}

func NovoInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		ativos: make(map[string]*domain.Ativo),
		cache:  make(map[string]inMemoryCacheItem),
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

func (s *InMemoryStore) ObterCache(ctx context.Context, chave string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, existe := s.cache[chave]
	if !existe {
		return nil, errors.New("chave não encontrada no cache")
	}
	if time.Now().After(item.expiraEm) {
		return nil, errors.New("chave expirada no cache")
	}
	return item.dados, nil
}

func (s *InMemoryStore) SalvarCache(ctx context.Context, chave string, dados []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[chave] = inMemoryCacheItem{
		dados:    dados,
		expiraEm: time.Now().Add(ttl),
	}
	return nil
}

// ----------------------------------------------------------------------
// 2. Adaptador Redis (Para Produção e Cache Distribuído de Alta Performance)
// ----------------------------------------------------------------------

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NovoRedisStore inicializa o cliente Redis e valida a conexão com Ping
func NovoRedisStore(endereco string, senha string, db int, ttl time.Duration) (*RedisStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         endereco,
		Password:     senha,
		DB:           db,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Valida se o Redis está vivo
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("falha ao conectar no Redis (%s): %w", endereco, err)
	}

	return &RedisStore{
		client: rdb,
		ttl:    ttl,
	}, nil
}

// Salvar armazena um ativo individualmente com TTL
func (r *RedisStore) Salvar(ativo *domain.Ativo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dados, err := json.Marshal(ativo)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "b3:snapshot:"+ativo.Ticker, dados, r.ttl).Err()
}

// SalvarLote utiliza REDIS PIPELINE para gravar centenas de ativos em um único round-trip
func (r *RedisStore) SalvarLote(ativos []*domain.Ativo) error {
	if len(ativos) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := r.client.Pipeline()

	// 1. Enfileira a gravação individual de cada ativo no pipeline
	for _, a := range ativos {
		dados, err := json.Marshal(a)
		if err == nil {
			pipe.Set(ctx, "b3:snapshot:"+a.Ticker, dados, r.ttl)
		}
	}

	// 2. Enfileira a gravação do snapshot consolidado completo (para leitura instantânea dos rankings)
	if todosJSON, err := json.Marshal(ativos); err == nil {
		pipe.Set(ctx, "b3:snapshot:todos", todosJSON, r.ttl)
	}

	// Executa todos os comandos de uma só vez na rede
	_, err := pipe.Exec(ctx)
	return err
}

// Obter busca um ativo individual no Redis pelo Ticker
func (r *RedisStore) Obter(ticker string) (*domain.Ativo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

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

// ListarTodos recupera a lista completa em < 1ms usando a chave agregada
func (r *RedisStore) ListarTodos() []*domain.Ativo {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Tenta ler direto do snapshot agregado
	val, err := r.client.Get(ctx, "b3:snapshot:todos").Result()
	if err == nil {
		var lista []*domain.Ativo
		if json.Unmarshal([]byte(val), &lista) == nil {
			return lista
		}
	}

	// 2. Fallback caso a chave agregada não exista
	keys, err := r.client.Keys(ctx, "b3:snapshot:*").Result()
	if err != nil {
		return nil
	}

	var lista []*domain.Ativo
	for _, k := range keys {
		if k == "b3:snapshot:todos" {
			continue
		}
		if v, err := r.client.Get(ctx, k).Result(); err == nil {
			var a domain.Ativo
			if json.Unmarshal([]byte(v), &a) == nil {
				lista = append(lista, &a)
			}
		}
	}
	return lista
}

// ObterCache recupera bytes brutos do Redis com timeout rápido
func (r *RedisStore) ObterCache(ctx context.Context, chave string) ([]byte, error) {
	ctxT, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.client.Get(ctxT, chave).Bytes()
}

// SalvarCache armazena bytes brutos no Redis com TTL específico
func (r *RedisStore) SalvarCache(ctx context.Context, chave string, dados []byte, ttl time.Duration) error {
	ctxT, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.client.Set(ctxT, chave, dados, ttl).Err()
}
