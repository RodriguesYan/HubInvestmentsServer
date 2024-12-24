DROP TABLE IF EXISTS aucAggregations;

CREATE TABLE aucAggregations (
    id SERIAL PRIMARY KEY,
    user_id integer REFERENCES users,
    available_balance DECIMAL(10, 2) NOT NULL,
    fixed_income_invested DECIMAL(10, 2) NOT NULL,
    stocks_invested DECIMAL(10, 2) NOT NULL,
    etfs_invested DECIMAL(10, 2) NOT NULL
);

INSERT INTO aucAggregations (id, user_id, available_balance, fixed_income_invested, stocks_invested, etfs_invested) 
VALUES (1, 1, 100, 200, 300, 400);

select * from aucAggregations;