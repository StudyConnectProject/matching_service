# Contribuyendo al Matching Service

¡Gracias por tu interés en contribuir! Sigue estas guías para hacer el proceso lo más fácil posible.

## Proceso General

1. **Fork** el repositorio
2. **Crea una rama** para tu feature: `git checkout -b feature/amazing-feature`
3. **Haz cambios** y **escribe tests**
4. **Asegúrate que los tests pasen**: `make test`
5. **Ejecuta lint**: `make lint`
6. **Commit con mensajes claros**: `git commit -m 'Add amazing feature'`
7. **Push a tu rama**: `git push origin feature/amazing-feature`
8. **Abre un Pull Request**

## Estándares de Código

### Go Style Guide

- Sigue el [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Usa `gofmt` para formatear: `make fmt`
- Nombres de variables descriptivos
- Funciones pequeñas (máximo 30 líneas)
- Comentarios en funciones públicas

### Estructura de Código

```go
// Comentario de función pública
func (s *MatchingService) PublicFunction(param string) error {
    // implementación
    return nil
}

// comentario de función privada
func (s *MatchingService) privateHelper() {
    // implementación
}
```

### Error Handling

```go
// ✅ Bueno
result, err := db.QueryRow(query).Scan(&value)
if err != nil {
    log.Printf("error querying database: %v", err)
    return fmt.Errorf("failed to get record: %w", err)
}

// ❌ Evitar
_ = db.QueryRow(query).Scan(&value)  // Ignorar errores
```

## Testing

### Requisitos para PRs

- Mínimo 70% de cobertura
- Tests para funciones públicas
- Tests de error cases
- Tests de integración para endpoints

### Estructura de Tests

```go
func TestFeatureName(t *testing.T) {
    // Arrange
    expected := "some value"
    svc := setupService()
    
    // Act
    result := svc.DoSomething()
    
    // Assert
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

## Convenciones de Commit

```
<tipo>: <descripción corta>

<descripción detallada opcional>

Fixes #123
```

### Tipos de Commit

- `feat`: Nueva feature
- `fix`: Correción de bug
- `docs`: Cambios en documentación
- `style`: Cambios de formato (sin código)
- `refactor`: Refactorización de código
- `test`: Agregar o actualizar tests
- `chore`: Cambios en dependencias o config

### Ejemplos

```
feat: add recommendation algorithm

Implement the tutor recommendation algorithm with weighted scoring
based on skills, availability, rating, and price.

Implements part of #42
```

```
fix: handle database connection timeout

Add proper timeout handling for database connections and implement
connection pooling correctly.

Fixes #156
```

## PR Checklist

- [ ] Tests agregados/actualizados
- [ ] Código formateado con `make fmt`
- [ ] Lint pasa con `make lint`
- [ ] Tests pasan con `make test`
- [ ] Documentación actualizada
- [ ] No hay secrets en el código
- [ ] Commits tienen mensajes descriptivos
- [ ] PR tiene descripción clara

## Areas de Contribución

### High Priority

- [ ] Implementar autenticación JWT completa
- [ ] Algoritmo de matching optimizado
- [ ] Caché en memoria
- [ ] Rate limiting
- [ ] Metricas y monitoring

### Medium Priority

- [ ] Mejorar manejo de errores
- [ ] Documentación de API completa
- [ ] Tests de integración
- [ ] Logging estructurado
- [ ] Validación de input avanzada

### Low Priority

- [ ] Dashboard de admin
- [ ] CLI tools
- [ ] Ejemplos adicionales
- [ ] Documentación en otros idiomas

## Preguntas?

- Abre una **Issue** para preguntas o bugs
- Discute ideas nuevas antes de implementar
- Revisa los **Discussions** existentes

## Licencia

Al contribuir, aceptas que tu código será licenciado bajo MIT.

---

¡Esperamos tu contribución! 🚀
