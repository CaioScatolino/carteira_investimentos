"use client";

import React, { useState } from "react";
import { ClasseAtivo, ResumoCarteira, ClasseConsolidada } from "@/types/portfolio";
import { TrendingUp, Building2, ShieldCheck, Globe, Coins, PieChart } from "lucide-react";

interface PortfolioDonutChartProps {
  resumo: ResumoCarteira;
  classeSelecionada?: ClasseAtivo | null;
  onSelecionarClasse: (classe: ClasseAtivo) => void;
}

const CLASSE_CONFIG: Record<
  ClasseAtivo,
  {
    nome: string;
    cor: string;
    corHover: string;
    corBorda: string;
    corBg: string;
    corTexto: string;
    icon: React.ReactNode;
  }
> = {
  ACAO_BR: {
    nome: "Ações Brasil",
    cor: "#10b981", // Emerald
    corHover: "#34d399",
    corBorda: "border-emerald-500/30",
    corBg: "bg-emerald-500/10",
    corTexto: "text-emerald-400",
    icon: <TrendingUp className="w-4 h-4" />,
  },
  FII: {
    nome: "Fundos Imobiliários",
    cor: "#a855f7", // Purple
    corHover: "#c084fc",
    corBorda: "border-purple-500/30",
    corBg: "bg-purple-500/10",
    corTexto: "text-purple-400",
    icon: <Building2 className="w-4 h-4" />,
  },
  RENDA_FIXA: {
    nome: "Renda Fixa",
    cor: "#3b82f6", // Blue
    corHover: "#60a5fa",
    corBorda: "border-blue-500/30",
    corBg: "bg-blue-500/10",
    corTexto: "text-blue-400",
    icon: <ShieldCheck className="w-4 h-4" />,
  },
  INTERNACIONAL: {
    nome: "Internacional / ETFs",
    cor: "#06b6d4", // Cyan
    corHover: "#22d3ee",
    corBorda: "border-cyan-500/30",
    corBg: "bg-cyan-500/10",
    corTexto: "text-cyan-400",
    icon: <Globe className="w-4 h-4" />,
  },
  CAIXA: {
    nome: "Caixa / Oportunidade",
    cor: "#f59e0b", // Amber
    corHover: "#fbbf24",
    corBorda: "border-amber-500/30",
    corBg: "bg-amber-500/10",
    corTexto: "text-amber-400",
    icon: <Coins className="w-4 h-4" />,
  },
};

