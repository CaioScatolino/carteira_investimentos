# Step 10: Persistência no MySQL e Catálogo Oficial de Ativos (Ticker <-> CNPJ)

## 1. Como o Golang se Conecta a Bancos de Dados (Driver Nativo vs ORM)

No ecossistema Node.js / JavaScript, é comum usar ORMs pesadas (como Prisma ou TypeORM) que geram overhead e abstrações desnecessárias.

Em **Golang**, a biblioteca padrão possui um dos melhores gerenciadores de banco de dados do mundo: o pacote **`database/sql`**.

### Por que o Driver Nativo (`database/sql` + `go-sql-driver/mysql`) é a Escolha de Alta Performance?
1. **Connection Pooling Nativo**: O Go gerencia automaticamente uma fila de conexões abertas, reusando sockets TCP sem consumir CPU nem abrir conexões desenfreadas.
2. **Consumo de Memória (< 20 MB)**: ORMs como GORM utilizam muita *reflection* em tempo de execução, consumindo mais RAM e dificultando o controle de queries complexas. Com `database/sql`, operamos no modelo mais veloz e enxuto possível para a nossa VPS.
3. **Transações ACID Seguras**: Controle direto de `tx, err := db.Begin()`, `tx.Commit()` e `tx.Rollback()`.

---

## 2. O que é o Import com Underline (`_ "github.com/go-sql-driver/mysql"`)?

Em Go, quando importamos com `_`:
```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)
```
Estamos a dizer ao compilador:
> *"Executa apenas a função `init()` interna deste pacote para registrar o driver MySQL dentro do `database/sql` da biblioteca padrão, sem expor funções diretamente."*

---

## 3. Estrutura da Tabela `assets` (Catálogo Oficial B3)

Criámos o script de migração em `apps/server/scripts/migrations/001_create_assets_table.sql`:
- **`ticker`**: Código de negociação único na B3 (ex: `PETR4`, `MXRF11`).
- **`cnpj`**: Identificador oficial utilizado pela CVM.
- **`classe`**: `ACAO`, `FII`, `ETF` ou `BDR`.
- **`tipo`**: `ON`, `PN`, `UNT` para ações; `CI` para cotas de fundos.
- **`razao_social`**: Nome empresarial registrado.
- **`codigo_cvm`**: Código de registro na CVM.

---

## 4. Variáveis de Ambiente e Configuração
Para suportar tanto o ambiente de desenvolvimento local (no Windows) quanto a VPS de produção sem alterar código:
- `DB_HOST`: `127.0.0.1` (local) ou `mysql_central` (VPS via Docker network)
- `DB_PORT`: `3306`
- `DB_USER`: `root`
- `DB_PASSWORD`: senha do MySQL
- `DB_NAME`: `carteira_investimentos`
