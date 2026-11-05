"use client";

import React, { useState, useEffect } from "react";
import {
  Coins,
  TrendingUp,
  Calendar,
  Layers,
  ArrowUpRight,
  Clock,
  Sparkles,
  AlertCircle,
  Loader2,
  CheckCircle2,
  Building2,
  DollarSign,
  ChevronRight,
  Lock,
} from "lucide-react";
import { API_V1_URL } from "@/config/api";
import { ResumoProventos, ProventoFuturo } from "@/types/portfolio";

interface ProventosViewProps {
  token: string | null;
  onAbrirImportadorB3?: () => void;
  onAbrirLogin?: () => void;
}

export function ProventosView({ token, onAbrirImportadorB3, onAbrirLogin }: ProventosViewProps) {
  const [resumo, setResumo] = useState<ResumoProventos | null>(null);
  const [radarFuturo, setRadarFuturo] = useState<ProventoFuturo[]>([]);
  const [loading, setLoading] = useState(true);
  const [erro, setErro] = useState<string | null>(null);

  const carregarDados = async () => {
    if (!token) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setErro(null);

    try {
      const [resHistorico, resRadar] = await Promise.all([
        fetch(`${API_V1_URL}/carteira/proventos/historico`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
        fetch(`${API_V1_URL}/carteira/proventos/radar-futuro`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
      ]);

      if (resHistorico.ok) {
        const dataHist: ResumoProventos = await resHistorico.json();
        setResumo(dataHist);
      }

      if (resRadar.ok) {
        const dataRadar = await resRadar.json();
        setRadarFuturo(dataRadar.proventos || []);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Falha ao carregar proventos";
      setErro(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    carregarDados();
  }, [token]);

  const formatarMoeda = (val: number) => {
    return new Intl.NumberFormat("pt-BR", {
      style: "currency",
      currency: "BRL",
    }).format(val);
  };

  if (!token) {
    return (
      <div className="py-16 text-center max-w-2xl mx-auto space-y-6">
        <div className="w-16 h-16 rounded-3xl bg-[#d4af37]/10 border border-[#d4af37]/30 flex items-center justify-center mx-auto text-[#d4af37] shadow-xl">
          <Lock className="w-8 h-8" />
        </div>
        <div>
          <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
            Histórico de Proventos & Renda Passiva
          </h2>
          <p className="text-sm text-neutral-400 mt-2">
            Consolide todos os seus rendimentos de FIIs, dividendos e JCP recebidos na B3 desde 2019, acompanhe a sazonalidade mês a mês e consulte o radar preditivo de futuros proventos.
          </p>
        </div>
        {onAbrirLogin && (
          <div className="flex items-center justify-center pt-2">
            <button
              onClick={onAbrirLogin}
              className="px-6 py-3 rounded-xl bg-[#d4af37] hover:bg-[#c49f2f] text-black font-bold text-sm shadow-xl shadow-[#d4af37]/25 transition-all font-mono"
            >
              Entrar ou Criar Minha Conta
            </button>
          </div>
        )}
      </div>
    );
  }

  if (loading) {
    return (
      <div className="py-20 flex flex-col items-center justify-center gap-3 text-neutral-400">
        <Loader2 className="w-8 h-8 animate-spin text-[#d4af37]" />
        <span className="text-xs font-mono">Consolidando histórico de dividendos e proventos...</span>
      </div>
    );
  }

  const temProventos = resumo && resumo.total_acumulado > 0;

  if (!temProventos && radarFuturo.length === 0) {
    return (
      <div className="py-16 px-4 text-center rounded-3xl bg-neutral-950/60 border border-neutral-800/80 max-w-xl mx-auto space-y-4">
        <div className="w-14 h-14 rounded-2xl bg-[#d4af37]/10 border border-[#d4af37]/30 flex items-center justify-center mx-auto text-[#d4af37]">
          <Coins className="w-7 h-7" />
        </div>
        <div className="space-y-1">
          <h4 className="font-extrabold text-white text-base font-mono">
            Nenhum Provento Registrado Ainda
          </h4>
          <p className="text-xs text-neutral-400">
            Importe seu extrato da Área do Investidor B3 para carregar automaticamente todos os rendimentos, dividendos e JCP recebidos desde 2019.
          </p>
        </div>
        {onAbrirImportadorB3 && (
          <button
            onClick={onAbrirImportadorB3}
            className="px-5 py-2.5 rounded-xl bg-[#d4af37] hover:bg-[#c29e2e] text-black font-extrabold text-xs shadow-lg shadow-[#d4af37]/20 transition-all font-mono"
          >
            📥 Importar Extrato da B3 Agora
          </button>
        )}
      </div>
    );
  }

  // Maior valor anual para normalizar as barras
  const maxAno = resumo?.total_por_ano.reduce((acc, item) => Math.max(acc, item.valor), 1) || 1;
  const maxMes = resumo?.sazonalidade_mensal.reduce((acc, item) => Math.max(acc, item.valor), 1) || 1;

  // Total futuro previsto
  const totalFuturoPrevisto = radarFuturo.reduce((acc, item) => acc + item.valor_previsto, 0);

  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      
      {/* 1. Header Cards de Performance de Renda Passiva */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3.5 font-mono">
        <div className="p-4 sm:p-5 rounded-2xl bg-neutral-950 border border-[#d4af37]/30 shadow-lg shadow-black/40">
          <div className="flex items-center justify-between text-neutral-400 mb-2">
            <span className="text-[10px] uppercase tracking-wider">Renda Passiva Acumulada</span>
            <Coins className="w-4 h-4 text-[#d4af37]" />
          </div>
          <span className="text-xl sm:text-2xl font-extrabold text-white block">
            {formatarMoeda(resumo?.total_acumulado || 0)}
          </span>
          <span className="text-[10px] text-emerald-400 mt-1 block">
            Desde 2019 na B3
          </span>
        </div>

        <div className="p-4 sm:p-5 rounded-2xl bg-neutral-950 border border-neutral-800">
          <div className="flex items-center justify-between text-neutral-400 mb-2">
            <span className="text-[10px] uppercase tracking-wider">Média Mensal (12M)</span>
            <Calendar className="w-4 h-4 text-emerald-400" />
          </div>
          <span className="text-xl sm:text-2xl font-extrabold text-emerald-400 block">
            {formatarMoeda(resumo?.media_mensal_12m || 0)}
          </span>
          <span className="text-[10px] text-neutral-500 mt-1 block">
            Fluxo mensal médio recente
          </span>
        </div>

        <div className="p-4 sm:p-5 rounded-2xl bg-neutral-950 border border-neutral-800">
          <div className="flex items-center justify-between text-neutral-400 mb-2">
            <span className="text-[10px] uppercase tracking-wider">Recebido em 12 Meses</span>
            <TrendingUp className="w-4 h-4 text-blue-400" />
          </div>
          <span className="text-xl sm:text-2xl font-extrabold text-white block">
            {formatarMoeda(resumo?.total_recebido_12m || 0)}
          </span>
          <span className="text-[10px] text-neutral-500 mt-1 block">
            Últimos 12 meses fechados
          </span>
        </div>

        <div className="p-4 sm:p-5 rounded-2xl bg-neutral-950 border border-amber-500/30">
          <div className="flex items-center justify-between text-neutral-400 mb-2">
            <span className="text-[10px] uppercase tracking-wider text-amber-400">Futuros Anunciados</span>
            <Clock className="w-4 h-4 text-amber-400" />
          </div>
          <span className="text-xl sm:text-2xl font-extrabold text-amber-400 block">
            {formatarMoeda(totalFuturoPrevisto)}
          </span>
          <span className="text-[10px] text-neutral-500 mt-1 block">
            {radarFuturo.length} provento(s) em carteira
          </span>
        </div>
      </div>

      {/* 2. Radar de Próximos Rendimentos Futuros */}
      {radarFuturo.length > 0 && (
        <div className="p-5 rounded-3xl bg-neutral-950/80 border border-amber-500/30 shadow-xl space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-amber-400" />
              <h4 className="font-extrabold text-sm sm:text-base text-white font-mono">
                Radar de Próximos Rendimentos & Dividendos Anunciados
              </h4>
            </div>
            <span className="text-[11px] font-mono text-amber-400 font-bold">
              Total Previsto: {formatarMoeda(totalFuturoPrevisto)}
            </span>
          </div>

          <div className="overflow-x-auto custom-scrollbar">
            <table className="w-full text-left text-xs font-mono">
              <thead className="bg-[#0e0e14] border-b border-neutral-800 text-[10px] text-neutral-400 uppercase tracking-wider">
                <tr>
                  <th className="p-2.5">Ticker</th>
                  <th className="p-2.5">Tipo</th>
                  <th className="p-2.5">Data-Com</th>
                  <th className="p-2.5">Pagamento</th>
                  <th className="p-2.5 text-right">Cotas</th>
                  <th className="p-2.5 text-right">Valor/Cota</th>
                  <th className="p-2.5 text-right text-amber-400">Total Previsto</th>
                  <th className="p-2.5 text-center">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-900">
                {radarFuturo.map((rf, idx) => (
                  <tr key={idx} className="hover:bg-neutral-900/50 transition-colors">
                    <td className="p-2.5 font-extrabold text-white text-sm">{rf.ticker}</td>
                    <td className="p-2.5 text-neutral-300">{rf.tipo}</td>
                    <td className="p-2.5 text-neutral-400">{rf.data_com || "-"}</td>
                    <td className="p-2.5 text-white font-bold">{rf.data_pagamento}</td>
                    <td className="p-2.5 text-right text-neutral-300">{rf.quantidade_cotas}</td>
                    <td className="p-2.5 text-right text-neutral-300">{formatarMoeda(rf.valor_unitario)}</td>
                    <td className="p-2.5 text-right font-extrabold text-amber-400">{formatarMoeda(rf.valor_previsto)}</td>
                    <td className="p-2.5 text-center">
                      <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold ${
                        rf.status === "DATA_COM_IMINENTE"
                          ? "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                          : "bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                      }`}>
                        {rf.status === "DATA_COM_IMINENTE" ? "Data-Com Iminente" : "Confirmado"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 3. Gráfico de Evolução Anual & Matriz de Sazonalidade */}
      {resumo && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 font-mono">
          
          {/* Evolução por Ano */}
          <div className="p-5 rounded-3xl bg-neutral-950 border border-neutral-800 space-y-4">
            <div className="flex items-center justify-between">
              <span className="font-extrabold text-sm text-white flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-emerald-400" />
                Evolução Anual da Renda Passiva
              </span>
              <span className="text-[10px] text-neutral-500 uppercase">Crescimento Acumulado</span>
            </div>

            <div className="space-y-3 pt-2">
              {resumo.total_por_ano.map((anoItem) => {
                const pct = (anoItem.valor / maxAno) * 100;
                return (
                  <div key={anoItem.ano} className="space-y-1">
                    <div className="flex items-center justify-between text-xs">
                      <span className="font-bold text-neutral-300">{anoItem.ano}</span>
                      <span className="font-extrabold text-white">{formatarMoeda(anoItem.valor)}</span>
                    </div>
                    <div className="h-3 w-full bg-neutral-900 rounded-full overflow-hidden p-0.5">
                      <div
                        className="h-full bg-gradient-to-r from-[#d4af37] to-emerald-400 rounded-full transition-all duration-500"
                        style={{ width: `${Math.max(pct, 4)}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Sazonalidade Mensal (Jan a Dez) */}
          <div className="p-5 rounded-3xl bg-neutral-950 border border-neutral-800 space-y-4">
            <div className="flex items-center justify-between">
              <span className="font-extrabold text-sm text-white flex items-center gap-2">
                <Calendar className="w-4 h-4 text-[#d4af37]" />
                Sazonalidade Mensal (Histórico)
              </span>
              <span className="text-[10px] text-neutral-500 uppercase">Meses mais fortes</span>
            </div>

            <div className="grid grid-cols-3 sm:grid-cols-4 gap-2 pt-2">
              {resumo.sazonalidade_mensal.map((m) => (
                <div key={m.mes} className="p-2.5 rounded-xl bg-neutral-900/60 border border-neutral-800/80 text-center space-y-1">
                  <span className="text-[10px] text-neutral-400 uppercase tracking-wider block font-bold">{m.nome.slice(0, 3)}</span>
                  <span className="text-xs font-extrabold text-emerald-400 block">{formatarMoeda(m.valor)}</span>
                </div>
              ))}
            </div>
          </div>

        </div>
      )}

      {/* 4. Maiores Pagadores & Últimos Pagamentos */}
      {resumo && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 font-mono">
          
          {/* Top Ativos Pagadores */}
          <div className="p-5 rounded-3xl bg-neutral-950 border border-neutral-800 space-y-4">
            <span className="font-extrabold text-sm text-white flex items-center gap-2">
              <Coins className="w-4 h-4 text-[#d4af37]" />
              Maiores Pagadores da Sua Carteira
            </span>

            <div className="space-y-2 max-h-80 overflow-y-auto custom-scrollbar pr-1">
              {resumo.top_ativos_pagadores.slice(0, 10).map((at, idx) => (
                <div
                  key={idx}
                  className="p-3 rounded-xl bg-neutral-900/50 border border-neutral-800/60 flex items-center justify-between text-xs"
                >
                  <div className="flex items-center gap-3">
                    <span className="w-6 text-center font-bold text-neutral-500 text-[11px]">{idx + 1}º</span>
                    <div>
                      <span className="font-extrabold text-white text-sm block">{at.ticker}</span>
                      <span className="text-[10px] text-neutral-500">{at.total_eventos} pagamentos</span>
                    </div>
                  </div>
                  <div className="text-right">
                    <span className="font-extrabold text-emerald-400 block">{formatarMoeda(at.total_recebido)}</span>
                    <span className="text-[10px] text-neutral-400 font-semibold">{at.percentual}% do total</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Últimos Pagamentos Recebidos */}
          <div className="p-5 rounded-3xl bg-neutral-950 border border-neutral-800 space-y-4">
            <span className="font-extrabold text-sm text-white flex items-center gap-2">
              <Clock className="w-4 h-4 text-blue-400" />
              Últimos Pagamentos Creditados na Conta
            </span>

            <div className="space-y-2 max-h-80 overflow-y-auto custom-scrollbar pr-1">
              {resumo.ultimos_pagamentos.slice(0, 10).map((p, idx) => (
                <div
                  key={idx}
                  className="p-3 rounded-xl bg-neutral-900/50 border border-neutral-800/60 flex items-center justify-between text-xs"
                >
                  <div className="space-y-0.5">
                    <div className="flex items-center gap-2">
                      <span className="font-extrabold text-white text-sm">{p.ticker}</span>
                      <span className="text-[10px] px-1.5 py-0.2 rounded bg-neutral-800 text-neutral-300 font-semibold">{p.tipo}</span>
                    </div>
                    <span className="text-[10px] text-neutral-500">{p.data_pagamento} • {p.corretora}</span>
                  </div>
                  <div className="text-right font-extrabold text-emerald-400">
                    +{formatarMoeda(p.valor_liquido)}
                  </div>
                </div>
              ))}
            </div>
          </div>

        </div>
      )}

    </div>
  );
}

