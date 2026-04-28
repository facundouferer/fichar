# AGENTS.md

Guia operativa para agentes de codigo en este repositorio.

## Estado real del proyecto (verificado)

- Stack activo: Go (backend), Astro + TypeScript (frontend), PostgreSQL, Docker Compose.
- Hay **dos proyectos Astro**:
  - `frontend/` (app principal real del sistema).
  - raiz (`package.json`, `src/pages/index.astro`) de plantilla base.
- Backend en `backend/` con tests unitarios y de handlers en Go.
- Infra en `docker-compose.yml` con servicios `postgres`, `backend`, `frontend`.

## Mapa rapido

- `backend/cmd/server/main.go` - bootstrap HTTP + middlewares + rutas.
- `backend/internal/` - domain, service, handler, middleware, repo.
- `backend/pkg/` - utilidades compartidas (DB, PDF).
- `frontend/src/pages/` - vistas Astro (admin y employee).
- `frontend/src/config/api.ts` - cliente API y tipos del frontend.
- `database/init/001_schema.sql` - schema inicial.
- `docs/operations.md` y `docs/testing-strategy.md` - operacion y testing.

## Cursor / Copilot

- `.cursor/rules/`: no existe.
- `.cursorrules`: no existe.
- `.github/copilot-instructions.md`: no existe.
- Si aparecen, pasan a ser politica prioritaria y este archivo debe actualizarse.

## Comandos de build / lint / test

## Requisitos generales

- Node: `>=22.12.0` (definido en `package.json` y `frontend/package.json`).
- Go: `1.23` (definido en `backend/go.mod`).
- Docker Compose disponible (`docker-compose` en docs/scripts actuales).

**Services**:
- `postgres`: PostgreSQL database on port 5432
- `backend`: Go API on port 8082
- `frontend`: Astro UI on port 4321

**Initial Admin**: set during first run or via seed script

## Key Constraints

- Authentication: JWT
- RBAC: ADMIN and EMPLOYEE roles
- Check-in/out logic: if no record for day → check-in; if record exists without check-out → check-out

This part provides guidelines for AI agents working on this codebase.

## Project Overview

| Property | Value |
|----------|-------|
| **Type** | Astro-based website |
| **Framework** | Astro v6 |
| **Language** | TypeScript (strict mode) |
| **Testing** | Playwright |
| **Node** | >=22.12.0 |

---

## Build & Development Commands

## Orquestacion con Docker

```bash
# Levantar stack completo
docker-compose up -d

# Ver logs
docker-compose logs -f

# Bajar stack
docker-compose down

# Rebuild de imagenes
docker-compose up -d --build

# Bajar y limpiar volumenes
docker-compose down -v

# Validar compose
docker-compose config

# Script de validacion
./validate-compose.sh
```

## Backend (Go)

```bash
# Desarrollo con hot reload
cd backend && air

# Build binario principal
cd backend && go build -o ./tmp/main ./cmd/server

# Build paquetes
cd backend && go build ./...

# Test suite completa
cd backend && go test ./...

# Coverage
cd backend && go test -cover ./...

# Tests cortos
cd backend && go test -short ./...

# Vet
cd backend && go vet ./...

# Formato
cd backend && gofmt -w .
```

## Frontend (app principal en `frontend/`)

```bash
cd frontend

# Instalar dependencias
npm install

# Dev server (custom node server)
npm run dev

# Build Astro
npm run build

# Preview / run server
npm run preview
```

Nota: hoy no hay script `lint` ni `test` en `frontend/package.json`.

---

## Frontend Structure

```
frontend/
├── src/
│   ├── pages/         # Route-based pages
│   ├── layouts/       # Page layouts
│   ├── config/        # Configuration files
│   ├── server/        # Custom Express server
│   └── entry.tsx      # Astro entry point
├── astro.config.mjs   # Astro configuration
├── package.json       # Dependencies
└── tsconfig.json      # TypeScript configuration
```

---

## Agent Workflow Requirements

When working on this project, agents **MUST** follow these steps:

1. **Explain Actions** — State what you're doing and why before making changes
2. **Verify** — Run `docker-compose config` to validate configuration
3. **Test** — Start services with `docker-compose up` and verify all containers are healthy
4. **Fix Errors** — Correct any reported errors before marking the task complete
5. **Close Issue** — After implementation is verified, close the GitHub issue:

   ```bash
   gh issue close <ISSUE_NUMBER> --comment "Implemented in this PR/branch"
   ```

6. **Commit Changes** — Commit all changes with a descriptive message:

   ```bash
   git add .
   git commit -m "feat: <description of what was implemented>"
   git push
   ```

---

## Code Style Guidelines

### TypeScript

- Uses `astro/tsconfigs/strict` — ALL strict rules enabled
- **Never use `any`** — Use `unknown` when type is truly unknown
- Enable strict mode in your editor
- Use explicit type annotations for function parameters and return types

### Astro Components (.astro files)

```
// File structure order:
1. --- frontmatter (imports, props, logic)
2. Template HTML
3. <style> block (at bottom)
```

- **Imports**: Sort alphabetically within groups
- **Indentation**: 2 spaces for HTML/JSX
- **Props**: Use TypeScript interfaces for component props
- **Scripts**: Use `is:global` for styles that need to leak to children

### CSS/Styling

- Use **scoped styles** within `<style>` blocks
- Prefer **CSS custom properties** for theme values
- Modern CSS features: flexbox, grid, `clamp()`, `calc()`
- Mobile-first responsive design
- Class naming: **kebab-case** (e.g., `.modal-overlay`)

