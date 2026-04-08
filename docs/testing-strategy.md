# Estrategia de Testing - Fichar

Este documento establece la estrategia de testing para el proyecto Fichar, incluyendo tipos de tests, estructura, y guías para contributors.

## Tipos de Tests

### 1. Tests Unitarios
Prueban funciones y métodos individuales de forma aislada.

**Ubicación:** Mismo paquete que el código bajo prueba, archivo `*_test.go`

**Ejemplo:**
```go
// backend/internal/service/auth_service_test.go
func TestAuthServiceLogin(t *testing.T) {
    // Arrange
    repo := NewMockEmployeeRepository()
    authSvc := NewAuthService(repo, "test-secret")
    
    // Act
    resp, err := authSvc.Login(ctx, &LoginRequest{
        DNI: "12345678",
        Password: "password123",
    })
    
    // Assert
    require.NoError(t, err)
    require.NotEmpty(t, resp.Token)
}
```

### 2. Tests de Integración
Prueban la interacción entre múltiples componentes o con la base de datos.

**Ubicación:** Paquete `integration/` o tests con标签 `-tags integration`

**Nota:** Actualmente los tests de integración están representados mediante mocks de repositorios. Los tests con base de datos real deben ejecutarse en un entorno controlado (Docker).

### 3. Tests de API / Handler
Prueban los endpoints HTTP y la validación de requests.

**Ubicación:** `backend/internal/handler/*_test.go`

**Ejemplo:**
```go
func TestCreateEmployeeValidation(t *testing.T) {
    tests := []struct {
        name    string
        body    map[string]interface{}
        wantErr string
    }{
        {
            name: "missing DNI returns error",
            body: map[string]interface{}{
                "first_name": "Juan",
                "last_name":  "Perez",
            },
            wantErr: "DNI is required",
        },
    }
    // ...
}
```

## Estructura de Archivos de Test

```
backend/
├── internal/
│   ├── handler/
│   │   ├── handler.go              # Implementación
│   │   ├── handler_test.go          # Tests
│   │   ├── handler_api_test.go      # Tests de API
│   │   └── handler_calculation_test.go  # Tests de cálculos
│   ├── service/
│   │   ├── service.go
│   │   ├── auth_service.go
│   │   ├── auth_service_test.go     # Tests de autenticación
│   │   └── service_calculation_test.go  # Tests de lógica de negocio
│   └── domain/
│       └── models.go
├── pkg/
│   └── pdf/
│       ├── report_service.go
│       └── report_service_test.go  # Tests de generación PDF
```

## Convenciones

### Naming
- `TestNombreFuncion` para tests de funciones específicas
- `TestNombreFuncion_Escenario` para tests con escenarios específicos
- Usar **table-driven tests** para múltiples casos de prueba

### Table-Driven Tests
```go
func TestCalculateHours(t *testing.T) {
    tests := []struct {
        name     string
        checkIn  string
        checkOut string
        expected float64
    }{
        {
            name:     "standard 8 hour day",
            checkIn:  "2026-04-07T08:00:00",
            checkOut: "2026-04-07T16:00:00",
            expected: 8.0,
        },
        {
            name:     "overtime 10 hours",
            checkIn:  "2026-04-07T08:00:00",
            checkOut: "2026-04-07T18:00:00",
            expected: 10.0,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := calculateHours(tt.checkIn, tt.checkOut)
            require.Equal(t, tt.expected, result)
        })
    }
}
```

### Mocks
Crear interfaces de mock en los archivos de test para evitar dependencias externas:

```go
type mockEmployeeRepo struct{}

func (m *mockEmployeeRepo) Create(ctx context.Context, emp *domain.Employee) error {
    return nil
}
func (m *mockEmployeeRepo) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
    return nil, nil
}
// ... implement all methods of repository.EmployeeRepository
```

## Ejecutar Tests

```bash
# Todos los tests
go test ./...

# Tests con coverage
go test -cover ./...

# Tests de un paquete específico
go test -v ./internal/service/...

# Tests por patrón
go test -run TestAuthService

# Ignorar tests largos (integración)
go test -short ./...
```

## Áreas Cubiertas

### Autenticación ✅
- Login exitoso
- Credenciales inválidas
- Cambio de contraseña obligatorio
- Validación de JWT
- Cambio de contraseña

### Cálculos ✅
- Horas trabajadas (con diferentes formatos de fecha)
- Detección de tardanzas (isLate)
- Días laborables por mes
- Rango de fechas

### Reportes ✅
- Generación de PDF
- Formateo de fechas/horas
- Dashboard summary

### Validación de Requests ✅
- Creación de empleados
- Check-in/check-out
- Creación de turnos

## Áreas que Necesitan Mejora

- Tests de integración con base de datos real
- Tests E2E completos
- Tests de concurrencia
- Tests de carga

## Recursos

- [Go Testing Patterns](https://go.dev/doc/code#Testing)
- [Testable Examples in Go](https://go.dev/blog/examples)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)