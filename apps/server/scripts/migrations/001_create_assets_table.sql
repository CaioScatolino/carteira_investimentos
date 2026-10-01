-- Migração 001: Criação da Tabela de Catálogo de Ativos da B3 (Ticker <-> CNPJ)

CREATE TABLE IF NOT EXISTS assets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ticker VARCHAR(10) NOT NULL UNIQUE,
    cnpj VARCHAR(18) NOT NULL,
    classe ENUM('ACAO', 'FII', 'ETF', 'BDR') NOT NULL,
    tipo VARCHAR(10) NULL COMMENT 'ON, PN, UNT para Ações; CI para FIIs',
    razao_social VARCHAR(255) NOT NULL,
    codigo_cvm VARCHAR(10) NULL,
    setor VARCHAR(100) NULL,
    ativo TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_assets_cnpj (cnpj),
    INDEX idx_assets_classe (classe)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Carga Inicial (Seed) com os ativos de maior liquidez da B3 (IBOV + IFIX)
INSERT INTO assets (ticker, cnpj, classe, tipo, razao_social, codigo_cvm, setor) VALUES
-- Ações Blue Chips e Dividendos
('PETR4', '33.000.167/0001-01', 'ACAO', 'PN',  'PETROLEO BRASILEIRO S.A. PETROBRAS', '009512', 'Petróleo e Gás'),
('VALE3', '33.592.510/0001-54', 'ACAO', 'ON',  'VALE S.A.', '004170', 'Mineração e Siderurgia'),
('BBAS3', '00.000.000/0001-91', 'ACAO', 'ON',  'BANCO DO BRASIL S.A.', '001023', 'Financeiro e Bancos'),
('ITUB4', '60.872.534/0001-23', 'ACAO', 'PN',  'ITAU UNIBANCO HOLDING S.A.', '019348', 'Financeiro e Bancos'),
('BBDC4', '60.746.948/0001-12', 'ACAO', 'PN',  'BANCO BRADESCO S.A.', '000906', 'Financeiro e Bancos'),
('WEGE3', '84.429.695/0001-11', 'ACAO', 'ON',  'WEG S.A.', '005410', 'Bens Industriais'),
('TAEE11','07.859.971/0001-30', 'ACAO', 'UNT', 'TRANSMISSORA ALIANCA DE ENERGIA ELETRICA S.A.', '020257', 'Energia Elétrica'),
('CPLE3', '04.368.898/0001-06', 'ACAO', 'ON',  'COMPANHIA PARANAENSE DE ENERGIA - COPEL', '014311', 'Energia Elétrica'),
('CSAN3', '50.746.577/0001-15', 'ACAO', 'ON',  'COSAN S.A.', '019836', 'Petróleo, Gás e Biocombustíveis'),
('RENT3', '16.670.085/0001-55', 'ACAO', 'ON',  'LOCALIZA RENT A CAR S.A.', '013064', 'Consumo Cíclico e Aluguel de Carros'),

-- Fundos Imobiliários mais negociados (IFIX)
('MXRF11','13.318.527/0001-79', 'FII',  'CI',  'MAXI RENDA FUNDO DE INVESTIMENTO IMOBILIARIO', NULL, 'Híbrido e Papel'),
('HGLG11','11.728.688/0001-47', 'FII',  'CI',  'CSHG LOGISTICA - FUNDO DE INVESTIMENTO IMOBILIARIO', NULL, 'Logística'),
('XPML11','28.750.990/0001-74', 'FII',  'CI',  'XP MALLS FII', NULL, 'Shoppings'),
('KNRI11','12.005.956/0001-65', 'FII',  'CI',  'KINEA RENDA IMOBILIARIA FII', NULL, 'Misto e Lajes Corporativas'),
('BTLG11','13.111.782/0001-84', 'FII',  'CI',  'BTG PACTUAL LOGISTICA FII', NULL, 'Logística'),
('VISC11','17.554.274/0001-25', 'FII',  'CI',  'VINCI SHOPPING CENTERS FII', NULL, 'Shoppings'),
('XPLG11','26.502.794/0001-85', 'FII',  'CI',  'XP LOG FII', NULL, 'Logística')
ON DUPLICATE KEY UPDATE 
    cnpj = VALUES(cnpj),
    classe = VALUES(classe),
    tipo = VALUES(tipo),
    razao_social = VALUES(razao_social),
    codigo_cvm = VALUES(codigo_cvm),
    setor = VALUES(setor);