### File Organization

```
src/
├── assets/          # Static assets (images, fonts)
├── components/      # Reusable Astro components
├── layouts/         # Page layouts
└── pages/           # Route-based page components
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Components | PascalCase | `CityScene.astro`, `Layout.astro` |
| Regular files | kebab-case | `my-component.astro` |
| CSS classes | kebab-case | `.user-panel`, `.modal-content` |
| Variables | camelCase | `isLoading`, `buildingConfig` |
| Constants | SCREAMING_SNAKE | `MAX_BUILDINGS`, `GT_BLUE` |
| Types/Interfaces | PascalCase | `BuildingConfig`, `ModalData` |

### Imports

- Use relative imports for local files
- Sort imports alphabetically: `import A from 'a'; import B from 'b';`
- Third-party imports first, then local imports
- Group by: external → internal → relative

---

## Error Handling

- Let Astro handle errors with built-in error pages
- For client-side code: wrap async operations in try/catch
- Use `console.error()` for recoverable errors, `throw` for fatal ones
- Validate DOM elements exist before use:

```typescript
const element = document.getElementById('my-id');
if (!element) {
  console.error('Required element #my-id not found');
  throw new Error('Canvas not found');
}
```

---

## Local Development Setup (Hot Reload)

**AVOID rebuilding Docker images for every change.** Use local development with hot reload instead.

### Backend (Go)

```bash
# 1. Keep PostgreSQL running in Docker
docker-compose up -d postgres

# 2. Run with hot reload (from backend/ directory)
air
```

- `air` auto-restarts on Go file changes
- Backend available at `http://localhost:8082`

### Frontend (Astro)

```bash
# 1. Install dependencies (first time only)
cd frontend && npm install

# 2. Stop Docker frontend to avoid port conflict
docker-compose stop frontend

# 3. Run dev server with hot reload
npm run dev
npm run build
npm run preview
```

## Ejecutar un solo test (muy importante)

## Go (backend)

```bash
# Un paquete
cd backend && go test ./internal/service -v

# Un test exacto
cd backend && go test ./internal/service -run '^TestAuthServiceLogin$' -v

# Otro ejemplo real
cd backend && go test ./internal/handler -run '^TestHealthEndpoint$' -v

# Por prefijo/patron
cd backend && go test ./internal/service -run 'TestAuthService' -v
```

## Frontend

Test API endpoints directly:
```bash
curl -X POST http://localhost:8082/api/attendance/check \
  -H "Content-Type: application/json" \
  -d '{"dni":"00000000","latitude":-27.46,"longitude":-58.98}'
```

## Convenciones de codigo

## Reglas generales

1. **Prefer local development over Docker rebuilds** - Use `air` and `npm run dev` for faster iteration
2. **Always test changes locally first** before committing
3. **Rebuild Docker only when necessary** - e.g., after adding new dependencies or changing Dockerfile
4. **Commit documentation updates** - Keep README.md and AGENTS.md in sync with actual workflows
5. **Update `.air.toml`** if new build flags or paths are added to the backend
6. **Do not create new branches unless explicitly asked to**

## Imports

- Go: stdlib -> internos del modulo -> terceros.
- TS/Astro: externos -> internos absolutos -> relativos.
- Eliminar imports no usados antes de cerrar tarea.

## Formato

- Go: `gofmt` obligatorio.
- TS/Astro: respetar estilo actual (2 espacios en `.astro`, semicolons en TS).
- No introducir otro formatter sin acuerdo del repo.

## Tipos

- Go: structs tipadas para dominio; evitar `interface{}` salvo borde dinamico.
- TypeScript: evitar `any`; usar tipos concretos o `unknown` + narrowing.
- Reusar tipos de `frontend/src/config/api.ts` antes de duplicar interfaces.

## Naming

- Go exportado: `PascalCase`; interno: `camelCase`.
- Tests Go: `TestXxx` / `TestXxx_Scenario`.
- Astro layouts/componentes: `PascalCase.astro`.
- Variables JS/TS: `camelCase`; constantes: `UPPER_SNAKE_CASE` solo si aplica.

## Manejo de errores

- Nunca silenciar errores.
- Backend HTTP: responder con status correcto y mensaje claro.
- En Go, envolver errores con contexto cuando se propagan.
- Frontend: capturar async con `try/catch` y mostrar mensaje usable.

## API y seguridad

- Validar input en handlers (body, query, params).
- Mantener JWT + RBAC (`ADMIN` / `EMPLOYEE`) en capas de middleware.
- No loggear secretos (`JWT_SECRET`, credenciales, tokens completos).
- Respetar rate limiting en endpoints publicos.

## Testing

- Cada cambio de logica de negocio requiere test o ajuste de test.
- Preferir table-driven tests en Go cuando hay multiples escenarios.
- Tests de handler deben cubrir casos validos + invalidos + bordes.

## Flujo recomendado de desarrollo

1. Levantar `postgres` con Docker.
2. Ejecutar backend local con `air`.
3. Ejecutar frontend desde `frontend/` con `npm run dev`.
4. Correr `go test ./...` en `backend/` antes de cerrar.
5. Validar compose con `docker-compose config` si tocaste infra.

## Notas para agentes

- No asumir que la app de raiz y `frontend/` son equivalentes; confirmar target.
- Si agregas scripts (`lint`, `test`, `typecheck`), actualiza este archivo.
- Si se incorporan reglas Cursor/Copilot, reflejarlas aca inmediatamente.