CREATE TABLE WINTHOR.PCPEDI (
    NUMPED NUMBER,                       -- Número do pedido (FK)
    NUMITEM NUMBER,                      -- Número do item
    CODPROD NUMBER,                      -- Código do produto
    QT NUMBER(10,3),                     -- Quantidade
    PVENDA NUMBER(10,4),                 -- Preço de venda
    VLCUSTOFIN NUMBER(10,4),             -- Custo financeiro
    BONIFIC CHAR(1)                     -- Se é bonificação
);
