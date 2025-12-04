# Quick Start Guide - Portal Service with Kong

This guide will get you up and running with the Portal Service and Kong API Gateway in minutes.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.24 or higher installed
- PostgreSQL 15 installed locally (for local database setup)
- `jq` installed for testing (optional but recommended)
- `make` installed (optional but recommended)

## Two Setup Options

### Option A: Local PostgreSQL (Recommended for Development)
Uses your local PostgreSQL instance for better performance and easier database management.

### Option B: Containerized PostgreSQL
Runs PostgreSQL in Docker containers for complete isolation.

---

## Option A: Local PostgreSQL Setup

### Step 1: Prepare Local PostgreSQL

Ensure PostgreSQL is running and create the database:

```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql  # Linux
# or
brew services list | grep postgresql  # macOS

# Connect to PostgreSQL
psql -U postgres

# Create the database
CREATE DATABASE portal_db;

# Exit psql
\q
```

### Step 2: Configure PostgreSQL for Docker Access

Edit PostgreSQL configuration to accept connections from Docker containers:

```bash
# For Linux (adjust version number as needed)
sudo nano /etc/postgresql/15/main/pg_hba.conf

# For macOS
nano /opt/homebrew/var/postgresql@15/pg_hba.conf

# Add this line to allow Docker network connections:
host    all             all             172.17.0.0/16           md5

# Restart PostgreSQL
sudo systemctl restart postgresql  # Linux
# or
brew services restart postgresql@15  # macOS
```

### Step 3: Setup Environment Variables

```bash
# Navigate to the project directory
cd portal-service

# Copy environment variables
cp .env.example .env

# Edit .env for local PostgreSQL
nano .env
```

Update these values in `.env`:
```bash
# Database Configuration (for local PostgreSQL)
DB_HOST=host.docker.internal  # Docker's way to access host
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres  # Your PostgreSQL password
DB_NAME=portal_db
DB_SSLMODE=disable
```

### Step 4: Start Infrastructure Services

Start Kong, Redis, and RabbitMQ (without PostgreSQL container):

```bash
# Using the local database configuration
docker-compose -f docker-compose.local.yml up -d

# Check all services are healthy
docker-compose -f docker-compose.local.yml ps
```

Wait for all services to be healthy (~15-20 seconds). Verify services:

```bash
# Check Kong Admin API
curl http://localhost:3602/

# Check Redis
docker-compose -f docker-compose.local.yml exec redis redis-cli -a redis123 ping

# Check RabbitMQ
curl http://localhost:3608  # Management UI
```

### Step 5: Database Migrations and Seeding

The migrations and seeding will run automatically when you start the services. Check the logs:

```bash
# Check migration logs
docker-compose -f docker-compose.local.yml logs portal-migrations

# Check seeder logs
docker-compose -f docker-compose.local.yml logs portal-seeder
```

Verify in PostgreSQL:
```bash
psql -U postgres -d portal_db

# Check tables
\dt

# Check data
SELECT * FROM users;
SELECT * FROM tenants;
SELECT * FROM roles;
```

### Step 6: Verify Portal Service

The portal service should be running automatically. Check its status:

```bash
# Check portal service logs
docker-compose -f docker-compose.local.yml logs -f portal-service

# Test health endpoint
curl http://localhost:3502/health
```

---

## Option B: Containerized PostgreSQL Setup

### Step 1: Setup Environment Variables

```bash
# Navigate to the project directory
cd portal-service

# Copy environment variables
cp .env.example .env

# Edit .env if needed (defaults should work)
nano .env
```

Keep default database settings in `.env`:
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=portal_user
DB_PASSWORD=portal_pass
DB_NAME=portal_db
DB_SSLMODE=disable
```

### Step 2: Start All Services

Start all services including PostgreSQL containers:

```bash
# Using Make (recommended)
make up

# Or using docker-compose directly
docker-compose up -d
```

Wait for all services to be healthy (~15-20 seconds). Check status:

```bash
# Show all running containers
make ps

