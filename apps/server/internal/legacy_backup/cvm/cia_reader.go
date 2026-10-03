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
	CNPJ         string
	DenomSocial  string
	DenomComerc  string
	CleanSocial  string
	CleanComerc  string
	IsBolsa      bool
	IsCategoriaA bool
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
			isBolsa := len(row) > 11 && strings.Contains(strings.ToUpper(row[11]), "BOLSA")
			isCatA := len(row) > 12 && strings.Contains(strings.ToUpper(row[12]), "CATEGORIA A")

			cias = append(cias, CiaAbertaCVM{
				CNPJ:         cleanCNPJ(row[0]),
				DenomSocial:  row[1],
				DenomComerc:  row[2],
				CleanSocial:  normalizar(row[1]),
				CleanComerc:  normalizar(row[2]),
				IsBolsa:      isBolsa,
				IsCategoriaA: isCatA,
			})
		}
	}

	cad := &CadastroCVM{
		cias: cias,
		aliases: map[string]string{
			"CMIG": "17155730000164", // CIA ENERG MINAS GERAIS - CEMIG (Holding Categoria A)
			"RDOR": "06047087000139", // REDE D'OR SÃO LUIZ S.A.
			"TAEE": "07859971000130", // TAESA
			"EQTL": "07703350000128", // EQUATORIAL ENERGIA S.A.
			"ITUB": "60701190000104", // ITAU UNIBANCO HOLDING S.A.
			"BBDC": "60746948000112", // BCO BRADESCO S.A.
			"BBAS": "00000000000191", // BCO DO BRASIL S.A.
			"SANB": "90400888000142", // BCO SANTANDER (BRASIL) S.A.
			"B3SA": "09346601000125", // B3 S.A. - BRASIL, BOLSA, BALCAO
			"VALE": "33592510000154", // VALE S.A.
			"PETR": "33000167000101", // PETROLEO BRASILEIRO S.A. PETROBRAS
			"WEGE": "84429695000111", // WEG S.A.
			"RENT": "02286479000108", // LOCALIZA RENT A CAR S.A.
			"BMOB": "09042817000105", // BEMOBI MOBILE TECH S.A.
			"ABEV": "07526557000100", // AMBEV S.A.
			"KLBN": "89637490000145", // KLABIN S.A.
			"CSNA": "33042730000104", // CIA SIDERURGICA NACIONAL
			"USIM": "60872504000140", // USINAS SIDER DE MINAS GERAIS S.A.
			"CPLE": "76483817000120", // COPEL
			"RADL": "61585865000151", // RAIA DROGASIL S.A.
			"LREN": "92754738000130", // LOJAS RENNER S.A.
			"BPAC": "30306294000145", // BANCO BTG PACTUAL S.A.
			"ASAI": "06057223000171", // SENDAS DISTRIBUIDORA S.A. (ASSAI)
			"EZTC": "08398414000198", // EZTEC
			"MRVE": "08343492000120", // MRV ENGENHARIA E PARTICIPACOES S.A.
			"BBSE": "17365717000172", // BB SEGURIDADE
			"RAIZ": "33453598000123", // RAIZEN S.A.
			"RECV": "03867610000182", // PETRORECONCAVO S.A.
			"SAPR": "76484013000145", // CIA SANEAMENTO DO PARANA - SANEPAR
			"BRAP": "00000000000191", // BRADESPAR S.A.
			"POMO": "88610314000130", // MARCOPOLO S.A.
			"CXSE": "34078697000118", // CAIXA SEGURIDADE PARTICIPACOES S.A.
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

	// 1. Prioridade absoluta: Alias com CNPJ direto de 14 dígitos
	if alias, ok := cad.aliases[radix]; ok && len(alias) == 14 {
		return alias
	}

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

	// 2. Primeira varredura: Prioriza empresas Categoria A e de BOLSA (Holdings negociadas)
	for _, termo := range buscaTerms {
		for i := range cad.cias {
			c := &cad.cias[i]
			if c.IsBolsa && c.IsCategoriaA {
				if strings.Contains(c.CleanSocial, termo) || strings.Contains(c.CleanComerc, termo) {
					return c.CNPJ
				}
			}
		}
	}

	// 3. Segunda varredura: Fallback para qualquer empresa ativa
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
