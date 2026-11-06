"use client";

import React from "react";
import { ResumoCarteira } from "@/types/portfolio";
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  PiggyBank,
  Briefcase,
  Coins,
} from "lucide-react";

interface PortfolioSummaryCardsProps {
  resumo: ResumoCarteira;
  ocultarValores?: boolean;
}

export function PortfolioSummaryCards({ resumo, ocultarValores = false }: PortfolioSummaryCardsProps) {
  const formatarMoeda = (val: number) =>
    new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(val);

  const renderMoeda = (val: number, prefixo = "") => {
    const texto = `${prefixo}${formatarMoeda(val)}`;
    return (
      <span
        className={
          ocultarValores
            ? "filter blur-[6px] select-none pointer-events-none transition-all duration-300 inline-block"
            : "transition-all duration-300"
        }
      >
        {texto}
      </span>
    );
  };

  // 1. Proventos totais acumulados
  const totalProventos =
    resumo.total_proventos_recebidos ??
    resumo.posicoes.reduce((acc, p) => acc + (p.total_proventos ?? 0), 0);

  // 2. Valorização de ativos pura (cotação de mercado - custo investido)
  const valorizacaoAtivos = resumo.lucro_total_valor;
  const rentabilidadeAtivosPct = resumo.rentabilidade_pct;

  // 3. Resultado Global da Carteira (Valorização + Proventos Recebidos)
  const resultadoGlobal =
    resumo.lucro_total_com_proventos ?? (valorizacaoAtivos + totalProventos);

  const rentabilidadeTotalPct =
    resumo.rentabilidade_total_pct ??
    (resumo.total_investido > 0
      ? (resultadoGlobal / resumo.total_investido) * 100
      : 0);

  const yieldOnCostGlobal =
    resumo.total_investido > 0
      ? (totalProventos / resumo.total_investido) * 100
      : 0;

  const isResultadoPositivo = resultadoGlobal >= 0;
  const isValorizacaoPositiva = valorizacaoAtivos >= 0;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4 mb-8">
      {/* 1. Patrimônio Total */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-[#d4af37]/40 transition-all flex flex-col justify-between">
        <div className="absolute top-0 right-0 w-24 h-24 bg-[#d4af37]/5 rounded-bl-full pointer-events-none group-hover:bg-[#d4af37]/10 transition-colors" />
        <div>
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
              Patrimônio Total
            </span>
            <div className="p-2 rounded-xl bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/25">
              <DollarSign className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-black text-white font-mono tracking-tight">
            {renderMoeda(resumo.patrimonio_total)}
          </div>
        </div>
        <div className="mt-3 flex items-center gap-1.5 text-xs text-neutral-500 font-mono">
          <span>Total consolidado a mercado</span>
        </div>
      </div>

      {/* 2. Total Investido */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-neutral-700 transition-all flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
              Total Investido
            </span>
            <div className="p-2 rounded-xl bg-neutral-800 text-neutral-300 border border-neutral-700">
              <PiggyBank className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-black text-neutral-200 font-mono tracking-tight">
            {renderMoeda(resumo.total_investido)}
          </div>
        </div>
        <div className="mt-3 flex items-center gap-1.5 text-xs text-neutral-500 font-mono">
          <span>Custo histórico de aquisição</span>
        </div>
      </div>

      {/* 3. Proventos Recebidos (Card Dedicado) */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-emerald-500/40 transition-all flex flex-col justify-between">
        <div className="absolute top-0 right-0 w-24 h-24 bg-emerald-500/5 rounded-bl-full pointer-events-none group-hover:bg-emerald-500/10 transition-colors" />
        <div>
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
              Proventos Recebidos
            </span>
            <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/25">
              <Coins className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-black text-emerald-400 font-mono tracking-tight">
            {renderMoeda(totalProventos, "+")}
          </div>
        </div>
        <div className="mt-3 flex items-center justify-between text-xs font-mono">
          <span className="text-emerald-400/90 font-bold">
            +{yieldOnCostGlobal.toFixed(2)}% YoC
          </span>
          <span className="text-neutral-500">renda acumulada</span>
        </div>
      </div>

      {/* 4. Resultado Global (Somatória de Valorização + Proventos com Detalhamento) */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-neutral-700 transition-all flex flex-col justify-between">
        <div
          className={`absolute top-0 right-0 w-24 h-24 rounded-bl-full pointer-events-none transition-colors ${
            isResultadoPositivo ? "bg-emerald-500/5 group-hover:bg-emerald-500/10" : "bg-red-500/5 group-hover:bg-red-500/10"
          }`}
        />
        <div>
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
              Resultado Global
            </span>
            <div
              className={`p-2 rounded-xl border ${
                isResultadoPositivo
                  ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/25"
                  : "bg-red-500/10 text-red-400 border-red-500/25"
              }`}
            >
              {isResultadoPositivo ? (
                <TrendingUp className="w-4 h-4" />
              ) : (
                <TrendingDown className="w-4 h-4" />
              )}
            </div>
          </div>
          <div
            className={`text-2xl font-black font-mono tracking-tight ${
              isResultadoPositivo ? "text-emerald-400" : "text-red-400"
            }`}
          >
            {renderMoeda(resultadoGlobal, isResultadoPositivo ? "+" : "")}
          </div>
          <div className="mt-1 flex items-center gap-1.5 text-xs font-mono">
            <span
              className={`font-bold ${
                isResultadoPositivo ? "text-emerald-400" : "text-red-400"
              }`}
            >
              {isResultadoPositivo ? "+" : ""}
              {rentabilidadeTotalPct.toFixed(2)}%
            </span>
            <span className="text-neutral-500">rentabilidade total</span>
          </div>
        </div>

        {/* Detalhamento: Valorização de Ativos vs Proventos Recebidos */}
        <div className="mt-3 pt-2.5 border-t border-neutral-800/70 flex items-center justify-between text-[11px] font-mono text-neutral-400">
          <span title="Ganho/Perda pura de cotação dos ativos">
            Ativos:{" "}
            <strong className={isValorizacaoPositiva ? "text-emerald-400" : "text-red-400"}>
              {renderMoeda(valorizacaoAtivos, isValorizacaoPositiva ? "+" : "")}
            </strong>
          </span>
          <span className="text-neutral-600">•</span>
          <span title="Total de proventos (dividendos e JCP) recebidos">
            Prov.:{" "}
            <strong className="text-emerald-400">
              {renderMoeda(totalProventos, "+")}
            </strong>
          </span>
        </div>
      </div>

      {/* 5. Total de Ativos em Custódia */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-neutral-700 transition-all flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
              Ativos em Custódia
            </span>
            <div className="p-2 rounded-xl bg-purple-500/10 text-purple-400 border border-purple-500/25">
              <Briefcase className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-black text-white font-mono tracking-tight">
            {resumo.posicoes.length} <span className="text-sm font-normal text-neutral-400">ativos</span>
          </div>
        </div>
        <div className="mt-3 flex items-center gap-1.5 text-xs text-neutral-500 font-mono">
          <span>{Object.keys(resumo.alocacao_atual_pct).length} macro-classes ativas</span>
        </div>
      </div>
    </div>
  );
}
