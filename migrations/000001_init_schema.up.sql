CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,         
    login TEXT UNIQUE NOT NULL,       
    password TEXT NOT NULL            
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,         
    user_id BIGINT REFERENCES users(id), 
    number TEXT UNIQUE NOT NULL,      
    status TEXT NOT NULL,             
    accrual DOUBLE PRECISION,         
    uploaded_at TIMESTAMP NOT NULL    
);

CREATE TABLE balance (
    user_id BIGINT UNIQUE REFERENCES users(id), 
    current DOUBLE PRECISION DEFAULT 0, 
    withdrawn DOUBLE PRECISION DEFAULT 0 
);

CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,         
    user_id BIGINT REFERENCES users(id), 
    order_number TEXT NOT NULL,       
    sum DOUBLE PRECISION NOT NULL,    
    processed_at TIMESTAMP NOT NULL   
);