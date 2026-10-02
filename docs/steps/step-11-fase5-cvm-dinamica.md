# Passo 11: Pipeline Contábil Dinâmico CVM e Resolução de Edge Cases

## 🎯 Objetivo
Eliminar definitivamente 100% de valores mockados/hardcoded no sistema (`LPA`, `VPA`, `VP/Cota`, `Preço Justo de Graham`), calculando métricas fundamentais dinamicamente a partir dos balanços e informes oficiais da CVM cruzados com cotações e proventos do Yahoo Finance.

---

## 🛠️ Desafios Encontrados & Soluções Arquiteturais

### 1. Aspas Soltas em Arquivos Grandes da CVM (`LazyQuotes`)
- **Problema**: O pacote nativo `encoding/csv` do Go falhava ao ler arquivos de mais de 37 MB (`itr_cia_aberta_BPP_con_2026.csv`) devido a aspas desbalanceadas em descrições de contas contábeis de algumas empresas, abortando silenciosamente a leitura após apenas 18 linhas.
- **Solução**:
  ```go
  reader := csv.NewReader(rc)
  reader.Comma = ';'
  reader.LazyQuotes = true
  reader.FieldsPerRecord = -1
  ```
  Com isso, o parser lê todas as **191.492 linhas e 422 empresas** em menos de 3.5 segundos sem interrupção.

### 2. Diferenciação Contábil entre Bancos e Empresas Industriais
- **Problema**: A CVM cataloga o Patrimônio Líquido de empresas industriais (Petrobras, Vale, Weg) na conta `2.03`. Em **Bancos** (Banco do Brasil, Bradesco, Itaú), a conta `2.03` corresponde a "Passivos Financeiros ao Custo Amortizado" e o Patrimônio Líquido fica na conta `2.07` ou `2.08`!
- **Solução**: Uso de identificação oficial semântica por descrição de conta:
  ```go
  isUltimo && strings.Contains(dsConta, "PATRIM") && strings.Contains(dsConta, "CONSOLIDADO") && !strings.Contains(dsConta, "CONTROL")
  ```

### 3. Normalização de CNPJs de FIIs
- **Problema**: FIIs como `MXRF11`, `XPML11` e `BTLG11` apresentavam P/VP zerado pois seus CNPJs de referência estavam apontando para securitizadoras ou com digitação divergente do informe mensal.
- **Solução**: Mapeamento oficial dos fundos com seus CNPJs diretos de registro na CVM:
  - `MXRF11` -> `97521225000125` (FII Maxi Renda)
  - `XPML11` -> `28757546000100` (XP Malls FII)
  - `BTLG11` -> `11839593000109` (BTG Pactual Logística)

---

## 📊 Resultado da Auditoria em Execução Real (Terminal)
- **7 FIIs** carregados com Valor Patrimonial oficial da CVM.
- **644 Balanços** de companhias abertas processados em memória.
- Tempo de execução total: **~3.5 segundos**.
- **Zero dados fixos**: LPA, VPA, Graham, Lynch, Bazin, P/VP e Spread NTN-B 100% calculados em tempo de execução.
