# 🗺️ Roadmap do Projeto: Plataforma B3 Core & Motor Inteligente

Este documento é a bússola viva do projeto. Ele mapeia as decisões arquiteturais, o progresso atual e os próximos passos para guiar o desenvolvimento incremental, seja nesta sessão ou em sessões futuras.

---

## 🏗️ Arquitetura das Duas Frentes Operacionais

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       FRENTE 2: SCANNER GERAL (BATCH)                       │
│                  (Worker concorrente em segundo plano via Go)               │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                ┌──────────────────────┴──────────────────────┐
                ▼                                             ▼
        FONTE 1: CVM                                   FONTE 2: B3 OFICIAL
    (Balanços & Informes)                            (Boletim Diário COTAHIST)
    • DFP/ITR: Lucro Líquido & PL                 • Cotações de Fechamento (PREULT)
    • Composição Acionária (Ações)                • Volume Financeiro Real (VOLTOT)
    • INF_MENSAL: VP/Cota FIIs                    • 1 único download diário (~450 KB)
                │                                 • Zero risco de rate limit / ban
                │                                             │
                └──────────────────────┬──────────────────────┘
                                       │
                                       ▼
                          Cálculo Dinâmico (Sem mocks):
                          • LPA = Lucro Líquido / Ações
                          • VPA = Patrimônio Líquido / Ações
                          • Motores: Bazin, Graham, Lynch, FII
                                       │
                                       ▼
                       SNAPSHOT CONSOLIDADO NO REDIS
                       (Chave: "b3:snapshot:<ticker>")
                                       │
═══════════════════════════════════════╪═══════════════════════════════════════
                                       │ (Leitura em < 1ms)
┌──────────────────────────────────────┴──────────────────────────────────────┐
│                  FRENTE 1: CARTEIRA DO USUÁRIO (ON-DEMAND)                  │
│                     (API REST & Server-Sent Events - SSE)                   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼
                       YAHOO FINANCE (Uso Restrito/Pontual)
                • Apenas cotações intradiárias ao vivo sob demanda
                • Poucos ativos da carteira do usuário (< 20 tickers)
                • Previne bloqueio de IP por excesso de chamadas
