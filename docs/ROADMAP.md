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
- [x] **Passo 6.1**: Serviço de Ingestão Diária da B3 (`internal/b3`):
  - `cotahist.go`: Structs do arquivo posicional da B3 (`CODNEG`, `PREULT`, `VOLTOT`, `PREABE`, `PREMAX`, `PREMIN`).
  - `client.go`: Download HTTP com fallback retroativo de datas para feriados e fins de semana.
  - `parser.go`: Leitura posicional em stream com `bufio.Scanner` e filtro de mercado à vista (`CODBDI == "02"`, `"12"` e `"14"` para ETFs).
- [x] **Passo 6.2**: Ponto de entrada CLI manual (`cmd/cli/main.go --sync-b3`) para disparo sob demanda (587 ativos em 9.9ms).
- [x] **Passo 6.3**: Provedor de Mercado B3 (`internal/provider/b3_provider.go`):
  - Download único diário do COTAHIST (~450 KB), parsing em memória (< 15ms), filtro de corte por liquidez financeira (ex: > R$ 500k) e cruzamento com catálogo MySQL.
- [x] **Passo 6.4**: Integração no Worker Batch (`internal/worker/sync_worker.go`):
  - União da cotação oficial B3 com os balanços contábeis da CVM (LPA, VPA, VP/Cota) sem nenhuma requisição ao Yahoo Finance na Frente 2.
- [x] **Passo 6.5**: Motores Avançados de Valuation, Score Fundamentalista e Pareceres (Concluído - Passo 15):
  - Modelo Gordon DDM (`internal/gordon`), Joel Greenblatt Earnings Yield e Peter Lynch PEG Ratio.
  - Score Fundamentalista Composto (0 a 100) segmentado por classe de ativo.
  - Pareceres individuais pedagógicos por escola (`APROVADO`, `ATENCAO`, `REPROVADO`).
  - Separação e ordenação de 3 rankings completos (FIIs, Ações e ETFs) com 388 ativos auditados em 14 segundos.
- [x] **Passo 6.6**: Adaptador Redis (`internal/storage/store.go`):
  - [x] Snapshot distribuído de alta performance com Redis Pipeline (< 5ms para ~400 ativos).
  - [x] Estratégia de chaves individuais (`b3:snapshot:<ticker>`) e agregada (`b3:snapshot:todos`) com TTL dinâmico de 25 min.
  - [x] Loop de atualização em background via StatusInvest a cada 20 minutos.
  - [x] Fallback automático para `InMemoryStore` caso o Redis esteja indisponível.
- [x] **Passo 6.7**: Histórico Diário de Rankings Pós-Pregão (Cron 19h) & Fator Temporal no Valuation:
  - [x] **Cron Automatizado 19h:** Rotina agendada diária executada às 19:00 em Go nativo (segunda a sexta-feira, pós-fechamento do pregão da B3 e homologação dos boletins diários).
  - [x] **Tabela MySQL `historico_rankings_diarios`:** Armazena o snapshot imutável de cada pregão (`data_pregao`, `ticker`, `classe`, `posicao_ranking`, `score`, `preco_atual`, `preco_teto_consolidado`, `premio_desconto_percentual`, `dy`, `pl`, `pvp`, `status`, `is_provento_atipico`, `modelos_json`).
  - [x] **Persistência Atômica & Idempotente:** Transação única em lote (`tx.BeginTx`) com `ON DUPLICATE KEY UPDATE` gravando centenas de ativos em < 50ms sem duplicidade.
  - [ ] **Motor de Valuation Temporal (Consistência & Momentum Fundamentalista):**
    - Analisa os rankings e scores dos pregões anteriores (7, 30 e 90 dias) como peso analítico.
    - *Fator de Consistência:* Bonifica ativos resilientes no topo e penaliza oscilações efêmeras.
    - *Detecção de Turnaround / Upgrades:* Identifica ativos em franca ascensão no ranking.
    - *Divergência Preço x Score:* Detecta oportunidades quando o preço cai mas o score sobe ou se mantém firme.
  - [ ] **Visualização Gráfica no Front-end:** Gráfico de evolução histórica de Score e Preço Teto nos modais de análise (`GET /api/v1/ativos/{ticker}/historico`).

---

