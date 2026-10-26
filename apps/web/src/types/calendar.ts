export interface MesProvento {
  mes: number;
  nome_mes: string;
  probabilidade: number;
  ocorrencias: number;
  total_anos: number;
  valor_medio: number;
  anos_com_provento: number[];
  tipos: string[];
}

export interface AtivoCalendario {
  ticker: string;
  nome: string;
  classe: "ACAO" | "FII" | "ETF";
  setor: string;
  segmento: string;
  preco_atual: number;
  dy: number;
  score: number;
  porte: string;
  status: "COMPRAR_MAIS" | "MANTER" | "ALERTA";
  meses: MesProvento[];
  total_meses_pagadores: number;
  mes_mais_provavel: string;
  maior_probabilidade: number;
  total_eventos_historico: number;
  possui_na_carteira: boolean;
}

export interface ResumoCalendario {
  total_ativos: number;
  limite?: number;
  offset?: number;
  tem_mais?: boolean;
  tipo_data: "DATA_COM" | "DATA_PAGAMENTO";
  periodo_anos: number;
  probabilidade_minima: number;
  meses_mais_frequentes: Record<string, number>;
  ativos: AtivoCalendario[];
}

export interface FiltroCalendarioFrontend {
  tipoData: "DATA_COM" | "DATA_PAGAMENTO";
  probMin: number;
  apenasCarteira: boolean;
  classe: string;
  setor: string;
  busca: string;
}
