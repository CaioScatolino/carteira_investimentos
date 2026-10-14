-- Migração 003: Tabela de Usuários com Protocolos de Cibersegurança
-- Armazena credenciais criptografadas, UUIDs públicos e controle de força bruta.

CREATE TABLE IF NOT EXISTS usuarios (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE COMMENT 'Identificador público seguro para APIs e tokens JWT',
    nome VARCHAR(100) NOT NULL COMMENT 'Nome completo ou de exibição do usuário',
    email VARCHAR(191) NOT NULL UNIQUE COMMENT 'E-mail para autenticação e notificações',
    senha_hash VARCHAR(255) NOT NULL COMMENT 'Hash criptográfico resistente (bcrypt / Argon2id)',
    role VARCHAR(20) NOT NULL DEFAULT 'USER' COMMENT 'Papel de acesso: USER ou ADMIN',
    status VARCHAR(20) NOT NULL DEFAULT 'ATIVO' COMMENT 'Status da conta: ATIVO, BLOQUEADO ou PENDENTE',
    tentativas_falhas INT NOT NULL DEFAULT 0 COMMENT 'Contador de falhas consecutivas de login',
    bloqueado_ate DATETIME NULL COMMENT 'Timestamp de expiração do bloqueio por brute-force',
    ultimo_login_at DATETIME NULL COMMENT 'Data e hora do último acesso com sucesso',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP(),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_usuarios_uuid (uuid),
    UNIQUE KEY uk_usuarios_email (email),
    INDEX idx_usuarios_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
