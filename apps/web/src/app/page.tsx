"use client";

import React, { useEffect, useState, useCallback } from "react";
import { Ativo, RespostaRankings, CenarioMacro } from "@/types/market";
import { ResumoCarteira, CriarPosicaoRequest, ResultadoRebalanceamento, PosicaoConsolidada, ClasseAtivo } from "@/types/portfolio";
import { API_BASE_URL, API_V1_URL } from "@/config/api";
import { Header } from "@/components/Header";
import { SummaryCards } from "@/components/SummaryCards";
import { MarketTable } from "@/components/MarketTable";
import { AssetDetailModal } from "@/components/AssetDetailModal";
import { PortfolioSummaryCards } from "@/components/PortfolioSummaryCards";
import { PortfolioDonutChart } from "@/components/PortfolioDonutChart";
import { MacroAllocationView } from "@/components/MacroAllocationView";
import { PortfolioTable } from "@/components/PortfolioTable";
import { AddPositionModal } from "@/components/AddPositionModal";
import { RebalanceSimulatorModal } from "@/components/RebalanceSimulatorModal";
import { TransactionHistoryModal } from "@/components/TransactionHistoryModal";
import { AssetChartModal } from "@/components/AssetChartModal";
import { AdvisorModal } from "@/components/AdvisorModal";
import { B3ImportModal } from "@/components/B3ImportModal";
import { ProventosView } from "@/components/ProventosView";
import { AuthModal } from "@/components/AuthModal";
import { DividendCalendar } from "@/components/DividendCalendar";
import { useAuth } from "@/context/AuthContext";
import {
  TrendingUp,
  Building2,
  Globe,
  AlertCircle,
  RefreshCw,
  Landmark,
  Plus,
  Scale,
  Briefcase,
  Lock,
  ArrowRight,
  ShieldCheck,
  History,
  Sparkles,
  Upload,
  Eye,
  EyeOff,
} from "lucide-react";

