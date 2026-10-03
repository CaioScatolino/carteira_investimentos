"use client";

import React, { useEffect, useState } from "react";
import { Ativo, RespostaRankings } from "@/types/market";
import { Header } from "@/components/Header";
import { SummaryCards } from "@/components/SummaryCards";
import { MarketTable } from "@/components/MarketTable";
import { AssetDetailModal } from "@/components/AssetDetailModal";
import { TrendingUp, Building2, Globe, AlertCircle, RefreshCw } from "lucide-react";

export default function Dashboard() {
  const [data, setData] = useState<RespostaRankings>({
    total: 0,
    acoes: [],
    fiis: [],
    etfs: [],
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [ativoSelecionado, setAtivoSelecionado] = useState<Ativo | null>(null);
  const [ultimaAtualizacao, setUltimaAtualizacao] = useState<string>("");

  const carregarRankings = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("http://localhost:8080/api/v1/rankings", {
        headers: { Accept: "application/json" },
      });
      if (!res.ok) {
        throw new Error(`Servidor respondeu com status ${res.status}`);
      }
      const json: RespostaRankings = await res.json();
      setData({
        total: json.total ?? 0,
        acoes: json.acoes ?? [],
        fiis: json.fiis ?? [],
        etfs: json.etfs ?? [],
      });
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

  useEffect(() => {
    carregarRankings();
  }, []);

  return (
    <div className="min-h-screen bg-[#070709] text-neutral-100 flex flex-col selection:bg-[#d4af37]/30 selection:text-[#f5df88]">
      {/* Header Fixo */}
      <Header
        totalAtivos={data.total}
        loading={loading}
        onRefresh={carregarRankings}
        ultimaAtualizacao={ultimaAtualizacao}
      />

      {/* Conteúdo Principal */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Banner de Erro com Ação de Tentar Novamente */}
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
      </main>

      {/* Modal de Detalhes do Ativo */}
      <AssetDetailModal
        ativo={ativoSelecionado}
        onClose={() => setAtivoSelecionado(null)}
      />

      {/* Rodapé Minimalista */}
      <footer className="border-t border-neutral-900 bg-[#060608] py-6 text-center text-xs text-neutral-500">
        <p>B3 Core Platform • Desenvolvido com Go e Next.js • Dados Oficiais B3 & CVM</p>
      </footer>
    </div>
  );
}
