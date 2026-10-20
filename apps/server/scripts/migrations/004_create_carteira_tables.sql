-- apps/server/scripts/migrations/004_create_carteira_tables.sql
-- Tabela de Posições Consolidadas da Carteira do Usuário
CREATE TABLE IF NOT EXISTS carteira_posicoes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    usuario_id BIGINT NOT NULL,
    ticker VARCHAR(20) NOT NULL,
    classe ENUM('RENDA_FIXA', 'FII', 'ACAO_BR', 'INTERNACIONAL', 'CAIXA') NOT NULL,
    categoria_macro VARCHAR(50) NOT NULL DEFAULT 'GERAL',
    quantidade DECIMAL(18, 6) NOT NULL DEFAULT 0.000000,
    preco_medio DECIMAL(18, 4) NOT NULL DEFAULT 0.0000,
    valor_aplicado DECIMAL(18, 4) NOT NULL DEFAULT 0.0000,
    taxa_contratada VARCHAR(50) NULL,
    indexador VARCHAR(20) NULL,
    vencimento DATE NULL,
    corretora VARCHAR(50) NOT NULL DEFAULT 'Principal',
    observacoes TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    CONSTRAINT fk_posicoes_usuario FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    UNIQUE KEY uq_usuario_ticker_corretora (usuario_id, ticker, corretora),
    INDEX idx_carteira_usuario (usuario_id),
    INDEX idx_carteira_classe (usuario_id, classe)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabela de Livro-Razão de Movimentações da Carteira
CREATE TABLE IF NOT EXISTS carteira_transacoes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    usuario_id BIGINT NOT NULL,
    posicao_id BIGINT NULL,
    ticker VARCHAR(20) NOT NULL,
    tipo ENUM('COMPRA', 'VENDA', 'APORTE', 'RESGATE') NOT NULL,
    quantidade DECIMAL(18, 6) NOT NULL,
    preco_unitario DECIMAL(18, 4) NOT NULL,
    custos_taxas DECIMAL(10, 4) NOT NULL DEFAULT 0.0000,
    valor_total DECIMAL(18, 4) NOT NULL,
    data_operacao DATETIME NOT NULL,
    corretora VARCHAR(50) NOT NULL DEFAULT 'Principal',
    notas TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    CONSTRAINT fk_transacoes_usuario FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_transacoes_posicao FOREIGN KEY (posicao_id) REFERENCES carteira_posicoes(id) ON DELETE SET NULL,
    INDEX idx_transacoes_usuario (usuario_id),
    INDEX idx_transacoes_ticker (usuario_id, ticker),
    INDEX idx_transacoes_data (usuario_id, data_operacao)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
