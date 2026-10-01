# Step 09: Background Worker em Go (`time.Ticker`), Universo de Ativos e Cache com Redis

## 1. A Visão de Produção: Por que Caching é Obrigatório?

Se um utilizador abrir o nosso dashboard e tivermos de consultar 100 ativos na internet na hora:
- A interface ficaria lenta (latência de rede).
- Poderíamos sofrer bloqueios de taxa (*Rate Limit*).

### A Arquitetura do Funil (Etapa 1):
1. **Background Worker em Goroutine Perpétua**:
   Corre a cada 15 minutos em segundo plano (`time.NewTicker(15 * time.Minute)`).
2. **Ingestão Concorrente**:
   Descarrega as cotações e proventos do universo de ativos da B3 via Goroutines.
3. **Auditoria Determinística em Memória**:
   Roda os 4 motores (Bazin, Graham, Lynch, FII).
4. **Armazenamento no Snapshot Store (Redis / In-Memory)**:
   Grava os ativos auditados com TTL.
5. **Consulta Instantânea (< 1 ms)**:
   Quando a API ou a IA precisa dos dados, lê diretamente do Redis em milissegundos!

---

## 2. Padrão Arquitetural: Repository Pattern para Cache

Para que possas desenvolver na tua máquina local sem precisar de instalar o Redis no Windows agora, aplicamos o padrão de **Portas e Adaptadores (Clean Architecture)**:

```go
type SnapshotStore interface {
    Salvar(ativo *domain.Ativo) error
    Obter(ticker string) (*domain.Ativo, error)
    ListarTodos() []*domain.Ativo
}
```

- **Ambiente Local (Windows):** `InMemoryStore` (usando `sync.RWMutex` do Go, rápido, seguro contra race conditions e sem precisar de instalar nada).
- **Ambiente VPS (Produção):** `RedisStore` (conectado ao contentor `redis_central:6379` na VPS).

---

## 3. O Temporizador Nativo: `time.Ticker` em Go

Em Node.js usamos `setInterval()`. No Go usamos canais de tempo:
```go
ticker := time.NewTicker(15 * time.Minute)
defer ticker.Stop()

for range ticker.C {
    // Executa a cada 15 minutos pontualmente!
}
```
O canal `ticker.C` emite um sinal a cada ciclo sem consumir nenhum ciclo de CPU enquanto aguarda.
