"use client";

import React, { useState, useEffect, useRef } from "react";
import {
  Sparkles,
  X,
  TrendingUp,
  Building2,
  ShieldCheck,
  Scale,
  Zap,
  CheckCircle2,
  AlertCircle,
  Plus,
  RefreshCw,
  Coins,
  ArrowRight,
  Lightbulb,
} from "lucide-react";
import { API_V1_URL } from "@/config/api";
import {
  ClasseAtivo,
  DiagnosticoAdvisor,
  PerfilAdvisor,
  SugestaoAporteItem,
} from "@/types/portfolio";

interface AdvisorModalProps {
  isOpen: boolean;
  onClose: () => void;
  token: string | null;
  onAportarEmAtivo?: (
    ticker: string,
    quantidade: number,
    precoEstimado: number,
    classe: ClasseAtivo
  ) => void;
}

export function AdvisorModal({
  isOpen,
  onClose,
  token,
  onAportarEmAtivo,
}: AdvisorModalProps) {
  const [valorAporte, setValorAporte] = useState<number>(1000);
  const [perfil, setPerfil] = useState<PerfilAdvisor>("EQUILIBRADO");
  const [permitirNovos, setPermitirNovos] = useState<boolean>(true);

  const [executando, setExecutando] = useState<boolean>(false);
  const [streamingTexto, setStreamingTexto] = useState<string>("");
  const [diagnostico, setDiagnostico] = useState<DiagnosticoAdvisor | null>(null);
  const [erro, setErro] = useState<string | null>(null);

  const eventSourceRef = useRef<EventSource | null>(null);
  const scrollRef = useRef<HTMLDivElement | null>(null);

  // Fecha eventSource se fechar modal
  useEffect(() => {
    if (!isOpen) {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      setExecutando(false);
    }
  }, [isOpen]);

  // Auto-scroll do texto da IA enquanto digita
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [streamingTexto]);

  if (!isOpen) return null;

  const valoresAtalhos = [200, 500, 1000, 2500, 5000];

  const iniciarConsulta = () => {
    if (!token) {
      setErro("Usuário não autenticado. Faça login para consultar seu assistente.");
      return;
    }
    if (valorAporte <= 0) {
      setErro("Informe um valor de aporte maior que zero.");
      return;
    }

    setErro(null);
    setStreamingTexto("");
    setDiagnostico(null);
    setExecutando(true);

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    const url = `${API_V1_URL}/advisor/stream?token=${encodeURIComponent(
      token
    )}&valor=${valorAporte}&perfil=${perfil}&permitir_novos=${permitirNovos}`;

    const es = new EventSource(url);
    eventSourceRef.current = es;

    es.addEventListener("diagnostico", (event) => {
      try {
        const dados: DiagnosticoAdvisor = JSON.parse(event.data);
        setDiagnostico(dados);
      } catch (e) {
        console.warn("Erro ao fazer parse do diagnóstico SSE:", e);
      }
    });

    es.addEventListener("token", (event) => {
      setStreamingTexto((prev) => prev + event.data);
    });

    es.addEventListener("done", () => {
      es.close();
      eventSourceRef.current = null;
      setExecutando(false);
    });

    es.addEventListener("error", (event: any) => {
      console.warn("SSE error ou conexão encerrada:", event);
      if (es.readyState === EventSource.CLOSED) {
        es.close();
        eventSourceRef.current = null;
        setExecutando(false);
      }
    });
  };

  const formatarMoeda = (val: number) => {
    return new Intl.NumberFormat("pt-BR", {
      style: "currency",
      currency: "BRL",
    }).format(val || 0);
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/85 backdrop-blur-md flex items-center justify-center p-3 sm:p-6 overflow-y-auto animate-in fade-in duration-200">
      <div className="relative w-full max-w-4xl bg-[#09090d] border border-[#d4af37]/40 rounded-2xl sm:rounded-3xl shadow-2xl shadow-[#d4af37]/10 flex flex-col max-h-[92vh] overflow-hidden">
        {/* Top Header */}
        <div className="p-5 sm:p-6 border-b border-neutral-800/80 bg-gradient-to-r from-[#12121a] via-[#0d0d14] to-[#09090d] flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#d4af37] via-[#a8892d] to-[#0a0a0d] p-[1px] shadow-lg shadow-[#d4af37]/20 flex items-center justify-center">
              <div className="w-full h-full rounded-xl bg-[#0b0b0f] flex items-center justify-center">
                <Sparkles className="w-5 h-5 text-[#d4af37] animate-pulse" />
              </div>
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-base sm:text-lg font-black tracking-wide text-white">
                  Assistente Cognitivo de Aportes
                </h3>
                <span className="px-2 py-0.5 rounded-full bg-[#d4af37]/15 text-[#d4af37] border border-[#d4af37]/30 text-[10px] font-bold font-mono uppercase tracking-wider">
                  IA Buy & Hold
                </span>
              </div>
              <p className="text-xs text-neutral-400">
                Auditoria algorítmica de defasagem patrimonial, assimetrias de valuation e recomendação precisa.
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-2 rounded-xl text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Scrollable Content Body */}
        <div className="p-5 sm:p-6 overflow-y-auto space-y-6 flex-1 text-sm custom-scrollbar">
          {/* Seletor de Parâmetros de Aporte */}
          <div className="p-4 sm:p-5 rounded-2xl bg-neutral-900/60 border border-neutral-800 space-y-4">
            <div>
              <label className="text-xs font-bold text-neutral-300 block mb-2 uppercase tracking-wider">
                1. Quanto você deseja aportar hoje?
              </label>

              <div className="flex flex-col sm:flex-row items-center gap-3">
                <div className="relative w-full sm:w-64">
                  <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-500 font-bold font-mono text-sm">
                    R$
                  </span>
                  <input
                    type="number"
                    min="1"
                    step="50"
                    value={valorAporte}
                    onChange={(e) => setValorAporte(Number(e.target.value))}
                    className="w-full pl-10 pr-4 py-2.5 rounded-xl bg-neutral-950 border border-neutral-700 focus:border-[#d4af37] text-white font-mono font-bold text-base outline-none transition-colors"
                    placeholder="1000"
                  />
                </div>

                <div className="flex items-center gap-1.5 flex-wrap w-full sm:w-auto">
                  {valoresAtalhos.map((val) => (
                    <button
                      key={val}
                      type="button"
                      onClick={() => setValorAporte(val)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-mono font-semibold transition-all ${
                        valorAporte === val
                          ? "bg-[#d4af37] text-black font-bold shadow-md shadow-[#d4af37]/20"
                          : "bg-neutral-950 text-neutral-400 hover:text-white border border-neutral-800"
                      }`}
                    >
                      +{val}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Seletor de Perfil Estratégico */}
            <div>
              <label className="text-xs font-bold text-neutral-300 block mb-2 uppercase tracking-wider">
                2. Qual o foco estratégico deste aporte?
              </label>

              <div className="grid grid-cols-1 sm:grid-cols-4 gap-2">
                {[
                  {
                    id: "EQUILIBRADO",
                    nome: "Equilibrado",
                    desc: "Recompor a caixinha mais atrasada da regra 40/22.5/21",
                    icone: <Scale className="w-4 h-4 text-[#d4af37]" />,
                  },
                  {
                    id: "DIVIDENDOS",
                    nome: "Foco Dividendos",
                    desc: "Priorizar FIIs e Ações com alto DY abaixo do Teto Bazin",
                    icone: <Coins className="w-4 h-4 text-emerald-400" />,
                  },
                  {
                    id: "OPORTUNIDADE",
                    nome: "Assimetrias",
                    desc: "Aproveitar descontos extremos e maior margem de segurança",
                    icone: <Zap className="w-4 h-4 text-amber-400" />,
                  },
                  {
                    id: "DEFENSIVO",
                    nome: "Defensivo",
                    desc: "Segurança em Renda Fixa e Blue Chips blindadas",
                    icone: <ShieldCheck className="w-4 h-4 text-cyan-400" />,
                  },
                ].map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => setPerfil(item.id as PerfilAdvisor)}
                    className={`p-3 rounded-xl border text-left transition-all flex flex-col justify-between gap-1.5 ${
                      perfil === item.id
                        ? "bg-[#d4af37]/10 border-[#d4af37] shadow-lg shadow-[#d4af37]/10"
                        : "bg-neutral-950/80 border-neutral-800/80 hover:border-neutral-700"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-white text-xs flex items-center gap-1.5">
                        {item.icone}
                        {item.nome}
                      </span>
                      {perfil === item.id && (
                        <CheckCircle2 className="w-3.5 h-3.5 text-[#d4af37]" />
                      )}
                    </div>
                    <p className="text-[11px] text-neutral-400 leading-tight">
                      {item.desc}
                    </p>
                  </button>
                ))}
              </div>
            </div>

            {/* Checkbox Opção Novos Ativos */}
            <div className="flex items-center gap-2 pt-1">
              <input
                type="checkbox"
                id="checkNovosAtivos"
                checked={permitirNovos}
                onChange={(e) => setPermitirNovos(e.target.checked)}
                className="w-4 h-4 rounded border-neutral-700 bg-neutral-950 text-[#d4af37] focus:ring-[#d4af37]"
              />
              <label
                htmlFor="checkNovosAtivos"
                className="text-xs text-neutral-300 cursor-pointer select-none"
              >
                Permitir que a IA sugira novos ativos líderes do ranking B3 caso minha classe defasada não possua posições suficientes.
              </label>
            </div>

            {/* Botão de Disparo */}
            <div className="pt-2 flex justify-end">
              <button
                onClick={iniciarConsulta}
                disabled={executando}
                className="w-full sm:w-auto px-6 py-3 rounded-xl bg-gradient-to-r from-[#d4af37] via-[#e5bf43] to-[#c49f2f] hover:from-[#e5bf43] hover:to-[#b8952b] text-black font-extrabold text-xs tracking-wider uppercase shadow-xl shadow-[#d4af37]/20 flex items-center justify-center gap-2 transition-all active:scale-95 disabled:opacity-50"
              >
                {executando ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin text-black" />
                    <span>Auditando Mercado & Carteira com IA...</span>
                  </>
                ) : (
                  <>
                    <Sparkles className="w-4 h-4 text-black" />
                    <span>Consultar Assistente de Aportes</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {erro && (
            <div className="p-4 rounded-xl bg-rose-950/40 border border-rose-500/40 text-rose-200 text-xs flex items-center gap-2.5">
              <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
              <span>{erro}</span>
            </div>
          )}

          {/* Cards das Sugestões Acionáveis Pré-Calculadas */}
          {diagnostico && diagnostico.sugestoes.length > 0 && (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="text-xs font-black uppercase tracking-wider text-white flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                  Alocação Recomendada ({formatarMoeda(diagnostico.valor_aporte_efetivo)})
                </h4>
                <span className="text-[11px] text-neutral-400 font-mono">
                  Sobra em Caixa: <strong className="text-neutral-200">{formatarMoeda(diagnostico.saldo_restante_caixa)}</strong>
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {diagnostico.sugestoes.map((sug, idx) => (
                  <div
                    key={idx}
                    className="p-4 rounded-xl bg-[#0f0f15] border border-neutral-800 hover:border-[#d4af37]/40 transition-all flex flex-col justify-between gap-3 shadow-sm"
                  >
                    <div>
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="font-extrabold text-white text-base font-mono">
                              {sug.ticker}
                            </span>
                            <span className="text-[10px] px-2 py-0.5 rounded bg-neutral-800 text-neutral-300 font-semibold font-mono">
                              {sug.classe}
                            </span>
                          </div>
                          <p className="text-xs text-neutral-400 mt-1 line-clamp-2">
                            {sug.justificativa}
                          </p>
                        </div>

                        <div className="text-right shrink-0">
                          <span className="text-[10px] text-neutral-500 block uppercase font-mono">Total</span>
                          <span className="text-sm font-bold text-emerald-400 font-mono">
                            {formatarMoeda(sug.valor_total_sugerido)}
                          </span>
                        </div>
                      </div>

                      <div className="grid grid-cols-3 gap-2 pt-3 mt-3 border-t border-neutral-800/80 font-mono text-[11px]">
                        <div>
                          <span className="text-[10px] text-neutral-500 block">Qtd</span>
                          <span className="font-semibold text-white">
                            {sug.quantidade_sugerida} cotas
                          </span>
                        </div>
                        <div>
                          <span className="text-[10px] text-neutral-500 block">Preço</span>
                          <span className="font-semibold text-white">
                            {formatarMoeda(sug.preco_estimado)}
                          </span>
                        </div>
                        <div>
                          <span className="text-[10px] text-neutral-500 block">Margem</span>
                          <span className={`font-semibold ${sug.margem_seguranca_pct >= 0 ? "text-emerald-400" : "text-amber-400"}`}>
                            {sug.margem_seguranca_pct > 0 ? `+${sug.margem_seguranca_pct.toFixed(1)}%` : "N/D"}
                          </span>
                        </div>
                      </div>
                    </div>

                    {onAportarEmAtivo && (
                      <button
                        onClick={() => {
                          onAportarEmAtivo(
                            sug.ticker,
                            sug.quantidade_sugerida,
                            sug.preco_estimado,
                            sug.classe
                          );
                          onClose();
                        }}
                        className="w-full py-2 rounded-lg bg-[#d4af37]/15 hover:bg-[#d4af37]/25 text-[#d4af37] border border-[#d4af37]/40 text-xs font-bold tracking-wide flex items-center justify-center gap-1.5 transition-colors"
                      >
                        <Plus className="w-3.5 h-3.5" />
                        <span>Aportar Este Ativo na Carteira</span>
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Área de Streaming do Parecer da IA */}
          {(streamingTexto || executando) && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold uppercase tracking-wider text-neutral-300 flex items-center gap-2">
                  <Lightbulb className="w-4 h-4 text-[#d4af37]" />
                  Parecer do Estrategista Buy & Hold
                </span>
                {executando && (
                  <span className="text-[10px] text-[#d4af37] font-mono flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-[#d4af37] animate-ping"></span>
                    Digitando...
                  </span>
                )}
              </div>

              <div
                ref={scrollRef}
                className="p-4 sm:p-5 rounded-2xl bg-neutral-950 border border-[#d4af37]/30 text-neutral-200 leading-relaxed font-sans text-xs sm:text-sm whitespace-pre-wrap max-h-72 overflow-y-auto custom-scrollbar shadow-inner"
              >
                {streamingTexto}
                {executando && (
                  <span className="inline-block w-2 h-4 ml-1 bg-[#d4af37] animate-pulse align-middle" />
                )}
              </div>
            </div>
          )}
        </div>

        {/* Modal Footer */}
        <div className="p-4 border-t border-neutral-800 bg-[#0a0a0e] flex items-center justify-between text-xs text-neutral-500 font-mono">
          <span>Metodologia: Décio Bazin • Graham • Buy & Hold</span>
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-300 transition-colors font-sans text-xs"
          >
            Fechar
          </button>
        </div>
      </div>
    </div>
  );
}
