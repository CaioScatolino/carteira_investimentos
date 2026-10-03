"use client";

import React from "react";
import { Ativo } from "@/types/market";
import { StatusBadge, ParecerBadge } from "./StatusBadge";
import { X, ExternalLink, ShieldCheck, TrendingUp, DollarSign, PieChart, Landmark } from "lucide-react";

interface AssetModalProps {
  ativo: Ativo | null;
  onClose: () => void;
}

export function AssetDetailModal({ ativo, onClose }: AssetModalProps) {
  if (!ativo) return null;

  const formatCurrency = (val?: number) => {
    if (val === undefined || isNaN(val)) return "R$ 0,00";
    return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(val);
  };

  const formatPercent = (val?: number) => {
    if (val === undefined || isNaN(val)) return "0,0%";
    return `${val.toFixed(1)}%`;
  };

  const formatVolume = (val?: number) => {
    if (!val) return "R$ 0";
    if (val >= 1_000_000_000) return `R$ ${(val / 1_000_000_000).toFixed(2)}B`;
    if (val >= 1_000_000) return `R$ ${(val / 1_000_000).toFixed(2)}M`;
    if (val >= 1_000) return `R$ ${(val / 1_000).toFixed(1)}K`;
    return `R$ ${val.toFixed(0)}`;
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div 
        className="relative w-full max-w-3xl max-h-[90vh] overflow-y-auto rounded-2xl bg-[#0d0d12] border border-[#d4af37]/30 shadow-2xl p-6 md:p-8 gold-glow"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Botão Fechar */}
        <button
          onClick={onClose}
          className="absolute top-5 right-5 p-2 rounded-full text-neutral-400 hover:text-[#d4af37] hover:bg-neutral-900 transition-colors"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Cabeçalho do Ativo */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-neutral-800/80 pb-6">
          <div>
            <div className="flex items-center gap-3">
              <span className="text-3xl font-extrabold tracking-tight text-white font-mono">{ativo.ticker}</span>
              <span className="text-xs uppercase px-2 py-0.5 rounded bg-neutral-900 text-[#d4af37] border border-[#d4af37]/30 font-semibold tracking-wider">
                {ativo.classe}
              </span>
              <StatusBadge status={ativo.status} />
            </div>
            <p className="text-sm text-neutral-400 mt-1 font-medium">{ativo.nome}</p>
            {ativo.cnpj && (
              <p className="text-xs text-neutral-500 font-mono mt-0.5">CNPJ: {ativo.cnpj}</p>
            )}
          </div>

          <div className="flex flex-col md:items-end">
            <span className="text-xs text-neutral-500 uppercase tracking-widest">Cotação Atual (B3)</span>
            <span className="text-3xl font-bold text-white tabular-numbers">{formatCurrency(ativo.preco_atual)}</span>
            <span className="text-xs text-neutral-400 mt-0.5">Volume Diário: {formatVolume(ativo.volume_total)}</span>
          </div>
        </div>

        {/* Score Fundamentalista */}
        <div className="my-6 p-4 rounded-xl bg-gradient-to-r from-[#14141c] to-[#121217] border border-[#d4af37]/20 flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-xl bg-[#d4af37]/10 border border-[#d4af37]/30 flex items-center justify-center text-[#d4af37]">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div>
              <h4 className="text-sm font-semibold text-neutral-200">Score Fundamentalista Geral</h4>
              <p className="text-xs text-neutral-400">Pontuação multi-critério consolidada baseada em CVM e B3</p>
            </div>
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-4xl font-extrabold text-[#d4af37] font-mono tabular-numbers">{ativo.score}</span>
            <span className="text-sm text-neutral-500 font-medium">/ 100</span>
          </div>
        </div>

        {/* Indicadores Principais */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 my-6">
          <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
            <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">Dividend Yield 12M</span>
            <span className="text-lg font-bold text-emerald-400 tabular-numbers">{formatPercent(ativo.dy)}</span>
            <span className="text-[11px] text-neutral-500 block mt-0.5">{formatCurrency(ativo.dividendos_12m)}/cota</span>
          </div>

          {ativo.classe === "FII" ? (
            <>
              <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">P/VP CVM</span>
                <span className={`text-lg font-bold tabular-numbers ${ativo.pvp <= 1.0 ? "text-emerald-400" : "text-amber-400"}`}>
                  {ativo.pvp.toFixed(2)}x
                </span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">VP: {formatCurrency(ativo.vp_cota)}</span>
              </div>
              <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">Spread vs NTN-B</span>
                <span className={`text-lg font-bold tabular-numbers ${ativo.spread_ntnb >= 2.0 ? "text-emerald-400" : "text-amber-400"}`}>
                  {ativo.spread_ntnb >= 0 ? `+${ativo.spread_ntnb.toFixed(2)}%` : `${ativo.spread_ntnb.toFixed(2)}%`}
                </span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">Prêmio s/ IPCA+ 6.5%</span>
              </div>
              <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">Teto Bazin (6%)</span>
                <span className="text-lg font-bold text-[#d4af37] tabular-numbers">{formatCurrency(ativo.preco_teto_bazin)}</span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">Margem: {ativo.margem_bazin.toFixed(1)}%</span>
              </div>
            </>
          ) : ativo.classe === "ACAO" ? (
            <>
              <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">P/L (Preço/Lucro)</span>
                <span className="text-lg font-bold text-neutral-200 tabular-numbers">{ativo.pl.toFixed(1)}x</span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">LPA: {formatCurrency(ativo.lpa)}</span>
              </div>
              <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">P/VP & ROE</span>
                <span className="text-lg font-bold text-neutral-200 tabular-numbers">{ativo.pvp_real.toFixed(2)}x</span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">ROE: {formatPercent(ativo.roe)}</span>
              </div>
              <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">Earnings Yield</span>
                <span className="text-lg font-bold text-emerald-400 tabular-numbers">{formatPercent(ativo.earnings_yield)}</span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">Lucro s/ Cotação</span>
              </div>
            </>
          ) : (
            <>
              <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800 col-span-3">
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">Liquidez Diária Institucional</span>
                <span className="text-lg font-bold text-emerald-400 tabular-numbers">{formatVolume(ativo.volume_total)}</span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">Negociação direta no pregão da B3</span>
              </div>
            </>
          )}
        </div>

        {/* Métricas Operacionais e Governança StatusInvest */}
        {(ativo.setor || ativo.segmento || ativo.roic !== undefined) && (
          <div className="mb-6 p-4 rounded-xl bg-neutral-900/30 border border-neutral-800/80">
            <div className="flex items-center justify-between mb-3 border-b border-neutral-800 pb-2">
              <span className="text-xs font-semibold text-neutral-400 uppercase tracking-wider">
                Métricas Operacionais & Setoriais (StatusInvest)
              </span>
              {ativo.setor && (
                <span className="text-[11px] text-[#d4af37] bg-[#d4af37]/10 px-2 py-0.5 rounded border border-[#d4af37]/20 font-medium">
                  {ativo.setor} {ativo.segmento ? `• ${ativo.segmento}` : ""}
                </span>
              )}
            </div>

            {ativo.classe === "ACAO" ? (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5 text-xs">
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">ROIC</span>
                  <span className="font-semibold text-neutral-200">{formatPercent(ativo.roic)}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">ROA</span>
                  <span className="font-semibold text-neutral-200">{formatPercent(ativo.roa)}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Margem Líquida</span>
                  <span className="font-semibold text-emerald-400">{formatPercent(ativo.margem_liquida)}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Margem Bruta</span>
                  <span className="font-semibold text-neutral-200">{formatPercent(ativo.margem_bruta)}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Dív. Líquida / PL</span>
                  <span className="font-semibold text-neutral-200">{ativo.divida_liquida_pl !== undefined ? `${ativo.divida_liquida_pl.toFixed(2)}x` : "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Dív. Líquida / EBIT</span>
                  <span className="font-semibold text-neutral-200">{ativo.divida_liquida_ebit !== undefined ? `${ativo.divida_liquida_ebit.toFixed(2)}x` : "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Liq. Corrente</span>
                  <span className="font-semibold text-neutral-200">{ativo.liquidez_corrente !== undefined ? `${ativo.liquidez_corrente.toFixed(2)}x` : "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Liq. Média Diária</span>
                  <span className="font-semibold text-[#d4af37]">{formatVolume(ativo.liquidez_media_diaria)}</span>
                </div>
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5 text-xs">
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Segmento FII</span>
                  <span className="font-semibold text-neutral-200">{ativo.segmento || "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Gestão</span>
                  <span className="font-semibold text-neutral-200">{ativo.gestao || "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Cotistas</span>
                  <span className="font-semibold text-neutral-200">{ativo.numero_cotistas ? ativo.numero_cotistas.toLocaleString("pt-BR") : "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Último Provento</span>
                  <span className="font-semibold text-emerald-400">{formatCurrency(ativo.last_dividend)}</span>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Pareceres Individuais das Escolas de Investimento */}
        {ativo.pareceres && Object.keys(ativo.pareceres).length > 0 && (
          <div className="mt-6">
            <h4 className="text-xs font-semibold text-neutral-400 uppercase tracking-wider mb-3">
              Pareceres Detalhados por Escola de Valuation
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {Object.entries(ativo.pareceres).map(([escola, p]) => (
                <div key={escola} className="p-3.5 rounded-xl bg-neutral-900/40 border border-neutral-800/90 flex flex-col justify-between">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-bold text-neutral-200">{escola}</span>
                    <ParecerBadge status={p.status} />
                  </div>
                  <div className="space-y-0.5">
                    <p className="text-xs font-mono text-[#d4af37]">{p.metrica}</p>
                    <p className="text-xs text-neutral-400">{p.detalhe}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Rodapé do Modal */}
        <div className="mt-8 pt-4 border-t border-neutral-800/80 flex items-center justify-between text-xs text-neutral-500">
          <span>Dados oficiais CVM e B3 • Sem atraso de cálculo</span>
          <button 
            onClick={onClose}
            className="px-4 py-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 text-neutral-200 transition-colors"
          >
            Fechar
          </button>
        </div>
      </div>
    </div>
  );
}
