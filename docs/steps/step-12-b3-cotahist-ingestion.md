# Passo 12: Ingestão de Cotações Diárias em Lote da B3 (COTAHIST)

## 🎯 Objetivo
Substituir o Yahoo Finance na **Frente 2 (Scanner Geral em Lote)** pelo boletim diário oficial da B3 (**COTAHIST_D**).
Isso elimina milhares de requisições de rede, contorna bloqueios de IP (HTTP 429) e fornece cotações de fechamento oficiais e volumes negociados de 100% dos ativos em um único download diário de ~450 KB.

---

## 🏗️ Estrutura do Pacote `internal/b3`

```
apps/server/internal/b3/
├── cotahist.go  # Entidades: CotacaoDiaria, ArquivoDiarioB3
├── client.go    # Download HTTP com resolução retroativa de pregão útil
└── parser.go    # Leitor em stream com bufio.Scanner e layout posicional
```

---

## 📋 Layout Posicional Oficial (Tabela de Offsets)

| Campo | Posição (Bytes) | Formato | Descrição |
| :--- | :---: | :---: | :--- |
| **`TIPREG`** | `0 a 2` | `X(02)` | `01` = Registro de cotação |
| **`DTPREG`** | `2 a 10` | `9(08)` | Data do pregão (`AAAAMMDD`) |
| **`CODBDI`** | `10 a 12` | `X(02)` | `02` = Lote padrão de ações/ETFs/BDRs; `12` = FIIs |
| **`CODNEG`** | `12 a 24` | `X(12)` | Código de negociação / Ticker (ex: `PETR4`, `MXRF11`) |
| **`TPMERC`** | `24 a 27` | `9(03)` | `010` = Mercado à Vista |
| **`NOMRES`** | `27 a 39` | `X(12)` | Nome resumido da empresa/fundo |
| **`PREABE`** | `56 a 69` | `11v99` | Preço de Abertura (divide por 100.0) |
| **`PREMAX`** | `69 a 82` | `11v99` | Preço Máximo do dia (divide por 100.0) |
| **`PREMIN`** | `82 a 95` | `11v99` | Preço Mínimo do dia (divide por 100.0) |
| **`PREULT`** | `108 a 121`| `11v99` | **Preço de Fechamento / Último negócio** (divide por 100.0) |
| **`VOLTOT`** | `170 a 188`| `16v99` | **Volume Financeiro Total Negociado** (divide por 100.0) |

---

## 💻 Ponto de Entrada CLI
- Comando manual para disparar a sincronização a qualquer momento:
  ```powershell
  go run cmd/cli/main.go --sync-b3
  ```
