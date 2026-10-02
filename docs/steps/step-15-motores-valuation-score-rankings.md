# Passo 15: Motores de Valuation Avançados, Score Fundamentalista e Rankings Segmentados

## 🎯 Objetivo Arquitetural
Expandir a capacidade analítica da plataforma B3 Core, transformando o sistema de um simples calculador de métricas individuais em um **motor institucional de screening e recomendação fundamentalista**.

A arquitetura agora conta com:
1. **Novos Motores de Valuation e Análise:**
   - **Gordon Growth Model (DDM):** Modelo de Desconto de Dividendos com taxa de desconto exigida ($k = 11\%$ a.a.) e crescimento sustentável perpétuo derivado do ROE e taxa de retenção ($\text{ROE} \times 40\%$, travado no teto conservador de $4.5\%$ a.a.).
   - **Joel Greenblatt (Earnings Yield):** Relação operacional $\frac{\text{LPA}}{\text{Preço Atual}}$, medindo a capacidade da empresa de gerar lucro frente ao investimento em cotações e comparando com a taxa livre de risco da renda fixa.
   - **Peter Lynch (PEG Ratio):** Relação $\frac{P/L}{\text{Crescimento 5A}}$, avaliando se a precificação da cotação está inflacionada em relação ao crescimento histórico de lucros.
2. **Score Fundamentalista Composto (0 a 100):**
   - Sistema de pontuação ponderada segmentado por classe de ativo:
     - **Ações:** Bazin (30 pts), Graham (30 pts), Lynch PEG (15 pts), Gordon DDM (15 pts) e ROE (10 pts).
     - **FIIs:** P/VP sobre laudo CVM (40 pts), Spread real vs NTN-B / IPCA+ 6.5% (40 pts) e Dividend Yield absoluto (20 pts).
     - **ETFs:** Liquidez institucional negociada no pregão diário da B3 (VOLTOT).
3. **Pareceres Educacionais Individuais por Escola:**
   - Em vez de uma caixa preta com apenas um semáforo genérico, cada escola emite seu próprio parecer (`APROVADO`, `ATENCAO`, `REPROVADO`) com métricas e justificativas transparentes.
4. **Rankings Segmentados Completos (100% dos Tickers Líquidos da B3):**
   - Identificação do BDI oficial da B3:
     - `CODBDI == "02"`: Lote Padrão Ações e Units.
     - `CODBDI == "12"`: Fundos de Investimento Imobiliário (FIIs).
     - `CODBDI == "14"`: Fundos de Índice (ETFs).
   - Divisão e ordenação decrescente por Score de 100% dos ativos analisados no pregão.

---

## 🧮 Modelagem Matemática e Critérios de Análise

### 1. Gordon Growth Model (DDM)
$$\text{Preço Teto}_{\text{Gordon}} = \frac{D_1}{k - g}$$
- $D_1 = \text{Dividendos}_{12M} \times (1 + g)$
- $k = 11\%$ a.a. (custo de oportunidade exigido para o mercado de capitais brasileiro)
- $g = \min(\text{ROE} \times 0.40, 4.5\%)$ (crescimento orgânico sustentável por reinvestimento)

### 2. Benjamin Graham
$$\text{Valor Intrínseco} = \sqrt{22.5 \times \text{LPA} \times \text{VPA}}$$
$$\text{Margem de Segurança} = \left(\frac{\text{Valor Intrínseco} - \text{Preço Atual}}{\text{Valor Intrínseco}}\right) \times 100$$

### 3. Décio Bazin
$$\text{Preço Teto}_{\text{Bazin}} = \frac{\text{Dividendos Líquidos}_{12M}}{0.06}$$
$$\text{Margem de Segurança} = \left(\frac{\text{Preço Teto} - \text{Preço Atual}}{\text{Preço Teto}}\right) \times 100$$

### 4. Peter Lynch (PEG Ratio)
$$\text{PEG Ratio} = \frac{P/L}{\text{Crescimento Anual do Lucro (\%) 5A}}$$
- $\text{PEG} \le 1.0$: Aprovado (Ação barata para o crescimento entregue).
- $1.0 < \text{PEG} \le 1.5$: Atenção (Preço justo para o crescimento).
- $\text{PEG} > 1.5$: Reprovado (Preço esticado em relação ao crescimento histórico).

### 5. Joel Greenblatt (Earnings Yield)
$$\text{Earnings Yield} = \frac{\text{LPA}}{\text{Preço Atual}} \times 100 = \frac{1}{P/L} \times 100$$
- $\ge 10\%$: Aprovado (Retorno em lucros superior ao custo de capital e renda fixa).
- $6\% \text{ a } 10\%$: Atenção (Retorno moderado).
- $< 6\%$: Reprovado (Retorno comprimido).

### 6. Fundos Imobiliários (FIIs)
- **P/VP sobre Laudo Oficial CVM:**
  - $0.85 \le P/VP \le 0.95$: Desconto ideal com margem de segurança.
  - $0.95 < P/VP \le 1.02$: Negociação em valor justo patrimonial.
  - $P/VP > 1.02$: Ágio patrimonial (reprovado/alerta).
- **Spread vs NTN-B:**
  - $\text{Spread} = \text{DY Efetivo} - 6.5\%$ (taxa do Tesouro IPCA+).

---

## 🏛️ Estrutura de Código Criada e Atualizada

| Pacote | Arquivo | Responsabilidade |
| :--- | :--- | :--- |
| `domain` | `ativo.go` | Inclusão de `ParecerItem`, `StatusParecer`, `Score`, `DY`, `PL`, `PVPReal`, `ROE`, `EarningsYield`, `PrecoTetoGordon`, `MargemBazin`, `MargemGraham` e suporte a ETFs no Semáforo. |
| `b3` | `parser.go` | Suporte ao BDI `"14"` (ETFs) junto com `"02"` (Ações) e `"12"` (FIIs). |
| `provider` | `b3_provider.go` | Mapeamento fidedigno de classe de ativo (`ClasseETF`, `ClasseFII`, `ClasseAcao`) no carregamento do COTAHIST. |
| `gordon` | `calculator.go` | Implementação do Analisador Gordon DDM conforme interface `domain.Analisador`. |
| `score` | `calculator.go` | Geração de pareceres analíticos individuais e pontuação ponderada consolidada (0 a 100). |
| `cmd/api` | `main.go` | Pipeline de auditoria completa de 388 ativos da B3 com ordenação e exibição de 3 rankings independentes. |

---

## 📊 Resultados da Validação

Execução em lote de 388 ativos líquidos oficiais da B3 e CVM:
- **Tempo de Execução:** ~14.7 segundos para varredura e auditoria integral de 388 ativos.
- **Divisão:** 200 Ações, 90 FIIs e 98 ETFs.
- **Top FIIs:** KNIP11, HGBS11, GARE11, BTHF11, ALZR11 com Score 100.
- **Top Ações:** ALLD3, PFRM3, MTRE3, PETR4 com Scores superiores a 75.
- **Top ETFs:** BOVA11, IVVB11, SMAL11 liderando em liquidez institucional.
