# Step 08: Ingestão de Dados de Mercado com Goroutines, Channels e HTTP Nativo

## 1. Concorrência no Go vs Assincronismo no JavaScript

No JavaScript / Node.js:
- O modelo baseia-se num **Event Loop de thread única**.
- Quando fazes `await Promise.all([fetch1, fetch2])`, o Node delega o I/O ao sistema operacional, mas a execução do código de parsing ocorre sequencialmente na mesma thread.

No **Golang**:
- O Go possui um **Runtime Scheduler em M:N** (M goroutines distribuídas sobre N threads reais do processador).
- Cada **Goroutine** pesa apenas **~2 KB** de memória inicial (em comparação a 1 MB de uma thread nativa de SO).
- Disparar 10, 100 ou 1.000 goroutines tem um custo insignificante de CPU e memória.

---

## 2. A Tríade da Concorrência Profissional em Go

Quando várias tarefas rodam ao mesmo tempo, precisamos de sincronização e comunicação segura:

### A. A Palavra-Chave `go`
Basta colocar `go` antes de uma função para que ela seja executada em paralelo em segundo plano:
```go
go buscarCotacao("PETR4")
```

### B. `sync.WaitGroup` (O equivalente elegante ao `Promise.all`)
Garante que o programa principal não termine antes que as goroutines terminem o seu trabalho:
```go
var wg sync.WaitGroup

wg.Add(1) // Avisa: "Mais uma goroutine em execução"
go func() {
    defer wg.Done() // Avisa: "Terminei o meu trabalho!"
    // ... buscar cotação ...
}()

wg.Wait() // Bloqueia a execução até que todas as goroutines tenham chamado Done()
```

### C. Channels (`chan`) - Comunicação Thread-Safe
O lema oficial dos criadores do Go:
> *"Não comunique compartilhando memória; compartilhe memória comunicando."*

Se 5 goroutines tentarem escrever no mesmo slice `carteira = append(carteira, ativo)` ao mesmo tempo, ocorrerá uma **Race Condition** (conflito de memória que corrompe os dados).

A solução idiomática é usar um **Canal (`chan`)**:
- Cada goroutine atua como remetente e envia o ativo pelo canal: `canal <- ativo`
- A função principal consome os ativos do canal com segurança: `ativo := <-canal`

---

## 3. O Cliente HTTP Nativo do Go (`net/http`)
O Go não requer Axios, Fetch ou Request. A biblioteca padrão inclui `net/http` e `encoding/json` que são extremamente velozes e seguras.

**Regra de Ouro em Produção:**
Nunca usar `http.Get(...)` direto sem timeout. Configuramos sempre um cliente com tempo limite explícito:
```go
cliente := &http.Client{
    Timeout: 10 * time.Second,
}
```
