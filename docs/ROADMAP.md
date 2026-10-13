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
- [ ] **Passo 6.6**: Adaptador Redis (`internal/storage/redis_store.go`) gravando Snapshot Consolidado ("b3:snapshot:<ticker>") com TTL.
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
- [ ] **Passo 7.1: Arquitetura de Cibersegurança & Autenticação (`internal/auth`)**:
  - **Tabela MySQL `usuarios`:** `id`, `uuid`, `nome`, `email`, `senha_hash`, `status`, `tentativas_falhas`, `bloqueado_ate`, `created_at`, `updated_at`.
  - **Hashing Criptográfico Resistente:** Utilização de **Argon2id** (padrão OWASP) ou **bcrypt** (fator de custo 12+) com salt único por usuário contra ataques de rainbow tables e brute-force via GPU.
  - **Autenticação JWT Stateless & Chaves Seguras:** Emissão de Access Tokens de curta duração (15 min) assinados criptograficamente, contendo apenas identificador público e perfil.
  - **Refresh Tokens Rotativos:** Armazenados com hash no banco, com revogação sob demanda e tráfego exclusivo via cookies `HttpOnly`, `Secure` e `SameSite=Strict`.
  - **Proteção Anti-Brute Force & Rate Limiting:** Middleware limitador de requisições por IP e por conta de usuário no endpoint de login (bloqueio temporário após 5 tentativas inválidas consecutivas).
  - **Prevenção OWASP Top 10:** Prepared Statements estritos contra SQL Injection, validação e sanitização de inputs, proteção contra IDOR e injeção de headers de segurança HTTP (HSTS, CSP, X-Frame-Options: DENY, X-Content-Type-Options: nosniff).
- [ ] **Passo 7.2: Gestão de Carteira Multi-Classe (`internal/portfolio`)**:
  - **Classes Suportadas:**
    - `RENDA_FIXA`: Títulos do Tesouro Direto (Selic/IPCA+), CDBs, LCIs, LCAs e Debêntures (com indexador, taxa e vencimento).
    - `FII`: Fundos Imobiliários auditados na B3.
    - `ACAO_BR`: Ações de companhias abertas brasileiras cotadas na B3.
    - `INTERNACIONAL`: BDRs, ETFs globais (ex: IVVB11, WRLD11) e Ações Internacionais (Stocks).
    - `CAIXA`: Saldo livre e reserva de oportunidade líquida.
  - **Tabela MySQL `carteira_posicoes`:** `id`, `usuario_id`, `ticker`, `classe`, `categoria_macro`, `quantidade`, `preco_medio`, `valor_aplicado`, `taxa_contratada`, `vencimento`, `corretora`, `observacoes`, `created_at`, `updated_at`.
  - **Tabela MySQL `carteira_transacoes`:** Histórico contábil de movimentações (compras e vendas) para cálculo automatizado de Preço Médio ponderado e apuração fiscal básica.
  - **Validação de Ativos B3:** Validação rigorosa dos tickers contra a tabela de catálogo oficial (`assets`), impedindo códigos inexistentes ou injeções de dados arbitrários.
  - **Isolamento Estrito Multi-Tenant (Anti-IDOR):** Todo acesso e mutação à carteira é obrigatoriamente filtrado por `WHERE usuario_id = ?`, extraído do contexto seguro do token JWT validado pelo middleware.
- [ ] **Passo 7.3: Motor de Rebalanceamento Inteligente & Auditoria da Carteira (`internal/portfolio/rebalance.go`)**:
  - **Regra Institucional de Macro Alocação Patrimonial:**
    - 🏛️ **40,0% Renda Fixa** (Segurança, estabilidade e carrego de juros).
    - 🏢 **22,5% Fundos Imobiliários** (Geração mensal de proventos isentos).
    - 📈 **21,0% Ações Brasileiras** (Crescimento de dividendos e valorização assimétrica).
    - 🌐 **16,0% Ações Internacionais** (Dolarização e proteção geográfica/geopolítica).
    - 💵 **0,5% Caixa / Reserva de Oportunidade** (Liquidez para compras táticas).
  - **Algoritmo de Rebalanceamento sem Venda (Aporte Direcionado):**
    - Calcula o patrimônio total e o desvio percentual ($\Delta = \%_{atual} - \%_{alvo}$) de cada macro classe.
    - **Filosofia Buy & Hold:** Não força vendas nem giros tributáveis desnecessários; canaliza 100% dos novos aportes e dividendos para a classe mais defasada em relação à meta.
    - **Rebalanceamento Hierárquico em 2 Níveis:**
      1. *Nível 1 (Macro Alocação):* Identifica qual caixinha patrimonial está mais atrasada (ex: se FIIs caíram para 19%, o aporte deve ir para FIIs).
      2. *Nível 2 (Micro Alocação Fundamentalista):* Dentro da classe escolhida, cruza com o **Score Fundamentalista e o Upside sobre o Preço Teto Consolidado** para escolher o melhor ativo, impedindo que o usuário aporte em ativos caros/reprovados.
  - **Simulador de Aportes:**
    - Usuário informa o valor a aportar no mês (ex: R$ 2.000,00) $\rightarrow$ O sistema calcula a alocação exata em cotas/ações para restabelecer o equilíbrio da carteira.
  - **Indicadores Individuais da Carteira:**
    - *Rentabilidade Real:* Preço Médio vs Preço Atual de Fechamento.
    - *Upside Potencial:* Margem de segurança sobre o Preço Teto Consolidado dos seus ativos.
    - *Yield on Cost (YoC):* Proventos reais recebidos em relação ao preço médio desembolsado.
