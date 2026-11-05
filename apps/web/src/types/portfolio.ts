export type ClasseAtivo = "RENDA_FIXA" | "FII" | "ACAO_BR" | "INTERNACIONAL" | "CAIXA";

export interface Posicao {
  id: number;
  usuario_id: number;
  ticker: string;
  classe: ClasseAtivo;
  categoria_macro: string;
  quantidade: number;
  preco_medio: number;
  valor_aplicado: number;
  taxa_contratada?: string | null;
  indexador?: string | null;
  vencimento?: string | null;
  corretora: string;
  observacoes?: string | null;
  created_at: string;
  updated_at: string;
}

export interface PosicaoConsolidada extends Posicao {
  cotacao_atual: number;
  valor_atual_mercado: number;
  lucro_prejuizo_valor: number;
  rentabilidade_pct: number;
  preco_teto_estimado: number;
  margem_seguranca_pct: number;
  score_fundamentalista: number;
  sinal_rebalanceamento: "COMPRAR_APORTE" | "AGUARDAR" | "REAVALIAR" | string;
  percentual_na_classe?: number;
  percentual_na_carteira?: number;
  total_proventos?: number;
  preco_medio_descontado?: number;
  lucro_com_proventos?: number;
  rentabilidade_total_pct?: number;
  proventos_yoc_pct?: number;
}

export interface ClasseConsolidada {
  classe: ClasseAtivo;
  nome: string;
  valor_atual_mercado: number;
  total_investido: number;
  lucro_prejuizo_valor: number;
  rentabilidade_pct: number;
  percentual_na_carteira: number;
  percentual_alvo: number;
  quantidade_ativos: number;
  posicoes: PosicaoConsolidada[];
  total_proventos?: number;
  lucro_com_proventos?: number;
}

export interface ResumoCarteira {
  patrimonio_total: number;
  total_investido: number;
  lucro_total_valor: number;
  rentabilidade_pct: number;
  posicoes: PosicaoConsolidada[];
  classes?: ClasseConsolidada[];
  alocacao_atual_pct: Record<ClasseAtivo, number>;
  alocacao_alvo_pct: Record<ClasseAtivo, number>;
  total_proventos_recebidos?: number;
  lucro_total_com_proventos?: number;
  rentabilidade_total_pct?: number;
}

export interface CriarPosicaoRequest {
  ticker: string;
  classe: ClasseAtivo;
  categoria_macro: string;
  quantidade: number;
  preco_medio: number;
  taxa_contratada?: string;
  indexador?: string;
  vencimento?: string;
  corretora: string;
  observacoes?: string;
  tipo_operacao?: "APORTE" | "SUBSTITUIR" | "COMPRA";
  data_operacao?: string;
}

export interface DiagnosticoMacroClasse {
  classe: ClasseAtivo;
  percentual_atual: number;
  percentual_alvo: number;
  delta_percentual: number;
  valor_atual: number;
  valor_ideal: number;
  necessidade_aporte: number;
  status: "DEFASADO_APORTAR" | "EQUILIBRADO" | "SOBREALOCADO" | string;
}

export interface SugestaoAporteItem {
  ticker: string;
  classe: ClasseAtivo;
  quantidade_sugerida: number;
  preco_estimado: number;
  valor_total_sugerido: number;
  score: number;
  margem_seguranca_pct: number;
  justificativa: string;
}

export interface ResultadoRebalanceamento {
  valor_aporte_informado: number;
  valor_aporte_efetivo: number;
  saldo_restante_caixa: number;
  diagnostico_macro: DiagnosticoMacroClasse[];
  sugestoes: SugestaoAporteItem[];
}

export interface Transacao {
  id: number;
  usuario_id: number;
  posicao_id?: number | null;
  ticker: string;
  tipo: "COMPRA" | "VENDA" | "APORTE" | "RESGATE";
  quantidade: number;
  preco_unitario: number;
  custos_taxas: number;
  valor_total: number;
  data_operacao: string;
  corretora: string;
  notas?: string | null;
  created_at: string;
}

