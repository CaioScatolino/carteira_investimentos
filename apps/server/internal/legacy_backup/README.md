# 🏛️ Legacy Backup: CVM, B3 & Yahoo Finance Parsers

Este diretório preserva 100% dos códigos, parsers, rotinas de download e integrações desenvolvidos anteriormente, antes do pivô arquitetural para a API consolidada do StatusInvest.

## 📦 Estrutura Preservada

- **`cvm/`**:
  - `cia_reader.go`: Resolução de CNPJs cadastrados na CVM (priorização de Categoria A e BOLSA).
  - `dre_reader.go`: Leitura direta de relatórios ITR/DFP (Composição de Capital, BPP e DRE consolidada/individual).
  - `fii_reader.go`: Leitura de Informes Mensais de FIIs da CVM (Valor Patrimonial e rendimentos).
- **`b3/`**:
  - `events_reader.go`: Leitura de Eventos Corporativos em Dinheiro da B3 (Dividendos e JCP líquidos com IR descontado).
  - `cotahist.go` / `parser.go` / `client.go`: Download e parsing de cotações históricas diárias oficiais da B3.
- **`provider/`**:
  - `b3_provider.go`: Provedor híbrido usando arquivos da B3.
  - `yahoo.go`: Integração com Yahoo Finance via Goroutines concorrentes.
  - `brapi.go`: Integração com a API Brapi.
- **`scripts/`**:
  - Scripts de teste, inspeção contábil e benchmarking de indicadores.

---
> 🔒 **Garantia de Preservação**: Nenhum código ou lógica contábil anterior foi deletado. Caso seja necessário reativar qualquer um dos componentes legados, os arquivos estão intactos e prontos para uso.
