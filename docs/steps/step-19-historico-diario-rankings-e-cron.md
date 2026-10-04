# Passo 19: Histórico Diário de Rankings Pós-Pregão (Cron 19h) & Fator Temporal no Valuation

## 🎯 Objetivo Arquitetural
Implementar um pipeline diário automatizado executado às **19:00** (de segunda a sexta-feira, logo após o fechamento do pregão da B3 e homologação das cotações e proventos), gravando um **snapshot consolidado diário dos rankings e múltiplos** no MySQL, e utilizando a série temporal histórica como uma dimensão adicional nos modelos de valuation (**Consistência Fundamentalista & Momentum de Ranking**).

---

## 🕒 Por que 19:00 (Pós-Pregão)?
1. **Fechamento Definitivo da B3:** O pregão regular da B3 encerra às 17h/18h. Entre 18h e 19h, ocorrem os ajustes pós-mercado, leilões de encerramento e consolidação dos boletins oficiais (COTAHIST e informes).
2. **Proventos & Fatos Relevantes:** As empresas e gestoras de FIIs divulgam anúncios de dividendos, JCP e fatos relevantes costumeiramente após as 18:00 (after market).
3. **Imutabilidade do Pregão:** Ao rodar às 19:00, o registro do dia representa fielmente o estado final do mercado para aquela data de referência.

---

## 🗄️ Modelagem da Tabela MySQL (`historico_rankings_diarios`)

```sql
CREATE TABLE IF NOT EXISTS historico_rankings_diarios (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    data_pregao DATE NOT NULL,
    ticker VARCHAR(10) NOT NULL,
    classe ENUM('ACAO', 'FII', 'ETF') NOT NULL,
    posicao_ranking INT NOT NULL,
    score DECIMAL(5,2) NOT NULL,
    preco_fechamento DECIMAL(10,2) NOT NULL,
    dy_12m DECIMAL(8,4) DEFAULT 0.0000,
    payout DECIMAL(8,2) DEFAULT 0.00,
    pl DECIMAL(10,2) DEFAULT 0.00,
    pvp DECIMAL(10,2) DEFAULT 0.00,
    roe DECIMAL(8,2) DEFAULT 0.00,
    roic DECIMAL(8,2) DEFAULT 0.00,
    preco_teto_bazin DECIMAL(10,2) DEFAULT 0.00,
    valor_graham DECIMAL(10,2) DEFAULT 0.00,
    status ENUM('COMPRAR_MAIS', 'MANTER', 'ALERTA') NOT NULL,
    is_provento_atipico BOOLEAN DEFAULT FALSE,
    alerta_risco VARCHAR(255) DEFAULT NULL,
    pareceres_json JSON DEFAULT NULL,
    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Chave única para garantir idempotência (evita duplicação caso a rotina execute novamente)
    UNIQUE KEY uq_data_ticker (data_pregao, ticker),
    INDEX idx_data_classe_posicao (data_pregao, classe, posicao_ranking),
    INDEX idx_ticker_data (ticker, data_pregao DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## 📈 Fator Temporal no Valuation (Momentum Fundamentalista & Consistência)

Atualmente, o valuation avalia um retrato instantâneo (*cross-section*) de hoje. Com a série temporal dos dias anteriores, introduzimos métricas de segunda ordem:

### 1. Fator de Consistência de Score (Estabilidade)
- **Problema:** Um ativo pode entrar no Top 5 em um único dia por causa de uma distorção pontual de fechamento ou evento de liquidez efêmero ("voo de galinha").
- **Solução:** Calcular a permanência e desvio padrão do score nos últimos 30 pregões:
  $$\text{Consistência} = \frac{\text{Dias no Top 20 nos últimos 30 pregões}}{30}$$
- Ativos com alta consistência ganham bonificação de resiliência.

### 2. Detecção de Turnaround & Upgrades de Ranking
- Comparar a posição do ranking atual ($Pos_t$) com a posição de 7 e 30 dias atrás ($Pos_{t-7}$, $Pos_{t-30}$):
  $$\Delta \text{Ranking}_{30D} = Pos_{t-30} - Pos_t$$
- Se um ativo saltou da 95ª para a 15ª posição de forma consistente, o sistema sinaliza **"Empresa em Momentum Fundamentalista Positivo (Upgrade Contínuo)"**.

### 3. Divergência Preço vs. Fundamentos (Janela de Oportunidade)
- Quando o Score Fundamentalista e o Preço Teto se mantêm altos ou sobem, mas a cotação do mercado sofre queda pontual (estresse de mercado, aversão a risco macro):
  - **Sinal:** Margem de segurança ampliando sem deterioração da qualidade contábil.
  - **Classificação:** Oportunidade institucional de compra com desconto.

---

## 🚀 Próximos Passos de Implementação
1. **Migrations / Criação da Tabela:** Script DDL executado no MySQL local e na VPS.
2. **Cron Scheduler em Go:** Implementar rotina de agendamento usando fuso `America/Sao_Paulo` disparando às 19:00 em dias úteis.
3. **Endpoint de Histórico:** Criar `GET /api/v1/ativos/{ticker}/historico?dias=30` retornando a série diária de Score, Preço e Ranking.
4. **Gráfico Temporal no Front-end:** Exibir mini-gráfico de linha interativo no `AssetDetailModal.tsx` mostrando a trajetória do Score ao longo dos pregões.