# Or
docker-compose ps
```

### Step 3: Verify Services

```bash
# Check health of all services
make health-check

# Check PostgreSQL
docker-compose exec postgres psql -U portal_user -d portal_db -c "\dt"

# Check Kong
curl http://localhost:3602/
```

---

## Common Steps (Both Options)

## Test the API

### Test Registration

Test the registration endpoint:

```bash
curl -X POST http://localhost:3502/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePass123!",
    "name": "Admin User",
    "company_name": "My Company"
  }'
```

Expected response:
```json
{
  "status": "success",
  "message": "Registration successful. User and tenant created.",
  "data": {
    "user_id": 1,
    "user_uuid": "550e8400-e29b-41d4-a716-446655440000",
    "email": "admin@example.com",
    "name": "Admin User",
    "tenant_id": 1,
    "tenant_name": "My Company",
    "tenant_slug": "my-company",
    "role": "owner"
  }
}
```

### Test Health Endpoint

```bash
curl http://localhost:3502/health
```

### Verify Kong Consumer

Check that the Kong consumer was created:

```bash
# Get the user_uuid from registration response
USER_UUID="550e8400-e29b-41d4-a716-446655440000"

# Check Kong consumer (Kong Admin API)
curl http://localhost:3602/consumers/$USER_UUID | jq

# Or list all consumers
curl http://localhost:3602/consumers | jq
```

## Verify Database

### For Local PostgreSQL (Option A):

```bash
# Connect to your local PostgreSQL
psql -U postgres -d portal_db
```

### For Containerized PostgreSQL (Option B):

```bash
# Connect to PostgreSQL container
docker-compose exec postgres psql -U portal_user -d portal_db
```

### Run Verification Queries:

```sql
-- Check users
SELECT pkid, uuid, email, name FROM users;

-- Check tenants
SELECT id, name, slug FROM tenants;

-- Check roles
SELECT id, tenant_id, name FROM roles;

-- Check memberships
SELECT m.id, u.email, t.name as tenant, r.name as role
FROM memberships m
JOIN users u ON m.user_id = u.pkid
JOIN tenants t ON m.tenant_id = t.id
JOIN roles r ON m.role_id = r.id;

-- Exit
\q
```

## Available Services

After starting the services, the following are available:

| Service | URL | Credentials | Notes |
|---------|-----|-------------|-------|
| Portal Service | http://localhost:3502 | - | Main application |
| Kong Proxy | http://localhost:3600 | - | API Gateway (public) |
| Kong Admin API | http://localhost:3602 | - | Kong management |
| PostgreSQL (Container) | localhost:3605 | portal_user / portal_pass | Option B only |
| PostgreSQL (Kong) | localhost:5433 | kong / kongpass | Kong's database |
| PostgreSQL (Local) | localhost:5432 | postgres / postgres | Option A only |
| Redis | localhost:3606 | Password: redis123 | Cache & sessions |
| RabbitMQ AMQP | localhost:3607 | guest / guest | Message broker |
| RabbitMQ UI | http://localhost:3608 | guest / guest | Management console |

## Common Commands

### For Local PostgreSQL Setup (docker-compose.local.yml):

```bash
# Start all services
docker-compose -f docker-compose.local.yml up -d

# Stop all services
docker-compose -f docker-compose.local.yml down

# View all logs
docker-compose -f docker-compose.local.yml logs -f

# View specific service logs
docker-compose -f docker-compose.local.yml logs -f portal-service
docker-compose -f docker-compose.local.yml logs -f kong

# Check service status
docker-compose -f docker-compose.local.yml ps

# Restart a service
docker-compose -f docker-compose.local.yml restart portal-service

# Rebuild and restart
docker-compose -f docker-compose.local.yml up -d --build portal-service
```

### For Containerized Setup (docker-compose.yml):

```bash
# Start all services
make up
# or
docker-compose up -d

