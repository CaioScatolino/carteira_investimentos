-- apps/server/scripts/migrations/005_create_proventos_tables.sql
-- Tabela de Histórico de Proventos e Renda Passiva Recebida
CREATE TABLE IF NOT EXISTS proventos_recebidos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    usuario_id BIGINT NOT NULL,
    ticker VARCHAR(20) NOT NULL,
    tipo ENUM('RENDIMENTO', 'DIVIDENDO', 'JCP', 'JUROS', 'RESTITUICAO_CAPITAL') NOT NULL,
    data_pagamento DATE NOT NULL,
    quantidade DECIMAL(18, 6) NOT NULL DEFAULT 0.000000,
    valor_unitario DECIMAL(18, 4) NOT NULL DEFAULT 0.0000,
    valor_liquido DECIMAL(18, 4) NOT NULL DEFAULT 0.0000,
    corretora VARCHAR(100) NOT NULL DEFAULT 'Principal',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    CONSTRAINT fk_proventos_usuario FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    UNIQUE KEY uq_usuario_provento_mov (usuario_id, ticker, data_pagamento, tipo, valor_liquido),
    INDEX idx_proventos_usuario (usuario_id),
    INDEX idx_proventos_data (usuario_id, data_pagamento),
    INDEX idx_proventos_ticker (usuario_id, ticker),
    INDEX idx_proventos_tipo (usuario_id, tipo)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
