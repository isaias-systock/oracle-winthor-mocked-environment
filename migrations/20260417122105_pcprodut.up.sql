CREATE TABLE WINTHOR.PCPRODUT (
    CODPROD NUMBER PRIMARY KEY,          -- Código do produto
    DESCRICAO VARCHAR2(100),             -- Descrição do produto
    CODEPTO NUMBER,                      -- Código do departamento
    CODSEC NUMBER,                       -- Código da seção
    PESOBRUTO NUMBER(10,3),              -- Peso bruto
    PESOLIQ NUMBER(10,3),                -- Peso líquido
    QTUNITCX NUMBER(10,3),               -- Unidades por caixa
    EMBALAGEM VARCHAR2(20),              -- Tipo de embalagem
    CUSTOREP NUMBER(10,4),               -- Custo de reposição
    CODAUXILIAR NUMBER(20,0)             -- Código Auxiliar do produto
)