export default function Dashboard() {
  const { usuario, token, loading: authLoading } = useAuth();

  // Estado de navegação
  const [abaAtiva, setAbaAtiva] = useState<"mercado" | "carteira" | "proventos" | "calendario">("mercado");

  // Dados do Mercado
  const [data, setData] = useState<RespostaRankings>({
    total: 0,
    acoes: [],
    fiis: [],
    etfs: [],
  });
  const [macro, setMacro] = useState<CenarioMacro | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [ativoSelecionado, setAtivoSelecionado] = useState<Ativo | null>(null);
  const [ultimaAtualizacao, setUltimaAtualizacao] = useState<string>("");

  // Dados da Carteira
  const [resumoCarteira, setResumoCarteira] = useState<ResumoCarteira | null>(null);
  const [loadingCarteira, setLoadingCarteira] = useState(false);
  const [erroCarteira, setErroCarteira] = useState<string | null>(null);
  const [classeFiltroDonut, setClasseFiltroDonut] = useState<ClasseAtivo | null>(null);

  // Privacidade: Ocultar valores em reais com efeito de Blur
  const [ocultarValores, setOcultarValores] = useState(false);

  useEffect(() => {
    try {
      const salvo = localStorage.getItem("carteira_ocultar_valores");
      if (salvo !== null) {
        setOcultarValores(salvo === "true");
      }
    } catch (e) {}
  }, []);

  const toggleOcultarValores = () => {
    setOcultarValores((prev) => {
      const novo = !prev;
      try {
        localStorage.setItem("carteira_ocultar_valores", String(novo));
      } catch (e) {}
      return novo;
    });
  };

  // Modais
  const [modalAddAberto, setModalAddAberto] = useState(false);
  const [modalImportB3Aberto, setModalImportB3Aberto] = useState(false);
  const [modalSimuladorAberto, setModalSimuladorAberto] = useState(false);
  const [modalAdvisorAberto, setModalAdvisorAberto] = useState(false);
  const [modalAuthAberto, setModalAuthAberto] = useState(false);
  const [modalHistoricoAberto, setModalHistoricoAberto] = useState(false);
  const [tickerHistoricoFiltro, setTickerHistoricoFiltro] = useState<string | null>(null);
  const [posicaoParaAporte, setPosicaoParaAporte] = useState<PosicaoConsolidada | null>(null);

  // Modal de Gráfico Técnico & Médias Móveis
  const [modalGraficoAberto, setModalGraficoAberto] = useState(false);
  const [tickerGrafico, setTickerGrafico] = useState<string | null>(null);
  const [nomeGrafico, setNomeGrafico] = useState<string>("");

  const handleAbrirGrafico = (ticker: string, nome?: string) => {
    setTickerGrafico(ticker);
    setNomeGrafico(nome || "");
    setModalGraficoAberto(true);
  };

  // 1. Carrega dados do mercado (Rankings B3 & Macro)
  const carregarRankings = async () => {
    setLoading(true);
    setError(null);
    try {
      const [resRankings, resMacro] = await Promise.all([
        fetch(`${API_V1_URL}/rankings`, { headers: { Accept: "application/json" } }),
        fetch(`${API_V1_URL}/macro`, { headers: { Accept: "application/json" } }).catch(() => null),
      ]);

      if (!resRankings.ok) {
        throw new Error(`Servidor respondeu com status ${resRankings.status}`);
      }
      const json: RespostaRankings = await resRankings.json();
      setData({
        total: json.total ?? 0,
        acoes: json.acoes ?? [],
        fiis: json.fiis ?? [],
        etfs: json.etfs ?? [],
      });

      if (resMacro && resMacro.ok) {
        const jsonMacro: CenarioMacro = await resMacro.json();
        setMacro(jsonMacro);
      }

      setUltimaAtualizacao(new Date().toLocaleTimeString("pt-BR"));
    } catch (err: any) {
      console.warn("Erro ao buscar dados da API B3 Core:", err);
      setError(
        `Não foi possível conectar ao servidor em ${API_BASE_URL}. Verifique se o backend Go está rodando.`
      );
    } finally {
      setLoading(false);
    }
  };

  // 2. Carrega dados da carteira do usuário logado
  const carregarCarteira = useCallback(async () => {
    if (!token) {
      setResumoCarteira(null);
      return;
    }

    setLoadingCarteira(true);
    setErroCarteira(null);
    try {
      const res = await fetch(`${API_V1_URL}/carteira`, {
        headers: {
          Accept: "application/json",
          Authorization: `Bearer ${token}`,
        },
      });

      if (!res.ok) {
        throw new Error(`Falha ao buscar carteira (status ${res.status})`);
      }

      const json: ResumoCarteira = await res.json();
      setResumoCarteira(json);
    } catch (err: any) {
      console.warn("Erro ao buscar carteira:", err);
      setErroCarteira(err.message || "Falha ao carregar posições da carteira.");
    } finally {
      setLoadingCarteira(false);
    }
  }, [token]);

  // Sincronização inicial com o path da URL (/carteira, /proventos, /calendario)
  useEffect(() => {
    if (typeof window !== "undefined") {
      const path = window.location.pathname.replace(/^\//, "").toLowerCase();
      if (path === "carteira") setAbaAtiva("carteira");
      else if (path === "proventos") setAbaAtiva("proventos");
      else if (path === "calendario" || path === "mdi") setAbaAtiva("calendario");
      else if (path === "" || path === "mercado") setAbaAtiva("mercado");
    }
  }, []);

  const handleMudarAba = (aba: "mercado" | "carteira" | "proventos" | "calendario") => {
    setAbaAtiva(aba);
    if (typeof window !== "undefined") {
      const targetPath = aba === "mercado" ? "/" : `/${aba}`;
      window.history.pushState(null, "", targetPath);
    }
  };

  useEffect(() => {
    carregarRankings();
  }, []);

  useEffect(() => {
    if (abaAtiva === "carteira" && usuario && token) {
      carregarCarteira();
    }
  }, [abaAtiva, usuario, token, carregarCarteira]);

  // Ações da Carteira
  const handleSalvarPosicao = async (req: CriarPosicaoRequest) => {
    if (!token) throw new Error("Usuário não autenticado.");

    const res = await fetch(`${API_V1_URL}/carteira/posicoes`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(req),
    });

    if (!res.ok) {
      const errJson = await res.json().catch(() => null);
      throw new Error(errJson?.erro || `Erro ao salvar ativo (${res.status})`);
    }

    await carregarCarteira();
  };

  const handleExcluirPosicao = async (id: number, ticker: string) => {
    if (!token) return;
    if (!window.confirm(`Tem certeza que deseja remover ${ticker} da sua carteira?`)) return;

    try {
      const res = await fetch(`${API_V1_URL}/carteira/posicoes/${id}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!res.ok) {
        throw new Error("Erro ao excluir posição.");
      }

      await carregarCarteira();
    } catch (err: any) {
      alert(err.message || "Falha ao excluir posição.");
    }
  };

  const handleSimularRebalanceamento = async (valorAporte: number): Promise<ResultadoRebalanceamento> => {
    if (!token) throw new Error("Usuário não autenticado.");

    const res = await fetch(`${API_V1_URL}/carteira/rebalancear`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ valor_aporte: valorAporte }),
    });

    if (!res.ok) {
      const errJson = await res.json().catch(() => null);
      throw new Error(errJson?.erro || "Falha ao calcular simulação.");
    }

    return await res.json();
  };

  const handleSincronizar = async () => {
    setLoading(true);
    setError(null);
    try {
      await fetch(`${API_V1_URL}/sync`, { method: "POST" });
    } catch (e) {
      console.warn("Falha ao solicitar sincronização no servidor:", e);
    }
    await carregarRankings();
    if (abaAtiva === "carteira" && usuario && token) {
      await carregarCarteira();
    }
  };

  return (
    <div className="min-h-screen bg-[#070709] text-neutral-100 flex flex-col selection:bg-[#d4af37]/30 selection:text-[#f5df88]">
      {/* Header Fixo com Seletor de Abas */}
      <Header
        totalAtivos={data.total}
        loading={loading || loadingCarteira}
        onRefresh={handleSincronizar}
        ultimaAtualizacao={ultimaAtualizacao}
        abaAtiva={abaAtiva}
        onMudarAba={handleMudarAba}
      />

      {/* Conteúdo Principal */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Seletor de Abas Mobile */}
        <div className="grid grid-cols-4 md:hidden items-center gap-1.5 mb-6 p-1 rounded-xl bg-neutral-900 border border-neutral-800">
          <button
            onClick={() => handleMudarAba("mercado")}
            className={`py-2 rounded-lg text-xs font-bold transition-all text-center ${
              abaAtiva === "mercado" ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20" : "text-neutral-400"
            }`}
          >
            Mercado
          </button>
          <button
            onClick={() => handleMudarAba("carteira")}
            className={`py-2 rounded-lg text-xs font-bold transition-all text-center ${
              abaAtiva === "carteira" ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20" : "text-neutral-400"
            }`}
          >
            Carteira
          </button>
          <button
            onClick={() => handleMudarAba("proventos")}
            className={`py-2 rounded-lg text-xs font-bold transition-all text-center ${
              abaAtiva === "proventos" ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20" : "text-neutral-400"
            }`}
          >
            Proventos
          </button>
          <button
            onClick={() => handleMudarAba("calendario")}
            className={`py-2 rounded-lg text-xs font-bold transition-all text-center ${
              abaAtiva === "calendario" ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20" : "text-neutral-400"
            }`}
          >
            MDI
          </button>
        </div>

        {/* ========================================================================= */}
        {/* ABA 1: MERCADO & RANKINGS                                                */}
        {/* ========================================================================= */}
        {abaAtiva === "mercado" && (
          <>
            {/* Banner de Erro */}
            {error && (
              <div className="mb-8 p-4 rounded-xl bg-amber-950/30 border border-amber-500/40 text-amber-200 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <AlertCircle className="w-5 h-5 text-amber-400 shrink-0" />
                  <div>
                    <h4 className="text-sm font-semibold text-white">Servidor Go Desconectado</h4>
                    <p className="text-xs text-amber-300/80">{error}</p>
                  </div>
                </div>
                <button
                  onClick={handleSincronizar}
                  className="px-3.5 py-1.5 rounded-lg bg-amber-500/20 hover:bg-amber-500/30 border border-amber-500/40 text-xs font-semibold text-amber-300 transition-colors shrink-0"
                >
                  Reconectar e Sincronizar
                </button>
              </div>
            )}

            {/* Aviso de Cache Vazio / Pós-Feriado */}
            {!loading && data.total === 0 && !error && (
              <div className="mb-8 p-6 rounded-2xl bg-[#13131a] border border-[#d4af37]/30 text-center flex flex-col items-center justify-center gap-3">
                <ShieldCheck className="w-10 h-10 text-[#d4af37]" />
                <h3 className="text-base font-bold text-white">Nenhum ativo sincronizado no momento</h3>
                <p className="text-xs text-neutral-400 max-w-md">
                  O pregão da B3 esteve fechado por feriado ou o cache local foi reiniciado. Clique no botão abaixo para puxar as cotações oficiais e laudos dos ativos agora.
                </p>
                <button
                  onClick={handleSincronizar}
                  disabled={loading}
                  className="mt-2 flex items-center gap-2 px-4 py-2 rounded-xl bg-[#d4af37] text-black font-bold text-xs hover:bg-[#e5bf43] transition-all"
                >
                  <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
                  Sincronizar Mercado Agora
                </button>
              </div>
            )}

            {/* Hero & Highlights de Mercado */}
            <div className="mb-6 flex flex-col md:flex-row md:items-end justify-between gap-4">
              <div>
                <h1 className="text-2xl sm:text-3xl font-black tracking-tight text-white">
                  Painel de <span className="text-[#d4af37]">Oportunidades</span> B3
                </h1>
                <p className="text-sm text-neutral-400 mt-1 max-w-2xl">
                  Varredura de mercado com valuation por Décio Bazin, Benjamin Graham, Peter Lynch, Gordon e laudos oficiais da CVM.
                </p>
              </div>

              <div className="flex items-center gap-2 text-xs text-neutral-400">
                <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
                <span>3 Tabelas Independentes</span>
                <span className="text-neutral-600">•</span>
                <span className="text-[#d4af37] font-medium">Scores 0-100</span>
              </div>
            </div>

            {/* Barômetro Macroeconômico */}
            {macro && (
              <div className="mb-6 p-4 rounded-2xl bg-gradient-to-r from-[#0d1624]/90 via-[#0f1422]/90 to-[#12121c]/90 border border-cyan-500/25 shadow-lg flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400 shrink-0">
                    <Landmark className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="text-xs font-bold text-white uppercase tracking-wider">
                        Barômetro Macroeconômico & Custo de Oportunidade
                      </h3>
                      <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 font-mono font-semibold border border-emerald-500/30">
                        Curva de Juros Ativa
                      </span>
                    </div>
                    <p className="text-[11px] text-neutral-400 mt-0.5">
                      Taxas oficiais usadas para calibrar preços tetos e exigir prêmio de risco real sobre a renda fixa
                    </p>
                  </div>
                </div>

                <div className="flex flex-wrap items-center gap-4 sm:gap-6 text-xs w-full md:w-auto justify-between md:justify-end border-t md:border-t-0 pt-3 md:pt-0 border-neutral-800 font-mono">
                  <div className="flex flex-col">
                    <span className="text-[10px] text-neutral-400 uppercase">Tesouro IPCA+ (NTN-B)</span>
                    <span className="font-extrabold text-emerald-400 text-sm">{macro.taxa_ntnb.toFixed(2)}% <span className="text-[10px] text-neutral-400 font-normal">a.a. real</span></span>
                  </div>
                  <div className="h-6 w-px bg-neutral-800 hidden sm:block"></div>
                  <div className="flex flex-col">
                    <div className="flex items-center gap-1">
                      <span className="text-[10px] text-neutral-400 uppercase">Selic Meta</span>
                      <span className="text-[8px] px-1 py-0.2 bg-amber-500/20 text-amber-300 rounded font-semibold">Bruta</span>
                    </div>
                    <span className="font-extrabold text-amber-400 text-sm">{macro.taxa_selic.toFixed(2)}% <span className="text-[10px] text-neutral-400 font-normal">a.a.</span></span>
                  </div>
                  <div className="h-6 w-px bg-neutral-800 hidden sm:block"></div>
                  <div className="flex flex-col">
                    <span className="text-[10px] text-neutral-400 uppercase">Hurdle Ações</span>
                    <span className="font-extrabold text-cyan-400 text-sm">{(macro.taxa_ntnb + macro.spread_minimo_acoes).toFixed(2)}% <span className="text-[10px] text-neutral-400 font-normal">a.a.</span></span>
                  </div>
                  <div className="h-6 w-px bg-neutral-800 hidden sm:block"></div>
                  <div className="flex flex-col">
                    <span className="text-[10px] text-neutral-400 uppercase">Hurdle FIIs</span>
                    <span className="font-extrabold text-purple-400 text-sm">{(macro.taxa_ntnb + macro.spread_minimo_fii_tijolo).toFixed(2)}% <span className="text-[10px] text-neutral-400 font-normal">a.a.</span></span>
                  </div>
                </div>
              </div>
            )}

            {/* Cards dos Líderes do Ranking */}
            <SummaryCards
              acoes={data.acoes}
              fiis={data.fiis}
              etfs={data.etfs}
              onSelectAtivo={(ativo) => setAtivoSelecionado(ativo)}
            />

            {/* 1. Tabela de Ações */}
            <section className="mb-12">
              <MarketTable
                titulo="Ranking Completo de Ações"
                subtitulo="Valuation multi-escola: Bazin, Graham, Peter Lynch, Gordon e Greenblatt"
                icone={<TrendingUp className="w-5 h-5" />}
                ativos={data.acoes}
                classe="ACAO"
                onSelectAtivo={(ativo) => setAtivoSelecionado(ativo)}
                onAbrirGrafico={(ativo) => handleAbrirGrafico(ativo.ticker, ativo.nome)}
              />
            </section>

            {/* 2. Tabela de FIIs */}
            <section className="mb-12">
              <MarketTable
                titulo="Ranking Completo de Fundos Imobiliários"
                subtitulo="P/VP sobre laudos CVM e Spread de rendimentos contra o Tesouro IPCA+ (NTN-B)"
                icone={<Building2 className="w-5 h-5" />}
                ativos={data.fiis}
                classe="FII"
                onSelectAtivo={(ativo) => setAtivoSelecionado(ativo)}
                onAbrirGrafico={(ativo) => handleAbrirGrafico(ativo.ticker, ativo.nome)}
              />
            </section>

            {/* 3. Tabela de ETFs */}
            <section className="mb-12">
              <MarketTable
                titulo="Ranking Completo de ETFs"
                subtitulo="Classificação oficial B3 por volume e liquidez institucional diária"
                icone={<Globe className="w-5 h-5" />}
                ativos={data.etfs}
                classe="ETF"
                onSelectAtivo={(ativo) => setAtivoSelecionado(ativo)}
                onAbrirGrafico={(ativo) => handleAbrirGrafico(ativo.ticker, ativo.nome)}
              />
            </section>
          </>
        )}

        {/* ========================================================================= */}
        {/* ABA 2: MINHA CARTEIRA & REBALANCEAMENTO                                  */}
        {/* ========================================================================= */}
        {abaAtiva === "carteira" && (
          <div>
            {!usuario ? (
              /* Estado Deslogado: Call to Action Pedagógico */
              <div className="py-16 text-center max-w-2xl mx-auto space-y-6">
                <div className="w-16 h-16 rounded-3xl bg-[#d4af37]/10 border border-[#d4af37]/30 flex items-center justify-center mx-auto text-[#d4af37] shadow-xl">
                  <Lock className="w-8 h-8" />
                </div>
                <div>
                  <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
                    Gestão de Carteira & Rebalanceamento Buy & Hold
                  </h2>
                  <p className="text-sm text-neutral-400 mt-2">
                    Cadastre suas posições em Ações, FIIs e Renda Fixa com proteção de dados multi-tenant, acompanhe a cotação oficial da B3 e simule aportes inteligentes sem giros desnecessários.
                  </p>
                </div>
                <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
                  <button
                    onClick={() => setModalAuthAberto(true)}
                    className="w-full sm:w-auto px-6 py-3 rounded-xl bg-[#d4af37] hover:bg-[#c49f2f] text-black font-bold text-sm shadow-xl shadow-[#d4af37]/25 flex items-center justify-center gap-2 transition-all"
                  >
                    <span>Entrar ou Criar Minha Conta</span>
                    <ArrowRight className="w-4 h-4" />
                  </button>
                </div>

                <div className="pt-8 border-t border-neutral-900 grid grid-cols-1 sm:grid-cols-3 gap-4 text-left text-xs font-mono text-neutral-400">
                  <div className="p-4 rounded-xl bg-neutral-950 border border-neutral-800/80">
                    <ShieldCheck className="w-4 h-4 text-[#d4af37] mb-2" />
                    <strong className="text-white block mb-1">Cibersegurança OWASP</strong>
                    Isolamento estrito por usuário e hash de senha bcrypt custo 12.
                  </div>
                  <div className="p-4 rounded-xl bg-neutral-950 border border-neutral-800/80">
                    <Scale className="w-4 h-4 text-emerald-400 mb-2" />
                    <strong className="text-white block mb-1">Meta 40/22.5/21/16</strong>
                    Rebalanceamento hierárquico em 2 níveis para estabilidade máxima.
                  </div>
                  <div className="p-4 rounded-xl bg-neutral-950 border border-neutral-800/80">
                    <Briefcase className="w-4 h-4 text-cyan-400 mb-2" />
                    <strong className="text-white block mb-1">Cotações B3 em Tempo Real</strong>
                    Consolidação automática de preços de mercado e preços tetos.
                  </div>
                </div>
              </div>
            ) : (
              /* Estado Logado: Painel da Carteira */
              <div className="space-y-6">
                {/* Header da Carteira & Toolbar de Ações Padronizada */}
                <div className="flex flex-col xl:flex-row xl:items-center justify-between gap-4 border-b border-neutral-800/80 pb-6">
                  <div>
                    <h1 className="text-2xl sm:text-3xl font-black tracking-tight text-white flex items-center gap-2.5">
                      Minha <span className="text-[#d4af37]">Carteira</span> & Alocação
                    </h1>
                    <p className="text-sm text-neutral-400 mt-1">
                      Monitoramento patrimonial com cruzamento instantâneo do motor de valuation oficial da B3.
                    </p>
                  </div>

                  {/* Toolbar Padronizada: mesma altura (h-9), mesma tipografia, ícones harmônicos sem emojis */}
                  <div className="flex items-center gap-2 flex-wrap xl:flex-nowrap">
                    {/* 1. Botão Destaque: Consultor IA */}
                    <button
                      onClick={() => setModalAdvisorAberto(true)}
                      className="h-9 px-3.5 rounded-xl bg-[#d4af37] hover:bg-[#e5c158] text-black font-bold text-xs shadow-md shadow-[#d4af37]/20 flex items-center gap-1.5 transition-all active:scale-95 whitespace-nowrap"
                      title="Consultor Buy & Hold algorítmico com IA"
                    >
                      <Sparkles className="w-3.5 h-3.5" />
                      <span>No que investir hoje?</span>
                    </button>

                    {/* 2. Adicionar Ativo */}
                    <button
                      onClick={() => {
                        setPosicaoParaAporte(null);
                        setModalAddAberto(true);
                      }}
                      className="h-9 px-3 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-200 hover:text-white border border-neutral-800 hover:border-neutral-700 font-semibold text-xs flex items-center gap-1.5 transition-all active:scale-95 whitespace-nowrap"
                    >
                      <Plus className="w-3.5 h-3.5 text-[#d4af37]" />
                      <span>Adicionar Ativo</span>
                    </button>

                    {/* 3. Importar B3 */}
                    <button
                      onClick={() => setModalImportB3Aberto(true)}
                      className="h-9 px-3 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-200 hover:text-white border border-neutral-800 hover:border-emerald-500/40 font-semibold text-xs flex items-center gap-1.5 transition-all active:scale-95 whitespace-nowrap group"
                      title="Importar extrato oficial (.xlsx) da Área do Investidor B3"
                    >
                      <Upload className="w-3.5 h-3.5 text-emerald-400 transition-transform group-hover:-translate-y-0.5" />
                      <span>Importar B3</span>
                    </button>

                    {/* 4. Simulador de Aporte */}
                    <button
                      onClick={() => setModalSimuladorAberto(true)}
                      className="h-9 px-3 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-200 hover:text-white border border-neutral-800 hover:border-neutral-700 font-semibold text-xs flex items-center gap-1.5 transition-all active:scale-95 whitespace-nowrap"
                      title="Simulador de Aportes com base nas metas da carteira"
                    >
                      <Scale className="w-3.5 h-3.5 text-neutral-400" />
                      <span>Simulador</span>
                    </button>

                    {/* 5. Extrato Contábil */}
                    <button
                      onClick={() => {
                        setTickerHistoricoFiltro(null);
                        setModalHistoricoAberto(true);
                      }}
                      className="h-9 px-3 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-200 hover:text-white border border-neutral-800 hover:border-neutral-700 font-semibold text-xs flex items-center gap-1.5 transition-all active:scale-95 whitespace-nowrap"
                      title="Ver extrato e histórico de aportes e transações"
                    >
                      <History className="w-3.5 h-3.5 text-neutral-400" />
                      <span>Extrato</span>
                    </button>

                    {/* 6. Modo Privacidade: Ocultar/Exibir Valores com Efeito Blur */}
                    <button
                      onClick={toggleOcultarValores}
                      className={`h-9 px-3 rounded-xl border font-semibold text-xs flex items-center gap-1.5 transition-all active:scale-95 whitespace-nowrap ${
                        ocultarValores
                          ? "bg-[#d4af37]/15 text-[#d4af37] border-[#d4af37]/40 shadow-sm shadow-[#d4af37]/10"
                          : "bg-neutral-900 hover:bg-neutral-800 text-neutral-400 hover:text-white border-neutral-800 hover:border-neutral-700"
                      }`}
                      title={
                        ocultarValores
                          ? "Exibir valores da carteira"
                          : "Ocultar valores em reais (Efeito Blur de Privacidade)"
                      }
                    >
                      {ocultarValores ? (
                        <>
                          <EyeOff className="w-3.5 h-3.5 text-[#d4af37]" />
                          <span>Valores Ocultos</span>
                        </>
                      ) : (
                        <>
                          <Eye className="w-3.5 h-3.5 text-neutral-400" />
                          <span>Ocultar Valores</span>
                        </>
                      )}
                    </button>

                    {/* 7. Atualizar Carteira (Sync) */}
                    <button
                      onClick={carregarCarteira}
                      disabled={loadingCarteira}
                      className="h-9 w-9 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-400 hover:text-white border border-neutral-800 hover:border-neutral-700 flex items-center justify-center transition-all active:scale-95 shrink-0"
                      title="Atualizar cotações e métricas da carteira"
                    >
                      <RefreshCw className={`w-3.5 h-3.5 ${loadingCarteira ? "animate-spin text-[#d4af37]" : ""}`} />
                    </button>
                  </div>
                </div>

                {erroCarteira && (
                  <div className="p-4 rounded-xl bg-red-950/40 border border-red-500/40 text-red-200 text-xs flex items-center gap-2">
                    <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />
                    <span>{erroCarteira}</span>
                  </div>
                )}

                {resumoCarteira ? (
                  <>
                    {/* 1. Cards de Resumo */}
                    <PortfolioSummaryCards
                      resumo={resumoCarteira}
                      ocultarValores={ocultarValores}
                    />

                    {/* 2. Gráfico Donut de Composição & Foco Interativo */}
                    <PortfolioDonutChart
                      resumo={resumoCarteira}
                      classeSelecionada={classeFiltroDonut}
                      ocultarValores={ocultarValores}
                      onSelecionarClasse={(cl) => {
                        setClasseFiltroDonut((prev) => (prev === cl ? null : cl));
                      }}
                    />

                    {/* 3. Tabela de Ativos em Custódia com Decomposição em Níveis */}
                    <PortfolioTable
                      posicoes={resumoCarteira.posicoes}
                      classes={resumoCarteira.classes}
                      patrimonioTotal={resumoCarteira.patrimonio_total}
                      classeFiltroDonut={classeFiltroDonut}
                      ocultarValores={ocultarValores}
                      onExcluirPosicao={handleExcluirPosicao}
                      onAportar={(pos) => {
                        setPosicaoParaAporte(pos);
                        setModalAddAberto(true);
                      }}
                      onVerHistorico={(ticker) => {
                        setTickerHistoricoFiltro(ticker || null);
                        setModalHistoricoAberto(true);
                      }}
                      onAbrirGrafico={(ticker) => handleAbrirGrafico(ticker)}
                    />

                    {/* 4. Macro Alocação das 5 Caixinhas */}
                    <MacroAllocationView
                      resumo={resumoCarteira}
                      ocultarValores={ocultarValores}
                    />
                  </>
                ) : (
                  <div className="p-12 text-center text-neutral-400 font-mono text-xs">
                    Carregando carteira de investimentos...
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* ========================================================================= */}
        {/* ABA 3: PROVENTOS & RENDA PASSIVA                                         */}
        {/* ========================================================================= */}
        {abaAtiva === "proventos" && (
          <ProventosView
            token={token}
            ocultarValores={ocultarValores}
            onAbrirImportadorB3={() => setModalImportB3Aberto(true)}
            onAbrirLogin={() => setModalAuthAberto(true)}
          />
        )}

        {/* ========================================================================= */}
        {/* ABA 4: CALENDÁRIO MDI (MAPA DE DIVIDENDOS INTELIGENTE)                    */}
        {/* ========================================================================= */}
        {abaAtiva === "calendario" && (
          <DividendCalendar
            onSelecionarAtivo={(ativo) => setAtivoSelecionado(ativo)}
            todosAtivosMercado={[...data.acoes, ...data.fiis, ...data.etfs]}
          />
        )}
      </main>

      {/* Modais da Aplicação */}
      <AssetDetailModal
        ativo={ativoSelecionado}
        macro={macro}
        onClose={() => setAtivoSelecionado(null)}
        onAbrirGrafico={(ticker) => handleAbrirGrafico(ticker, ativoSelecionado?.nome)}
      />

      <AddPositionModal
        isOpen={modalAddAberto}
        onClose={() => {
          setModalAddAberto(false);
          setPosicaoParaAporte(null);
        }}
        onSalvar={handleSalvarPosicao}
        ativosDisponiveis={[...data.acoes, ...data.fiis, ...data.etfs]}
        posicaoParaAporte={posicaoParaAporte}
        posicoesAtuais={resumoCarteira?.posicoes || []}
      />

      <RebalanceSimulatorModal
        isOpen={modalSimuladorAberto}
        onClose={() => setModalSimuladorAberto(false)}
        onSimular={handleSimularRebalanceamento}
      />

      <AdvisorModal
        isOpen={modalAdvisorAberto}
        onClose={() => setModalAdvisorAberto(false)}
        token={token}
        onAportarEmAtivo={(ticker, quantidade, precoEstimado, classe) => {
          setPosicaoParaAporte({
            id: 0,
            usuario_id: usuario?.id || 0,
            ticker: ticker,
            classe: classe,
            categoria_macro: "Aporte Recomendado por IA",
            quantidade: quantidade,
            preco_medio: precoEstimado,
            valor_aplicado: quantidade * precoEstimado,
            corretora: "Principal",
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            cotacao_atual: precoEstimado,
            valor_atual_mercado: quantidade * precoEstimado,
            lucro_prejuizo_valor: 0,
            rentabilidade_pct: 0,
            preco_teto_estimado: 0,
            margem_seguranca_pct: 0,
            score_fundamentalista: 100,
            sinal_rebalanceamento: "COMPRAR_APORTE",
          });
          setModalAddAberto(true);
        }}
      />

      <TransactionHistoryModal
        isOpen={modalHistoricoAberto}
        onClose={() => {
          setModalHistoricoAberto(false);
          setTickerHistoricoFiltro(null);
        }}
        tickerFiltroInicial={tickerHistoricoFiltro}
      />

      <B3ImportModal
        isOpen={modalImportB3Aberto}
        onClose={() => setModalImportB3Aberto(false)}
        token={token}
        onImportadoSucesso={() => {
          carregarCarteira();
        }}
      />

      <AuthModal
        isOpen={modalAuthAberto}
        onClose={() => setModalAuthAberto(false)}
      />

      <AssetChartModal
        isOpen={modalGraficoAberto}
        onClose={() => {
          setModalGraficoAberto(false);
          setTickerGrafico(null);
        }}
        ticker={tickerGrafico}
        nomeInicial={nomeGrafico}
      />

      {/* Rodapé Minimalista */}
      <footer className="border-t border-neutral-900 bg-[#060608] py-6 text-center text-xs text-neutral-500">
        <p>B3 Core Platform • Desenvolvido com Go e Next.js • Dados Oficiais B3 & CVM</p>
      </footer>
    </div>
  );
}
