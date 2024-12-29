DROP TABLE IF EXISTS positions;

CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    user_id integer REFERENCES users,
    instrument_id integer references instruments,
    quantity decimal NOT null,
    average_price decimal not null
);

INSERT INTO positions (id, user_id, quantity, average_price) 
VALUES (1, 1, 1, 5, 150.5);

select * from aucAggregations;