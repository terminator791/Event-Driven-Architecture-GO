-- PostgreSQL initialization script
-- Create databases for the microservices

-- Create orderdb database
CREATE DATABASE orderdb OWNER postgres;

-- Create paymentdb database
CREATE DATABASE paymentdb OWNER postgres;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE orderdb TO postgres;
GRANT ALL PRIVILEGES ON DATABASE paymentdb TO postgres;
