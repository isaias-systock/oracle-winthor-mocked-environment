CREATE TABLE WINTHOR.PCCLIENT (
    CODCLI NUMBER PRIMARY KEY,           -- Código do cliente
    CLIENTE VARCHAR2(100),               -- Nome/razão social
    CGC VARCHAR2(18),                    -- CNPJ/CPF
    ENDERECO VARCHAR2(100),              -- Endereço
    BAIRRO VARCHAR2(50),                 -- Bairro
    CIDADE VARCHAR2(50),                 -- Cidade
    ESTADO CHAR(2),                      -- Estado
    CEP VARCHAR2(10),                    -- CEP
	CODCOB VARCHAR2(4)                   -- Cobrança
)