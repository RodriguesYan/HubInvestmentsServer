DROP TABLE IF EXISTS balances;

CREATE TABLE balances (
    id SERIAL PRIMARY KEY,
    user_id integer REFERENCES users,
    available_balance DECIMAL(10, 2) NOT NULL
);

INSERT INTO balances (id, user_id, available_balance) VALUES (1, 1, 10000);

select * from balances;