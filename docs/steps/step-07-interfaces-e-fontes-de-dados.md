# Step 07: Fontes de Dados do Mercado Brasileiro e Polimorfismo com Interfaces em Go

## 1. De Onde Vêm os Dados? A Arquitetura de Ingestão da B3

Para não depender de serviços pagos exorbitantes (como Bloomberg ou Economatica), a nossa plataforma utiliza uma arquitetura híbrida de 3 fontes de altíssima confiabilidade:

### A. Dados Cadastrais e Balanços Oficiais: Portal CVM (Comissão de Valores Mobiliários)
- **Origem:** `dados.cvm.gov.br` (Público, Gratuito e Oficial do Governo Federal).
- **O que fornece:**
  - **FIIs:** Informes Mensais com Valor Patrimonial por cota, rendimentos distribuídos, vacância física e número de cotistas.
  - **Ações:** Demonstrações Financeiras Padronizadas (DFP anual) e Informações Trimestrais (ITR) com Lucro Líquido, Patrimônio Líquido, Receita e Dívida.
- **Como o Go consome:** O Go descarrega ficheiros `.zip` compactados e processa os `.csv` em *streaming* com Goroutines, consumindo quase zero de RAM.

### B. Cotações em Tempo Real e Indicadores de Mercado
- **Brapi (`brapi.dev`):** API brasileira excelente com camada gratuita para tickers da B3, cotações e histórico de proventos.
- **Yahoo Finance (v8 endpoint nativo):** Endpoint público gratuito adicionando o sufixo `.SA` (ex: `PETR4.SA`, `VALE3.SA`, `MXRF11.SA`).

### C. Taxas Livres de Risco e Inflação: Banco Central do Brasil (BACEN)
- **API SGS do Banco Central:** Fornece a taxa Selic diária e a taxa média do Tesouro IPCA+ (NTN-B) em formato JSON sem necessidade de chave de API.

---

## 2. O Conceito Revolucionário de Interfaces em Go

Em TypeScript ou Java, quando se cria uma interface, é obrigatório declarar explicitamente a implementação:
```typescript
// TypeScript exige a palavra 'implements'
class BazinAnalisador implements Analisador { ... }
```

No **Golang**, as interfaces são **implícitas** (*Structural Typing* ou *Duck Typing* em tempo de compilação):
> *"Se anda como um pato e grasna como um pato, o compilador Go considera-o um pato."*

Se definirmos:
```go
type Analisador interface {
    Nome() string
    Executar(ativo *Ativo) error
}
```
Qualquer struct no nosso projeto que possua os métodos `Nome() string` e `Executar(ativo *Ativo) error` **automaticamente implementa a interface `Analisador`**, sem precisar de palavras-chave como `implements`!

Isto permite que o nosso motor itere sobre uma lista de analisadores:
```go
analisadores := []domain.Analisador{
    bazin.Novo(0.06),
    graham.Novo(),
    lynch.Novo(),
    fii.Novo(0.065),
}

for _, motor := range analisadores {
    motor.Executar(ativo) // Polimorfismo puro!
}
```
