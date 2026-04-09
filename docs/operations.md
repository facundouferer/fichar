# Operaciones y Monitoreo

Este documento describe los endpoints de salud y métricas disponibles, junto con lineamientos para el monitoreo operacional del sistema.

## Endpoints de Salud

### Liveness Check
**GET `/health`** o **GET `/health/live`**

Verifica que el servicio esté corriendo. No verifica dependencias.

```json
{
  "status": "ok",
  "service": "fichar-backend"
}
```

### Readiness Check
**GET `/health/ready`**

Verifica que el servicio esté listo para recibir tráfico, incluyendo conectividad con la base de datos.

```json
{
  "status": "ready",
  "service": "fichar-backend",
  "database": "connected"
}
```

Si la base de datos no está disponible:
```json
{
  "status": "unavailable",
  "service": "fichar-backend",
  "database": "not connected"
}
```

## Métricas

### GET `/metrics`

Retorna métricas operacionales básicas:

```json
{
  "requests_total": 1234,
  "database_healthy": true,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

| Métrica | Descripción |
|---------|-------------|
| `requests_total` | Total de requests procesados desde el inicio del servicio |
| `database_healthy` | Estado de conectividad con la base de datos |
| `timestamp` | Timestamp UTC de la medición |

## Verificación con Docker

### Verificar estado de salud de los servicios

```bash
# Ver estado de los contenedores
docker-compose ps

# Verificar que todos los servicios estén healthy
docker-compose healthcheck
```

### Healthchecks configurados

| Servicio | Healthcheck |
|----------|-------------|
| postgres | `pg_isready -U fichar_user -d fichar` |
| backend | `wget -q --spider http://localhost:8080/health` |

### Verificar manualmente los endpoints

```bash
# Liveness
curl http://localhost:8080/health

# Readiness (incluye DB)
curl http://localhost:8080/health/ready

# Métricas
curl http://localhost:8080/metrics
```

## Lineamientos de Monitoreo

### Nivel 1: Disponibilidad básica
- Monitorizar que los contenedores estén corriendo (`docker ps`)
- Healthcheck de `/health` devuelve 200 OK

### Nivel 2: Dependencias críticas
- Healthcheck de `/health/ready` devuelve 200 OK
- PostgreSQL accesible y respondiendo

### Nivel 3: Métricas operacionales
-收集 `requests_total` para capacidad
- Alertar si `database_healthy` es `false`

### Recomendaciones para producción

1. **Prometheus/Grafana**: Configurar scrape de `/metrics` cada 30s
2. **Alertas**:
   - `/health` devuelve error → reiniciar contenedor
   - `/health/ready` devuelve error → notificar inmediatamente
   - `database_healthy: false` → verificar conectividad DB
3. **Logs**: Agregar cliente_id para trazabilidad
4. **Métricas adicionales a futuro**:
   - Latencia de requests (p50, p95, p99)
   - Uso de memoria y CPU
   - Conexiones activas a DB
   - Tasa de errores por endpoint

## Verificar el sistema

```bash
# Iniciar servicios
docker-compose up -d

# Verificar healthchecks
docker-compose ps

# Probar endpoints
curl -f http://localhost:8080/health && echo "OK"
curl -f http://localhost:8080/health/ready && echo "Ready"
curl -f http://localhost:8080/metrics && echo "Metrics OK"
```