# Sistema de Registro de Ingreso y Egreso de Empleados

## Introducción

Este documento describe el diseño y los requisitos de un sistema de control de ingreso y egreso de empleados basado en el uso del DNI como identificador principal. El sistema permitirá registrar las entradas y salidas de los empleados, gestionar turnos de trabajo, generar informes y analizar el cumplimiento de las horas laborales. La solución estará construida utilizando **Go** para el backend, **Astro** para el frontend y **PostgreSQL** como base de datos. Toda la aplicación estará **dockerizada** para facilitar su despliegue y la base de datos se inicializará automáticamente con su estructura al momento de la implementación.

El sistema contará con dos tipos de usuarios: **administradores** y **empleados**. Los administradores podrán gestionar usuarios, turnos y configuraciones del sistema, mientras que los empleados podrán visualizar sus registros de asistencia.

---

## Objetivos del Sistema

- Registrar el ingreso y egreso diario de los empleados mediante su DNI.
- Gestionar turnos laborales con horario de inicio y fin.
- Detectar llegadas tarde y horas trabajadas adicionales.
- Calcular el cumplimiento de horas laborales mensuales.
- Permitir cambios de turno dentro del mismo mes.
- Generar informes por día, semana, mes o rango horario.
- Exportar informes mensuales en formato PDF.
- Registrar logs de todas las acciones del sistema.
- Permitir a los empleados consultar su propio historial.

---

## Arquitectura Tecnológica

El sistema estará compuesto por tres componentes principales:

### Backend

- Lenguaje: **Go (Golang)**
- API: REST
- Framework sugerido: Fiber / Gin
- Manejo de autenticación y roles
- Generación de informes
- Generación de PDF
- Sistema de logging

### Frontend

- Framework: **Astro**
- Interfaz para registro rápido por DNI
- Panel administrativo
- Panel de empleado
- Visualización de reportes

### Base de Datos

- Motor: **PostgreSQL**
- Gestión de empleados
- Registro de asistencias
- Gestión de turnos
- Registro de cambios de turno
- Sistema de logs

### Contenedores

El sistema será distribuido mediante **Docker**, incluyendo:

- Contenedor backend
- Contenedor frontend
- Contenedor base de datos PostgreSQL

Se utilizará **docker-compose** para orquestar los servicios.

---

## Funcionalidades del Sistema

### Registro de Ingreso/Egreso

El registro de asistencia se realizará ingresando el **DNI del empleado**.

Flujo:

1. El usuario ingresa su DNI en la pantalla principal.
2. El sistema identifica al empleado.
3. Se registra automáticamente si corresponde a **entrada o salida**.
4. Se muestra la información del registro realizado.

Cada registro guardará:

- DNI
- Fecha
- Hora de ingreso
- Hora de egreso

---

### Gestión de Empleados

El sistema almacenará la siguiente información:

- Nombre
- Apellido
- DNI
- Rol (Administrador / Empleado)
- Turno asignado

Los administradores podrán:

- Crear empleados
- Editar empleados
- Asignar turnos
- Cambiar turnos

---

### Gestión de Turnos

Los turnos estarán definidos por:

- Nombre del turno
- Hora de inicio
- Hora de fin
- Cantidad de horas laborales

Los administradores podrán:

- Crear turnos
- Modificar turnos
- Asignar turnos a empleados

---

### Cambios de Turno

El sistema permitirá registrar **cambios de turno dentro del mes**.

Esto permitirá calcular correctamente:

- Horas trabajadas
- Horas esperadas
- Diferencias

Cada cambio de turno deberá registrar:

- Empleado
- Turno anterior
- Nuevo turno
- Fecha de inicio del cambio

---

### Control de Horas

El sistema podrá calcular:

- Horas trabajadas en el día
- Horas trabajadas en el mes
- Horas esperadas según turno
- Horas faltantes para cumplir la cuota mensual
- Horas extras realizadas

---

### Informes

El sistema permitirá generar informes filtrando por:

- Empleado
- Día
- Mes
- Rango de fechas

Los informes mostrarán:

- Entradas
- Salidas
- Horas trabajadas
- Horas faltantes
- Llegadas tarde

