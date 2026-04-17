CREATE TABLE WINTHOR.PCPEDC (
    NUMPED NUMBER PRIMARY KEY,           -- Número do pedido
    CODCLI NUMBER,                       -- Código do cliente
    CODUSUR NUMBER,                      -- Código do usuário/vendedor
    DATA DATE,                           -- Data do pedido
    POSICAO CHAR(1),                    -- Status do pedido
    CODFILIAL CHAR(2),                  -- Código da filial
    CONDVENDA NUMBER,                    -- Condição de venda
    DTCANCEL DATE,                       -- Data de cancelamento
    VLTOTAL NUMBER(10,2)                -- Valor total do pedido
);
