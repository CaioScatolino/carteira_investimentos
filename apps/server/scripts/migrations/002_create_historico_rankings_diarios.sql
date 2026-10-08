-- Migração 002: Tabela de Histórico Diário de Rankings Pós-Pregão (Cron 19h)
-- Armazena o snapshot fundamentalista congelado ao final de cada dia de negociação.

CREATE TABLE IF NOT EXISTS historico_rankings_diarios (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    data_pregao DATE NOT NULL COMMENT 'Data de fechamento do pregão (YYYY-MM-DD)',
    ticker VARCHAR(12) NOT NULL COMMENT 'Código do ativo (ex: PETR4, HGLG11)',
    classe VARCHAR(10) NOT NULL COMMENT 'ACAO, FII ou ETF',
    posicao_ranking INT NOT NULL COMMENT 'Posição no ranking geral da sua classe (1º, 2º, etc.)',
    score DECIMAL(5,2) NOT NULL COMMENT 'Score fundamentalista composto (0 a 100)',
    preco_atual DECIMAL(10,2) NOT NULL COMMENT 'Preço oficial de fechamento no pregão',
    preco_teto_consolidado DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT 'Preço teto pelo consenso multi-modelo',
    premio_desconto_percentual DECIMAL(6,2) NOT NULL DEFAULT 0.00 COMMENT '% de upside/desconto em relação ao teto',
    dy DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'Dividend Yield nos últimos 12 meses (%)',
    pl DECIMAL(8,2) NULL COMMENT 'Preço sobre Lucro (P/L)',
    pvp DECIMAL(8,2) NULL COMMENT 'Preço sobre Valor Patrimonial (P/VP)',
    status VARCHAR(25) NOT NULL COMMENT 'COMPRAR_MAIS, MANTER ou ALERTA',
    is_provento_atipico TINYINT(1) DEFAULT 0 COMMENT '1 se provento for não-recorrente/yield trap',
    modelos_json TEXT NULL COMMENT 'JSON com detalhes dos tetos de cada motor (Bazin, Graham, DCF, etc)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_pregao_ticker (data_pregao, ticker),
    INDEX idx_ticker_pregao (ticker, data_pregao),
    INDEX idx_classe_pregao_posicao (classe, data_pregao, posicao_ranking)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