export function PortfolioDonutChart({
  resumo,
  classeSelecionada,
  onSelecionarClasse,
}: PortfolioDonutChartProps) {
  const [hoveredClasse, setHoveredClasse] = useState<ClasseAtivo | null>(null);

  const formatarMoeda = (val: number) =>
    new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(val);

  // Calcula segmentos do Donut
  const total = resumo.patrimonio_total > 0 ? resumo.patrimonio_total : 0;
  const classesOrdem: ClasseAtivo[] = ["ACAO_BR", "FII", "RENDA_FIXA", "INTERNACIONAL", "CAIXA"];

  // Monta lista de dados por classe
  const dadosClasses = classesOrdem.map((c) => {
    const info = CLASSE_CONFIG[c];
    // Se o backend mandou classes[], usamos
    const classeObj = resumo.classes?.find((cl) => cl.classe === c);
    const valorMercado =
      classeObj?.valor_atual_mercado ??
      (resumo.alocacao_atual_pct[c] ? (resumo.alocacao_atual_pct[c] / 100) * total : 0);
    const pctAtual = total > 0 ? (valorMercado / total) * 100 : 0;
    const pctAlvo = resumo.alocacao_alvo_pct[c] ?? 0;
    const qtdAtivos = classeObj?.quantidade_ativos ?? resumo.posicoes.filter((p) => p.classe === c).length;

    return {
      classe: c,
      nome: info.nome,
      cor: info.cor,
      corHover: info.corHover,
      valorMercado,
      pctAtual,
      pctAlvo,
      qtdAtivos,
      info,
    };
  });

  // Raio e perímetro do Donut
  const size = 220;
  const strokeWidth = 26;
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;

  // Calcula offset cumulativo para os arcos SVG
  let acumuladoPct = 0;
  const arcos = dadosClasses.map((item) => {
    const strokeDasharray = `${(item.pctAtual / 100) * circumference} ${circumference}`;
    const strokeDashoffset = -((acumuladoPct / 100) * circumference);
    acumuladoPct += item.pctAtual;
    return {
      ...item,
      strokeDasharray,
      strokeDashoffset,
    };
  });

  const classeAtiva = hoveredClasse || classeSelecionada;
  const itemAtivo = classeAtiva ? dadosClasses.find((d) => d.classe === classeAtiva) : null;

  return (
    <div className="p-6 rounded-2xl bg-[#0f1015]/90 border border-neutral-800/80 shadow-2xl mb-8 backdrop-blur-sm">
      {/* Título & Instrução */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-6 border-b border-neutral-800/60 pb-4">
        <div>
          <h3 className="text-base font-bold text-white flex items-center gap-2">
            <PieChart className="w-4 h-4 text-[#d4af37]" />
            Composição e Alocação por Classe de Ativos
          </h3>
          <p className="text-xs text-neutral-400 mt-1">
            Passe o mouse para inspecionar os detalhes ou clique em uma fatia para expandir os ativos na tabela abaixo.
          </p>
        </div>
        <div className="text-xs font-mono text-neutral-400 self-start sm:self-auto bg-neutral-900/80 px-3 py-1.5 rounded-lg border border-neutral-800 flex items-center gap-2">
          <span>Patrimônio:</span>
          <span className="text-[#d4af37] font-bold">{formatarMoeda(total)}</span>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
        {/* Gráfico Donut SVG no Centro/Esquerda */}
        <div className="lg:col-span-5 flex flex-col items-center justify-center relative">
          <div className="relative w-[220px] h-[220px] flex items-center justify-center">
            <svg
              width={size}
              height={size}
              viewBox={`0 0 ${size} ${size}`}
              className="rotate-[-90deg] transition-all duration-300"
            >
              {/* Círculo de Fundo */}
              <circle
                cx={size / 2}
                cy={size / 2}
                r={radius}
                fill="none"
                stroke="#18181b"
                strokeWidth={strokeWidth}
              />

              {/* Fatias das Classes */}
              {total > 0 &&
                arcos.map((arco) => {
                  if (arco.pctAtual <= 0) return null;
                  const isHovered = hoveredClasse === arco.classe;
                  const isSelected = classeSelecionada === arco.classe;

                  return (
                    <circle
                      key={arco.classe}
                      cx={size / 2}
                      cy={size / 2}
                      r={radius}
                      fill="none"
                      stroke={isHovered || isSelected ? arco.corHover : arco.cor}
                      strokeWidth={isHovered || isSelected ? strokeWidth + 4 : strokeWidth}
                      strokeDasharray={arco.strokeDasharray}
                      strokeDashoffset={arco.strokeDashoffset}
                      strokeLinecap="round"
                      className="cursor-pointer transition-all duration-300"
                      style={{
                        filter:
                          isHovered || isSelected
                            ? `drop-shadow(0 0 8px ${arco.cor}80)`
                            : "none",
                      }}
                      onMouseEnter={() => setHoveredClasse(arco.classe)}
                      onMouseLeave={() => setHoveredClasse(null)}
                      onClick={() => onSelecionarClasse(arco.classe)}
                    />
                  );
                })}
            </svg>

            {/* Conteúdo Central do Donut */}
            <div className="absolute inset-0 flex flex-col items-center justify-center text-center pointer-events-none px-4">
              {itemAtivo ? (
                <>
                  <span className="text-[10px] font-mono text-neutral-400 uppercase tracking-wider mb-0.5">
                    {itemAtivo.nome}
                  </span>
                  <span className="text-lg font-black text-white font-mono tracking-tight">
                    {formatarMoeda(itemAtivo.valorMercado)}
                  </span>
                  <div className="flex items-center gap-1.5 mt-1">
                    <span
                      className="text-xs font-bold font-mono px-2 py-0.5 rounded-full"
                      style={{
                        backgroundColor: `${itemAtivo.cor}25`,
                        color: itemAtivo.cor,
                      }}
                    >
                      {itemAtivo.pctAtual.toFixed(1)}% do total
                    </span>
                  </div>
                </>
              ) : (
                <>
                  <span className="text-[10px] font-mono text-neutral-400 uppercase tracking-wider mb-0.5">
                    Patrimônio Total
                  </span>
                  <span className="text-xl font-black text-white font-mono tracking-tight">
                    {formatarMoeda(total)}
                  </span>
                  <span className="text-[11px] font-mono text-neutral-400 mt-1">
                    {resumo.posicoes.length} {resumo.posicoes.length === 1 ? "ativo" : "ativos"} cadastrados
                  </span>
                </>
              )}
            </div>
          </div>

          <div className="text-[11px] text-neutral-400 mt-4 text-center font-mono">
            {classeSelecionada ? (
              <span className="text-[#d4af37]">
                ✓ Filtro ativo: <strong>{CLASSE_CONFIG[classeSelecionada].nome}</strong> (clique novamente para ver todas)
              </span>
            ) : (
              <span>Toque em uma fatia para expandir seus ativos</span>
            )}
          </div>
        </div>

        {/* Legenda Interativa e Métricas por Classe na Direita */}
        <div className="lg:col-span-7 grid grid-cols-1 sm:grid-cols-2 gap-3">
          {dadosClasses.map((item) => {
            const isHovered = hoveredClasse === item.classe;
            const isSelected = classeSelecionada === item.classe;
            const deltaPct = item.pctAtual - item.pctAlvo;

            return (
              <div
                key={item.classe}
                onMouseEnter={() => setHoveredClasse(item.classe)}
                onMouseLeave={() => setHoveredClasse(null)}
                onClick={() => onSelecionarClasse(item.classe)}
                className={`p-3.5 rounded-xl border cursor-pointer transition-all ${
                  isSelected
                    ? "bg-neutral-900 border-[#d4af37] shadow-lg shadow-[#d4af37]/15 ring-1 ring-[#d4af37]"
                    : isHovered
                    ? "bg-neutral-900/90 border-neutral-700 shadow-md"
                    : "bg-neutral-950/60 border-neutral-800/80 hover:border-neutral-700"
                }`}
              >
                <div className="flex items-center justify-between gap-2 mb-2">
                  <div className="flex items-center gap-2">
                    <span
                      className="w-3 h-3 rounded-full shrink-0 shadow-sm"
                      style={{ backgroundColor: item.cor }}
                    />
                    <span className="text-xs font-bold text-white truncate">{item.nome}</span>
                  </div>
                  <span className="text-[10px] font-mono text-neutral-400 bg-neutral-900 px-1.5 py-0.5 rounded border border-neutral-800 shrink-0">
                    {item.qtdAtivos} {item.qtdAtivos === 1 ? "ativo" : "ativos"}
                  </span>
                </div>

                <div className="flex items-baseline justify-between gap-2 mb-2">
                  <span className="text-sm font-extrabold text-white font-mono">
                    {formatarMoeda(item.valorMercado)}
                  </span>
                  <div className="text-right">
                    <span className="text-xs font-black font-mono text-white">
                      {item.pctAtual.toFixed(1)}%
                    </span>
                    <span className="text-[10px] font-mono text-neutral-400 ml-1">
                      / {item.pctAlvo.toFixed(1)}% meta
                    </span>
                  </div>
                </div>

                {/* Mini Barra de Progresso */}
                <div className="w-full h-1.5 bg-neutral-900 rounded-full overflow-hidden flex">
                  <div
                    className="h-full rounded-full transition-all duration-500"
                    style={{
                      width: `${Math.min(100, item.pctAtual)}%`,
                      backgroundColor: item.cor,
                    }}
                  />
                </div>

                <div className="flex items-center justify-between text-[10px] font-mono text-neutral-400 mt-2 pt-1 border-t border-neutral-800/40">
                  <span>Desvio:</span>
                  <span
                    className={`font-bold ${
                      deltaPct < -2.0
                        ? "text-cyan-400"
                        : deltaPct > 3.0
                        ? "text-amber-400"
                        : "text-neutral-400"
                    }`}
                  >
                    {deltaPct >= 0 ? `+${deltaPct.toFixed(1)}%` : `${deltaPct.toFixed(1)}%`}
                    {deltaPct < -2.0 && " • Aporte"}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
