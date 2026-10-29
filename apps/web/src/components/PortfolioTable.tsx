"use client";

import React, { useState, useEffect } from "react";
import { ClasseAtivo, PosicaoConsolidada, ClasseConsolidada } from "@/types/portfolio";
import {
  Trash2,
  TrendingUp,
  TrendingDown,
  ShieldAlert,
  CheckCircle2,
  Clock,
  PlusCircle,
  History,
  BarChart2,
  ChevronDown,
  ChevronRight,
  Layers,
  List,
  Maximize2,
  Minimize2,
  ShieldCheck,
  Building2,
  Globe,
  Coins,
} from "lucide-react";

interface PortfolioTableProps {
  posicoes: PosicaoConsolidada[];
  classes?: ClasseConsolidada[];
  patrimonioTotal?: number;
  classeFiltroDonut?: ClasseAtivo | null;
  onExcluirPosicao: (id: number, ticker: string) => void;
  onAportar?: (posicao: PosicaoConsolidada) => void;
  onVerHistorico?: (ticker?: string) => void;
  onAbrirGrafico?: (ticker: string) => void;
}

const CLASSE_INFO: Record<
  ClasseAtivo,
  {
    nome: string;
    icon: React.ReactNode;
    cor: string;
    corBg: string;
    corBorda: string;
    corTexto: string;
  }
> = {
  ACAO_BR: {
    nome: "Ações Brasil",
    icon: <TrendingUp className="w-4 h-4 text-emerald-400" />,
    cor: "#10b981",
    corBg: "bg-emerald-500/10",
    corBorda: "border-emerald-500/30",
    corTexto: "text-emerald-400",
  },
  FII: {
    nome: "Fundos Imobiliários",
    icon: <Building2 className="w-4 h-4 text-purple-400" />,
    cor: "#a855f7",
    corBg: "bg-purple-500/10",
    corBorda: "border-purple-500/30",
    corTexto: "text-purple-400",
  },
  RENDA_FIXA: {
    nome: "Renda Fixa",
    icon: <ShieldCheck className="w-4 h-4 text-blue-400" />,
    cor: "#3b82f6",
    corBg: "bg-blue-500/10",
    corBorda: "border-blue-500/30",
    corTexto: "text-blue-400",
  },
  INTERNACIONAL: {
    nome: "Internacional / ETFs",
    icon: <Globe className="w-4 h-4 text-cyan-400" />,
    cor: "#06b6d4",
    corBg: "bg-cyan-500/10",
    corBorda: "border-cyan-500/30",
    corTexto: "text-cyan-400",
  },
  CAIXA: {
    nome: "Caixa / Oportunidade",
    icon: <Coins className="w-4 h-4 text-amber-400" />,
    cor: "#f59e0b",
    corBg: "bg-amber-500/10",
    corBorda: "border-amber-500/30",
    corTexto: "text-amber-400",
  },
};

