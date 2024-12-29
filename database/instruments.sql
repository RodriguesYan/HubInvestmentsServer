DROP TABLE IF EXISTS instruments;

CREATE TABLE instruments (
    id SERIAL PRIMARY KEY,
    symbol varchar(50) not null,
    name varchar(50) not null,
    category integer not null,
    last_price decimal not null
);

INSERT INTO instruments (id, symbol, name, category, last_price) 
VALUES (2, 'AMZN', 'Amazon prime', 1, 140.5)
VALUES (3, 'DIS', 'Disneylandia', 1, 240.5)
VALUES (4, 'VOO', 'Vanguard SP 500', 2, 240.5);

select * from instruments;