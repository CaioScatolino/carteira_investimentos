# Passo 22: Análise Gráfica, Indicadores Técnicos Institucionais & Preço Teto no Gráfico (FASE 16)

## 📌 Contexto & Objetivos
Unir o mundo da **Análise Fundamentalista** (Valuation, Preço Teto Bazin, Valor Intrínseco de Graham, Margens de Segurança e Proventos) com o **Price Action e a Análise Técnica Clássica** (Velas / Candlesticks, Médias Móveis Institucionais e Osciladores de Sobrevenda).

Nenhuma plataforma aberta no Brasil combina, na mesma interface de velas, a linha horizontal do Preço Teto de dividendos com médias de tendência e osciladores de exaustão.

---

## 🏗️ Arquitetura Implementada

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       1. CLIENTE (FRONTEND NEXT.JS)                         │
│   • Componente 'AssetChartModal' com 'lightweight-charts' (TradingView)     │
│   • Velas (Candles) + Linhas MMA 8/21/200 + Linha Dourada Teto Bazin + IFR   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ GET /api/v1/ativos/{ticker}/grafico
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         2. BACKEND GO (APPS/SERVER)                         │
│               (internal/technical/service.go e indicators.go)               │
└──────────────────┬───────────────────────────────────────┬──────────────────┘
                   │                                       │
    (Cache HIT)    ▼                        (Cache MISS)   ▼
┌───────────────────────────────┐        ┌────────────────────────────────────┐
│      3. REDIS (VPS CENTRAL)   │        │     4. YAHOO FINANCE HISTÓRICO     │
│   • Chave: "b3:chart:PETR4:1d"│        │   • Endpoint público oficial v8    │
│   • Retorno em < 1ms          │        │   • Baixa 1 ano de candles (OHLCV) │
│   • TTL: 4 horas              │        └─────────────────┬──────────────────┘
└───────────────────────────────┘                          │
                                                           ▼
                                         ┌────────────────────────────────────┐
                                         │   5. MOTOR MATEMÁTICO EM GO        │
                                         │   • Calcula MMA 8, 21 e 200        │
                                         │   • Calcula IFR 14 de Wilder       │
                                         │   • Injeta Preço Teto Bazin/Graham │
                                         │   • Avalia Sinal Técnico Confluente│
                                         │   • Salva no Redis e responde      │
                                         └────────────────────────────────────┘
