#!/bin/bash
# Script para crear la estructura de directorios necesarios

mkdir -p cmd/server
mkdir -p cmd/migrate
mkdir -p internal/domain
mkdir -p internal/handler
mkdir -p internal/middleware
mkdir -p internal/repository
mkdir -p internal/service
mkdir -p pkg/config
mkdir -p pkg/database
mkdir -p migrations
mkdir -p docker

echo "✅ Estructura de directorios creada"
