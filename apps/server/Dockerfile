# ==============================================================================
# ESTÁGIO 1: Compilação Estática em Go
# ==============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Instala certificados SSL e timezone
RUN apk add --no-cache ca-certificates tzdata

# Cache de dependências do Go
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte do back-end
COPY . .

# Compilação estática com remoção de símbolos de debug (-s -w) para peso mínimo
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/b3-api ./cmd/api/main.go

# ==============================================================================
# ESTÁGIO 2: Imagem Final de Execução Ultra-Enxuta (< 25 MB)
# ==============================================================================
FROM alpine:3.20

WORKDIR /app

# Copia certificados SSL de raiz e base de fuso horário oficial do Brasil
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=America/Sao_Paulo

# Cria usuário não-privilegiado (segurança de container)
RUN adduser -D -g '' appuser

# Copia o binário compilado
COPY --from=builder /app/bin/b3-api /app/b3-api

# Copia migrations SQL para automação de schema se necessário
COPY --from=builder /app/scripts/migrations /app/scripts/migrations

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/b3-api"]
