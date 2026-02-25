-- Create databases for each microservice
-- This script runs on first PostgreSQL container start

-- Create databases
CREATE DATABASE identity_db;
CREATE DATABASE catalog_db;
CREATE DATABASE listing_db;
CREATE DATABASE booking_db;
CREATE DATABASE deal_db;
CREATE DATABASE payment_db;
CREATE DATABASE doc_db;
CREATE DATABASE review_db;
CREATE DATABASE services_db;
CREATE DATABASE engagement_db;
CREATE DATABASE integrity_db;
CREATE DATABASE analytics_db;
CREATE DATABASE media_db;

-- Connect to identity_db and create extensions
\c identity_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to catalog_db and create extensions
\c catalog_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to listing_db and create extensions
\c listing_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to booking_db and create extensions
\c booking_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to deal_db and create extensions
\c deal_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to payment_db and create extensions
\c payment_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to doc_db and create extensions
\c doc_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to review_db and create extensions
\c review_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to services_db and create extensions
\c services_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to engagement_db and create extensions
\c engagement_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to integrity_db and create extensions
\c integrity_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to analytics_db and create extensions
\c analytics_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to media_db and create extensions
\c media_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
