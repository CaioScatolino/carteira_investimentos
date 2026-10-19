# Passo 21: Autenticação Segura (Backend OWASP & Especificação de Frontend)

## 🎯 Objetivo Arquitetural
Construir uma camada completa de autenticação institucional para a plataforma de investimentos, permitindo o isolamento multi-tenant de carteiras com cibersegurança rigorosa (OWASP Top 10) no backend em Go e uma experiência de usuário (UX) fluida, sofisticada e minimalista no frontend em Next.js 16.

---

## 🛡️ 1. O que foi Implementado no Backend (Passo 7.1)

### 1.1 Modelagem Relacional Segura (`usuarios`)
- **Tabela MySQL / MariaDB (`003_create_usuarios_table.sql`):**
  - `id`: Chave primária numérica interna para alta performance de índices relacionais.
  - `uuid`: Identificador universal v4 criptográfico (`google/uuid`) exposto publicamente nas APIs, eliminando vulnerabilidades de enumeração de contas e IDOR (*Insecure Direct Object References*).
  - `email`: `VARCHAR(191)` com índice único, garantindo compatibilidade estrita com `utf8mb4`.
  - `senha_hash`: Armazenamento de hash irreversível com `json:"-"` na struct Go para impedir vazamento acidental em respostas JSON.
  - `tentativas_falhas` e `bloqueado_ate`: Campos de controle ativo contra ataques de força bruta.

### 1.2 Motor Criptográfico
- **Algoritmo de Hashing:** **bcrypt** com Fator de Custo 12 ($2^{12} = 4.096$ iterações) e salt dinâmico embutido de 128 bits.
- **Proteção contra *Timing Attacks*:** Comparação em tempo constante via `bcrypt.CompareHashAndPassword`.
- **Validação de Senha Forte (OWASP):**
  - Mínimo de 8 caracteres.
  - Ao menos uma letra maiúscula.
  - Ao menos uma letra minúscula.
  - Ao menos um dígito numérico.
  - Ao menos um caractere especial (`!@#$%...`).

### 1.3 Sessão Stateless com JWT
- **Padrão:** JSON Web Tokens (RFC 7519) assinados com **HMAC-SHA256**.
- **Claims Registradas:** `sub` (UUID público), `uid` (ID interno), `email`, `role`, `iss` ("b3-core-api"), `iat` e `exp`.
- **Middleware HTTP (`ExigirAutenticacao`):**
  - Intercepta o cabeçalho `Authorization: Bearer <token>`.
  - Valida assinatura e expiração.
  - Injeta os dados do usuário autenticado no `context.Context` da requisição (`r.Context()`), permitindo acesso imediato nas rotas subsequentes.

### 1.4 Endpoints Disponíveis
| Método | Endpoint | Acesso | Descrição |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Público | Cadastro com validação de senha forte |
| `POST` | `/api/v1/auth/login` | Público | Autenticação com proteção anti-brute force |
| `GET` | `/api/v1/auth/me` | Protegido | Retorna o perfil do usuário logado via JWT |

---

## 🎨 2. Especificação do Frontend de Autenticação (Passo 7.2)

### 2.1 Identidade Visual e UX Institucional
- **Paleta de Cores:**
  - Fundo Ônix Profundo: `#070709` / `#0c0c10`
  - Cards e Modais Grafite: `#131318`
  - Dourado Champagne (Acentos e Bordas): `#d4af37` / `#f5df88`
  - Alertas Semânticos: Verde Esmeralda (`#10b981`), Âmbar (`#f59e0b`) e Vermelho Rubi (`#ef4444`).
- **Padrão de Apresentação:**
  - **Modal Inteligente de Autenticação (`AuthModal.tsx`):** Abre sobre o dashboard com efeito *glassmorphism* (backdrop blur) e animação suave de entrada.
  - Alternância rápida com abas: **[Entrar]** e **[Criar Conta]**.

### 2.2 Medidor de Força de Senha em Tempo Real (Cadastro)
Durante a digitação da senha no formulário de registro, o usuário recebe feedback visual imediato com 4 checagens interativas:
- [ ] No mínimo 8 caracteres
- [ ] Letra maiúscula (A-Z)
- [ ] Letra minúscula (a-z)
- [ ] Número (0-9) e caractere especial (@, #, $, etc.)

*O botão "Cadastrar" permanece desabilitado com opacidade reduzida até que todos os 4 requisitos estejam verdes.*

### 2.3 Gerenciamento de Estado Global (`AuthContext.tsx` e `useAuth`)
```typescript
interface Usuario {
  uuid: string;
  nome: string;
  email: string;
  role: string;
}

interface AuthContextType {
  usuario: Usuario | null;
  token: string | null;
  carregando: boolean;
  login: (email: string, senha: string) => Promise<void>;
  registro: (nome: string, email: string, senha: string) => Promise<void>;
  logout: () => void;
}
```
- **Persistência de Sessão:** Armazena o token no `localStorage` e nos Cookies do navegador.
- **Ciclo de Inicialização:** Ao carregar a página, se houver token salvo, dispara `GET /api/v1/auth/me` em segundo plano para validar a sessão e hidratar os dados do usuário.

### 2.4 Integração com o Header do Dashboard
- **Quando Deslogado:** Exibe o botão dourado com ícone de cadeado: `[ 🔒 Entrar / Cadastrar ]`.
- **Quando Logado:** Substitui o botão por:
  - Avatar circular com as iniciais do usuário (ex: "CD").
  - Nome ("Olá, Caio").
  - Botão de atalho para a futura tela "💼 Minha Carteira".
  - Botão discreto de logout `[ Sair ]`.

### 2.5 Tratamento de Erros e Segurança Visual
- Exibe mensagens institucionais claras sem vazar informações sensíveis:
  - *"E-mail ou senha incorretos"* (Anti-enumeração).
  - *"Conta temporariamente bloqueada por excesso de tentativas"* (Anti-brute force com tempo restante).
  - *"Este e-mail já está cadastrado na plataforma"*.
