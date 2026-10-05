package macro

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestMacroSincronizacao(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := "https://api.bcb.gov.br/dados/serie/bcdata.sgs.432/dados/ultimos/1?formato=json"
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Erro ao chamar BCB: %v", err)
	}
	defer resp.Body.Close()

	var resultados []RespostaBCBSGS
	json.NewDecoder(resp.Body).Decode(&resultados)

	t.Logf("Resultado bruto do BCB: %+v", resultados)

	SincronizarTaxasOficiais()
	cenario := ObterCenario()

	t.Logf("Taxa Selic no cenário: %.2f%% a.a.", cenario.TaxaSelic)
}
