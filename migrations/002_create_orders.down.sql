DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS order_type;
DROP TYPE IF EXISTS order_side;


-- why IF is in here ---> without if there is no table -> error,   
-- with if no error manage easly 