"use client";

import React from "react";
import { Check, X } from "lucide-react";

interface PasswordStrengthProps {
  senha: string;
}

export function PasswordStrength({ senha }: PasswordStrengthProps) {
  const temTamanhoMinimo = senha.length >= 8;
  const temMaiuscula = /[A-Z]/.test(senha);
  const temMinuscula = /[a-z]/.test(senha);
  const temNumero = /[0-9]/.test(senha);
  const temEspecial = /[^A-Za-z0-9]/.test(senha);

  const requisitos = [
    { label: "Mínimo 8 caracteres", valido: temTamanhoMinimo },
    { label: "Uma letra maiúscula (A-Z)", valido: temMaiuscula },
    { label: "Uma letra minúscula (a-z)", valido: temMinuscula },
    { label: "Um número (0-9) e caractere especial (!@#$)", valido: temNumero && temEspecial },
  ];

  const totalValidos = requisitos.filter((r) => r.valido).length;

  const getCorBarra = () => {
    if (totalValidos === 4) return "bg-emerald-400";
    if (totalValidos >= 2) return "bg-amber-400";
    return "bg-rose-500";
  };

  const getTextoForca = () => {
    if (senha.length === 0) return "";
    if (totalValidos === 4) return "Forte (Padrão OWASP)";
    if (totalValidos >= 2) return "Média";
    return "Fraca";
  };

  if (senha.length === 0) {
    return null;
  }

  return (
    <div className="space-y-2 mt-2 pt-1 border-t border-neutral-800/60">
      {/* Barra de Progresso */}
      <div className="flex items-center justify-between text-[11px]">
        <span className="text-neutral-400 font-medium">Força da Senha:</span>
        <span
          className={`font-semibold ${
            totalValidos === 4
              ? "text-emerald-400"
              : totalValidos >= 2
              ? "text-amber-400"
              : "text-rose-400"
          }`}
        >
          {getTextoForca()}
        </span>
      </div>

      <div className="grid grid-cols-4 gap-1.5 h-1">
        {[1, 2, 3, 4].map((passo) => (
          <div
            key={passo}
            className={`h-full rounded-full transition-all duration-300 ${
              totalValidos >= passo ? getCorBarra() : "bg-neutral-800"
            }`}
          />
        ))}
      </div>

      {/* Checklist Interativo */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5 pt-1">
        {requisitos.map((r, idx) => (
          <div key={idx} className="flex items-center gap-1.5 text-[11px]">
            {r.valido ? (
              <Check className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
            ) : (
              <X className="w-3.5 h-3.5 text-neutral-600 shrink-0" />
            )}
            <span className={r.valido ? "text-neutral-300" : "text-neutral-500"}>
              {r.label}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
