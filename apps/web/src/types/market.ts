export type StatusRecomendacao = "COMPRAR_MAIS" | "MANTER" | "ALERTA";

export type StatusParecer = "APROVADO" | "ATENCAO" | "REPROVADO" | "NAO_APLICA";

export interface ParecerItem {
  status: StatusParecer;
  metrica: string;
  detalhe: string;
}

export interface Ativo {
  ticker: string;
  nome: string;
  classe: "ACAO" | "FII" | "ETF";
  cnpj?: string;
  volume_total: number;
  preco_atual: number;
  dividendos_12m: number;
  lpa: number;
  vpa: number;
  crescimento_lucro_5a: number;
  vp_cota: number;
  preco_teto_bazin: number;
  valor_graham: number;
  preco_justo_lynch: number;
  peg_ratio: number;
  pvp: number;
  spread_ntnb: number;
  score: number;
  dy: number;
  pl: number;
  pvp_real: number;
  roe: number;
  earnings_yield: number;
  preco_teto_gordon: number;
  margem_bazin: number;
  margem_graham: number;

  // Novos campos StatusInvest
  roic?: number;
  roa?: number;
  giro_ativos?: number;
  margem_bruta?: number;
  margem_ebit?: number;
  margem_liquida?: number;
  divida_liquida_pl?: number;
  divida_liquida_ebit?: number;
  liquidez_corrente?: number;
  liquidez_media_diaria?: number;
  valor_mercado?: number;
  setor?: string;
  subsetor?: string;
  segmento?: string;
  gestao?: string;
  percentual_caixa?: number;
  numero_cotistas?: number;
  last_dividend?: number;

  // Qualidade e Sustentabilidade de Proventos
  payout?: number;
  is_provento_atipico?: boolean;
  alerta_risco?: string;
  preco_teto_bazin_sustentavel?: number;

  pareceres?: Record<string, ParecerItem>;
  status: StatusRecomendacao;
}

export interface RespostaRankings {
  total: number;
  acoes: Ativo[];
  fiis: Ativo[];
  etfs: Ativo[];
}
