"use client";

import React, { useState, useEffect, useRef } from "react";
import { X, RefreshCw, AlertCircle, Loader2, TrendingUp, TrendingDown, Eye, EyeOff, BarChart2, ShieldCheck } from "lucide-react";
import {
  createChart,
  ColorType,
  CandlestickSeries,
  LineSeries,
  AreaSeries,
  LineStyle,
  IChartApi,
  ISeriesApi,
  Time,
} from "lightweight-charts";
import { RespostaGrafico } from "@/types/chart";
import { API_V1_URL } from "@/config/api";

interface AssetChartModalProps {
  isOpen: boolean;
  onClose: () => void;
  ticker: string | null;
  nomeInicial?: string;
}

export function AssetChartModal({
  isOpen,
  onClose,
  ticker,
  nomeInicial = "",
}: AssetChartModalProps) {
  const [intervalo, setIntervalo] = useState<"1d" | "1wk">("1wk");
  const [periodo, setPeriodo] = useState<"6mo" | "1y" | "2y" | "5y">("5y");
  const [loading, setLoading] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [dados, setDados] = useState<RespostaGrafico | null>(null);

  // Filtros de visualização dos indicadores
  const [mostrarMMA8, setMostrarMMA8] = useState(true);
  const [mostrarMMA21, setMostrarMMA21] = useState(true);
  const [mostrarMMA20, setMostrarMMA20] = useState(true);
  const [mostrarMMA52, setMostrarMMA52] = useState(true);
  const [mostrarMMA200, setMostrarMMA200] = useState(true);
  const [mostrarBollinger, setMostrarBollinger] = useState(true);
  const [mostrarTeto, setMostrarTeto] = useState(true);
  const [mostrarRSI, setMostrarRSI] = useState(true);

  // Referências DOM para os containers dos gráficos
  const chartContainerRef = useRef<HTMLDivElement | null>(null);
  const rsiContainerRef = useRef<HTMLDivElement | null>(null);

  // Instâncias dos gráficos
  const mainChartRef = useRef<IChartApi | null>(null);
  const rsiChartRef = useRef<IChartApi | null>(null);

  // 1. Busca os dados no backend Go
  const carregarDadosGrafico = async () => {
    if (!ticker) return;
    setLoading(true);
    setErro(null);

    try {
      const res = await fetch(
        `${API_V1_URL}/ativos/${ticker}/grafico?intervalo=${intervalo}&periodo=${periodo}`,
        { headers: { Accept: "application/json" } }
      );

      if (!res.ok) {
        const errJson = await res.json().catch(() => null);
        throw new Error(errJson?.erro || `Falha ao carregar gráfico (${res.status})`);
      }

      const json: RespostaGrafico = await res.json();
      setDados(json);
    } catch (err: any) {
      console.warn("Erro ao buscar dados do gráfico:", err);
      setErro(err.message || "Erro de conexão ao carregar série histórica.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen && ticker) {
      carregarDadosGrafico();
    } else {
      setDados(null);
      setErro(null);
    }
  }, [isOpen, ticker, intervalo, periodo]);

  // Fecha com ESC
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  // 2. Renderização do Gráfico com Lightweight Charts
  useEffect(() => {
    if (!dados || !chartContainerRef.current) return;

    // Destrói instâncias anteriores
    if (mainChartRef.current) {
      mainChartRef.current.remove();
      mainChartRef.current = null;
    }
    if (rsiChartRef.current) {
      rsiChartRef.current.remove();
      rsiChartRef.current = null;
    }

    const containerWidth = chartContainerRef.current.clientWidth;
    const mainHeight = mostrarRSI ? 380 : 500;

    // --- GRÁFICO PRINCIPAL (CANDLES + MÉDIAS + TETO) ---
    const chart = createChart(chartContainerRef.current, {
      width: containerWidth,
      height: mainHeight,
      layout: {
        background: { type: ColorType.Solid, color: "#090a0f" },
        textColor: "#94a3b8",
        fontSize: 11,
      },
      grid: {
        vertLines: { color: "rgba(255, 255, 255, 0.04)" },
        horzLines: { color: "rgba(255, 255, 255, 0.04)" },
      },
      crosshair: {
        vertLine: { color: "rgba(212, 175, 55, 0.4)", width: 1, style: LineStyle.Dashed },
        horzLine: { color: "rgba(212, 175, 55, 0.4)", width: 1, style: LineStyle.Dashed },
      },
      timeScale: {
        borderColor: "rgba(255, 255, 255, 0.1)",
        timeVisible: true,
        secondsVisible: false,
      },
      rightPriceScale: {
        borderColor: "rgba(255, 255, 255, 0.1)",
        autoScale: true,
        minimumWidth: 105,
      },
    });

    mainChartRef.current = chart;

    // Série de Candlesticks
    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: "#10b981",
      downColor: "#ef4444",
      borderVisible: false,
      wickUpColor: "#10b981",
      wickDownColor: "#ef4444",
    });

    const formattedCandles = dados.candles.map((c) => ({
      time: c.time as Time,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }));
    candleSeries.setData(formattedCandles);

    // Linha Reta Contínua de Preço Teto Consolidado (Dourada Institucional)
    // + Shaded Zone esverdeada sutil da linha do Teto para baixo (Zona de Compra Fundamentalista)
    const tetoValor =
      (dados.preco_teto_consolidado && dados.preco_teto_consolidado > 0)
        ? dados.preco_teto_consolidado
        : dados.preco_teto_bazin;

    if (mostrarTeto && tetoValor > 0 && dados.candles.length > 0) {
      // 1. Shaded Zone sutil esverdeada do TETO CONSOLIDADO para baixo
      const tetoArea = chart.addSeries(AreaSeries, {
        topColor: "rgba(16, 185, 129, 0.10)", // Verde esmeralda suave
        bottomColor: "rgba(16, 185, 129, 0.01)", // Gradiente até transparente
        lineColor: "transparent",
        lineWidth: 1,
        priceLineVisible: false,
        lastValueVisible: false,
        crosshairMarkerVisible: false,
      });
      tetoArea.setData(
        dados.candles.map((c) => ({
          time: c.time as Time,
          value: tetoValor,
        }))
      );

      // 2. Linha Reta Contínua Dourada do TETO CONSOLIDADO
      const tetoSeries = chart.addSeries(LineSeries, {
        color: "#d4af37",
        lineWidth: 2,
        lineStyle: LineStyle.Solid,
        priceLineVisible: true,
        lastValueVisible: true,
        title: "TETO CONSOLIDADO",
      });

      tetoSeries.setData(
        dados.candles.map((c) => ({
          time: c.time as Time,
          value: tetoValor,
        }))
      );

      // 3. Faixa de Sobrepreço (+15% acima do Teto - Zona de Cautela/Venda)
      const sobrepreco15 = tetoValor * 1.15;
      candleSeries.createPriceLine({
        price: sobrepreco15,
        color: "rgba(239, 68, 68, 0.65)",
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        axisLabelVisible: true,
        title: "SOBREPREÇO (+15%)",
      });
    }

    // Linha de Valor de Graham (Ciano Pontilhada)
    if (mostrarTeto && dados.valor_graham > 0) {
      candleSeries.createPriceLine({
        price: dados.valor_graham,
        color: "#06b6d4",
        lineWidth: 1,
        lineStyle: LineStyle.Dotted,
        axisLabelVisible: true,
        title: "GRAHAM",
      });
    }

    // --- MÉDIAS MÓVEIS & INDICADORES CONFORME TEMPO GRÁFICO ---
    if (intervalo === "1wk") {
      // 1S (Semanal) - Tríade Institucional:
      // MMA 20 Semanal (Gatilho de Direção ~5 meses)
      if (mostrarMMA20 && dados.mma20 && dados.mma20.length > 0) {
        const mma20Series = chart.addSeries(LineSeries, {
          color: "#eab308", // Amarelo ouro
          lineWidth: 2,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "MMA 20 (SEM)",
        });
        mma20Series.setData(dados.mma20.map((pt) => ({ time: pt.time as Time, value: pt.value })));
      }

      // MMA 52 Semanal (O Eixo de 1 Ano)
      if (mostrarMMA52 && dados.mma52 && dados.mma52.length > 0) {
        const mma52Series = chart.addSeries(LineSeries, {
          color: "#06b6d4", // Ciano
          lineWidth: 2,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "MMA 52 (1 ANO)",
        });
        mma52Series.setData(dados.mma52.map((pt) => ({ time: pt.time as Time, value: pt.value })));
      }

      // MMA 200 Semanal (O Chão de Concreto ~4 anos)
      if (mostrarMMA200 && dados.mma200 && dados.mma200.length > 0) {
        const mma200Series = chart.addSeries(LineSeries, {
          color: "#a855f7", // Roxo Institucional
          lineWidth: 2,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "MMA 200 (SEM)",
        });
        mma200Series.setData(dados.mma200.map((pt) => ({ time: pt.time as Time, value: pt.value })));
      }

      // Bandas de Bollinger (20, 2) no Semanal - Áreas de Compra e Venda
      if (
        mostrarBollinger &&
        dados.bollinger_superior &&
        dados.bollinger_inferior &&
        dados.bollinger_superior.length > 0
      ) {
        const bollSupSeries = chart.addSeries(LineSeries, {
          color: "rgba(244, 63, 94, 0.75)", // Rosa/Vermelho (Área de Venda / Esticada)
          lineWidth: 1,
          lineStyle: LineStyle.Dashed,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "BOLL SUP (VENDA)",
        });
        bollSupSeries.setData(
          dados.bollinger_superior.map((pt) => ({ time: pt.time as Time, value: pt.value }))
        );

        const bollInfSeries = chart.addSeries(LineSeries, {
          color: "rgba(16, 185, 129, 0.8)", // Verde (Área de Compra Técnica)
          lineWidth: 1,
          lineStyle: LineStyle.Dashed,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "BOLL INF (COMPRA)",
        });
        bollInfSeries.setData(
          dados.bollinger_inferior.map((pt) => ({ time: pt.time as Time, value: pt.value }))
        );
      }
    } else {
      // 1D (Diário) - Tríade Clássica:
      if (mostrarMMA8 && dados.mma8 && dados.mma8.length > 0) {
        const mma8Series = chart.addSeries(LineSeries, {
          color: "#eab308",
          lineWidth: 1,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "MMA 8",
        });
        mma8Series.setData(dados.mma8.map((pt) => ({ time: pt.time as Time, value: pt.value })));
      }

      if (mostrarMMA21 && dados.mma21 && dados.mma21.length > 0) {
        const mma21Series = chart.addSeries(LineSeries, {
          color: "#06b6d4",
          lineWidth: 2,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "MMA 21",
        });
        mma21Series.setData(dados.mma21.map((pt) => ({ time: pt.time as Time, value: pt.value })));
      }

      if (mostrarMMA200 && dados.mma200 && dados.mma200.length > 0) {
        const mma200Series = chart.addSeries(LineSeries, {
          color: "#a855f7",
          lineWidth: 2,
          priceLineVisible: false,
          lastValueVisible: true,
          title: "MMA 200",
        });
        mma200Series.setData(dados.mma200.map((pt) => ({ time: pt.time as Time, value: pt.value })));
      }
    }

    // --- SUB-GRÁFICO INFERIOR: IFR / RSI (14) ---
    if (mostrarRSI && rsiContainerRef.current && dados.rsi14 && dados.rsi14.length > 0) {
      const rsiChart = createChart(rsiContainerRef.current, {
        width: containerWidth,
        height: 140,
        layout: {
          background: { type: ColorType.Solid, color: "#07080c" },
          textColor: "#64748b",
          fontSize: 10,
        },
        grid: {
          vertLines: { color: "rgba(255, 255, 255, 0.03)" },
          horzLines: { color: "rgba(255, 255, 255, 0.03)" },
        },
        timeScale: {
          borderColor: "rgba(255, 255, 255, 0.1)",
          visible: true,
          timeVisible: true,
          secondsVisible: false,
        },
        rightPriceScale: {
          borderColor: "rgba(255, 255, 255, 0.1)",
          scaleMargins: { top: 0.1, bottom: 0.1 },
          minimumWidth: 105,
        },
      });

      rsiChartRef.current = rsiChart;

      const rsiSeries = rsiChart.addSeries(LineSeries, {
        color: "#38bdf8",
        lineWidth: 2,
        priceLineVisible: false,
        lastValueVisible: true,
        title: "IFR (14)",
      });

      // Alinhamento temporal 1-para-1 com os candles:
      // O IFR (14) exige 14 períodos para o primeiro ponto. Usamos WhitespaceData para preencher
      // as primeiras barras vazias com o timestamp exato dos candles, garantindo que o índice
      // de cada barra no gráfico principal corresponda perfeitamente ao mesmo índice no gráfico de IFR.
      const rsiMap = new Map<number, number>(dados.rsi14.map((pt) => [pt.time, pt.value]));
      const fullRsiData = dados.candles.map((c) => {
        const val = rsiMap.get(c.time);
        if (val !== undefined) {
          return { time: c.time as Time, value: val };
        }
        return { time: c.time as Time };
      });

      rsiSeries.setData(fullRsiData as any);

      // Linhas de Nível 30/35 (Sobrevenda) e 70 (Sobrecompra)
      const nivelSobrevenda = intervalo === "1wk" ? 35 : 30;
      const labelSobrevenda = intervalo === "1wk" ? "35 SOBREVENDA (1S)" : "30 SOBREVENDA";

      rsiSeries.createPriceLine({
        price: 70,
        color: "#ef4444",
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        axisLabelVisible: true,
        title: "70 SOBRECOMPRA",
      });

      rsiSeries.createPriceLine({
        price: nivelSobrevenda,
        color: "#10b981",
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        axisLabelVisible: true,
        title: labelSobrevenda,
      });

      // Garante que o eixo de preços do IFR mantenha SEMPRE a escala mínima entre 20 e 80
      // para que a linha 30/35 verde e a linha 70 vermelha estejam 100% visíveis em tela!
      rsiSeries.applyOptions({
        autoscaleInfoProvider: (original: any) => {
          const res = original ? original() : null;
          if (!res || !res.priceRange) {
            return { priceRange: { minValue: 20, maxValue: 80 } };
          }
          return {
            priceRange: {
              minValue: Math.min(res.priceRange.minValue, 20),
              maxValue: Math.max(res.priceRange.maxValue, 80),
            },
          };
        },
      });

      // Sincronização bidirecional do zoom e rolagem temporal protegida contra loops
      let isSyncingRange = false;
      const syncFromMain = (range: any) => {
        if (isSyncingRange || !range || !rsiChartRef.current) return;
        isSyncingRange = true;
        rsiChartRef.current.timeScale().setVisibleLogicalRange(range);
        isSyncingRange = false;
      };

      const syncFromRsi = (range: any) => {
        if (isSyncingRange || !range || !mainChartRef.current) return;
        isSyncingRange = true;
        mainChartRef.current.timeScale().setVisibleLogicalRange(range);
        isSyncingRange = false;
      };

      chart.timeScale().subscribeVisibleLogicalRangeChange(syncFromMain);
      rsiChart.timeScale().subscribeVisibleLogicalRangeChange(syncFromRsi);

      // Sincronização do Crosshair (linha vertical guia em ambos os gráficos)
      chart.subscribeCrosshairMove((param) => {
        if (!rsiChartRef.current || !rsiSeries) return;
        if (!param.point || !param.time) {
          rsiChartRef.current.clearCrosshairPosition();
        } else {
          const val = rsiMap.get(param.time as number) ?? 50;
          rsiChartRef.current.setCrosshairPosition(val, param.time, rsiSeries);
        }
      });

      rsiChart.subscribeCrosshairMove((param) => {
        if (!mainChartRef.current || !candleSeries) return;
        if (!param.point || !param.time) {
          mainChartRef.current.clearCrosshairPosition();
        } else {
          const candle = dados.candles.find((c) => c.time === param.time);
          if (candle) {
            mainChartRef.current.setCrosshairPosition(candle.close, param.time, candleSeries);
          }
        }
      });
    }

    // Ajusta ambos os gráficos para caber na tela com o mesmo enquadramento inicial
    chart.timeScale().fitContent();
    if (rsiChartRef.current) {
      rsiChartRef.current.timeScale().fitContent();
    }

    // Redimensionamento responsivo automático
    const handleResize = () => {
      if (chartContainerRef.current) {
        const w = chartContainerRef.current.clientWidth;
        chart.applyOptions({ width: w });
        if (rsiChartRef.current) {
          rsiChartRef.current.applyOptions({ width: w });
        }
      }
    };

    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      if (mainChartRef.current) {
        mainChartRef.current.remove();
        mainChartRef.current = null;
      }
      if (rsiChartRef.current) {
        rsiChartRef.current.remove();
        rsiChartRef.current = null;
      }
    };
  }, [
    dados,
    intervalo,
    mostrarMMA8,
    mostrarMMA21,
    mostrarMMA20,
    mostrarMMA52,
    mostrarMMA200,
    mostrarBollinger,
    mostrarTeto,
    mostrarRSI,
  ]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-black/85 backdrop-blur-md animate-in fade-in duration-200">
      <div className="bg-[#0b0c11] border border-neutral-800 rounded-2xl w-full max-w-6xl max-h-[95vh] flex flex-col shadow-2xl overflow-hidden">
        
        {/* Header do Modal */}
        <div className="p-4 sm:p-5 border-b border-neutral-800/80 flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-gradient-to-r from-neutral-950 via-[#0e0f17] to-neutral-950">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-[#d4af37]/15 border border-[#d4af37]/30 flex items-center justify-center text-[#d4af37] font-black text-sm shadow-md">
              <BarChart2 className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-xl sm:text-2xl font-black text-white tracking-wide">
                  {ticker}
                </h2>
                <span className="text-[10px] px-2 py-0.5 rounded bg-neutral-800 text-neutral-400 font-mono">
                  B3 OFICIAL
                </span>
                {dados?.sinal_tecnico && (
                  <span
                    className={`text-[10px] font-bold px-2 py-0.5 rounded border ${
                      dados.sinal_tecnico.includes("COMPRA") || dados.sinal_tecnico.includes("SOBREVENDIDO")
                        ? "bg-emerald-500/15 text-emerald-400 border-emerald-500/30"
                        : dados.sinal_tecnico.includes("ALERTA")
                        ? "bg-red-500/15 text-red-400 border-red-500/30"
                        : "bg-neutral-800 text-neutral-300 border-neutral-700"
                    }`}
                  >
                    {dados.sinal_tecnico.replaceAll("_", " ")}
                  </span>
                )}
              </div>
              <p className="text-xs text-neutral-400 truncate max-w-md mt-0.5">
                {dados?.nome || nomeInicial || "Série histórica de preços e médias institucionais"}
              </p>
            </div>
          </div>

          {/* Cards de Métricas Rápidas */}
          <div className="flex items-center gap-3 overflow-x-auto pb-1 lg:pb-0 font-mono text-xs">
            <div className="px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 flex flex-col">
              <span className="text-[9px] text-neutral-500 uppercase">Preço Atual</span>
              <span className="font-bold text-white">
                R$ {dados ? dados.preco_atual.toFixed(2) : "--"}
              </span>
            </div>

            {dados && ((dados.preco_teto_consolidado && dados.preco_teto_consolidado > 0) || (dados.preco_teto_bazin && dados.preco_teto_bazin > 0)) ? (
              <div className="px-3 py-1.5 rounded-lg bg-[#d4af37]/10 border border-[#d4af37]/30 flex flex-col">
                <span className="text-[9px] text-[#f5df88] uppercase font-bold flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-[#d4af37]"></span>
                  Teto Consolidado
                </span>
                <span className="font-extrabold text-[#d4af37]">
                  R$ {(dados.preco_teto_consolidado || dados.preco_teto_bazin).toFixed(2)}
                </span>
                {dados.modelos_teto && dados.modelos_teto.length > 0 && (
                  <span className="text-[8px] text-neutral-400 font-sans tracking-tight">
                    {dados.modelos_teto.join(" + ")}
                  </span>
                )}
              </div>
            ) : null}

            {dados && (
              <div className="px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 flex flex-col">
                <span className="text-[9px] text-neutral-500 uppercase">Margem Seg.</span>
                <span
                  className={`font-bold flex items-center gap-1 ${
                    dados.margem_seguranca >= 0 ? "text-emerald-400" : "text-red-400"
                  }`}
                >
                  {dados.margem_seguranca >= 0 ? (
                    <TrendingUp className="w-3 h-3" />
                  ) : (
                    <TrendingDown className="w-3 h-3" />
                  )}
                  {dados.margem_seguranca >= 0 ? "+" : ""}
                  {dados.margem_seguranca.toFixed(1)}%
                </span>
              </div>
            )}

            {dados && dados.rsi14 && dados.rsi14.length > 0 && (
              <div className="px-3 py-1.5 rounded-lg bg-neutral-900 border border-neutral-800 flex flex-col">
                <span className="text-[9px] text-neutral-500 uppercase">IFR (14)</span>
                <span
                  className={`font-bold ${
                    dados.rsi14[dados.rsi14.length - 1].value <= 35
                      ? "text-emerald-400"
                      : dados.rsi14[dados.rsi14.length - 1].value >= 70
                      ? "text-red-400"
                      : "text-neutral-200"
                  }`}
                >
                  {dados.rsi14[dados.rsi14.length - 1].value.toFixed(1)}
                </span>
              </div>
            )}

            <button
              onClick={onClose}
              className="p-2 rounded-xl bg-neutral-900 hover:bg-neutral-800 text-neutral-400 hover:text-white border border-neutral-800 transition-colors shrink-0 ml-1"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Barra de Controles e Legenda */}
        <div className="px-4 py-2.5 bg-neutral-950 border-b border-neutral-800 flex flex-wrap items-center justify-between gap-3 text-xs">
          {/* Seletor de Tempo e Período */}
          <div className="flex items-center gap-2">
            <div className="flex items-center rounded-lg bg-neutral-900 p-0.5 border border-neutral-800">
              <button
                onClick={() => setIntervalo("1d")}
                className={`px-2.5 py-1 rounded text-xs font-bold transition-all ${
                  intervalo === "1d" ? "bg-[#d4af37] text-black" : "text-neutral-400 hover:text-white"
                }`}
              >
                1D (Diário)
              </button>
              <button
                onClick={() => setIntervalo("1wk")}
                className={`px-2.5 py-1 rounded text-xs font-bold transition-all ${
                  intervalo === "1wk" ? "bg-[#d4af37] text-black" : "text-neutral-400 hover:text-white"
                }`}
              >
                1S (Semanal)
              </button>
            </div>

            <div className="h-4 w-px bg-neutral-800 hidden sm:block"></div>

            <div className="flex items-center rounded-lg bg-neutral-900 p-0.5 border border-neutral-800">
              <button
                onClick={() => setPeriodo("6mo")}
                className={`px-2 py-1 rounded text-[11px] font-mono transition-all ${
                  periodo === "6mo" ? "bg-neutral-700 text-white font-bold" : "text-neutral-400 hover:text-white"
                }`}
              >
                6M
              </button>
              <button
                onClick={() => setPeriodo("1y")}
                className={`px-2 py-1 rounded text-[11px] font-mono transition-all ${
                  periodo === "1y" ? "bg-neutral-700 text-white font-bold" : "text-neutral-400 hover:text-white"
                }`}
              >
                1A
              </button>
              <button
                onClick={() => setPeriodo("2y")}
                className={`px-2 py-1 rounded text-[11px] font-mono transition-all ${
                  periodo === "2y" ? "bg-neutral-700 text-white font-bold" : "text-neutral-400 hover:text-white"
                }`}
              >
                2A
              </button>
              <button
                onClick={() => setPeriodo("5y")}
                className={`px-2 py-1 rounded text-[11px] font-mono transition-all ${
                  periodo === "5y" ? "bg-neutral-700 text-white font-bold" : "text-neutral-400 hover:text-white"
                }`}
                title="5 Anos (recomendado para MMA 200 Semanal)"
              >
                5A
              </button>
            </div>

            <button
              onClick={carregarDadosGrafico}
              disabled={loading}
              className="p-1.5 rounded-lg bg-neutral-900 hover:bg-neutral-800 text-neutral-400 hover:text-white border border-neutral-800 transition-colors"
              title="Recarregar cotações"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-[#d4af37]" : ""}`} />
            </button>
          </div>

          {/* Alternadores de Indicadores (Toggles) */}
          <div className="flex items-center gap-2 flex-wrap">
            <button
              onClick={() => setMostrarTeto(!mostrarTeto)}
              className={`px-2.5 py-1 rounded border text-[11px] font-medium flex items-center gap-1.5 transition-all ${
                mostrarTeto
                  ? "bg-[#d4af37]/15 border-[#d4af37]/50 text-[#f5df88]"
                  : "bg-neutral-900 border-neutral-800 text-neutral-500"
              }`}
              title="Exibir ou ocultar linha reta de Teto Consolidado e Faixa de Compra"
            >
              <span className="w-3 h-0.5 bg-[#d4af37] rounded-full inline-block"></span>
              <span>Teto Consolidado</span>
            </button>

            {intervalo === "1wk" ? (
              <>
                <button
                  onClick={() => setMostrarMMA20(!mostrarMMA20)}
                  className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                    mostrarMMA20
                      ? "bg-yellow-500/15 border-yellow-500/40 text-yellow-300"
                      : "bg-neutral-900 border-neutral-800 text-neutral-500"
                  }`}
                  title="MMA 20 Semanal (~5 Meses) - Gatilho de Direção"
                >
                  <span className="w-2 h-2 rounded-full bg-yellow-400"></span>
                  <span>MMA 20 (Gatilho)</span>
                </button>

                <button
                  onClick={() => setMostrarMMA52(!mostrarMMA52)}
                  className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                    mostrarMMA52
                      ? "bg-cyan-500/15 border-cyan-500/40 text-cyan-300"
                      : "bg-neutral-900 border-neutral-800 text-neutral-500"
                  }`}
                  title="MMA 52 Semanal - O Eixo de 1 Ano"
                >
                  <span className="w-2 h-2 rounded-full bg-cyan-400"></span>
                  <span>MMA 52 (Eixo 1A)</span>
                </button>

                <button
                  onClick={() => setMostrarMMA200(!mostrarMMA200)}
                  className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                    mostrarMMA200
                      ? "bg-purple-500/15 border-purple-500/40 text-purple-300"
                      : "bg-neutral-900 border-neutral-800 text-neutral-500"
                  }`}
                  title="MMA 200 Semanal (~4 Anos) - Chão de Concreto"
                >
                  <span className="w-2 h-2 rounded-full bg-purple-400"></span>
                  <span>MMA 200 (Chão)</span>
                </button>

                <button
                  onClick={() => setMostrarBollinger(!mostrarBollinger)}
                  className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                    mostrarBollinger
                      ? "bg-emerald-500/15 border-emerald-500/40 text-emerald-300"
                      : "bg-neutral-900 border-neutral-800 text-neutral-500"
                  }`}
                  title="Bandas de Bollinger (20, 2) no Semanal - Áreas de Compra e Venda"
                >
                  <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
                  <span>Bollinger (20,2)</span>
                </button>
              </>
            ) : (
              <>
                <button
                  onClick={() => setMostrarMMA8(!mostrarMMA8)}
                  className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                    mostrarMMA8
                      ? "bg-yellow-500/15 border-yellow-500/40 text-yellow-300"
                      : "bg-neutral-900 border-neutral-800 text-neutral-500"
                  }`}
                  title="MMA 8 Diária - Momentum Curto Prazo"
                >
                  <span className="w-2 h-2 rounded-full bg-yellow-400"></span>
                  <span>MMA 8</span>
                </button>

                <button
                  onClick={() => setMostrarMMA21(!mostrarMMA21)}
                  className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                    mostrarMMA21
                      ? "bg-cyan-500/15 border-cyan-500/40 text-cyan-300"
                      : "bg-neutral-900 border-neutral-800 text-neutral-500"
                  }`}
                  title="MMA 21 Diária - Linha d'água de Médio Prazo"
                >
                  <span className="w-2 h-2 rounded-full bg-cyan-400"></span>
                  <span>MMA 21</span>
                </button>

                <button
                  onClick={() => setMostrarMMA200(!mostrarMMA200)}
                  className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                    mostrarMMA200
                      ? "bg-purple-500/15 border-purple-500/40 text-purple-300"
                      : "bg-neutral-900 border-neutral-800 text-neutral-500"
                  }`}
                  title="MMA 200 Diária - Tendência Primária Institucional"
                >
                  <span className="w-2 h-2 rounded-full bg-purple-400"></span>
                  <span>MMA 200</span>
                </button>
              </>
            )}

            <button
              onClick={() => setMostrarRSI(!mostrarRSI)}
              className={`px-2 py-1 rounded border text-[11px] flex items-center gap-1.5 transition-all ${
                mostrarRSI
                  ? "bg-sky-500/15 border-sky-500/40 text-sky-300"
                  : "bg-neutral-900 border-neutral-800 text-neutral-500"
              }`}
            >
              <span className="w-2 h-2 rounded-full bg-sky-400"></span>
              <span>IFR 14</span>
            </button>
          </div>
        </div>

        {/* Parecer Técnico e Fundamentalista Sintetizado */}
        {dados?.descricao_sinal && (
          <div className="px-4 py-2 bg-neutral-900/60 border-b border-neutral-800/80 flex items-center gap-2 text-xs">
            <ShieldCheck className="w-4 h-4 text-[#d4af37] shrink-0" />
            <span className="text-neutral-300 font-medium">{dados.descricao_sinal}</span>
          </div>
        )}

        {/* Área do Gráfico */}
        <div className="flex-1 overflow-y-auto p-4 bg-[#070709] relative flex flex-col justify-center min-h-[420px]">
          {loading && (
            <div className="absolute inset-0 z-10 flex flex-col items-center justify-center bg-black/60 backdrop-blur-xs">
              <Loader2 className="w-8 h-8 text-[#d4af37] animate-spin mb-2" />
              <p className="text-xs text-neutral-400 font-mono">
                Calculando velas e médias institucionais para {ticker}...
              </p>
            </div>
          )}

          {erro ? (
            <div className="max-w-md mx-auto p-6 rounded-xl bg-red-950/30 border border-red-500/40 text-center">
              <AlertCircle className="w-8 h-8 text-red-400 mx-auto mb-2" />
              <h3 className="text-sm font-bold text-white mb-1">Falha ao carregar gráfico</h3>
              <p className="text-xs text-red-200 mb-4">{erro}</p>
              <button
                onClick={carregarDadosGrafico}
                className="px-4 py-2 rounded-lg bg-red-500 hover:bg-red-600 text-white font-bold text-xs"
              >
                Tentar Novamente
              </button>
            </div>
          ) : (
            <div className="w-full flex flex-col gap-2">
              {/* Gráfico Principal de Velas */}
              <div
                ref={chartContainerRef}
                className="w-full rounded-xl overflow-hidden border border-neutral-800/70"
              />

              {/* Sub-gráfico de IFR / RSI */}
              {mostrarRSI && (
                <div className="w-full">
                  <div className="flex items-center justify-between text-[10px] text-neutral-500 font-mono px-1 pb-1">
                    <span>OSCILADOR IFR (ÍNDICE DE FORÇA RELATIVA - 14)</span>
                    <span className="text-neutral-400">
                      Linha {intervalo === "1wk" ? "35" : "30"} (Sobrevenda) | Linha 70 (Sobrecompra)
                    </span>
                  </div>
                  <div
                    ref={rsiContainerRef}
                    className="w-full rounded-xl overflow-hidden border border-neutral-800/70"
                  />
                </div>
              )}
            </div>
          )}
        </div>

        {/* Rodapé Informativo */}
        <div className="p-3 bg-[#08090d] border-t border-neutral-800 text-[11px] text-neutral-500 flex flex-col sm:flex-row items-center justify-between gap-2">
          <div className="flex items-center gap-3">
            <span>
              <strong className="text-neutral-400">Dica:</strong> Use o scroll do mouse para Zoom e arraste para rolar a série temporal.
            </span>
          </div>
          <span className="text-neutral-500 font-mono text-[10px]">
            Engine: Lightweight Charts (TradingView) • B3 Cache Redis
          </span>
        </div>

      </div>
    </div>
  );
}