export function PortfolioTable({
  posicoes,
  classes,
  patrimonioTotal = 0,
  classeFiltroDonut,
  onExcluirPosicao,
  onAportar,
  onVerHistorico,
  onAbrirGrafico,
}: PortfolioTableProps) {
  const [modoHierarquico, setModoHierarquico] = useState(true);
  const [filtroClasse, setFiltroClasse] = useState<ClasseAtivo | "TODAS">("TODAS");
  const [busca, setBusca] = useState("");
  // Estado de expansão das classes no modo hierárquico
  const [classesExpandidas, setClassesExpandidas] = useState<Record<string, boolean>>({
    ACAO_BR: true,
    FII: true,
    RENDA_FIXA: true,
    INTERNACIONAL: true,
    CAIXA: true,
  });

  // Se o usuário clicou em uma fatia do Donut, foca e expande a respectiva classe
  useEffect(() => {
    if (classeFiltroDonut) {
      setFiltroClasse(classeFiltroDonut);
      setClassesExpandidas((prev) => ({
        ...prev,
        [classeFiltroDonut]: true,
      }));
    } else {
      setFiltroClasse("TODAS");
    }
  }, [classeFiltroDonut]);

  // Se o usuário digitou uma busca, expande todas as classes automaticamente
  useEffect(() => {
    if (busca && busca.trim().length > 0) {
      setClassesExpandidas({
        ACAO_BR: true,
        FII: true,
        RENDA_FIXA: true,
        INTERNACIONAL: true,
        CAIXA: true,
      });
    }
  }, [busca]);

  const toggleClasse = (classe: string) => {
    setClassesExpandidas((prev) => ({
      ...prev,
      [classe]: !prev[classe],
    }));
  };

  const expandirTodas = (expandir: boolean) => {
    setClassesExpandidas({
      ACAO_BR: expandir,
      FII: expandir,
      RENDA_FIXA: expandir,
      INTERNACIONAL: expandir,
      CAIXA: expandir,
    });
  };

  const formatarMoeda = (val: number) =>
    new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(val);

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

  // Monta classes agrupadas caso o backend não tenha enviado
  const classesOrdem: ClasseAtivo[] = ["ACAO_BR", "FII", "RENDA_FIXA", "INTERNACIONAL", "CAIXA"];

  const classesParaExibir = classesOrdem.map((c) => {
    const info = CLASSE_INFO[c];
    const objBackend = classes?.find((cl) => cl.classe === c);

    const posicoesDaClasse = posicoes
      .filter((p) => p.classe === c)
      .filter((p) => {
        if (!busca || !busca.trim()) return true;
        const termo = busca.toLowerCase().trim();
        return (
          (p.ticker && p.ticker.toLowerCase().includes(termo)) ||
          (p.corretora && p.corretora.toLowerCase().includes(termo)) ||
          (p.categoria_macro && p.categoria_macro.toLowerCase().includes(termo))
        );
      });

    const valorAtualMercado =
      objBackend?.valor_atual_mercado ??
      posicoesDaClasse.reduce((acc, p) => acc + p.valor_atual_mercado, 0);

    const totalInvestido =
      objBackend?.total_investido ??
      posicoesDaClasse.reduce((acc, p) => acc + p.valor_aplicado, 0);

    const lucroTotal = valorAtualMercado - totalInvestido;
    const rentabPct = totalInvestido > 0 ? (lucroTotal / totalInvestido) * 100 : 0;
    const pctCarteira =
      patrimonioTotal > 0 ? (valorAtualMercado / patrimonioTotal) * 100 : 0;

    return {
      classe: c,
      nome: info.nome,
      info,
      valorAtualMercado,
      totalInvestido,
      lucroTotal,
      rentabPct,
      pctCarteira,
      posicoes: posicoesDaClasse,
      expandida: !!classesExpandidas[c],
    };
  });

  // Lista plana tradicional para quando o modo hierárquico estiver desligado
  const posicoesFiltradasPlanas = posicoes.filter((p) => {
    if (filtroClasse !== "TODAS" && p.classe !== filtroClasse) return false;
    if (busca && busca.trim()) {
      const termo = busca.toLowerCase().trim();
      return (
        (p.ticker && p.ticker.toLowerCase().includes(termo)) ||
        (p.corretora && p.corretora.toLowerCase().includes(termo)) ||
        (p.categoria_macro && p.categoria_macro.toLowerCase().includes(termo))
      );
    }
    return true;
  });

  return (
    <div className="rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-2xl overflow-hidden backdrop-blur-sm">
      {/* Barra de Filtros, Modos de Visão e Busca */}
      <div className="p-4 sm:p-5 border-b border-neutral-800/80 flex flex-col md:flex-row md:items-center justify-between gap-4">
        {/* Lado Esquerdo: Seletor de Modo & Abas */}
        <div className="flex items-center gap-2 flex-wrap">
          {/* Alternador de Modo: Hierárquico vs Lista Plana */}
          <div className="flex items-center bg-neutral-900 p-1 rounded-xl border border-neutral-800">
            <button
              onClick={() => setModoHierarquico(true)}
              className={`px-3 py-1.5 rounded-lg text-xs font-bold flex items-center gap-1.5 transition-all ${
                modoHierarquico
                  ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20"
                  : "text-neutral-400 hover:text-white"
              }`}
              title="Visão Hierárquica: Agrupada por Classes com expansão de ativos"
            >
              <Layers className="w-3.5 h-3.5" />
              <span>Visão em Níveis</span>
            </button>
            <button
              onClick={() => setModoHierarquico(false)}
              className={`px-3 py-1.5 rounded-lg text-xs font-bold flex items-center gap-1.5 transition-all ${
                !modoHierarquico
                  ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20"
                  : "text-neutral-400 hover:text-white"
              }`}
              title="Visão Plana: Todos os ativos em lista contínua"
            >
              <List className="w-3.5 h-3.5" />
              <span>Lista Completa</span>
            </button>
          </div>

          {/* Abas de Classes para Modo Plano ou Filtro Rápido */}
          <div className="flex items-center gap-1.5 overflow-x-auto scrollbar-none">
            {(["TODAS", "ACAO_BR", "FII", "RENDA_FIXA", "INTERNACIONAL", "CAIXA"] as const).map(
              (aba) => (
                <button
                  key={aba}
                  onClick={() => setFiltroClasse(aba)}
                  className={`px-2.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-all ${
                    filtroClasse === aba
                      ? "bg-neutral-800 text-white font-bold border border-neutral-700"
                      : "bg-neutral-950/60 hover:bg-neutral-900 text-neutral-400 border border-neutral-800/80"
                  }`}
                >
                  {aba === "TODAS"
                    ? `Todos (${posicoes.length})`
                    : aba === "ACAO_BR"
                    ? "Ações"
                    : aba === "FII"
                    ? "FIIs"
                    : aba === "RENDA_FIXA"
                    ? "R. Fixa"
                    : aba === "INTERNACIONAL"
                    ? "Internac."
                    : "Caixa"}
                </button>
              )
            )}
          </div>
        </div>

        {/* Lado Direito: Ações de Expansão & Input de Busca */}
        <div className="flex items-center gap-2">
          {modoHierarquico && (
            <div className="flex items-center gap-1 bg-neutral-950 p-1 rounded-xl border border-neutral-800">
              <button
                onClick={() => expandirTodas(true)}
                className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-900 text-xs transition-colors"
                title="Expandir todas as classes"
              >
                <Maximize2 className="w-3.5 h-3.5" />
              </button>
              <button
                onClick={() => expandirTodas(false)}
                className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-900 text-xs transition-colors"
                title="Recolher todas as classes"
              >
                <Minimize2 className="w-3.5 h-3.5" />
              </button>
            </div>
          )}

          <div className="w-full sm:w-56">
            <input
              type="text"
              placeholder="Buscar por ticker, corretora..."
              value={busca}
              onChange={(e) => setBusca(e.target.value)}
              className="w-full px-3.5 py-1.5 rounded-xl bg-neutral-950 border border-neutral-800 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-[#d4af37]/60"
            />
          </div>
        </div>
      </div>

      {/* ===================================================================== */}
      {/* MODO 1: VISÃO HIERÁRQUICA EM NÍVEIS (ACCORDION POR CLASSE)            */}
      {/* ===================================================================== */}
      {modoHierarquico ? (
        <div className="divide-y divide-neutral-800/80">
          {classesParaExibir
            .filter((cl) => filtroClasse === "TODAS" || cl.classe === filtroClasse)
            .map((cl) => {
              const temAtivos = cl.posicoes.length > 0;
              const isPositivoClasse = cl.lucroTotal >= 0;

              return (
                <div key={cl.classe} className="bg-neutral-950/40 transition-colors">
                  {/* Linha da Classe (Nível 1) - Clicável */}
                  <div
                    onClick={() => toggleClasse(cl.classe)}
                    className="p-4 sm:px-6 cursor-pointer hover:bg-neutral-900/50 flex flex-col md:flex-row md:items-center justify-between gap-3 border-l-4 transition-all select-none"
                    style={{
                      borderLeftColor: cl.expandida ? cl.info.cor : "transparent",
                    }}
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className="w-6 h-6 rounded-lg flex items-center justify-center transition-transform"
                        style={{ color: cl.info.cor }}
                      >
                        {cl.expandida ? (
                          <ChevronDown className="w-5 h-5" />
                        ) : (
                          <ChevronRight className="w-5 h-5 text-neutral-500" />
                        )}
                      </div>

                      <div
                        className="w-8 h-8 rounded-xl flex items-center justify-center"
                        style={{ backgroundColor: `${cl.info.cor}18` }}
                      >
                        {cl.info.icon}
                      </div>

                      <div>
                        <div className="flex items-center gap-2">
                          <h4 className="text-sm font-extrabold text-white tracking-wide">
                            {cl.nome}
                          </h4>
                          <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-neutral-900 text-neutral-400 border border-neutral-800">
                            {cl.posicoes.length} {cl.posicoes.length === 1 ? "ativo" : "ativos"}
                          </span>
                        </div>
                        <p className="text-[11px] text-neutral-400 font-mono mt-0.5">
                          Investido: {formatarMoeda(cl.totalInvestido)}
                        </p>
                      </div>
                    </div>

                    {/* Métricas Consolidadas da Classe */}
                    <div className="flex items-center gap-6 self-end md:self-auto pl-9 md:pl-0">
                      {/* Lucro / Rentabilidade da Classe */}
                      <div className="text-right font-mono">
                        <div
                          className={`text-xs font-bold ${
                            isPositivoClasse ? "text-emerald-400" : "text-red-400"
                          }`}
                        >
                          {isPositivoClasse ? "+" : ""}
                          {formatarMoeda(cl.lucroTotal)}
                        </div>
                        <div
                          className={`text-[10px] ${
                            isPositivoClasse ? "text-emerald-400" : "text-red-400"
                          }`}
                        >
                          {isPositivoClasse ? "+" : ""}
                          {cl.rentabPct.toFixed(2)}%
                        </div>
                      </div>

                      {/* Valor de Mercado & % da Carteira */}
                      <div className="text-right font-mono min-w-[140px]">
                        <div className="text-sm font-black text-white">
                          {formatarMoeda(cl.valorAtualMercado)}
                        </div>
                        <div className="flex items-center justify-end gap-1.5 mt-0.5">
                          <span
                            className="text-xs font-bold"
                            style={{ color: cl.info.cor }}
                          >
                            {cl.pctCarteira.toFixed(1)}%
                          </span>
                          <span className="text-[10px] text-neutral-400">da carteira</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Sub-Tabela com os Ativos da Classe (Nível 2) */}
                  {cl.expandida && (
                    <div className="border-t border-neutral-800/60 bg-[#070709]/80 overflow-x-auto">
                      {!temAtivos ? (
                        <div className="py-8 text-center text-xs text-neutral-500 font-mono">
                          Nenhum ativo cadastrado nesta classe no momento.
                        </div>
                      ) : (
                        <table className="w-full text-left border-collapse text-xs">
                          <thead>
                            <tr className="border-b border-neutral-800/60 text-neutral-400 font-mono uppercase text-[10px] bg-neutral-950/70">
                              <th className="py-2.5 px-4 font-semibold">Ativo / Corretora</th>
                              <th className="py-2.5 px-4 font-semibold text-right">Qtd / Preço Médio</th>
                              <th className="py-2.5 px-4 font-semibold text-right">Cotação Atual</th>
                              <th className="py-2.5 px-4 font-semibold text-right">Valor Atual</th>
                              <th className="py-2.5 px-4 font-semibold text-right min-w-[130px]">
                                % na Classe
                              </th>
                              <th className="py-2.5 px-4 font-semibold text-right">
                                % na Carteira
                              </th>
                              <th className="py-2.5 px-4 font-semibold text-right">Lucro / Retorno</th>
                              <th className="py-2.5 px-4 font-semibold text-right">Preço Teto / Upside</th>
                              <th className="py-2.5 px-4 font-semibold text-center">Score</th>
                              <th className="py-2.5 px-4 font-semibold text-center">Sinal</th>
                              <th className="py-2.5 px-4 font-semibold text-center">Ações</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-neutral-800/40">
                            {cl.posicoes.map((p) => {
                              const isPositivo = p.lucro_prejuizo_valor >= 0;
                              const pctClasse =
                                p.percentual_na_classe ??
                                (cl.valorAtualMercado > 0
                                  ? (p.valor_atual_mercado / cl.valorAtualMercado) * 100
                                  : 0);
                              const pctTotal =
                                p.percentual_na_carteira ??
                                (patrimonioTotal > 0
                                  ? (p.valor_atual_mercado / patrimonioTotal) * 100
                                  : 0);

                              return (
                                <tr
                                  key={p.id}
                                  className="hover:bg-neutral-800/30 transition-colors group"
                                >
                                  {/* Ativo / Corretora */}
                                  <td className="py-3 px-4">
                                    <div className="flex items-center gap-2">
                                      <span className="font-extrabold text-white text-sm font-mono tracking-wide">
                                        {p.ticker}
                                      </span>
                                    </div>
                                    <div className="flex items-center gap-2 text-[10px] text-neutral-400 font-mono mt-0.5">
                                      <span>{p.corretora}</span>
                                      {p.taxa_contratada && (
                                        <>
                                          <span>•</span>
                                          <span className="text-cyan-400 font-semibold">
                                            {p.taxa_contratada}
                                          </span>
                                        </>
                                      )}
                                    </div>
                                  </td>

                                  {/* Qtd & PM */}
                                  <td className="py-3 px-4 text-right font-mono">
                                    <div className="text-white font-medium">
                                      {p.quantidade.toLocaleString("pt-BR", {
                                        maximumFractionDigits: 4,
                                      })}
                                    </div>
                                    <div className="text-[10px] text-neutral-400">
                                      PM: {formatarMoeda(p.preco_medio)}
                                    </div>
                                  </td>

                                  {/* Cotação Atual */}
                                  <td className="py-3 px-4 text-right font-mono">
                                    <div className="font-bold text-white">
                                      {formatarMoeda(p.cotacao_atual)}
                                    </div>
                                    <span className="text-[9px] text-neutral-400">B3 Oficial</span>
                                  </td>

                                  {/* Valor Atual */}
                                  <td className="py-3 px-4 text-right font-mono font-bold text-white">
                                    {formatarMoeda(p.valor_atual_mercado)}
                                  </td>

                                  {/* % na Classe (Exposição Relativa) */}
                                  <td className="py-3 px-4 text-right font-mono">
                                    <div className="flex items-center justify-end gap-2">
                                      <span
                                        className="font-bold text-xs"
                                        style={{ color: cl.info.cor }}
                                      >
                                        {pctClasse.toFixed(1)}%
                                      </span>
                                    </div>
                                    <div className="w-full h-1 bg-neutral-900 rounded-full mt-1.5 overflow-hidden flex">
                                      <div
                                        className="h-full rounded-full transition-all"
                                        style={{
                                          width: `${Math.min(100, pctClasse)}%`,
                                          backgroundColor: cl.info.cor,
                                        }}
                                      />
                                    </div>
                                  </td>

                                  {/* % na Carteira (Exposição Absoluta) */}
                                  <td className="py-3 px-4 text-right font-mono">
                                    <span className="text-xs font-bold text-neutral-200 bg-neutral-900 px-2 py-0.5 rounded border border-neutral-800">
                                      {pctTotal.toFixed(1)}%
                                    </span>
                                  </td>

                                  {/* Lucro e Retorno */}
                                  <td className="py-3 px-4 text-right font-mono">
                                    <div
                                      className={`font-extrabold ${
                                        isPositivo ? "text-emerald-400" : "text-red-400"
                                      }`}
                                    >
                                      {isPositivo ? "+" : ""}
                                      {formatarMoeda(p.lucro_prejuizo_valor)}
                                    </div>
                                    <div
                                      className={`text-[10px] flex items-center justify-end gap-0.5 ${
                                        isPositivo ? "text-emerald-400" : "text-red-400"
                                      }`}
                                    >
                                      {isPositivo ? (
                                        <TrendingUp className="w-3 h-3" />
                                      ) : (
                                        <TrendingDown className="w-3 h-3" />
                                      )}
                                      {isPositivo ? "+" : ""}
                                      {p.rentabilidade_pct.toFixed(2)}%
                                    </div>
                                  </td>

                                  {/* Preço Teto / Upside */}
                                  <td className="py-3 px-4 text-right font-mono">
                                    {p.preco_teto_estimado > 0 ? (
                                      <>
                                        <div className="text-neutral-200">
                                          {formatarMoeda(p.preco_teto_estimado)}
                                        </div>
                                        <div
                                          className={`text-[10px] font-bold ${
                                            p.margem_seguranca_pct >= 0
                                              ? "text-emerald-400"
                                              : "text-amber-400"
                                          }`}
                                        >
                                          {p.margem_seguranca_pct >= 0 ? "+" : ""}
                                          {p.margem_seguranca_pct.toFixed(1)}%
                                        </div>
                                      </>
                                    ) : (
                                      <span className="text-neutral-500 text-[10px]">-</span>
                                    )}
                                  </td>

                                  {/* Score */}
                                  <td className="py-3 px-4 text-center font-mono">
                                    {p.score_fundamentalista > 0 ? (
                                      <span
                                        className={`px-2 py-0.5 rounded text-[11px] font-bold ${
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
                                      <span className="text-neutral-500 font-mono text-[10px]">
                                        N/A
                                      </span>
                                    )}
                                  </td>

                                  {/* Sinal */}
                                  <td className="py-3 px-4 text-center">
                                    {getBadgeSinal(p.sinal_rebalanceamento)}
                                  </td>

                                  {/* Ações */}
                                  <td className="py-3 px-4 text-center">
                                    <div className="flex items-center justify-center gap-1.5">
                                      {p.classe !== "RENDA_FIXA" && (
                                        <button
                                          onClick={() => onAbrirGrafico?.(p.ticker)}
                                          className="p-1.5 rounded-lg text-neutral-400 hover:text-[#d4af37] hover:bg-[#d4af37]/10 transition-colors"
                                          title="Abrir gráfico técnico de velas e médias"
                                        >
                                          <BarChart2 className="w-3.5 h-3.5" />
                                        </button>
                                      )}
                                      <button
                                        onClick={() => onAportar?.(p)}
                                        className="px-2 py-1 rounded-lg bg-emerald-500/15 hover:bg-emerald-500/25 text-emerald-400 border border-emerald-500/30 text-[11px] font-semibold flex items-center gap-1 transition-colors"
                                        title="Fazer novo aporte (recalcular preço médio)"
                                      >
                                        <PlusCircle className="w-3 h-3" />
                                        <span className="hidden sm:inline">Aporte</span>
                                      </button>
                                      <button
                                        onClick={() => onVerHistorico?.(p.ticker)}
                                        className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors"
                                        title="Ver histórico de transações"
                                      >
                                        <History className="w-3.5 h-3.5" />
                                      </button>
                                      <button
                                        onClick={() => onExcluirPosicao(p.id, p.ticker)}
                                        className="p-1.5 rounded-lg text-neutral-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                                        title="Remover ativo da carteira"
                                      >
                                        <Trash2 className="w-3.5 h-3.5" />
                                      </button>
                                    </div>
                                  </td>
                                </tr>
                              );
                            })}
                          </tbody>
                        </table>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
        </div>
      ) : (
        /* ===================================================================== */
        /* MODO 2: VISÃO PLANA (LISTA CONTÍNUA TRADICIONAL)                      */
        /* ===================================================================== */
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-neutral-800/80 bg-neutral-950/60 text-neutral-400 font-mono uppercase text-[11px]">
                <th className="py-3.5 px-4 font-semibold">Ativo / Corretora</th>
                <th className="py-3.5 px-4 font-semibold text-right">Qtd / Preço Médio</th>
                <th className="py-3.5 px-4 font-semibold text-right">Cotação Atual</th>
                <th className="py-3.5 px-4 font-semibold text-right">Valor Mercado</th>
                <th className="py-3.5 px-4 font-semibold text-right">% na Classe</th>
                <th className="py-3.5 px-4 font-semibold text-right">% no Total</th>
                <th className="py-3.5 px-4 font-semibold text-right">Lucro / Retorno</th>
                <th className="py-3.5 px-4 font-semibold text-right">Preço Teto / Upside</th>
                <th className="py-3.5 px-4 font-semibold text-center">Score</th>
                <th className="py-3.5 px-4 font-semibold text-center">Sinal Aporte</th>
                <th className="py-3.5 px-4 font-semibold text-center">Ações</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-800/50">
              {posicoesFiltradasPlanas.length === 0 ? (
                <tr>
                  <td colSpan={11} className="py-12 text-center text-neutral-500 font-mono">
                    Nenhum ativo encontrado para o filtro selecionado.
                  </td>
                </tr>
              ) : (
                posicoesFiltradasPlanas.map((p) => {
                  const isPositivo = p.lucro_prejuizo_valor >= 0;
                  const pctTotal =
                    p.percentual_na_carteira ??
                    (patrimonioTotal > 0 ? (p.valor_atual_mercado / patrimonioTotal) * 100 : 0);

                  return (
                    <tr
                      key={p.id}
                      className="hover:bg-neutral-800/25 transition-colors group"
                    >
                      <td className="py-3.5 px-4">
                        <div className="flex items-center gap-2">
                          <span className="font-extrabold text-white text-sm font-mono tracking-wide">
                            {p.ticker}
                          </span>
                          <span
                            className="text-[10px] font-mono px-2 py-0.5 rounded border"
                            style={{
                              backgroundColor: `${CLASSE_INFO[p.classe]?.cor}18`,
                              borderColor: `${CLASSE_INFO[p.classe]?.cor}40`,
                              color: CLASSE_INFO[p.classe]?.cor,
                            }}
                          >
                            {CLASSE_INFO[p.classe]?.nome}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 text-[11px] text-neutral-400 font-mono mt-0.5">
                          <span>{p.corretora}</span>
                        </div>
                      </td>

                      <td className="py-3.5 px-4 text-right font-mono">
                        <div className="text-white font-medium">
                          {p.quantidade.toLocaleString("pt-BR", { maximumFractionDigits: 4 })}
                        </div>
                        <div className="text-[11px] text-neutral-400">
                          PM: {formatarMoeda(p.preco_medio)}
                        </div>
                      </td>

                      <td className="py-3.5 px-4 text-right font-mono font-bold text-white">
                        {formatarMoeda(p.cotacao_atual)}
                      </td>

                      <td className="py-3.5 px-4 text-right font-mono font-bold text-white">
                        {formatarMoeda(p.valor_atual_mercado)}
                      </td>

                      <td className="py-3.5 px-4 text-right font-mono font-bold">
                        {p.percentual_na_classe ? `${p.percentual_na_classe.toFixed(1)}%` : "-"}
                      </td>

                      <td className="py-3.5 px-4 text-right font-mono font-bold text-neutral-300">
                        {pctTotal.toFixed(1)}%
                      </td>

                      <td className="py-3.5 px-4 text-right font-mono">
                        <div
                          className={`font-extrabold ${
                            isPositivo ? "text-emerald-400" : "text-red-400"
                          }`}
                        >
                          {isPositivo ? "+" : ""}
                          {formatarMoeda(p.lucro_prejuizo_valor)}
                        </div>
                        <div
                          className={`text-[11px] flex items-center justify-end gap-0.5 ${
                            isPositivo ? "text-emerald-400" : "text-red-400"
                          }`}
                        >
                          {isPositivo ? (
                            <TrendingUp className="w-3 h-3" />
                          ) : (
                            <TrendingDown className="w-3 h-3" />
                          )}
                          {isPositivo ? "+" : ""}
                          {p.rentabilidade_pct.toFixed(2)}%
                        </div>
                      </td>

                      <td className="py-3.5 px-4 text-right font-mono">
                        {p.preco_teto_estimado > 0 ? (
                          <>
                            <div className="text-neutral-200">
                              {formatarMoeda(p.preco_teto_estimado)}
                            </div>
                            <div
                              className={`text-[11px] font-bold ${
                                p.margem_seguranca_pct >= 0 ? "text-emerald-400" : "text-amber-400"
                              }`}
                            >
                              {p.margem_seguranca_pct >= 0 ? "+" : ""}
                              {p.margem_seguranca_pct.toFixed(1)}%
                            </div>
                          </>
                        ) : (
                          <span className="text-neutral-500 text-[11px]">-</span>
                        )}
                      </td>

                      <td className="py-3.5 px-4 text-center font-mono">
                        {p.score_fundamentalista > 0 ? (
                          <span className="px-2 py-0.5 rounded text-xs font-bold bg-neutral-800 text-neutral-200">
                            {p.score_fundamentalista}
                          </span>
                        ) : (
                          <span className="text-neutral-500 text-[11px]">N/A</span>
                        )}
                      </td>

                      <td className="py-3.5 px-4 text-center">
                        {getBadgeSinal(p.sinal_rebalanceamento)}
                      </td>

                      <td className="py-3.5 px-4 text-center">
                        <div className="flex items-center justify-center gap-1.5">
                          {p.classe !== "RENDA_FIXA" && (
                            <button
                              onClick={() => onAbrirGrafico?.(p.ticker)}
                              className="p-1.5 rounded-lg text-neutral-400 hover:text-[#d4af37] hover:bg-[#d4af37]/10 transition-colors"
                              title="Abrir gráfico técnico"
                            >
                              <BarChart2 className="w-4 h-4" />
                            </button>
                          )}
                          <button
                            onClick={() => onAportar?.(p)}
                            className="px-2 py-1 rounded-lg bg-emerald-500/15 hover:bg-emerald-500/25 text-emerald-400 border border-emerald-500/30 text-xs font-semibold flex items-center gap-1 transition-colors"
                          >
                            <PlusCircle className="w-3.5 h-3.5" />
                            <span className="hidden sm:inline">Aporte</span>
                          </button>
                          <button
                            onClick={() => onVerHistorico?.(p.ticker)}
                            className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors"
                          >
                            <History className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => onExcluirPosicao(p.id, p.ticker)}
                            className="p-1.5 rounded-lg text-neutral-500 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
