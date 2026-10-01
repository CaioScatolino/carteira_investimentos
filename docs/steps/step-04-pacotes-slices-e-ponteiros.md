# Step 04: Organização em Pacotes (`internal/`), Slices e Tags JSON

## 1. O Padrão de Arquitetura Go (`cmd/` vs `internal/`)

Nos ecossistemas convencionais, criamos pastas genéricas. No ecossistema profissional de Go existe um padrão da indústria (Standard Go Layout):

- **`cmd/`**: O ponto de partida da aplicação (*Entrypoint*). Deve conter o mínimo de código possível — apenas inicializar configurações, dependências e disparar o servidor.
- **`internal/`**: Diretório protegido pelo próprio compilador do Go. Qualquer pacote dentro de `internal/` **não pode** ser importado por módulos externos, garantindo que as regras de negócio permaneçam estritamente privadas ao nosso projeto.

---

## 2. Pacotes Locais e Importação

Quando inicializámos o módulo com:
```bash
go mod init carteira_investimentos/server
```
Definimos que o prefixo de todos os nossos ficheiros internos é `carteira_investimentos/server`.

Por isso, quando criamos `apps/server/internal/domain/ativo.go` com `package domain`, nós importamo-lo noutros ficheiros assim:
```go
import "carteira_investimentos/server/internal/domain"
```

---

## 3. O que são Slices em Go? (Adeus `array.map()` e `array.filter()`)

Em JavaScript:
```javascript
const carteira = [ativo1, ativo2];
const recomendados = carteira.filter(a => a.preco <= a.teto);
```

Em Go, usamos **Slices** (fatias dinâmicas de arrays):
```go
// Um slice de ativos
carteira := []domain.Ativo{ petr4, vale3, mxrf11 }

// Iteração idiomática em Go:
for i, ativo := range carteira {
    fmt.Printf("[%d] %s: R$ %.2f\n", i, ativo.Ticker, ativo.PrecoAtual)
}
```
> **Por que Go não tem `.map()` ou `.filter()` nativos?**
> Funções como `.map()` e `.filter()` alocam novas funções (callbacks) e novos arrays na memória a cada chamada. O loop `for ... range` do Go é compilado diretamente para instruções assembly ultra-otimizadas que consomem quase zero de CPU e memória.

---

## 4. Tags JSON (`json:"ticker"`)

Como o nosso back-end irá enviar dados para o Next.js via JSON / Server-Sent Events (SSE), precisamos de ensinar o Go como serializar os nomes dos campos:

```go
type Ativo struct {
    Ticker     string  `json:"ticker"`
    PrecoAtual float64 `json:"preco_atual"`
}
```
- Em Go, o campo é PascalCase (`Ticker`) para ser público.
- A anotação entre crases (backticks) `` `json:"ticker"` `` diz ao serializador para gerar `{ "ticker": "PETR4", "preco_atual": 38.50 }` em minúsculas (camelCase/snake_case), padrão exigido pelas APIs web.