### ⏳ FASE 7: Autenticação Segura (Cybersegurança OWASP) & Gestão de Carteira do Usuário
- [x] **Passo 7.1: Arquitetura de Cibersegurança & Autenticação Backend (`internal/auth`)**:
  - [x] **Tabela MySQL `usuarios`:** Migração `003_create_usuarios_table.sql` com UUID v4 público, e-mail único, hash bcrypt, controle de tentativas falhas e bloqueio temporário.
  - [x] **Hashing Criptográfico Resistente:** Hashing com **bcrypt (custo 12)** com salt dinâmico e comparações em tempo constante contra *timing attacks*.
  - [x] **Regras de Senha Forte (OWASP):** Validador estrito exigindo no mínimo 8 caracteres com maiúscula, minúscula, número e símbolo.
  - [x] **Autenticação JWT Stateless:** Tokens assinados com HMAC-SHA256, expiração configurável e injeção de identidade no `context.Context` da requisição.
  - [x] **Proteção Anti-Brute Force & Anti-Enumeração:** Bloqueio automático de 15 minutos após 5 erros consecutivos e mensagens neutras de erro.
  - [x] **Endpoints REST:** `POST /api/v1/auth/register`, `POST /api/v1/auth/login` e rota protegida `GET /api/v1/auth/me`.
- [x] **Passo 7.2: Interface de Autenticação no Front-end Next.js (`apps/web`)**:
  - [x] **Modal de Autenticação Dual (Entrar / Criar Conta):** Design terminal/institucional em preto ônix e dourado champagne, alternância suave entre abas.
  - [x] **Medidor Visual de Força de Senha (UX Pedagógica):** Checklist dinâmico OWASP com 4 critérios e barra de progresso colorida em tempo real.
  - [x] **Gerenciador de Estado Global (`AuthContext` / `useAuth`):** Armazenamento de JWT, auto-hidratação via `/api/v1/auth/me`, ações `login`, `register`, `logout`.
  - [x] **Header do Dashboard Integrado:** Alternância dinâmica entre botão "Entrar / Cadastrar" e badge com iniciais do usuário autenticado + "Sair".
  - [x] **Tratamento Visual de Erros:** Alertas em vermelho carmim para credenciais inválidas e bloqueio temporário anti-brute force.
- [x] **Passo 7.3: Gestão de Carteira Multi-Classe (`internal/portfolio`)**:
  - [x] **Classes Suportadas:** `RENDA_FIXA`, `FII`, `ACAO_BR`, `INTERNACIONAL`, `CAIXA`.
  - [x] **Tabela MySQL `carteira_posicoes`:** Migração `004_create_carteira_tables.sql` com controle multi-tenant, quantidade ponderada, corretoras e indexadores de renda fixa.
  - [x] **Tabela MySQL `carteira_transacoes`:** Livro-razão contábil para histórico de operações (compras, vendas, aportes) e rastreabilidade contábil.
  - [x] **Consolidação em Tempo Real com Snapshot B3:** Enriquecimento automático com cotação atual, Preço Teto consensual, Score Fundamentalista e Margem de Segurança.
  - [x] **Isolamento Estrito Multi-Tenant (Anti-IDOR):** Todas as operações auditadas e filtradas por `WHERE usuario_id = ?` extraído do JWT.
  - [x] **Endpoints REST:** `GET /api/v1/carteira`, `POST /api/v1/carteira/posicoes`, `DELETE /api/v1/carteira/posicoes/{id}`, `GET /api/v1/carteira/transacoes`.
- [x] **Passo 7.4: Motor de Rebalanceamento Inteligente & Auditoria da Carteira (`internal/portfolio/rebalance.go`)**:
  - [x] **Regra Institucional de Macro Alocação:** Metas oficiais (40% RF, 22.5% FIIs, 21% Ações BR, 16% Int, 0.5% Caixa).
  - [x] **Algoritmo de Rebalanceamento sem Venda (Aporte Direcionado):** Filosofia Buy & Hold em 2 níveis (Nível 1: Macro Alocação; Nível 2: Micro Alocação Fundamentalista por Score e Margem de Segurança).
  - [x] **Simulador de Aportes Financeiros (`POST /api/v1/carteira/rebalancear`):** Distribuição matemática ótima com suporte a cotas inteiras para ações/FIIs e aportes fracionários em Renda Fixa.
  - [x] **Diagnóstico Macro com Semáforo de Desvio:** Identificação em tempo real de classes `DEFASADO_APORTAR`, `EQUILIBRADO` e `SOBREALOCADO`.
