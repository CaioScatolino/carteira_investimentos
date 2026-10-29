export interface CandleData {
  time: number; // Unix timestamp em segundos
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface IndicatorPoint {
  time: number;
  value: number;
}

export interface RespostaGrafico {
  ticker: string;
  nome: string;
  intervalo: string;
  periodo: string;
  preco_atual: number;
  preco_teto_consolidado?: number;
  modelos_teto?: string[];
  preco_teto_bazin: number;
  valor_graham: number;
  margem_seguranca: number;
  sinal_tecnico: string;
  descricao_sinal: string;
  candles: CandleData[];
  mma8: IndicatorPoint[];
  mma21: IndicatorPoint[];
  mma20?: IndicatorPoint[];
  mma52?: IndicatorPoint[];
  mma200: IndicatorPoint[];
  bollinger_superior?: IndicatorPoint[];
  bollinger_inferior?: IndicatorPoint[];
  rsi14: IndicatorPoint[];
}
