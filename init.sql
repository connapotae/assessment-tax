CREATE TABLE IF NOT EXISTS tax_level (
	id serial PRIMARY KEY,
	level int NOT NULL,
	label varchar(20) NOT NULL,
	min_amount numeric NOT NULL,
	max_amount numeric NOT NULL,
	tax_percent int NOT NULL
);

INSERT INTO tax_level (level,label,min_amount,max_amount,tax_percent) VALUES
	 (1,'0-150,000',0,150000,0),
	 (2,'150,001-500,000',150000,500000,10),
	 (3,'500,001-1,000,000',500000,1000000,15),
	 (4,'1,000,001-2,000,000',1000000,2000000,20),
	 (5,'2,000,001 ขึ้นไป',2000000,'infinity'::numeric,35);

CREATE TABLE IF NOT EXISTS deduction (
	id serial PRIMARY KEY,
	deduct_type varchar(20) NOT NULL,
	deduct_amount numeric NOT NULL
);

INSERT INTO deduction (deduct_type,deduct_amount) VALUES
	('personal',60000),
	('donation',100000),
	('k-receipt',50000);