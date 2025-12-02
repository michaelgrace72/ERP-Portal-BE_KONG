-- ====================================================
-- INITIALIZE MULTIPLE DATABASES IN POSTGRESQL
-- ====================================================
-- This script creates separate databases for:
-- - Portal Service (portal_db)
-- - Kong Gateway (kong_db)
-- - Konga Admin UI (konga_db)

-- Create Portal Database
CREATE DATABASE portal_db;
GRANT ALL PRIVILEGES ON DATABASE portal_db TO erp_user;

-- Create Kong Database
CREATE DATABASE kong_db;
GRANT ALL PRIVILEGES ON DATABASE kong_db TO erp_user;

-- Create Konga Database
CREATE DATABASE konga_db;
GRANT ALL PRIVILEGES ON DATABASE konga_db TO erp_user;

-- Display created databases
\l
