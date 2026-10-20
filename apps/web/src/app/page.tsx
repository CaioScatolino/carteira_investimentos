"use client";

import React, { useEffect, useState, useCallback } from "react";
import { Ativo, RespostaRankings, CenarioMacro } from "@/types/market";
import { ResumoCarteira, CriarPosicaoRequest, ResultadoRebalanceamento } from "@/types/portfolio";
import { Header } from "@/components/Header";
import { SummaryCards } from "@/components/SummaryCards";
import { MarketTable } from "@/components/MarketTable";
import { AssetDetailModal } from "@/components/AssetDetailModal";
import { PortfolioSummaryCards } from "@/components/PortfolioSummaryCards";
import { MacroAllocationView } from "@/components/MacroAllocationView";
import { PortfolioTable } from "@/components/PortfolioTable";
import { AddPositionModal } from "@/components/AddPositionModal";
import { RebalanceSimulatorModal } from "@/components/RebalanceSimulatorModal";
import { AuthModal } from "@/components/AuthModal";
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
} from "lucide-react";

export default function Dashboard() {
  const { usuario, token, loading: authLoading } = useAuth();

  // Estado de navegação
  const [abaAtiva, setAbaAtiva] = useState<"mercado" | "carteira">("mercado");

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

  // Modais
  const [modalAddAberto, setModalAddAberto] = useState(false);
  const [modalSimuladorAberto, setModalSimuladorAberto] = useState(false);
  const [modalAuthAberto, setModalAuthAberto] = useState(false);

  // 1. Carrega dados do mercado (Rankings B3 & Macro)
  const carregarRankings = async () => {
    setLoading(true);
    setError(null);
    try {
      const [resRankings, resMacro] = await Promise.all([
        fetch("http://localhost:8080/api/v1/rankings", { headers: { Accept: "application/json" } }),
        fetch("http://localhost:8080/api/v1/macro", { headers: { Accept: "application/json" } }).catch(() => null),
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
        "Não foi possível conectar ao servidor em http://localhost:8080. Verifique se o backend Go está rodando."
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
      const res = await fetch("http://localhost:8080/api/v1/carteira", {
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

    const res = await fetch("http://localhost:8080/api/v1/carteira/posicoes", {
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
      const res = await fetch(`http://localhost:8080/api/v1/carteira/posicoes/${id}`, {
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

    const res = await fetch("http://localhost:8080/api/v1/carteira/rebalancear", {
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

  return (
    <div className="min-h-screen bg-[#070709] text-neutral-100 flex flex-col selection:bg-[#d4af37]/30 selection:text-[#f5df88]">
      {/* Header Fixo com Seletor de Abas */}
      <Header
        totalAtivos={data.total}
        loading={loading || loadingCarteira}
        onRefresh={() => {
          if (abaAtiva === "mercado") carregarRankings();
          else carregarCarteira();
        }}
        ultimaAtualizacao={ultimaAtualizacao}
        abaAtiva={abaAtiva}
        onMudarAba={(aba) => setAbaAtiva(aba)}
      />

      {/* Conteúdo Principal */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Seletor de Abas Mobile */}
        <div className="flex md:hidden items-center gap-2 mb-6 p-1 rounded-xl bg-neutral-900 border border-neutral-800">
          <button
            onClick={() => setAbaAtiva("mercado")}
            className={`flex-1 py-2 rounded-lg text-xs font-bold transition-all ${
              abaAtiva === "mercado" ? "bg-[#d4af37] text-black" : "text-neutral-400"
            }`}
          >
            Mercado B3
          </button>
          <button
            onClick={() => setAbaAtiva("carteira")}
            className={`flex-1 py-2 rounded-lg text-xs font-bold transition-all ${
              abaAtiva === "carteira" ? "bg-[#d4af37] text-black" : "text-neutral-400"
            }`}
          >
            Minha Carteira
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
                  onClick={carregarRankings}
                  className="px-3.5 py-1.5 rounded-lg bg-amber-500/20 hover:bg-amber-500/30 border border-amber-500/40 text-xs font-semibold text-amber-300 transition-colors shrink-0"
                >
                  Reconectar Agora
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
                {/* Header da Carteira & Ações Rápidas */}
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-neutral-800 pb-6">
                  <div>
                    <h1 className="text-2xl sm:text-3xl font-black tracking-tight text-white flex items-center gap-2.5">
                      Minha <span className="text-[#d4af37]">Carteira</span> & Alocação
                    </h1>
                    <p className="text-sm text-neutral-400 mt-1">
                      Monitoramento patrimonial com cruzamento instantâneo do motor de valuation oficial da B3.
                    </p>
                  </div>

                  <div className="flex items-center gap-3 flex-wrap">
                    <button
                      onClick={() => setModalAddAberto(true)}
                      className="px-4 py-2.5 rounded-xl bg-[#d4af37] hover:bg-[#c49f2f] text-black font-bold text-xs shadow-lg shadow-[#d4af37]/20 flex items-center gap-2 transition-all"
                    >
                      <Plus className="w-4 h-4" />
                      <span>Adicionar Ativo</span>
                    </button>

                    <button
                      onClick={() => setModalSimuladorAberto(true)}
                      className="px-4 py-2.5 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-white font-bold text-xs border border-neutral-700 flex items-center gap-2 transition-all"
                    >
                      <Scale className="w-4 h-4 text-[#d4af37]" />
                      <span>Simulador de Aporte</span>
                    </button>

                    <button
                      onClick={carregarCarteira}
                      disabled={loadingCarteira}
                      className="p-2.5 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-400 hover:text-white border border-neutral-800 transition-colors"
                      title="Atualizar Carteira"
                    >
                      <RefreshCw className={`w-4 h-4 ${loadingCarteira ? "animate-spin text-[#d4af37]" : ""}`} />
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
                    <PortfolioSummaryCards resumo={resumoCarteira} />

                    {/* 2. Macro Alocação das 5 Caixinhas */}
                    <MacroAllocationView resumo={resumoCarteira} />

                    {/* 3. Tabela de Ativos em Custódia */}
                    <PortfolioTable
                      posicoes={resumoCarteira.posicoes}
                      onExcluirPosicao={handleExcluirPosicao}
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
      </main>

      {/* Modais da Aplicação */}
      <AssetDetailModal
        ativo={ativoSelecionado}
        macro={macro}
        onClose={() => setAtivoSelecionado(null)}
      />

      <AddPositionModal
        isOpen={modalAddAberto}
        onClose={() => setModalAddAberto(false)}
        onSalvar={handleSalvarPosicao}
      />

      <RebalanceSimulatorModal
        isOpen={modalSimuladorAberto}
        onClose={() => setModalSimuladorAberto(false)}
        onSimular={handleSimularRebalanceamento}
      />

      <AuthModal
        isOpen={modalAuthAberto}
        onClose={() => setModalAuthAberto(false)}
      />

      {/* Rodapé Minimalista */}
      <footer className="border-t border-neutral-900 bg-[#060608] py-6 text-center text-xs text-neutral-500">
        <p>B3 Core Platform • Desenvolvido com Go e Next.js • Dados Oficiais B3 & CVM</p>
      </footer>
    </div>
  );
}