```

---

## 📊 Indicadores e Fórmulas Matemáticas

### 1. Média Móvel Aritmética (MMA)
* **Tempo Gráfico 1D (Diário):**
  * **MMA 8:** Rastreamento de momentum e primeiro suporte de curto prazo.
  * **MMA 21:** Linha d'água de tendência de médio prazo.
  * **MMA 200:** Tendência primária institucional.
* **Tempo Gráfico 1S (Semanal - Padrão Institucional):**
  * **MMA 20 Semanal (Gatilho de Direção — ~5 meses):** Substitui a 8 e 21 no semanal. Preço acima = fluxo comprador dominante; preço abaixo = perda de tração e início de correção.
  * **MMA 52 Semanal (O Eixo de 1 Ano):** Média exata das 52 semanas do ano. Preço trabalhando em direção a ela sinaliza saída de euforia e entrada na área de desconto anual.
  * **MMA 200 Semanal (O Chão de Concreto — ~4 anos):** Raramente rompida por empresas sólidas da B3; região de compra pesada de fundo de ciclo secular.

### 2. Bandas de Bollinger (20, 2) no Semanal
Calculadas com base na MMA 20 semanas e $\pm 2$ desvios padrão ($\sigma$):
* **Banda Superior (Área de Venda / Esticada):** $\text{MMA 20} + 2\sigma$. Preço tocando esta banda acima do Teto Consolidado indica ativo estatisticamente sobrecomprado no curto prazo.
* **Linha Central:** Própria MMA 20 semanal como eixo de suporte/resistência dinâmico.
* **Banda Inferior (Área de Compra Técnica):** $\text{MMA 20} - 2\sigma$. Preço tocando esta banda abaixo do Preço Teto Bazin/Consolidado gera o alinhamento perfeito entre desconto fundamentalista e exaustão técnica de queda.

### 3. Índice de Força Relativa (IFR / RSI de 14 períodos de Wilder)
Mede a magnitude dos ganhos em relação às perdas para detectar zonas de esgotamento:
$$\text{RS} = \frac{\text{Média Suavizada de Ganhos}}{\text{Média Suavizada de Perdas}}$$
$$\text{IFR} = 100 - \left( \frac{100}{1 + \text{RS}} \right)$$
* **Semanal (1S):** Linha 35 (Sobrevenda institucional) e Linha 70 (Sobrecompra).
* **Diário (1D):** Linha 30 (Sobrevenda clássica) e Linha 70 (Sobrecompra).
* **Autoscale Travado:** Garantido no intervalo [20, 80] para que ambas as linhas horizontais estejam sempre visíveis.

### 3. Confluência Técnica & Fundamentalista (O Sinal de Ouro)
Cruza em tempo real a margem de segurança do valuation com o Price Action:
* **`ZONA_COMPRA_SOBREVENDIDO`**: $\text{Preço} \le \text{Preço Teto Bazin} \quad \text{E} \quad \text{IFR} \le 35$. Momento assimétrico de aporte onde uma empresa lucrativa e boa pagadora de proventos é vendida com desconto em meio ao pânico do mercado.
* **`COMPRA_TENDENCIA_ALTA`**: $\text{Preço} \le \text{Preço Teto Bazin} \quad \text{E} \quad \text{Preço} \ge \text{MMA 21} \ge \text{MMA 200}$. Compra na tendência primária de alta com margem de segurança positiva.
* **`SOBRECOMPRA_ALERTA`**: $\text{IFR} \ge 70$. Risco de correção iminente.
* **`AGIO_SEM_MARGEM`**: $\text{Preço} > \text{Preço Teto Bazin}$. Aguardar recuo técnico antes de aportar.

---

## 🔌 Endpoints REST

* **`GET /api/v1/ativos/{ticker}/grafico`**
  * **Query Params:**
    * `intervalo`: `1d` (padrão) ou `1wk`.
    * `periodo`: `6mo`, `1y` (padrão), `2y`, `5y`.
  * **Retorno:**
    * `candles`: Array com `{time, open, high, low, close, volume}`.
    * `mma8`, `mma21`, `mma200`: Arrays com `{time, value}`.
    * `rsi14`: Array com `{time, value}`.
    * `preco_teto_consolidado`: Preço teto consensual multi-modelo (Bazin 5A, Graham, Gordon).
    * `modelos_teto`: Lista descritiva dos modelos e ponderações que compõem o consenso.
    * `preco_teto_bazin`, `valor_graham`, `margem_seguranca`.
    * `sinal_tecnico`, `descricao_sinal`.
  * **Renderização Gráfica do Teto Consolidado:**
    * Renderizado diretamente no gráfico via `LineSeries` contínua em tom dourado champanhe (`#d4af37`), traçando uma linha reta horizontal perfeita do início ao fim da série.
    * Mantém o auto-scale do eixo Y perfeitamente calibrado, garantindo visibilidade imediata mesmo quando o teto de valuation estiver acima dos topos históricos do ativo.

---

## 🎨 Componentes Frontend

* **`AssetChartModal.tsx`**: Modal imersivo com tema escuro (`#070709`), velas em verde esmeralda e vermelho carmim, sub-gráfico sincronizado de IFR 14 com níveis 30/70, alternador Diário/Semanal e checkboxes para ligar/desligar indicadores.
* **Acesso Unificado:**
  1. **Na Tabela do Mercado B3 (`MarketTable`):** Botão gráfico de barras embutido ao lado de cada ticker.
  2. **Na Carteira do Usuário (`PortfolioTable`):** Botão de gráfico na coluna de ações ao lado de Aporte e Histórico.
  3. **No Modal Analítico (`AssetDetailModal`):** Botão em destaque "Gráfico Técnico (Candles & Teto)".
