"use client";

import React, { useState, useEffect } from "react";
import { ResultadoRebalanceamento } from "@/types/portfolio";
import { X, Scale, Sparkles, CheckCircle, AlertTriangle, ArrowRight, Loader2 } from "lucide-react";

interface RebalanceSimulatorModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSimular: (valorAporte: number) => Promise<ResultadoRebalanceamento>;
}

export function RebalanceSimulatorModal({ isOpen, onClose, onSimular }: RebalanceSimulatorModalProps) {
  const [valorInput, setValorInput] = useState<string>("2000");
  const [loading, setLoading] = useState(false);
  const [resultado, setResultado] = useState<ResultadoRebalanceamento | null>(null);
  const [erro, setErro] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const handleSimular = async (valor: number) => {
    if (isNaN(valor) || valor <= 0) {
      setErro("Informe um valor de aporte válido maior que zero.");
      return;
    }

    setErro(null);
    setLoading(true);
    try {
      const res = await onSimular(valor);
      setResultado(res);
    } catch (err: any) {
      setErro(err.message || "Falha ao calcular simulação de rebalanceamento.");
    } finally {
      setLoading(false);
    }
  };

  const handleQuickAdd = (valor: number) => {
    setValorInput(valor.toString());
    handleSimular(valor);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in duration-200">
      <div
        className="w-full max-w-4xl rounded-2xl bg-[#0c0d12] border border-neutral-800 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header do Modal */}
        <div className="p-5 border-b border-neutral-800 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-[#d4af37]/15 text-[#d4af37] border border-[#d4af37]/30">
              <Scale className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white flex items-center gap-2">
                Simulador de Aporte Inteligente • Rebalanceamento Buy & Hold
              </h3>
              <p className="text-xs text-neutral-400">
                Alocação ótima em 2 níveis (Macro Alocação + Valuation) sem forçar vendas ou gerar tributos
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Corpo do Simulador */}
        <div className="p-6 overflow-y-auto space-y-6 flex-1">
          {/* Campo de Entrada de Aporte */}
          <div className="p-5 rounded-2xl bg-neutral-900/60 border border-neutral-800 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <label className="block text-xs font-mono font-bold uppercase text-neutral-400 mb-1">
                Valor Disponível para Aporte Mensal (R$)
              </label>
              <div className="relative">
                <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-400 font-mono text-sm">
                  R$
                </span>
                <input
                  type="number"
                  step="100"
                  value={valorInput}
                  onChange={(e) => setValorInput(e.target.value)}
                  placeholder="Ex: 2000"
                  className="w-full sm:w-64 pl-10 pr-4 py-2.5 rounded-xl bg-neutral-950 border border-neutral-700 text-white font-mono font-bold text-base focus:outline-none focus:border-[#d4af37]"
                />
              </div>
            </div>

            {/* Atalhos Rápidos */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
              <div className="flex items-center gap-1.5">
                {[500, 1000, 2000, 5000].map((v) => (
                  <button
                    key={v}
                    onClick={() => handleQuickAdd(v)}
                    className="px-2.5 py-1.5 rounded-lg bg-neutral-800 hover:bg-neutral-700 text-neutral-300 font-mono text-xs border border-neutral-700 transition-colors"
                  >
                    +R$ {v}
                  </button>
                ))}
              </div>

              <button
                onClick={() => handleSimular(parseFloat(valorInput))}
                disabled={loading}
                className="px-5 py-2.5 rounded-xl bg-[#d4af37] hover:bg-[#c49f2f] text-black font-bold text-xs shadow-lg shadow-[#d4af37]/25 flex items-center gap-2 disabled:opacity-50 transition-all shrink-0"
              >
                {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4" />}
                <span>Calcular Rebalanceamento</span>
              </button>
            </div>
          </div>

          {erro && (
            <div className="p-4 rounded-xl bg-red-950/40 border border-red-500/40 text-red-200 text-xs flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-red-400 shrink-0" />
              <span>{erro}</span>
            </div>
          )}

          {/* Resultado da Simulação */}
          {resultado && (
            <div className="space-y-6 animate-in fade-in-50 duration-300">
              {/* Resumo Rápido */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div className="p-4 rounded-xl bg-neutral-900/60 border border-neutral-800 font-mono">
                  <span className="text-[11px] text-neutral-400 uppercase block">Aporte Informado</span>
                  <span className="text-lg font-bold text-white">
                    {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(
                      resultado.valor_aporte_informado
                    )}
                  </span>
                </div>

                <div className="p-4 rounded-xl bg-emerald-950/20 border border-emerald-500/30 font-mono">
                  <span className="text-[11px] text-emerald-400 uppercase block">Aporte Efetivo Sugerido</span>
                  <span className="text-lg font-bold text-emerald-300">
                    {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(
                      resultado.valor_aporte_efetivo
                    )}
                  </span>
                </div>

                <div className="p-4 rounded-xl bg-neutral-900/60 border border-neutral-800 font-mono">
                  <span className="text-[11px] text-neutral-400 uppercase block">Saldo Restante Caixa</span>
                  <span className="text-lg font-bold text-amber-400">
                    {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(
                      resultado.saldo_restante_caixa
                    )}
                  </span>
                </div>
              </div>

              {/* Tabela de Compras Recomendadas */}
              <div>
                <h4 className="text-sm font-bold text-white mb-3 flex items-center gap-2 font-mono">
                  <CheckCircle className="w-4 h-4 text-emerald-400" />
                  Sugestões de Compras & Aportes Direcionados:
                </h4>

                {resultado.sugestoes && resultado.sugestoes.length > 0 ? (
                  <div className="rounded-xl border border-neutral-800 overflow-hidden bg-neutral-950">
                    <table className="w-full text-left border-collapse text-xs">
                      <thead>
                        <tr className="border-b border-neutral-800 bg-neutral-900/80 text-neutral-400 font-mono uppercase text-[10px]">
                          <th className="py-3 px-4">Ativo Recomendado</th>
                          <th className="py-3 px-4">Classe</th>
                          <th className="py-3 px-4 text-right">Qtd Sugerida</th>
                          <th className="py-3 px-4 text-right">Preço Est.</th>
                          <th className="py-3 px-4 text-right">Total Alocado</th>
                          <th className="py-3 px-4">Justificativa Financeira</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-neutral-800/60 font-mono">
                        {resultado.sugestoes.map((s, idx) => (
                          <tr key={idx} className="hover:bg-neutral-900/40">
                            <td className="py-3 px-4 font-extrabold text-white text-sm">
                              {s.ticker}
                            </td>
                            <td className="py-3 px-4 text-[11px] text-neutral-300">
                              {s.classe}
                            </td>
                            <td className="py-3 px-4 text-right font-bold text-emerald-400">
                              {s.quantidade_sugerida.toLocaleString("pt-BR", { maximumFractionDigits: 2 })}
                            </td>
                            <td className="py-3 px-4 text-right text-neutral-300">
                              {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(s.preco_estimado)}
                            </td>
                            <td className="py-3 px-4 text-right font-extrabold text-[#d4af37]">
                              {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(s.valor_total_sugerido)}
                            </td>
                            <td className="py-3 px-4 text-[11px] font-sans text-neutral-400 max-w-xs">
                              {s.justificativa}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <div className="p-6 rounded-xl bg-neutral-900/30 border border-neutral-800 text-center text-xs text-neutral-400">
                    Nenhuma compra necessária no momento. A carteira já está equilibrada com as posições atuais.
                  </div>
                )}
              </div>

              {/* Diagnóstico das 5 Macro Classes */}
              <div>
                <h4 className="text-sm font-bold text-white mb-3 flex items-center gap-2 font-mono">
                  <ArrowRight className="w-4 h-4 text-cyan-400" />
                  Diagnóstico de Equilíbrio das Macro Classes:
                </h4>

                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                  {resultado.diagnostico_macro.map((d) => (
                    <div
                      key={d.classe}
                      className="p-3.5 rounded-xl bg-neutral-900/50 border border-neutral-800 flex flex-col justify-between"
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-bold text-xs text-white">{d.classe}</span>
                        <span
                          className={`text-[10px] font-mono font-semibold px-2 py-0.5 rounded ${
                            d.status === "DEFASADO_APORTAR"
                              ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/30"
                              : d.status === "SOBREALOCADO"
                              ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                              : "bg-neutral-800 text-neutral-300"
                          }`}
                        >
                          {d.status}
                        </span>
                      </div>

                      <div className="text-[11px] font-mono space-y-1 text-neutral-400">
                        <div className="flex justify-between">
                          <span>Atual vs Alvo:</span>
                          <span className="text-white font-semibold">
                            {d.percentual_atual.toFixed(1)}% / {d.percentual_alvo.toFixed(1)}%
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span>Desvio:</span>
                          <span className={d.delta_percentual >= 0 ? "text-amber-400" : "text-cyan-400"}>
                            {d.delta_percentual >= 0 ? "+" : ""}
                            {d.delta_percentual.toFixed(1)}%
                          </span>
                        </div>
                        <div className="flex justify-between border-t border-neutral-800/80 pt-1">
                          <span>Necessidade:</span>
                          <span className="text-emerald-400 font-bold">
                            {new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(
                              d.necessidade_aporte
                            )}
                          </span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Rodapé do Modal */}
        <div className="p-4 border-t border-neutral-800 flex items-center justify-end">
          <button
            onClick={onClose}
            className="px-5 py-2 rounded-xl bg-neutral-800 hover:bg-neutral-700 text-xs font-semibold text-white transition-colors"
          >
            Fechar
          </button>
        </div>
      </div>
    </div>
  );
}
