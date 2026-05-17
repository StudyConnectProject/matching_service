# Quick Start Guide - Matching Service

## Instalación en 5 minutos

### Opción A: Con Docker (Recomendado) 🐳

**Requisitos:**
- Docker 20.10+
- Docker Compose 2.0+

```bash
# 1. Clonar o navegar al repositorio
cd matching_service

# 2. Copiar variables de entorno
cp .env.example .env

# 3. Iniciar servicios
docker-compose up

# 4. Verificar que funciona
curl http://localhost:8084/health
```

**Listo!** El servicio está corriendo en `http://localhost:8084`

### Opción B: Desarrollo Local 💻

**Requisitos:**
- Go 1.21+
- PostgreSQL 14+
- Make

```bash
# 1. Instalar dependencias de Go
make install-deps

# 2. Copiar variables de entorno
cp .env.example .env

# 3. Iniciar PostgreSQL (si no está corriendo)
# En Linux/macOS:
docker run -d \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:15-alpine

# 4. Ejecutar migraciones
psql -U postgres -f 001_create_initial_schema.sql

# 5. (Opcional) Cargar datos de ejemplo
psql -U postgres -f seed_data.sql

# 6. Iniciar servidor
make run-local
```

**Listo!** El servicio está corriendo en `http://localhost:8084`

---

## Primeros Pasos

### 1. Health Check

```bash
curl http://localhost:8084/health
```

Respuesta esperada:
```json
{
  "status": "healthy",
  "message": "Matching Service is running",
  "database": "up"
}
```

### 2. Crear una Solicitud de Matching

```bash
# Asume JWT_TOKEN disponible
curl -X POST http://localhost:8084/api/matching/request \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Mathematics",
    "level": "beginner",
    "description": "Need help with algebra"
  }'
```

### 3. Listar Solicitudes

```bash
curl http://localhost:8084/api/matching \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### 4. Procesar Matches

```bash
curl -X POST http://localhost:8084/api/matching/process \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

## Commandos Útiles

```bash
# Ver logs en tiempo real
make logs

# Ejecutar tests
make test

# Lint del código
make lint

# Conectar a la base de datos
make db-shell

# Ver estado de contenedores
make ps

# Detener servicios
make stop
```

---

## Troubleshooting

### Error: "Connection refused"
```bash
# Verificar que PostgreSQL está corriendo
docker-compose ps

# Reiniciar
docker-compose restart postgres
```

### Error: "Port 8084 already in use"
```bash
# Cambiar puerto en .env
APP_PORT=8085

# O matar el proceso (Linux/macOS)
lsof -i :8084 | grep LISTEN | awk '{print $2}' | xargs kill -9
```

### Error: "Database locked"
```bash
# Reiniciar toda la pila
docker-compose down
docker-compose up
```

### Los tests fallan
```bash
# Asegurar que PostgreSQL está disponible
docker-compose up postgres -d

# Ejecutar tests nuevamente
make test
```

---

## Variables de Entorno

Archivo `.env` con valores por defecto:

```env
APP_PORT=8084                    # Puerto del servidor
DB_HOST=localhost                # Host de PostgreSQL
DB_PORT=5432                     # Puerto de PostgreSQL
DB_USER=postgres                 # Usuario de BD
DB_PASSWORD=postgres             # Contraseña de BD
DB_NAME=matching_db              # Nombre de la BD
DB_POOL_MAX=10                   # Conexiones máximas
JWT_SECRET=your-secret-key       # Cambiar en producción!
NOTIFICATION_WEBHOOK_URL=...     # URL de notificaciones
LOG_LEVEL=info                   # debug, info, warn, error
```

---

## Estructura del Proyecto

```
matching_service/
├── main.go                    # Servidor principal
├── config.go                  # Configuración
├── models.go                  # Modelos de datos
├── 001_create_initial_schema.sql  # Migraciones BD
├── seed_data.sql             # Datos de ejemplo
├── Dockerfile                # Imagen Docker
├── docker-compose.yml        # Stack desarrollo
├── docker-compose.prod.yml   # Stack producción
├── Makefile                  # Comandos útiles
└── docs/
    ├── README.md             # Este archivo
    ├── DEVELOPMENT.md        # Guía de desarrollo
    ├── API_SPEC.md          # Especificación de API
    ├── ARCHITECTURE.md      # Documentación de arquitectura
    └── ROADMAP.md           # Roadmap del proyecto
```

---

## Arquitectura General

```
┌──────────────────┐
│   API Gateway    │  (Java Spring Cloud)
└────────┬─────────┘
         │ /api/matching/**
         ↓
┌──────────────────┐
│  Matching Service │  (Go) ← TÚ ESTÁS AQUÍ
├──────────────────┤
│ - REST API
│ - PostgreSQL
│ - Event processing
└────────┬─────────┘
         │
    ┌────┴──────┐
    ↓           ↓
┌────────┐  ┌──────────────┐
│  Auth  │  │Notification  │
│Service │  │Service       │
└────────┘  └──────────────┘
```

---

## Siguiente Pasos

1. **Leer documentación:**
   - [API_SPEC.md](API_SPEC.md) - Endpoints disponibles
   - [DEVELOPMENT.md](DEVELOPMENT.md) - Guía de desarrollo
   - [ARCHITECTURE.md](ARCHITECTURE.md) - Diseño técnico

2. **Explorar ejemplos:**
   - Ver `examples.sh` para ejemplos de curl

3. **Escribir código:**
   - Crear feature branch: `git checkout -b feature/my-feature`
   - Hacer cambios
   - Ejecutar tests: `make test`
   - Enviar PR

4. **Producción:**
   - Ver [docker-compose.prod.yml](docker-compose.prod.yml)
   - Configurar variables de entorno seguras
   - Deploy a Kubernetes o similar

---

## Contacto y Soporte

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Documentation**: Revisar archivos .md

---

## Licencia

MIT License - Ver [LICENSE](LICENSE)

---

**¿Necesitas ayuda?** Revisa [DEVELOPMENT.md](DEVELOPMENT.md) o abre un issue en GitHub.
