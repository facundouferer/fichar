# AGENTS.md

## Project State

**Project initialized** - Docker Compose setup complete with PostgreSQL, backend and frontend scaffolding. See `docs/documento_inicial.md` for specification.

## Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Astro
- **Database**: PostgreSQL
- **Infrastructure**: Docker + docker-compose

## Architecture

Clean Architecture / Hexagonal with these layers:
- `cmd/` - Entry points
- `internal/` - Domain, repository, service, handler, middleware, config
- `pkg/` - logger, pdf, database utilities

## API Endpoints (planned)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/attendance/check` | Register check-in/out by DNI |
| GET | `/api/employees/{id}/attendances` | Get employee attendance |
| POST | `/api/admin/employees` | Create employee |
| POST | `/api/admin/shifts` | Create shift |
| POST | `/api/admin/employee-shifts` | Assign shift to employee |
| GET | `/api/reports/monthly` | Generate monthly report |

## Database Schema

Tables: `employees`, `shifts`, `employee_shift_assignments`, `attendances`, `logs`

Initial admin credentials: `admin` / `admin` (must change on first login)

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

```bash
# Setup
cp .env.example .env  # Copy environment variables

# Start all services (PostgreSQL, Backend, Frontend)
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Rebuild services
docker-compose up -d --build

# Stop and remove volumes (for clean slate)
docker-compose down -v

# Validate docker-compose configuration
docker-compose config
```

**Services**:
- `postgres`: PostgreSQL database on port 5432
- `backend`: Go API on port 8080
- `frontend`: Astro UI on port 4321

**Initial Admin**: `admin` / `admin` (must change on first login)

---

## Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

**Important**: Frontend uses custom Astro build + Express server approach for Docker deployment.

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