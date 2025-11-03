.PHONY: all infra gateway auth users achievement down clean

help:
	@echo "Available commands:"
	@echo "  make all      - Start all services"
	@echo "  make service  - Start selected service selected from minio, haproxy, API-Gateway, Auth, Users, achinement_service"
	@echo "  make down     - Stop all services" 
	@echo "  make logs     - Show logs in selected service"
	@echo "  make status   - Show service status"

# Запуск всего
all: haproxy minio gateway auth users achievement

haproxy:
	cd devops && docker-compose -f haproxy-compose.yml up -d

minio:
	cd devops && docker-compose -f minio-compose.yml up -d

gateway:
	cd API-Gateway && docker-compose -f api-gateway-compose.yml up -d

auth:
	cd Auth && docker-compose -f auth-compose.yml up -d

users:
	cd Users && docker-compose -f user-compose.yml up -d

achievement:
	cd achievement_service && docker-compose -f achinement_service/achivement-compose.yml up -d

# Управление
down:
	docker-compose -f API-Gateway/api-gateway-compose.yml down
	docker-compose -f Auth/auth-compose.yml down
	docker-compose -f Users/user-compose.yml down
	docker-compose -f achinement_service/achivement-compose.yml down
	docker-compose -f devops/docker-compose.core.yml down

logs:
	# Логи конкретного сервиса
	docker-compose -f $(service)/docker-compose.yml logs -f

status:
	@echo "=== Service Status ==="
	@for service in minio haproxy API-Gateway Auth Users achinement_service; do \
		if [ -f "$$service/docker-compose.yml" ]; then \
			echo "$$service:"; \
			docker-compose -f $$service/docker-compose.yml ps; \
			echo; \
		fi; \
	done