# Step 01: Preparação do Ambiente Go e Fundamentos Iniciais

## 1. Visão Geral e Contexto Arquitetural
Estamos a construir uma plataforma de alta performance para auditoria de carteira na B3, cálculo de Preço Teto (Bazin, Graham) e recomendação inteligente via IA (Gemini Pro).

### Por que Golang para o Back-end?
- **Binário Estático Único:** Em vez de depender de um runtime pesado (como Node.js ou JVM), o compilador Go compila todo o código-fonte e suas dependências diretamente em código de máquina (binário executável).
- **Consumo Mínimo de Memória (< 40 MB em repouso):** Como a nossa VPS dispõe de 4 GB de RAM compartilhados com MySQL, Redis e Nginx Proxy Manager, cada megabyte economizado garante estabilidade e espaço para concorrência de goroutines.
- **Concorrência Nativa (Goroutines & Channels):** Threads gerenciadas pelo runtime do Go que pesam apenas ~2 KB cada (em contraste com threads de SO de ~1-2 MB), ideais para processamento simultâneo de relatórios da CVM e scraping de relatórios institucionais.
- **Tratamento Explícito de Erros:** Não existem `try/catch` nem exceções não capturadas derrubando o servidor em produção. Todo erro é um valor retornado explicitamente (`val, err := funcao()`).

---

## 2. Instalação do Compilador Go no Windows

### Opção A: Via Winget (Linha de Comandos no PowerShell)
Abra o PowerShell como Utilizador ou Administrador e execute:
```powershell
winget install GoLang.Go
```
*Após a instalação, reinicie o terminal PowerShell para que as variáveis de ambiente sejam recarregadas.*

### Opção B: Download Manual do Instalador Oficial
1. Aceda a [go.dev/dl](https://go.dev/dl/).
2. Descarregue a versão estável mais recente para Windows (ex: `go1.23.x.windows-amd64.msi` ou superior).
3. Execute o instalador mantendo os caminhos padrão (`C:\Program Files\Go`).

---

## 3. Variáveis de Ambiente Essenciais em Go
O Go trabalha com três caminhos fundamentais:
- **`GOROOT`**: Onde o compilador e a biblioteca padrão do Go estão instalados (ex: `C:\Program Files\Go`). Configurado automaticamente pelo instalador.
- **`GOPATH`**: Diretório padrão para o workspace de ferramentas e cache de pacotes (por defeito: `%USERPROFILE%\go`).
- **`PATH`**: Deve conter `C:\Program Files\Go\bin` e `%USERPROFILE%\go\bin` para permitir invocar comandos como `go` e ferramentas instaladas via terminal.

---

## 4. Validação da Instalação
No terminal, execute:
```powershell
go version
```
A saída esperada deve ser semelhante a:
```text
go version go1.23.x windows/amd64
```

Em seguida, confira as variáveis do ambiente com:
```powershell
go env
```

---

## 5. Estrutura Proposta para o Repositório
```text
carteira_investimentos/
├── apps/
│   ├── server/                   # Back-end em Golang (Binário Único)
│   │   ├── cmd/
│   │   │   └── api/              # Ponto de entrada (main.go)
│   │   ├── internal/             # Lógica de domínio privada da aplicação
│   │   │   ├── bazin/            # Lógica matemática e regras de Bazin
│   │   │   ├── graham/           # Lógica matemática e margem de segurança de Graham
│   │   │   ├── scraper/          # Scraping concorrente de carteiras recomendadas
│   │   │   ├── cvm/              # Ingestão e streaming de relatórios CVM
│   │   │   ├── ai/               # Integração com Gemini Pro (funil de 2 etapas)
│   │   │   ├── storage/          # Camada de banco (MySQL/sqlc) e cache (Redis)
│   │   │   └── sse/              # Handlers para Server-Sent Events nativo
│   │   ├── skills/               # Diretrizes analíticas e prompts modulares da IA
│   │   │   ├── bazin-method/
│   │   │   ├── graham-method/
│   │   │   └── consensus-filter/
│   │   ├── go.mod                # Gerenciador de dependências do módulo Go
│   │   └── Dockerfile            # Multi-stage build para imagem ultra-leve (< 25 MB)
│   │
│   └── web/                      # Front-end em Next.js (Standalone)
│       ├── src/
│       ├── Dockerfile
│       └── package.json
│
├── deploy/                       # Infraestrutura e orquestração na VPS
│   ├── docker-compose.yml        # Conexão com a rede dos containers centrais da VPS
│   └── scripts/                  # Scripts de deploy automatizado via SSH
│
└── docs/                         # Documentação viva do projeto e mentoria
    └── steps/                    # Histórico incremental passo a passo
        └── step-01-ambiente-go.md
```

---

## 6. Próximo Passo
Após a validação do `go version`, avançaremos para o **Step 02**:
- Inicialização do módulo Go (`go mod init`).
- Criação do primeiro "Hello World" com servidor HTTP nativo.
- Comparação do modelo mental: Event Loop do Node.js vs Go Runtime & Goroutines.
