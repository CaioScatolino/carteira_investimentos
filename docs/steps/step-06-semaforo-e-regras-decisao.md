# Step 06: Modelagem do Semáforo de Decisão e Enums Tipados

## 1. Por que Enums Tipados em Go?
Em JavaScript, é comum passar strings soltas como `"comprar"`, `"manter"` ou `"alerta"`. Se alguém cometer um erro de digitação como `"compra"` ou `"Comprar"`, o código falha silenciosamente.

Em Golang, criamos **tipos fortes** baseados em string:
```go
type StatusRecomendacao string

const (
    StatusComprarMais StatusRecomendacao = "COMPRAR_MAIS"
    StatusManter      StatusRecomendacao = "MANTER"
    StatusAlerta      StatusRecomendacao = "ALERTA"
)
```
Desta forma:
- O compilador garante que apenas valores válidos sejam atribuídos.
- No JSON da API / Server-Sent Events, o valor gerado é claro e padronizado.

---

## 2. A Lógica de Decisão do Semáforo (Bazin + Graham)

A nossa plataforma combina os dois maiores métodos de análise fundamentalista:

### Para Ações:
1. **`COMPRAR_MAIS` (Verde ✅):**
   A cotação está **abaixo** do Preço Teto de Bazin (garantindo yield $\ge 6\%$) **E** **abaixo** do Valor Intrínseco de Graham (garantindo margem patrimonial e de lucros).
2. **`MANTER` (Amarelo ⚠️):**
   Aprovado em apenas um dos critérios (ex: bom pagador de dividendos, mas acima do valor patrimonial de Graham, ou vice-versa).
3. **`ALERTA` (Vermelho 🚨):**
   Cotação acima de ambos os tetos. Não há margem de segurança.

### Para FIIs (Fundos Imobiliários):
Como FIIs distribuem 95% do lucro caixa por lei e não possuem lucros corporativos tradicionais (LPA), a auditoria avalia primariamente o fluxo de proventos de **Bazin**:
1. **`COMPRAR_MAIS` (Verde ✅):** Preço Atual $\le$ Preço Teto Bazin.
2. **`ALERTA / AGUARDAR` (Amarelo ⚠️):** Preço Atual $>$ Preço Teto Bazin.

---

## 3. O `switch` sem `break` do Go
No ecossistema JS/C, um `switch` sem a palavra `break` cai no caso de baixo (*fallthrough* bug).
No Go:
- Cada `case` termina automaticamente (o `break` é implícito).
- O `switch` pode ser usado sem variável condicional, funcionando como uma série de `if / else if` muito mais limpa e legível.
