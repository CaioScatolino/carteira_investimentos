"use client";

import React, { useState, useRef } from "react";
import {
  X,
  Upload,
  FileSpreadsheet,
  CheckCircle2,
  AlertCircle,
  Loader2,
  Calendar,
  Layers,
  DollarSign,
  TrendingUp,
  Coins,
  Check,
  ArrowRight,
  ShieldAlert,
  Info,
} from "lucide-react";
import { API_V1_URL } from "@/config/api";
import { RelatorioAnaliseB3, ResultadoImportacao, ClasseAtivo } from "@/types/portfolio";

interface B3ImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  token: string | null;
  onImportadoSucesso?: () => void;
}

export function B3ImportModal({
  isOpen,
  onClose,
  token,
  onImportadoSucesso,
}: B3ImportModalProps) {
  const [etapa, setEtapa] = useState<"UPLOAD" | "PREVIEW" | "SUCESSO">("UPLOAD");
  const [arquivo, setArquivo] = useState<File | null>(null);
  const [arrastando, setArrastando] = useState(false);
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);

  // Opções de importação
  const [limparAnteriores, setLimparAnteriores] = useState(true);
  const [importarProventos, setImportarProventos] = useState(true);

  // Dados do preview
  const [relatorio, setRelatorio] = useState<RelatorioAnaliseB3 | null>(null);
  const [abaPreview, setAbaPreview] = useState<"POSICOES" | "OPERACOES" | "PROVENTOS">("POSICOES");
  const [resultado, setResultado] = useState<ResultadoImportacao | null>(null);

  const fileInputRef = useRef<HTMLInputElement | null>(null);

  if (!isOpen) return null;

  const resetar = () => {
    setEtapa("UPLOAD");
    setArquivo(null);
    setRelatorio(null);
    setResultado(null);
    setErro(null);
    setCarregando(false);
  };

  const handleFechar = () => {
    resetar();
    onClose();
  };

  const formatarMoeda = (val: number) => {
    return new Intl.NumberFormat("pt-BR", {
      style: "currency",
      currency: "BRL",
    }).format(val);
  };

  const formatarData = (dtStr: string) => {
    if (!dtStr) return "";
    try {
      const dt = new Date(dtStr);
      if (isNaN(dt.getTime())) return dtStr;
      return dt.toLocaleDateString("pt-BR");
    } catch {
      return dtStr;
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      processarArquivo(e.target.files[0]);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setArrastando(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      processarArquivo(e.dataTransfer.files[0]);
    }
  };

  const processarArquivo = async (file: File) => {
    if (!file.name.toLowerCase().endsWith(".xlsx")) {
      setErro("Por favor, selecione um arquivo Excel (.xlsx) oficial da Área do Investidor B3.");
      return;
    }

    setArquivo(file);
    setCarregando(true);
    setErro(null);

    const formData = new FormData();
    formData.append("arquivo", file);

    try {
      const res = await fetch(`${API_V1_URL}/carteira/importar-b3/preview`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.erro || `Falha ao processar arquivo (${res.status})`);
      }

      const data: RelatorioAnaliseB3 = await res.json();
      setRelatorio(data);
      setEtapa("PREVIEW");
    } catch (err: unknown) {
      const mensagem = err instanceof Error ? err.message : "Erro inesperado ao analisar planilha.";
      setErro(mensagem);
      setArquivo(null);
    } finally {
      setCarregando(false);
    }
  };

  const handleConfirmarImportacao = async () => {
    if (!relatorio) return;

    setCarregando(true);
    setErro(null);

    try {
      const res = await fetch(`${API_V1_URL}/carteira/importar-b3/confirmar`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          limpar_dados_anteriores: limparAnteriores,
          importar_proventos: importarProventos,
          posicoes: relatorio.posicoes,
          operacoes: relatorio.operacoes,
          proventos: relatorio.proventos,
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.erro || `Falha ao salvar importação (${res.status})`);
      }

      const data: ResultadoImportacao = await res.json();
      setResultado(data);
      setEtapa("SUCESSO");

      if (onImportadoSucesso) {
        onImportadoSucesso();
      }
    } catch (err: unknown) {
      const mensagem = err instanceof Error ? err.message : "Erro ao persistir carteira.";
      setErro(mensagem);
    } finally {
      setCarregando(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-black/85 backdrop-blur-md animate-in fade-in duration-200">
      <div className="relative w-full max-w-4xl max-h-[92vh] flex flex-col rounded-3xl bg-[#09090d] border border-neutral-800 shadow-2xl shadow-black/80 overflow-hidden font-sans text-white">
        
        {/* Modal Header */}
        <div className="p-5 sm:p-6 border-b border-neutral-800/80 bg-[#0c0c12] flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-gradient-to-br from-[#d4af37]/20 to-amber-500/5 border border-[#d4af37]/40 flex items-center justify-center shadow-lg shadow-[#d4af37]/10">
              <FileSpreadsheet className="w-5 h-5 text-[#d4af37]" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="font-extrabold text-base sm:text-lg text-white font-mono tracking-tight">
                  Importador Oficial B3
                </h3>
                <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 font-semibold tracking-wider uppercase font-mono">
                  Área do Investidor
                </span>
              </div>
              <p className="text-xs text-neutral-400">
                Povoamento instantâneo da custódia, preço médio ponderado e histórico de proventos
              </p>
            </div>
          </div>

          <button
            onClick={handleFechar}
            className="p-2 rounded-xl text-neutral-400 hover:text-white hover:bg-neutral-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="flex-1 overflow-y-auto p-5 sm:p-6 space-y-6 custom-scrollbar">
          
          {erro && (
            <div className="p-4 rounded-2xl bg-rose-950/40 border border-rose-500/40 text-rose-300 text-xs flex items-center gap-3 animate-in fade-in">
              <AlertCircle className="w-5 h-5 text-rose-400 shrink-0" />
              <span>{erro}</span>
            </div>
          )}

          {/* ------------------------------------------------------------- */}
          {/* ETAPA 1: UPLOAD                                               */}
          {/* ------------------------------------------------------------- */}
          {etapa === "UPLOAD" && (
            <div className="space-y-6">
              
              {/* Dropzone */}
              <div
                onDragOver={(e) => {
                  e.preventDefault();
                  setArrastando(true);
                }}
                onDragLeave={() => setArrastando(false)}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                className={`border-2 border-dashed rounded-3xl p-8 sm:p-12 text-center cursor-pointer transition-all duration-300 flex flex-col items-center justify-center gap-4 ${
                  arrastando
                    ? "border-[#d4af37] bg-[#d4af37]/10 scale-[1.01]"
                    : "border-neutral-800 hover:border-[#d4af37]/50 bg-neutral-950/50 hover:bg-neutral-900/50"
                }`}
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".xlsx"
                  className="hidden"
                  onChange={handleFileChange}
                />

                <div className="w-16 h-16 rounded-3xl bg-[#d4af37]/10 border border-[#d4af37]/30 flex items-center justify-center text-[#d4af37] shadow-xl shadow-[#d4af37]/10">
                  {carregando ? (
                    <Loader2 className="w-8 h-8 animate-spin" />
                  ) : (
                    <Upload className="w-8 h-8" />
                  )}
                </div>

                <div className="space-y-1">
                  <p className="font-bold text-sm sm:text-base text-white">
                    {carregando ? "Processando histórico da B3..." : "Arraste sua planilha .xlsx aqui ou clique para selecionar"}
                  </p>
                  <p className="text-xs text-neutral-400 max-w-md mx-auto">
                    Exportada diretamente da Área do Investidor B3 (Extratos → Movimentação ou Negociação).
                  </p>
                </div>

                <button
                  type="button"
                  disabled={carregando}
                  className="px-5 py-2.5 rounded-xl bg-[#d4af37] hover:bg-[#c29e2e] text-black font-extrabold text-xs shadow-lg shadow-[#d4af37]/20 transition-all"
                >
                  {carregando ? "Analisando..." : "Selecionar Arquivo do Computador"}
                </button>
              </div>

              {/* Opções de Importação */}
              <div className="p-4 sm:p-5 rounded-2xl bg-neutral-950 border border-neutral-800/80 space-y-3">
                <span className="text-xs font-bold text-neutral-300 uppercase tracking-wider font-mono block">
                  ⚙️ Configurações da Importação
                </span>

                <label className="flex items-start gap-3 cursor-pointer group">
                  <input
                    type="checkbox"
                    checked={limparAnteriores}
                    onChange={(e) => setLimparAnteriores(e.target.checked)}
                    className="mt-0.5 rounded border-neutral-700 bg-neutral-900 text-[#d4af37] focus:ring-[#d4af37] accent-[#d4af37]"
                  />
                  <div>
                    <span className="text-xs font-bold text-white group-hover:text-[#d4af37] transition-colors">
                      Substituir dados manuais anteriores (Recomendado)
                    </span>
                    <p className="text-[11px] text-neutral-400">
                      Remove ativos e transações cadastrados anteriormente de forma manual para evitar duplicidades, reconstruindo toda a custódia a partir do histórico oficial da B3.
                    </p>
                  </div>
                </label>

                <label className="flex items-start gap-3 cursor-pointer group">
                  <input
                    type="checkbox"
                    checked={importarProventos}
                    onChange={(e) => setImportarProventos(e.target.checked)}
                    className="mt-0.5 rounded border-neutral-700 bg-neutral-900 text-[#d4af37] focus:ring-[#d4af37] accent-[#d4af37]"
                  />
                  <div>
                    <span className="text-xs font-bold text-white group-hover:text-[#d4af37] transition-colors">
                      Importar histórico de proventos e rendimentos recebidos
                    </span>
                    <p className="text-[11px] text-neutral-400">
                      Grava todos os dividendos, JCP e rendimentos recebidos desde 2019 no módulo de renda passiva da plataforma.
                    </p>
                  </div>
                </label>
              </div>

              {/* Guia Rápido */}
              <div className="p-4 rounded-2xl bg-[#0c0c12] border border-neutral-800/60 flex items-start gap-3 text-xs text-neutral-400">
                <Info className="w-5 h-5 text-neutral-400 shrink-0 mt-0.5" />
                <div className="space-y-1">
                  <span className="font-bold text-neutral-300">Como obter o arquivo na B3?</span>
                  <p className="text-[11px] leading-relaxed">
                    Acesse <strong>investidor.b3.com.br</strong> → Menu <strong>Extratos</strong> → <strong>Movimentação</strong> → Selecione todas as instituições e o maior período → Clique em <strong>Exportar (.xlsx)</strong>.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* ------------------------------------------------------------- */}
          {/* ETAPA 2: PREVIEW ANALÍTICO                                    */}
          {/* ------------------------------------------------------------- */}
          {etapa === "PREVIEW" && relatorio && (
            <div className="space-y-5 animate-in fade-in duration-200">
              
              {/* Barra de Síntese */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 font-mono">
                <div className="p-3.5 rounded-2xl bg-neutral-950 border border-neutral-800">
                  <span className="text-[10px] text-neutral-500 uppercase tracking-wider block">Ativos em Custódia</span>
                  <span className="text-lg font-extrabold text-white">{relatorio.total_ativos_ativos} tickers</span>
                </div>
                <div className="p-3.5 rounded-2xl bg-neutral-950 border border-neutral-800">
                  <span className="text-[10px] text-neutral-500 uppercase tracking-wider block">Total Investido (Custo)</span>
                  <span className="text-lg font-extrabold text-[#d4af37]">{formatarMoeda(relatorio.total_investido_custo)}</span>
                </div>
                <div className="p-3.5 rounded-2xl bg-neutral-950 border border-neutral-800">
                  <span className="text-[10px] text-neutral-500 uppercase tracking-wider block">Proventos Históricos</span>
                  <span className="text-lg font-extrabold text-emerald-400">{formatarMoeda(relatorio.total_proventos_historicos)}</span>
                </div>
                <div className="p-3.5 rounded-2xl bg-neutral-950 border border-neutral-800">
                  <span className="text-[10px] text-neutral-500 uppercase tracking-wider block">Movimentações B3</span>
                  <span className="text-lg font-extrabold text-neutral-300">{relatorio.total_movimentacoes} eventos</span>
                </div>
              </div>

              {/* Seletor de Abas */}
              <div className="flex items-center gap-1.5 p-1 rounded-xl bg-neutral-950 border border-neutral-800 text-xs font-mono">
                <button
                  onClick={() => setAbaPreview("POSICOES")}
                  className={`flex-1 py-2 px-3 rounded-lg font-bold transition-all flex items-center justify-center gap-2 ${
                    abaPreview === "POSICOES"
                      ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20"
                      : "text-neutral-400 hover:text-white"
                  }`}
                >
                  <Layers className="w-3.5 h-3.5" />
                  <span>Posições Atuais ({relatorio.posicoes.length})</span>
                </button>
                <button
                  onClick={() => setAbaPreview("OPERACOES")}
                  className={`flex-1 py-2 px-3 rounded-lg font-bold transition-all flex items-center justify-center gap-2 ${
                    abaPreview === "OPERACOES"
                      ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20"
                      : "text-neutral-400 hover:text-white"
                  }`}
                >
                  <TrendingUp className="w-3.5 h-3.5" />
                  <span>Livro-Razão ({relatorio.operacoes.length})</span>
                </button>
                <button
                  onClick={() => setAbaPreview("PROVENTOS")}
                  className={`flex-1 py-2 px-3 rounded-lg font-bold transition-all flex items-center justify-center gap-2 ${
                    abaPreview === "PROVENTOS"
                      ? "bg-[#d4af37] text-black shadow-md shadow-[#d4af37]/20"
                      : "text-neutral-400 hover:text-white"
                  }`}
                >
                  <Coins className="w-3.5 h-3.5" />
                  <span>Proventos ({relatorio.proventos.length})</span>
                </button>
              </div>

              {/* Conteúdo da Aba 1: Posições Finais */}
              {abaPreview === "POSICOES" && (
                <div className="max-h-80 overflow-y-auto custom-scrollbar border border-neutral-800/80 rounded-2xl bg-neutral-950">
                  <table className="w-full text-left text-xs font-mono">
                    <thead className="bg-[#0e0e14] border-b border-neutral-800 text-[10px] text-neutral-400 uppercase tracking-wider sticky top-0">
                      <tr>
                        <th className="p-3">Ticker / Ativo</th>
                        <th className="p-3">Classe</th>
                        <th className="p-3 text-right">Quantidade</th>
                        <th className="p-3 text-right">Preço Médio</th>
                        <th className="p-3 text-right">Custo Total</th>
                        <th className="p-3 text-right text-emerald-400">Proventos</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-900">
                      {relatorio.posicoes.map((p, idx) => (
                        <tr key={idx} className="hover:bg-neutral-900/50 transition-colors">
                          <td className="p-3">
                            <span className="font-extrabold text-white text-sm block">{p.ticker}</span>
                            <span className="text-[10px] text-neutral-500 font-sans truncate block max-w-[200px]">
                              {p.razao_social || p.corretora}
                            </span>
                          </td>
                          <td className="p-3">
                            <span className="text-[10px] px-2 py-0.5 rounded bg-neutral-900 border border-neutral-800 text-neutral-300 font-semibold">
                              {p.classe}
                            </span>
                          </td>
                          <td className="p-3 text-right font-bold text-white">{p.quantidade.toLocaleString("pt-BR", { maximumFractionDigits: 4 })}</td>
                          <td className="p-3 text-right text-neutral-300">{formatarMoeda(p.preco_medio)}</td>
                          <td className="p-3 text-right font-extrabold text-[#d4af37]">{formatarMoeda(p.valor_aplicado)}</td>
                          <td className="p-3 text-right text-emerald-400 font-bold">{formatarMoeda(p.total_proventos_recebidos)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Conteúdo da Aba 2: Livro-Razão */}
              {abaPreview === "OPERACOES" && (
                <div className="max-h-80 overflow-y-auto custom-scrollbar border border-neutral-800/80 rounded-2xl bg-neutral-950">
                  <table className="w-full text-left text-xs font-mono">
                    <thead className="bg-[#0e0e14] border-b border-neutral-800 text-[10px] text-neutral-400 uppercase tracking-wider sticky top-0">
                      <tr>
                        <th className="p-3">Data</th>
                        <th className="p-3">Tipo</th>
                        <th className="p-3">Ticker</th>
                        <th className="p-3 text-right">Qtd</th>
                        <th className="p-3 text-right">Preço</th>
                        <th className="p-3 text-right">Valor Total</th>
                        <th className="p-3">Notas</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-900">
                      {relatorio.operacoes.slice(0, 100).map((op, idx) => (
                        <tr key={idx} className="hover:bg-neutral-900/50 transition-colors text-[11px]">
                          <td className="p-3 text-neutral-400">{formatarData(op.data)}</td>
                          <td className="p-3">
                            <span className={`text-[10px] px-1.5 py-0.5 rounded font-bold ${
                              op.tipo === "COMPRA"
                                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                                : "bg-rose-500/10 text-rose-400 border border-rose-500/30"
                            }`}>
                              {op.tipo}
                            </span>
                          </td>
                          <td className="p-3 font-extrabold text-white">{op.ticker}</td>
                          <td className="p-3 text-right font-bold">{op.quantidade}</td>
                          <td className="p-3 text-right text-neutral-300">{formatarMoeda(op.preco_unitario)}</td>
                          <td className="p-3 text-right font-bold text-white">{formatarMoeda(op.valor_total)}</td>
                          <td className="p-3 text-[10px] text-neutral-500 truncate max-w-[150px] font-sans">{op.notas}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Conteúdo da Aba 3: Proventos */}
              {abaPreview === "PROVENTOS" && (
                <div className="max-h-80 overflow-y-auto custom-scrollbar border border-neutral-800/80 rounded-2xl bg-neutral-950">
                  <table className="w-full text-left text-xs font-mono">
                    <thead className="bg-[#0e0e14] border-b border-neutral-800 text-[10px] text-neutral-400 uppercase tracking-wider sticky top-0">
                      <tr>
                        <th className="p-3">Data Pagamento</th>
                        <th className="p-3">Ticker</th>
                        <th className="p-3">Tipo</th>
                        <th className="p-3 text-right">Cotas</th>
                        <th className="p-3 text-right">Valor Unitário</th>
                        <th className="p-3 text-right text-emerald-400">Total Líquido</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-900">
                      {relatorio.proventos.slice(0, 100).map((pv, idx) => (
                        <tr key={idx} className="hover:bg-neutral-900/50 transition-colors text-[11px]">
                          <td className="p-3 text-neutral-400">{formatarData(pv.data_pagamento)}</td>
                          <td className="p-3 font-extrabold text-white">{pv.ticker}</td>
                          <td className="p-3">
                            <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold">
                              {pv.tipo}
                            </span>
                          </td>
                          <td className="p-3 text-right text-neutral-300">{pv.quantidade}</td>
                          <td className="p-3 text-right text-neutral-300">{formatarMoeda(pv.valor_unitario)}</td>
                          <td className="p-3 text-right font-bold text-emerald-400">{formatarMoeda(pv.valor_total)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Resumo de Ação */}
              <div className="p-4 rounded-2xl bg-[#0c0c12] border border-amber-500/30 flex items-center justify-between gap-4">
                <div className="text-xs">
                  <span className="font-bold text-amber-300 block">
                    {limparAnteriores
                      ? "⚠️ Modo de Substituição Ativo: os dados manuais anteriores serão substituídos."
                      : "ℹ️ Modo de Mesclagem: os novos dados serão combinados com a carteira existente."}
                  </span>
                  <span className="text-neutral-400 text-[11px]">
                    {relatorio.posicoes.length} ativos serão cadastrados na sua custódia e {relatorio.operacoes.length} transações no livro-razão.
                  </span>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setEtapa("UPLOAD")}
                    className="px-4 py-2.5 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-300 text-xs font-semibold transition-colors"
                  >
                    Voltar
                  </button>
                  <button
                    onClick={handleConfirmarImportacao}
                    disabled={carregando}
                    className="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-[#d4af37] hover:bg-[#c29e2e] text-black font-extrabold text-xs shadow-lg shadow-[#d4af37]/20 transition-all active:scale-95 disabled:opacity-50"
                  >
                    {carregando ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        <span>Gravando...</span>
                      </>
                    ) : (
                      <>
                        <Check className="w-4 h-4" />
                        <span>Confirmar e Importar</span>
                      </>
                    )}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* ------------------------------------------------------------- */}
          {/* ETAPA 3: SUCESSO                                              */}
          {/* ------------------------------------------------------------- */}
          {etapa === "SUCESSO" && resultado && (
            <div className="py-8 text-center space-y-6 animate-in zoom-in-95 duration-200">
              <div className="w-20 h-20 rounded-full bg-emerald-500/10 border-2 border-emerald-500/40 text-emerald-400 flex items-center justify-center mx-auto shadow-2xl shadow-emerald-500/20">
                <CheckCircle2 className="w-10 h-10" />
              </div>

              <div className="space-y-2">
                <h4 className="text-xl font-extrabold text-white font-mono">
                  Carteira Sincronizada com Sucesso!
                </h4>
                <p className="text-xs text-neutral-400 max-w-md mx-auto">
                  Todos os ativos oficiais, histórico contábil e rendimentos foram importados e integrados aos motores de valuation.
                </p>
              </div>

              <div className="grid grid-cols-3 gap-3 max-w-lg mx-auto font-mono text-xs">
                <div className="p-3.5 rounded-2xl bg-neutral-950 border border-neutral-800">
                  <span className="text-[10px] text-neutral-500 uppercase block">Posições Ativas</span>
                  <span className="font-extrabold text-emerald-400 text-base">{resultado.posicoes_salvas}</span>
                </div>
                <div className="p-3.5 rounded-2xl bg-neutral-950 border border-neutral-800">
                  <span className="text-[10px] text-neutral-500 uppercase block">Transações</span>
                  <span className="font-extrabold text-white text-base">{resultado.operacoes_salvas}</span>
                </div>
                <div className="p-3.5 rounded-2xl bg-neutral-950 border border-neutral-800">
                  <span className="text-[10px] text-neutral-500 uppercase block">Proventos Salvos</span>
                  <span className="font-extrabold text-[#d4af37] text-base">{resultado.proventos_salvos}</span>
                </div>
              </div>

              <button
                onClick={handleFechar}
                className="px-8 py-3 rounded-2xl bg-[#d4af37] hover:bg-[#c29e2e] text-black font-extrabold text-xs shadow-xl shadow-[#d4af37]/20 transition-all font-mono"
              >
                Concluir e Ver Carteira Atualizada
              </button>
            </div>
          )}

        </div>

        {/* Modal Footer */}
        <div className="p-4 border-t border-neutral-800 bg-[#0a0a0e] flex items-center justify-between text-xs text-neutral-500 font-mono">
          <span>Área do Investidor B3 • Protocolo Seguro & Idempotente</span>
          <button
            onClick={handleFechar}
            className="px-4 py-2 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-300 transition-colors font-sans text-xs"
          >
            Fechar
          </button>
        </div>
      </div>
    </div>
  );
}

