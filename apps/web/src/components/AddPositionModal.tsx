"use client";

import React, { useState, useEffect, useMemo, useRef } from "react";
import { ClasseAtivo, CriarPosicaoRequest, PosicaoConsolidada } from "@/types/portfolio";
import { Ativo } from "@/types/market";
import { X, Plus, AlertCircle, Loader2, Search, Check, Sparkles, Building2, TrendingUp, ArrowDown, ArrowUp, RefreshCw, Calculator } from "lucide-react";

interface AddPositionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSalvar: (req: CriarPosicaoRequest) => Promise<void>;
  ativosDisponiveis?: Ativo[];
  posicaoParaAporte?: PosicaoConsolidada | null;
  posicoesAtuais?: PosicaoConsolidada[];
}

export function AddPositionModal({
  isOpen,
  onClose,
  onSalvar,
  ativosDisponiveis = [],
  posicaoParaAporte = null,
  posicoesAtuais = [],
}: AddPositionModalProps) {
  const [ticker, setTicker] = useState("");
  const [classe, setClasse] = useState<ClasseAtivo>("ACAO_BR");
  const [categoriaMacro, setCategoriaMacro] = useState("");
  const [quantidade, setQuantidade] = useState("");
  const [precoMedio, setPrecoMedio] = useState("");
  const [modoOperacao, setModoOperacao] = useState<"APORTE" | "SUBSTITUIR">("APORTE");

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

  // Inicializa quando abre o modal ou quando posição para aporte é passada
  useEffect(() => {
    if (posicaoParaAporte) {
      setTicker(posicaoParaAporte.ticker);
      setClasse(posicaoParaAporte.classe);
      setCategoriaMacro(posicaoParaAporte.categoria_macro || "");
      setQuantidade("");
      const cotacao = posicaoParaAporte.cotacao_atual > 0
        ? posicaoParaAporte.cotacao_atual
        : posicaoParaAporte.preco_medio;
      setPrecoMedio(cotacao ? cotacao.toFixed(2) : "");
      setModoOperacao("APORTE");
      setTaxaContratada(posicaoParaAporte.taxa_contratada || "");
      setIndexador(posicaoParaAporte.indexador || "CDI");
    } else {
      setTicker("");
      setQuantidade("");
      setPrecoMedio("");
      setCategoriaMacro("");
      setAtivoEncontrado(null);
      setTaxaContratada("");
      setModoOperacao("APORTE");
    }
    setErro(null);
  }, [posicaoParaAporte, isOpen]);

  // Detecta se o ticker digitado já existe na carteira
  const posicaoExistente = useMemo(() => {
    if (posicaoParaAporte) return posicaoParaAporte;
    const tkUpper = ticker.trim().toUpperCase();
    if (!tkUpper || !posicoesAtuais.length) return null;
    return posicoesAtuais.find((p) => p.ticker.toUpperCase() === tkUpper) || null;
  }, [posicaoParaAporte, ticker, posicoesAtuais]);

  const isModoAporteAtivo = !!posicaoExistente && modoOperacao === "APORTE";

  // Simulação Contábil em Tempo Real (Média Ponderada)
  const simulacaoAporte = useMemo(() => {
    if (!posicaoExistente) return null;
    const qtdAporte = parseFloat(quantidade.replace(",", "."));
    const precoAporte = parseFloat(precoMedio.replace(",", "."));
    if (isNaN(qtdAporte) || qtdAporte <= 0 || isNaN(precoAporte) || precoAporte <= 0) {
      return null;
    }

    const qtdAtual = posicaoExistente.quantidade;
    const pmAtual = posicaoExistente.preco_medio;
    const valAtual = posicaoExistente.valor_aplicado > 0 ? posicaoExistente.valor_aplicado : qtdAtual * pmAtual;

    const valAporte = qtdAporte * precoAporte;
    const qtdTotal = qtdAtual + qtdAporte;
    const novoPrecoMedio = (valAtual + valAporte) / (qtdTotal || 1);
    const novoValorTotal = valAtual + valAporte;
    const diferencaPM = novoPrecoMedio - pmAtual;

    return {
      qtdAtual,
      pmAtual,
      valAtual,
      qtdAporte,
      precoAporte,
      valAporte,
      qtdTotal,
      novoPrecoMedio,
      novoValorTotal,
      diferencaPM,
    };
  }, [posicaoExistente, quantidade, precoMedio]);

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
    if (!termo || !ativosDisponiveis.length || posicaoParaAporte) return [];

    return ativosDisponiveis
      .filter((a) => {
        const matchTicker = a.ticker.toUpperCase().includes(termo);
        const matchNome = a.nome.toUpperCase().includes(termo);
        return matchTicker || matchNome;
      })
      .sort((a, b) => {
        const aStart = a.ticker.toUpperCase().startsWith(termo);
        const bStart = b.ticker.toUpperCase().startsWith(termo);
        if (aStart && !bStart) return -1;
        if (!aStart && bStart) return 1;
        return a.ticker.localeCompare(b.ticker);
      })
      .slice(0, 8);
  }, [ticker, ativosDisponiveis, posicaoParaAporte]);

  // Aplica os dados do ativo selecionado
  const selecionarAtivo = (a: Ativo) => {
    setTicker(a.ticker.toUpperCase());
    setAtivoEncontrado(a);
    setMenuAberto(false);

    if (a.classe === "FII") {
      setClasse("FII");
    } else if (a.classe === "ETF") {
      setClasse("INTERNACIONAL");
    } else {
      setClasse("ACAO_BR");
    }

    const seg = a.segmento || a.setor || (a.classe === "FII" ? "Fundos Imobiliários" : "Ações B3");
    setCategoriaMacro(seg);

    if (!precoMedio && a.preco_atual > 0) {
      setPrecoMedio(a.preco_atual.toFixed(2));
    }
  };

  // Trata digitação manual do ticker
  const handleTickerChange = (valor: string) => {
    const tkUpper = valor.toUpperCase();
    setTicker(tkUpper);
    setMenuAberto(true);

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
        corretora: "Principal",
        taxa_contratada: classe === "RENDA_FIXA" && taxaContratada.trim() ? taxaContratada.trim() : undefined,
        indexador: classe === "RENDA_FIXA" && indexador.trim() ? indexador.trim() : undefined,
        vencimento: classe === "RENDA_FIXA" && vencimento.trim() ? vencimento.trim() : undefined,
        tipo_operacao: posicaoExistente ? modoOperacao : "COMPRA",
      });

      setTicker("");
      setQuantidade("");
      setPrecoMedio("");
      setCategoriaMacro("");
      setAtivoEncontrado(null);
      setTaxaContratada("");
      onClose();
    } catch (err: any) {
      setErro(err.message || "Falha ao processar operação na carteira.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
      <div
        className="w-full max-w-lg rounded-2xl bg-[#0e0f14] border border-[#d4af37]/30 shadow-2xl overflow-hidden flex flex-col max-h-[92vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header do Modal */}
        <div className="p-5 border-b border-neutral-800/80 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className={`p-2 rounded-xl border ${
              isModoAporteAtivo
                ? "bg-emerald-500/15 text-emerald-400 border-emerald-500/30"
                : "bg-[#d4af37]/10 text-[#d4af37] border-[#d4af37]/30"
            }`}>
              {isModoAporteAtivo ? <TrendingUp className="w-5 h-5" /> : <Plus className="w-5 h-5" />}
            </div>
            <div>
              <h3 className="text-base font-bold text-white flex items-center gap-2">
                {isModoAporteAtivo ? `Registrar Aporte • ${ticker}` : "Cadastrar Novo Ativo na Carteira"}
                {isModoAporteAtivo && (
                  <span className="text-[10px] font-mono uppercase px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                    Média Ponderada
                  </span>
                )}
              </h3>
              <p className="text-xs text-neutral-400">
                {isModoAporteAtivo
                  ? "Soma automática de cotas e recálculo contábil do Preço Médio"
                  : "Preenchimento automático inteligente integrado à B3"}
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

        {/* Formulário */}
        <form onSubmit={handleSubmit} className="p-5 sm:p-6 space-y-4 overflow-y-auto flex-1">
          {erro && (
            <div className="p-3.5 rounded-xl bg-red-950/40 border border-red-500/40 text-red-200 text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />
              <span>{erro}</span>
            </div>
          )}

          {/* Banner de Posição Existente Detectada */}
          {posicaoExistente && (
            <div className="p-3.5 rounded-xl bg-gradient-to-r from-[#141824] to-[#11131c] border border-cyan-500/30 text-xs space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1.5 text-cyan-300 font-bold font-mono">
                  <Calculator className="w-4 h-4" />
                  <span>Posição Atual em Custódia</span>
                </div>
                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={() => setModoOperacao("APORTE")}
                    className={`px-2 py-0.5 rounded text-[10px] font-bold font-mono transition-colors ${
                      modoOperacao === "APORTE"
                        ? "bg-[#d4af37] text-black"
                        : "bg-neutral-800 text-neutral-400 hover:text-white"
                    }`}
                  >
                    Somar Aporte
                  </button>
                  <button
                    type="button"
                    onClick={() => setModoOperacao("SUBSTITUIR")}
                    className={`px-2 py-0.5 rounded text-[10px] font-bold font-mono transition-colors ${
                      modoOperacao === "SUBSTITUIR"
                        ? "bg-[#d4af37] text-black"
                        : "bg-neutral-800 text-neutral-400 hover:text-white"
                    }`}
                    title="Substituir dados sem somar média ponderada (correção de erro)"
                  >
                    Substituir
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-2 pt-1 font-mono text-[11px] text-neutral-300 border-t border-neutral-800">
                <div>
                  <span className="text-neutral-500 block text-[10px]">Cotas Atuais</span>
                  <strong className="text-white">{posicaoExistente.quantidade.toLocaleString("pt-BR")}</strong>
                </div>
                <div>
                  <span className="text-neutral-500 block text-[10px]">Preço Médio Atual</span>
                  <strong className="text-[#d4af37]">R$ {posicaoExistente.preco_medio.toFixed(2)}</strong>
                </div>
                <div>
                  <span className="text-neutral-500 block text-[10px]">Total Investido</span>
                  <strong className="text-white">
                    R$ {(posicaoExistente.valor_aplicado || posicaoExistente.quantidade * posicaoExistente.preco_medio).toFixed(2)}
                  </strong>
                </div>
              </div>
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
                  disabled={!!posicaoParaAporte}
                  placeholder="Ex: UNIP6, PETR4, HGLG11"
                  value={ticker}
                  onChange={(e) => handleTickerChange(e.target.value)}
                  onFocus={() => !posicaoParaAporte && setMenuAberto(true)}
                  className={`w-full pl-9 pr-3.5 py-2.5 rounded-xl border text-sm font-mono text-white placeholder-neutral-600 focus:outline-none uppercase ${
                    posicaoParaAporte
                      ? "bg-neutral-900 border-neutral-700 cursor-not-allowed opacity-90"
                      : "bg-neutral-950 border-neutral-800 focus:border-[#d4af37]/80"
                  }`}
                />
                <Search className="w-4 h-4 text-neutral-500 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
              </div>

              {/* Menu Suspenso de Sugestões */}
              {menuAberto && sugestoes.length > 0 && !posicaoParaAporte && (
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
                          <span className="text-[10px] uppercase px-1.5 py-0.2 rounded bg-neutral-800 text-neutral-300 font-mono">
                            {item.classe}
                          </span>
                        </div>
                        <p className="text-[11px] text-neutral-400 truncate max-w-[200px]">
                          {item.nome}
                        </p>
                      </div>
                      <div className="text-right font-mono">
                        <span className="text-xs text-emerald-400 font-semibold block">
                          R$ {item.preco_atual.toFixed(2)}
                        </span>
                        <span className="text-[10px] text-neutral-500">
                          DY: {item.dy.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Macro Classe */}
            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                Classe do Ativo
              </label>
              <select
                value={classe}
                onChange={(e) => setClasse(e.target.value as ClasseAtivo)}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm font-mono text-white focus:outline-none focus:border-[#d4af37]/80"
              >
                <option value="ACAO_BR">Ação Brasileira (B3)</option>
                <option value="FII">Fundo Imobiliário (FII)</option>
                <option value="RENDA_FIXA">Renda Fixa</option>
                <option value="INTERNACIONAL">Internacional (ETF / BDR)</option>
                <option value="CAIXA">Caixa / Oportunidade</option>
              </select>
            </div>
          </div>

          {/* Card de Ativo Detectado da B3 com botão Cotação */}
          {ativoEncontrado && (
            <div className="p-3 rounded-xl bg-neutral-950 border border-[#d4af37]/30 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2.5">
                <div className="p-1.5 rounded-lg bg-[#d4af37]/10 text-[#d4af37]">
                  <Building2 className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-white">
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

          {/* Linha 2: Quantidade & Preço Pago */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                {isModoAporteAtivo ? "Quantidade Comprada *" : "Quantidade *"}
              </label>
              <input
                type="number"
                step="any"
                required
                placeholder={isModoAporteAtivo ? "Ex: 10 ou 100" : "Ex: 100 ou 0.5"}
                value={quantidade}
                onChange={(e) => setQuantidade(e.target.value)}
                className="w-full px-3.5 py-2.5 rounded-xl bg-neutral-950 border border-neutral-800 text-sm font-mono text-white placeholder-neutral-600 focus:outline-none focus:border-[#d4af37]/80"
              />
            </div>

            <div>
              <label className="block text-xs font-mono font-semibold uppercase text-neutral-400 mb-1">
                {isModoAporteAtivo ? "Preço Pago na Compra (R$) *" : "Preço Médio (R$) *"}
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

          {/* Card de Simulação Contábil Ponderada em Tempo Real */}
          {isModoAporteAtivo && simulacaoAporte && (
            <div className="p-4 rounded-xl bg-gradient-to-r from-[#0d171e] via-[#0e1818] to-[#121c17] border border-emerald-500/40 space-y-2.5 animate-in fade-in duration-150">
              <div className="flex items-center justify-between text-xs border-b border-emerald-500/20 pb-2">
                <span className="font-bold text-emerald-400 flex items-center gap-1.5 font-mono">
                  <Calculator className="w-4 h-4" />
                  Prévia Contábil do Aporte
                </span>
                <span className="text-[10px] text-neutral-400 font-mono">
                  Média Ponderada Instantânea
                </span>
              </div>

              <div className="grid grid-cols-3 gap-2 font-mono text-xs">
                <div>
                  <span className="text-[10px] text-neutral-400 block">Nova Qtd Total</span>
                  <span className="font-bold text-white text-sm">
                    {simulacaoAporte.qtdTotal.toLocaleString("pt-BR")}
                  </span>
                  <span className="text-[10px] text-emerald-400 block">
                    (+{simulacaoAporte.qtdAporte})
                  </span>
                </div>

                <div>
                  <span className="text-[10px] text-neutral-400 block">Novo Preço Médio</span>
                  <span className="font-bold text-[#d4af37] text-sm">
                    R$ {simulacaoAporte.novoPrecoMedio.toFixed(2)}
                  </span>
                  <span className={`text-[10px] flex items-center gap-0.5 ${
                    simulacaoAporte.diferencaPM < 0 ? "text-emerald-400" : "text-amber-400"
                  }`}>
                    {simulacaoAporte.diferencaPM < 0 ? (
                      <>
                        <ArrowDown className="w-3 h-3" />
                        -R$ {Math.abs(simulacaoAporte.diferencaPM).toFixed(2)}
                      </>
                    ) : (
                      <>
                        <ArrowUp className="w-3 h-3" />
                        +R$ {simulacaoAporte.diferencaPM.toFixed(2)}
                      </>
                    )}
                  </span>
                </div>

                <div>
                  <span className="text-[10px] text-neutral-400 block">Novo Total Investido</span>
                  <span className="font-bold text-white text-sm">
                    R$ {simulacaoAporte.novoValorTotal.toFixed(2)}
                  </span>
                  <span className="text-[10px] text-neutral-400 block">
                    (+R$ {simulacaoAporte.valAporte.toFixed(2)})
                  </span>
                </div>
              </div>
            </div>
          )}

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
              className={`px-5 py-2.5 rounded-xl text-xs font-bold transition-all shadow-lg flex items-center gap-2 disabled:opacity-50 ${
                isModoAporteAtivo
                  ? "bg-emerald-500 hover:bg-emerald-400 text-black shadow-emerald-500/20"
                  : "bg-[#d4af37] hover:bg-[#c49f2f] text-black shadow-[#d4af37]/20"
              }`}
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              <span>
                {isModoAporteAtivo
                  ? "Confirmar Aporte e Atualizar PM"
                  : "Cadastrar Ativo"}
              </span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