- [x] **Passo 7.5: Endpoints da Carteira REST & Interface de Gestão no Frontend Next.js**:
  - [x] **Endpoints REST:** `GET /api/v1/carteira`, `POST /api/v1/carteira/posicoes`, `DELETE /api/v1/carteira/posicoes/{id}`, `POST /api/v1/carteira/rebalancear`.
  - [x] **Dashboard de Custódia (`PortfolioSummaryCards`):** Cards de Patrimônio Total consolidado a mercado, Custo Histórico, Lucro/Prejuízo nominal e rentabilidade percentual.
  - [x] **Módulo de Macro Alocação (`MacroAllocationView`):** Comparativo visual da regra institucional 40% RF | 22.5% FIIs | 21% Ações BR | 16% Int | 0.5% Caixa com barras de progresso e semáforo de desvio ($\Delta$).
  - [x] **Tabela Interativa de Posições (`PortfolioTable`):** Filtros por classe, cotação oficial em tempo real da B3, preço teto consensual, margem de segurança, score e botão de exclusão.
  - [x] **Modal de Cadastro Dinâmico (`AddPositionModal`):** Suporte completo para Ações BR, FIIs e Renda Fixa (taxa contratada, indexador, vencimento).
  - [x] **Simulador Interativo de Rebalanceamento (`RebalanceSimulatorModal`):** Cálculo automatizado de aportes direcionados Buy & Hold em cotas e frações sem vendas forçadas.

---

### ⏳ FASE 8: Servidor HTTP Nativo & Streaming com Server-Sent Events (SSE)
- [x] **Passo 8.1**: Servidor HTTP em Go (`net/http`) com `http.NewServeMux()` (Concluído - Passo 16):
  - `GET /health` (Health check com total de ativos carregados).
  - `GET /api/v1/rankings` (Rankings completos de Ações, FIIs e ETFs ordenados por Score).
  - `GET /api/v1/ativos/{ticker}` (Análise profunda e pareceres individuais).
  - Middlewares de CORS nativo e Logging de performance em microssegundos.
- [ ] **Passo 8.2**: Endpoint SSE (`GET /api/audit/stream`):
  - Streaming em tempo real para o frontend com efeito máquina de escrever.

---

### ⏳ FASE 9: Agente de IA com Gemini Pro & Skills Analíticas (Funil Etapa 2)
- [ ] **Passo 9.1**: Integração com a API do Gemini Pro via Go SDK nativo.
- [ ] **Passo 9.2**: Prompts das skills modulares (`skills/bazin-method`, `skills/graham-method`, `skills/consensus-filter`).
- [ ] **Passo 9.3**: Funil cognitivo: O Go envia os 15 melhores finalistas pré-filtrados e a IA redige o parecer fundamentalista detalhado do Top 10 via streaming SSE.

---

### ⏳ FASE 10: Front-end Next.js (Mobile-First & Desktop)
- [x] **Passo 10.1**: Inicialização do Next.js standalone em `apps/web/` com Tailwind CSS v4 e TypeScript (Concluído - Passo 17).
- [x] **Passo 10.2**: Consumo dos endpoints REST (`/api/v1/rankings` e `/api/v1/ativos/{ticker}`) em tempo real.
- [x] **Passo 10.3**: Interface responsiva institucional em preto com dourado, 3 tabelas independentes (Ações, FIIs, ETFs) com paginação, busca e modais detalhados de valuation.
- [ ] **Passo 10.4**: Visualização temporal da trajetória de rankings e scores dos últimos 30 dias.

---

### ✅ FASE 11: Infraestrutura, Dockerização e Deploy na VPS (`179.236.230.139`)
- [x] **Passo 11.1: Dockerização Multi-Stage de Alta Performance**:
  - [x] **Back-end em Go (`apps/server/Dockerfile`):**
    - Multi-stage build com `golang:alpine` (`GOTOOLCHAIN=auto`), compilação estática (`CGO_ENABLED=0 GOOS=linux`) e imagem final `alpine:3.20` ultra-enxuta (< 25 MB).
    - Inclusão de `ca-certificates` e `tzdata` (`America/Sao_Paulo`).
    - Execução com usuário `appuser` não-privilegiado por segurança de container.
  - [x] **Front-end Next.js (`apps/web/Dockerfile`):**
    - Build otimizado standalone (`output: 'standalone'`) reduzindo o container para < 120 MB.
