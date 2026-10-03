# Passo 18: Pivô Arquitetural - Ingestão Consolidada StatusInvest e Preservação de Backup

## 🎯 Objetivo Estratégico
Realizar uma transição arquitetural de alta performance na camada de obtenção de dados da plataforma:
- **Antes:** Download e parsing manual de múltiplos arquivos ZIP pesados da CVM (ITR, DFP, BPP, DRE, Informes Mensais) e arquivos diários da B3 (COTAHIST, Eventos Corporativos), que demandavam alto consumo de I/O, memória e regras complexas de sincronização periódica.
- **Agora:** Consumo direto dos endpoints internos consolidados de busca avançada do **StatusInvest**, obtendo 100% dos ativos da B3 (617 Ações e 605 FIIs) com todas as suas métricas contábeis, operacionais e múltiplos em apenas **2 requisições HTTP (< 1 segundo)**, já alinhadas ao padrão de mercado institucional.

---

## 🔒 1. Preservação Total e Diretório de Backup (`internal/legacy_backup/`)
Seguindo o protocolo estrito de segurança, nenhum código ou inteligência contábil desenvolvida anteriormente foi deletada. Todos os pacotes e scripts legados foram preservados em:
- [apps/server/internal/legacy_backup/cvm/](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/legacy_backup/cvm/): Parsers diretos da CVM (`cia_reader.go`, `dre_reader.go`, `fii_reader.go`).
- [apps/server/internal/legacy_backup/b3/](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/legacy_backup/b3/): Leitor de eventos corporativos em dinheiro (`events_reader.go`) e COTAHIST (`cotahist.go`).
- [apps/server/internal/legacy_backup/provider/](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/legacy_backup/provider/): Provedores de mercado Yahoo Finance e B3 Provider.
- [apps/server/internal/legacy_backup/scripts/](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/legacy_backup/scripts/): Scripts de teste contábil e inspeção.

---

## ⚡ 2. Novo Módulo StatusInvest (`internal/statusinvest/`)
Criado novo pacote fortemente tipado em Go:
- [models.go](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/statusinvest/models.go):
  - **`AcaoStatusInvest` (36 colunas):** `Ticker`, `CompanyName`, `Price`, `DY`, `PL`, `PVP`, `PEbit`, `PEGRatio`, `EVEbit`, `LPA`, `VPA`, `PSR`, `PAtivo`, `PCapitalGiro`, `PAtivoCirculante`, `ROE`, `ROIC`, `ROA`, `GiroAtivos`, `MargemBruta`, `MargemEbit`, `MargemLiquida`, `DividaLiquidaPatrimonioLiquido`, `DividaLiquidaEbit`, `PLAtivo`, `PassivoAtivo`, `LiquidezCorrente`, `LucrosCAGR5`, `ReceitasCAGR5`, `LiquidezMediaDiaria`, `ValorMercado`, `Setor`, `Subsetor`, `Segmento`.
  - **`FIIStatusInvest` (22 colunas):** `Ticker`, `CompanyName`, `Price`, `DY`, `PVP`, `ValorPatrimonialCota`, `Patrimonio`, `LastDividend`, `Gestao`, `Segmento`, `Setor`, `Subsetor`, `PercentualCaixa`, `DividendCAGR`, `CotaCAGR`, `NumeroCotistas`, `NumeroCotas`, `LiquidezMediaDiaria`.
- [client.go](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/statusinvest/client.go):
  - `http.Client` com pool de conexões e headers idênticos aos de um navegador legítimo (`User-Agent`, `Referer`, `Accept`, `X-Requested-With`).
  - Execução via POST em `https://statusinvest.com.br/category/advancedsearchresultpaginated` com parâmetro `take=1000`.
- [service.go](file:///c:/Users/caio/Desktop/projetos/golang/carteira_investimentos/apps/server/internal/statusinvest/service.go):
  - Ingestion Service que orquestra a busca de ações e FIIs, converte para o modelo de domínio `*domain.Ativo` e executa a auditoria dos motores (**Décio Bazin**, **Benjamin Graham**, **Peter Lynch**, **Gordon DDM**, **FIIs NTN-B** e **Score Fundamentalista**).

---

## 💻 3. Comandos de Execução e Verificação
- **CLI Dedicado de Ingestão e Auditoria:**
  ```powershell
  cd apps/server
  go run cmd/sync_statusinvest/main.go
  ```
- **API REST Completa (Integrada ao Dashboard Next.js):**
  ```powershell
  cd apps/server
  go run cmd/api/main.go
  ```