# Stop all services
make down
# or
docker-compose down

# View logs
make logs
# or
docker-compose logs -f

# View specific service logs
make logs-blue  # Blue instance
docker-compose logs -f portal-service

# Check health
make health-check

# Check service status
make ps

# Restart services
docker-compose restart portal-service

# Clean up (remove volumes)
make clean
```

### Database Commands:

```bash
# For local PostgreSQL
psql -U postgres -d portal_db

# For containerized PostgreSQL
docker-compose exec postgres psql -U portal_user -d portal_db

# Backup database (local)
pg_dump -U postgres portal_db > backup.sql

# Restore database (local)
psql -U postgres portal_db < backup.sql
```

## API Endpoints

### Registration
```bash
POST http://localhost:3502/api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "User Name",
  "company_name": "Company Name"
}
```

### Health Check
```bash
GET http://localhost:3502/health
```

## Troubleshooting

### Kong is not starting
```bash
# Check Kong logs
docker-compose logs kong
# or for local setup
docker-compose -f docker-compose.local.yml logs kong

# Check Kong database
docker-compose logs kong-database

# Restart Kong
docker-compose restart kong
```

### Database connection failed (Option A - Local PostgreSQL)

```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql  # Linux
brew services list | grep postgresql  # macOS

# Test connection from Docker
docker run --rm postgres:15-alpine psql -h host.docker.internal -U postgres -d portal_db

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-15-main.log  # Linux
tail -f /opt/homebrew/var/log/postgresql@15.log  # macOS

# Verify pg_hba.conf allows Docker network
cat /etc/postgresql/15/main/pg_hba.conf | grep 172.17
```

### Database connection failed (Option B - Containerized)

```bash
# Check PostgreSQL container is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Verify connection
docker-compose exec postgres psql -U portal_user -d portal_db

# Restart PostgreSQL
docker-compose restart postgres
```

### Port already in use
```bash
# Check what's using the port
sudo lsof -i :3502  # Portal Service
sudo lsof -i :3600  # Kong Proxy
sudo lsof -i :3602  # Kong Admin API
sudo lsof -i :3605  # PostgreSQL
sudo lsof -i :5432  # Local PostgreSQL

# Stop the conflicting service or change ports in docker-compose.yml
```

### Migration errors
```bash
# Check migration logs
docker-compose logs portal-migrations
# or
docker-compose -f docker-compose.local.yml logs portal-migrations

# Manually run migrations (if needed)
docker-compose exec portal-service /app/migrate up

# Check migration table in database
psql -U postgres -d portal_db -c "SELECT * FROM schema_migrations;"
```

### Service not accessible
```bash
# Check if service is running
docker-compose ps

# Check service logs
docker-compose logs portal-service

# Check health endpoint
curl http://localhost:3502/health

# Restart the service
docker-compose restart portal-service

# Rebuild and restart
docker-compose up -d --build portal-service
```

### Redis connection issues
```bash
# Check Redis is running
docker-compose ps redis

# Test Redis connection
docker-compose exec redis redis-cli -a redis123 ping

# Check Redis logs
docker-compose logs redis
```

### RabbitMQ connection issues
```bash
# Check RabbitMQ is running
docker-compose ps rabbitmq

# Check RabbitMQ logs
docker-compose logs rabbitmq

