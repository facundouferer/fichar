# Skill Registry - Fichar Project

Generated: 2026-04-23

## Project Skills

| Skill | Location | Trigger |
|-------|----------|---------|
| uncodixfy | `.agents/skills/uncodixfy/` | Frontend UI code generation |

## User Skills (SDD)

These skills are available from the user's configuration and apply to all projects.

| Skill | Purpose |
|-------|---------|
| sdd-init | Initialize SDD context |
| sdd-explore | Explore and investigate ideas |
| sdd-propose | Create change proposals |
| sdd-spec | Write specifications |
| sdd-design | Technical design documents |
| sdd-tasks | Implementation task checklists |
| sdd-apply | Implement tasks |
| sdd-verify | Validate implementation |
| sdd-archive | Archive completed changes |
| sdd-onboard | Guided SDD walkthrough |
| judgment-day | Adversarial code review |
| skill-registry | Update skill registry |
| skill-creator | Create new AI skills |

## Project Conventions

- **AGENTS.md** exists in project root - contains all project-specific developer guidelines
- **Architecture**: Clean Architecture / Hexagonal with layers: `cmd/`, `internal/`, `pkg/`
- **Frontend**: Astro with TypeScript strict mode
- **Backend**: Go with pgx/v5 driver

## Testing Capabilities

### Backend (Go)
- **Test Runner**: `go test` (built-in)
- **Unit Tests**: Available (*_test.go files)
- **No integration/E2E testing frameworks installed**

### Frontend (Astro)
- **No test framework installed**
- AGENTS.md mentions Playwright but it's not in dependencies