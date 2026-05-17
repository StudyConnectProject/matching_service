# Matching Service API Specification

## Overview

The Matching Service provides a REST API for managing student-tutor matching requests and recommendations.

## Base URL

```
http://localhost:8084/api/matching
```

## Authentication

All endpoints (except `/health`) require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <JWT_TOKEN>
```

## Response Format

All responses are in JSON format. Success responses include relevant data, error responses include an error message:

```json
{
  "error": "Description of the error"
}
```

## Endpoints

### Health Check

#### `GET /health`

Returns the health status of the service and database.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "message": "Matching Service is running",
  "database": "up"
}
```

---

### Match Requests

#### `POST /api/matching/request`

Create a new matching request from a student.

**Request Body:**
```json
{
  "subject": "Mathematics",
  "level": "beginner",
  "description": "Need help with algebra fundamentals",
  "preferred_schedule": "afternoon",
  "preferred_language": "Spanish",
  "modality": "virtual",
  "max_price": 50
}
```

**Parameters:**
- `subject` (string, required): Subject to learn (e.g., "Mathematics", "Physics")
- `level` (string, required): Student level - `beginner | intermediate | advanced`
- `description` (string, optional): Additional details about the request
- `preferred_schedule` (string, optional): `morning | afternoon | evening | weekend`
- `preferred_language` (string, optional): Preferred language (e.g., "English", "Spanish")
- `modality` (string, optional): `virtual | in-person | hybrid`
- `max_price` (integer, optional): Maximum hourly rate in currency units

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "student_id": "123e4567-e89b-12d3-a456-426614174000",
  "subject": "Mathematics",
  "level": "beginner",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

#### `GET /api/matching`

List all matching requests with optional filters.

**Query Parameters:**
- `status` (string, optional): Filter by status - `pending | processing | completed | rejected | cancelled`
- `subject` (string, optional): Filter by subject (case-insensitive, partial match)
- `student_id` (string, optional): Filter by student ID
- `limit` (integer, optional): Results per page (default: 20, max: 100)
- `offset` (integer, optional): Pagination offset (default: 0)

**Request:**
```
GET /api/matching?status=pending&subject=Mathematics&limit=10
```

**Response (200 OK):**
```json
{
  "requests": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "student_id": "123e4567-e89b-12d3-a456-426614174000",
      "subject": "Mathematics",
      "level": "beginner",
      "description": "Need help with algebra",
      "status": "pending",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0,
    "total": 45
  }
}
```

---

#### `GET /api/matching/{id}`

Get details of a specific matching request.

**Parameters:**
- `id` (string, required): Matching request ID (UUID)

**Request:**
```
GET /api/matching/550e8400-e29b-41d4-a716-446655440000
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "student_id": "123e4567-e89b-12d3-a456-426614174000",
  "subject": "Mathematics",
  "level": "beginner",
  "description": "Need help with algebra fundamentals",
  "status": "pending",
  "preferences": {
    "preferred_schedule": "afternoon",
    "preferred_language": "Spanish",
    "modality": "virtual",
    "max_price": 50
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Match request not found"
}
```

---

#### `PUT /api/matching/{id}`

Update the editable fields of a matching request.

**Parameters:**
- `id` (string, required): Matching request ID (UUID)

**Request Body:**
```json
{
  "subject": "Mathematics",
  "level": "intermediate",
  "description": "Updated details"
}
```

- `subject` (string, required)
- `level` (string, required): `beginner | intermediate | advanced`
- `description` (string, optional)

**Response (200 OK):**
```json
{
  "message": "Match request updated",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "subject": "Mathematics",
  "level": "intermediate"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Match request not found"
}
```

---

#### `DELETE /api/matching/{id}`

Cancel a matching request.

**Parameters:**
- `id` (string, required): Matching request ID (UUID)

**Request:**
```
DELETE /api/matching/550e8400-e29b-41d4-a716-446655440000
```

