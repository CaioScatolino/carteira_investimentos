-- apps/server/scripts/migrations/006_expand_ticker_columns.sql
-- Expande a coluna ticker para VARCHAR(60) em todas as tabelas de carteira e proventos
-- para suportar com folga Títulos Públicos (Tesouro Direto), CDBs e nomes descritivos de ativos

ALTER TABLE carteira_posicoes MODIFY COLUMN ticker VARCHAR(60) NOT NULL;
ALTER TABLE carteira_transacoes MODIFY COLUMN ticker VARCHAR(60) NOT NULL;
ALTER TABLE proventos_recebidos MODIFY COLUMN ticker VARCHAR(60) NOT NULL;
