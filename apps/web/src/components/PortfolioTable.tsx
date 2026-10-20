"use client";

import React, { useState } from "react";
import { ClasseAtivo, PosicaoConsolidada } from "@/types/portfolio";
import { Trash2, TrendingUp, TrendingDown, ShieldAlert, CheckCircle2, Clock } from "lucide-react";

interface PortfolioTableProps {
  posicoes: PosicaoConsolidada[];
  onExcluirPosicao: (id: number, ticker: string) => void;
}

export function PortfolioTable({ posicoes, onExcluirPosicao }: PortfolioTableProps) {
  const [filtroClasse, setFiltroClasse] = useState<ClasseAtivo | "TODAS">("TODAS");
  const [busca, setBusca] = useState("");

  const filtradas = posicoes.filter((p) => {
    if (filtroClasse !== "TODAS" && p.classe !== filtroClasse) return false;
    if (busca.trim()) {
      const termo = busca.toLowerCase();
      return (
        p.ticker.toLowerCase().includes(termo) ||
        p.corretora.toLowerCase().includes(termo) ||
        p.categoria_macro.toLowerCase().includes(termo)
      );
    }
    return true;
  });

  const getBadgeClasse = (classe: ClasseAtivo) => {
    switch (classe) {
      case "RENDA_FIXA":
        return <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-blue-500/15 text-blue-300 border border-blue-500/30">Renda Fixa</span>;
      case "FII":
        return <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-purple-500/15 text-purple-300 border border-purple-500/30">FII</span>;
      case "ACAO_BR":
        return <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/15 text-emerald-300 border border-emerald-500/30">Ação BR</span>;
      case "INTERNACIONAL":
        return <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-cyan-500/15 text-cyan-300 border border-cyan-500/30">Internacional</span>;
      case "CAIXA":
        return <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-amber-500/15 text-amber-300 border border-amber-500/30">Caixa</span>;
    }
  };

  const getBadgeSinal = (sinal: string) => {
    switch (sinal) {
      case "COMPRAR_APORTE":
        return (
          <span className="flex items-center gap-1 text-[11px] font-bold font-mono px-2.5 py-1 rounded-md bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
            <CheckCircle2 className="w-3.5 h-3.5" /> Comprar Aporte
          </span>
        );
      case "REAVALIAR":
        return (
          <span className="flex items-center gap-1 text-[11px] font-bold font-mono px-2.5 py-1 rounded-md bg-red-500/15 text-red-400 border border-red-500/30">
            <ShieldAlert className="w-3.5 h-3.5" /> Reavaliar
          </span>
        );
      default:
        return (
          <span className="flex items-center gap-1 text-[11px] font-mono px-2.5 py-1 rounded-md bg-neutral-800 text-neutral-400 border border-neutral-700">
            <Clock className="w-3.5 h-3.5" /> Aguardar
          </span>
        );
    }
  };

  return (
    <div className="rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-2xl overflow-hidden backdrop-blur-sm">
      {/* Barra de Filtros e Busca */}
      <div className="p-4 sm:p-5 border-b border-neutral-800/80 flex flex-col md:flex-row md:items-center justify-between gap-4">
        {/* Abas de Classes */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-2 md:pb-0 scrollbar-none">
          {(["TODAS", "ACAO_BR", "FII", "RENDA_FIXA", "INTERNACIONAL", "CAIXA"] as const).map((aba) => (
            <button
              key={aba}
              onClick={() => setFiltroClasse(aba)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold whitespace-nowrap transition-all ${
                filtroClasse === aba
                  ? "bg-[#d4af37] text-black font-bold shadow-md shadow-[#d4af37]/20"
                  : "bg-neutral-900/80 hover:bg-neutral-800 text-neutral-300 border border-neutral-800"
              }`}
            >
              {aba === "TODAS"
                ? `Todos (${posicoes.length})`
                : aba === "ACAO_BR"
                ? "Ações BR"
                : aba === "FII"
                ? "FIIs"
                : aba === "RENDA_FIXA"
                ? "Renda Fixa"
                : aba === "INTERNACIONAL"
                ? "Internacional"
                : "Caixa"}
            </button>
          ))}
        </div>

        {/* Input de Busca */}
        <div className="w-full md:w-64">
          <input
            type="text"
            placeholder="Buscar por ticker, corretora..."
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            className="w-full px-3.5 py-2 rounded-xl bg-neutral-950 border border-neutral-800 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-[#d4af37]/60"
          />
        </div>
      </div>

      {/* Tabela de Posições */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-neutral-800/80 bg-neutral-950/60 text-neutral-400 font-mono uppercase text-[11px]">
              <th className="py-3.5 px-4 font-semibold">Ativo / Corretora</th>
              <th className="py-3.5 px-4 font-semibold text-right">Qtd / Preço Médio</th>
              <th className="py-3.5 px-4 font-semibold text-right">Cotação Atual</th>
              <th className="py-3.5 px-4 font-semibold text-right">Total Investido</th>
              <th className="py-3.5 px-4 font-semibold text-right">Valor de Mercado</th>
              <th className="py-3.5 px-4 font-semibold text-right">Lucro / Retorno</th>
              <th className="py-3.5 px-4 font-semibold text-right">Preço Teto / Upside</th>
              <th className="py-3.5 px-4 font-semibold text-center">Score</th>
              <th className="py-3.5 px-4 font-semibold text-center">Sinal Aporte</th>
              <th className="py-3.5 px-4 font-semibold text-center">Ações</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-neutral-800/50">
            {filtradas.length === 0 ? (
              <tr>
                <td colSpan={10} className="py-12 text-center text-neutral-500 font-mono">
                  Nenhum ativo encontrado para o filtro selecionado.
                </td>
              </tr>
            ) : (
              filtradas.map((p) => {
                const isPositivo = p.lucro_prejuizo_valor >= 0;

                return (
                  <tr
                    key={p.id}
                    className="hover:bg-neutral-800/25 transition-colors group"
                  >
                    {/* Ticker & Corretora */}
                    <td className="py-3.5 px-4">
                      <div className="flex items-center gap-2">
                        <span className="font-extrabold text-white text-sm font-mono tracking-wide">
                          {p.ticker}
                        </span>
                        {getBadgeClasse(p.classe)}
                      </div>
                      <div className="flex items-center gap-2 text-[11px] text-neutral-400 font-mono mt-0.5">
                        <span>{p.corretora}</span>
                        {p.taxa_contratada && (
                          <>
                            <span>•</span>
                            <span className="text-cyan-400 font-semibold">{p.taxa_contratada}</span>
                          </>
                        )}
                      </div>
                    </td>

                    {/* Qtd & PM */}
                    <td className="py-3.5 px-4 text-right font-mono">
                      <div className="text-white font-medium">
                        {p.quantidade.toLocaleString("pt-BR", { maximumFractionDigits: 4 })}
                      </div>
                      <div className="text-[11px] text-neutral-400">
                        PM: {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(p.preco_medio)}
                      </div>
                    </td>

                    {/* Cotação Atual */}
                    <td className="py-3.5 px-4 text-right font-mono">
                      <div className="font-bold text-white">
                        {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(p.cotacao_atual)}
                      </div>
                      <span className="text-[10px] text-neutral-500">B3 Oficial</span>
                    </td>

                    {/* Total Investido */}
                    <td className="py-3.5 px-4 text-right font-mono text-neutral-300">
                      {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(p.valor_aplicado)}
                    </td>

                    {/* Valor de Mercado */}
                    <td className="py-3.5 px-4 text-right font-mono font-bold text-white">
                      {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(p.valor_atual_mercado)}
                    </td>

                    {/* Lucro e Rentabilidade */}
                    <td className="py-3.5 px-4 text-right font-mono">
                      <div className={`font-extrabold ${isPositivo ? "text-emerald-400" : "text-red-400"}`}>
                        {isPositivo ? "+" : ""}
                        {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(p.lucro_prejuizo_valor)}
                      </div>
                      <div className={`text-[11px] flex items-center justify-end gap-0.5 ${isPositivo ? "text-emerald-400" : "text-red-400"}`}>
                        {isPositivo ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
                        {isPositivo ? "+" : ""}
                        {p.rentabilidade_pct.toFixed(2)}%
                      </div>
                    </td>

                    {/* Preço Teto / Upside */}
                    <td className="py-3.5 px-4 text-right font-mono">
                      {p.preco_teto_estimado > 0 ? (
                        <>
                          <div className="text-neutral-200">
                            {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(p.preco_teto_estimado)}
                          </div>
                          <div
                            className={`text-[11px] font-bold ${
                              p.margem_seguranca_pct >= 0 ? "text-emerald-400" : "text-amber-400"
                            }`}
                          >
                            {p.margem_seguranca_pct >= 0 ? "+" : ""}
                            {p.margem_seguranca_pct.toFixed(1)}% margem
                          </div>
                        </>
                      ) : (
                        <span className="text-neutral-500 text-[11px]">-</span>
                      )}
                    </td>

                    {/* Score */}
                    <td className="py-3.5 px-4 text-center font-mono">
                      {p.score_fundamentalista > 0 ? (
                        <span
                          className={`px-2 py-0.5 rounded text-xs font-bold ${
                            p.score_fundamentalista >= 80
                              ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                              : p.score_fundamentalista >= 60
                              ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                              : "bg-red-500/20 text-red-300 border border-red-500/30"
                          }`}
                        >
                          {p.score_fundamentalista}
                        </span>
                      ) : (
                        <span className="text-neutral-500 font-mono text-[11px]">N/A</span>
                      )}
                    </td>

                    {/* Sinal de Aporte */}
                    <td className="py-3.5 px-4 text-center">
                      {getBadgeSinal(p.sinal_rebalanceamento)}
                    </td>

                    {/* Ações */}
                    <td className="py-3.5 px-4 text-center">
                      <button
                        onClick={() => onExcluirPosicao(p.id, p.ticker)}
                        className="p-1.5 rounded-lg text-neutral-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                        title="Remover ativo da carteira"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
