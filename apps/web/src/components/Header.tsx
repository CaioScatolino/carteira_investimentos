"use client";

import React from "react";
import { RefreshCw, Activity, ShieldCheck } from "lucide-react";

interface HeaderProps {
  totalAtivos: number;
  loading: boolean;
  onRefresh: () => void;
  ultimaAtualizacao?: string;
}

export function Header({ totalAtivos, loading, onRefresh, ultimaAtualizacao }: HeaderProps) {
  return (
    <header className="border-b border-[#1f1f26] bg-[#09090c]/90 backdrop-blur-md sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-20 flex items-center justify-between">
        {/* Brand / Logo */}
        <div className="flex items-center gap-3.5">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#d4af37] via-[#917726] to-[#0a0a0d] p-[1px] shadow-lg">
            <div className="w-full h-full rounded-xl bg-[#0b0b0f] flex items-center justify-center">
              <ShieldCheck className="w-5 h-5 text-[#d4af37]" />
            </div>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xl font-extrabold tracking-wider text-white font-mono">
                B3<span className="text-[#d4af37]">CORE</span>
              </span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-[#d4af37]/10 text-[#d4af37] border border-[#d4af37]/30 font-semibold tracking-widest uppercase">
                Terminal
              </span>
            </div>
            <p className="text-[11px] text-neutral-400 hidden sm:block tracking-wide">
              Auditoria Institucional de Mercado & Valuation Fundamentalista
            </p>
          </div>
        </div>

        {/* Live Status & Actions */}
        <div className="flex items-center gap-4">
          <div className="hidden md:flex items-center gap-2 px-3 py-1.5 rounded-lg bg-neutral-900/80 border border-neutral-800 text-xs">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
            <span className="text-neutral-400">Feed B3/CVM:</span>
            <span className="text-white font-mono font-medium">{totalAtivos} Ativos</span>
          </div>

          {ultimaAtualizacao && (
            <span className="text-[11px] text-neutral-500 hidden lg:inline font-mono">
              Atualizado: {ultimaAtualizacao}
            </span>
          )}

          <button
            onClick={onRefresh}
            disabled={loading}
            className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-[#d4af37]/10 hover:bg-[#d4af37]/20 border border-[#d4af37]/40 text-[#d4af37] text-xs font-semibold tracking-wider transition-all duration-200 active:scale-95 disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
            <span className="hidden sm:inline">Sincronizar</span>
          </button>
        </div>
      </div>
    </header>
  );
}
