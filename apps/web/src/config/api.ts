// apps/web/src/config/api.ts
// Obtém a URL base da API a partir de variável de ambiente com fallback para desenvolvimento local
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "http://localhost:8080";

export const API_V1_URL = `${API_BASE_URL}/api/v1`;
