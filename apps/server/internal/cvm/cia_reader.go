package cvm

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// CiaAbertaCVM armazena as informações cadastrais de uma empresa na CVM
type CiaAbertaCVM struct {
	CNPJ        string
	DenomSocial string
	DenomComerc string
	CleanSocial string
	CleanComerc string
}

// CadastroCVM gerencia o índice de companhias em memória para resolução ultrarrápida
type CadastroCVM struct {
	cias    []CiaAbertaCVM
	aliases map[string]string
}

var (
	cacheCadastro *CadastroCVM
	cacheMu       sync.RWMutex
)

// normalizar remove acentuações e caracteres especiais para correspondência textual
func normalizar(s string) string {
	s = strings.ToUpper(s)
	replacer := strings.NewReplacer(
		"Á", "A", "À", "A", "Ã", "A", "Â", "A", "Ä", "A",
		"É", "E", "È", "E", "Ê", "E", "Ë", "E",
		"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
		"Ó", "O", "Ò", "O", "Õ", "O", "Ô", "O", "Ö", "O",
		"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
		"Ç", "C",
		".", "", "/", "", "-", "", ",", "", "(", "", ")", "", "'", "",
	)
	return strings.TrimSpace(replacer.Replace(s))
}

// ObterCadastroCVM baixa o arquivo oficial da CVM (cad_cia_aberta.csv) e cria o índice
func (c *CVMClient) ObterCadastroCVM() (*CadastroCVM, error) {
	cacheMu.RLock()
	if cacheCadastro != nil {
		defer cacheMu.RUnlock()
		return cacheCadastro, nil
	}
	cacheMu.RUnlock()

	url := "https://dados.cvm.gov.br/dados/CIA_ABERTA/CAD/DADOS/cad_cia_aberta.csv"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao baixar cadastro da CVM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CVM retornou HTTP %d para cadastro de cias", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(strings.NewReader(string(body)))
	r.Comma = ';'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	// Ignora cabeçalho
	_, _ = r.Read()

	var cias []CiaAbertaCVM
	cleanCNPJ := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), "/", ""), "-", "")
	}

	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		// Coluna 7: SIT (deve ser "ATIVO")
		if len(row) > 7 && row[7] == "ATIVO" {
			cias = append(cias, CiaAbertaCVM{
				CNPJ:        cleanCNPJ(row[0]),
				DenomSocial: row[1],
				DenomComerc: row[2],
				CleanSocial: normalizar(row[1]),
				CleanComerc: normalizar(row[2]),
			})
		}
	}

	cad := &CadastroCVM{
		cias: cias,
		aliases: map[string]string{
			"ITUB": "ITAU",
			"BBDC": "BRADESCO",
			"BBAS": "BANCO DO BRASIL",
			"SANB": "SANTANDER",
			"B3SA": "B3 SA",
			"ASAI": "ASSAI",
			"EZTC": "EZTEC",
			"MRVE": "MRV",
			"TAEE": "TAESA",
			"BBSE": "BB SEGURIDADE",
			"RECV": "PETRORECONCAVO",
			"RAIZ": "RAIZEN",
			"VALE": "VALE",
			"PETR": "PETROBRAS",
			"WEGE": "WEG",
			"RENT": "LOCALIZA",
			"CMIG": "CEMIG",
			"CPLE": "COPEL",
		},
	}

	cacheMu.Lock()
	cacheCadastro = cad
	cacheMu.Unlock()

	return cad, nil
}

// ResolverCNPJ encontra o CNPJ oficial da companhia pelo ticker e nome B3
func (cad *CadastroCVM) ResolverCNPJ(ticker, nomeB3 string) string {
	if len(ticker) < 4 {
		return ""
	}
	radix := ticker[:4]
	normNome := normalizar(nomeB3)
	palavras := strings.Fields(normNome)

	var buscaTerms []string
	if alias, ok := cad.aliases[radix]; ok {
		buscaTerms = append(buscaTerms, alias)
	}
	if len(palavras) > 0 && len(palavras[0]) >= 3 {
		buscaTerms = append(buscaTerms, palavras[0])
	}
	if len(normNome) >= 3 {
		buscaTerms = append(buscaTerms, normNome)
	}
	buscaTerms = append(buscaTerms, radix)

	for _, termo := range buscaTerms {
		for i := range cad.cias {
			c := &cad.cias[i]
			if strings.Contains(c.CleanSocial, termo) || strings.Contains(c.CleanComerc, termo) {
				return c.CNPJ
			}
		}
	}

	return ""
}
