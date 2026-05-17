# Guía de Desarrollo - Matching Service

## Quick Start

### Opción 1: Con Docker (Recomendado)

```bash
# Clonar o navegar al repositorio
cd matching_service

# Iniciar servicios
make run

# Verificar que el servicio está up
curl http://localhost:8084/health
```

### Opción 2: Desarrollo Local

```bash
# Instalar dependencias
make install-deps

# Asegurar que existe .env
cp .env.example .env

# Iniciar PostgreSQL localmente (si no está corriendo)
# En Linux/macOS:
docker run -d \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=matching_db \
  -p 5432:5432 \
  postgres:15-alpine

# Ejecutar migraciones
psql -U postgres -d matching_db < 001_create_initial_schema.sql

# Cargar datos de ejemplo (opcional)
psql -U postgres -d matching_db < seed_data.sql

# Iniciar servidor
make run-local
```

## Estructura del Proyecto

```
matching_service/
├── main.go                      # Punto de entrada principal
├── config.go                    # Configuración y carga de .env
├── models.go                    # Modelos de dominio
├── 001_create_initial_schema.sql # Migraciones de BD
├── seed_data.sql               # Datos de ejemplo
├── go.mod                       # Dependencias
├── Dockerfile                   # Imagen Docker
├── docker-compose.yml          # Stack local
├── Makefile                    # Comandos de desarrollo
├── API_SPEC.md                 # Especificación de API
└── DEVELOPMENT.md              # Este archivo
```

## Dependencias

- **Go 1.21+**: Lenguaje de programación
- **PostgreSQL 14+**: Base de datos
- **Docker & Docker Compose**: Para contenerización
- **Chi v5**: Router HTTP
- **Logrus**: Logging
- **JWT**: Autenticación
- **uuid**: Generación de UUIDs

## Variables de Entorno

Copia `.env.example` a `.env` y ajusta según necesidad:

```bash
APP_PORT=8084                  # Puerto del servidor
DB_HOST=localhost              # Host de PostgreSQL
DB_PORT=5432                   # Puerto de PostgreSQL
DB_USER=postgres               # Usuario de BD
DB_PASSWORD=postgres           # Contraseña de BD
DB_NAME=matching_db            # Nombre de la BD
DB_POOL_MAX=10                 # Conexiones máximas
JWT_SECRET=your-secret-key     # IMPORTANTE: cambiar en producción
NOTIFICATION_WEBHOOK_URL=...   # URL del servicio de notificaciones
LOG_LEVEL=info                 # Nivel de logging
```

## Comandos Útiles

```bash
# Instalar dependencias
make install-deps

# Iniciar servicios con Docker
make run

# Ejecutar localmente
make run-local

# Ver logs
make logs

# Detener servicios
make stop

# Ejecutar tests
make test

# Lint del código
make lint

# Shell de PostgreSQL
make db-shell

# Ver estado de contenedores
make ps

# Limpiar todo (destruye contenedores y datos)
make clean
```

## Arquitectura

### Componentes Principales

1. **Handlers** (REST API)
   - Reciben requests HTTP
   - Validan entrada
   - Retornan respuestas JSON

2. **Service Layer**
   - Lógica de negocio
   - Algoritmo de matching
   - Scoring de tutores

3. **Repository Layer**
   - Acceso a datos
   - Queries a PostgreSQL
   - Transacciones

4. **Models/Domain**
   - Estructuras de datos
   - Validaciones
   - Tipos personalizados

### Flujo de Matching

```
Solicitud de Estudiante
    ↓
Crear MatchRequest
    ↓
Procesar (POST /api/matching/process)
    ↓
Buscar Tutores Candidatos
    ↓
Calcular Scores (algoritmo)
    ↓
Crear MatchResults
    ↓
Notificar (webhook)
    ↓
Estudiante y Tutor Aceptan/Rechazan
```

### Algoritmo de Scoring

