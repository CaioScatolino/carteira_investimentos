"use client";

import React, { useState, useEffect } from "react";
import { ClasseAtivo, CriarPosicaoRequest } from "@/types/portfolio";
import { X, Plus, AlertCircle, Loader2 } from "lucide-react";

interface AddPositionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSalvar: (req: CriarPosicaoRequest) => Promise<void>;
}

export function AddPositionModal({ isOpen, onClose, onSalvar }: AddPositionModalProps) {
  const [ticker, setTicker] = useState("");
  const [classe, setClasse] = useState<ClasseAtivo>("ACAO_BR");
  const [categoriaMacro, setCategoriaMacro] = useState("");
  const [quantidade, setQuantidade] = useState("");
  const [precoMedio, setPrecoMedio] = useState("");
  const [corretora, setCorretora] = useState("BTG Pactual");
  const [taxaContratada, setTaxaContratada] = useState("");
  const [indexador, setIndexador] = useState("CDI");
  const [vencimento, setVencimento] = useState("");
  const [observacoes, setObservacoes] = useState("");

  const [loading, setLoading] = useState(false);
  const [erro, setErro] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErro(null);

    const qtdNum = parseFloat(quantidade.replace(",", "."));
    const pmNum = parseFloat(precoMedio.replace(",", "."));

    if (!ticker.trim()) {
      setErro("Informe o código/ticker do ativo.");
      return;
    }
    if (isNaN(qtdNum) || qtdNum <= 0) {
      setErro("A quantidade deve ser um número maior que zero.");
      return;
    }
    if (isNaN(pmNum) || pmNum <= 0) {
      setErro("O preço médio deve ser maior que zero.");
      return;
    }

    setLoading(true);
    try {
      await onSalvar({
        ticker: ticker.trim().toUpperCase(),
        classe,
        categoria_macro: categoriaMacro.trim() || "GERAL",
        quantidade: qtdNum,
        preco_medio: pmNum,
        corretora: corretora.trim() || "Principal",
        taxa_contratada: classe === "RENDA_FIXA" && taxaContratada.trim() ? taxaContratada.trim() : undefined,
        indexador: classe === "RENDA_FIXA" && indexador.trim() ? indexador.trim() : undefined,
        vencimento: classe === "RENDA_FIXA" && vencimento.trim() ? vencimento.trim() : undefined,
        observacoes: observacoes.trim() || undefined,
      });

      // Limpa os campos
      setTicker("");
      setQuantidade("");
      setPrecoMedio("");
      setTaxaContratada("");
      setObservacoes("");
      onClose();
    } catch (err: any) {
      setErro(err.message || "Falha ao cadastrar ativo na carteira.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
      <div
        className="w-full max-w-lg rounded-2xl bg-[#0e0f14] border border-neutral-800 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header do Modal */}
        <div className="p-5 border-b border-neutral-800/80 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/30">
              <Plus className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-white">Cadastrar Novo Ativo na Carteira</h3>
              <p className="text-xs text-neutral-400">Suporte a Renda Variável, FIIs e Renda Fixa</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Formulário */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4 overflow-y-auto flex-1">
          {erro && (
            <div className="p-3.5 rounded-xl bg-red-950/40 border border-red-500/40 text-red-200 text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />
              <span>{erro}</span>
            </div>
          )}

          {/* Linha 1: Ticker & Classe */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Ticker / Código *
              </label>
              <input
                type="text"
                required
                placeholder="Ex: PETR4, HGLG11"
                value={ticker}
                onChange={(e) => setTicker(e.target.value.toUpperCase())}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm font-mono text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80"
              />
            </div>

            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Macro Classe *
              </label>
              <select
                value={classe}
                onChange={(e) => setClasse(e.target.value as ClasseAtivo)}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm font-mono text-white focus:outline-none focus:border-[#d4af37]/80"
              >
                <option value="ACAO_BR">Ações Brasileiras (Meta 21%)</option>
                <option value="FII">Fundos Imobiliários (Meta 22.5%)</option>
                <option value="RENDA_FIXA">Renda Fixa / Tesouro (Meta 40%)</option>
                <option value="INTERNACIONAL">Internacional / ETFs (Meta 16%)</option>
                <option value="CAIXA">Caixa / Reserva (Meta 0.5%)</option>
              </select>
            </div>
          </div>

          {/* Linha 2: Quantidade & Preço Médio */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Quantidade *
              </label>
              <input
                type="number"
                step="any"
                required
                placeholder="Ex: 100 ou 0.5"
                value={quantidade}
                onChange={(e) => setQuantidade(e.target.value)}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm font-mono text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80"
              />
            </div>

            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Preço Médio (R$) *
              </label>
              <input
                type="number"
                step="0.0001"
                required
                placeholder="Ex: 35.80"
                value={precoMedio}
                onChange={(e) => setPrecoMedio(e.target.value)}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm font-mono text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80"
              />
            </div>
          </div>

          {/* Linha 3: Corretora & Categoria */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Corretora / Custodiante
              </label>
              <input
                type="text"
                placeholder="Ex: BTG, XP, NuInvest"
                value={corretora}
                onChange={(e) => setCorretora(e.target.value)}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80"
              />
            </div>

            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Sub-Segmento / Categoria
              </label>
              <input
                type="text"
                placeholder="Ex: Bancos, Logística, Tesouro"
                value={categoriaMacro}
                onChange={(e) => setCategoriaMacro(e.target.value)}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80"
              />
            </div>
          </div>

          {/* Campos Exclusivos para RENDA FIXA */}
          {classe === "RENDA_FIXA" && (
            <div className="p-4 rounded-xl bg-blue-950/20 border border-blue-500/30 space-y-3">
              <span className="text-xs font-bold text-blue-300 font-mono uppercase block">
                Detalhes de Renda Fixa
              </span>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div>
                  <label className="block text-[11px] font-mono text-neutral-400 mb-1">Taxa Contratada</label>
                  <input
                    type="text"
                    placeholder="Ex: 115% CDI"
                    value={taxaContratada}
                    onChange={(e) => setTaxaContratada(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-neutral-950 border border-neutral-800 text-xs text-white"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-mono text-neutral-400 mb-1">Indexador</label>
                  <select
                    value={indexador}
                    onChange={(e) => setIndexador(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-neutral-950 border border-neutral-800 text-xs text-white"
                  >
                    <option value="CDI">CDI</option>
                    <option value="IPCA">IPCA</option>
                    <option value="PRE">Prefixado</option>
                    <option value="SELIC">Selic</option>
                  </select>
                </div>

                <div>
                  <label className="block text-[11px] font-mono text-neutral-400 mb-1">Vencimento</label>
                  <input
                    type="date"
                    value={vencimento}
                    onChange={(e) => setVencimento(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg bg-neutral-950 border border-neutral-800 text-xs text-white"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Observações */}
          <div>
            <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
              Notas / Tese de Investimento (Opcional)
            </label>
            <textarea
              rows={2}
              placeholder="Ex: Compra focada em dividendos defensivos para rebalanceamento."
              value={observacoes}
              onChange={(e) => setObservacoes(e.target.value)}
              className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-xs text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80 resize-none"
            />
          </div>

          {/* Botões de Ação */}
          <div className="pt-2 flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2.5 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-xs font-semibold text-neutral-400 hover:text-white transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-5 py-2.5 rounded-xl bg-[#d4af37] hover:bg-[#c49f2f] text-xs font-bold text-black transition-all shadow-lg shadow-[#d4af37]/20 flex items-center gap-2 disabled:opacity-50"
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              <span>Cadastrar Ativo</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
