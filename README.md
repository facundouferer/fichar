# Fichar - Sistema de Control de Asistencia

Sistema de control de asistencia con autenticación basada en ubicación GPS.

## Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Astro
- **Database**: PostgreSQL
- **Infrastructure**: Docker + docker-compose

## Quick Start

```bash
# Start all services
docker-compose up -d

# Access the app
open http://localhost:4321
```

## Development Setup

### Backend (Go with Hot Reload)

**1. Keep PostgreSQL running in Docker:**
```bash
docker-compose up -d postgres
```

**2. Run backend with hot reload:**
```bash
cd backend
air
```

The backend will be available at `http://localhost:8080`.

Changes to Go files will automatically restart the server.

### Frontend (Astro with Hot Reload)

**1. Install dependencies:**
```bash
cd frontend
npm install
```

**2. Run dev server:**
```bash
npm run dev
```

The frontend will be available at `http://localhost:4321`.

Changes to Astro/TypeScript files will automatically refresh.

### Running Backend + Frontend Together

```bash
# Terminal 1: PostgreSQL + Backend with hot reload
docker-compose up -d postgres
cd backend && air

# Terminal 2: Frontend dev server
cd frontend && npm run dev
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/attendance/check` | Register check-in/out by DNI |
| POST | `/api/auth/login` | Admin/Employee login |
| GET | `/api/employees/{id}/attendances` | Get employee attendance |
| POST | `/api/admin/employees` | Create employee |
| POST | `/api/admin/shifts` | Create shift |
| POST | `/api/admin/employee-shifts` | Assign shift to employee |
| GET | `/api/reports/monthly` | Generate monthly report |

## Environment Variables

### Backend
```env
POSTGRES_DB=fichar
POSTGRES_USER=fichar_user
POSTGRES_PASSWORD=changeme
DATABASE_URL=postgres://fichar_user:changeme@localhost:5432/fichar?sslmode=disable
JWT_SECRET=change-me-in-production
OFFICE_LATITUDE=-27.46768274122434
OFFICE_LONGITUDE=-58.98517836698102
OFFICE_RADIUS_KM=5
```

### Frontend
```env
PUBLIC_API_URL=http://localhost:8080
```

## Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Rebuild services
docker-compose up -d --build

# Stop and remove volumes (clean slate)
docker-compose down -v

# Validate docker-compose configuration
docker-compose config
```

## Initial Admin Credentials

- **Username**: `00000000`
- **Password**: `admin` (must change on first login)

## Database Schema

Tables: `employees`, `shifts`, `employee_shift_assignments`, `attendances`, `logs`

## Architecture

Clean Architecture / Hexagonal with these layers:
- `cmd/` - Entry points
- `internal/` - Domain, repository, service, handler, middleware, config
- `pkg/` - logger, pdf, database utilities
