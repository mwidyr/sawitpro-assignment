-- This is the SQL script that will be used to initialize the database schema.
-- We will evaluate you based on how well you design your database.
-- 1. How you design the tables.
-- 2. How you choose the data types and keys.
-- 3. How you name the fields.
-- In this assignment we will use PostgreSQL as the database.

-- This is test table. Remove this table and replace with your own tables. 
CREATE TABLE test (
	id serial PRIMARY KEY,
	name VARCHAR ( 50 ) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS estates (
    estate_id UUID PRIMARY KEY,
    estate_length INT NOT NULL CHECK (estate_length BETWEEN 1 AND 50000),
    estate_width  INT NOT NULL CHECK (estate_width BETWEEN 1 AND 50000)
);

CREATE TABLE IF NOT EXISTS trees (
    tree_id UUID PRIMARY KEY,
    estate_id UUID NOT NULL REFERENCES estates(estate_id) ON DELETE CASCADE,
    tree_height INT NOT NULL CHECK (tree_height BETWEEN 1 AND 30),
    coordinate_x INT NOT NULL,
    coordinate_y INT NOT NULL,
    UNIQUE (estate_id, coordinate_x, coordinate_y)
);

CREATE INDEX idx_trees_estate_id ON trees(estate_id);


