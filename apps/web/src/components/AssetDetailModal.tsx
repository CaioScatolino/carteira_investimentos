"use client";

import React, { useState, useEffect } from "react";
import { Ativo, AnaliseProventos5A, CenarioMacro } from "@/types/market";
import { StatusBadge, ParecerBadge } from "./StatusBadge";
import { X, ExternalLink, ShieldCheck, TrendingUp, DollarSign, PieChart, Landmark, BarChart3, Coins, AlertCircle, Target, Scale } from "lucide-react";

interface AssetModalProps {
  ativo: Ativo | null;
  macro?: CenarioMacro | null;
  onClose: () => void;
}

export function AssetDetailModal({ ativo, macro, onClose }: AssetModalProps) {
  if (!ativo) return null;

  const [analiseB3, setAnaliseB3] = useState<AnaliseProventos5A | null>(null);
  const [carregandoB3, setCarregandoB3] = useState<boolean>(false);

  useEffect(() => {
    let cancelado = false;
    if (ativo && (ativo.classe === "ACAO" || ativo.classe === "FII")) {
      setCarregandoB3(true);
      fetch(`http://localhost:8080/api/v1/ativos/${ativo.ticker}/dividendos`)
        .then((res) => {
          if (!res.ok) throw new Error("Erro HTTP");
          return res.json();
        })
        .then((data: AnaliseProventos5A) => {
          if (!cancelado) {
            setAnaliseB3(data);
          }
        })
        .catch(() => {
          // Mantém dados existentes se falhar
        })
        .finally(() => {
          if (!cancelado) setCarregandoB3(false);
        });
    } else {
      setAnaliseB3(null);
    }

    return () => {
      cancelado = true;
    };
  }, [ativo?.ticker, ativo?.classe]);

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

  // Prepara histórico anual para o gráfico
  const mapaHistorico = analiseB3?.proventos_por_ano || ativo.historico_dividendos_anual || {};
  const entradasAnos = Object.entries(mapaHistorico).sort(([a], [b]) => a.localeCompare(b));
  const valoresAno = entradasAnos.map(([, v]) => v);
  const maxValorAno = Math.max(...valoresAno, 1.0);

  const teto5A = analiseB3?.preco_teto_bazin_5a || ativo.preco_teto_bazin_5a || ativo.preco_teto_bazin;
  const margem5A = analiseB3?.margem_bazin_5a ?? ativo.margem_bazin_5a ?? ativo.margem_bazin;
  const media5A = analiseB3?.media_dividendos_5a || ativo.media_dividendos_5a || ativo.dividendos_12m;
  const media5ANorm = analiseB3?.media_dividendos_5a_normalizada || ativo.media_dividendos_5a_normalizada || media5A;
  const teveOutlier = analiseB3?.teve_outlier_5a ?? ativo.teve_outlier_5a ?? false;
  const obsOutlier = analiseB3?.observacao_outlier || ativo.observacao_outlier || "";

  const tetoConsolidado = ativo.preco_teto_consolidado && ativo.preco_teto_consolidado > 0 ? ativo.preco_teto_consolidado : teto5A;
  const premioPct = ativo.premio_desconto_percentual ?? (((tetoConsolidado - ativo.preco_atual) / (ativo.preco_atual || 1)) * 100);

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
            <div className="flex items-center gap-2.5 flex-wrap">
              <span className="text-3xl font-extrabold tracking-tight text-white font-mono">{ativo.ticker}</span>
              <span className="text-xs uppercase px-2 py-0.5 rounded bg-neutral-900 text-[#d4af37] border border-[#d4af37]/30 font-semibold tracking-wider">
                {ativo.classe}
              </span>
              {ativo.porte === "BLUE_CHIP" && (
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-[#d4af37]/15 text-[#d4af37] border border-[#d4af37]/40 font-bold flex items-center gap-1 shadow-sm">
                  👑 Blue Chip
                </span>
              )}
              {ativo.porte === "FII_GIGANTE" && (
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-[#d4af37]/15 text-[#d4af37] border border-[#d4af37]/40 font-bold flex items-center gap-1 shadow-sm">
                  🏰 FII Baleia
                </span>
              )}
              {ativo.porte === "MID_CAP" && (
                <span className="text-xs px-2 py-0.5 rounded-full bg-neutral-800 text-neutral-300 border border-neutral-700 font-medium">
                  Mid Cap
                </span>
              )}
              {ativo.porte === "SMALL_CAP" && (
                <span className="text-xs px-2 py-0.5 rounded-full bg-neutral-900 text-neutral-400 border border-neutral-800 font-medium">
                  Small Cap
                </span>
              )}
              {ativo.porte === "MICRO_CAP" && (
                <span className="text-xs px-2 py-0.5 rounded-full bg-red-500/10 text-red-400 border border-red-500/20 font-medium">
                  Microcap
                </span>
              )}
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
            <span className="text-xs text-neutral-400 mt-0.5">
              {ativo.valor_mercado ? `MktCap: ${formatVolume(ativo.valor_mercado)}` : `Cotistas: ${ativo.numero_cotistas?.toLocaleString("pt-BR")}`} • Vol: {formatVolume(ativo.volume_total)}
            </span>
          </div>
        </div>

        {/* Alerta de Risco para Proventos Atípicos ou Amortizações Extraordinárias */}
        {ativo.is_provento_atipico && (
          <div className="my-6 p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-start gap-3.5">
            <span className="text-2xl mt-0.5">⚠️</span>
            <div>
              <h5 className="text-xs font-bold text-amber-400 uppercase tracking-wider">
                Alerta de Sustentabilidade: Provento Atípico / Amortização
              </h5>
              <p className="text-xs text-neutral-300 mt-1 leading-relaxed">
                {ativo.alerta_risco || "Este ativo distribuiu proventos não recorrentes que distorcem o Dividend Yield. O Preço Teto e o Score foram ajustados para a capacidade real de geração de caixa."}
              </p>
            </div>
          </div>
        )}

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

        {/* Preço Teto Consensual (Multi-Modelo) & Potencial de Ganho (% do Prêmio) */}
        <div className="my-6 p-5 rounded-2xl bg-gradient-to-r from-[#0c1815] via-[#101924] to-[#12121c] border border-emerald-500/30 shadow-xl relative overflow-hidden">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center gap-3.5">
              <div className="w-12 h-12 rounded-xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 shrink-0">
                <Target className="w-6 h-6" />
              </div>
              <div>
                <div className="flex items-center gap-2 flex-wrap">
                  <h4 className="text-base font-bold text-white">Preço Teto Consensual (Multi-Modelo)</h4>
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 font-mono font-bold">
                    Consenso
                  </span>
                </div>
                <p className="text-xs text-neutral-300 mt-0.5">
                  Consenso ponderado das escolas de valuation cruzando lucro, patrimônio e dividendos históricos
                </p>
              </div>
            </div>

            <div className="flex flex-col md:items-end">
              <span className="text-xs text-neutral-400 uppercase tracking-widest font-mono">Teto Estimado</span>
              <div className="flex items-baseline gap-2">
                <span className="text-3xl font-black text-white font-mono tabular-numbers">
                  {formatCurrency(tetoConsolidado)}
                </span>
                <span
                  className={`text-xs font-bold px-2 py-0.5 rounded-full ${
                    premioPct >= 15
                      ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                      : premioPct >= 0
                      ? "bg-amber-500/20 text-amber-400 border border-amber-500/40"
                      : "bg-rose-500/20 text-rose-400 border border-rose-500/40"
                  }`}
                >
                  {premioPct >= 0 ? "+" : ""}{premioPct.toFixed(1)}% prêmio
                </span>
              </div>
              <span className="text-[11px] text-neutral-400 mt-1">
                {premioPct >= 0 ? (
                  <span>Espaço de alta de <strong className="text-emerald-400 font-mono">{formatCurrency(tetoConsolidado - ativo.preco_atual)}</strong> por cota</span>
                ) : (
                  <span>Negociando <strong className="text-rose-400 font-mono">{formatCurrency(ativo.preco_atual - tetoConsolidado)}</strong> acima do teto</span>
                )}
              </span>
            </div>
          </div>

          {/* Modelos que participaram do Consenso */}
          {ativo.modelos_teto_consolidado && ativo.modelos_teto_consolidado.length > 0 && (
            <div className="mt-4 pt-3 border-t border-neutral-800/80 flex flex-wrap items-center gap-2">
              <span className="text-[11px] text-neutral-400 font-mono">Modelos no Consenso:</span>
              {ativo.modelos_teto_consolidado.map((m, idx) => (
                <span 
                  key={idx}
                  className="px-2 py-0.5 text-[11px] rounded-md bg-neutral-900 text-neutral-200 border border-neutral-700/80 font-mono"
                >
                  {m}
                </span>
              ))}
            </div>
          )}

          {/* Alerta de Outlier Normalizado */}
          {teveOutlier && (
            <div className="mt-3 p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/30 flex items-center gap-2 text-xs text-amber-300">
              <span className="text-base shrink-0">🛡️</span>
              <span>
                <strong>Blindagem Anti-Outlier Ativa:</strong> {obsOutlier || "Distribuições extraordinárias foram normalizadas com Winsorização para evitar teto artificialmente inflado."}
              </span>
            </div>
          )}
        </div>

        {/* Card do Analista Sênior: Risco vs Renda Fixa (NTN-B e Selic) */}
        <div className="my-6 p-5 rounded-2xl bg-gradient-to-r from-[#0d1624] via-[#0f1422] to-[#12121c] border border-cyan-500/30 shadow-xl">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center gap-3.5">
              <div className={`w-12 h-12 rounded-xl flex items-center justify-center shrink-0 ${
                ativo.veredito_risco === "COMPENSA_RISCO"
                  ? "bg-emerald-500/15 border border-emerald-500/40 text-emerald-400"
                  : ativo.veredito_risco === "NEUTRO"
                  ? "bg-amber-500/15 border border-amber-500/40 text-amber-400"
                  : "bg-rose-500/15 border border-rose-500/40 text-rose-400"
              }`}>
                <Scale className="w-6 h-6" />
              </div>
              <div>
                <div className="flex items-center gap-2 flex-wrap">
                  <h4 className="text-base font-bold text-white">Análise do Analista Sênior: Vale a Pena o Risco?</h4>
                  <span className={`text-[10px] px-2.5 py-0.5 rounded-full font-mono font-bold ${
                    ativo.veredito_risco === "COMPENSA_RISCO"
                      ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                      : ativo.veredito_risco === "NEUTRO"
                      ? "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                      : "bg-rose-500/20 text-rose-300 border border-rose-500/40"
                  }`}>
                    {ativo.veredito_risco === "COMPENSA_RISCO"
                      ? "🟢 Compensa o Risco"
                      : ativo.veredito_risco === "NEUTRO"
                      ? "🟡 Risco Neutro / Equilibrado"
                      : "🔴 Risco Descompensado"}
                  </span>
                </div>
                <p className="text-xs text-neutral-300 mt-1 leading-relaxed">
                  {ativo.justificativa_risco || (ativo.spread_ntnb >= 2 ? "Oferece spread de retorno expressivo em relação à renda fixa soberana." : "Rendimento próximo ou abaixo do custo de oportunidade da renda fixa.")}
                </p>
              </div>
            </div>

            <div className="flex flex-row md:flex-col items-center md:items-end justify-between gap-2 border-t md:border-t-0 pt-3 md:pt-0 border-neutral-800">
              <div>
                <span className="text-[10px] text-neutral-400 uppercase tracking-widest font-mono block">Spread vs NTN-B</span>
                <span className={`text-2xl font-black font-mono tabular-numbers ${
                  ativo.spread_ntnb >= 2.0 ? "text-emerald-400" : ativo.spread_ntnb >= 0 ? "text-amber-400" : "text-rose-400"
                }`}>
                  {ativo.spread_ntnb !== undefined ? (ativo.spread_ntnb >= 0 ? `+${ativo.spread_ntnb.toFixed(2)}%` : `${ativo.spread_ntnb.toFixed(2)}%`) : "—"}
                </span>
              </div>
              <div className="text-right">
                <span className="text-[10px] text-neutral-500 font-mono">Hurdle Mínimo: <strong className="text-neutral-300">{ativo.yield_exigido ? `${ativo.yield_exigido.toFixed(1)}% a.a.` : "8.0% a.a."}</strong></span>
                {ativo.tir_projetada && (
                  <span className="block text-[10px] text-cyan-400 font-mono">TIR Estimada: <strong>{ativo.tir_projetada.toFixed(1)}% a.a.</strong></span>
                )}
              </div>
            </div>
          </div>

          {/* Régua comparativa com a Renda Fixa Soberana */}
          <div className="mt-4 pt-3 border-t border-neutral-800/60 grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
            <div className="p-2.5 rounded-lg bg-neutral-900/80 border border-neutral-800 flex items-center justify-between">
              <span className="text-neutral-400">Yield Efetivo do Ativo:</span>
              <span className="font-bold text-white font-mono">{formatPercent(ativo.dy)}</span>
            </div>
            <div className="p-2.5 rounded-lg bg-neutral-900/80 border border-neutral-800 flex items-center justify-between">
              <span className="text-neutral-400">Tesouro IPCA+ (NTN-B):</span>
              <span className="font-bold text-emerald-400 font-mono">
                {macro ? `${macro.taxa_ntnb.toFixed(2)}% a.a. real` : "6.50% a.a. real"}
              </span>
            </div>
            <div className="p-2.5 rounded-lg bg-neutral-900/80 border border-neutral-800 flex items-center justify-between">
              <div className="flex flex-col">
                <span className="text-neutral-400">Taxa Selic Meta:</span>
                <span className="text-[10px] text-neutral-500 font-mono">
                  Líq. IR (15%): ~{((macro?.taxa_selic ?? 13.75) * 0.85).toFixed(2)}% a.a.
                </span>
              </div>
              <span className="font-bold text-amber-400 font-mono">
                {macro ? `${macro.taxa_selic.toFixed(2)}% a.a.` : "13.75% a.a."}
                <span className="text-[10px] text-neutral-400 font-normal ml-1">(Bruta)</span>
              </span>
            </div>
          </div>
        </div>

        {/* Indicadores Principais */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 my-6">
          <div className="p-3 rounded-lg bg-neutral-900/60 border border-neutral-800">
            <div className="flex items-center justify-between">
              <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">Dividend Yield 12M</span>
              {ativo.is_provento_atipico && (
                <span className="text-[9px] px-1 py-0.2 bg-amber-500/20 text-amber-400 rounded font-bold">Atípico</span>
              )}
            </div>
            <span className={`text-lg font-bold tabular-numbers ${ativo.is_provento_atipico ? "text-amber-400" : "text-emerald-400"}`}>
              {formatPercent(ativo.dy)}
            </span>
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
                <span className="text-[11px] text-neutral-400 uppercase tracking-wider block">Teto Bazin (5A vs 12M)</span>
                <span className="text-lg font-bold text-[#d4af37] tabular-numbers">{formatCurrency(teto5A)}</span>
                <span className="text-[11px] text-neutral-500 block mt-0.5">12M: {formatCurrency(ativo.preco_teto_bazin)}</span>
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

        {/* ========================================================================= */}
        {/* SEÇÃO ESPECIALISTA: DÉCIO BAZIN - MÉDIA 5 ANOS & AUDITORIA B3 vs STATUSINVEST */}
        {/* ========================================================================= */}
        <div className="my-6 p-5 rounded-2xl bg-gradient-to-b from-[#14141d] to-[#0c0c12] border border-[#d4af37]/40 shadow-xl">
          {/* Header com Ícone e Título */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-neutral-800 pb-3 mb-4">
            <div className="flex items-center gap-2.5">
              <div className="p-2 rounded-xl bg-[#d4af37]/15 text-[#d4af37]">
                <Landmark className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-sm md:text-base font-bold text-white flex items-center gap-2">
                  Método Décio Bazin: Preço Teto & Média dos Últimos 5 Anos
                  <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded-full bg-[#d4af37]/20 text-[#d4af37] border border-[#d4af37]/40 font-bold">
                    {ativo.classe === "FII" ? "CVM / B3" : "B3 Oficial"}
                  </span>
                </h3>
                <p className="text-xs text-neutral-400">
                  Compara a capacidade real do ciclo histórico (5A) com a fotografia recente (12M StatusInvest)
                </p>
              </div>
            </div>
            {/* Badge de Auditoria B3 vs StatusInvest */}
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-950/40 border border-emerald-500/40 text-emerald-400 text-xs font-mono self-start sm:self-auto">
              <ShieldCheck className="w-4 h-4 shrink-0" />
              <span>{analiseB3?.aderencia_statusinvest || ativo.aderencia_statusinvest || "✓ B3 vs StatusInvest: Conferido"}</span>
            </div>
          </div>

          {/* Cards Comparativos: Teto 5 Anos vs Teto 12 Meses */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-5">
            {/* Card 1: 5 Anos Recomendado */}
            <div className="p-4 rounded-xl bg-neutral-900/80 border border-[#d4af37]/40 relative overflow-hidden">
              <div className="absolute top-0 right-0 w-24 h-24 bg-[#d4af37]/5 rounded-bl-full pointer-events-none" />
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-bold uppercase tracking-wider text-[#d4af37] flex items-center gap-1.5">
                  👑 Preço Teto Bazin (Média 5 Anos)
                </span>
                <span className="text-[9px] px-1.5 py-0.5 rounded bg-[#d4af37]/20 text-[#d4af37] font-bold">
                  Recomendado
                </span>
              </div>
              <div className="mt-2.5 flex items-baseline gap-2">
                <span className="text-2xl font-black text-white font-mono">
                  {formatCurrency(teto5A)}
                </span>
                <span className={`text-xs font-bold font-mono ${margem5A >= 0 ? "text-emerald-400" : "text-rose-400"}`}>
                  {margem5A >= 0 ? "+" : ""}{margem5A.toFixed(1)}% margem
                </span>
              </div>
              <p className="text-xs text-neutral-300 mt-1">
                {ativo.classe === "FII" ? "Média Isenta 5A: " : "Média Líquida 5A: "}
                <strong className="text-[#d4af37] font-mono">{formatCurrency(media5ANorm)}/ano</strong>
                {teveOutlier && (
                  <span className="text-[10px] text-neutral-400 ml-1.5 font-mono">(Bruta: {formatCurrency(media5A)})</span>
                )}
              </p>
              {teveOutlier ? (
                <div className="mt-2 pt-2 border-t border-amber-500/20 text-[10px] text-amber-300 flex items-start gap-1">
                  <span className="shrink-0">🛡️</span>
                  <span><strong>Blindagem Anti-Outlier:</strong> {obsOutlier || "Distorção extraordinária ajustada com Winsorização."}</span>
                </div>
              ) : (
                <span className="text-[10px] text-neutral-400 block mt-2 border-t border-neutral-800/60 pt-2">
                  • Disciplina Bazin: Dilui picos atípicos e oscilações de mercado, revelando o teto sustentável de longo prazo.
                </span>
              )}
            </div>

            {/* Card 2: 12 Meses StatusInvest */}
            <div className="p-4 rounded-xl bg-neutral-900/60 border border-neutral-800 relative">
              <span className="text-[11px] font-bold uppercase tracking-wider text-neutral-400 flex items-center gap-1.5">
                ⏱️ Preço Teto Bazin (Últimos 12M)
              </span>
              <div className="mt-2.5 flex items-baseline gap-2">
                <span className="text-2xl font-bold text-neutral-300 font-mono">
                  {formatCurrency(ativo.preco_teto_bazin)}
                </span>
                <span className={`text-xs font-bold font-mono ${ativo.margem_bazin >= 0 ? "text-emerald-400" : "text-rose-400"}`}>
                  {ativo.margem_bazin >= 0 ? "+" : ""}{ativo.margem_bazin.toFixed(1)}% margem
                </span>
              </div>
              <p className="text-xs text-neutral-300 mt-1">
                Proventos 12M: <strong className="text-neutral-200 font-mono">{formatCurrency(ativo.dividendos_12m)}/ano</strong> (DY: {formatPercent(ativo.dy)})
              </p>
              <span className="text-[10px] text-neutral-400 block mt-2 border-t border-neutral-800/60 pt-2">
                • Fotografia dos últimos 12 meses (StatusInvest). Sensível a eventos extraordinários recentes.
              </span>
            </div>
          </div>

          {/* Gráfico / Histórico de Proventos Ano a Ano */}
          <div className="p-4 rounded-xl bg-neutral-950/70 border border-neutral-800/80">
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-semibold text-neutral-300 flex items-center gap-2">
                <BarChart3 className="w-4 h-4 text-[#d4af37]" />
                {ativo.classe === "FII" ? "Histórico de Rendimentos Isentos por Ano (CVM / B3)" : "Histórico de Proventos Líquidos por Ano (B3 Oficial)"}
              </span>
              <span className="text-[11px] text-[#d4af37] font-mono font-semibold">
                Média 5A: {formatCurrency(media5A)}/ano
              </span>
            </div>

            {/* Barras Anuais */}
            {entradasAnos.length > 0 ? (
              <div className="grid grid-cols-4 sm:grid-cols-6 gap-2 pt-2 items-end">
                {entradasAnos.map(([ano, valor]) => {
                  const alturaPercentual = Math.min(100, Math.max(14, (valor / maxValorAno) * 100));
                  const isMaior = valor === maxValorAno;
                  return (
                    <div key={ano} className="flex flex-col items-center gap-1.5 group">
                      <span className="text-[10px] font-mono text-neutral-400 font-semibold group-hover:text-white transition-colors">
                        {formatCurrency(valor)}
                      </span>
                      <div className="w-full bg-neutral-900 rounded-t-lg h-24 flex items-end p-1 relative overflow-hidden">
                        <div 
                          style={{ height: `${alturaPercentual}%` }}
                          className={`w-full rounded-t transition-all duration-500 ${
                            isMaior 
                              ? "bg-gradient-to-t from-[#997d26] to-[#ffd700]" 
                              : "bg-gradient-to-t from-emerald-800 to-emerald-500"
                          }`}
                        />
                      </div>
                      <span className="text-[11px] font-mono font-bold text-neutral-300">
                        {ano}
                      </span>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="py-4 text-center text-xs text-neutral-500">
                {carregandoB3 ? "Consultando eventos em dinheiro na B3 e CVM..." : "Proventos dos 5 anos estimados via crescimento composto de lucros."}
              </div>
            )}

            {/* Linha Informativa de Auditoria de Conferência */}
            <div className="mt-4 pt-3 border-t border-neutral-800/60 flex flex-wrap items-center justify-between text-[11px] text-neutral-400 gap-2">
              <div className="flex items-center gap-2 flex-wrap">
                <span className="w-2 h-2 rounded-full bg-emerald-400" />
                <span>{ativo.classe === "FII" ? "Rendimentos 12M: " : "B3 12M Bruto: "}<strong className="text-white font-mono">{formatCurrency(analiseB3?.dividendos_12m_b3_bruto || ativo.dividendos_12m_b3 || ativo.dividendos_12m)}</strong></span>
                <span>•</span>
                <span>StatusInvest 12M: <strong className="text-white font-mono">{formatCurrency(analiseB3?.dividendos_12m_statusinvest || ativo.dividendos_12m)}</strong></span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-neutral-500">Conferência Ticker:</span>
                <span className="font-mono text-emerald-400 font-bold">
                  {analiseB3?.aderencia_percentual ? `${analiseB3.aderencia_percentual}% aderente` : "Conferido"}
                </span>
              </div>
            </div>
          </div>
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
                  <span className="text-neutral-500 block text-[10px] uppercase">Piotroski F-Score</span>
                  <span className={`font-semibold ${ativo.piotroski_score && ativo.piotroski_score >= 7 ? "text-emerald-400" : ativo.piotroski_score && ativo.piotroski_score >= 5 ? "text-amber-400" : "text-rose-400"}`}>
                    {ativo.piotroski_score !== undefined && ativo.piotroski_score > 0 ? `${ativo.piotroski_score}/9 pts` : "—"}
                  </span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Preço Teto DCF</span>
                  <span className="font-semibold text-cyan-400">
                    {ativo.preco_teto_dcf && ativo.preco_teto_dcf > 0 ? formatCurrency(ativo.preco_teto_dcf) : "Inaplicável"}
                  </span>
                </div>
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
                  <span className="text-neutral-500 block text-[10px] uppercase">Payout Histórico</span>
                  <span className={`font-semibold ${ativo.payout && ativo.payout > 100 ? "text-amber-400" : "text-neutral-200"}`}>
                    {ativo.payout !== undefined && ativo.payout > 0
                      ? ativo.payout > 500
                        ? "> 500%"
                        : `${ativo.payout.toFixed(0)}%`
                      : "—"}
                  </span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Dív. Líquida / PL</span>
                  <span className="font-semibold text-neutral-200">{ativo.divida_liquida_pl !== undefined ? `${ativo.divida_liquida_pl.toFixed(2)}x` : "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">P/EBIT</span>
                  <span className="font-semibold text-neutral-200">{ativo.p_ebit !== undefined && ativo.p_ebit > 0 ? `${ativo.p_ebit.toFixed(1)}x` : "-"}</span>
                </div>
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5 text-xs">
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Cap Rate Implícito</span>
                  <span className="font-semibold text-emerald-400">
                    {ativo.cap_rate_implicito ? formatPercent(ativo.cap_rate_implicito) : (ativo.pvp && ativo.pvp > 0 ? formatPercent(ativo.dy / ativo.pvp) : "—")}
                  </span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Segmento FII</span>
                  <span className="font-semibold text-neutral-200">{ativo.segmento || "-"}</span>
                </div>
                <div className="p-2 rounded bg-neutral-900/80 border border-neutral-800/60">
                  <span className="text-neutral-500 block text-[10px] uppercase">Gestão</span>
                  <span className="font-semibold text-neutral-200">{ativo.gestao || "-"}</span>
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
          <span>Dados oficiais CVM e B3 • Auditoria Multi-Fonte Ativa</span>
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