- [ ] **Passo 7.4: Endpoints da API REST & Interface no Frontend Next.js**:
  - `POST /api/v1/auth/register` (Cadastro com regras de complexidade de senha).
  - `POST /api/v1/auth/login` (Autenticação e geração de sessão segura).
  - `POST /api/v1/auth/refresh` & `POST /api/v1/auth/logout`.
  - `GET /api/v1/carteira` (Listagem consolidada das posições do usuário autenticado).
  - `POST /api/v1/carteira/posicoes` (Cadastro/adição de ticker com quantidade e preço médio).
  - `PUT /api/v1/carteira/posicoes/{id}` & `DELETE /api/v1/carteira/posicoes/{id}`.
  - `POST /api/v1/carteira/rebalancear` (Simulador de aporte com cálculo automático de distribuição).
  - `GET /api/v1/carteira/auditoria` (Diagnóstico completo, desvios percentuais e semáforo da carteira).
  - **Front-end:** Telas de Login, Cadastro, Dashboard de Carteira com Gráfico Donut de Macro Alocação (Atual vs Alvo: 40% RF | 22.5% FIIs | 21% Ações BR | 16% Int | 0.5% Caixa), e Módulo Interativo de Simulação de Aporte Inteligente.

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

### ⏳ FASE 11: Infraestrutura, Dockerização e Deploy na VPS (`179.236.230.139`)
- [ ] **Passo 11.1: Dockerização Multi-Stage de Alta Performance**:
  - **Back-end em Go (`apps/server/Dockerfile`):**
    - Multi-stage build com `golang:alpine` para compilação estática (`CGO_ENABLED=0 GOOS=linux`) e imagem final `alpine` ou `scratch` ultra-enxuta (< 25 MB).
    - Inclusão de `ca-certificates` (para requisições HTTPS à CVM/B3) e `tzdata` (para o Cron 19h respeitar o timezone `America/Sao_Paulo`).
    - Execução com usuário não-privilegiado (*non-root*) por segurança de container.
  - **Front-end Next.js (`apps/web/Dockerfile`):**
    - Build otimizado com modo standalone (`output: 'standalone'` no Next.js) reduzindo o container de 1 GB para < 120 MB.
- [ ] **Passo 11.2: Orquestração com Docker Compose (`docker-compose.prod.yml`)**:
  - Serviços orquestrados:
    - `b3-api`: Servidor HTTP Go com healthcheck automático (`/health`).
    - `b3-web`: Servidor Next.js conectado internamente à API.
  - Integração com a infraestrutura existente na VPS:
    - Conexão com os containers existentes via rede Docker compartilhada (`mysql_central`, `redis_central`).
  - Configurações de resiliência: `restart: unless-stopped`, limites de recursos (RAM/CPU) e volumes para logs persistentes.
- [ ] **Passo 11.3: Configuração do Domínio, Nginx Reverse Proxy e SSL/HTTPS**:
  - **Roteamento por Domínio Próprio:**
    - `invest.seudominio.com.br` $\rightarrow$ Front-end Next.js (`b3-web:3000`).
    - `api.invest.seudominio.com.br` (ou `/api/`) $\rightarrow$ Back-end Go (`b3-api:8080`).
  - **Certificado SSL/TLS Gratuito:** Let's Encrypt com renovação automática via Certbot ou interface do Nginx Proxy Manager.
  - **Configurações Especiais no Nginx:**
    - Streaming de Server-Sent Events (SSE): `proxy_buffering off;` e `proxy_read_timeout 3600s;` para a transmissão em tempo real do assistente/IA sem cortes.
    - Headers de Cibersegurança: HSTS (`Strict-Transport-Security`), X-Frame-Options (`DENY`), X-Content-Type-Options (`nosniff`).
- [ ] **Passo 11.4: Script de Deploy Automatizado (Zero Downtime)**:
  - Script bash `scripts/deploy.sh` executável via SSH:
    - Execução automática de novas migrações SQL no MySQL da VPS (`scripts/migrations/*.sql`).
    - Build e subida limpa dos containers com `docker compose -f docker-compose.prod.yml up -d --build`.
    - Health check pós-deploy confirmando que a API e o Web estão respondendo 200 OK.

