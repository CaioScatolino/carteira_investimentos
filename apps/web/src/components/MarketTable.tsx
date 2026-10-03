"use client";

import React, { useState, useMemo } from "react";
import { Ativo } from "@/types/market";
import { StatusBadge } from "./StatusBadge";
import { Search, ChevronLeft, ChevronRight, ArrowUpDown, ChevronDown } from "lucide-react";

interface MarketTableProps {
  titulo: string;
  subtitulo: string;
  icone: React.ReactNode;
  ativos: Ativo[];
  classe: "ACAO" | "FII" | "ETF";
  onSelectAtivo: (ativo: Ativo) => void;
}

export function MarketTable({
  titulo,
  subtitulo,
  icone,
  ativos,
  classe,
  onSelectAtivo,
}: MarketTableProps) {
  const [busca, setBusca] = useState("");
  const [pagina, setPagina] = useState(1);
  const [itensPorPagina, setItensPorPagina] = useState(8);

  // Filtro de busca por Ticker ou Nome
  const filtrados = useMemo(() => {
    if (!busca.trim()) return ativos;
    const termo = busca.toLowerCase().trim();
    return ativos.filter(
      (a) =>
        a.ticker.toLowerCase().includes(termo) ||
        (a.nome && a.nome.toLowerCase().includes(termo))
    );
  }, [ativos, busca]);

  // Paginação
  const totalPaginas = Math.max(1, Math.ceil(filtrados.length / itensPorPagina));
  const paginaAtual = Math.min(pagina, totalPaginas);
  const inicioIdx = (paginaAtual - 1) * itensPorPagina;
  const paginados = filtrados.slice(inicioIdx, inicioIdx + itensPorPagina);

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
    if (val >= 1_000_000) return `R$ ${(val / 1_000_000).toFixed(1)}M`;
    if (val >= 1_000) return `R$ ${(val / 1_000).toFixed(0)}K`;
    return `R$ ${val.toFixed(0)}`;
  };

  return (
    <div className="rounded-2xl bg-[#0c0c10] border border-[#22222a] hover:border-[#d4af37]/30 transition-all duration-300 shadow-xl overflow-hidden flex flex-col mb-8">
      {/* Barra Superior da Tabela */}
      <div className="p-5 md:p-6 border-b border-neutral-800/80 flex flex-col md:flex-row md:items-center justify-between gap-4 bg-[#0e0e14]">
        <div className="flex items-center gap-3.5">
          <div className="w-10 h-10 rounded-xl bg-[#d4af37]/10 border border-[#d4af37]/30 flex items-center justify-center text-[#d4af37]">
            {icone}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-lg font-bold text-white tracking-wide">{titulo}</h3>
              <span className="text-xs px-2 py-0.5 rounded-full bg-neutral-900 text-[#d4af37] border border-[#d4af37]/30 font-mono font-semibold">
                {filtrados.length}
              </span>
            </div>
            <p className="text-xs text-neutral-400 mt-0.5">{subtitulo}</p>
          </div>
        </div>

        {/* Campo de Pesquisa */}
        <div className="flex items-center gap-3">
          <div className="relative w-full md:w-64">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-neutral-500" />
            <input
              type="text"
              placeholder={`Buscar ${classe === "ACAO" ? "ação" : classe}...`}
              value={busca}
              onChange={(e) => {
                setBusca(e.target.value);
                setPagina(1);
              }}
              className="w-full pl-9 pr-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-neutral-200 placeholder-neutral-500 focus:outline-none focus:border-[#d4af37]/60 transition-colors"
            />
          </div>
        </div>
      </div>

      {/* Conteúdo Desktop: Tabela Institucional */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-neutral-800/60 bg-[#09090d] text-neutral-400 font-mono uppercase tracking-wider text-[11px]">
              <th className="py-3 px-4 w-12 text-center">#</th>
              <th className="py-3 px-4">Ativo</th>
              <th className="py-3 px-4 text-right">Cotação</th>
              <th className="py-3 px-4 text-center">Score</th>
              
              {classe === "ACAO" && (
                <>
                  <th className="py-3 px-4 text-right">P/L</th>
                  <th className="py-3 px-4 text-right">P/VP</th>
                  <th className="py-3 px-4 text-right">ROE</th>
                  <th className="py-3 px-4 text-right">Yield 12M</th>
                  <th className="py-3 px-4 text-right">Teto Bazin</th>
                  <th className="py-3 px-4 text-right">VI Graham</th>
                </>
              )}

              {classe === "FII" && (
                <>
                  <th className="py-3 px-4 text-right">P/VP CVM</th>
                  <th className="py-3 px-4 text-right">VP Cota</th>
                  <th className="py-3 px-4 text-right">Yield 12M</th>
                  <th className="py-3 px-4 text-right">Spread NTN-B</th>
                  <th className="py-3 px-4 text-right">Teto Bazin</th>
                </>
              )}

              {classe === "ETF" && (
                <>
                  <th className="py-3 px-4 text-right">Volume Diário B3</th>
                  <th className="py-3 px-4">Nível de Liquidez</th>
                </>
              )}

              <th className="py-3 px-4 text-center">Semáforo</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-neutral-800/30">
            {paginados.length === 0 ? (
              <tr>
                <td colSpan={10} className="py-8 text-center text-neutral-500">
                  Nenhum ativo encontrado para o filtro digitado.
                </td>
              </tr>
            ) : (
              paginados.map((ativo, idx) => {
                const rankingGeral = inicioIdx + idx + 1;
                return (
                  <tr
                    key={ativo.ticker}
                    onClick={() => onSelectAtivo(ativo)}
                    className="hover:bg-[#13131b] transition-colors cursor-pointer group"
                  >
                    <td className="py-3 px-4 text-center font-mono text-neutral-500 group-hover:text-[#d4af37]">
                      {rankingGeral}º
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-white font-mono group-hover:text-[#d4af37] transition-colors">
                          {ativo.ticker}
                        </span>
                        <span className="text-[11px] text-neutral-500 truncate max-w-[120px]">
                          {ativo.nome}
                        </span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-right font-medium text-white tabular-numbers">
                      {formatCurrency(ativo.preco_atual)}
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center justify-center gap-2">
                        <div className="w-12 h-1.5 rounded-full bg-neutral-800 overflow-hidden">
                          <div
                            className={`h-full rounded-full ${
                              ativo.score >= 70
                                ? "bg-gradient-to-r from-[#d4af37] to-emerald-400"
                                : ativo.score >= 50
                                ? "bg-amber-400"
                                : "bg-neutral-600"
                            }`}
                            style={{ width: `${Math.min(100, ativo.score)}%` }}
                          />
                        </div>
                        <span className="font-mono font-bold text-neutral-200 tabular-numbers w-6 text-right">
                          {ativo.score}
                        </span>
                      </div>
                    </td>

                    {/* Colunas específicas de Ações */}
                    {classe === "ACAO" && (
                      <>
                        <td className="py-3 px-4 text-right font-mono text-neutral-300 tabular-numbers">
                          {ativo.pl > 0 ? `${ativo.pl.toFixed(1)}x` : "—"}
                        </td>
                        <td className="py-3 px-4 text-right font-mono text-neutral-300 tabular-numbers">
                          {ativo.pvp_real > 0 ? `${ativo.pvp_real.toFixed(2)}x` : "—"}
                        </td>
                        <td className="py-3 px-4 text-right font-mono text-neutral-300 tabular-numbers">
                          {ativo.roe > 0 ? formatPercent(ativo.roe) : "—"}
                        </td>
                        <td className="py-3 px-4 text-right font-mono font-semibold text-emerald-400 tabular-numbers">
                          {formatPercent(ativo.dy)}
                        </td>
                        <td className="py-3 px-4 text-right font-mono text-neutral-300 tabular-numbers">
                          {ativo.preco_teto_bazin > 0 ? formatCurrency(ativo.preco_teto_bazin) : "—"}
                        </td>
                        <td className="py-3 px-4 text-right font-mono text-neutral-300 tabular-numbers">
                          {ativo.valor_graham > 0 ? formatCurrency(ativo.valor_graham) : "—"}
                        </td>
                      </>
                    )}

                    {/* Colunas específicas de FIIs */}
                    {classe === "FII" && (
                      <>
                        <td className="py-3 px-4 text-right font-mono tabular-numbers">
                          <span
                            className={
                              ativo.pvp <= 0.95
                                ? "text-emerald-400 font-semibold"
                                : ativo.pvp <= 1.02
                                ? "text-neutral-300"
                                : "text-amber-400"
                            }
                          >
                            {ativo.pvp > 0 ? `${ativo.pvp.toFixed(2)}x` : "—"}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-right font-mono text-neutral-400 tabular-numbers">
                          {formatCurrency(ativo.vp_cota)}
                        </td>
                        <td className="py-3 px-4 text-right font-mono font-semibold text-emerald-400 tabular-numbers">
                          {formatPercent(ativo.dy)}
                        </td>
                        <td className="py-3 px-4 text-right font-mono tabular-numbers">
                          <span
                            className={
                              ativo.spread_ntnb >= 2.0
                                ? "text-emerald-400 font-semibold"
                                : ativo.spread_ntnb >= 0
                                ? "text-amber-400"
                                : "text-rose-400"
                            }
                          >
                            {ativo.spread_ntnb >= 0
                              ? `+${ativo.spread_ntnb.toFixed(2)}%`
                              : `${ativo.spread_ntnb.toFixed(2)}%`}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-right font-mono text-neutral-300 tabular-numbers">
                          {formatCurrency(ativo.preco_teto_bazin)}
                        </td>
                      </>
                    )}

                    {/* Colunas específicas de ETFs */}
                    {classe === "ETF" && (
                      <>
                        <td className="py-3 px-4 text-right font-mono font-semibold text-neutral-200 tabular-numbers">
                          {formatVolume(ativo.volume_total)}
                        </td>
                        <td className="py-3 px-4 text-neutral-400">
                          {ativo.volume_total >= 50_000_000 ? (
                            <span className="text-emerald-400 font-medium">Altíssima Liquidez</span>
                          ) : ativo.volume_total >= 5_000_000 ? (
                            <span className="text-neutral-300 font-medium">Boa Liquidez</span>
                          ) : (
                            <span className="text-amber-400 font-medium">Liquidez Regular</span>
                          )}
                        </td>
                      </>
                    )}

                    <td className="py-3 px-4 text-center">
                      <StatusBadge status={ativo.status} />
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Barra Inferior com Paginação Independente */}
      <div className="p-4 border-t border-neutral-800/80 bg-[#09090d] flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-neutral-400">
        <div>
          Mostrando <span className="text-white font-medium">{paginados.length > 0 ? inicioIdx + 1 : 0}</span> a{" "}
          <span className="text-white font-medium">{Math.min(inicioIdx + itensPorPagina, filtrados.length)}</span> de{" "}
          <span className="text-[#d4af37] font-semibold">{filtrados.length}</span> ativos
        </div>

        <div className="flex items-center gap-2">
          {/* Seletor de itens por página */}
          <div className="flex items-center gap-1.5 mr-2">
            <span>Linhas:</span>
            <select
              value={itensPorPagina}
              onChange={(e) => {
                setItensPorPagina(Number(e.target.value));
                setPagina(1);
              }}
              className="bg-neutral-900 border border-neutral-800 text-neutral-300 rounded px-2 py-1 text-xs focus:outline-none focus:border-[#d4af37]"
            >
              <option value={5}>5</option>
              <option value={8}>8</option>
              <option value={15}>15</option>
              <option value={25}>25</option>
            </select>
          </div>

          {/* Botões de Navegação */}
          <button
            onClick={() => setPagina((p) => Math.max(1, p - 1))}
            disabled={paginaAtual <= 1}
            className="p-1.5 rounded-lg bg-neutral-900 border border-neutral-800 hover:border-neutral-700 disabled:opacity-40 disabled:cursor-not-allowed text-neutral-300 hover:text-white"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="px-2 font-mono text-neutral-300">
            {paginaAtual} / {totalPaginas}
          </span>
          <button
            onClick={() => setPagina((p) => Math.min(totalPaginas, p + 1))}
            disabled={paginaAtual >= totalPaginas}
            className="p-1.5 rounded-lg bg-neutral-900 border border-neutral-800 hover:border-neutral-700 disabled:opacity-40 disabled:cursor-not-allowed text-neutral-300 hover:text-white"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