```

---

## 📋 Status das Fases de Desenvolvimento

### ✅ FASE 1: Fundamentos da Linguagem Go (Concluída)
- [x] Instalação do compilador Go no Windows (`go1.27.0 windows/amd64`).
- [x] Inicialização do módulo oficial (`apps/server/go.mod`).
- [x] Estrutura Standard Go Layout (`cmd/api/main.go`, `internal/domain`, etc.).
- [x] Tipagem estática forte, Structs e Tags JSON (`json:"ticker"`).
- [x] Múltiplos retornos e tratamento explícito de erros (`if err != nil`).
- [x] Ponteiros na prática (`&` address-of e `*` pointer type) evitando cópias descartáveis de memória.

### ✅ FASE 2: Motores de Valuation e Polimorfismo (Concluída)
- [x] Método Décio Bazin (`internal/bazin`): Preço Teto por sustentabilidade de proventos.
- [x] Método Benjamin Graham (`internal/graham`): Valor Intrínseco por LPA e VPA com `math.Sqrt`.
- [x] Método Peter Lynch (`internal/lynch`): PEG Ratio e Preço Justo por crescimento.
- [x] Método FIIs (`internal/fii`): P/VP patrimonial e Spread vs Tesouro IPCA+ (NTN-B).
- [x] Polimorfismo via Interfaces em Go (`domain.Analisador`).
- [x] Semáforo de recomendação unificado (`COMPRAR_MAIS`, `MANTER`, `ALERTA`).

### ✅ FASE 3: Concorrência e Ingestão de Cotações (Concluída)
- [x] Provedor de cotações em tempo real (`internal/provider/yahoo.go`).
- [x] Goroutines concorrentes, `sync.WaitGroup` e Channels (`chan *domain.Ativo`).
- [x] Tratamento de Units da B3 (`TAEE11`, etc.) diferenciando Ações de FIIs.
- [x] Teste real do download oficial da CVM em stream (`internal/cvm/fii_reader.go`).

---

### ✅ FASE 4: Catálogo no MySQL e Mapeamento Ticker <-> CNPJ (Concluída)
- [x] **Passo 4.1**: Modelar a tabela `assets` no MySQL com tipos e índices.
- [x] **Passo 4.2**: Configurar conexão `database/sql` nativa em Go com pool de conexões e `.env`.
- [x] **Passo 4.3**: Criar o repositório `internal/catalog/repository.go` com métodos de busca e inserção em lote.
- [x] **Passo 4.4**: Sincronizador dinâmico de catálogo B3 + CVM: **2.000 ativos oficiais da B3 cadastrados no MySQL local sem dados fixos no código**.

---

### ✅ FASE 5: Pipeline CVM Dinâmico para Ações (Eliminação Total de Mocks - Concluída)
- [x] **Passo 5.1**: Parser dos demonstrativos ITR da CVM em stream (`internal/cvm/dre_reader.go`):
  - Extrair Lucro Líquido (DRE Con) e Patrimônio Líquido (BPP Con) do ZIP oficial da CVM com `LazyQuotes: true` e suporte multi-ano.
- [x] **Passo 5.2**: Extrair composição acionária oficial (`composicao_capital`) e derivar:
  - $LPA = \frac{\text{Lucro Líquido}}{\text{Total de Ações}}$
  - $VPA = \frac{\text{Patrimônio Líquido}}{\text{Total de Ações}}$
- [x] **Passo 5.3**: Remover definitivamente o `switch` estático de LPA/VPA do `sync_worker.go` e CNPJs oficiais em `fii_reader.go` (100% dos FIIs com VP/Cota e Ações com LPA/VPA dinâmicos).

---

### ⏳ FASE 6: Frente 2 - Ingestão B3 (COTAHIST) & Scanner Batch em Redis
- [x] **Passo 6.1**: Serviço de Ingestão Diária da B3 (`internal/b3`) (Concluído):
  - `cotahist.go`: Structs do arquivo posicional da B3 (`CODNEG`, `PREULT`, `VOLTOT`, `PREABE`, `PREMAX`, `PREMIN`).
  - `client.go`: Download HTTP com fallback retroativo de datas para feriados e fins de semana.
  - `parser.go`: Leitura posicional em stream com `bufio.Scanner` e filtro de mercado à vista (`CODBDI == "02"` e `"12"`).
- [x] **Passo 6.2**: Ponto de entrada CLI manual (`cmd/cli/main.go --sync-b3`) para disparo sob demanda (Concluído: 587 ativos em 9.9ms).
- [x] **Passo 6.3**: Provedor de Mercado B3 (`internal/provider/b3_provider.go`) (Concluído):
  - Download único diário do COTAHIST (~450 KB), parsing em memória (< 15ms), filtro de corte por liquidez financeira (ex: > R$ 500k) e cruzamento com catálogo MySQL.
- [x] **Passo 6.4**: Integração no Worker Batch (`internal/worker/sync_worker.go`) (Concluído):
  - União da cotação oficial B3 com os balanços contábeis da CVM (LPA, VPA, VP/Cota) sem nenhuma requisição ao Yahoo Finance na Frente 2.
- [ ] **Passo 6.5**: Motor de Proventos Dinâmicos ($\Sigma_{12M}$ TTM) & Resolução de CNPJs Oficiais:
  - FIIs: Leitura dos rendimentos mensais dos últimos 12 meses direto do informe CVM (`inf_mensal_fii_complemento`).
  - Ações: Processamento de eventos corporativos em dinheiro (Dividendos e JCP com dedução de 15% de IR) na janela móvel de 365 dias (`Data >= Hoje - 365d`).
  - Cálculo oficial do Preço Teto de Bazin ($\frac{\text{Proventos Líquidos 12M}}{0.06}$) e semáforo dinâmico.
- [ ] **Passo 6.6**: Adaptador Redis (`internal/storage/redis_store.go`) gravando Snapshot Consolidado ("b3:snapshot:<ticker>") com TTL.

---

### ⏳ FASE 7: Frente 1 - Auditoria e Carteira do Usuário (On-Demand)
- [ ] **Passo 7.1**: Estrutura `internal/portfolio`:
  - `Posicao`: Ticker, Quantidade, Preço Médio, Custo Total, Valor Atual, Lucro/Prejuízo, Yield on Cost (YoC).
- [ ] **Passo 7.2**: Serviço de Auditoria Rápida:
  - Cruza as posições do usuário com o Snapshot do Redis (< 10ms).
  - Emite o parecer consolidado e os aportes recomendados por classe.

---

### ⏳ FASE 8: Servidor HTTP Nativo & Streaming com Server-Sent Events (SSE)
- [ ] **Passo 8.1**: Servidor HTTP em Go (`net/http`) com `http.NewServeMux()`:
  - `GET /health` (Health check para Docker/Nginx Proxy Manager).
  - `GET /api/market/snapshot` (JSON de mercado).
  - `POST /api/portfolio/audit` (Auditoria rápida da carteira).
- [ ] **Passo 8.2**: Endpoint SSE (`GET /api/audit/stream`):
  - Streaming em tempo real para o frontend com efeito máquina de escrever.

---

### ⏳ FASE 9: Agente de IA com Gemini Pro & Skills Analíticas (Funil Etapa 2)
- [ ] **Passo 9.1**: Integração com a API do Gemini Pro via Go SDK nativo.
- [ ] **Passo 9.2**: Prompts das skills modulares (`skills/bazin-method`, `skills/graham-method`, `skills/consensus-filter`).
- [ ] **Passo 9.3**: Funil cognitivo: O Go envia os 15 melhores finalistas pré-filtrados e a IA redige o parecer fundamentalista detalhado do Top 10 via streaming SSE.

---

### ⏳ FASE 10: Front-end Next.js (Mobile-First & Desktop)
- [ ] **Passo 10.1**: Inicialização do Next.js standalone em `apps/web/`.
- [ ] **Passo 10.2**: Consumo do fluxo SSE em tempo real.
- [ ] **Passo 10.3**: Interface responsiva de auditoria com semáforo visual e simulador de aportes.

---

### ⏳ FASE 11: Infraestrutura e Deploy na VPS (`179.236.230.139`)
- [ ] **Passo 11.1**: `Dockerfile` multi-stage build do back-end em Go (< 25 MB).
- [ ] **Passo 11.2**: `docker-compose.yml` conectado à rede existente da VPS (`mysql_central`, `redis_central`, `nginx_proxy_manager`).
- [ ] **Passo 11.3**: Script de deploy automatizado via SSH (`root@179.236.230.139`).
