# Passo 13: Provedor de Mercado B3 & Scanner em Lote sem Yahoo Finance

## 🎯 Objetivo
Criar o `B3MarketProvider` no pacote `internal/provider/` que implementa a interface `provider.MarketProvider`.
Ele substituirá o `YahooFinanceProvider` no Scanner Geral de Mercado (Frente 2), alimentando todo o pipeline de auditoria a partir do arquivo oficial diário **COTAHIST_D** da B3 e cruzando com o Catálogo no MySQL.

---

## 🧩 O que o `B3MarketProvider` faz:

1. **Download Oficial Único:**
   - Baixa o arquivo do pregão diário mais recente via `b3.B3Client` (~450 KB).
   - Faz o parse em streaming com `b3.ParseCOTAHIST` (< 10 ms).

2. **Filtro de Liquidez:**
   - Filtra apenas ativos com volume financeiro relevante no pregão (ex: `VolumeTotal >= R$ 500.000,00`), descartando "micos" e ativos sem negociação real.

3. **Cruzamento com o Catálogo MySQL:**
   - Identifica se o ticker é **AÇÃO** ou **FII** e obtém seu **CNPJ oficial** cadastrado no banco.

4. **Compatibilidade com o Worker:**
   - Retorna instâncias de `*domain.Ativo` com `PrecoAtual`, `VolumeTotal`, `Classe` e `Ticker` preenchidos oficialmente.

---

## 📂 Arquivos Envolvidos

- `apps/server/internal/provider/b3_provider.go` (Novo)
- `apps/server/internal/worker/sync_worker.go` (Atualização para usar catálogo MySQL de CNPJs)
- `apps/server/cmd/api/main.go` ou `cmd/cli/main.go` (Execução do scanner completo)
