#!/bin/bash

# Configurações do teste
DURATION="1m"          # Duração do teste
CONCURRENT=50          # Número de requisições concorrentes
TOTAL_REQUESTS=1000   # Número total de requisições
API_URL="http://localhost:8080"  # URL da API

echo "Iniciando teste de carga..."
echo "Duração: $DURATION"
echo "Requisições concorrentes: $CONCURRENT"
echo "Total de requisições: $TOTAL_REQUESTS"
echo "URL da API: $API_URL"
echo ""

# Teste de carga normal
echo "Testando endpoint /load..."
hey -z $DURATION -c $CONCURRENT -n $TOTAL_REQUESTS $API_URL/load

# Teste de erros
echo "Testando endpoint /error..."
hey -z $DURATION -c $CONCURRENT -n $TOTAL_REQUESTS $API_URL/error

echo ""
echo "Teste de carga concluído!" 