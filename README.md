# Matching Service - StudyConnect

Microservicio de emparejamiento (matching) entre estudiantes y tutores para la plataforma StudyConnect.

**Estado del Proyecto**: 🟢 [Ver PROJECT_STATUS.md](PROJECT_STATUS.md)

## Características

- ✅ Emparejamiento asincrónico basado en eventos
- ✅ Algoritmo de scoring avanzado
- ✅ Respuesta en < 200ms (95% de los casos)
- ✅ Procesamiento de eventos en < 5 segundos
- ✅ Integración con servicios de autenticación y notificaciones
- ✅ API REST con validación JWT
- ✅ Base de datos PostgreSQL

## Stack Técnico

- **Lenguaje**: Go 1.21
- **Router**: Chi v5
- **Base de Datos**: PostgreSQL
- **Logging**: Logrus
- **Autenticación**: JWT
- **Containerización**: Docker & Docker Compose

## Configuración Rápida

### Requisitos Previos

- Go 1.21+
- PostgreSQL 14+
- Docker & Docker Compose

### Variables de Entorno

Copia `.env.example` a `.env`:

```bash
APP_PORT=8084
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=matching_db
DB_POOL_MAX=10
JWT_SECRET=your-secret-key-change-in-production
NOTIFICATION_WEBHOOK_URL=http://notification-service/internal/events
LOG_LEVEL=info
```

### Instalación Local

```bash
# Instalar dependencias
go mod download

# Ejecutar migraciones de BD
go run cmd/migrate/main.go

# Iniciar servidor
go run cmd/server/main.go
```

### Con Docker Compose

```bash
docker-compose up -d
```

El servicio estará disponible en `http://localhost:8084`

## API Endpoints

### Solicitudes de Emparejamiento

- `POST   /api/matching/request` - Crear nueva solicitud
- `GET    /api/matching` - Listar solicitudes
- `GET    /api/matching/{id}` - Detalle de solicitud
- `DELETE /api/matching/{id}` - Cancelar solicitud
- `PATCH  /api/matching/{id}/status` - Actualizar estado

### Procesamiento

- `POST   /api/matching/process` - Procesar matches pendientes

### Recomendaciones

- `GET    /api/matching/recommendations/{userId}` - Top tutores recomendados
- `GET    /api/matching/{userId}/matches` - Matches activos del usuario

### Acciones en Matches

- `POST   /api/matching/{id}/accept` - Aceptar match
- `POST   /api/matching/{id}/reject` - Rechazar match

### Sistema

- `GET    /health` - Health check

## Estructura del Proyecto

```
matching_service/
├── cmd/
│   ├── server/
│   │   └── main.go              # Punto de entrada del servidor
│   └── migrate/
│       └── main.go              # Herramienta de migraciones
├── internal/
│   ├── domain/                  # Modelos de dominio
│   ├── handler/                 # Handlers HTTP
│   ├── middleware/              # Middlewares (JWT, logging, etc)
│   ├── repository/              # Acceso a datos
│   └── service/                 # Lógica de negocio
├── pkg/
│   ├── config/                  # Configuración
│   └── database/                # Conexión a BD
├── migrations/                  # Migraciones SQL
├── docker/
│   ├── Dockerfile               # Imagen del servicio
│   └── entrypoint.sh            # Script de entrada
├── docker-compose.yml           # Stack local
├── .env.example                 # Variables de entorno
├── go.mod                       # Dependencias Go
└── README.md                    # Este archivo
```

## Algoritmo de Matching

El servicio usa un algoritmo de scoring ponderado:

```
Score = (skill_match * 0.4) + (availability_match * 0.3) + 
        (rating_score * 0.2) + (price_match * 0.1)
```

Donde:
- **skill_match**: % de habilidades del tutor que coinciden con la solicitud
- **availability_match**: % de disponibilidad compatible con preferencias
- **rating_score**: Rating del tutor normalizado (0-100)
- **price_match**: Coincidencia de precio (0-100)

## Testing

```bash
# Ejecutar tests
go test ./...

# Con cobertura
go test -cover ./...

# Verbose
go test -v ./...
```

## Documentación Adicional

- [Modelo de BD](docs/database-schema.md)
- [Especificación de API](docs/api-spec.md)
- [Guía de Desarrollo](docs/development.md)

## Contribución

Por favor lee [CONTRIBUTING.md](CONTRIBUTING.md) antes de hacer cambios.

## Licencia

MIT License - Ver [LICENSE](LICENSE) para detalles
