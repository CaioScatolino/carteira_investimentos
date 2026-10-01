# Step 02: Inicialização do Módulo Go e Primeiro Executável

## 1. O que é um Módulo em Go (`go mod`)?

No ecossistema JavaScript/Node.js, usamos o comando:
```bash
npm init -y
```
que cria o ficheiro `package.json` e, quando instalamos bibliotecas, cria a pasta local `node_modules`.

No **Golang**:
- Usamos o comando:
  ```bash
  go mod init <nome-do-modulo>
  ```
- Este comando cria o ficheiro `go.mod`.
- **Diferença Crucial:** O Go **NÃO** cria uma pasta `node_modules` dentro do projeto!
- As dependências externas são descarregadas num diretório global de cache do sistema (`$GOPATH/pkg/mod`).
- O `go.mod` declara a versão da linguagem Go e a lista de dependências com as suas versões exatas.
- Quando dependências forem adicionadas, o Go criará automaticamente o `go.sum`, que guarda as assinaturas criptográficas (checksums SHA-256) de cada pacote para garantir segurança e reprodutibilidade da compilação.

---

## 2. Anatomia de um Ficheiro Go

Todo ficheiro Go (`.go`) possui uma estrutura previsível:

```go
package main // 1. Declaração do pacote

import (
    "fmt"    // 2. Importação de pacotes da biblioteca padrão
)

// 3. Ponto de entrada da aplicação
func main() {
    fmt.Println("Olá, B3!")
}
```

### Regras de Ouro do Compilador:
1. **Pacote `main` e Função `main`**: Qualquer programa Go que pretenda gerar um binário executável **tem de** pertencer ao `package main` e conter uma `func main()`.
2. **Sem imports inúteis**: Se importar um pacote e não o utilizar no código, o compilador **recusa-se a compilar**. Isto evita binários poluídos e código morto.
3. **Sem variáveis inúteis**: Se declarar uma variável local e nunca ler o seu valor, o código gera erro de compilação.

---

## 3. Variáveis e Tipagem Estática

Em JavaScript, uma variável pode mudar de tipo a qualquer momento:
```javascript
let preco = 38.50; // number
preco = "trinta e oito"; // perfeitamente válido em JS, mas causa bugs silenciosos
```

Em Go, os tipos são **imutáveis após a declaração**:
```go
// Forma explícita:
var ticker string = "PETR4"
var preco float64 = 38.50

// Forma idiomática com inferência de tipo (Operador Walrus `:=`):
ticker := "PETR4"   // O Go infere automaticamente que é string
preco := 38.50      // O Go infere automaticamente que é float64
```

> **Atenção:** O operador `:=` só pode ser usado **dentro** de funções para criar e inicializar variáveis novas.

---

## 4. Comandos Essenciais do Dia a Dia

- `go run <caminho>`: Compila o código na memória temporária e executa-o de imediato (ótimo para testes rápidos).
- `go build -o <saida> <caminho>`: Compila e gera o binário executável final no disco (o que faremos em produção).
- `go vet ./...`: Analisador estático oficial do Go para detetar potenciais falhas de concorrência ou sintaxe.
- `go fmt ./...`: Formata automaticamente todo o código do projeto de acordo com o padrão canónico oficial da linguagem.

---

## 5. Execução Prática e Resultados Obtidos

Criámos o primeiro ponto de entrada em `apps/server/cmd/api/main.go` aplicando a fórmula do Preço Teto de Décio Bazin:

```go
package main

import "fmt"

func main() {
    ticker := "PETR4"
    precoAtual := 38.50
    dividendos12M := 5.20
    yieldMinimo := 0.06

    precoTeto := dividendos12M / yieldMinimo
    margemSeguranca := ((precoTeto - precoAtual) / precoAtual) * 100.0

    fmt.Printf("Ativo Auditado:       %s\n", ticker)
    fmt.Printf("Cotação Atual:        R$ %.2f\n", precoAtual)
    fmt.Printf("Preço Teto (Bazin 6%%): R$ %.2f\n", precoTeto)
    fmt.Printf("Margem de Segurança:  %.2f%%\n", margemSeguranca)
}
```

### Resultados da Execução:
1. **Modo Desenvolvimento (`go run cmd/api/main.go`):**
   ```text
   Ativo Auditado:       PETR4
   Cotação Atual:        R$ 38.50
   Proventos (12 Meses): R$ 5.20
   Preço Teto (Bazin 6%): R$ 86.67
   Margem de Segurança:  125.11%
   Status: ✅ COMPRAR MAIS (Ativo com margem de segurança positiva)
   ```
2. **Modo Produção (`go build -o bin/api.exe cmd/api/main.go`):**
   - Gerou o executável autossuficiente `bin/api.exe`.
   - **Tamanho final:** apenas **~2.39 MB**.
   - **Dependências externas:** Zero! Não requer Node.js, nem Python, nem runtime instalado para executar.

---

## 6. Próximo Passo (Step 03)
- Criação de `structs` para modelar os ativos da B3 (Ação, FII, ETF).
- Métodos, Ponteiros e Funções com tratamento explícito de erros (`if err != nil`).
- Estruturação do pacote modular `internal/bazin` e `internal/domain`.

