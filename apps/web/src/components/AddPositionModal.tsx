"use client";

import React, { useState, useEffect, useMemo, useRef } from "react";
import { ClasseAtivo, CriarPosicaoRequest } from "@/types/portfolio";
import { Ativo } from "@/types/market";
import { X, Plus, AlertCircle, Loader2, Search, Check, Sparkles, Building2 } from "lucide-react";

interface AddPositionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSalvar: (req: CriarPosicaoRequest) => Promise<void>;
  ativosDisponiveis?: Ativo[];
}

export function AddPositionModal({
  isOpen,
  onClose,
  onSalvar,
  ativosDisponiveis = [],
}: AddPositionModalProps) {
  const [ticker, setTicker] = useState("");
  const [classe, setClasse] = useState<ClasseAtivo>("ACAO_BR");
  const [categoriaMacro, setCategoriaMacro] = useState("");
  const [quantidade, setQuantidade] = useState("");
  const [precoMedio, setPrecoMedio] = useState("");

  // Renda Fixa (campos específicos)
  const [taxaContratada, setTaxaContratada] = useState("");
  const [indexador, setIndexador] = useState("CDI");
  const [vencimento, setVencimento] = useState("");

  // Estado do autocomplete
  const [menuAberto, setMenuAberto] = useState(false);
  const [ativoEncontrado, setAtivoEncontrado] = useState<Ativo | null>(null);

  const [loading, setLoading] = useState(false);
  const [erro, setErro] = useState<string | null>(null);

  const searchContainerRef = useRef<HTMLDivElement | null>(null);

  // Fecha o dropdown se clicar fora
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        searchContainerRef.current &&
        !searchContainerRef.current.contains(e.target as Node)
      ) {
        setMenuAberto(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Fecha com ESC
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  // Lista filtrada de sugestões conforme o usuário digita
  const sugestoes = useMemo(() => {
    const termo = ticker.trim().toUpperCase();
    if (!termo || !ativosDisponiveis.length) return [];

    return ativosDisponiveis
      .filter((a) => {
        const matchTicker = a.ticker.toUpperCase().includes(termo);
        const matchNome = a.nome.toUpperCase().includes(termo);
        return matchTicker || matchNome;
      })
      .sort((a, b) => {
        // Prioriza quem começa exatamente com o termo digitado
        const aStart = a.ticker.toUpperCase().startsWith(termo);
        const bStart = b.ticker.toUpperCase().startsWith(termo);
        if (aStart && !bStart) return -1;
        if (!aStart && bStart) return 1;
        return a.ticker.localeCompare(b.ticker);
      })
      .slice(0, 8);
  }, [ticker, ativosDisponiveis]);

  // Aplica os dados do ativo selecionado
  const selecionarAtivo = (a: Ativo) => {
    setTicker(a.ticker.toUpperCase());
    setAtivoEncontrado(a);
    setMenuAberto(false);

    // 1. Auto-seleciona a Macro Classe
    if (a.classe === "FII") {
      setClasse("FII");
    } else if (a.classe === "ETF") {
      setClasse("INTERNACIONAL");
    } else {
      setClasse("ACAO_BR");
    }

    // 2. Auto-preenche o Sub-segmento / Categoria
    const seg = a.segmento || a.setor || (a.classe === "FII" ? "Fundos Imobiliários" : "Ações B3");
    setCategoriaMacro(seg);

    // 3. Sugere o preço atual se o preço médio ainda não foi informado
    if (!precoMedio && a.preco_atual > 0) {
      setPrecoMedio(a.preco_atual.toFixed(2));
    }
  };

  // Trata digitação manual do ticker
  const handleTickerChange = (valor: string) => {
    const tkUpper = valor.toUpperCase();
    setTicker(tkUpper);
    setMenuAberto(true);

    // Verifica se coincide exatamente com um ativo do catálogo
    const exato = ativosDisponiveis.find((a) => a.ticker.toUpperCase() === tkUpper.trim());
    if (exato) {
      setAtivoEncontrado(exato);
      if (exato.classe === "FII") setClasse("FII");
      else if (exato.classe === "ETF") setClasse("INTERNACIONAL");
      else setClasse("ACAO_BR");

      const seg = exato.segmento || exato.setor || (exato.classe === "FII" ? "Fundos Imobiliários" : "Ações B3");
      setCategoriaMacro(seg);
    } else {
      setAtivoEncontrado(null);

      // Heurística rápida por sufixo de código B3 se não tiver correspondência exata
      if (tkUpper.endsWith("11") && !tkUpper.startsWith("IVVB") && !tkUpper.startsWith("SPXI")) {
        setClasse("FII");
        setCategoriaMacro("Fundos Imobiliários");
      } else if (tkUpper.endsWith("3") || tkUpper.endsWith("4") || tkUpper.endsWith("5") || tkUpper.endsWith("6")) {
        setClasse("ACAO_BR");
        setCategoriaMacro("Ações B3");
      }
    }
  };

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
        categoria_macro: categoriaMacro.trim() || (classe === "FII" ? "Fundos Imobiliários" : "Ações"),
        quantidade: qtdNum,
        preco_medio: pmNum,
        corretora: "Principal", // Simplificado: não requer digitação manual
        taxa_contratada: classe === "RENDA_FIXA" && taxaContratada.trim() ? taxaContratada.trim() : undefined,
        indexador: classe === "RENDA_FIXA" && indexador.trim() ? indexador.trim() : undefined,
        vencimento: classe === "RENDA_FIXA" && vencimento.trim() ? vencimento.trim() : undefined,
      });

      // Limpa os campos
      setTicker("");
      setQuantidade("");
      setPrecoMedio("");
      setCategoriaMacro("");
      setAtivoEncontrado(null);
      setTaxaContratada("");
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
              <p className="text-xs text-neutral-400">Preenchimento automático inteligente integrado à B3</p>
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

          {/* Linha 1: Ticker com Autocomplete Inteligente & Macro Classe */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/* Campo Ticker com Sugestões Suspensas */}
            <div ref={searchContainerRef} className="relative">
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Ticker / Código *
              </label>
              <div className="relative">
                <input
                  type="text"
                  required
                  placeholder="Ex: UNIP6, PETR4, HGLG11"
                  value={ticker}
                  onChange={(e) => handleTickerChange(e.target.value)}
                  onFocus={() => setMenuAberto(true)}
                  className="w-full pl-9 pr-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm font-mono text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80 uppercase"
                />
                <Search className="w-4 h-4 text-neutral-500 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
              </div>

              {/* Menu Suspenso de Sugestões */}
              {menuAberto && sugestoes.length > 0 && (
                <div className="absolute top-full left-0 right-0 mt-1.5 bg-[#12131a] border border-[#d4af37]/40 rounded-xl shadow-2xl z-50 max-h-60 overflow-y-auto divide-y divide-neutral-800/60 animate-in fade-in zoom-in-95 duration-100">
                  {sugestoes.map((item) => (
                    <div
                      key={item.ticker}
                      onClick={() => selecionarAtivo(item)}
                      className="p-2.5 hover:bg-neutral-800/80 cursor-pointer transition-colors flex items-center justify-between gap-2"
                    >
                      <div>
                        <div className="flex items-center gap-1.5">
                          <span className="font-mono font-bold text-white text-xs">
                            {item.ticker}
                          </span>
                          <span
                            className={`text-[9px] px-1.5 py-0.2 rounded font-bold uppercase ${
                              item.classe === "FII"
                                ? "bg-purple-950 text-purple-300 border border-purple-800/40"
                                : item.classe === "ETF"
                                ? "bg-amber-950 text-amber-300 border border-amber-800/40"
                                : "bg-blue-950 text-blue-300 border border-blue-800/40"
                            }`}
                          >
                            {item.classe}
                          </span>
                        </div>
                        <span className="text-[11px] text-neutral-400 truncate max-w-[190px] block font-sans">
                          {item.nome}
                        </span>
                      </div>
                      <div className="text-right shrink-0">
                        <span className="text-xs font-mono font-bold text-[#d4af37]">
                          R$ {item.preco_atual.toFixed(2)}
                        </span>
                        {item.segmento && (
                          <span className="text-[10px] text-neutral-500 block">
                            {item.segmento}
                          </span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Macro Classe (Ajustada para nunca cortar texto) */}
            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Macro Classe *
              </label>
              <select
                value={classe}
                onChange={(e) => setClasse(e.target.value as ClasseAtivo)}
                className="w-full px-3 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-xs sm:text-sm text-neutral-200 focus:outline-none focus:border-[#d4af37]/80 cursor-pointer"
              >
                <option value="ACAO_BR">Ações Brasil (Meta 21%)</option>
                <option value="FII">FIIs Imobiliários (Meta 22.5%)</option>
                <option value="RENDA_FIXA">Renda Fixa / Tesouro (Meta 40%)</option>
                <option value="INTERNACIONAL">Internacional / ETFs (Meta 16%)</option>
                <option value="CAIXA">Caixa / Reserva (Meta 0.5%)</option>
              </select>
            </div>
          </div>

          {/* Card com Detalhes do Ativo Identificado Automaticamente */}
          {ativoEncontrado && (
            <div className="p-3 rounded-xl bg-[#14151f] border border-[#d4af37]/30 flex items-center justify-between gap-3 text-xs animate-in fade-in slide-in-from-top-1 duration-150">
              <div className="flex items-center gap-2.5">
                <div className="p-1.5 rounded-lg bg-[#d4af37]/10 text-[#d4af37]">
                  <Building2 className="w-4 h-4" />
                </div>
                <div>
                  <div className="font-bold text-white">
                    {ativoEncontrado.nome}
                  </div>
                  <div className="text-[11px] text-neutral-400 flex items-center gap-2">
                    <span>Segmento: <strong className="text-neutral-300">{categoriaMacro || "Geral"}</strong></span>
                    <span>•</span>
                    <span>Cotação B3: <strong className="text-emerald-400 font-mono">R$ {ativoEncontrado.preco_atual.toFixed(2)}</strong></span>
                  </div>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setPrecoMedio(ativoEncontrado.preco_atual.toFixed(2))}
                className="px-2.5 py-1 rounded-lg bg-[#d4af37]/15 hover:bg-[#d4af37]/25 text-[#d4af37] border border-[#d4af37]/30 text-[11px] font-bold shrink-0 transition-colors"
                title="Copiar cotação atual para Preço Médio"
              >
                Usar Cotação
              </button>
            </div>
          )}

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

          {/* Campos Exclusivos para RENDA FIXA (se selecionado) */}
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

          {/* Botões de Ação */}
          <div className="pt-3 flex items-center justify-end gap-3 border-t border-neutral-800/80">
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
