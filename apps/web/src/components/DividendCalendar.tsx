"use client";

import React, { useState, useEffect, useMemo, useCallback, useRef } from "react";
import {
  Calendar as CalendarIcon,
  Search,
  CheckCircle2,
  Clock,
  Sparkles,
  AlertCircle,
  ChevronDown,
  Loader2,
  X,
  Layers,
  LayoutGrid,
  Table as TableIcon,
  TrendingUp,
  Building2,
  ExternalLink,
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

const LIMITE_POR_PAGINA = 20;

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

  // Modo de visualização: Tabela Matriz vs Visão Calendário (Cards de 12 Meses)
  const [modoVisualizacao, setModoVisualizacao] = useState<"tabela" | "calendario">("tabela");
  const [modalCalendarioAberto, setModalCalendarioAberto] = useState<boolean>(false);

  // Estado dos dados e Paginação (Top 20 com Infinite Scroll)
  const [listaAtivos, setListaAtivos] = useState<AtivoCalendario[]>([]);
  const [totalAtivosGeral, setTotalAtivosGeral] = useState<number>(0);
  const [temMais, setTemMais] = useState<boolean>(false);
  const [periodoAnos, setPeriodoAnos] = useState<number>(5);

  const [loadingInicial, setLoadingInicial] = useState<boolean>(true);
  const [loadingMais, setLoadingMais] = useState<boolean>(false);
  const [erro, setErro] = useState<string | null>(null);

  // Tooltip flutuante não-cortável (position: fixed com detecção de borda)
  const [tooltip, setTooltip] = useState<{
    ativo: AtivoCalendario;
    mes: MesProvento;
    x: number;
    y: number;
    placeBelow: boolean;
  } | null>(null);

  // Mês fixado via clique (para visualização estendida / mobile)
  const [detalheFixado, setDetalheFixado] = useState<{
    ativo: AtivoCalendario;
    mes: MesProvento;
  } | null>(null);

  // Sentinel para infinite scroll
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  // Mapeamento de Ativos por ticker para o Modal de Detalhes
  const mapaAtivos = useMemo(() => {
    const mapa = new Map<string, Ativo>();
    for (const at of todosAtivosMercado) {
      mapa.set(at.ticker.toUpperCase(), at);
    }
    return mapa;
  }, [todosAtivosMercado]);

  // Função central de busca com paginação
  const carregarPagina = useCallback(
    async (offsetAtual: number, append: boolean) => {
      if (append) {
        setLoadingMais(true);
      } else {
        setLoadingInicial(true);
      }
      setErro(null);

      try {
        const params = new URLSearchParams();
        params.set("tipo_data", tipoData);
        params.set("limite", LIMITE_POR_PAGINA.toString());
        params.set("offset", offsetAtual.toString());
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
        const novosAtivos = json.ativos || [];

        setTotalAtivosGeral(json.total_ativos || 0);
        setTemMais(Boolean(json.tem_mais));
        setPeriodoAnos(json.periodo_anos || 5);

        if (append) {
          setListaAtivos((prev) => {
            const existentes = new Set(prev.map((a) => a.ticker));
            const filtrados = novosAtivos.filter((a) => !existentes.has(a.ticker));
            return [...prev, ...filtrados];
          });
        } else {
          setListaAtivos(novosAtivos);
        }
      } catch (err: any) {
        console.warn("Erro ao buscar calendário:", err);
        setErro(err.message || "Não foi possível carregar a matriz de proventos.");
      } finally {
        setLoadingInicial(false);
        setLoadingMais(false);
      }
    },
    [tipoData, probMin, apenasCarteira, classe, setorSelecionado, busca, token]
  );

  // Recarrega do início sempre que os filtros mudarem (com debounce de 200ms)
  useEffect(() => {
    const timer = setTimeout(() => {
      carregarPagina(0, false);
    }, 200);
    return () => clearTimeout(timer);
  }, [carregarPagina]);

  // Carrega próxima página
  const carregarMais = useCallback(() => {
    if (loadingInicial || loadingMais || !temMais) return;
    carregarPagina(listaAtivos.length, true);
  }, [loadingInicial, loadingMais, temMais, listaAtivos.length, carregarPagina]);

  // Infinite Scroll via IntersectionObserver
  useEffect(() => {
    if (!sentinelRef.current || !temMais || loadingMais || loadingInicial) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          carregarMais();
        }
      },
      { rootMargin: "300px" }
    );

    observer.observe(sentinelRef.current);
    return () => observer.disconnect();
  }, [temMais, loadingMais, loadingInicial, carregarMais]);

  // Setores únicos a partir de todos os ativos do mercado
  const listaSetores = useMemo(() => {
    const set = new Set<string>();
    for (const a of todosAtivosMercado) {
      if (a.setor && a.setor.trim() !== "") {
        set.add(a.setor.trim());
      }
    }
    return Array.from(set).sort();
  }, [todosAtivosMercado]);

  // Concentração dinâmica calculada a partir dos ativos exibidos e do corte ativo
  const concentracaoPorMes = useMemo(() => {
    return NOMES_MESES_COMPLETOS.map((nomeMes, idx) => {
      const mesNum = idx + 1;
      const sigla = NOMES_MESES[idx];
      const pagadores: {
        ticker: string;
        nome: string;
        probabilidade: number;
        classe: string;
        valorMedio: number;
        tipos: string[];
      }[] = [];

      for (const a of listaAtivos) {
        const m = a.meses?.find((item) => item.mes === mesNum);
        if (m) {
          const atende = probMin === 0 || m.probabilidade >= probMin;
          if (atende && m.probabilidade > 0) {
            pagadores.push({
              ticker: a.ticker,
              nome: a.nome,
              probabilidade: m.probabilidade,
              classe: a.classe,
              valorMedio: m.valor_medio,
              tipos: m.tipos,
            });
          }
        }
      }

      pagadores.sort((a, b) => b.probabilidade - a.probabilidade);

      return {
        mes: mesNum,
        nomeMes,
        sigla,
        totalAtivos: pagadores.length,
        pagadores,
      };
    });
  }, [listaAtivos, probMin]);

  // Meses ordenados por volume de proventos e depois cronologicamente
  const mesesComPagadores = useMemo(() => {
    return concentracaoPorMes
      .filter((m) => m.totalAtivos > 0)
      .sort((a, b) => {
        if (b.totalAtivos !== a.totalAtivos) return b.totalAtivos - a.totalAtivos;
        return a.mes - b.mes;
      });
  }, [concentracaoPorMes]);

  const maxPagadoresNoMes = useMemo(() => {
    return Math.max(...concentracaoPorMes.map((m) => m.totalAtivos), 1);
  }, [concentracaoPorMes]);

  // Estilo visual da probabilidade
  const obterEstiloProbabilidade = (prob: number, atendeFiltro: boolean) => {
    if (prob <= 0 || !atendeFiltro) {
      return "bg-neutral-900/30 text-neutral-600 border border-transparent";
    }
    if (prob >= 100) {
      return "bg-emerald-400 text-neutral-950 font-black border border-emerald-300 shadow-md shadow-emerald-400/30";
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

  // Posicionamento inteligente do Tooltip (nunca corta por overflow)
  const handleCellMouseEnter = (
    e: React.MouseEvent<HTMLTableCellElement>,
    ativo: AtivoCalendario,
    mes: MesProvento
  ) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const placeBelow = rect.top < 240;
    setTooltip({
      ativo,
      mes,
      x: rect.left + rect.width / 2,
      y: placeBelow ? rect.bottom + 8 : rect.top - 8,
      placeBelow,
    });
  };

  const handleCellMouseLeave = () => {
    setTooltip(null);
  };

  const handleCellClick = (ativo: AtivoCalendario, mes: MesProvento) => {
    setDetalheFixado({ ativo, mes });
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
      {/* 1. Header do MDI com Metodologia */}
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
              <strong className="text-neutral-200">B3 Oficial dos últimos 5 anos (2021 a 2026)</strong>.
              Identifique com exatidão os meses de maior concentração para antecipar compras antes do fechamento do pregão.
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
                title="Data com direito: último dia em que você precisa ter a ação para garantir os proventos"
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
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-6 pt-5 border-t border-neutral-800/80">
          {/* Modo de Análise */}
          <div className="p-3.5 rounded-xl bg-neutral-950/60 border border-neutral-800">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-neutral-400 font-mono block mb-1">
              Critério Ativo
            </span>
            <div className="text-sm font-bold text-white flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
              {tipoData === "DATA_COM" ? "Data COM (Direito)" : "Data de Pagamento"}
            </div>
            <span className="text-[10px] text-neutral-500 block mt-0.5">
              {tipoData === "DATA_COM" ? "Posicionamento para garantir" : "Data de desembolso"}
            </span>
          </div>

          {/* Ativos Carregados vs Total */}
          <div className="p-3.5 rounded-xl bg-neutral-950/60 border border-neutral-800">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-neutral-400 font-mono block mb-1">
              Ativos Exibidos
            </span>
            <div className="text-sm font-bold text-[#d4af37] font-mono">
              {listaAtivos.length} <span className="text-xs text-neutral-500 font-normal">de</span> {totalAtivosGeral}
            </div>
            <span className="text-[10px] text-neutral-500 block mt-0.5">
              Top volume B3 ({periodoAnos}A)
            </span>
          </div>

          {/* Meses com Maior Concentração de Proventos (Sem cortes indevidos + Botão para Calendário) */}
          <div className="p-3.5 rounded-xl bg-neutral-950/60 border border-neutral-800 col-span-2 sm:col-span-2">
            <div className="flex items-center justify-between mb-1.5">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-neutral-400 font-mono">
                Meses com Proventos ({mesesComPagadores.length} meses ativos)
              </span>
              <button
                onClick={() => setModalCalendarioAberto(true)}
                className="text-[10px] text-[#d4af37] hover:text-[#f5df88] font-bold flex items-center gap-1 hover:underline transition-all"
              >
                <CalendarIcon className="w-3 h-3" />
                <span>Ver Calendário Geral (12M)</span>
              </button>
            </div>
            <div className="flex flex-wrap items-center gap-1.5">
              {mesesComPagadores.length > 0 ? (
                mesesComPagadores.map((m) => (
                  <button
                    key={m.mes}
                    onClick={() => setModalCalendarioAberto(true)}
                    className="px-2 py-0.5 rounded-md bg-emerald-950/60 hover:bg-emerald-900/60 border border-emerald-500/30 text-[11px] font-mono font-semibold text-emerald-300 transition-colors cursor-pointer"
                    title={`Clique para ver as ${m.totalAtivos} ações de ${m.nomeMes}`}
                  >
                    {m.sigla}: <strong className="text-white">{m.totalAtivos}</strong>
                  </button>
                ))
              ) : (
                <span className="text-xs text-neutral-500">
                  Nenhum mês atende ao filtro de probabilidade (≥ {probMin}%).
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 2. Barra de Filtros, Busca e Alternador de Visão */}
      <div className="p-4 rounded-xl bg-[#0b0b0f] border border-neutral-800 flex flex-wrap items-center justify-between gap-4">
        {/* Lado Esquerdo: Filtros */}
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
                login
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

          {/* Filtro de Probabilidade Mínima (Múltiplos de 20% baseados na janela de 5 anos) */}
          <select
            value={probMin}
            onChange={(e) => setProbMin(Number(e.target.value))}
            className="px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-neutral-200 font-medium focus:outline-none focus:border-[#d4af37]"
          >
            <option value={0}>Probabilidade: Todas (≥ 0%)</option>
            <option value={20}>Probabilidade ≥ 20% (1 de 5 anos)</option>
            <option value={40}>Probabilidade ≥ 40% (2 de 5 anos)</option>
            <option value={60}>Probabilidade ≥ 60% (3 de 5 anos)</option>
            <option value={80}>Probabilidade ≥ 80% (4 de 5 anos)</option>
            <option value={100}>Probabilidade = 100% (5 de 5 anos)</option>
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

        {/* Lado Direito: Alternador de Visualização & Campo de Busca */}
        <div className="flex items-center gap-2 w-full sm:w-auto justify-between sm:justify-end">
          {/* Alternador Matriz vs Calendário */}
          <div className="flex items-center p-0.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs">
            <button
              onClick={() => setModoVisualizacao("tabela")}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md font-bold transition-all ${
                modoVisualizacao === "tabela"
                  ? "bg-[#d4af37] text-neutral-950 shadow-sm"
                  : "text-neutral-400 hover:text-white"
              }`}
              title="Exibir como Matriz Tabela (12 Colunas)"
            >
              <TableIcon className="w-3.5 h-3.5" />
              <span>Matriz</span>
            </button>
            <button
              onClick={() => setModoVisualizacao("calendario")}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md font-bold transition-all ${
                modoVisualizacao === "calendario"
                  ? "bg-[#d4af37] text-neutral-950 shadow-sm"
                  : "text-neutral-400 hover:text-white"
              }`}
              title="Exibir como Cards de Calendário Anual (12 Meses)"
            >
              <LayoutGrid className="w-3.5 h-3.5" />
              <span>Calendário Anual</span>
            </button>
          </div>

          {/* Campo de Busca */}
          <div className="relative min-w-[200px]">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-neutral-500" />
            <input
              type="text"
              placeholder="Buscar ticker ou empresa..."
              value={busca}
              onChange={(e) => setBusca(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-white placeholder-neutral-500 focus:outline-none focus:border-[#d4af37] font-mono"
            />
          </div>
        </div>
      </div>

      {/* Painel de Detalhe Fixado por Clique */}
      {detalheFixado && (
        <div className="p-4 rounded-xl bg-[#0f1118] border border-[#d4af37]/50 shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 animate-in fade-in slide-in-from-top-2 duration-150">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-[#d4af37]/15 border border-[#d4af37]/30 flex items-center justify-center font-mono font-black text-[#d4af37] text-sm">
              {detalheFixado.ativo.ticker.slice(0, 4)}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono font-black text-white text-base">
                  {detalheFixado.ativo.ticker}
                </span>
                <span className="text-xs text-neutral-400 font-medium">
                  {detalheFixado.ativo.nome}
                </span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-mono font-bold border border-emerald-500/40">
                  {NOMES_MESES_COMPLETOS[detalheFixado.mes.mes - 1]}: {detalheFixado.mes.probabilidade.toFixed(0)}%
                </span>
              </div>
              <div className="flex flex-wrap items-center gap-3 text-xs text-neutral-400 mt-1">
                <span>
                  Frequência: <strong className="text-white font-mono">{detalheFixado.mes.ocorrencias} de {detalheFixado.mes.total_anos} anos</strong>
                </span>
                {detalheFixado.mes.anos_com_provento && (
                  <span>
                    Anos: <strong className="text-neutral-200 font-mono">[{detalheFixado.mes.anos_com_provento.join(", ")}]</strong>
                  </span>
                )}
                {detalheFixado.mes.valor_medio > 0 && (
                  <span>
                    Valor Médio: <strong className="text-[#d4af37] font-mono">R$ {detalheFixado.mes.valor_medio.toFixed(2)}</strong>
                  </span>
                )}
                {detalheFixado.mes.tipos && (
                  <span>
                    Tipo: <strong className="text-emerald-400 font-mono">{detalheFixado.mes.tipos.join(" • ")}</strong>
                  </span>
                )}
              </div>
            </div>
          </div>
          <button
            onClick={() => setDetalheFixado(null)}
            className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors self-end sm:self-center"
            title="Fechar detalhes"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Legenda Visual */}
      <div className="flex flex-wrap items-center justify-between text-[11px] text-neutral-400 px-2 gap-2">
        <div className="flex items-center gap-3">
          <span className="font-semibold text-neutral-300">Intensidade Histórica (5 Anos):</span>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-400 border border-emerald-300"></span>
            <span>100% (5 de 5)</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-500/90 border border-emerald-400"></span>
            <span>80% (4 de 5)</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-600/70 border border-emerald-500/50"></span>
            <span>60% (3 de 5)</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-800/50 border border-emerald-600/30"></span>
            <span>40% (2 de 5)</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-emerald-950/40 border border-emerald-800/20"></span>
            <span>20% (1 de 5)</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3.5 h-3.5 rounded bg-neutral-900/40 border border-neutral-800"></span>
            <span>0% / Abaixo</span>
          </div>
        </div>

        <div className="text-neutral-500 italic">
          *Passe o mouse ou clique no mês para fixar detalhes
        </div>
      </div>

      {/* Banner de Erro */}
      {erro && (
        <div className="p-4 rounded-xl bg-amber-950/30 border border-amber-500/40 text-amber-200 text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 text-amber-400 shrink-0" />
          <span>{erro}</span>
        </div>
      )}

      {/* ===================================================================================== */}
      {/* MODO 1: VISÃO CALENDÁRIO ANUAL (12 CARDS DE MESES)                                    */}
      {/* ===================================================================================== */}
      {modoVisualizacao === "calendario" && (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 animate-in fade-in duration-200">
          {concentracaoPorMes.map((mesInfo) => {
            const hasProventos = mesInfo.totalAtivos > 0;
            const barWidth = `${Math.min(100, Math.round((mesInfo.totalAtivos / maxPagadoresNoMes) * 100))}%`;

            return (
              <div
                key={mesInfo.mes}
                className={`p-4 rounded-2xl border transition-all flex flex-col justify-between ${
                  hasProventos
                    ? "bg-[#0b0c12] border-neutral-800 hover:border-[#d4af37]/50 shadow-lg"
                    : "bg-[#08080a] border-neutral-900/60 opacity-60"
                }`}
              >
                <div>
                  {/* Cabeçalho do Mês */}
                  <div className="flex items-center justify-between pb-2 mb-2 border-b border-neutral-800/80">
                    <span className="text-sm font-black text-white font-mono uppercase tracking-wider">
                      {mesInfo.nomeMes}
                    </span>
                    <span
                      className={`text-xs font-mono font-bold px-2 py-0.5 rounded-lg ${
                        hasProventos
                          ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                          : "bg-neutral-900 text-neutral-500"
                      }`}
                    >
                      {mesInfo.totalAtivos} {mesInfo.totalAtivos === 1 ? "ativo" : "ativos"}
                    </span>
                  </div>

                  {/* Barra de Concentração */}
                  {hasProventos && (
                    <div className="w-full h-1.5 bg-neutral-900 rounded-full mb-3 overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-emerald-500 to-[#d4af37] rounded-full transition-all duration-300"
                        style={{ width: barWidth }}
                      />
                    </div>
                  )}

                  {/* Lista de Ativos Pagadores */}
                  <div className="space-y-1.5 max-h-48 overflow-y-auto pr-1">
                    {hasProventos ? (
                      mesInfo.pagadores.map((p) => (
                        <div
                          key={p.ticker}
                          onClick={() => handleRowClick(p.ticker)}
                          className="p-2 rounded-xl bg-neutral-950/80 hover:bg-neutral-900 border border-neutral-800/80 hover:border-[#d4af37]/40 cursor-pointer transition-all flex items-center justify-between gap-2 group"
                        >
                          <div className="flex items-center gap-1.5">
                            <span className="font-mono font-black text-white text-xs group-hover:text-[#d4af37] transition-colors">
                              {p.ticker}
                            </span>
                            <span className="text-[9px] px-1 rounded bg-neutral-800 text-neutral-400 uppercase font-bold">
                              {p.classe}
                            </span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <span
                              className={`text-[10px] font-mono font-black px-1.5 py-0.2 rounded ${
                                p.probabilidade >= 100
                                  ? "bg-emerald-400 text-black"
                                  : p.probabilidade >= 80
                                  ? "bg-emerald-500/80 text-black font-bold"
                                  : "bg-emerald-950 text-emerald-300 border border-emerald-800/40"
                              }`}
                            >
                              {p.probabilidade.toFixed(0)}%
                            </span>
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className="py-6 text-center text-xs text-neutral-600 font-sans italic">
                        Sem proventos neste mês
                      </div>
                    )}
                  </div>
                </div>

                {hasProventos && (
                  <div className="pt-2 mt-2 border-t border-neutral-900 text-[10px] text-neutral-500 text-right font-mono">
                    Critério: {tipoData === "DATA_COM" ? "Data COM" : "Pagamento"}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* ===================================================================================== */}
      {/* MODO 2: VISÃO TABELA MATRIZ (12 MESES)                                                */}
      {/* ===================================================================================== */}
      {modoVisualizacao === "tabela" && (
        <div className="rounded-2xl bg-[#09090d] border border-neutral-800/90 shadow-2xl overflow-hidden relative">
          {loadingInicial && (
            <div className="absolute inset-0 bg-[#09090d]/80 backdrop-blur-xs flex items-center justify-center z-30">
              <div className="flex items-center gap-3 px-4 py-2.5 rounded-xl bg-neutral-900 border border-[#d4af37]/40 text-white text-xs font-semibold shadow-xl">
                <Loader2 className="w-4 h-4 text-[#d4af37] animate-spin" />
                <span>Carregando Top 20 ativos da B3...</span>
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
                  {NOMES_MESES.map((nome) => (
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
                {listaAtivos.length > 0 ? (
                  listaAtivos.map((ativo) => (
                    <tr
                      key={ativo.ticker}
                      className="hover:bg-neutral-900/50 transition-colors group"
                    >
                      {/* Coluna Fixa: Ticker e Nome */}
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
                          const atendeFiltro = probMin === 0 || m.probabilidade >= probMin;
                          const estilo = obterEstiloProbabilidade(m.probabilidade, atendeFiltro);
                          const isAtiva =
                            detalheFixado?.ativo.ticker === ativo.ticker &&
                            detalheFixado?.mes.mes === m.mes;

                          return (
                            <td
                              key={m.mes}
                              className="py-2.5 px-1.5 text-center border-l border-neutral-800/40"
                              onMouseEnter={(e) => handleCellMouseEnter(e, ativo, m)}
                              onMouseLeave={handleCellMouseLeave}
                              onClick={() => handleCellClick(ativo, m)}
                            >
                              <div
                                className={`w-full py-1.5 rounded-md text-[11px] transition-all cursor-pointer select-none ${estilo} ${
                                  isAtiva ? "ring-2 ring-white scale-105 z-20" : ""
                                }`}
                              >
                                {atendeFiltro && m.probabilidade > 0 ? `${Math.round(m.probabilidade)}%` : "-"}
                              </div>
                            </td>
                          );
                        })}
                    </tr>
                  ))
                ) : (
                  !loadingInicial && (
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
                  )
                )}
              </tbody>
            </table>
          </div>

          {/* Barra de Rodapé / Infinite Scroll */}
          <div className="p-4 bg-neutral-950/90 border-t border-neutral-800 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-neutral-400">
            <div className="flex items-center gap-2">
              <span>
                Exibindo <strong className="text-white font-mono">{listaAtivos.length}</strong> de{" "}
                <strong className="text-white font-mono">{totalAtivosGeral}</strong> ativos
              </span>
              <span className="text-neutral-600">•</span>
              <span className="text-neutral-500">Ordenados por Maior Volume & Relevância</span>
            </div>

            {/* Botão Carregar Mais (+20) */}
            {temMais ? (
              <button
                onClick={carregarMais}
                disabled={loadingMais}
                className="flex items-center gap-2 px-4 py-2 rounded-xl bg-[#d4af37]/15 hover:bg-[#d4af37]/25 text-[#d4af37] border border-[#d4af37]/40 font-bold transition-all disabled:opacity-50"
              >
                {loadingMais ? (
                  <>
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                    <span>Carregando +20 ativos...</span>
                  </>
                ) : (
                  <>
                    <span>Carregar mais 20 ativos</span>
                    <ChevronDown className="w-3.5 h-3.5" />
                  </>
                )}
              </button>
            ) : (
              listaAtivos.length > 0 && (
                <span className="text-emerald-400/80 font-mono text-[11px] flex items-center gap-1">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  Todos os ativos carregados
                </span>
              )
            )}
          </div>

          {/* Sentinel invisível para acionar carregamento automático ao rolar a página */}
          <div ref={sentinelRef} className="h-4 w-full pointer-events-none" />
        </div>
      )}

      {/* ===================================================================================== */}
      {/* MODAL DE CALENDÁRIO GERAL (12 MESES CONSOLIDADOS COM LISTA DE AÇÕES)                  */}
      {/* ===================================================================================== */}
      {modalCalendarioAberto && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in duration-200">
          <div
            className="w-full max-w-5xl rounded-2xl bg-[#0e0f14] border border-neutral-800 shadow-2xl overflow-hidden flex flex-col max-h-[92vh]"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header do Modal */}
            <div className="p-5 border-b border-neutral-800 flex items-center justify-between bg-neutral-950/60">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-xl bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/30">
                  <CalendarIcon className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-white flex items-center gap-2">
                    Calendário Geral de Proventos (12 Meses)
                    <span className="text-xs font-mono px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 border border-emerald-500/40">
                      {apenasCarteira ? "Minha Carteira" : "Mercado B3"}
                    </span>
                  </h3>
                  <p className="text-xs text-neutral-400">
                    Concentração de ativos pagadores em {tipoData === "DATA_COM" ? "Data COM (Direito)" : "Data de Pagamento"} com probabilidade ≥ {probMin}%
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                {/* Seletor rápido de probabilidade no modal */}
                <select
                  value={probMin}
                  onChange={(e) => setProbMin(Number(e.target.value))}
                  className="px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-neutral-200 font-medium focus:outline-none focus:border-[#d4af37]"
                >
                  <option value={0}>Todas (≥ 0%)</option>
                  <option value={20}>≥ 20% (1 de 5A)</option>
                  <option value={40}>≥ 40% (2 de 5A)</option>
                  <option value={60}>≥ 60% (3 de 5A)</option>
                  <option value={80}>≥ 80% (4 de 5A)</option>
                  <option value={100}>100% (5 de 5A)</option>
                </select>

                <button
                  onClick={() => setModalCalendarioAberto(false)}
                  className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>
            </div>

            {/* Conteúdo: 12 Cards do Ano */}
            <div className="p-6 overflow-y-auto grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 flex-1">
              {concentracaoPorMes.map((mesInfo) => {
                const hasProventos = mesInfo.totalAtivos > 0;
                const barWidth = `${Math.min(100, Math.round((mesInfo.totalAtivos / maxPagadoresNoMes) * 100))}%`;

                return (
                  <div
                    key={mesInfo.mes}
                    className={`p-4 rounded-xl border flex flex-col justify-between ${
                      hasProventos
                        ? "bg-[#12131a] border-neutral-800 hover:border-[#d4af37]/40 shadow-md"
                        : "bg-[#090a0d] border-neutral-900/60 opacity-50"
                    }`}
                  >
                    <div>
                      {/* Cabeçalho */}
                      <div className="flex items-center justify-between pb-2 mb-2 border-b border-neutral-800">
                        <span className="text-xs font-black text-white font-mono uppercase tracking-wider">
                          {mesInfo.nomeMes}
                        </span>
                        <span
                          className={`text-[11px] font-mono font-bold px-2 py-0.5 rounded ${
                            hasProventos
                              ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                              : "bg-neutral-900 text-neutral-500"
                          }`}
                        >
                          {mesInfo.totalAtivos} {mesInfo.totalAtivos === 1 ? "ativo" : "ativos"}
                        </span>
                      </div>

                      {/* Barra de Progresso */}
                      {hasProventos && (
                        <div className="w-full h-1.5 bg-neutral-900 rounded-full mb-2.5 overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-emerald-500 to-[#d4af37] rounded-full"
                            style={{ width: barWidth }}
                          />
                        </div>
                      )}

                      {/* Lista de Ações */}
                      <div className="space-y-1 max-h-44 overflow-y-auto pr-1">
                        {hasProventos ? (
                          mesInfo.pagadores.map((p) => (
                            <div
                              key={p.ticker}
                              onClick={() => {
                                setModalCalendarioAberto(false);
                                handleRowClick(p.ticker);
                              }}
                              className="p-1.5 rounded-lg bg-neutral-950 hover:bg-neutral-900 border border-neutral-800/80 hover:border-[#d4af37]/40 cursor-pointer transition-all flex items-center justify-between gap-1 group"
                              title={`${p.nome} • Clique para ver Valuation`}
                            >
                              <div className="flex items-center gap-1">
                                <span className="font-mono font-bold text-white text-xs group-hover:text-[#d4af37]">
                                  {p.ticker}
                                </span>
                                <span className="text-[9px] px-1 rounded bg-neutral-900 text-neutral-400 font-bold uppercase">
                                  {p.classe}
                                </span>
                              </div>
                              <span
                                className={`text-[10px] font-mono font-black px-1.5 py-0.2 rounded ${
                                  p.probabilidade >= 100
                                    ? "bg-emerald-400 text-black"
                                    : p.probabilidade >= 80
                                    ? "bg-emerald-500/80 text-black"
                                    : "bg-emerald-950 text-emerald-300 border border-emerald-800/40"
                                }`}
                              >
                                {p.probabilidade.toFixed(0)}%
                              </span>
                            </div>
                          ))
                        ) : (
                          <div className="py-4 text-center text-[11px] text-neutral-600 italic">
                            Sem proventos
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Rodapé do Modal */}
            <div className="p-4 border-t border-neutral-800 flex items-center justify-between bg-neutral-950/80 text-xs text-neutral-400">
              <span>*Clique em qualquer ativo para abrir a análise fundamentalista completa.</span>
              <button
                onClick={() => setModalCalendarioAberto(false)}
                className="px-4 py-2 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-white font-bold transition-colors"
              >
                Fechar
              </button>
            </div>
          </div>
        </div>
      )}

      {/* TOOLTIP FLUTUANTE GLOBAL (RENDERIZADO VIA FIXED FORA DO OVERFLOW SCROLL) */}
      {tooltip && (
        <div
          style={{
            position: "fixed",
            left: `${Math.max(150, Math.min(typeof window !== "undefined" ? window.innerWidth - 150 : 500, tooltip.x))}px`,
            top: tooltip.placeBelow ? `${tooltip.y}px` : "auto",
            bottom: !tooltip.placeBelow ? `${(typeof window !== "undefined" ? window.innerHeight : 800) - tooltip.y}px` : "auto",
            transform: "translateX(-50%)",
            zIndex: 99999,
          }}
          className="w-72 p-3.5 rounded-xl bg-[#0e0f14]/98 border border-[#d4af37]/60 shadow-[0_12px_40px_rgba(0,0,0,0.9)] text-left pointer-events-none backdrop-blur-md animate-in fade-in zoom-in-95 duration-100"
        >
          <div className="flex items-center justify-between pb-1.5 mb-1.5 border-b border-neutral-800">
            <span className="font-black text-white text-xs font-mono">
              {tooltip.ativo.ticker} • {NOMES_MESES_COMPLETOS[tooltip.mes.mes - 1]}
            </span>
            <span
              className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${
                tooltip.mes.probabilidade >= 70
                  ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                  : "bg-neutral-800 text-neutral-300 border border-neutral-700"
              }`}
            >
              {tooltip.mes.probabilidade.toFixed(0)}% Prob.
            </span>
          </div>

          {probMin > 0 && tooltip.mes.probabilidade < probMin && (
            <div className="mb-2 px-2 py-1 rounded bg-amber-950/40 border border-amber-500/30 text-[10px] text-amber-300 font-medium">
              Abaixo do filtro ativo (≥ {probMin}%)
            </div>
          )}

          <div className="space-y-1.5 text-[11px] font-sans">
            <div className="flex justify-between text-neutral-400">
              <span>Frequência Histórica:</span>
              <strong className="text-white font-mono">
                {tooltip.mes.ocorrencias} de {tooltip.mes.total_anos} anos
              </strong>
            </div>

            {tooltip.mes.anos_com_provento && tooltip.mes.anos_com_provento.length > 0 && (
              <div className="flex justify-between text-neutral-400">
                <span>Anos com Provento:</span>
                <span className="text-neutral-200 font-mono text-[10px]">
                  {tooltip.mes.anos_com_provento.join(", ")}
                </span>
              </div>
            )}

            {tooltip.mes.valor_medio > 0 && (
              <div className="flex justify-between text-neutral-400">
                <span>Valor Médio / Cota:</span>
                <strong className="text-[#d4af37] font-mono">
                  R$ {tooltip.mes.valor_medio.toFixed(2)}
                </strong>
              </div>
            )}

            {tooltip.mes.tipos && tooltip.mes.tipos.length > 0 && (
              <div className="pt-1 border-t border-neutral-800/80 flex items-center gap-1 text-[10px] text-neutral-400">
                <span>Tipo:</span>
                <span className="text-emerald-400 font-mono">
                  {tooltip.mes.tipos.join(" • ")}
                </span>
              </div>
            )}

            <div className="pt-1 text-[9px] text-neutral-500 italic">
              Fonte: B3 Oficial ({periodoAnos} Anos)
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
