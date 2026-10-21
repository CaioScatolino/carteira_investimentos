"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from "react";
import { UsuarioPublico, RespostaAuth } from "@/types/auth";

interface AuthContextType {
  usuario: UsuarioPublico | null;
  token: string | null;
  loading: boolean;
  login: (email: string, senha: string) => Promise<void>;
  register: (nome: string, email: string, senha: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

import { API_V1_URL } from "@/config/api";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [usuario, setUsuario] = useState<UsuarioPublico | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // Inicialização: Hidrata a sessão se houver token salvo
  useEffect(() => {
    const carregarSessao = async () => {
      try {
        const tokenSalvo = localStorage.getItem("b3_token");
        if (!tokenSalvo) {
          setLoading(false);
          return;
        }

        setToken(tokenSalvo);

        // Valida o token no backend chamando /api/v1/auth/me
        const res = await fetch(`${API_V1_URL}/auth/me`, {
          headers: {
            Authorization: `Bearer ${tokenSalvo}`,
            Accept: "application/json",
          },
        });

        if (res.ok) {
          const dados: UsuarioPublico = await res.json();
          setUsuario(dados);
        } else {
          // Token expirado ou inválido
          localStorage.removeItem("b3_token");
          setToken(null);
          setUsuario(null);
        }
      } catch (err) {
        console.warn("Falha ao validar sessão prévia:", err);
      } finally {
        setLoading(false);
      }
    };

    carregarSessao();
  }, []);

  const login = async (email: string, senha: string) => {
    const res = await fetch(`${API_V1_URL}/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, senha }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.erro || "Falha ao realizar login");
    }

    const authData: RespostaAuth = data;
    setToken(authData.token);
    setUsuario(authData.usuario);
    localStorage.setItem("b3_token", authData.token);
  };

  const register = async (nome: string, email: string, senha: string) => {
    const res = await fetch(`${API_V1_URL}/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nome, email, senha }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.erro || "Falha ao registrar conta");
    }

    const authData: RespostaAuth = data;
    setToken(authData.token);
    setUsuario(authData.usuario);
    localStorage.setItem("b3_token", authData.token);
  };

  const logout = () => {
    setToken(null);
    setUsuario(null);
    localStorage.removeItem("b3_token");
  };

  return (
    <AuthContext.Provider value={{ usuario, token, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth deve ser utilizado dentro de um AuthProvider");
  }
  return context;
}
