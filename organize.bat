@echo off
REM Estructura de carpetas para Matching Service
REM Ejecutar desde la raíz del proyecto: organize.bat

echo Creating directory structure...

REM Crear directorios principales
if not exist "cmd" mkdir cmd
if not exist "cmd\server" mkdir cmd\server
if not exist "cmd\migrate" mkdir cmd\migrate

if not exist "internal" mkdir internal
if not exist "internal\domain" mkdir internal\domain
if not exist "internal\handler" mkdir internal\handler
if not exist "internal\middleware" mkdir internal\middleware
if not exist "internal\repository" mkdir internal\repository
if not exist "internal\service" mkdir internal\service

if not exist "pkg" mkdir pkg
if not exist "pkg\config" mkdir pkg\config
if not exist "pkg\database" mkdir pkg\database

if not exist "migrations" mkdir migrations
if not exist "docker" mkdir docker
if not exist "tests" mkdir tests
if not exist "tests\integration" mkdir tests\integration
if not exist "tests\unit" mkdir tests\unit

echo.
echo Directory structure created!
echo.
echo Next steps to reorganize files:
echo.
echo 1. Move main.go and config.go to cmd/server/
echo    BEFORE: main.go, config.go
echo    AFTER:  cmd/server/main.go, cmd/server/config.go
echo.
echo 2. Move models.go to internal/domain/
echo    BEFORE: models.go
echo    AFTER:  internal/domain/models.go
echo.
echo 3. Create handler methods in internal/handler/
echo    CREATE: internal/handler/matching.go
echo    CREATE: internal/handler/health.go
echo.
echo 4. Create middleware in internal/middleware/
echo    CREATE: internal/middleware/auth.go
echo    CREATE: internal/middleware/cors.go
echo    CREATE: internal/middleware/logger.go
echo.
echo 5. Create repository in internal/repository/
echo    CREATE: internal/repository/repository.go (interface)
echo    CREATE: internal/repository/postgres.go (implementation)
echo.
echo 6. Create service in internal/service/
echo    CREATE: internal/service/matching.go
echo.
echo 7. Move SQL files to migrations/
echo    BEFORE: 001_create_initial_schema.sql, seed_data.sql
echo    AFTER:  migrations/001_create_initial_schema.sql
echo    AFTER:  migrations/seed_data.sql
echo.
echo 8. Move Docker files to docker/
echo    BEFORE: Dockerfile
echo    AFTER:  docker/Dockerfile
echo.
echo 9. Move tests to tests/
echo    BEFORE: main_test.go
echo    AFTER:  tests/unit/main_test.go
echo.
echo Done!