export type PerfilAdvisor = "EQUILIBRADO" | "DIVIDENDOS" | "OPORTUNIDADE" | "DEFENSIVO";

export interface AssimetriaDetectada {
  ticker: string;
  nome: string;
  classe: ClasseAtivo;
  cotacao_atual: number;
  preco_teto: number;
  margem_seguranca_pct: number;
  score: number;
  dividend_yield: number;
  motivo: string;
}

export interface CenarioMacro {
  taxa_selic: number;
  taxa_ntnb: number;
  spread_minimo_acoes?: number;
  data_atualizacao?: string;
  fonte?: string;
}

export interface DiagnosticoAdvisor {
  valor_aporte_informado: number;
  valor_aporte_efetivo: number;
  saldo_restante_caixa: number;
  patrimonio_atual: number;
  patrimonio_projetado: number;
  classe_mais_defasada: ClasseAtivo;
  delta_defasagem: number;
  cenario_macro?: CenarioMacro;
  assimetrias: AssimetriaDetectada[];
  sugestoes: SugestaoAporteItem[];
  diagnostico_macro: DiagnosticoMacroClasse[];
  parecer_sintese?: string;
}

export interface PosicaoImportada {
  ticker: string;
  razao_social: string;
  classe: ClasseAtivo;
  quantidade: number;
  preco_medio: number;
  valor_aplicado: number;
  corretora: string;
  total_compras: number;
  total_vendas: number;
  total_proventos_recebidos: number;
}

export interface OperacaoImportada {
  ticker: string;
  classe: ClasseAtivo;
  tipo: "COMPRA" | "VENDA";
  data: string;
  quantidade: number;
  preco_unitario: number;
  valor_total: number;
  corretora: string;
  origem: string;
  notas: string;
}

export interface ProventoImportado {
  ticker: string;
  tipo: string;
  data_pagamento: string;
  quantidade: number;
  valor_unitario: number;
  valor_total: number;
  corretora: string;
}

export interface RelatorioAnaliseB3 {
  total_movimentacoes: number;
  data_inicio: string;
  data_fim: string;
  total_ativos_ativos: number;
  total_investido_custo: number;
  total_proventos_historicos: number;
  posicoes: PosicaoImportada[];
  operacoes: OperacaoImportada[];
  proventos: ProventoImportado[];
}

export interface ResultadoImportacao {
  posicoes_salvas: number;
  operacoes_salvas: number;
  proventos_salvos: number;
  total_investido: number;
  total_proventos: number;
  dados_substituidos: boolean;
}

export interface TotalAnoProvento {
  ano: number;
  valor: number;
}

export interface SazonalidadeMesProvento {
  mes: number;
  nome: string;
  valor: number;
}

export interface AtivoPagador {
  ticker: string;
  total_recebido: number;
  percentual: number;
  total_eventos: number;
}

export interface ProventoRecebidoItem {
  id: number;
  ticker: string;
  tipo: string;
  data_pagamento: string;
  quantidade: number;
  valor_unitario: number;
  valor_liquido: number;
  corretora: string;
}

export interface ResumoProventos {
  total_acumulado: number;
  media_mensal_12m: number;
  total_recebido_12m: number;
  total_por_ano: TotalAnoProvento[];
  sazonalidade_mensal: SazonalidadeMesProvento[];
  top_ativos_pagadores: AtivoPagador[];
  ultimos_pagamentos: ProventoRecebidoItem[];
}

export interface ProventoFuturo {
  ticker: string;
  tipo: string;
  data_com: string;
  data_pagamento: string;
  valor_unitario: number;
  quantidade_cotas: number;
  valor_previsto: number;
  status: "DATA_COM_IMINENTE" | "AGUARDANDO_PAGAMENTO" | "CREDITADO" | string;
}