- [x] **Passo 11.2: Orquestração com Docker Compose (`docker-compose.prod.yml`)**:
  - [x] Serviços `b3_invest_api` e `b3_invest_web` orquestrados com restart automático e healthcheck calibrado (90s start period para ingestão dos 1092 ativos da B3).
  - [x] Integração perfeita com a infraestrutura compartilhada na rede `rede-compartilhada` consumindo `mysql_central` e `redis_central`.
- [x] **Passo 11.3: Configuração do Domínio, Nginx Reverse Proxy e SSL/HTTPS**:
  - [x] Roteamento de produção em **`https://invest.caioscatolino.com.br/`** com SSL Let's Encrypt ativo.
  - [x] Backend API isolado e roteado via Nginx Proxy Manager.
- [x] **Passo 11.4: Script de Deploy e Pipeline CI/CD Automatizado**:
  - [x] Script `scripts/deploy.sh` com automação de migrations MySQL (`001` a `004`), auto-detecção de rede e subida limpa dos containers.
  - [x] Pipeline GitHub Actions `.github/workflows/deploy.yml` com disparo automático a cada push na branch `main`.

---

### 🚀 FASE 12: Modelo Multi-Fator Institucional de Ações & Desacoplamento Macroeconômico NTN-B
- [x] **Passo 12.1: Diagnóstico e Desacoplamento Macroeconômico de Risco em Ações**:
  - [x] **Eliminação da Penalidade Rígida de Dividend Yield:** Ações não são títulos de renda fixa; cessou a punição de -8 pontos para ações com DY < 6.5% quando reinvestem capital de forma lucrativa (ex: SLCE3, WEGE3, PRIO3).
  - [x] **Introdução do Fed Model (Earnings Yield vs NTN-B):** Avaliação do poder de lucro operacional real ($1/\text{P/L}$) contra a taxa livre de risco da NTN-B, reconhecendo empresas de alto retorno operacional mesmo em fundos de ciclo de commodities/agro.
- [x] **Passo 12.2: Implementação do Score Fundamentalista de 4 Pilares (0 a 100)**:
  - [x] **Pilar 1: Qualidade & Eficiência Operacional (até 28 pts):** ROIC/ROE (até 14 pts), Margens Operacionais/Líquidas (até 7 pts), Crescimento de Lucros 5A (até 7 pts).
  - [x] **Pilar 2: Valuation & Margem de Segurança (até 30 pts):** P/L e Earnings Yield (até 10 pts), P/VP e Desconto Patrimonial Real (até 8 pts), Margem de Segurança Graham/Greenblatt (até 12 pts).
  - [x] **Pilar 3: Saúde Financeira & Solvência (até 22 pts):** Piotroski F-Score (até 12 pts), Alavancagem Dívida Líquida/EBITDA e Porte (até 10 pts).
  - [x] **Pilar 4: Retorno Total ao Acionista (até 20 pts):** Dividend Yield Normalizado (máximo entre 12M e média de 5 anos) (até 12 pts), Alinhamento de Preço Teto Consensual (até 8 pts).
  - [x] **Bônus Macroeconômico & Travas Prudenciais:** Bônus de +3 pts para Earnings Yield elevado; travas anti-Value Trap mantidas para prejuízo (LPA <= 0), PL negativo e microcaps ilíquidas.
- [x] **Passo 12.3: Atualização da Interface do Usuário (Frontend Next.js)**:
  - [x] **Substituição na Tabela Principal de Ações:** Remoção da coluna de ruído "Spread NTN-B" e rótulo indevido de "Descompensado", substituindo pela métrica rainha de rentabilidade: **ROIC**.
  - [x] **Filtros de Ordenação Dedicados:** Botão "💎 ROIC / Rentab." adicionado para ações; filtro "⚖️ Risco vs NTN-B" mantido exclusivamente para FIIs onde a métrica é mandatória.
  - [x] **Preservação de Contexto:** Análise de Spread NTN-B mantida exclusivamente no modal analítico de detalhes e na tabela de Fundos Imobiliários.
- [ ] **Passo 12.4: Integração de Scraper Automático da NTN-B (Tesouro Direto / Anbima)**:
  - [ ] Sincronização em tempo real do juro real das NTN-Bs longas (Tesouro IPCA+ 2035/2045) via API oficial ou scraper diário, complementando a Selic do BCB SGS.