---

### Exportación de Informes

El sistema permitirá exportar **informes mensuales por empleado en formato PDF**.

El PDF incluirá:

- Datos del empleado
- Resumen mensual
- Registro diario de entradas y salidas
- Total de horas trabajadas

---

### Sistema de Logs

Todas las acciones del sistema deberán registrarse en un sistema de logs.

Ejemplos:

- Registro de ingreso
- Registro de salida
- Creación de usuario
- Cambio de turno
- Generación de informes

Cada log deberá guardar:

- Fecha
- Usuario
- Acción
- Descripción

---

## Roles del Sistema

### Administrador

Permisos:

- Crear usuarios
- Modificar usuarios
- Crear turnos
- Cambiar turnos de empleados
- Ver informes de todos los empleados
- Exportar informes

### Empleado

Permisos:

- Registrar ingreso/egreso
- Ver sus registros de asistencia
- Ver informes mensuales o anuales

---

## Despliegue

El sistema se desplegará mediante **Docker**.

El proceso de despliegue incluirá:

- Inicialización automática de la base de datos
- Creación automática de las tablas
- Levantamiento de los servicios

Se utilizará un archivo **docker-compose.yml** para gestionar los contenedores.

---

## Futuras Mejoras

- Integración con lector de tarjetas o QR
- Integración con sistemas de recursos humanos
- Dashboard analítico
- Notificaciones automáticas

---

## Modelo de Datos (PostgreSQL)

### Tabla: employees

| Campo | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador interno |
| dni | VARCHAR(20) | DNI del empleado (único) |
| first_name | VARCHAR(100) | Nombre |
| last_name | VARCHAR(100) | Apellido |
| role | VARCHAR(20) | ADMIN / EMPLOYEE |
| created_at | TIMESTAMP | Fecha de creación |
| updated_at | TIMESTAMP | Fecha de actualización |

### Tabla: shifts

| Campo | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador del turno |
| name | VARCHAR(100) | Nombre del turno |
| start_time | TIME | Hora de inicio |
| end_time | TIME | Hora de fin |
| expected_hours | NUMERIC | Horas esperadas |
| created_at | TIMESTAMP | Fecha creación |

### Tabla: employee_shift_assignments

Permite asignar turnos históricos.

| Campo | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador |
| employee_id | UUID | Empleado |
| shift_id | UUID | Turno |
| start_date | DATE | Inicio del turno |
| end_date | DATE | Fin del turno (nullable) |

### Tabla: attendances

| Campo | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador |
| employee_id | UUID | Empleado |
| date | DATE | Fecha |
| check_in | TIMESTAMP | Hora de entrada |
| check_out | TIMESTAMP | Hora de salida |
| worked_hours | NUMERIC | Horas calculadas |
| late | BOOLEAN | Llegada tarde |

### Tabla: logs

| Campo | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador |
| user_id | UUID | Usuario que ejecuta acción |
| action | VARCHAR(100) | Tipo de acción |
| description | TEXT | Detalle |
| created_at | TIMESTAMP | Fecha |

---

## Estructura del Backend (Go)

Arquitectura recomendada: **Clean Architecture / Hexagonal**.

```
/backend

cmd/
  server/
    main.go

internal/

  domain/
    employee.go
    attendance.go
    shift.go

  repository/
    employee_repository.go
    attendance_repository.go
    shift_repository.go

  service/
    attendance_service.go
    shift_service.go

  handler/
    employee_handler.go
    attendance_handler.go

  middleware/
    auth.go
    logging.go

  config/

pkg/
  logger/
  pdf/
  database/
```

---

## API REST

### Registro de ingreso/egreso

POST

```
/api/attendance/check
```

Body

```
{
  "dni": "12345678"
}
```

Respuesta

```
{
  "status": "checkin",
  "employee": "Juan Perez",
  "time": "08:00"
}
```

---

### Obtener registros de un empleado

GET

```
/api/employees/{id}/attendances
```

Filtros

```
?month=05
?year=2026
```

---

### Crear empleado

POST

```
/api/admin/employees
```

---

### Crear turno

POST

