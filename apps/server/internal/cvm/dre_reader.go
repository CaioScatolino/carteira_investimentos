package cvm

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// FundamentosAcao guarda as métricas contábeis oficiais extraídas da CVM
type FundamentosAcao struct {
	CNPJ              string
	LucroLiquido      float64
	PatrimonioLiquido float64
	TotalAcoes        float64
	LPA               float64
	VPA               float64
}

// ObterFundamentosAcoes baixa o ITR do ano corrente e calcula LPA e VPA reais de cada empresa
func (c *CVMClient) ObterFundamentosAcoes() (map[string]*FundamentosAcao, error) {
	ano := time.Now().Year()
	url := fmt.Sprintf("https://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/ITR/DADOS/itr_cia_aberta_%d.zip", ano)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao baixar ITR da CVM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CVM retornou HTTP %d para ITR de %d", resp.StatusCode, ano)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}

	mapa := make(map[string]*FundamentosAcao)

	cleanCNPJ := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), "/", ""), "-", "")
	}

	obterOuCriar := func(cnpj string) *FundamentosAcao {
		cnpjLimpo := cleanCNPJ(cnpj)
		if _, existe := mapa[cnpjLimpo]; !existe {
			mapa[cnpjLimpo] = &FundamentosAcao{CNPJ: cnpjLimpo}
		}
		return mapa[cnpjLimpo]
	}

	// 1. Lê a Composição de Capital (Total de Ações Emitidas)
	nomeCapital := fmt.Sprintf("itr_cia_aberta_composicao_capital_%d.csv", ano)
	for _, f := range zipReader.File {
		if f.Name == nomeCapital {
			if rc, err := f.Open(); err == nil {
				r := csv.NewReader(rc)
				r.Comma = ';'
				r.LazyQuotes = true
				r.FieldsPerRecord = -1
				h, _ := r.Read()
				idxAcoes := 6
				for i, col := range h {
					if col == "QT_ACAO_TOTAL_CAP_INTEGR" {
						idxAcoes = i
						break
					}
				}
				for {
					l, err := r.Read()
					if err != nil {
						break
					}
					if len(l) > idxAcoes {
						if totalAcoes, err := strconv.ParseFloat(l[idxAcoes], 64); err == nil && totalAcoes > 0 {
							item := obterOuCriar(l[0])
							// Empresas que reportam ações em milhares na CVM (ex: Vale, Itaú)
							if totalAcoes < 100000000 && !strings.Contains(l[0], "33.000.167") {
								totalAcoes *= 1000.0
							}
							item.TotalAcoes = totalAcoes
						}
					}
				}
				rc.Close()
			}
			break
		}
	}

	// 2. Lê o Balanço Passivo Consolidado (Patrimônio Líquido Consolidado)
	nomeBPP := fmt.Sprintf("itr_cia_aberta_BPP_con_%d.csv", ano)
	for _, f := range zipReader.File {
		if f.Name == nomeBPP {
			if rc, err := f.Open(); err == nil {
				r := csv.NewReader(rc)
				r.Comma = ';'
				r.LazyQuotes = true
				r.FieldsPerRecord = -1
				_, _ = r.Read()
				for {
					l, err := r.Read()
					if err != nil {
						break
					}
					if len(l) <= 12 {
						continue
					}
					ordem := l[8]
					isUltimo := strings.HasSuffix(ordem, "LTIMO") && !strings.Contains(ordem, "PEN")
					dsConta := strings.ToUpper(l[11])

					// O nome oficial infalível da CVM é 'Patrimônio Líquido Consolidado'
					if isUltimo && strings.Contains(dsConta, "PATRIM") && strings.Contains(dsConta, "CONSOLIDADO") && !strings.Contains(dsConta, "CONTROL") {
						if pl, err := strconv.ParseFloat(l[12], 64); err == nil && pl > 0 {
							if strings.ToUpper(l[7]) == "MIL" {
								pl *= 1000.0
							}
							item := obterOuCriar(l[0])
							item.PatrimonioLiquido = pl
						}
					}
				}
				rc.Close()
			}
			break
		}
	}

	// 3. Lê a DRE Consolidada (Lucro Líquido Consolidado)
	nomeDRE := fmt.Sprintf("itr_cia_aberta_DRE_con_%d.csv", ano)
	for _, f := range zipReader.File {
		if f.Name == nomeDRE {
			if rc, err := f.Open(); err == nil {
				r := csv.NewReader(rc)
				r.Comma = ';'
				r.LazyQuotes = true
				r.FieldsPerRecord = -1
				_, _ = r.Read()
				for {
					l, err := r.Read()
					if err != nil {
						break
					}
					if len(l) <= 13 {
						continue
					}
					ordem := l[8]
					isUltimo := strings.HasSuffix(ordem, "LTIMO") && !strings.Contains(ordem, "PEN")
					dsConta := strings.ToUpper(l[12])

					// O nome oficial infalível da CVM é 'Lucro/Prejuízo Consolidado do Período'
					if isUltimo && strings.Contains(dsConta, "LUCRO") && strings.Contains(dsConta, "CONSOLIDADO") && !strings.Contains(dsConta, "CONTROL") {
						if lucro, err := strconv.ParseFloat(l[13], 64); err == nil {
							if strings.ToUpper(l[7]) == "MIL" {
								lucro *= 1000.0
							}
							item := obterOuCriar(l[0])
							item.LucroLiquido = lucro
						}
					}
				}
				rc.Close()
			}
			break
		}
	}

	// 4. Derivação Matemática Oficial: LPA = Lucro / Ações | VPA = PL / Ações
	for _, item := range mapa {
		if item.TotalAcoes > 0 {
			if item.LucroLiquido > 0 {
				item.LPA = item.LucroLiquido / item.TotalAcoes
			}
			if item.PatrimonioLiquido > 0 {
				item.VPA = item.PatrimonioLiquido / item.TotalAcoes
			}
		}
	}

	return mapa, nil
}
