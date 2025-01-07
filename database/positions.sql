DROP TABLE IF EXISTS positions;

CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    user_id integer REFERENCES users,
    instrument_id integer references instruments,
    quantity decimal NOT null,
    average_price decimal not null
);

INSERT INTO positions (id, user_id, instrument_id, quantity, average_price) 
VALUES 	(4, 1, 5, 5, 350.5),
		(1, 1, 1, 5, 150.5),
		(2, 1, 2, 6.7, 250.5),
		(3, 1, 4, 16.7, 350.5);

select * from positions;
select * from instruments;

SELECT 	  
		sum(p.average_price * p.quantity) as totalInvested,
		sum(p.quantity * i.last_price) as currentTotal,
		(sum(p.average_price * p.quantity) - sum(p.quantity * i.last_price)) as pnl
FROM positions p
JOIN instruments i ON p.instrument_id = i.id
WHERE i.category = 1

update positions 
set average_price = 90
where id = 4


WITH Calculations AS (
    SELECT 
        p.average_price * p.quantity AS total_invested,
        i.last_price * p.quantity AS current_total
    FROM positions p
	JOIN instruments i ON p.instrument_id = i.id
	WHERE i.category = 2
)
SELECT 
    SUM(total_invested) AS total_invested,
    SUM(current_total) AS current_total,
    SUM(current_total - total_invested) AS pnl
FROM 
    Calculations;