# Access RabbitMQ management UI
open http://localhost:3608
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Applications                       │
│                    (Portal FE, ERP FE, Mobile)                   │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                    ┌───────────▼──────────┐
                    │   Kong API Gateway   │
                    │   (Port 3600/3602)   │
                    │  - JWT Validation    │
                    │  - Rate Limiting     │
                    │  - Routing           │
                    └───────────┬──────────┘
                                │
                    ┌───────────▼──────────┐
                    │   Portal Service     │
                    │    (Port 3502)       │
                    │  - Authentication    │
                    │  - User Management   │
                    │  - Tenant Management │
                    └─────┬─────┬─────┬────┘
                          │     │     │
        ┌─────────────────┘     │     └─────────────────┐
        │                       │                        │
   ┌────▼─────┐          ┌─────▼────┐          ┌───────▼──────┐
   │PostgreSQL│          │  Redis   │          │  RabbitMQ    │
   │(Port 5432│          │(Port 3606│          │(Port 3607)   │
   │ or 3605) │          │          │          │              │
   │- Users   │          │- Sessions│          │- Events      │
   │- Tenants │          │- Tokens  │          │- Async Tasks │
   │- Roles   │          │- Cache   │          │              │
   └──────────┘          └──────────┘          └──────────────┘
```

## Next Steps

1. **Read Architecture Documentation** - See [ARCHITECTURE_REFERENCE.md](../ARCHITECTURE_REFERENCE.md)
2. **Configure Kong Routes** - Use the provided Kong configuration scripts
3. **Implement Authentication Flow** - See [AUTH_FLOW_SIMPLE.md](../AUTH_FLOW_SIMPLE.md)
4. **Add More Services** - Integrate inventory, manufacturing services
5. **Setup CI/CD** - Configure deployment pipelines
6. **Enable Monitoring** - Setup logging and metrics collection

## Development Workflow

### Daily Development:

1. **Start services** (choose your setup):
   ```bash
   # Local PostgreSQL (Option A)
   docker-compose -f docker-compose.local.yml up -d
   
   # Or Containerized (Option B)
   docker-compose up -d
   ```

2. **Make code changes** in your editor

3. **Rebuild and restart** after changes:
   ```bash
   # Local setup
   docker-compose -f docker-compose.local.yml up -d --build portal-service
   
   # Containerized setup
   docker-compose up -d --build portal-service
   ```

4. **Test your changes**:
   ```bash
   curl http://localhost:3502/health
   curl -X POST http://localhost:3502/api/v1/auth/register ...
   ```

5. **Check logs** if something goes wrong:
   ```bash
   docker-compose logs -f portal-service
   ```

### Adding New Migrations:

```bash
# Create a new migration file in migrations/ directory
# Example: 003_add_new_table.up.sql and 003_add_new_table.down.sql

# Restart migrations service to apply
docker-compose restart portal-migrations

# Or manually run
docker-compose exec portal-service /app/migrate up
```

## Production Deployment

For production deployment, consider:

1. **Security**:
   - Use strong passwords in `.env`
   - Enable HTTPS on Kong (configure SSL certificates)
   - Set `APP_ENV=production`
   - Use proper secret management (AWS Secrets Manager, Vault, etc.)
   - Configure PostgreSQL with SSL enabled

2. **High Availability**:
   - Set up Kong in DB-less mode or with clustering
   - Use PostgreSQL replication (master-slave)
   - Configure Redis Sentinel for high availability
   - Set up RabbitMQ clustering

3. **Monitoring**:
   - Configure Kong logging plugins
   - Setup Prometheus metrics
   - Use ELK stack or similar for log aggregation
   - Configure health checks and alerts

4. **Performance**:
   - Configure proper database connection pooling
   - Enable Redis persistence
   - Configure Kong rate limiting
   - Use CDN for static assets

5. **Backup**:
   - Schedule regular PostgreSQL backups
   - Backup Redis data if needed
   - Version control your configurations

## Additional Resources

- **Kong Documentation**: https://docs.konghq.com/
- **Gin Framework**: https://gin-gonic.com/docs/
- **GORM**: https://gorm.io/docs/
- **PostgreSQL**: https://www.postgresql.org/docs/
- **Redis**: https://redis.io/documentation
- **RabbitMQ**: https://www.rabbitmq.com/documentation.html

## Support & Contributing

For issues or questions:
1. Check the troubleshooting section above
2. Review the architecture documentation
3. Check Kong and application logs
4. Open an issue in the repository

## License

MIT
