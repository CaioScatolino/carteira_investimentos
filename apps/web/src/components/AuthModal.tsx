"use client";

import React, { useState, useEffect } from "react";
import { X, Lock, Mail, User, Eye, EyeOff, Loader2, ShieldCheck, AlertCircle } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { PasswordStrength } from "./PasswordStrength";

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialTab?: "login" | "register";
}

export function AuthModal({ isOpen, onClose, initialTab = "login" }: AuthModalProps) {
  const { login, register } = useAuth();
  const [tab, setTab] = useState<"login" | "register">(initialTab);

  // Form states
  const [nome, setNome] = useState("");
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [mostrarSenha, setMostrarSenha] = useState(false);

  // Status states
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);

  useEffect(() => {
    setTab(initialTab);
    setErro(null);
  }, [initialTab, isOpen]);

  // Tecla Escape para fechar o modal
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) {
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setCarregando(true);
    setErro(null);

    try {
      if (tab === "login") {
        await login(email, senha);
      } else {
        await register(nome, email, senha);
      }
      onClose();
    } catch (err: any) {
      setErro(err.message || "Ocorreu um erro ao processar sua solicitação");
    } finally {
      setCarregando(false);
    }
  };

  const senhaValida =
    senha.length >= 8 &&
    /[A-Z]/.test(senha) &&
    /[a-z]/.test(senha) &&
    /[0-9]/.test(senha) &&
    /[^A-Za-z0-9]/.test(senha);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md transition-opacity duration-200"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-md bg-[#0c0c10] border border-[#d4af37]/30 rounded-2xl p-6 sm:p-8 shadow-2xl shadow-black/80 transition-all duration-300 transform scale-100"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-titulo"
      >
        {/* Botão Fechar */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 text-neutral-400 hover:text-white rounded-lg hover:bg-neutral-800/60 transition-colors"
          aria-label="Fechar modal"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Topo / Logo */}
        <div className="flex flex-col items-center text-center mb-6">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-[#d4af37] via-[#917726] to-[#0a0a0d] p-[1px] shadow-lg mb-3">
            <div className="w-full h-full rounded-xl bg-[#0b0b0f] flex items-center justify-center">
              <ShieldCheck className="w-6 h-6 text-[#d4af37]" />
            </div>
          </div>
          <h2 id="modal-titulo" className="text-xl font-bold tracking-tight text-white font-mono">
            B3<span className="text-[#d4af37]">CORE</span> Terminal
          </h2>
          <p className="text-xs text-neutral-400 mt-1">
            Plataforma Institucional de Inteligência & Carteiras
          </p>
        </div>

        {/* Seletor de Abas (Entrar / Criar Conta) */}
        <div className="grid grid-cols-2 p-1 bg-neutral-900/90 border border-neutral-800 rounded-xl mb-6">
          <button
            type="button"
            onClick={() => {
              setTab("login");
              setErro(null);
            }}
            className={`py-2 text-xs font-semibold rounded-lg transition-all ${
              tab === "login"
                ? "bg-[#d4af37] text-black shadow-md font-bold"
                : "text-neutral-400 hover:text-white"
            }`}
          >
            Acessar Conta
          </button>
          <button
            type="button"
            onClick={() => {
              setTab("register");
              setErro(null);
            }}
            className={`py-2 text-xs font-semibold rounded-lg transition-all ${
              tab === "register"
                ? "bg-[#d4af37] text-black shadow-md font-bold"
                : "text-neutral-400 hover:text-white"
            }`}
          >
            Criar Nova Conta
          </button>
        </div>

        {/* Mensagem de Erro com Alerta */}
        {erro && (
          <div className="flex items-start gap-2.5 p-3.5 mb-5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs animate-shake">
            <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
            <span className="leading-relaxed">{erro}</span>
          </div>
        )}

        {/* Formulário */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {tab === "register" && (
            <div>
              <label className="block text-xs font-medium text-neutral-300 mb-1.5" htmlFor="nome">
                Nome Completo
              </label>
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-neutral-500">
                  <User className="w-4 h-4" />
                </div>
                <input
                  id="nome"
                  type="text"
                  required
                  autoComplete="name"
                  value={nome}
                  onChange={(e) => setNome(e.target.value)}
                  placeholder="Seu nome"
                  className="w-full pl-10 pr-4 py-2.5 bg-neutral-900/90 border border-neutral-800 rounded-xl text-sm text-white placeholder-neutral-500 focus:outline-none focus:border-[#d4af37] focus:ring-1 focus:ring-[#d4af37] transition-all"
                />
              </div>
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-neutral-300 mb-1.5" htmlFor="email">
              E-mail Institucional
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-neutral-500">
                <Mail className="w-4 h-4" />
              </div>
              <input
                id="email"
                type="email"
                required
                autoComplete="username"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="seu.email@exemplo.com"
                className="w-full pl-10 pr-4 py-2.5 bg-neutral-900/90 border border-neutral-800 rounded-xl text-sm text-white placeholder-neutral-500 focus:outline-none focus:border-[#d4af37] focus:ring-1 focus:ring-[#d4af37] transition-all"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-neutral-300 mb-1.5" htmlFor="senha">
              Senha de Acesso
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-neutral-500">
                <Lock className="w-4 h-4" />
              </div>
              <input
                id="senha"
                type={mostrarSenha ? "text" : "password"}
                required
                autoComplete={tab === "register" ? "new-password" : "current-password"}
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
                placeholder="••••••••••••"
                className="w-full pl-10 pr-11 py-2.5 bg-neutral-900/90 border border-neutral-800 rounded-xl text-sm text-white placeholder-neutral-500 focus:outline-none focus:border-[#d4af37] focus:ring-1 focus:ring-[#d4af37] transition-all"
              />
              <button
                type="button"
                onClick={() => setMostrarSenha(!mostrarSenha)}
                className="absolute inset-y-0 right-0 pr-3.5 flex items-center text-neutral-500 hover:text-neutral-300 transition-colors"
                aria-label={mostrarSenha ? "Ocultar senha" : "Exibir senha"}
              >
                {mostrarSenha ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>

            {/* Medidor visual de força para novo cadastro */}
            {tab === "register" && <PasswordStrength senha={senha} />}
          </div>

          {/* Botão de Submissão */}
          <button
            type="submit"
            disabled={carregando || (tab === "register" && !senhaValida)}
            className="w-full mt-2 py-3 px-4 rounded-xl bg-gradient-to-r from-[#d4af37] to-[#b39129] hover:from-[#e5bf43] hover:to-[#c4a031] text-black font-semibold text-sm tracking-wide shadow-lg shadow-[#d4af37]/20 transition-all duration-200 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {carregando ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>Processando...</span>
              </>
            ) : tab === "login" ? (
              "Acessar Terminal"
            ) : (
              "Criar Conta Segura"
            )}
          </button>
        </form>

        {/* Rodapé / Nota de Cibersegurança */}
        <div className="mt-6 pt-4 border-t border-neutral-800/60 text-center">
          <p className="text-[11px] text-neutral-500 flex items-center justify-center gap-1.5">
            <ShieldCheck className="w-3.5 h-3.5 text-[#d4af37]" />
            Criptografia de ponta a ponta com bcrypt & JWT Stateless
          </p>
        </div>
      </div>
    </div>
  );
}
