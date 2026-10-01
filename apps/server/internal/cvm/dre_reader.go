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

	// Helper para obter ou instanciar struct do CNPJ
	obterOuCriar := func(cnpj string) *FundamentosAcao {
		cnpjLimpo := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(cnpj, ".", ""), "/", ""), "-", "")
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
				_, _ = r.Read() // pula cabeçalho
				for {
					l, err := r.Read()
					if err != nil || len(l) <= 6 {
						break
					}
					cnpj := l[0]
					totalAcoesStr := strings.TrimSpace(l[6]) // QT_ACAO_TOTAL_CAP_INTEGR
					if totalAcoes, err := strconv.ParseFloat(totalAcoesStr, 64); err == nil && totalAcoes > 0 {
						item := obterOuCriar(cnpj)
						item.TotalAcoes = totalAcoes
					}
				}
				rc.Close()
			}
			break
		}
	}

	// 2. Lê o Balanço Passivo Consolidado (Patrimônio Líquido - Conta 2.03)
	nomeBPP := fmt.Sprintf("itr_cia_aberta_BPP_con_%d.csv", ano)
	for _, f := range zipReader.File {
		if f.Name == nomeBPP {
			if rc, err := f.Open(); err == nil {
				r := csv.NewReader(rc)
				r.Comma = ';'
				_, _ = r.Read()
				for {
					l, err := r.Read()
					if err != nil || len(l) <= 12 {
						break
					}
					cnpj := l[0]
					cdConta := strings.TrimSpace(l[10]) // CD_CONTA
					ordem := strings.TrimSpace(l[8])    // ORDEM_EXERC

					// Conta 2.03 é Patrimônio Líquido Consolidado do exercício mais recente
					if cdConta == "2.03" && (ordem == "ÚLTIMO" || ordem == "ULTIMO") {
						valStr := strings.TrimSpace(l[12])
						if pl, err := strconv.ParseFloat(valStr, 64); err == nil {
							item := obterOuCriar(cnpj)
							// A CVM geralmente informa em milhares de Reais se ESCALA_MOEDA == "MIL"
							escala := strings.ToUpper(strings.TrimSpace(l[7]))
							if escala == "MIL" {
								pl = pl * 1000.0
							}
							item.PatrimonioLiquido = pl
						}
					}
				}
				rc.Close()
			}
			break
		}
	}

	// 3. Lê a DRE Consolidada (Lucro Líquido - Conta 3.11 ou 3.99)
	nomeDRE := fmt.Sprintf("itr_cia_aberta_DRE_con_%d.csv", ano)
	for _, f := range zipReader.File {
		if f.Name == nomeDRE {
			if rc, err := f.Open(); err == nil {
				r := csv.NewReader(rc)
				r.Comma = ';'
				_, _ = r.Read()
				for {
					l, err := r.Read()
					if err != nil || len(l) <= 12 {
						break
					}
					cnpj := l[0]
					cdConta := strings.TrimSpace(l[10])
					ordem := strings.TrimSpace(l[8])

					// 3.11 ou 3.99 correspondem ao Lucro Líquido Consolidado
					if (cdConta == "3.11" || cdConta == "3.99") && (ordem == "ÚLTIMO" || ordem == "ULTIMO") {
						valStr := strings.TrimSpace(l[12])
						if lucro, err := strconv.ParseFloat(valStr, 64); err == nil {
							item := obterOuCriar(cnpj)
							escala := strings.ToUpper(strings.TrimSpace(l[7]))
							if escala == "MIL" {
								lucro = lucro * 1000.0
							}
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
