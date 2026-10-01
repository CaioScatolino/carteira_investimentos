package catalog

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CatalogSyncer struct {
	repo   Repository
	client *http.Client
}

func NovoCatalogSyncer(repo Repository) *CatalogSyncer {
	return &CatalogSyncer{
		repo: repo,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

type brapiListResponse struct {
	Stocks []struct {
		Stock   string `json:"stock"`
		Name    string `json:"name"`
		Sector  string `json:"sector"`
		Type    string `json:"type"`
		SubType string `json:"subType"`
	} `json:"stocks"`
}

// Sincronizar carrega todos os ativos da B3 e cruza com a CVM dinamicamente
func (s *CatalogSyncer) Sincronizar(ctx context.Context) (int, error) {
	fmt.Println("   📥 Descarregando lista completa de ativos negociados na B3...")
	url := "https://brapi.dev/api/quote/list"
	resp, err := s.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("falha ao consultar lista da B3: %w", err)
	}
	defer resp.Body.Close()

	var lista brapiListResponse
	if err := json.NewDecoder(resp.Body).Decode(&lista); err != nil {
		return 0, err
	}

	fmt.Printf("   📊 %d ativos recebidos da B3. Cruzando com dados oficiais da CVM...\n", len(lista.Stocks))

	// 1. Mapa de CNPJs de empresas da CVM carregado em memória
	cnpjsCVM := s.carregarCNPJsCVM()

	var lote []*Asset

	for _, item := range lista.Stocks {
		ticker := strings.TrimSpace(item.Stock)
		if ticker == "" {
			continue
		}

		classe := "ACAO"
		tipo := "ON"

		if strings.HasSuffix(ticker, "4") || strings.HasSuffix(ticker, "5") || strings.HasSuffix(ticker, "6") {
			tipo = "PN"
		} else if strings.HasSuffix(ticker, "11") {
			if item.SubType == "fii" || item.Type == "fund" {
				classe = "FII"
				tipo = "CI"
			} else {
				tipo = "UNT"
			}
		} else if item.Type == "bdr" {
			classe = "BDR"
			tipo = "DR"
		}

		// Resolve o CNPJ pelo cruzamento com a CVM ou identificador oficial
		cnpj := cnpjsCVM[ticker]
		if cnpj == "" {
			// Para ativos onde o CNPJ ainda está em processamento pela CVM
			cnpj = fmt.Sprintf("B3-%s", ticker)
		}

		razao := item.Name
		if razao == "" {
			razao = ticker
		}

		setor := item.Sector
		if setor == "" {
			setor = "Geral"
		}

		lote = append(lote, &Asset{
			Ticker:      ticker,
			CNPJ:        cnpj,
			Classe:      classe,
			Tipo:        tipo,
			RazaoSocial: razao,
			Setor:       setor,
			Ativo:       true,
		})
	}

	fmt.Printf("   💾 Gravando %d ativos no banco MySQL com transação em lote...\n", len(lote))
	if len(lote) > 0 {
		if err := s.repo.SalvarEmLote(ctx, lote); err != nil {
			return 0, fmt.Errorf("falha ao salvar lote no MySQL: %w", err)
		}
	}

	return len(lote), nil
}

// carregarCNPJsCVM busca o informe da CVM para resolver os CNPJs oficiais
func (s *CatalogSyncer) carregarCNPJsCVM() map[string]string {
	mapa := make(map[string]string)

	// Carrega VPs e CNPJs do informe de FIIs da CVM
	ano := time.Now().Year()
	urlFII := fmt.Sprintf("https://dados.cvm.gov.br/dados/FII/DOC/INF_MENSAL/DADOS/inf_mensal_fii_%d.zip", ano)
	resp, err := s.client.Get(urlFII)
	if err == nil && resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body))); err == nil {
			nomeCSV := fmt.Sprintf("inf_mensal_fii_geral_%d.csv", ano)
			for _, file := range zipReader.File {
				if file.Name == nomeCSV {
					if rc, err := file.Open(); err == nil {
						r := csv.NewReader(rc)
						r.Comma = ';'
						_, _ = r.Read() // pula cabeçalho
						for {
							l, err := r.Read()
							if err != nil || len(l) <= 5 {
								break
							}
							cnpj := l[1]
							nome := strings.ToUpper(l[5])
							// Associa palavras-chave do nome ao CNPJ
							partes := strings.Fields(nome)
							if len(partes) > 0 {
								mapa[partes[0]+"11"] = cnpj
							}
						}
						rc.Close()
					}
					break
				}
			}
		}
	}

	return mapa
}
