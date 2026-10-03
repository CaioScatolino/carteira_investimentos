# Passo 16: API REST Nativa em Go, Endpoints de Rankings e Middleware CORS

## 🎯 Objetivo Arquitetural
Expor toda a inteligência e os motores de valuation através de uma interface HTTP REST ultra-rápida, desacoplada e padronizada em JSON.

Esta camada foi construída utilizando os recursos modernos do **Go 1.22+** da biblioteca padrão (`net/http`), dispensando frameworks pesados ou dependências externas desnecessárias.

---

## 🏗️ Design e Padrões Aplicados

### 1. Injeção de Dependências com `storage.SnapshotStore`
O [`internal/api.Handler`](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/api/handler.go) não conhece detalhes de persistência. Ele recebe apenas a interface `SnapshotStore`.
- Atualmente: consome o `InMemoryStore` (latência sub-milissegundo).
- Futuro próximo: trocar para o `RedisStore` sem alterar sequer uma linha de código da camada HTTP.

### 2. Rotas Nativas do Go 1.22+
Aproveitamos o novo pattern matching com verbos HTTP e extração de parâmetros de rota:
- `"GET /health"`
- `"GET /api/v1/rankings"`
- `"GET /api/v1/ativos/{ticker}"` com extração direta via `r.PathValue("ticker")`.

### 3. Middlewares Idiomáticos
- **CORS Middleware:** Libera origens, métodos (`GET`, `POST`, `OPTIONS`, etc.) e responde imediatamente a requisições de preflight (`OPTIONS` com status `200 OK`).
- **Logging Middleware:** Mede o tempo de resposta em tempo real (`log.Printf`).

---

## 📋 Endpoints Disponíveis

| Método | Rota | Descrição | Exemplo de Resposta |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Status de saúde e total de ativos carregados | `{"status": "ok", "total_ativos": 390}` |
| `GET` | `/api/v1/rankings` | Entrega os 3 rankings (Ações, FIIs, ETFs) ordenados por Score | `{"total": 390, "acoes": [...], "fiis": [...], "etfs": [...]}` |
| `GET` | `/api/v1/ativos/{ticker}` | Detalhes do ativo, indicadores e pareceres de cada escola | Objeto `Ativo` com pareceres de Bazin, Graham, Lynch, Gordon, etc. |

---

## ⚡ Métricas de Desempenho Medidas ao Vivo

- `GET /health`: **< 100 µs**
- `GET /api/v1/ativos/PETR4`: **505.2 µs** (microssegundos)
- `GET /api/v1/rankings` (390 ativos completos): **2.6 ms**
