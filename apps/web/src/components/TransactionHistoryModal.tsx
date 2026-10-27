"use client";

import React, { useState, useEffect, useMemo } from "react";
import { X, ReceiptText, Calendar, Filter, Loader2, ArrowUpRight, ArrowDownRight, Tag } from "lucide-react";
import { Transacao } from "@/types/portfolio";
import { API_V1_URL } from "@/config/api";
import { useAuth } from "@/context/AuthContext";

interface TransactionHistoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  tickerFiltroInicial?: string | null;
}

export function TransactionHistoryModal({
  isOpen,
  onClose,
  tickerFiltroInicial = null,
}: TransactionHistoryModalProps) {
  const { token } = useAuth();
  const [transacoes, setTransacoes] = useState<Transacao[]>([]);
  const [loading, setLoading] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [filtroTicker, setFiltroTicker] = useState<string>("");
  const [filtroTipo, setFiltroTipo] = useState<string>("TODOS");

  // Atualiza filtro quando o modal abre com ticker inicial
  useEffect(() => {
    if (tickerFiltroInicial) {
      setFiltroTicker(tickerFiltroInicial.toUpperCase());
    } else {
      setFiltroTicker("");
    }
  }, [tickerFiltroInicial, isOpen]);

  // Fecha com ESC
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  // Busca as transações na API
  useEffect(() => {
    if (!isOpen || !token) return;

    async function buscarTransacoes() {
      setLoading(true);
      setErro(null);
      try {
        const res = await fetch(`${API_V1_URL}/carteira/transacoes`, {
          headers: {
            Authorization: `Bearer ${token}`,
            Accept: "application/json",
          },
        });

        if (!res.ok) {
          throw new Error("Não foi possível carregar o extrato de transações.");
        }

        const data: Transacao[] = await res.json();
        setTransacoes(data || []);
      } catch (err: any) {
        setErro(err.message || "Erro de conexão ao carregar transações.");
      } finally {
        setLoading(false);
      }
    }

    buscarTransacoes();
  }, [isOpen, token]);

  // Lista filtrada
  const transacoesFiltradas = useMemo(() => {
    return transacoes.filter((t) => {
      if (filtroTicker.trim() && !t.ticker.toUpperCase().includes(filtroTicker.trim().toUpperCase())) {
        return false;
      }
      if (filtroTipo !== "TODOS" && t.tipo !== filtroTipo) {
        return false;
      }
      return true;
    });
  }, [transacoes, filtroTicker, filtroTipo]);

  // Métricas rápidas
  const totalVolume = useMemo(() => {
    return transacoesFiltradas.reduce((acc, t) => acc + (t.valor_total || 0), 0);
  }, [transacoesFiltradas]);

  if (!isOpen) return null;

  const formatarData = (dtStr: string) => {
    try {
      const dt = new Date(dtStr);
      return new Intl.DateTimeFormat("pt-BR", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      }).format(dt);
    } catch {
      return dtStr;
    }
  };

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat("pt-BR", {
      style: "currency",
      currency: "BRL",
    }).format(val);
  };

  const getBadgeTipo = (tipo: string) => {
    switch (tipo) {
      case "COMPRA":
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold font-mono bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex items-center gap-1 w-fit">
            <ArrowDownRight className="w-3 h-3" /> COMPRA
          </span>
        );
      case "APORTE":
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold font-mono bg-[#d4af37]/20 text-[#d4af37] border border-[#d4af37]/40 flex items-center gap-1 w-fit">
            <ArrowUpRight className="w-3 h-3" /> APORTE (PM)
          </span>
        );
      case "VENDA":
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold font-mono bg-rose-500/15 text-rose-400 border border-rose-500/30 flex items-center gap-1 w-fit">
            <ArrowUpRight className="w-3 h-3" /> VENDA
          </span>
        );
      case "RESGATE":
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold font-mono bg-amber-500/15 text-amber-400 border border-amber-500/30 flex items-center gap-1 w-fit">
            RESGATE
          </span>
        );
      default:
        return (
          <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-neutral-800 text-neutral-300">
            {tipo}
          </span>
        );
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in duration-200">
      <div
        className="w-full max-w-4xl rounded-2xl bg-[#0c0d12] border border-[#d4af37]/30 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header do Modal */}
        <div className="p-5 border-b border-neutral-800 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-[#d4af37]/15 text-[#d4af37] border border-[#d4af37]/30">
              <ReceiptText className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white flex items-center gap-2">
                Extrato de Transações Contábeis • Livro-Razão
                {filtroTicker && (
                  <span className="text-xs px-2 py-0.5 rounded bg-neutral-800 text-[#d4af37] border border-neutral-700 font-mono">
                    {filtroTicker}
                  </span>
                )}
              </h3>
              <p className="text-xs text-neutral-400">
                Histórico imutável de compras, novos aportes e evolução contábil de Preço Médio
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

        {/* Barra de Filtros e Métricas Rápidas */}
        <div className="p-4 bg-[#0a0a0f] border-b border-neutral-800/80 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
          <div className="flex items-center gap-2 flex-wrap">
            <div className="w-36">
              <input
                type="text"
                placeholder="Filtrar Ticker..."
                value={filtroTicker}
                onChange={(e) => setFiltroTicker(e.target.value.toUpperCase())}
                className="w-full px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 text-xs text-white placeholder-neutral-500 uppercase font-mono focus:outline-none focus:border-[#d4af37]"
              />
            </div>

            <div className="flex items-center gap-1 bg-neutral-900/80 p-0.5 rounded-lg border border-neutral-800">
              {(["TODOS", "APORTE", "COMPRA", "VENDA"] as const).map((tipo) => (
                <button
                  key={tipo}
                  onClick={() => setFiltroTipo(tipo)}
                  className={`px-2.5 py-1 rounded text-[11px] font-medium transition-all ${
                    filtroTipo === tipo
                      ? "bg-[#d4af37] text-black font-bold shadow-sm"
                      : "text-neutral-400 hover:text-white"
                  }`}
                >
                  {tipo === "TODOS" ? "Todos" : tipo}
                </button>
              ))}
            </div>

            {(filtroTicker || filtroTipo !== "TODOS") && (
              <button
                onClick={() => {
                  setFiltroTicker("");
                  setFiltroTipo("TODOS");
                }}
                className="text-[11px] text-neutral-400 hover:text-white underline ml-1"
              >
                Limpar filtros
              </button>
            )}
          </div>

          <div className="flex items-center gap-3 font-mono text-[11px] text-neutral-400 self-end sm:self-auto">
            <span>
              Lançamentos: <strong className="text-white">{transacoesFiltradas.length}</strong>
            </span>
            <span>•</span>
            <span>
              Volume Total: <strong className="text-emerald-400">{formatCurrency(totalVolume)}</strong>
            </span>
          </div>
        </div>

        {/* Tabela de Transações */}
        <div className="flex-1 overflow-y-auto p-4 sm:p-6">
          {loading ? (
            <div className="py-16 flex flex-col items-center justify-center gap-2 text-neutral-400 font-mono text-xs">
              <Loader2 className="w-6 h-6 animate-spin text-[#d4af37]" />
              <span>Consultando livro-razão contábil...</span>
            </div>
          ) : erro ? (
            <div className="p-4 rounded-xl bg-red-950/30 border border-red-500/30 text-red-300 text-xs text-center font-mono">
              {erro}
            </div>
          ) : transacoesFiltradas.length === 0 ? (
            <div className="py-16 text-center text-neutral-500 font-mono text-xs space-y-2">
              <ReceiptText className="w-8 h-8 text-neutral-600 mx-auto" />
              <p>Nenhuma transação contábil encontrada para o filtro selecionado.</p>
              {filtroTicker && (
                <p className="text-[11px] text-neutral-400">
                  Experimente limpar o filtro de ticker para ver todo o histórico da carteira.
                </p>
              )}
            </div>
          ) : (
            <div className="rounded-xl border border-neutral-800/80 overflow-hidden bg-neutral-950/40">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-neutral-800 bg-neutral-900/60 text-neutral-400 font-mono uppercase text-[10px]">
                    <th className="py-3 px-3.5">Data / Hora</th>
                    <th className="py-3 px-3.5">Ativo</th>
                    <th className="py-3 px-3.5">Tipo</th>
                    <th className="py-3 px-3.5 text-right">Quantidade</th>
                    <th className="py-3 px-3.5 text-right">Preço Pago</th>
                    <th className="py-3 px-3.5 text-right">Total Operação</th>
                    <th className="py-3 px-3.5">Detalhes / Nota Contábil</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-800/40 font-mono">
                  {transacoesFiltradas.map((t) => (
                    <tr key={t.id} className="hover:bg-neutral-800/20 transition-colors">
                      <td className="py-3 px-3.5 text-neutral-400 whitespace-nowrap text-[11px]">
                        {formatarData(t.data_operacao)}
                      </td>
                      <td className="py-3 px-3.5">
                        <span className="font-extrabold text-white text-xs tracking-wide">
                          {t.ticker}
                        </span>
                        <span className="text-[10px] text-neutral-500 block font-sans">
                          {t.corretora}
                        </span>
                      </td>
                      <td className="py-3 px-3.5 whitespace-nowrap">
                        {getBadgeTipo(t.tipo)}
                      </td>
                      <td className="py-3 px-3.5 text-right text-white font-medium">
                        {t.quantidade.toLocaleString("pt-BR", { maximumFractionDigits: 4 })}
                      </td>
                      <td className="py-3 px-3.5 text-right text-neutral-200">
                        {formatCurrency(t.preco_unitario)}
                      </td>
                      <td className="py-3 px-3.5 text-right font-bold text-white">
                        {formatCurrency(t.valor_total)}
                      </td>
                      <td className="py-3 px-3.5 text-neutral-400 text-[11px] font-sans max-w-xs truncate">
                        {t.notas || "-"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Footer do Modal */}
        <div className="p-4 border-t border-neutral-800/80 bg-neutral-950/60 flex items-center justify-between text-xs">
          <span className="text-neutral-500 font-mono text-[11px]">
            Conformidade Contábil Buy & Hold • Registro Ponderado
          </span>
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-200 font-medium transition-colors"
          >
            Fechar Extrato
          </button>
        </div>
      </div>
    </div>
  );
}
