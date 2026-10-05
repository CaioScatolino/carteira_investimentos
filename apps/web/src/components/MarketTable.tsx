"use client";

import React, { useState, useMemo } from "react";
import { Ativo } from "@/types/market";
import { StatusBadge } from "./StatusBadge";
import { Search, ChevronLeft, ChevronRight, ArrowUpDown, ChevronDown, Target, TrendingUp, ShieldCheck } from "lucide-react";

interface MarketTableProps {
  titulo: string;
  subtitulo: string;
  icone: React.ReactNode;
  ativos: Ativo[];
  classe: "ACAO" | "FII" | "ETF";
  onSelectAtivo: (ativo: Ativo) => void;
}

type SortKey = "SCORE" | "PREMIO" | "SPREAD_NTNB" | "DY" | "MARGEM" | "COTACAO";

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
  const [filtroPorte, setFiltroPorte] = useState<string>("TODOS");
  const [criterioOrdenacao, setCriterioOrdenacao] = useState<SortKey>("SCORE");
  const [direcaoOrdenacao, setDirecaoOrdenacao] = useState<"ASC" | "DESC">("DESC");

  const alternarOrdenacao = (key: SortKey) => {
    if (criterioOrdenacao === key) {
      setDirecaoOrdenacao((prev) => (prev === "DESC" ? "ASC" : "DESC"));
    } else {
      setCriterioOrdenacao(key);
      setDirecaoOrdenacao("DESC");
    }
    setPagina(1);
  };

  // Filtro de busca por Ticker ou Nome, Porte e Ordenação Dinâmica
  const filtrados = useMemo(() => {
    let lista = [...ativos];

    if (filtroPorte === "BLUE_CHIPS") {
      lista = lista.filter((a) => a.porte === "BLUE_CHIP" || a.porte === "FII_GIGANTE");
    } else if (filtroPorte === "MID_SMALL") {
      lista = lista.filter(
        (a) => a.porte === "MID_CAP" || a.porte === "SMALL_CAP" || a.porte === "FII_CONSOLIDADO"
      );
    }

    if (busca.trim()) {
      const termo = busca.toLowerCase().trim();
      lista = lista.filter(
        (a) =>
          a.ticker.toLowerCase().includes(termo) ||
          (a.nome && a.nome.toLowerCase().includes(termo))
      );
    }

    // Ordenação dinâmica solicitada
    lista.sort((a, b) => {
      let valA = 0;
      let valB = 0;

      switch (criterioOrdenacao) {
        case "PREMIO":
          valA = a.premio_desconto_percentual ?? (((a.preco_teto_consolidado || a.preco_teto_bazin_5a || a.preco_teto_bazin || 0) - a.preco_atual) / (a.preco_atual || 1) * 100);
          valB = b.premio_desconto_percentual ?? (((b.preco_teto_consolidado || b.preco_teto_bazin_5a || b.preco_teto_bazin || 0) - b.preco_atual) / (b.preco_atual || 1) * 100);
          break;
        case "SPREAD_NTNB":
          valA = a.spread_ntnb ?? -999;
          valB = b.spread_ntnb ?? -999;
          break;
        case "DY":
          valA = a.dy || 0;
          valB = b.dy || 0;
          break;
        case "MARGEM":
          valA = a.margem_seguranca_consolidada ?? a.margem_bazin_5a ?? a.margem_bazin ?? 0;
          valB = b.margem_seguranca_consolidada ?? b.margem_bazin_5a ?? b.margem_bazin ?? 0;
          break;
        case "COTACAO":
          valA = a.preco_atual || 0;
          valB = b.preco_atual || 0;
          break;
        case "SCORE":
        default:
          valA = a.score || 0;
          valB = b.score || 0;
          if (valA === valB) {
            valA = a.volume_total || 0;
            valB = b.volume_total || 0;
          }
          break;
      }

      if (direcaoOrdenacao === "ASC") {
        return valA - valB;
      }
      return valB - valA;
    });

    return lista;
  }, [ativos, busca, filtroPorte, criterioOrdenacao, direcaoOrdenacao]);

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

      {/* Sub-barra de Filtros Rápidos de Porte e Classificação / Ordenação */}
      <div className="px-5 py-3 bg-[#0a0a0e] border-b border-neutral-800/60 flex flex-col md:flex-row md:items-center justify-between gap-3 text-xs">
        {/* Lado Esquerdo: Filtros de Porte (Ações e FIIs) */}
        {classe !== "ETF" ? (
          <div className="flex items-center gap-2 overflow-x-auto pb-1 md:pb-0">
            <span className="text-[11px] text-neutral-500 uppercase tracking-wider font-mono mr-1">Porte:</span>
            <button
              onClick={() => {
                setFiltroPorte("TODOS");
                setPagina(1);
              }}
              className={`px-3 py-1 rounded-lg font-medium transition-all ${
                filtroPorte === "TODOS"
                  ? "bg-[#d4af37] text-black font-semibold shadow-sm"
                  : "bg-neutral-900 text-neutral-400 hover:text-white border border-neutral-800"
              }`}
            >
              Todos ({ativos.length})
            </button>

            <button
              onClick={() => {
                setFiltroPorte("BLUE_CHIPS");
                setPagina(1);
              }}
              className={`px-3 py-1 rounded-lg font-medium transition-all flex items-center gap-1.5 ${
                filtroPorte === "BLUE_CHIPS"
                  ? "bg-[#d4af37] text-black font-semibold shadow-sm"
                  : "bg-neutral-900 text-neutral-400 hover:text-white border border-neutral-800"
              }`}
            >
              <span>{classe === "ACAO" ? "👑 Blue Chips" : "🏰 FIIs Baleia"}</span>
              <span
                className={`text-[10px] px-1.5 rounded-full ${
                  filtroPorte === "BLUE_CHIPS" ? "bg-black/20 text-black font-bold" : "bg-neutral-800 text-[#d4af37]"
                }`}
              >
                {ativos.filter((a) => a.porte === "BLUE_CHIP" || a.porte === "FII_GIGANTE").length}
              </span>
            </button>

            <button
              onClick={() => {
                setFiltroPorte("MID_SMALL");
                setPagina(1);
              }}
              className={`px-3 py-1 rounded-lg font-medium transition-all flex items-center gap-1.5 ${
                filtroPorte === "MID_SMALL"
                  ? "bg-[#d4af37] text-black font-semibold shadow-sm"
                  : "bg-neutral-900 text-neutral-400 hover:text-white border border-neutral-800"
              }`}
            >
              <span>{classe === "ACAO" ? "⚡ Mid & Small" : "🏢 Consolidados"}</span>
              <span
                className={`text-[10px] px-1.5 rounded-full ${
                  filtroPorte === "MID_SMALL" ? "bg-black/20 text-black font-bold" : "bg-neutral-800 text-neutral-300"
                }`}
              >
                {ativos.filter((a) => a.porte === "MID_CAP" || a.porte === "SMALL_CAP" || a.porte === "FII_CONSOLIDADO").length}
              </span>
            </button>
          </div>
        ) : (
          <div className="flex items-center gap-2">
            <span className="text-[11px] text-neutral-500 uppercase tracking-wider font-mono">Listando ETFs por Liquidez</span>
          </div>
        )}

        {/* Lado Direito: Seletor de Ordenação Dinâmica (Score, Prêmio %, Yield, Margem) */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0">
          <span className="text-[11px] text-neutral-500 uppercase tracking-wider font-mono mr-1">Ordenar por:</span>
          
          <button
            onClick={() => alternarOrdenacao("SCORE")}
            title="Ordenar por Score Fundamentalista Composto"
            className={`px-2.5 py-1 rounded-lg font-medium transition-all flex items-center gap-1 shrink-0 ${
              criterioOrdenacao === "SCORE"
                ? "bg-[#d4af37] text-black font-semibold shadow-sm"
                : "bg-neutral-900 text-neutral-400 hover:text-white border border-neutral-800"
            }`}
          >
            <span>🛡️ Score</span>
            {criterioOrdenacao === "SCORE" && (
              <span className="text-[11px] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
            )}
          </button>

          <button
            onClick={() => alternarOrdenacao("PREMIO")}
            title="Ordenar pela maior % de Prêmio / Upside (diferença entre preço atual e preço teto)"
            className={`px-2.5 py-1 rounded-lg font-medium transition-all flex items-center gap-1 shrink-0 ${
              criterioOrdenacao === "PREMIO"
                ? "bg-emerald-400 text-black font-semibold shadow-sm"
                : "bg-neutral-900 text-neutral-400 hover:text-emerald-400 border border-neutral-800"
            }`}
          >
            <span>🎯 Maior Prêmio (%)</span>
            {criterioOrdenacao === "PREMIO" && (
              <span className="text-[11px] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
            )}
          </button>

          <button
            onClick={() => alternarOrdenacao("SPREAD_NTNB")}
            title="Ordenar pelo maior Spread de Retorno sobre a NTN-B (Tesouro IPCA+)"
            className={`px-2.5 py-1 rounded-lg font-medium transition-all flex items-center gap-1 shrink-0 ${
              criterioOrdenacao === "SPREAD_NTNB"
                ? "bg-cyan-400 text-black font-semibold shadow-sm"
                : "bg-neutral-900 text-neutral-400 hover:text-cyan-400 border border-neutral-800"
            }`}
          >
            <span>⚖️ Risco vs NTN-B</span>
            {criterioOrdenacao === "SPREAD_NTNB" && (
              <span className="text-[11px] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
            )}
          </button>

          <button
            onClick={() => alternarOrdenacao("DY")}
            title="Ordenar por maior Dividend Yield nos últimos 12M"
            className={`px-2.5 py-1 rounded-lg font-medium transition-all flex items-center gap-1 shrink-0 ${
              criterioOrdenacao === "DY"
                ? "bg-[#d4af37] text-black font-semibold shadow-sm"
                : "bg-neutral-900 text-neutral-400 hover:text-white border border-neutral-800"
            }`}
          >
            <span>💰 Yield 12M</span>
            {criterioOrdenacao === "DY" && (
              <span className="text-[11px] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
            )}
          </button>

          <button
            onClick={() => alternarOrdenacao("MARGEM")}
            title="Ordenar por maior Margem de Segurança Consolidada"
            className={`px-2.5 py-1 rounded-lg font-medium transition-all flex items-center gap-1 shrink-0 ${
              criterioOrdenacao === "MARGEM"
                ? "bg-[#d4af37] text-black font-semibold shadow-sm"
                : "bg-neutral-900 text-neutral-400 hover:text-white border border-neutral-800"
            }`}
          >
            <span>📐 Margem</span>
            {criterioOrdenacao === "MARGEM" && (
              <span className="text-[11px] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
            )}
          </button>
        </div>
      </div>

      {/* Conteúdo Desktop: Tabela Institucional */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-neutral-800/60 bg-[#09090d] text-neutral-400 font-mono uppercase tracking-wider text-[11px]">
              <th className="py-3 px-4 w-12 text-center">#</th>
              <th className="py-3 px-4">Ativo</th>
              <th 
                onClick={() => alternarOrdenacao("COTACAO")} 
                className="py-3 px-4 text-right cursor-pointer hover:text-white transition-colors"
                title="Clique para ordenar por Cotação"
              >
                <div className="flex items-center justify-end gap-1">
                  <span>Cotação</span>
                  {criterioOrdenacao === "COTACAO" && (
                    <span className="text-[#d4af37] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                  )}
                </div>
              </th>
              <th 
                onClick={() => alternarOrdenacao("SCORE")} 
                className="py-3 px-4 text-center cursor-pointer hover:text-white transition-colors"
                title="Clique para ordenar por Score Fundamentalista"
              >
                <div className="flex items-center justify-center gap-1">
                  <span>Score</span>
                  {criterioOrdenacao === "SCORE" && (
                    <span className="text-[#d4af37] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                  )}
                </div>
              </th>
              
              {classe === "ACAO" && (
                <>
                  <th className="py-3 px-4 text-right">P/L</th>
                  <th className="py-3 px-4 text-right">P/VP</th>
                  <th className="py-3 px-4 text-right">ROE</th>
                  <th 
                    onClick={() => alternarOrdenacao("DY")} 
                    className="py-3 px-4 text-right cursor-pointer hover:text-white transition-colors"
                    title="Clique para ordenar por Dividend Yield"
                  >
                    <div className="flex items-center justify-end gap-1">
                      <span>Yield 12M</span>
                      {criterioOrdenacao === "DY" && (
                        <span className="text-[#d4af37] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                      )}
                    </div>
                  </th>
                  <th 
                    onClick={() => alternarOrdenacao("SPREAD_NTNB")} 
                    className="py-3 px-4 text-right cursor-pointer hover:text-cyan-400 transition-colors"
                    title="Clique para ordenar por Spread de Retorno sobre a NTN-B (IPCA+ 6.5%)"
                  >
                    <div className="flex items-center justify-end gap-1">
                      <span>Spread NTN-B</span>
                      {criterioOrdenacao === "SPREAD_NTNB" && (
                        <span className="text-cyan-400 font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                      )}
                    </div>
                  </th>
                  <th 
                    onClick={() => alternarOrdenacao("PREMIO")} 
                    className="py-3 px-4 text-right cursor-pointer hover:text-emerald-400 transition-colors bg-emerald-950/10 border-x border-emerald-500/20"
                    title="Clique para ordenar por % do Prêmio / Upside (diferença entre preço atual e preço teto)"
                  >
                    <div className="flex items-center justify-end gap-1 text-emerald-400 font-semibold">
                      <span>🎯 Teto & Prêmio (%)</span>
                      {criterioOrdenacao === "PREMIO" ? (
                        <span className="text-emerald-300 font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                      ) : (
                        <ArrowUpDown className="w-3 h-3 opacity-60" />
                      )}
                    </div>
                  </th>
                  <th className="py-3 px-4 text-right">VI Graham</th>
                </>
              )}

              {classe === "FII" && (
                <>
                  <th className="py-3 px-4 text-right">P/VP CVM</th>
                  <th className="py-3 px-4 text-right">VP Cota</th>
                  <th 
                    onClick={() => alternarOrdenacao("DY")} 
                    className="py-3 px-4 text-right cursor-pointer hover:text-white transition-colors"
                    title="Clique para ordenar por Dividend Yield"
                  >
                    <div className="flex items-center justify-end gap-1">
                      <span>Yield 12M</span>
                      {criterioOrdenacao === "DY" && (
                        <span className="text-[#d4af37] font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                      )}
                    </div>
                  </th>
                  <th 
                    onClick={() => alternarOrdenacao("SPREAD_NTNB")} 
                    className="py-3 px-4 text-right cursor-pointer hover:text-cyan-400 transition-colors"
                    title="Clique para ordenar por Spread sobre o Tesouro IPCA+ (6.5%)"
                  >
                    <div className="flex items-center justify-end gap-1">
                      <span>Spread NTN-B</span>
                      {criterioOrdenacao === "SPREAD_NTNB" && (
                        <span className="text-cyan-400 font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                      )}
                    </div>
                  </th>
                  <th 
                    onClick={() => alternarOrdenacao("PREMIO")} 
                    className="py-3 px-4 text-right cursor-pointer hover:text-emerald-400 transition-colors bg-emerald-950/10 border-x border-emerald-500/20"
                    title="Clique para ordenar por % do Prêmio / Upside (diferença entre preço atual e preço teto)"
                  >
                    <div className="flex items-center justify-end gap-1 text-emerald-400 font-semibold">
                      <span>🎯 Teto & Prêmio (%)</span>
                      {criterioOrdenacao === "PREMIO" ? (
                        <span className="text-emerald-300 font-bold">{direcaoOrdenacao === "DESC" ? "↓" : "↑"}</span>
                      ) : (
                        <ArrowUpDown className="w-3 h-3 opacity-60" />
                      )}
                    </div>
                  </th>
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
                      <div className="flex items-center gap-1.5 flex-wrap">
                        <span className="font-bold text-white font-mono group-hover:text-[#d4af37] transition-colors">
                          {ativo.ticker}
                        </span>
                        {ativo.porte === "BLUE_CHIP" && (
                          <span
                            title="Blue Chip: Gigante de mercado (> R$ 15B) com altíssima liquidez institucional"
                            className="px-1.5 py-0.5 text-[9px] font-bold rounded bg-[#d4af37]/15 text-[#d4af37] border border-[#d4af37]/40 flex items-center gap-1 shrink-0"
                          >
                            👑 Blue Chip
                          </span>
                        )}
                        {ativo.porte === "FII_GIGANTE" && (
                          <span
                            title="FII Baleia: Mais de 150 mil cotistas e alta liquidez no mercado"
                            className="px-1.5 py-0.5 text-[9px] font-bold rounded bg-[#d4af37]/15 text-[#d4af37] border border-[#d4af37]/40 flex items-center gap-1 shrink-0"
                          >
                            🏰 Baleia
                          </span>
                        )}
                        {ativo.porte === "MID_CAP" && (
                          <span
                            title="Mid Cap: Empresa consolidada de médio porte (> R$ 3B)"
                            className="px-1.5 py-0.5 text-[9px] font-medium rounded bg-neutral-800 text-neutral-300 border border-neutral-700 shrink-0"
                          >
                            Mid Cap
                          </span>
                        )}
                        {ativo.is_provento_atipico && (
                          <span
                            title={ativo.alerta_risco || "Provento atípico / Yield Trap"}
                            className="px-1.5 py-0.5 text-[9px] font-bold rounded bg-amber-500/15 text-amber-400 border border-amber-500/30 flex items-center gap-1 cursor-help shrink-0"
                          >
                            ⚠️ Atípico
                          </span>
                        )}
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
                        <td className={`py-3 px-4 text-right font-mono font-semibold tabular-numbers ${ativo.is_provento_atipico ? "text-amber-400" : "text-emerald-400"}`}>
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
                            {ativo.spread_ntnb !== undefined
                              ? `${ativo.spread_ntnb >= 0 ? "+" : ""}${ativo.spread_ntnb.toFixed(2)}%`
                              : "—"}
                          </span>
                          {ativo.veredito_risco && (
                            <span
                              className={`block text-[9px] font-bold ${
                                ativo.veredito_risco === "COMPENSA_RISCO"
                                  ? "text-emerald-400"
                                  : ativo.veredito_risco === "NEUTRO"
                                  ? "text-amber-400"
                                  : "text-rose-400"
                              }`}
                            >
                              {ativo.veredito_risco === "COMPENSA_RISCO"
                                ? "🟢 Compensa"
                                : ativo.veredito_risco === "NEUTRO"
                                ? "🟡 Neutro"
                                : "🔴 Descompensado"}
                            </span>
                          )}
                        </td>
                        <td className="py-3 px-4 text-right font-mono tabular-numbers bg-emerald-950/5 border-x border-emerald-500/10">
                          {ativo.preco_teto_consolidado && ativo.preco_teto_consolidado > 0 ? (
                            <div>
                              <div className="flex items-center justify-end gap-1.5">
                                <span className="text-white font-bold text-xs">{formatCurrency(ativo.preco_teto_consolidado)}</span>
                                {ativo.teve_outlier_5a && (
                                  <span 
                                    title={ativo.observacao_outlier || "Outlier 5A normalizado com Winsorização"} 
                                    className="text-[9px] px-1 py-0.2 rounded bg-amber-500/20 text-amber-400 border border-amber-500/30 cursor-help"
                                  >
                                    🛡️
                                  </span>
                                )}
                              </div>
                              <div className="flex items-center justify-end gap-1 mt-0.5">
                                <span
                                  className={`text-[10px] font-bold px-1.5 py-0.2 rounded-full ${
                                    (ativo.premio_desconto_percentual ?? 0) >= 15
                                      ? "bg-emerald-500/15 text-emerald-400 border border-emerald-500/30"
                                      : (ativo.premio_desconto_percentual ?? 0) >= 0
                                      ? "bg-amber-500/15 text-amber-400 border border-amber-500/30"
                                      : "bg-rose-500/15 text-rose-400 border border-rose-500/30"
                                  }`}
                                >
                                  {(ativo.premio_desconto_percentual ?? 0) >= 0 ? "+" : ""}
                                  {(ativo.premio_desconto_percentual ?? 0).toFixed(1)}% prêmio
                                </span>
                              </div>
                              <span className="block text-[9px] text-neutral-500 font-normal mt-0.5">
                                Bazin 5A: {formatCurrency(ativo.preco_teto_bazin_5a || ativo.preco_teto_bazin)}
                              </span>
                            </div>
                          ) : ativo.preco_teto_bazin_5a && ativo.preco_teto_bazin_5a > 0 ? (
                            <div>
                              <span className="text-[#d4af37] font-semibold">{formatCurrency(ativo.preco_teto_bazin_5a)}</span>
                              <span className="block text-[10px] text-neutral-500 font-normal">12M: {formatCurrency(ativo.preco_teto_bazin)}</span>
                            </div>
                          ) : (
                            formatCurrency(ativo.preco_teto_bazin)
                          )}
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
                        <td className={`py-3 px-4 text-right font-mono font-semibold tabular-numbers ${ativo.is_provento_atipico ? "text-amber-400" : "text-emerald-400"}`}>
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
                            {ativo.spread_ntnb !== undefined
                              ? `${ativo.spread_ntnb >= 0 ? "+" : ""}${ativo.spread_ntnb.toFixed(2)}%`
                              : "—"}
                          </span>
                          {ativo.veredito_risco && (
                            <span
                              className={`block text-[9px] font-bold ${
                                ativo.veredito_risco === "COMPENSA_RISCO"
                                  ? "text-emerald-400"
                                  : ativo.veredito_risco === "NEUTRO"
                                  ? "text-amber-400"
                                  : "text-rose-400"
                              }`}
                            >
                              {ativo.veredito_risco === "COMPENSA_RISCO"
                                ? "🟢 Compensa"
                                : ativo.veredito_risco === "NEUTRO"
                                ? "🟡 Neutro"
                                : "🔴 Descompensado"}
                            </span>
                          )}
                        </td>
                        <td className="py-3 px-4 text-right font-mono tabular-numbers bg-emerald-950/5 border-x border-emerald-500/10">
                          {ativo.preco_teto_consolidado && ativo.preco_teto_consolidado > 0 ? (
                            <div>
                              <div className="flex items-center justify-end gap-1.5">
                                <span className="text-white font-bold text-xs">{formatCurrency(ativo.preco_teto_consolidado)}</span>
                                {ativo.teve_outlier_5a && (
                                  <span 
                                    title={ativo.observacao_outlier || "Outlier 5A normalizado"} 
                                    className="text-[9px] px-1 py-0.2 rounded bg-amber-500/20 text-amber-400 border border-amber-500/30 cursor-help"
                                  >
                                    🛡️
                                  </span>
                                )}
                              </div>
                              <div className="flex items-center justify-end gap-1 mt-0.5">
                                <span
                                  className={`text-[10px] font-bold px-1.5 py-0.2 rounded-full ${
                                    (ativo.premio_desconto_percentual ?? 0) >= 5
                                      ? "bg-emerald-500/15 text-emerald-400 border border-emerald-500/30"
                                      : (ativo.premio_desconto_percentual ?? 0) >= 0
                                      ? "bg-amber-500/15 text-amber-400 border border-amber-500/30"
                                      : "bg-rose-500/15 text-rose-400 border border-rose-500/30"
                                  }`}
                                >
                                  {(ativo.premio_desconto_percentual ?? 0) >= 0 ? "+" : ""}
                                  {(ativo.premio_desconto_percentual ?? 0).toFixed(1)}% prêmio
                                </span>
                              </div>
                              <span className="block text-[9px] text-neutral-500 font-normal mt-0.5">
                                Teto Bazin: {formatCurrency(ativo.preco_teto_bazin_5a || ativo.preco_teto_bazin)}
                              </span>
                            </div>
                          ) : ativo.preco_teto_bazin_5a && ativo.preco_teto_bazin_5a > 0 ? (
                            <div>
                              <span className="text-[#d4af37] font-semibold">{formatCurrency(ativo.preco_teto_bazin_5a)}</span>
                              <span className="block text-[10px] text-neutral-500 font-normal">12M: {formatCurrency(ativo.preco_teto_bazin)}</span>
                            </div>
                          ) : (
                            formatCurrency(ativo.preco_teto_bazin)
                          )}
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
