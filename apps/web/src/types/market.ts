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
  yield_exigido?: number;
  veredito_risco?: "COMPENSA_RISCO" | "NEUTRO" | "RISCO_DESCOMPENSADO";
  justificativa_risco?: string;
  tir_projetada?: number;
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

  // Porte e Robustez Institucional
  porte?: "BLUE_CHIP" | "MID_CAP" | "SMALL_CAP" | "MICRO_CAP" | "FII_GIGANTE" | "FII_CONSOLIDADO" | "FII_MEDIO" | "FII_CONCENTRADO";

  // Proventos Décio Bazin 5 Anos (B3 Oficial vs StatusInvest)
  media_dividendos_5a?: number;
  media_dividendos_5a_normalizada?: number;
  preco_teto_bazin_5a?: number;
  margem_bazin_5a?: number;
  dividendos_12m_b3?: number;
  historico_dividendos_anual?: Record<string, number>;
  diferenca_12m_b3_statusinvest?: number;
  aderencia_statusinvest?: string;
  teve_outlier_5a?: boolean;
  observacao_outlier?: string;

  // Preço Teto Consolidado e Prêmio de Valorização (Consenso Multi-Modelo)
  preco_teto_consolidado?: number;
  premio_desconto_percentual?: number;
  margem_seguranca_consolidada?: number;
  modelos_teto_consolidado?: string[];

  pareceres?: Record<string, ParecerItem>;
  status: StatusRecomendacao;
}

export interface AnaliseProventos5A {
  ticker: string;
  trading_name: string;
  total_eventos: number;
  proventos_por_ano: Record<string, number>;
  media_dividendos_5a: number;
  media_dividendos_5a_normalizada?: number;
  preco_teto_bazin_5a: number;
  margem_bazin_5a: number;
  dividendos_12m_b3_liquido: number;
  dividendos_12m_b3_bruto: number;
  dividendos_12m_statusinvest: number;
  diferenca_12m: number;
  aderencia_percentual: number;
  aderencia_statusinvest: string;
  teve_outlier_5a?: boolean;
  observacao_outlier?: string;
  observacao: string;
}

export interface RespostaRankings {
  total: number;
  acoes: Ativo[];
  fiis: Ativo[];
  etfs: Ativo[];
}

export interface CenarioMacro {
  taxa_selic: number;
  taxa_ntnb: number;
  spread_minimo_acoes: number;
  spread_minimo_fii_tijolo: number;
  equity_risk_premium: number;
  data_atualizacao: string;
  fonte: string;
}

