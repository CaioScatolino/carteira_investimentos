"use client";

import React from "react";
import { AnaliseConcentracaoHHI, SetorConsolidado, ClasseAtivo } from "@/types/portfolio";
import {
  ShieldCheck,
  Building2,
  TrendingUp,
  Globe,
  Coins,
  AlertTriangle,
  CheckCircle2,
  Layers,
  Activity,
  Info,
} from "lucide-react";

interface SectorConcentrationViewProps {
  analise?: AnaliseConcentracaoHHI;
  patrimonioTotal: number;
  ocultarValores?: boolean;
}

export function SectorConcentrationView({
  analise,
  patrimonioTotal,
  ocultarValores = false,
}: SectorConcentrationViewProps) {
  if (!analise || analise.setores.length === 0) {
    return (
      <div className="p-6 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 text-center font-mono text-neutral-400">
        Nenhum ativo ou setor computado para a análise de concentração.
      </div>
    );
  }

  const formatarMoeda = (val: number) =>
    new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(val);

  const renderMoeda = (val: number) => (
    <span
      className={
        ocultarValores
          ? "filter blur-[6px] select-none pointer-events-none transition-all duration-300 inline-block"
          : "transition-all duration-300"
      }
    >
      {formatarMoeda(val)}
    </span>
  );

  const obterIconeClasse = (classe: ClasseAtivo) => {
    switch (classe) {
      case "RENDA_FIXA":
        return <ShieldCheck className="w-3.5 h-3.5 text-blue-400" />;
      case "FII":
        return <Building2 className="w-3.5 h-3.5 text-purple-400" />;
      case "ACAO_BR":
        return <TrendingUp className="w-3.5 h-3.5 text-emerald-400" />;
      case "INTERNACIONAL":
        return <Globe className="w-3.5 h-3.5 text-cyan-400" />;
      default:
        return <Coins className="w-3.5 h-3.5 text-amber-400" />;
    }
  };

  const badgeHHI = (classificacao: string, tipo: "ATIVOS" | "SETORES") => {
    switch (classificacao) {
      case "BEM_DIVERSIFICADO":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
            {tipo === "ATIVOS" ? "Bem Diversificado (< 1.500)" : "Bem Descorrelacionado (< 1.800)"}
          </span>
        );
      case "MODERADO":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
            Concentração Moderada
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-rose-400" />
            {tipo === "ATIVOS" ? "Alta Concentração (≥ 2.500)" : "Risco Monocíclico (≥ 3.000)"}
          </span>
        );
    }
  };

  const temAlertaCritico = analise.alertas.some((a) => a.includes("🚨"));
  const temAlertaModerado = analise.alertas.some((a) => a.includes("⚠️"));

  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      
      {/* 1. Trinca de Cards HHI & Blindagem */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 font-mono">
        
        {/* Card 1: HHI Ativos */}
        <div className="p-5 rounded-2xl bg-neutral-950/90 border border-neutral-800 shadow-xl space-y-3 relative overflow-hidden">
          <div className="flex items-center justify-between text-neutral-400">
            <div className="flex items-center gap-2">
              <Layers className="w-4 h-4 text-[#d4af37]" />
              <span className="text-xs uppercase tracking-wider font-bold text-neutral-300">
                HHI Emissores (Ativos)
              </span>
            </div>
            {badgeHHI(analise.classificacao_ativos, "ATIVOS")}
          </div>

          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-extrabold text-white tracking-tight">
              {analise.hhi_ativos.toLocaleString("pt-BR", { minimumFractionDigits: 1 })}
            </span>
            <span className="text-xs text-neutral-500">/ 10.000 pts</span>
          </div>

          <div className="pt-1 border-t border-neutral-900 space-y-1">
            <div className="text-[11px] text-neutral-400 flex items-center justify-between">
              <span>Maior Posição:</span>
              <span className="text-white font-bold">
                {analise.maior_ativo_ticker || "-"} ({analise.maior_ativo_pct.toFixed(1)}%)
              </span>
            </div>
            <p className="text-[10px] text-neutral-500 leading-relaxed">
              {analise.descricao_hhi_ativos}
            </p>
          </div>
        </div>

        {/* Card 2: HHI Setores */}
        <div className="p-5 rounded-2xl bg-neutral-950/90 border border-neutral-800 shadow-xl space-y-3 relative overflow-hidden">
          <div className="flex items-center justify-between text-neutral-400">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-cyan-400" />
              <span className="text-xs uppercase tracking-wider font-bold text-neutral-300">
                HHI Setorial (Ciclos)
              </span>
            </div>
            {badgeHHI(analise.classificacao_setores, "SETORES")}
          </div>

          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-extrabold text-white tracking-tight">
              {analise.hhi_setores.toLocaleString("pt-BR", { minimumFractionDigits: 1 })}
            </span>
            <span className="text-xs text-neutral-500">/ 10.000 pts</span>
          </div>

          <div className="pt-1 border-t border-neutral-900 space-y-1">
            <div className="text-[11px] text-neutral-400 flex items-center justify-between">
              <span>Maior Setor:</span>
              <span className="text-white font-bold truncate max-w-[150px]">
                {analise.maior_setor_nome || "-"} ({analise.maior_setor_pct.toFixed(1)}%)
              </span>
            </div>
            <p className="text-[10px] text-neutral-500 leading-relaxed">
              {analise.descricao_hhi_setores}
            </p>
          </div>
        </div>

        {/* Card 3: Auditoria de Blindagem & Alertas */}
        <div
          className={`p-5 rounded-2xl bg-neutral-950/90 border shadow-xl space-y-3 relative overflow-hidden ${
            temAlertaCritico
              ? "border-rose-500/40"
              : temAlertaModerado
              ? "border-amber-500/40"
              : "border-emerald-500/30"
          }`}
        >
          <div className="flex items-center justify-between text-neutral-400">
            <div className="flex items-center gap-2">
              {temAlertaCritico ? (
                <AlertTriangle className="w-4 h-4 text-rose-400" />
              ) : temAlertaModerado ? (
                <AlertTriangle className="w-4 h-4 text-amber-400" />
              ) : (
                <CheckCircle2 className="w-4 h-4 text-emerald-400" />
              )}
              <span className="text-xs uppercase tracking-wider font-bold text-neutral-300">
                Status de Blindagem
              </span>
            </div>
            <span
              className={`text-[10px] px-2 py-0.5 rounded-full font-bold ${
                temAlertaCritico
                  ? "bg-rose-500/10 text-rose-300 border border-rose-500/30"
                  : temAlertaModerado
                  ? "bg-amber-500/10 text-amber-300 border border-amber-500/30"
                  : "bg-emerald-500/10 text-emerald-300 border border-emerald-500/30"
              }`}
            >
              {temAlertaCritico
                ? "Risco Específico"
                : temAlertaModerado
                ? "Atenção Prudencial"
                : "Blindada"}
            </span>
          </div>

          <div className="space-y-1.5 max-h-24 overflow-y-auto custom-scrollbar pr-1 pt-1">
            {analise.alertas.map((alerta, idx) => (
              <div
                key={idx}
                className="text-[11px] leading-tight text-neutral-300 flex items-start gap-1.5"
              >
                <span>{alerta}</span>
              </div>
            ))}
          </div>

          <div className="pt-1 border-t border-neutral-900 text-[10px] text-neutral-500 flex items-center gap-1">
            <Info className="w-3 h-3 text-[#d4af37] shrink-0" />
            <span>Teto sugerido: máx 15% por ativo e 30% por setor econômico.</span>
          </div>
        </div>

      </div>

      {/* 2. Decomposição Setorial B3 (Barras de Distribuição) */}
      <div className="p-6 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-2xl backdrop-blur-sm space-y-5">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-neutral-800/60 pb-4">
          <div>
            <h3 className="text-base font-bold text-white flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-cyan-400" />
              Decomposição & Concentração por Setores Oficiais B3
            </h3>
            <p className="text-xs text-neutral-400 mt-1">
              Mapeamento do capital alocado por segmentos econômicos e sensibilidade a ciclos de mercado.
            </p>
          </div>
          <div className="text-xs font-mono text-neutral-400 self-start sm:self-auto bg-neutral-900/80 px-3 py-1.5 rounded-lg border border-neutral-800">
            Total Setores Ativos: <span className="text-white font-bold">{analise.setores.length}</span>
          </div>
        </div>

        {/* Lista de Setores */}
        <div className="space-y-4">
          {analise.setores.map((setor, idx) => {
            const isSobre = setor.percentual_na_carteira > 30.0;
            const isMod = setor.percentual_na_carteira > 20.0 && !isSobre;

            let corBarra = "bg-gradient-to-r from-emerald-500 to-cyan-500";
            if (isSobre) {
              corBarra = "bg-gradient-to-r from-rose-500 to-amber-500";
            } else if (isMod) {
              corBarra = "bg-gradient-to-r from-amber-500 to-[#d4af37]";
            }

            return (
              <div
                key={idx}
                className="p-4 rounded-xl bg-neutral-950/70 border border-neutral-800/70 hover:border-neutral-700/80 transition-all space-y-2.5"
              >
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs font-mono">
                  <div className="flex items-center gap-2.5">
                    <span className="w-6 h-6 rounded-lg bg-neutral-900 border border-neutral-800 flex items-center justify-center shrink-0">
                      {obterIconeClasse(setor.classe_principal)}
                    </span>
                    <div>
                      <span className="font-extrabold text-white text-sm block">
                        {setor.nome}
                      </span>
                      <span className="text-[10px] text-neutral-500">
                        {setor.quantidade_ativos} ativo(s) em carteira
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 self-end sm:self-auto">
                    {/* Badge de Concentração */}
                    {isSobre ? (
                      <span className="text-[10px] font-bold px-2 py-0.5 rounded-md bg-rose-500/10 text-rose-400 border border-rose-500/30">
                        Sobreconcentrado (&gt; 30%)
                      </span>
                    ) : isMod ? (
                      <span className="text-[10px] font-bold px-2 py-0.5 rounded-md bg-amber-500/10 text-amber-400 border border-amber-500/30">
                        Atenção (&gt; 20%)
                      </span>
                    ) : (
                      <span className="text-[10px] font-bold px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                        Equilibrado
                      </span>
                    )}

                    <div className="text-right">
                      <span className="font-bold text-white text-sm block">
                        {renderMoeda(setor.valor_atual_mercado)}
                      </span>
                      <span className="text-[11px] text-[#d4af37] font-extrabold">
                        {setor.percentual_na_carteira.toFixed(1)}% da carteira
                      </span>
                    </div>
                  </div>
                </div>

                {/* Barra de Progresso do Setor */}
                <div className="h-2 w-full bg-neutral-900 rounded-full overflow-hidden p-0.5 border border-neutral-800/40">
                  <div
                    className={`h-full rounded-full transition-all duration-700 ${corBarra}`}
                    style={{ width: `${Math.min(setor.percentual_na_carteira, 100)}%` }}
                  />
                </div>

                {/* Tickers participantes */}
                <div className="flex flex-wrap items-center gap-1.5 pt-1">
                  <span className="text-[10px] text-neutral-500 font-mono">Ativos:</span>
                  {setor.tickers.map((tk, tIdx) => (
                    <span
                      key={tIdx}
                      className="px-2 py-0.5 rounded bg-neutral-900 border border-neutral-800 text-[10px] font-bold text-neutral-300 font-mono"
                    >
                      {tk}
                    </span>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>

    </div>
  );
}

