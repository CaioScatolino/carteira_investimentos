"use client";

import React from "react";
import { ResumoCarteira } from "@/types/portfolio";
import { DollarSign, TrendingUp, TrendingDown, PiggyBank, Briefcase } from "lucide-react";

interface PortfolioSummaryCardsProps {
  resumo: ResumoCarteira;
}

export function PortfolioSummaryCards({ resumo }: PortfolioSummaryCardsProps) {
  const isPositivo = resumo.lucro_total_valor >= 0;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      {/* 1. Patrimônio Total */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-[#d4af37]/40 transition-all">
        <div className="absolute top-0 right-0 w-24 h-24 bg-[#d4af37]/5 rounded-bl-full pointer-events-none group-hover:bg-[#d4af37]/10 transition-colors" />
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
            Patrimônio Total
          </span>
          <div className="p-2 rounded-xl bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/25">
            <DollarSign className="w-4 h-4" />
          </div>
        </div>
        <div className="text-2xl font-black text-white font-mono tracking-tight">
          {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(resumo.patrimonio_total)}
        </div>
        <div className="mt-2 flex items-center gap-1.5 text-xs">
          <span className="text-neutral-500 font-mono">Total consolidado a mercado</span>
        </div>
      </div>

      {/* 2. Total Investido */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-neutral-700 transition-all">
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
            Total Investido
          </span>
          <div className="p-2 rounded-xl bg-neutral-800 text-neutral-300 border border-neutral-700">
            <PiggyBank className="w-4 h-4" />
          </div>
        </div>
        <div className="text-2xl font-black text-neutral-200 font-mono tracking-tight">
          {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(resumo.total_investido)}
        </div>
        <div className="mt-2 flex items-center gap-1.5 text-xs text-neutral-500 font-mono">
          <span>Custo histórico de aquisição</span>
        </div>
      </div>

      {/* 3. Lucro / Prejuízo Nominal */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-neutral-700 transition-all">
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-neutral-400 font-mono">
            Lucro / Prejuízo
          </span>
          <div
            className={`p-2 rounded-xl border ${
              isPositivo
                ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/25"
                : "bg-red-500/10 text-red-400 border-red-500/25"
            }`}
          >
            {isPositivo ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
          </div>
        </div>
        <div
          className={`text-2xl font-black font-mono tracking-tight ${
            isPositivo ? "text-emerald-400" : "text-red-400"
          }`}
        >
          {isPositivo ? "+" : ""}
          {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(resumo.lucro_total_valor)}
        </div>
        <div className="mt-2 flex items-center gap-1.5 text-xs font-mono">
          <span
            className={`font-bold ${isPositivo ? "text-emerald-400" : "text-red-400"}`}
          >
            {isPositivo ? "+" : ""}
            {resumo.rentabilidade_pct.toFixed(2)}%
          </span>
          <span className="text-neutral-500">rentabilidade total</span>
        </div>
      </div>

      {/* 4. Total de Posições */}
      <div className="p-5 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden group hover:border-neutral-700 transition-all">
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
        <div className="mt-2 flex items-center gap-1.5 text-xs text-neutral-500 font-mono">
          <span>{Object.keys(resumo.alocacao_atual_pct).length} macro-classes ativas</span>
        </div>
      </div>
    </div>
  );
}