```
/api/admin/shifts
```

---

### Asignar turno

POST

```
/api/admin/employee-shifts
```

---

### Generar informe

GET

```
/api/reports/monthly
```

---

## Lógica de Negocio

### Determinar Entrada o Salida

Algoritmo:

1. Buscar registro del día.
2. Si no existe → registrar entrada.
3. Si existe y no tiene salida → registrar salida.

---

### Cálculo de Horas Trabajadas

```
worked_hours = checkout - checkin
```

---

### Determinar Llegada Tarde

```
late = checkin > shift.start_time
```

---

### Horas Esperadas Mensuales

```
expected = shift_hours * working_days
```

---

### Horas Faltantes

```
missing_hours = expected - worked
```

---

## Generación de PDF

Librerías sugeridas en Go:

- gofpdf
- wkhtmltopdf

Contenido del PDF:

- Datos del empleado
- Tabla diaria
- Total mensual

---

## Frontend (Astro)

Estructura sugerida:

```
/frontend

src/

pages/
  index.astro
  login.astro

  admin/
    dashboard.astro
    employees.astro
    shifts.astro

  employee/
    dashboard.astro
    reports.astro

components/

layouts/

services/

styles/
```

---

## Dockerización

Servicios:

- backend
- frontend
- postgres

Ejemplo docker-compose:

```
services:

  postgres:
    image: postgres

  backend:
    build: ./backend

  frontend:
    build: ./frontend
```

La base se inicializa con **scripts SQL en /database/init**.

---

## Logging

Se recomienda utilizar logging estructurado.

Librerías:

- zerolog
- logrus

Eventos a registrar:

- Login
- Registro asistencia
- Cambios de turno
- Generación de reportes

---

## Seguridad

- Autenticación con JWT
- Roles RBAC
- Validación de entrada
- Rate limit en endpoint de DNI

---

## Observabilidad

Recomendado:

- métricas prometheus
- health check
- tracing opcional

---

## Estrategia de Testing

Tipos de tests:

- Unit tests
- Integration tests
- API tests

Herramientas:

- Go testing
- Testcontainers

---

## Roadmap Inicial

1. Infraestructura docker ✅
2. Esquema base de datos ✅
3. API registro asistencia ✅
4. CRUD empleados ✅
5. CRUD turnos ✅
6. Cálculo horas ✅
7. Reportes ✅
8. Exportación PDF ✅
9. Frontend ✅
10. Logging ✅
11. Observabilidad ✅
12. Testing ✅

## Roadmap Completado (Abril 2026)

