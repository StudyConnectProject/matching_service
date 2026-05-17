# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto se adhiere al [Versionado Semántico](https://semver.org/lang/es/).

## [Unreleased]

### Added
- Estructura base del proyecto Go
- Modelos de dominio para matching
- Handlers REST API básicos
- Configuración con variables de entorno
- Docker y docker-compose para desarrollo
- Migraciones de base de datos PostgreSQL
- Documentación de arquitectura
- Documentación de API
- Tests unitarios básicos

### Todo
- [ ] Implementar autenticación JWT completa
- [ ] Algoritmo de matching avanzado
- [ ] Caché en memoria para tutores
- [ ] Rate limiting
- [ ] Monitoreo y métricas
- [ ] Logging estructurado
- [ ] Integration tests
- [ ] Load testing
- [ ] Optimizaciones de performance

## [1.0.0] - TBD

### Added
- Inicio del proyecto

---

## Formato de Secciones

- **Added** para nuevas funcionalidades
- **Changed** para cambios en funcionalidades existentes
- **Deprecated** para funcionalidades pronto a ser removidas
- **Removed** para funcionalidades removidas
- **Fixed** para correcciones de bugs
- **Security** para vulnerabilidades

## Cómo Reportar Cambios

1. Crear una rama feature
2. Implementar cambios
3. Actualizar CHANGELOG.md
4. Incluir versión y fecha si es release
5. Crear PR

### Ejemplo de Actualización

```markdown
## [1.0.1] - 2024-01-20

### Fixed
- Corrección de error en el cálculo de scores (#42)
- Manejo de timeout en conexión a BD

### Added
- Validación adicional de entrada en requests
```

## Versionado

Usamos [Semantic Versioning](https://semver.org/):
- MAJOR (X.0.0): Cambios incompatibles
- MINOR (0.X.0): Nuevas características compatibles
- PATCH (0.0.X): Correcciones de bugs
