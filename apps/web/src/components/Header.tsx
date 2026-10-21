"use client";

import React, { useState } from "react";
import { RefreshCw, ShieldCheck, Lock, LogOut, User } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { AuthModal } from "./AuthModal";

interface HeaderProps {
  totalAtivos: number;
  loading: boolean;
  onRefresh: () => void;
  ultimaAtualizacao?: string;
  abaAtiva?: "mercado" | "carteira";
  onMudarAba?: (aba: "mercado" | "carteira") => void;
}

export function Header({
  totalAtivos,
  loading,
  onRefresh,
  ultimaAtualizacao,
  abaAtiva = "mercado",
  onMudarAba,
}: HeaderProps) {
  const { usuario, logout, loading: authLoading } = useAuth();
  const [modalAuthAberto, setModalAuthAberto] = useState(false);
  const [tabInicial, setTabInicial] = useState<"login" | "register">("login");

  const abrirLogin = () => {
    setTabInicial("login");
    setModalAuthAberto(true);
  };

  const getIniciais = (nome?: string) => {
    if (!nome || typeof nome !== "string") return "US";
    const limpo = nome.trim();
    if (!limpo) return "US";
    const partes = limpo.split(" ").filter(Boolean);
    if (partes.length === 0) return "US";
    if (partes.length === 1) return partes[0].slice(0, 2).toUpperCase();
    return (partes[0][0] + (partes[partes.length - 1]?.[0] || "")).toUpperCase();
  };

  const primeiroNome = usuario?.nome && typeof usuario.nome === "string"
    ? usuario.nome.trim().split(" ")[0]
    : (usuario?.email ? usuario.email.split("@")[0] : "Investidor");

  return (
    <>
      <header className="border-b border-[#1f1f26] bg-[#09090c]/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-20 flex items-center justify-between">
          {/* Brand / Logo */}
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-3.5 cursor-pointer" onClick={() => onMudarAba?.("mercado")}>
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

            {/* Navegação Principal */}
            {onMudarAba && (
              <nav className="hidden md:flex items-center gap-1.5 p-1 rounded-xl bg-neutral-900/90 border border-neutral-800">
                <button
                  onClick={() => onMudarAba("mercado")}
                  className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                    abaAtiva === "mercado"
                      ? "bg-[#d4af37] text-black font-bold shadow-md shadow-[#d4af37]/20"
                      : "text-neutral-400 hover:text-white"
                  }`}
                >
                  Mercado & Rankings
                </button>
                <button
                  onClick={() => onMudarAba("carteira")}
                  className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 ${
                    abaAtiva === "carteira"
                      ? "bg-[#d4af37] text-black font-bold shadow-md shadow-[#d4af37]/20"
                      : "text-neutral-400 hover:text-white"
                  }`}
                >
                  <span>Minha Carteira</span>
                  <span className="text-[9px] px-1.5 py-0.2 rounded-full bg-[#d4af37]/20 text-[#d4af37] font-mono border border-[#d4af37]/40">
                    40/22.5/21
                  </span>
                </button>
              </nav>
            )}
          </div>

          {/* Live Status & Actions */}
          <div className="flex items-center gap-3 sm:gap-4">
            <div className="hidden md:flex items-center gap-2 px-3 py-1.5 rounded-lg bg-neutral-900/80 border border-neutral-800 text-xs">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
              <span className="text-neutral-400">Feed B3/CVM:</span>
              <span className="text-white font-mono font-medium">{totalAtivos} Ativos</span>
            </div>

            {ultimaAtualizacao && (
              <span className="text-[11px] text-neutral-500 hidden lg:inline font-mono">
                {ultimaAtualizacao}
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

            {/* Divisor */}
            <div className="h-6 w-[1px] bg-neutral-800 hidden sm:block"></div>

            {/* Área de Autenticação */}
            {!authLoading && (
              usuario ? (
                <div className="flex items-center gap-3">
                  {/* Badge do Usuário Logado */}
                  <div className="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-neutral-900/90 border border-[#d4af37]/30 shadow-sm">
                    <div className="w-6 h-6 rounded-lg bg-[#d4af37] text-black font-bold text-[11px] flex items-center justify-center font-mono">
                      {getIniciais(usuario?.nome || usuario?.email || "US")}
                    </div>
                    <span className="text-xs font-medium text-neutral-200 hidden sm:inline">
                      Olá, <span className="font-semibold text-white">{primeiroNome}</span>
                    </span>
                  </div>

                  {/* Botão Sair */}
                  <button
                    onClick={logout}
                    title="Encerrar Sessão"
                    className="p-2 rounded-xl text-neutral-400 hover:text-rose-400 hover:bg-rose-500/10 border border-transparent hover:border-rose-500/20 transition-all"
                    aria-label="Sair da conta"
                  >
                    <LogOut className="w-4 h-4" />
                  </button>
                </div>
              ) : (
                <button
                  onClick={abrirLogin}
                  className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-gradient-to-r from-[#d4af37] to-[#b8962d] hover:from-[#e5bf43] hover:to-[#c4a031] text-black text-xs font-bold tracking-wide shadow-md shadow-[#d4af37]/15 transition-all duration-200 active:scale-95"
                >
                  <Lock className="w-3.5 h-3.5" />
                  <span>Entrar / Cadastrar</span>
                </button>
              )
            )}
          </div>
        </div>
      </header>

      {/* Modal de Autenticação */}
      <AuthModal
        isOpen={modalAuthAberto}
        onClose={() => setModalAuthAberto(false)}
        initialTab={tabInicial}
      />
    </>
  );
}