```
Score = (skill_match × 0.4) + 
        (availability_match × 0.3) + 
        (rating_score × 0.2) + 
        (price_match × 0.1)

Donde:
- skill_match: Porcentaje de habilidades que coinciden (0-100)
- availability_match: Compatibilidad de horarios (0-100)
- rating_score: Rating del tutor normalizado a 0-100
- price_match: Compatibilidad de precio (0-100)
```

## Testing

```bash
# Ejecutar todos los tests
make test

# Test con cobertura
go test -cover ./...

# Test específico
go test -v ./internal/service -run TestMatchingScore

# Ver cobertura en HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Debugging

### Con logs
```bash
# En development mode
LOG_LEVEL=debug make run-local

# Ver logs de contenedor
docker-compose logs -f matching_service
```

### Con breakpoints (VSCode)
```bash
# Crear .vscode/launch.json
# Configurar debugger de Go
# F5 para iniciar debug
```

### Base de datos
```bash
# Acceder a psql
make db-shell

# Ver tablas
\dt

# Ver datos de solicitudes
SELECT * FROM match_requests;

# Ver matches
SELECT * FROM match_results;
```

## Desarrollo de Nuevas Features

### 1. Crear modelo en models.go
```go
type NewFeature struct {
    ID    uuid.UUID
    Name  string
}
```

### 2. Crear tabla en migraciones
```sql
CREATE TABLE new_features (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL
);
```

### 3. Crear repository methods
```go
func (s *MatchingService) CreateNewFeature(...) error {
    // implementar
}
```

### 4. Crear handler
```go
func (s *MatchingService) HandleNewFeature(w http.ResponseWriter, r *http.Request) {
    // lógica
}
```

### 5. Registrar ruta
```go
r.Post("/api/matching/new-feature", service.HandleNewFeature)
```

### 6. Agregar tests
```go
func TestNewFeature(t *testing.T) {
    // tests
}
```

## Performance

### Optimizaciones implementadas

1. **Connection pooling**
   - MaxOpenConns: 10
   - MaxIdleConns: 2

2. **Índices de BD**
   - Índices en foreign keys
   - Índices en campos de búsqueda
   - Índices en status y scores

3. **Caché en memoria**
   - TBD: Implementar caché de tutores

4. **Async processing**
   - Matches se procesan en background
   - No bloquean requests

### Benchmarks esperados

- Health check: < 5ms
- Crear solicitud: < 100ms
- Listar solicitudes: < 50ms
- Procesar matches: < 5s (background)

## Seguridad

### Implementado

1. **JWT Authentication**
   - Validar token en cada request
   - Renovación de tokens
   - JWT_SECRET en variables de entorno

2. **CORS**
   - Configurar origins permitidos
   - Headers seguros

3. **SQL Injection Prevention**
   - Prepared statements
   - Validación de entrada

4. **Rate Limiting**
   - TBD: Implementar

### TODO

- [ ] Rate limiting
- [ ] API keys
- [ ] HTTPS
- [ ] Request signing
- [ ] Audit logging

## Troubleshooting

### Error: "Connection refused"
```bash
# Verificar que PostgreSQL está corriendo
docker-compose ps

# Reiniciar contenedor
docker-compose restart postgres
```

### Error: "Database locked"
```bash
# Revisar conexiones activas
make db-shell
# SELECT * FROM pg_stat_activity;
```

### Error: "Port 8084 already in use"
```bash
# Cambiar puerto en .env
# O matar el proceso
lsof -i :8084 | grep LISTEN | awk '{print $2}' | xargs kill -9
```

### Tests failing
```bash
# Limpiar base de datos
make clean
make run

# Ejecutar tests nuevamente
make test
```

## Recursos

- [Documentación de Go](https://golang.org/doc)
- [Chi Router](https://github.com/go-chi/chi)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)
- [JWT.io](https://jwt.io/)
- [API Spec](./API_SPEC.md)

## Contribución

1. Crear rama: `git checkout -b feature/my-feature`
2. Hacer cambios y tests
3. Commit con mensajes claros
4. Push y crear PR
5. Pasar code review

## Licencia

MIT License - Ver LICENSE para detalles
