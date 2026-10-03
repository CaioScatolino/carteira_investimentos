# Passo 17: Dashboard Web Institucional em Next.js e Tailwind CSS v4

## 🎯 Objetivo Arquitetural
Construir uma interface moderna, minimalista e de nível institucional (estilo Bloomberg Terminal / Koyfin), conectando-se diretamente à API REST em Go (`http://localhost:8080`).

---

## 🎨 Identidade Visual e Experiência do Usuário (UX)
- **Paleta de Cores:** Fundo preto ônix profundo (`#070709` / `#0c0c10`), recipientes em grafite escuro (`#131318`) e detalhes sofisticados em **dourado champagne** (`#d4af37` / `#f5df88`).
- **Sem visual genérico de IA:** Tipografia de dados tabulares (`tabular-nums`, `font-mono`), badges semânticos de alta legibilidade (`COMPRAR`, `MANTER`, `ALERTA`) e bordas douradas sutis com micro-brilho (`gold-glow`).
- **Arquitetura de Dados em Tela Única:** 
  1. 📈 **Tabela de Ações (206 ativos):** P/L, P/VP, ROE, DY 12M, Teto Bazin, VI Graham e Semáforo.
  2. 🏢 **Tabela de Fundos Imobiliários (85 FIIs):** P/VP CVM, VP Cota, DY 12M, Spread NTN-B e Teto Bazin.
  3. 🌐 **Tabela de ETFs (99 ativos):** Volume Diário B3 e Classificação de Liquidez Institucional.
- **Funcionalidades Interativas:**
  - Paginação individual e independente por tabela (com seletor de linhas por página).
  - Filtro e pesquisa instantânea por Ticker ou Nome da empresa/fundo.
  - Modal profundo de inspeção ao clicar em qualquer ativo, com a decomposição de todas as escolas de valuation (**Décio Bazin**, **Benjamin Graham**, **Peter Lynch PEG**, **Gordon DDM**, **Joel Greenblatt** e laudo oficial CVM).
- **Responsividade:** Estrutura pronta para Desktop e Mobile com barras horizontais e quebras fluidas.
