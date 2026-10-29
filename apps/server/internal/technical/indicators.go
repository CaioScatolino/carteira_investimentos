package technical

import (
	"math"
)

// CalcularMMA calcula a Média Móvel Aritmética para um dado período
func CalcularMMA(candles []Candle, periodo int) []IndicatorPoint {
	if len(candles) < periodo || periodo <= 0 {
		return nil
	}

	var pontos []IndicatorPoint
	soma := 0.0

	// Calcula a primeira janela
	for i := 0; i < periodo; i++ {
		soma += candles[i].Close
	}

	primeiraMedia := math.Round((soma/float64(periodo))*100) / 100
	pontos = append(pontos, IndicatorPoint{
		Time:  candles[periodo-1].Time,
		Value: primeiraMedia,
	})

	// Desliza a janela para os próximos candles
	for i := periodo; i < len(candles); i++ {
		soma += candles[i].Close - candles[i-periodo].Close
		media := math.Round((soma/float64(periodo))*100) / 100
		pontos = append(pontos, IndicatorPoint{
			Time:  candles[i].Time,
			Value: media,
		})
	}

	return pontos
}

// CalcularBollinger calcula as Bandas de Bollinger (Superior e Inferior) com base na MMA e desvio padrão
func CalcularBollinger(candles []Candle, periodo int, multiplicador float64) ([]IndicatorPoint, []IndicatorPoint) {
	if len(candles) < periodo || periodo <= 0 {
		return nil, nil
	}

	var superior, inferior []IndicatorPoint

	for i := periodo - 1; i < len(candles); i++ {
		soma := 0.0
		for j := i - periodo + 1; j <= i; j++ {
			soma += candles[j].Close
		}
		media := soma / float64(periodo)

		varianciaSoma := 0.0
		for j := i - periodo + 1; j <= i; j++ {
			diff := candles[j].Close - media
			varianciaSoma += diff * diff
		}
		desvioPadrao := math.Sqrt(varianciaSoma / float64(periodo))

		sup := math.Round((media+multiplicador*desvioPadrao)*100) / 100
		inf := math.Round((media-multiplicador*desvioPadrao)*100) / 100

		superior = append(superior, IndicatorPoint{
			Time:  candles[i].Time,
			Value: sup,
		})
		inferior = append(inferior, IndicatorPoint{
			Time:  candles[i].Time,
			Value: inf,
		})
	}

	return superior, inferior
}

// CalcularRSI calcula o Índice de Força Relativa de Wilder (RSI) para o período padrão (ex: 14)
func CalcularRSI(candles []Candle, periodo int) []IndicatorPoint {
	if len(candles) <= periodo || periodo <= 0 {
		return nil
	}

	var pontos []IndicatorPoint

	var ganhos, perdas float64
	for i := 1; i <= periodo; i++ {
		diff := candles[i].Close - candles[i-1].Close
		if diff >= 0 {
			ganhos += diff
		} else {
			perdas += math.Abs(diff)
		}
	}

	mediaGanho := ganhos / float64(periodo)
	mediaPerda := perdas / float64(periodo)

	// Primeiro RSI no índice 'periodo'
	rsi := calcularValorRSI(mediaGanho, mediaPerda)
	pontos = append(pontos, IndicatorPoint{
		Time:  candles[periodo].Time,
		Value: math.Round(rsi*100) / 100,
	})

	// Suavização exponencial de Wilder para os candles subsequentes
	for i := periodo + 1; i < len(candles); i++ {
		diff := candles[i].Close - candles[i-1].Close
		var ganhoAtual, perdaAtual float64
		if diff >= 0 {
			ganhoAtual = diff
		} else {
			perdaAtual = math.Abs(diff)
		}

		mediaGanho = (mediaGanho*float64(periodo-1) + ganhoAtual) / float64(periodo)
		mediaPerda = (mediaPerda*float64(periodo-1) + perdaAtual) / float64(periodo)

		rsi = calcularValorRSI(mediaGanho, mediaPerda)
		pontos = append(pontos, IndicatorPoint{
			Time:  candles[i].Time,
			Value: math.Round(rsi*100) / 100,
		})
	}

	return pontos
}

func calcularValorRSI(mediaGanho, mediaPerda float64) float64 {
	if mediaPerda == 0 {
		if mediaGanho == 0 {
			return 50.0
		}
		return 100.0
	}
	if mediaGanho == 0 {
		return 0.0
	}

	rs := mediaGanho / mediaPerda
	return 100.0 - (100.0 / (1.0 + rs))
}

// AvaliarSinalTecnico sintetiza os sinais de Price Action e os cruza com o Preço Teto Fundamentalista
func AvaliarSinalTecnico(ultimoFechamento, bazinTeto float64, ultimoRSI, mma21, mma200 float64) (string, string) {
	temTeto := bazinTeto > 0
	abaixoDoTeto := temTeto && ultimoFechamento <= bazinTeto

	// 1. Oportunidade Suprema: Abaixo do Teto + IFR em Sobrevenda
	if abaixoDoTeto && ultimoRSI > 0 && ultimoRSI <= 35 {
		return "ZONA_COMPRA_SOBREVENDIDO", "🟢 Oportunidade Alta: Abaixo do Teto Bazin e em sobrevenda técnica (IFR <= 35)"
	}

	// 2. Compra Forte em Tendência: Abaixo do Teto + Acima das Médias 21 e 200
	if abaixoDoTeto && mma21 > 0 && ultimoFechamento >= mma21 && (mma200 == 0 || mma21 >= mma200) {
		return "COMPRA_TENDENCIA_ALTA", "🟢 Compra Forte: Abaixo do Teto Bazin e em tendência primária de alta (acima da MMA 21 e MMA 200)"
	}

	// 3. Abaixo do Teto consolidado
	if abaixoDoTeto {
		return "ZONA_COMPRA_BAZIN", "🟢 Zona de Compra: Abaixo do Preço Teto Bazin com margem de segurança positiva"
	}

	// 4. Sobrecompra Técnica Esticada
	if ultimoRSI >= 70 {
		return "SOBRECOMPRA_ALERTA", "⚠️ Alerta de Sobrecompra: IFR esticado (>= 70), risco elevado de correção no curto prazo"
	}

	// 5. Acima do Teto + Abaixo da MMA 200 (Bearish)
	if temTeto && ultimoFechamento > bazinTeto && mma200 > 0 && ultimoFechamento < mma200 {
		return "CORRECAO_TENDENCIA_BAIXA", "🔴 Cautela: Acima do Preço Teto e operando abaixo da MMA 200 em tendência de baixa"
	}

	// 6. Ágio sobre o Teto
	if temTeto && ultimoFechamento > bazinTeto {
		return "AGIO_SEM_MARGEM", "🟡 Ágio: Preço acima do Teto Bazin; aguardar correção técnica para aportar"
	}

	// 7. Acima da MMA 200 geral
	if mma200 > 0 && ultimoFechamento >= mma200 {
		return "TENDENCIA_ALTA", "📈 Tendência de Alta: Operando acima da média móvel institucional de 200 períodos"
	}

	return "NEUTRO", "⚪ Confluência Neutra: Monitorando oscilações e médias móveis"
}
