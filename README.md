# 🏛️ B3 Core Platform & Motor de Valuation Inteligente

> Plataforma institucional de alta performance para análise fundamentalista multi-escola, mapa preditivo de dividendos (MDI), rebalanceamento automatizado de carteira Buy & Hold e análise gráfica com confluência de preço teto.

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?style=flat&logo=next.js)](https://nextjs.org)
[![Docker](https://img.shields.io/badge/Docker-Production-2496ED?style=flat&logo=docker)](https://docker.com)
[![SSL](https://img.shields.io/badge/SSL-A%2B%20RSA%20HTTPS-10b981?style=flat)](https://invest.caioscatolino.com.br)

---

## 🌐 Ambientes Oficiais em Produção

* **Web UI (Frontend Next.js):** [https://invest.caioscatolino.com.br/](https://invest.caioscatolino.com.br/)
* **API REST (Backend Go):** [https://api.invest.caioscatolino.com.br/](https://api.invest.caioscatolino.com.br/)
* **Health Check:** [https://api.invest.caioscatolino.com.br/health](https://api.invest.caioscatolino.com.br/health)

---

## 🏗️ Arquitetura do Sistema

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          APLICAÇÃO CLIENTE (WEB)                            │
│                 Next.js 16 (Standalone) • Tailwind CSS v4                  │
│             Design Preto Ônix & Dourado Champagne • Dark Mode               │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ HTTPS / REST (CORS Seguro)
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            BACKEND GO NATIVO                                │
│                   net/http • Concorrência via Goroutines                    │
│            Isolamento Multi-Tenant (Anti-IDOR) • JWT HMAC-SHA256            │
└───────────┬──────────────────────────┬──────────────────────────┬───────────┘
            │                          │                          │
            ▼                          ▼                          ▼
 ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
 │    MYSQL CENTRAL     │  │    REDIS CENTRAL     │  │   FONTES DE DADOS    │
 │ • Usuários & Bcrypt  │  │ • Snapshots Mercado  │  │ • B3 Oficial         │
 │ • Posições Custódia  │  │ • Cache Gráfico (4h) │  │ • CVM Demonstrativos │
 │ • Livro-Razão Trans. │  │ • Leitura em < 1ms   │  │ • StatusInvest Ingest│
 │ • Histórico Diário   │  │ • Fallback In-Memory │  │ • Yahoo Fin. Histórico│
 └──────────────────────┘  └──────────────────────┘  └──────────────────────┘
```

---

## 💎 Principais Funcionalidades

### 1. 📈 Valuation Multi-Escola & Score Composto (0 a 100)
* **Décio Bazin (6% e Histórico 5 Anos):** Preço teto por dividendo médio com auditoria oficial de proventos B3.
* **Benjamin Graham:** Valor Intrínseco por $VI = \sqrt{22.5 \times LPA \times VPA}$.
* **Peter Lynch:** PEG Ratio e Preço Justo baseado no crescimento de lucros de 5 anos.
* **Gordon DDM:** Custo de capital calibrado dinamicamente contra a Selic e NTN-B.
* **Joel Greenblatt:** Magic Formula combinando EV/EBIT e ROIC.
* **Joseph Piotroski:** F-Score de saúde contábil (0 a 9) com filtros anti-alavancagem.
* **Teto Consensual & Margem de Segurança:** Preço teto ponderado multi-modelo e percentual de desconto.

### 2. ⚖️ Gestão de Carteira & Rebalanceamento Buy & Hold
* **Macro Alocação Oficial em 5 Caixinhas:**
  * 40.0% Renda Fixa (Pós-Fixado, IPCA+, Prefixado)
  * 22.5% Fundos Imobiliários (FIIs)
  * 21.0% Ações Brasil
  * 16.0% Ativos Internacionais (ETFs/BDRs)
  * 0.5% Caixa Estratégico
* **Rebalanceamento Hierárquico sem Venda:** O robô direciona novos aportes exclusivamente para as classes e ativos mais defasados e com maior margem de segurança fundamentalista.
* **Recálculo Automático de Preço Médio (Média Ponderada):**
  $$PM_{\text{novo}} = \frac{(Q_{\text{antiga}} \times PM_{\text{antigo}}) + (Q_{\text{compra}} \times P_{\text{compra}})}{Q_{\text{antiga}} + Q_{\text{compra}}}$$
* **Extrato e Livro-Razão Contábil:** Rastreabilidade cronológica de todas as compras, vendas e aportes para fins fiscais e de IR.

### 3. 📅 Mapa de Dividendos Inteligente (MDI)
* Motor de sazonalidade estatística processando 5 anos de proventos.
* Modalidades analíticas por **Data-Com** e **Data de Pagamento**.
* Filtros de probabilidade mínima ($\ge 40\%, 60\%, 80\%, 100\%$) e visão cruzada com a custódia do investidor.

### 4. 📊 Análise Gráfica & Indicadores Técnicos Institucionais
* Engine gráfica proprietária com **Lightweight Charts (TradingView)**.
* **Preço Teto Plotado nas Velas:** Linha horizontal dourada mostrando graficamente se a cotação está na zona de compra ou no ágio.
* **Médias Móveis de Tendência:** MMA 8 (curto prazo), MMA 21 (médio prazo) e MMA 200 (tendência primária institucional).
* **Oscilador IFR / RSI (14 períodos):** Painel inferior sincronizado identificando sobrevenda ($\le 30$) e sobrecompra ($\ge 70$).
* **Sinal Técnico Confluente:** Alerta inteligente cruzando valuation com timing de entrada (ex: *ZONA COMPRA SOBREVENDIDO*).

---

## 🛠️ Executando Localmente

### Pré-requisitos
* Go 1.26+
* Node.js 20+ e npm
* MySQL e Redis (ou modo fallback In-Memory automático)

### 1. Backend em Go
```bash
cd apps/server
go mod download
go test ./...
go run cmd/api/main.go
```
*A API subirá em `http://localhost:8080`.*

### 2. Frontend Next.js
```bash
cd apps/web
npm install
npm run dev
```
*A interface web subirá em `http://localhost:3000`.*

---

## 🐳 Deploy Automatizado com Docker na VPS

A plataforma conta com pipeline de CI/CD via GitHub Actions (`.github/workflows/deploy.yml`) que dispara a cada push na branch `main`:

```bash
# Execução manual na VPS
./scripts/deploy.sh
```

---

## 📜 Licença e Propriedade
Desenvolvido por **Caio Scatolino**. Todos os direitos reservados.
