#!/bin/bash
# Ejemplos de uso de la API del Matching Service
# Cambiar JWT_TOKEN con un token válido

JWT_TOKEN="your-jwt-token-here"
BASE_URL="http://localhost:8084"

echo "=== Matching Service API Examples ==="
echo ""

# 1. Health Check (sin autenticación)
echo "1. Verificar salud del servicio:"
curl -X GET "$BASE_URL/health" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 2. Crear una solicitud de matching
echo "2. Crear solicitud de matching:"
curl -X POST "$BASE_URL/api/matching/request" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Mathematics",
    "level": "beginner",
    "description": "Need help with algebra fundamentals",
    "preferred_schedule": "afternoon",
    "preferred_language": "Spanish",
    "modality": "virtual",
    "max_price": 50
  }'
echo -e "\n\n"

# 3. Listar solicitudes de matching
echo "3. Listar solicitudes de matching:"
curl -X GET "$BASE_URL/api/matching?status=pending&limit=10" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 4. Obtener detalles de una solicitud (reemplazar {id})
echo "4. Obtener detalles de una solicitud:"
# Necesitas reemplazar 550e8400-e29b-41d4-a716-446655440000 con un ID real
curl -X GET "$BASE_URL/api/matching/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 5. Procesar matches pendientes
echo "5. Procesar matches pendientes (async):"
curl -X POST "$BASE_URL/api/matching/process" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 6. Obtener recomendaciones de tutores
echo "6. Obtener recomendaciones de tutores:"
curl -X GET "$BASE_URL/api/matching/recommendations/123e4567-e89b-12d3-a456-426614174000?limit=5" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 7. Obtener matches activos del usuario
echo "7. Obtener matches activos del usuario:"
curl -X GET "$BASE_URL/api/matching/123e4567-e89b-12d3-a456-426614174000/matches" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 8. Aceptar un match (reemplazar {id})
echo "8. Aceptar un match:"
curl -X POST "$BASE_URL/api/matching/550e8400-e29b-41d4-a716-446655440111/accept" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 9. Rechazar un match (reemplazar {id})
echo "9. Rechazar un match:"
curl -X POST "$BASE_URL/api/matching/550e8400-e29b-41d4-a716-446655440111/reject" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

# 10. Actualizar estado de solicitud
echo "10. Actualizar estado de solicitud:"
curl -X PATCH "$BASE_URL/api/matching/550e8400-e29b-41d4-a716-446655440000/status" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "processing"
  }'
echo -e "\n\n"

# 11. Cancelar una solicitud
echo "11. Cancelar una solicitud:"
curl -X DELETE "$BASE_URL/api/matching/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n\n"

echo "=== Fin de ejemplos ==="