Todas las issues iniciales (#1-#16) han sido implementadas:

| # | Issue | Estado |
|---|-------|--------|
| #1 | Bootstrap infraestructura | ✅ CLOSED |
| #2 | DB schema | ✅ CLOSED |
| #3 | Backend bootstrap | ✅ CLOSED |
| #4 | Auth JWT + RBAC | ✅ CLOSED |
| #5 | Employees CRUD | ✅ CLOSED |
| #6 | Shifts CRUD | ✅ CLOSED |
| #7 | Attendance check | ✅ CLOSED |
| #8 | Attendance calculations | ✅ CLOSED |
| #9 | Audit logging | ✅ CLOSED |
| #10 | Reports API | ✅ CLOSED |
| #11 | PDF export | ✅ CLOSED |
| #12 | Frontend public/auth | ✅ CLOSED |
| #13 | Frontend admin | ✅ CLOSED |
| #14 | Frontend employee | ✅ CLOSED |
| #15 | Observability | ✅ CLOSED |
| #16 | Testing | ✅ CLOSED |

## Funcionalidades Extensiones (Pendientes)

Las siguientes features fueron identificadas post-launch:
- #19 - Seteo de cantidad de horas por empleado
- #20 - Permitir correcciones manuales
- #21 - Informe Especial (PDF con texto libre)

---

## Estrategia de Issues

Cada feature debe dividirse en issues:

Ejemplo:

```
ISSUE-1
crear esquema base datos

ISSUE-2
endpoint registro asistencia

ISSUE-3
algoritmo calculo horas
```

---

## Guía para Contribuidores

1. Crear branch desde main
2. Implementar feature
3. Tests obligatorios
4. Pull request

---

## Métricas de Negocio

El sistema debe generar métricas que permitan a la gerencia comprender el desempeño de asistencia de los empleados y detectar problemas operativos.

### Métricas Principales

**1. Tasa de puntualidad**

Porcentaje de días en los que los empleados ingresan dentro del horario de su turno.

```
punctuality_rate = on_time_days / worked_days
```

---

**2. Horas trabajadas vs horas esperadas**

Comparación entre las horas efectivamente trabajadas y las horas que deberían haberse trabajado según el turno.

```
compliance_rate = worked_hours / expected_hours
```

---

**3. Horas extras acumuladas**

Cantidad de horas trabajadas por encima de las horas definidas por el turno.

---

**4. Horas faltantes**

Cantidad de horas que un empleado aún debe cumplir para alcanzar su cuota mensual.

---

**5. Cantidad de llegadas tarde**

Número de días en que un empleado ingresó luego del horario de su turno.

---

**6. Promedio de horas trabajadas por día**

```
average_hours = total_worked_hours / worked_days
```

---

**7. Índice de asistencia**

Relación entre días trabajados y días laborales esperados.

```
attendance_index = worked_days / expected_working_days
```

---

## Tipos de Informes

El sistema deberá generar distintos tipos de informes dependiendo del rol del usuario.

---

## Informes Disponibles para Administradores

Los administradores podrán acceder a información global del sistema.

### Informe de Asistencia General

Resumen del estado de asistencia de todos los empleados.

Incluye:

- Cantidad de empleados activos
- Cantidad de empleados presentes en el día
- Cantidad de ausentes
- Cantidad de llegadas tarde

---

### Informe Mensual por Empleado

Detalle completo de asistencia de un empleado durante un mes.

Incluye:

- Entradas y salidas por día
- Horas trabajadas
- Horas esperadas
- Horas faltantes
- Horas extras

Este informe puede exportarse a **PDF**.

---

### Informe de Cumplimiento de Horas

Muestra el grado de cumplimiento de las horas laborales por empleado.

Incluye:

- Horas esperadas
- Horas trabajadas
- Diferencia

---

### Informe de Llegadas Tarde

Listado de empleados con llegadas tarde dentro de un período.

Filtros disponibles:

- por empleado
- por mes
- por rango de fechas

---

### Informe de Horas Extras

Reporte que muestra:

- empleados con horas extras
- cantidad de horas extras
- fechas en las que se produjeron

---

### Dashboard Gerencial

Vista resumida con indicadores clave:

- empleados presentes hoy
- llegadas tarde hoy
- horas trabajadas totales del mes
- cumplimiento promedio del personal

---

## Informes Disponibles para Empleados

Los empleados solo podrán visualizar información relacionada con su propia asistencia.

### Informe Mensual Personal

Incluye:

- entradas y salidas diarias
- horas trabajadas
- horas esperadas
- horas faltantes
- horas extras

Este informe puede exportarse a **PDF**.

---

### Informe Anual Personal

Resumen anual de asistencia.

Incluye:

- horas totales trabajadas
- horas esperadas
- cantidad de llegadas tarde

---

### Estado de Horas del Mes

Vista rápida que muestra:

- horas trabajadas
- horas esperadas
- horas faltantes para cumplir la cuota

---

## Usuario Administrador Inicial

Durante la inicialización del sistema se deberá crear automáticamente un usuario administrador por defecto.

Credenciales iniciales:

```
usuario: admin
password: admin
rol: ADMIN
```

Este usuario permitirá ingresar por primera vez al sistema y configurar:

- empleados
- turnos
- asignaciones de turno

### Requisito de Seguridad

En el primer inicio de sesión el sistema deberá solicitar obligatoriamente:

- cambio de contraseña

---

## Conclusión

Esta documentación define una base técnica completa para que un equipo de desarrollo pueda iniciar la implementación del sistema de control de asistencia. Incluye arquitectura, modelo de datos, API, infraestructura, lógica de negocio, métricas de negocio e informes gerenciales, permitiendo dividir el trabajo en tareas claras y escalables.

