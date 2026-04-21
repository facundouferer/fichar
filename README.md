# Fichar — Control de Asistencia

Sistema de registro de asistencia laboral con autenticación basada en ubicación GPS. Los empleados registran su ingreso y egreso desde una ubicación autorizada.

## Propósito

- Registrar check-in/check-out de empleados vía DNI
- Autenticación por ubicación (oficina autorizada)
- Gestión de turnos y empleados (admin)
- Reportes mensuales de asistencia

## Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Astro
- **Database**: PostgreSQL
- **Infraestructura**: Docker + docker-compose

---

## Inicio Rápido

```bash
# Copiar variables de entorno
cp .env.example .env

# Iniciar servicios
docker-compose up -d

# Acceder a la app
open http://localhost:4321
```

**Credenciales iniciales:**
- Usuario: `00000000`
- Contraseña: `admin` (cambiar en primer login)

---

## Desarrollo Local

### Requisitos

- Docker + docker-compose
- Node.js >= 22.12.0 (para frontend)
- Go 1.21+ (para backend)

### Backend con hot-reload

```bash
# 1. Mantener PostgreSQL en Docker
docker-compose up -d postgres

# 2. Ejecutar backend con air
cd backend && air
```

El backend corre en `http://localhost:8080`.

### Frontend con hot-reload

```bash
cd frontend && npm install
npm run dev
```

El frontend corre en `http://localhost:4321`.

### Flujo de desarrollo

```bash
# Terminal 1: PostgreSQL + Backend
docker-compose up -d postgres
cd backend && air

# Terminal 2: Frontend
cd frontend && npm run dev
```

---

## API Endpoints

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/api/attendance/check` | Registrar check-in/out por DNI |
| POST | `/api/auth/login` | Login admin/empleado |
| GET | `/api/employees/{id}/attendances` | Ver asistencia de empleado |
| POST | `/api/admin/employees` | Crear empleado |
| POST | `/api/admin/shifts` | Crear turno |
| POST | `/api/admin/employee-shifts` | Asignar turno a empleado |
| GET | `/api/reports/monthly` | Reporte mensual |

---

## Variables de Entorno

### Backend

```env
POSTGRES_DB=fichar
POSTGRES_USER=fichar_user
POSTGRES_PASSWORD=changeme
DATABASE_URL=postgres://fichar_user:changeme@localhost:5432/fichar?sslmode=disable
JWT_SECRET=change-me-in-production
OFFICE_LATITUDE=-27.46
OFFICE_LONGITUDE=-58.98
OFFICE_RADIUS_KM=5
```

### Frontend

```env
PUBLIC_API_URL=http://localhost:8080
```

---

## Arquitectura

```
backend/
├── cmd/              # Puntos de entrada
├── internal/
│   ├── domain/       # Entidades y interfaces
│   ├── repository/  # Acceso a datos
│   ├── service/    # Lógica de negocio
│   ├── handler/    # HTTP handlers
│   ├── middleware/ # Autenticación, RBAC
│   └── config/     # Configuración
└── pkg/             # Utilidades (logger, PDF)
```

---

## Reglas para Contribuidores

### Commits

Usar [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: resolve a bug
docs: update documentation
refactor: restructure without behavior change
test: add or update tests
chore: maintenance, deps, config
```

### Issues

**Antes de crear un issue:**
1. Buscar si ya existe uno similar
2. Usar templates de GitHub
3. Incluir:
   - Pasos para reproducir
   - Comportamiento esperado vs actual
   - Screenshots/logs si aplica

### Pull Requests

**Antes de abrir un PR:**
1. linkeditar un issue relacionado
2. Hacer fork del repositorio
3. Crear branch desde `main`: `git checkout -b feature/descripcion`
4. Seguir code style del proyecto
5. Verificar que pasa `docker-compose config`

**En el PR:**
- Descripción clara del cambio
- Steps to test
- Screenshots si hay cambio visual

**No hacer:**
- PRs sin issue vinculada
- Commits tipo "fix", "update", "wip"
- Mezclar múltiples features

### Code Review

- Ser constructivo — explicar el por qué
- Apretar buenos cambios, no solo encontrar errores
- Responder en 48hs

---

## Comandos Docker

```bash
# Iniciar servicios
docker-compose up -d

# Ver logs
docker-compose logs -f

# Detener servicios
docker-compose down

# Rebuild
docker-compose up -d --build

# Clean slate
docker-compose down -v

# Validate config
docker-compose config
```

---

## Licencia

MIT