"use client";

import React from "react";
import { ClasseAtivo, ResumoCarteira } from "@/types/portfolio";
import { ShieldCheck, Building2, TrendingUp, Globe, Coins } from "lucide-react";

interface MacroAllocationViewProps {
  resumo: ResumoCarteira;
}

const META_CONFIG: Record<
  ClasseAtivo,
  { nome: string; icon: React.ReactNode; corBarra: string; corTexto: string; desc: string }
> = {
  RENDA_FIXA: {
    nome: "Renda Fixa",
    icon: <ShieldCheck className="w-4 h-4" />,
    corBarra: "bg-blue-500",
    corTexto: "text-blue-400",
    desc: "Tesouro Direto, CDBs, LCIs e Debêntures (Carrego de Juros)",
  },
  FII: {
    nome: "Fundos Imobiliários",
    icon: <Building2 className="w-4 h-4" />,
    corBarra: "bg-purple-500",
    corTexto: "text-purple-400",
    desc: "Tijolo e Papel auditados na B3 (Proventos Mensais Isentos)",
  },
  ACAO_BR: {
    nome: "Ações Brasileiras",
    icon: <TrendingUp className="w-4 h-4" />,
    corBarra: "bg-emerald-500",
    corTexto: "text-emerald-400",
    desc: "Companhias abertas na B3 com foco em dividendos e assimetria",
  },
  INTERNACIONAL: {
    nome: "Ações Internacionais",
    icon: <Globe className="w-4 h-4" />,
    corBarra: "bg-cyan-500",
    corTexto: "text-cyan-400",
    desc: "ETFs globais, BDRs e Stocks (Proteção Geográfica/Dolarizada)",
  },
  CAIXA: {
    nome: "Caixa / Oportunidade",
    icon: <Coins className="w-4 h-4" />,
    corBarra: "bg-amber-500",
    corTexto: "text-amber-400",
    desc: "Liquidez imediata (D+0) para compras táticas em momentos de estresse",
  },
};

export function MacroAllocationView({ resumo }: MacroAllocationViewProps) {
  const classes: ClasseAtivo[] = ["RENDA_FIXA", "FII", "ACAO_BR", "INTERNACIONAL", "CAIXA"];

  return (
    <div className="p-6 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-2xl mb-8 backdrop-blur-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-6 border-b border-neutral-800/60 pb-4">
        <div>
          <h3 className="text-base font-bold text-white flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-[#d4af37]" />
            Macro Alocação Patrimonial • Regra Oficial Institucional
          </h3>
          <p className="text-xs text-neutral-400 mt-1">
            Distribuição balanceada para estabilidade, proventos isentos e crescimento patrimonial sem giros fiscais.
          </p>
        </div>
        <div className="text-xs font-mono text-neutral-400 self-start sm:self-auto bg-neutral-900/80 px-3 py-1.5 rounded-lg border border-neutral-800">
          Patrimônio Base: <span className="text-white font-bold">{new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(resumo.patrimonio_total)}</span>
        </div>
      </div>

      <div className="space-y-5">
        {classes.map((classe) => {
          const config = META_CONFIG[classe];
          const atualPct = resumo.alocacao_atual_pct[classe] || 0;
          const alvoPct = resumo.alocacao_alvo_pct[classe] || 0;
          const deltaPct = atualPct - alvoPct;

          const valorAtualClasse = (atualPct / 100) * resumo.patrimonio_total;
          const valorAlvoClasse = (alvoPct / 100) * resumo.patrimonio_total;

          // Status pedagógico
          let statusBadge = (
            <span className="text-[11px] font-mono px-2 py-0.5 rounded-md bg-neutral-800 text-neutral-300 border border-neutral-700">
              Equilibrado
            </span>
          );
          if (deltaPct < -2.0) {
            statusBadge = (
              <span className="text-[11px] font-mono font-semibold px-2 py-0.5 rounded-md bg-cyan-500/15 text-cyan-300 border border-cyan-500/30">
                Defasado • Priorizar Aporte
              </span>
            );
          } else if (deltaPct > 3.0) {
            statusBadge = (
              <span className="text-[11px] font-mono font-semibold px-2 py-0.5 rounded-md bg-amber-500/15 text-amber-300 border border-amber-500/30">
                Sobrealocado • Aguardar
              </span>
            );
          }

          return (
            <div key={classe} className="p-4 rounded-xl bg-neutral-900/40 border border-neutral-800/60 hover:border-neutral-700/80 transition-all">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-2.5">
                <div className="flex items-center gap-2.5">
                  <div className={`p-1.5 rounded-lg bg-neutral-800 border border-neutral-700 ${config.corTexto}`}>
                    {config.icon}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-bold text-white">{config.nome}</span>
                      {statusBadge}
                    </div>
                    <span className="text-[11px] text-neutral-400 hidden sm:inline">{config.desc}</span>
                  </div>
                </div>

                <div className="flex items-center gap-4 text-xs font-mono">
                  <div className="text-right">
                    <span className="text-neutral-400 text-[11px] block">Atual vs Meta</span>
                    <span className={`font-bold ${config.corTexto}`}>{atualPct.toFixed(1)}%</span>
                    <span className="text-neutral-500"> / {alvoPct.toFixed(1)}%</span>
                  </div>
                  <div className="text-right hidden sm:block">
                    <span className="text-neutral-400 text-[11px] block">Valor Atual</span>
                    <span className="text-white font-medium">{new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(valorAtualClasse)}</span>
                  </div>
                </div>
              </div>

              {/* Barra de Progresso Comparativa */}
              <div className="w-full h-3 rounded-full bg-neutral-800/80 overflow-hidden relative border border-neutral-700/40">
                {/* Linha indicadora da Meta Alvo */}
                <div
                  className="absolute top-0 bottom-0 w-0.5 bg-white z-10 shadow-[0_0_8px_rgba(255,255,255,0.8)]"
                  style={{ left: `${Math.min(alvoPct, 100)}%` }}
                  title={`Meta: ${alvoPct}%`}
                />
                {/* Barra do Percentual Atual */}
                <div
                  className={`h-full ${config.corBarra} transition-all duration-500 rounded-full`}
                  style={{ width: `${Math.min(atualPct, 100)}%` }}
                />
              </div>

              <div className="mt-2 flex items-center justify-between text-[11px] font-mono text-neutral-500">
                <span>
                  Desvio:{" "}
                  <strong className={deltaPct >= 0 ? "text-amber-400" : "text-cyan-400"}>
                    {deltaPct >= 0 ? "+" : ""}
                    {deltaPct.toFixed(1)}%
                  </strong>
                </span>
                <span>
                  Meta Financeira: {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(valorAlvoClasse)}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
