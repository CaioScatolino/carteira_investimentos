# Passo 14: Motor de Proventos Dinâmicos (Trailing 12 Months - TTM) e Resolução CVM/B3

## 🎯 Objetivo Arquitetural
Eliminar a dependência de APIs externas (Yahoo/Brapi) para obtenção de proventos.
Em vez de buscar um campo estático inexistente, derivar matematicamente a soma móvel de proventos dos últimos 12 meses ($\Sigma_{12M}$) a partir dos **eventos discretos oficiais de pagamento** da CVM e da B3.

---

## 📋 Checklist de Tarefas (TODO para Amanhã)

- [ ] **Tarefa 1: Modelagem no Domínio (`domain/provento.go`)**
  - Criar a struct `EventoProvento` com campos: `Identificador` (Ticker ou CNPJ), `Tipo` (`DIVIDENDO`, `JCP`, `RENDIMENTO_FII`), `DataPagamento` (`time.Time`) e `ValorUnitario` (`float64`).
- [ ] **Tarefa 2: Algoritmo de Janela Deslizante de 365 Dias (`domain.CalcularProventos12M`)**
  - Implementar o filtro: `DataPagamento >= Hoje.AddDate(-1, 0, 0)`.
  - Aplicar dedução de 15% de IR sobre JCP (Juros sobre Capital Próprio) para apuração de provento líquido exigido pelo método Décio Bazin.
- [ ] **Tarefa 3: Leitor Oficial de Rendimentos de FIIs da CVM (`internal/cvm`)**
  - Extrair os rendimentos mensais declarados no arquivo `inf_mensal_fii_complemento_YYYY.csv` da CVM (`Rendimento_Distribuido_Por_Cota`).
  - Somar as competências dos últimos 12 meses para cada CNPJ de FII.
- [ ] **Tarefa 4: Leitor de Eventos de Proventos de Ações**
  - Processar eventos corporativos em dinheiro (Dividendos e JCP deliberados e pagos).
- [ ] **Tarefa 5: Integração no `MarketSyncWorker`**
  - O worker injeta `Dividendos12M` calculado dinamicamente na entidade `domain.Ativo`.
  - Décio Bazin calcula: $\text{Preço Teto} = \frac{\text{Dividendos12M}}{0.06}$.
  - Derivação do Dividend Yield real auditado: $\text{DY} = \frac{\text{Dividendos12M}}{\text{PrecoFechamento B3}} \times 100$.
- [ ] **Tarefa 6: Validação Geral com Semáforo**
  - Executar `go run cmd/api/main.go` e validar a classificação automática dos ativos (`🟢 COMPRAR MAIS`, `🟡 MANTER`, `🔴 ALERTA`).

---

## 🧮 Fórmulas e Regras de Negócio

1. **Janela Deslizante:**
   $$\text{Proventos Líquidos}_{12M} = \sum_{\substack{e \in \text{Eventos} \\ e.\text{Data} \ge \text{Hoje} - 365\text{d}}} \text{ValorLíquido}(e)$$

2. **Dedução Fiscal (Regra Décio Bazin):**
   $$\text{ValorLíquido}(e) = \begin{cases} e.\text{ValorUnitario} \times 0.85, & \text{se } e.\text{Tipo} = \text{JCP} \\ e.\text{ValorUnitario}, & \text{se } e.\text{Tipo} \in \{\text{Dividendo}, \text{FII}\} \end{cases}$$

3. **Preço Teto Bazin:**
   $$\text{Preço Teto} = \frac{\text{Proventos Líquidos}_{12M}}{0.06}$$

4. **Margem de Segurança:**
   $$\text{Margem de Segurança (\%)} = \left(\frac{\text{Preço Teto} - \text{Cotação Fechamento B3}}{\text{Cotação Fechamento B3}}\right) \times 100$$
