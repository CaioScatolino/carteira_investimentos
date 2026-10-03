import React from "react";
import { StatusRecomendacao, StatusParecer } from "@/types/market";

export function StatusBadge({ status }: { status: StatusRecomendacao }) {
  switch (status) {
    case "COMPRAR_MAIS":
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold tracking-wider bg-emerald-950/70 text-emerald-400 border border-emerald-500/30">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
          COMPRAR
        </span>
      );
    case "MANTER":
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold tracking-wider bg-amber-950/60 text-amber-300 border border-amber-500/30">
          <span className="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
          MANTER
        </span>
      );
    default:
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold tracking-wider bg-red-950/60 text-rose-400 border border-red-500/30">
          <span className="w-1.5 h-1.5 rounded-full bg-rose-400"></span>
          ALERTA
        </span>
      );
  }
}

export function ParecerBadge({ status }: { status?: StatusParecer }) {
  if (!status) return null;

  switch (status) {
    case "APROVADO":
      return (
        <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-emerald-950/60 text-emerald-300 border border-emerald-500/20">
          ✓ Aprovado
        </span>
      );
    case "ATENCAO":
      return (
        <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-amber-950/50 text-amber-300 border border-amber-500/20">
          ⚠ Atenção
        </span>
      );
    case "REPROVADO":
      return (
        <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-rose-950/50 text-rose-300 border border-rose-500/20">
          ✕ Reprovado
        </span>
      );
    default:
      return (
        <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium bg-neutral-900 text-neutral-400 border border-neutral-800">
          N/A
        </span>
      );
  }
}
