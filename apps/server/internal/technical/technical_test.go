package technical

import (
	"testing"
)

func TestCalcularMMA(t *testing.T) {
	candles := []Candle{
		{Time: 1000, Close: 10.0},
		{Time: 2000, Close: 11.0},
		{Time: 3000, Close: 12.0},
		{Time: 4000, Close: 13.0},
		{Time: 5000, Close: 14.0},
	}

	mma3 := CalcularMMA(candles, 3)
	if len(mma3) != 3 {
		t.Fatalf("esperava 3 pontos de MMA3, obteve %d", len(mma3))
	}

	// Primeiro ponto: (10 + 11 + 12)/3 = 11.00
	if mma3[0].Value != 11.00 {
		t.Errorf("esperava MMA3[0] = 11.00, obteve %.2f", mma3[0].Value)
	}
	if mma3[0].Time != 3000 {
		t.Errorf("esperava Time 3000, obteve %d", mma3[0].Time)
	}

	// Segundo ponto: (11 + 12 + 13)/3 = 12.00
	if mma3[1].Value != 12.00 {
		t.Errorf("esperava MMA3[1] = 12.00, obteve %.2f", mma3[1].Value)
	}

	// Terceiro ponto: (12 + 13 + 14)/3 = 13.00
	if mma3[2].Value != 13.00 {
		t.Errorf("esperava MMA3[2] = 13.00, obteve %.2f", mma3[2].Value)
	}
}

func TestCalcularRSI(t *testing.T) {
	// Série de 16 candles alternados
	var candles []Candle
	precos := []float64{
		44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08,
		45.89, 46.03, 45.61, 46.28, 46.28, 46.00,
	}
	for i, p := range precos {
		candles = append(candles, Candle{
			Time:  int64(i * 1000),
			Close: p,
		})
	}

	rsi := CalcularRSI(candles, 14)
	if len(rsi) != 2 {
		t.Fatalf("esperava 2 pontos de RSI 14 para 16 candles, obteve %d", len(rsi))
	}

	for _, pt := range rsi {
		if pt.Value < 0 || pt.Value > 100 {
			t.Errorf("RSI fora do intervalo [0, 100]: %.2f", pt.Value)
		}
	}
}

func TestAvaliarSinalTecnico(t *testing.T) {
	// 1. Oportunidade Suprema: Abaixo do Teto Bazin + IFR em Sobrevenda (<= 35)
	sinal, _ := AvaliarSinalTecnico(30.0, 40.0, 28.5, 32.0, 31.0)
	if sinal != "ZONA_COMPRA_SOBREVENDIDO" {
		t.Errorf("esperava ZONA_COMPRA_SOBREVENDIDO, obteve %s", sinal)
	}

	// 2. Compra em Tendência: Abaixo do Teto Bazin + Acima da MMA21 e MMA200
	sinal, _ = AvaliarSinalTecnico(35.0, 40.0, 55.0, 34.0, 30.0)
	if sinal != "COMPRA_TENDENCIA_ALTA" {
		t.Errorf("esperava COMPRA_TENDENCIA_ALTA, obteve %s", sinal)
	}

	// 3. Sobrecompra Técnica (IFR >= 70)
	sinal, _ = AvaliarSinalTecnico(45.0, 40.0, 75.0, 42.0, 38.0)
	if sinal != "SOBRECOMPRA_ALERTA" {
		t.Errorf("esperava SOBRECOMPRA_ALERTA, obteve %s", sinal)
	}

	// 4. Ágio sobre o Teto Bazin
	sinal, _ = AvaliarSinalTecnico(45.0, 40.0, 50.0, 46.0, 48.0)
	if sinal != "CORRECAO_TENDENCIA_BAIXA" { // mma200 = 48.0 > preco = 45.0
		t.Errorf("esperava CORRECAO_TENDENCIA_BAIXA, obteve %s", sinal)
	}

	// 5. Ágio sem correção baixa
	sinal, _ = AvaliarSinalTecnico(45.0, 40.0, 50.0, 43.0, 42.0)
	if sinal != "AGIO_SEM_MARGEM" {
		t.Errorf("esperava AGIO_SEM_MARGEM, obteve %s", sinal)
	}
}
