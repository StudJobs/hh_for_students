#!/bin/bash
set -e

echo "🚀 Starting infrastructure services..."

docker-compose -f devops/docker-compose.core.yml up -d

services=(
    "APIGateway:API-Gateway/api-gateway-compose.yml"
    "Auth:Auth/auth-compose.yml" 
    "Users:Users/user-compose.yml"
    "Achievement:achinement_service/achivement-compose.yml"
    "Haproxy:devops/haproxy-compose.yml"
    "Minio:devops/minio-compose.yml"
)

declare -A pids

for service_pair in "${services[@]}"; do
    IFS=':' read -r service_name compose_file <<< "$service_pair"
    
    echo "Starting $service_name..."
    docker-compose -f "$compose_file" up -d &
    pids[$service_name]=$!
done

# 3. Ждем завершения всех запусков
for service_name in "${!pids[@]}"; do
    wait ${pids[$service_name]}
    echo "✅ $service_name started"
done

for service_pair in "${services[@]}"; do
    IFS=':' read -r service_name compose_file <<< "$service_pair"
    
    echo "Checking health of $service_name..."
    container_id=$(docker-compose -f "$compose_file" ps -q | head -1)
    
    for i in {1..30}; do
        health=$(docker inspect --format='{{.State.Health.Status}}' "$container_id" 2>/dev/null || echo "starting")
        
        case $health in
            "healthy")
                echo "✅ $service_name is healthy"
                break
                ;;
            "unhealthy")
                echo "❌ $service_name is unhealthy"
                docker-compose -f "$compose_file" logs
                exit 1
                ;;
            *)
                if [ $i -eq 30 ]; then
                    echo "⏰ Timeout waiting for $service_name"
                    docker-compose -f "$compose_file" logs
                    exit 1
                fi
                sleep 5
                ;;
        esac
    done
done

echo "All services deployed successfully!"