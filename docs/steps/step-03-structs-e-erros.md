# Step 03: Structs, Métodos, Ponteiros e Tratamento Explícito de Erros

## 1. Por que não existem Classes em Go?
Em JavaScript ou Python, estamos habituados a `class`, `constructor`, `this` e herança (`extends`).
O Go foi desenhado para ser simples e previsível. Em vez de classes e heranças complexas, o Go utiliza:
- **`struct`**: Uma coleção nomeada de campos de dados.
- **Composição**: Em vez de herdar, embutimos structs dentro de outras structs.
- **Métodos com Receivers**: Funções vinculadas diretamente a uma `struct`.

---

## 2. A Regra de Visibilidade (Exported vs Unexported)
O Go não tem palavras-chave como `public`, `private` ou `export`. A visibilidade é decidida pela **primeira letra do identificador**:

- **Letra Maiúscula (`Ticker`, `PrecoAtual`, `Calcular`):** É **PÚBLICO** (exportado). Pode ser acedido por qualquer outro pacote.
- **Letra Minúscula (`ticker`, `precoAtual`, `calcular`):** É **PRIVADO** (não exportado). Só pode ser lido ou executado dentro do mesmo pacote (`package`).

---

## 3. Ausência de Exceptions (`try/catch`)
No ecossistema JS/Node.js, se uma função disparar um `throw new Error(...)` e não houver um `catch`, a aplicação inteira pode colapsar.

Em Go:
1. **Funções podem retornar mais do que um valor** ao mesmo tempo.
2. O padrão idiomático da linguagem para operações que podem falhar é retornar `(resultado, error)`.
3. Se a operação tiver sucesso, o erro retornado é `nil` (o equivalente ao `null` do JS).
4. O tratamento é sempre explícito através da famosa verificação:
```go
resultado, err := operacao()
if err != nil {
    // Tratar a falha conscientemente
    return err
}
```

---

## 4. O que é um Receiver (Método da Struct)?
Para associar uma função a uma struct, declaramos um *receiver* antes do nome da função:

```go
// (a Ativo) é o receiver. É como se fosse o 'this' em JS, mas nomeado explicitamente por ti.
func (a Ativo) CalcularPrecoTetoBazin(yield float64) (float64, error) {
    if yield <= 0 {
        return 0, errors.New("o yield mínimo não pode ser zero nem negativo")
    }
    return a.Dividendos12M / yield, nil
}
```

---

## 5. Ponteiros Desmistificados (`&` e `*`)
- **Sem ponteiro (Passagem por Valor):** O Go cria uma cópia exata de toda a memória da struct. Se alterares um campo dentro da função, a struct original permanece inalterada.
- **Com ponteiro (`*`):** Passamos o endereço de memória (`&`) da struct. É ultrarrápido (não duplica memória) e permite que a função modifique o estado da struct original.
