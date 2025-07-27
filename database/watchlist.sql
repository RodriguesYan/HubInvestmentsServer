DROP TABLE IF EXISTS watchlist;

CREATE TABLE watchlist (
    id SERIAL PRIMARY KEY,
    user_id integer REFERENCES users,
    symbols varchar(1000) not null
);

INSERT INTO watchlist (id, user_id, symbols) 
VALUES 	(1, 1, 'AAPL,AMZN,DIS,VOO,VBR');

select * from watchlist;