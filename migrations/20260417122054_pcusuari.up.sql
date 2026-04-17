CREATE TABLE WINTHOR.PCUSUARI (
    CODUSUR NUMBER PRIMARY KEY,          -- Código do usuário
    NOME VARCHAR2(100),                  -- Nome completo
    ATIVO CHAR(1),                       -- Se está ativo
    EMAIL VARCHAR2(100)                 -- Email
);
