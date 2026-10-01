# Step 05: Ponteiros na Prática (Passagem por Valor vs Referência) e o Método de Graham

## 1. A Maior Pegadinha do Go para quem vem do JavaScript

Em JavaScript:
```javascript
const carteira = [{ ticker: "PETR4", precoTeto: 0 }];
const ativo = carteira[0]; // ativo aponta para o mesmo objeto na memória (referência)
ativo.precoTeto = 86.67;
console.log(carteira[0].precoTeto); // Imprime 86.67! O array original foi alterado.
```

Em **Golang**:
Structs são tipos de valor (*Value Types*). Quando fazes:
```go
for _, ativo := range carteira {
    ativo.PrecoTeto = 86.67 // ATENÇÃO: 'ativo' é apenas uma CÓPIA local descartável!
}
// Se consultares carteira[0].PrecoTeto, continuará ZERO (0.0)!
```

### Como resolver?
Existem duas formas idiomáticas:
1. **Acessar diretamente o índice do slice:**
   ```go
   for i := range carteira {
       carteira[i].PrecoTeto = teto // Altera a memória da fatia diretamente
   }
   ```
2. **Utilizar Slices de Ponteiros (`[]*domain.Ativo`):**
   ```go
   carteira := []*domain.Ativo{ &petr4, &vale3 }
   for _, ativo := range carteira {
       ativo.PrecoTeto = teto // Agora sim! 'ativo' guarda o endereço de memória (&)
   }
   ```

---

## 2. O que são os operadores `&` e `*`?
- **`&` (E comercial / Address-of):** "Dá-me o endereço de memória onde esta variável está guardada".
- **`*` (Asterisco na tipagem / Pointer type):** "Esta variável não guarda o valor diretamente, mas sim um endereço de memória para um valor do tipo X" (ex: `*domain.Ativo`).
- **`*` (Asterisco antes da variável / Dereference):** "Lê ou escreve no valor real que está guardado naquele endereço".

---

## 3. O Método Benjamin Graham (Valor Intrínseco)
Benjamin Graham (mentor de Warren Buffett) postulou que uma ação defensiva não deve negociar acima de:
$$P/L \le 15 \quad \text{e} \quad P/VP \le 1.5$$
Multiplicando ambos os limites máximos: $15 \times 1.5 = 22.5$.
A fórmula do Valor Intrínseco (VI) é:
$$VI = \sqrt{22.5 \times LPA \times VPA}$$

Onde:
- **LPA:** Lucro por Ação.
- **VPA:** Valor Patrimonial por Ação.
- Se a empresa tiver lucro negativo ou patrimônio líquido negativo, a fórmula é matematicamente inválida (não se extrai raiz de número negativo em finanças).

---

## 4. O Pacote Nativo `"math"`
Para calcular a raiz quadrada em Go, usamos a biblioteca padrão:
```go
import "math"

resultado := math.Sqrt(22.5 * lpa * vpa)
```
Como Go é fortemente tipado, `math.Sqrt` exige e devolve estritamente `float64`.