**Response (200 OK):**
```json
{
  "message": "Match request cancelled",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

#### `PATCH /api/matching/{id}/status`

Update the status of a matching request.

**Parameters:**
- `id` (string, required): Matching request ID (UUID)

**Request Body:**
```json
{
  "status": "processing"
}
```

**Response (200 OK):**
```json
{
  "message": "Status updated",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing"
}
```

---

### Match Processing

#### `POST /api/matching/process`

Run the matching algorithm over every `pending` request. For each request it
scores all available tutors (by skill relevance, bio and rating) and stores the
relevant ones as `suggested` match results.

> Note: the algorithm also runs automatically when a request is created
> (`POST /api/matching/request`), so the student sees suggestions immediately.
> This endpoint is for reprocessing requests that were still pending.

**Request:**
```
POST /api/matching/process
```

**Response (200 OK):**
```json
{
  "message": "Processing completed",
  "requests_processed": 4,
  "matches_created": 11
}
```

---

### Match Results and Recommendations

#### `GET /api/matching/recommendations/{userId}`

Get top recommended tutors for a user (student or tutor context).

**Parameters:**
- `userId` (string, required): User ID (UUID)
- `limit` (integer, optional): Number of recommendations (default: 10, max: 100)

**Request:**
```
GET /api/matching/recommendations/123e4567-e89b-12d3-a456-426614174000?limit=5
```

**Response (200 OK):**
```json
{
  "recommendations": [
    {
      "tutor_id": "789e4567-e89b-12d3-a456-426614174999",
      "name": "John Smith",
      "bio": "Experienced mathematics tutor with 5 years of experience",
      "hourly_rate": 45,
      "rating": 4.8,
      "score": 92.5,
      "skills": ["Algebra", "Geometry", "Calculus"],
      "availability": "Weekday afternoons"
    }
  ]
}
```

---

#### `GET /api/matching/{userId}/matches`

Get all active matches for a user.

**Parameters:**
- `userId` (string, required): User ID (UUID)
- `status` (string, optional): Filter by status - `suggested | accepted | rejected`

**Request:**
```
GET /api/matching/123e4567-e89b-12d3-a456-426614174000/matches?status=accepted
```

**Response (200 OK):**
```json
{
  "matches": [
    {
      "match_id": "550e8400-e29b-41d4-a716-446655440111",
      "request_id": "550e8400-e29b-41d4-a716-446655440000",
      "tutor_id": "789e4567-e89b-12d3-a456-426614174999",
      "student_id": "123e4567-e89b-12d3-a456-426614174000",
      "score": 92.5,
      "status": "accepted",
      "created_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

---

### Match Actions

#### `POST /api/matching/{id}/accept`

Accept a suggested match.

**Parameters:**
- `id` (string, required): Match result ID (UUID)

**Request:**
```
POST /api/matching/550e8400-e29b-41d4-a716-446655440111/accept
```

**Response (200 OK):**
```json
{
  "message": "Match accepted",
  "match_id": "550e8400-e29b-41d4-a716-446655440111",
  "status": "accepted"
}
```

---

#### `POST /api/matching/{id}/reject`

Reject a suggested match.

**Parameters:**
- `id` (string, required): Match result ID (UUID)

**Request:**
```
POST /api/matching/550e8400-e29b-41d4-a716-446655440111/reject
```

**Response (200 OK):**
```json
{
  "message": "Match rejected",
  "match_id": "550e8400-e29b-41d4-a716-446655440111",
  "status": "rejected"
}
```

---

#### `POST /api/matching/{id}/offer`

A tutor offers themselves for a student's match request. Creates a match
result, moves the request to `processing` and logs the action in the history.

**Parameters:**
- `id` (string, required): Match **request** ID (UUID)

**Request Body:**
```json
{
  "tutor_id": "789e4567-e89b-12d3-a456-426614174999",
  "score": 80
}
```

- `tutor_id` (string, required): ID of the tutor making the offer
- `score` (number, optional): Compatibility score 0–100 (default: 80)

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440111",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "tutor_id": "789e4567-e89b-12d3-a456-426614174999",
  "score": 80,
  "status": "suggested"
}
```

**Error Responses:**
- `404` — request not found
- `409` — request is closed, or the tutor already offered for it

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request body",
  "details": "Subject and Level are required"
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "details": "Invalid or missing JWT token"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found",
  "details": "Match request with ID xxx not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "details": "An unexpected error occurred"
}
```

---

## Rate Limiting

- **Default**: 100 requests per minute per IP
- **Authenticated**: 1000 requests per minute per user

Headers returned:
- `X-RateLimit-Limit`: Request limit
- `X-RateLimit-Remaining`: Requests remaining
- `X-RateLimit-Reset`: Reset timestamp (Unix)

---

## Pagination

List endpoints support pagination via:
- `limit`: Results per page (default: 20, max: 100)
- `offset`: Starting position (default: 0)

Response includes pagination info:
```json
{
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 150
  }
}
```

---

## WebHooks

When match events occur, notifications are sent to the configured webhook URL:

```
POST {NOTIFICATION_WEBHOOK_URL}
```

**Payload:**
```json
{
  "event": "match.created|match.accepted|match.rejected",
  "data": {
    "match_id": "...",
    "request_id": "...",
    "tutor_id": "...",
    "student_id": "...",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

---

## Examples

### Complete Workflow

1. **Create a match request:**
   ```bash
   curl -X POST http://localhost:8084/api/matching/request \
     -H "Authorization: Bearer YOUR_JWT" \
     -H "Content-Type: application/json" \
     -d '{
       "subject": "Mathematics",
       "level": "beginner",
       "description": "Help with algebra",
       "preferred_schedule": "afternoon",
       "modality": "virtual",
       "max_price": 50
     }'
   ```

2. **List pending requests:**
   ```bash
   curl http://localhost:8084/api/matching?status=pending \
     -H "Authorization: Bearer YOUR_JWT"
   ```

3. **Process matches:**
   ```bash
   curl -X POST http://localhost:8084/api/matching/process \
     -H "Authorization: Bearer YOUR_JWT"
   ```

4. **Get recommendations:**
   ```bash
   curl http://localhost:8084/api/matching/recommendations/USER_ID \
     -H "Authorization: Bearer YOUR_JWT"
   ```

5. **Accept a match:**
   ```bash
   curl -X POST http://localhost:8084/api/matching/MATCH_ID/accept \
     -H "Authorization: Bearer YOUR_JWT"
   ```
