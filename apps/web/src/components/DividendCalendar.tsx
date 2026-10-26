"use client";

import React, { useState, useEffect, useMemo, useCallback } from "react";
import {
  Calendar as CalendarIcon,
  Filter,
  Search,
  CheckCircle2,
  Clock,
  Sparkles,
  ArrowUpDown,
  TrendingUp,
  Building2,
  Layers,
  Info,
  DollarSign,
  AlertCircle,
  HelpCircle,
} from "lucide-react";
import { API_V1_URL } from "@/config/api";
import { useAuth } from "@/context/AuthContext";
import { ResumoCalendario, AtivoCalendario, MesProvento } from "@/types/calendar";
import { Ativo } from "@/types/market";

interface DividendCalendarProps {
  onSelecionarAtivo?: (ativo: Ativo) => void;
  todosAtivosMercado?: Ativo[];
}

const NOMES_MESES = [
  "Jan", "Fev", "Mar", "Abr", "Mai", "Jun",
  "Jul", "Ago", "Set", "Out", "Nov", "Dez"
];

const NOMES_MESES_COMPLETOS = [
  "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
  "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"
];

export function DividendCalendar({
  onSelecionarAtivo,
  todosAtivosMercado = [],
}: DividendCalendarProps) {
  const { usuario, token } = useAuth();

  // Filtros
  const [tipoData, setTipoData] = useState<"DATA_COM" | "DATA_PAGAMENTO">("DATA_COM");
  const [apenasCarteira, setApenasCarteira] = useState<boolean>(false);
  const [classe, setClasse] = useState<string>("");
  const [probMin, setProbMin] = useState<number>(0);
  const [busca, setBusca] = useState<string>("");
  const [setorSelecionado, setSetorSelecionado] = useState<string>("");

  // Estado dos dados
  const [calendario, setCalendario] = useState<ResumoCalendario | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [erro, setErro] = useState<string | null>(null);

  // Tooltip / Detalhes flutuantes
  const [celulaAtiva, setCelulaAtiva] = useState<{
    ativo: AtivoCalendario;
    mes: MesProvento;
  } | null>(null);

  // Mapeamento de Ativos por ticker para o Modal de Detalhes
  const mapaAtivos = useMemo(() => {
    const mapa = new Map<string, Ativo>();
    for (const at of todosAtivosMercado) {
      mapa.set(at.ticker.toUpperCase(), at);
    }
    return mapa;
  }, [todosAtivosMercado]);

  // Carregar dados da API
  const carregarCalendario = useCallback(async () => {
    setLoading(true);
    setErro(null);
    try {
      const params = new URLSearchParams();
      params.set("tipo_data", tipoData);
      if (probMin > 0) params.set("prob_min", probMin.toString());
      if (apenasCarteira) params.set("apenas_carteira", "true");
      if (classe) params.set("classe", classe);
      if (setorSelecionado) params.set("setor", setorSelecionado);
      if (busca.trim()) params.set("busca", busca.trim());

      const headers: Record<string, string> = {
        Accept: "application/json",
      };
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const res = await fetch(`${API_V1_URL}/proventos/calendario?${params.toString()}`, {
        headers,
      });

      if (!res.ok) {
        throw new Error(`Falha ao obter calendário (status ${res.status})`);
      }

      const json: ResumoCalendario = await res.json();
      setCalendario(json);
    } catch (err: any) {
      console.warn("Erro ao buscar calendário:", err);
      setErro(err.message || "Não foi possível carregar a matriz preditiva de proventos.");
    } finally {
      setLoading(false);
    }
  }, [tipoData, probMin, apenasCarteira, classe, setorSelecionado, busca, token]);

  useEffect(() => {
    const timer = setTimeout(() => {
      carregarCalendario();
    }, 200);
    return () => clearTimeout(timer);
  }, [carregarCalendario]);

  // Lista única de setores para o seletor
  const listaSetores = useMemo(() => {
    if (!calendario?.ativos) return [];
    const set = new Set<string>();
    for (const a of calendario.ativos) {
      if (a.setor && a.setor.trim() !== "") {
        set.add(a.setor.trim());
      }
    }
    return Array.from(set).sort();
  }, [calendario?.ativos]);

  // Estilo visual da probabilidade
  const obterEstiloProbabilidade = (prob: number) => {
    if (prob <= 0) {
      return "bg-neutral-900/30 text-neutral-600 border border-transparent";
    }
    if (prob >= 80) {
      return "bg-emerald-500/90 hover:bg-emerald-400 text-neutral-950 font-black border border-emerald-400/80 shadow-md shadow-emerald-500/20";
    }
    if (prob >= 60) {
      return "bg-emerald-600/70 hover:bg-emerald-500 text-white font-bold border border-emerald-500/50 shadow-sm shadow-emerald-700/20";
    }
    if (prob >= 40) {
      return "bg-emerald-800/50 hover:bg-emerald-700/60 text-emerald-100 font-semibold border border-emerald-600/30";
    }
    return "bg-emerald-950/40 hover:bg-emerald-900/50 text-emerald-300/80 font-medium border border-emerald-800/20";
  };

  const handleRowClick = (ticker: string) => {
    if (!onSelecionarAtivo) return;
    const ativo = mapaAtivos.get(ticker.toUpperCase());
    if (ativo) {
      onSelecionarAtivo(ativo);
    }
  };

  return (
    <div className="space-y-6">
      {/* 1. Header do MDI com Explicação de Metodologia */}
      <div className="p-6 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-xl backdrop-blur-sm relative overflow-hidden">
        <div className="absolute top-0 right-0 w-80 h-80 bg-gradient-to-br from-[#d4af37]/10 via-emerald-500/5 to-transparent rounded-bl-full pointer-events-none" />
        
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 relative z-10">
          <div>
            <div className="flex items-center gap-2.5 mb-2">
              <span className="p-2 rounded-xl bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/30">
                <CalendarIcon className="w-5 h-5" />
              </span>
              <h2 className="text-xl md:text-2xl font-black text-white tracking-wide font-mono flex items-center gap-2">
                MAPA DE DIVIDENDOS INTELIGENTE <span className="text-[#d4af37]">(MDI)</span>
              </h2>
            </div>
            <p className="text-xs md:text-sm text-neutral-400 max-w-3xl leading-relaxed">
              Matriz preditiva de sazonalidade e probabilidade histórica calculada sobre os proventos oficiais da{" "}
              <strong className="text-neutral-200">B3 Oficial dos últimos 5 anos (2021 a 2026)</strong>. Identifique com exatidão quando as empresas anunciam proventos para antecipar compras antes do fechamento do pregão.
            </p>
          </div>

          {/* Toggle Data COM vs Data Pagamento */}
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 shrink-0">
            <div className="flex items-center p-1 rounded-xl bg-neutral-900 border border-neutral-800 shadow-inner">
              <button
                onClick={() => setTipoData("DATA_COM")}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                  tipoData === "DATA_COM"
                    ? "bg-[#d4af37] text-neutral-950 shadow-md shadow-[#d4af37]/20 font-black"
                    : "text-neutral-400 hover:text-white"
                }`}
                title="Data com direito: último dia em que você precisa ter a ação na carteira"
              >
                <Sparkles className="w-3.5 h-3.5" />
                <span>DATA COM (Direito)</span>
              </button>
              <button
                onClick={() => setTipoData("DATA_PAGAMENTO")}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                  tipoData === "DATA_PAGAMENTO"
                    ? "bg-[#d4af37] text-neutral-950 shadow-md shadow-[#d4af37]/20 font-black"
                    : "text-neutral-400 hover:text-white"
                }`}
                title="Data de Pagamento: quando os dividendos caem na conta"
              >
                <Clock className="w-3.5 h-3.5" />
                <span>DATA PAGAMENTO</span>
              </button>
            </div>
          </div>
        </div>

        {/* Barra de Cards Resumo */}
        {calendario && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-6 pt-5 border-t border-neutral-800/80">
            {/* Modo de Análise */}
            <div className="p-3.5 rounded-xl bg-neutral-950/60 border border-neutral-800">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-neutral-400 font-mono block mb-1">
                Critério de Data
              </span>
              <div className="text-sm font-bold text-white flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                {tipoData === "DATA_COM" ? "Data COM (Direito)" : "Data de Pagamento"}
              </div>
              <span className="text-[10px] text-neutral-500 block mt-0.5">
                {tipoData === "DATA_COM" ? "Posicionamento para garantir" : "Desembolso de provento"}
              </span>
            </div>

            {/* Total de Ativos Analisados */}
            <div className="p-3.5 rounded-xl bg-neutral-950/60 border border-neutral-800">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-neutral-400 font-mono block mb-1">
                Ativos na Matriz
              </span>
              <div className="text-sm font-bold text-[#d4af37] font-mono">
                {calendario.total_ativos} Ativos
              </div>
              <span className="text-[10px] text-neutral-500 block mt-0.5">
                Histórico B3 {calendario.periodo_anos} anos
              </span>
            </div>

            {/* Meses Mais Frequentes */}
            <div className="p-3.5 rounded-xl bg-neutral-950/60 border border-neutral-800 col-span-2 sm:col-span-2">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-neutral-400 font-mono block mb-1">
                Picos de Proventos no Ano
              </span>
              <div className="flex flex-wrap items-center gap-1.5">
                {Object.entries(calendario.meses_mais_frequentes || {})
                  .sort((a, b) => b[1] - a[1])
                  .slice(0, 4)
                  .map(([mes, qtd]) => (
                    <span
                      key={mes}
                      className="px-2 py-0.5 rounded-md bg-emerald-950/60 border border-emerald-500/30 text-[11px] font-mono font-semibold text-emerald-300"
                    >
                      {mes}: <strong className="text-white">{qtd}</strong>
                    </span>
                  ))}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 2. Barra de Filtros e Busca */}
      <div className="p-4 rounded-xl bg-[#0b0b0f] border border-neutral-800 flex flex-wrap items-center justify-between gap-4">
        {/* Lado Esquerdo: Alternador Carteira vs Mercado + Filtros de Classe */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Botão Minha Carteira */}
          <button
            onClick={() => setApenasCarteira(!apenasCarteira)}
            className={`px-3 py-1.5 rounded-lg text-xs font-bold border transition-all flex items-center gap-1.5 ${
              apenasCarteira
                ? "bg-[#d4af37] text-neutral-950 border-[#d4af37] shadow-md shadow-[#d4af37]/20"
                : "bg-neutral-900 text-neutral-300 border-neutral-800 hover:border-neutral-700"
            }`}
          >
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>Minha Carteira</span>
            {!usuario && (
              <span className="text-[9px] px-1 rounded bg-neutral-800 text-neutral-400">
                requer login
              </span>
            )}
          </button>

          {/* Filtro de Classe */}
          <div className="flex items-center p-0.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs">
            <button
              onClick={() => setClasse("")}
              className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                classe === "" ? "bg-neutral-800 text-white font-bold" : "text-neutral-400 hover:text-white"
              }`}
            >
              Todos
            </button>
            <button
              onClick={() => setClasse("ACAO")}
              className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                classe === "ACAO" ? "bg-neutral-800 text-white font-bold" : "text-neutral-400 hover:text-white"
              }`}
            >
              Ações
            </button>
            <button
              onClick={() => setClasse("FII")}
              className={`px-2.5 py-1 rounded-md font-medium transition-all ${
                classe === "FII" ? "bg-neutral-800 text-white font-bold" : "text-neutral-400 hover:text-white"
              }`}
            >
              FIIs
            </button>
          </div>

          {/* Filtro de Probabilidade Mínima */}
          <select
            value={probMin}
            onChange={(e) => setProbMin(Number(e.target.value))}
            className="px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-neutral-200 font-medium focus:outline-none focus:border-[#d4af37]"
          >
            <option value={0}>Probabilidade: Todas (≥ 0%)</option>
            <option value={40}>Probabilidade ≥ 40% (Provável)</option>
            <option value={50}>Probabilidade ≥ 50% (Consistente)</option>
            <option value={70}>Probabilidade ≥ 70% (Frequente)</option>
            <option value={80}>Probabilidade ≥ 80% (Altamente Recorrente)</option>
          </select>

          {/* Filtro por Setor */}
          {listaSetores.length > 0 && (
            <select
              value={setorSelecionado}
              onChange={(e) => setSetorSelecionado(e.target.value)}
              className="px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-neutral-200 font-medium focus:outline-none focus:border-[#d4af37] max-w-[180px]"
            >
              <option value="">Setores: Todos</option>
              {listaSetores.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          )}
        </div>

        {/* Lado Direito: Campo de Busca */}
        <div className="relative min-w-[220px]">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-neutral-500" />
          <input
            type="text"
            placeholder="Buscar ticker ou nome..."
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-[#d4af37] font-mono"
          />
        </div>
      </div>

      {/* Legenda Visual de Probabilidade */}
      <div className="flex flex-wrap items-center justify-between text-[11px] text-neutral-400 px-2 gap-2">
        <div className="flex items-center gap-3">
          <span className="font-semibold text-neutral-300">Intensidade Histórica:</span>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-500/90 border border-emerald-400"></span>
            <span>≥ 80% (Altamente Recorrente)</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-600/70 border border-emerald-500/50"></span>
            <span>60% - 79%</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-800/50 border border-emerald-600/30"></span>
            <span>40% - 59%</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-neutral-900/40 border border-neutral-800"></span>
            <span>0% (Sem histórico no mês)</span>
          </div>
        </div>

        <div className="text-neutral-500 italic">
          *Passe o mouse ou clique sobre qualquer mês para ver anos e valores médios
        </div>
      </div>

      {/* Banner de Erro */}
      {erro && (
        <div className="p-4 rounded-xl bg-amber-950/30 border border-amber-500/40 text-amber-200 text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 text-amber-400 shrink-0" />
          <span>{erro}</span>
        </div>
      )}

      {/* 3. Tabela Matriz de 12 Meses */}
      <div className="rounded-2xl bg-[#09090d] border border-neutral-800/90 shadow-2xl overflow-hidden relative">
        {loading && (
          <div className="absolute inset-0 bg-[#09090d]/70 backdrop-blur-xs flex items-center justify-center z-30">
            <div className="flex items-center gap-3 px-4 py-2 rounded-xl bg-neutral-900 border border-[#d4af37]/30 text-white text-xs font-semibold">
              <span className="w-2.5 h-2.5 rounded-full bg-[#d4af37] animate-ping" />
              Calculando matriz preditiva B3...
            </div>
          </div>
        )}

        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left text-xs">
            <thead>
              <tr className="bg-neutral-950/90 border-b border-neutral-800 text-[11px] font-semibold text-neutral-400 uppercase tracking-wider font-mono">
                <th className="py-3.5 px-4 sticky left-0 bg-neutral-950 z-20 min-w-[200px] border-r border-neutral-800/80">
                  Ativo / Empresa
                </th>
                <th className="py-3.5 px-3 text-center min-w-[70px]">Classe</th>
                <th className="py-3.5 px-3 text-right min-w-[80px]">DY 12M</th>
                <th className="py-3.5 px-3 text-right min-w-[80px]">Score</th>
                {/* 12 Meses */}
                {NOMES_MESES.map((nome, idx) => (
                  <th
                    key={nome}
                    className="py-3.5 px-2 text-center min-w-[65px] font-mono text-neutral-300 border-l border-neutral-800/40"
                  >
                    {nome}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-800/60 font-mono">
              {calendario?.ativos && calendario.ativos.length > 0 ? (
                calendario.ativos.map((ativo) => (
                  <tr
                    key={ativo.ticker}
                    className="hover:bg-neutral-900/50 transition-colors group"
                  >
                    {/* Coluna Fixa: Ticker, Nome e Tag Carteira */}
                    <td className="py-3 px-4 sticky left-0 bg-[#09090d] group-hover:bg-neutral-900/90 transition-colors z-10 border-r border-neutral-800/80">
                      <div className="flex items-center justify-between gap-2">
                        <div
                          onClick={() => handleRowClick(ativo.ticker)}
                          className="cursor-pointer group-hover:text-[#d4af37] transition-colors"
                        >
                          <div className="flex items-center gap-1.5">
                            <span className="font-black text-white text-sm tracking-wide">
                              {ativo.ticker}
                            </span>
                            {ativo.possui_na_carteira && (
                              <span className="text-[9px] px-1.5 py-0.2 rounded bg-[#d4af37]/20 text-[#d4af37] border border-[#d4af37]/40 font-bold uppercase">
                                Carteira
                              </span>
                            )}
                          </div>
                          <span
                            className="text-[11px] text-neutral-400 font-sans truncate max-w-[170px] block"
                            title={ativo.nome}
                          >
                            {ativo.nome}
                          </span>
                        </div>
                      </div>
                    </td>

                    {/* Classe */}
                    <td className="py-3 px-3 text-center">
                      <span
                        className={`text-[10px] px-2 py-0.5 rounded font-bold uppercase ${
                          ativo.classe === "FII"
                            ? "bg-purple-950/60 text-purple-300 border border-purple-800/40"
                            : "bg-blue-950/60 text-blue-300 border border-blue-800/40"
                        }`}
                      >
                        {ativo.classe}
                      </span>
                    </td>

                    {/* DY 12M */}
                    <td className="py-3 px-3 text-right font-bold text-neutral-200">
                      {ativo.dy > 0 ? `${ativo.dy.toFixed(1)}%` : "-"}
                    </td>

                    {/* Score */}
                    <td className="py-3 px-3 text-right">
                      <span
                        className={`text-xs font-black ${
                          ativo.score >= 80
                            ? "text-emerald-400"
                            : ativo.score >= 60
                            ? "text-[#d4af37]"
                            : "text-amber-400"
                        }`}
                      >
                        {ativo.score.toFixed(0)}
                      </span>
                    </td>

                    {/* 12 Células de Meses */}
                    {ativo.meses &&
                      ativo.meses.map((m) => {
                        const estilo = obterEstiloProbabilidade(m.probabilidade);
                        const isAtiva =
                          celulaAtiva?.ativo.ticker === ativo.ticker &&
                          celulaAtiva?.mes.mes === m.mes;

                        return (
                          <td
                            key={m.mes}
                            className="py-2.5 px-1.5 text-center border-l border-neutral-800/40 relative"
                            onMouseEnter={() => setCelulaAtiva({ ativo, mes: m })}
                            onMouseLeave={() => setCelulaAtiva(null)}
                            onClick={() => setCelulaAtiva({ ativo, mes: m })}
                          >
                            <div
                              className={`w-full py-1.5 rounded-md text-[11px] transition-all cursor-pointer select-none ${estilo} ${
                                isAtiva ? "ring-2 ring-white scale-105 z-20" : ""
                              }`}
                            >
                              {m.probabilidade > 0 ? `${Math.round(m.probabilidade)}%` : "-"}
                            </div>

                            {/* Tooltip Interativo */}
                            {isAtiva && (
                              <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-64 p-3 rounded-xl bg-[#0e0f14] border border-[#d4af37]/40 shadow-2xl text-left z-50 pointer-events-none backdrop-blur-md">
                                <div className="flex items-center justify-between pb-1.5 mb-1.5 border-b border-neutral-800">
                                  <span className="font-black text-white text-xs font-mono">
                                    {ativo.ticker} • {NOMES_MESES_COMPLETOS[m.mes - 1]}
                                  </span>
                                  <span
                                    className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${
                                      m.probabilidade >= 70
                                        ? "bg-emerald-500/20 text-emerald-400"
                                        : "bg-neutral-800 text-neutral-300"
                                    }`}
                                  >
                                    {m.probabilidade.toFixed(0)}% Prob.
                                  </span>
                                </div>

                                <div className="space-y-1 text-[11px] font-sans">
                                  <div className="flex justify-between text-neutral-400">
                                    <span>Frequência Histórica:</span>
                                    <strong className="text-white font-mono">
                                      {m.ocorrencias} de {m.total_anos} anos
                                    </strong>
                                  </div>

                                  {m.anos_com_provento && m.anos_com_provento.length > 0 && (
                                    <div className="flex justify-between text-neutral-400">
                                      <span>Anos Pagos:</span>
                                      <span className="text-neutral-200 font-mono text-[10px]">
                                        {m.anos_com_provento.join(", ")}
                                      </span>
                                    </div>
                                  )}

                                  {m.valor_medio > 0 && (
                                    <div className="flex justify-between text-neutral-400">
                                      <span>Valor Médio / Cota:</span>
                                      <strong className="text-[#d4af37] font-mono">
                                        R$ {m.valor_medio.toFixed(2)}
                                      </strong>
                                    </div>
                                  )}

                                  {m.tipos && m.tipos.length > 0 && (
                                    <div className="pt-1 border-t border-neutral-800/80 flex items-center gap-1 text-[10px] text-neutral-400">
                                      <span>Tipo:</span>
                                      <span className="text-emerald-400 font-mono">
                                        {m.tipos.join(" • ")}
                                      </span>
                                    </div>
                                  )}

                                  <div className="pt-1 text-[9px] text-neutral-500 italic">
                                    Fonte: B3 Oficial ({calendario.periodo_anos}A)
                                  </div>
                                </div>
                              </div>
                            )}
                          </td>
                        );
                      })}
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={16} className="py-12 text-center text-neutral-500">
                    <CalendarIcon className="w-8 h-8 mx-auto mb-2 opacity-30" />
                    <p className="text-sm font-medium">Nenhum ativo encontrado para os filtros selecionados.</p>
                    <p className="text-xs text-neutral-600 mt-1">
                      {apenasCarteira && !usuario
                        ? "Faça login para visualizar apenas os ativos da sua carteira."
                        : "Tente reduzir a probabilidade mínima ou remover o termo de busca."}
                    </p>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
