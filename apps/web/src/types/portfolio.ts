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
