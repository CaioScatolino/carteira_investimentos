#!/bin/bash
set -e

echo "=========================================================="
echo "🚀 INICIANDO DEPLOY AUTOMATIZADO: B3 INVESTIMENTOS PLATFORM"
echo "=========================================================="

# 1. Carrega variáveis de ambiente de produção
if [ -f .env.prod ]; then
    set -a
    . ./.env.prod
    set +a
else
    echo "⚠️  Aviso: Arquivo .env.prod não encontrado! Usando variáveis do sistema."
fi

# 2. Garante que o banco de dados carteira_investimentos existe no MySQL Central
echo "📦 Verificando/Criando schema no mysql_central..."
docker exec -i mysql_central mysql -u root -p"${DB_PASSWORD}" -e "CREATE DATABASE IF NOT EXISTS carteira_investimentos CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 3. Executa as migrações SQL em ordem
echo "🔄 Executando Migrations SQL..."
for f in apps/server/scripts/migrations/*.sql; do
    [ -f "$f" ] || continue
    echo "   • Aplicando migração: $f"
    docker exec -i mysql_central mysql -u root -p"${DB_PASSWORD}" carteira_investimentos < "$f" || true
done

# 4. Detecta automaticamente o nome da rede Docker compartilhada
echo "🌐 Detectando rede Docker da infra-compartilhada..."
if [ -z "$DOCKER_NETWORK" ]; then
    DETECTED_NETWORK=$(docker inspect mysql_central --format '{{range $k, $v := .NetworkSettings.Networks}}{{println $k}}{{end}}' 2>/dev/null | grep -v '^bridge$' | head -n 1 || echo "")
    if [ -n "$DETECTED_NETWORK" ]; then
        echo "   • Rede detectada no mysql_central: $DETECTED_NETWORK"
        export DOCKER_NETWORK="$DETECTED_NETWORK"
    else
        export DOCKER_NETWORK="rede-compartilhada"
        echo "   • Usando rede fallback: $DOCKER_NETWORK"
    fi
else
    echo "   • Rede customizada via .env.prod: $DOCKER_NETWORK"
fi

# 5. Build e subida dos containers com Docker Compose
echo "🐳 Construindo e subindo containers..."
if [ -f .env.prod ]; then
    docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build --remove-orphans
else
    docker compose -f docker-compose.prod.yml up -d --build --remove-orphans
fi

# 6. Limpeza de imagens antigas não utilizadas (economiza disco na VPS)
docker image prune -f

echo "=========================================================="
echo "✅ DEPLOY CONCLUÍDO COM SUCESSO!"
echo "=========================================================="
