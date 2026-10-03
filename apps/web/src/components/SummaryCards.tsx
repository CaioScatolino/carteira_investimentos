import React from "react";
import { Ativo } from "@/types/market";
import { TrendingUp, Building2, Globe, Award } from "lucide-react";

interface SummaryProps {
  acoes: Ativo[];
  fiis: Ativo[];
  etfs: Ativo[];
  onSelectAtivo: (ativo: Ativo) => void;
}

export function SummaryCards({ acoes, fiis, etfs, onSelectAtivo }: SummaryProps) {
  const topAcao = acoes[0];
  const topFII = fiis[0];
  const topETF = etfs[0];

  const formatCurrency = (val?: number) => {
    if (val === undefined || isNaN(val)) return "R$ 0,00";
    return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(val);
  };

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
      {/* Top Ação */}
      {topAcao && (
        <div
          onClick={() => onSelectAtivo(topAcao)}
          className="p-5 rounded-2xl bg-gradient-to-br from-[#111117] to-[#0d0d12] border border-[#22222a] hover:border-[#d4af37]/50 transition-all duration-300 cursor-pointer shadow-lg group relative overflow-hidden"
        >
          <div className="absolute top-0 right-0 w-24 h-24 bg-[#d4af37]/5 rounded-full blur-2xl group-hover:bg-[#d4af37]/10 transition-colors pointer-events-none" />
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-neutral-400 uppercase tracking-wider flex items-center gap-1.5">
              <TrendingUp className="w-3.5 h-3.5 text-[#d4af37]" />
              Top Ação (Score Geral)
            </span>
            <span className="px-2 py-0.5 rounded text-[11px] font-mono font-bold bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/30">
              Score: {topAcao.score}
            </span>
          </div>
          <div className="flex items-baseline justify-between">
            <div>
              <span className="text-2xl font-black text-white font-mono group-hover:text-[#d4af37] transition-colors">
                {topAcao.ticker}
              </span>
              <p className="text-xs text-neutral-400 mt-0.5 truncate max-w-[150px]">{topAcao.nome}</p>
            </div>
            <div className="text-right">
              <span className="text-lg font-bold text-white tabular-numbers">{formatCurrency(topAcao.preco_atual)}</span>
              <p className="text-xs font-medium text-emerald-400">DY: {topAcao.dy.toFixed(1)}%</p>
            </div>
          </div>
        </div>
      )}

      {/* Top FII */}
      {topFII && (
        <div
          onClick={() => onSelectAtivo(topFII)}
          className="p-5 rounded-2xl bg-gradient-to-br from-[#111117] to-[#0d0d12] border border-[#22222a] hover:border-[#d4af37]/50 transition-all duration-300 cursor-pointer shadow-lg group relative overflow-hidden"
        >
          <div className="absolute top-0 right-0 w-24 h-24 bg-[#d4af37]/5 rounded-full blur-2xl group-hover:bg-[#d4af37]/10 transition-colors pointer-events-none" />
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-neutral-400 uppercase tracking-wider flex items-center gap-1.5">
              <Building2 className="w-3.5 h-3.5 text-[#d4af37]" />
              Top Fundo Imobiliário
            </span>
            <span className="px-2 py-0.5 rounded text-[11px] font-mono font-bold bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/30">
              Score: {topFII.score}
            </span>
          </div>
          <div className="flex items-baseline justify-between">
            <div>
              <span className="text-2xl font-black text-white font-mono group-hover:text-[#d4af37] transition-colors">
                {topFII.ticker}
              </span>
              <p className="text-xs text-neutral-400 mt-0.5 truncate max-w-[150px]">{topFII.nome}</p>
            </div>
            <div className="text-right">
              <span className="text-lg font-bold text-white tabular-numbers">{formatCurrency(topFII.preco_atual)}</span>
              <p className="text-xs font-medium text-emerald-400">P/VP: {topFII.pvp.toFixed(2)}x</p>
            </div>
          </div>
        </div>
      )}

      {/* Top ETF */}
      {topETF && (
        <div
          onClick={() => onSelectAtivo(topETF)}
          className="p-5 rounded-2xl bg-gradient-to-br from-[#111117] to-[#0d0d12] border border-[#22222a] hover:border-[#d4af37]/50 transition-all duration-300 cursor-pointer shadow-lg group relative overflow-hidden"
        >
          <div className="absolute top-0 right-0 w-24 h-24 bg-[#d4af37]/5 rounded-full blur-2xl group-hover:bg-[#d4af37]/10 transition-colors pointer-events-none" />
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs font-semibold text-neutral-400 uppercase tracking-wider flex items-center gap-1.5">
              <Globe className="w-3.5 h-3.5 text-[#d4af37]" />
              Maior Liquidez (ETF)
            </span>
            <span className="px-2 py-0.5 rounded text-[11px] font-mono font-bold bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/30">
              Score: {topETF.score}
            </span>
          </div>
          <div className="flex items-baseline justify-between">
            <div>
              <span className="text-2xl font-black text-white font-mono group-hover:text-[#d4af37] transition-colors">
                {topETF.ticker}
              </span>
              <p className="text-xs text-neutral-400 mt-0.5 truncate max-w-[150px]">{topETF.nome}</p>
            </div>
            <div className="text-right">
              <span className="text-lg font-bold text-white tabular-numbers">{formatCurrency(topETF.preco_atual)}</span>
              <p className="text-xs font-medium text-[#d4af37]">R$ {(topETF.volume_total / 1_000_000).toFixed(0)}M/dia</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
