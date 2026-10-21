#!/bin/bash
set -e

echo "=========================================================="
echo "🚀 INICIANDO DEPLOY AUTOMATIZADO: B3 INVESTIMENTOS PLATFORM"
echo "=========================================================="

# 1. Carrega variáveis de ambiente de produção
if [ -f .env.prod ]; then
    export $(cat .env.prod | xargs)
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

# 4. Build e subida dos containers com Docker Compose
echo "🐳 Construindo e subindo containers..."
docker compose -f docker-compose.prod.yml up -d --build --remove-orphans

# 5. Limpeza de imagens antigas não utilizadas (economiza disco na VPS)
docker image prune -f

echo "=========================================================="
echo "✅ DEPLOY CONCLUÍDO COM SUCESSO!"
echo "=========================================================="
