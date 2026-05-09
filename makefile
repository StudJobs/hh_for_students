.PHONY: all infra gateway auth users achievement skills es search microtasks down clean help reindex obs obs-down loadtest

help:
	@echo "Available commands:"
	@echo "  make all      - Start all services"
	@echo "  make service  - Start selected service selected from minio, haproxy, API-Gateway, Auth, Users, achievement, company, vacancy, skills"
	@echo "  make down     - Stop all services"
	@echo "  make logs     - Show logs in selected service"
	@echo "  make status   - Show service status"

# Запуск всего в правильном порядке
all: haproxy minio es auth users achievement company vacancy skills search microtasks gateway

# Elasticsearch — нужен Search-сервису и индексаторам в Users/Vacancy
es:
	cd devops && docker-compose -f elasticsearch-compose.yml up -d
	@echo "Waiting for Elasticsearch (this can take ~30s on first start)..."
	@timeout 60 bash -c 'until curl -fs http://localhost:9200/_cluster/health 2>/dev/null | grep -E "(green|yellow)" >/dev/null; do sleep 3; echo "Waiting for Elasticsearch..."; done'
	@echo "✓ Elasticsearch is healthy!"

# Инфраструктурные сервисы
#haproxy:
#	cd devops && docker-compose -f haproxy-compose.yml up -d
#	@echo "Waiting for HAProxy..."
#	@timeout 30 bash -c 'until nc -z localhost 80 || nc -z localhost 443 || nc -z localhost 8443; do sleep 2; echo "Waiting for HAProxy..."; done'
#	@echo "✓ HAProxy is healthy!"
#
#minio: haproxy
#	cd devops && docker-compose -f minio-compose.yml up -d
#	@echo "Waiting for MinIO..."
#	@timeout 30 bash -c 'until curl -f http://localhost:9000/minio/health/live >/dev/null 2>&1; do sleep 2; echo "Waiting for MinIO..."; done'
#	@echo "✓ MinIO is healthy!"

# Микросервисы с зависимостями и проверками через grpcurl
auth:
	cd Auth && docker-compose -f auth-compose.yml up -d
	@echo "Waiting for auth service..."
	@timeout 10 bash -c 'until ./grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for auth..."; done'
	@echo "✓ Auth service is healthy!"

users: auth
	cd Users && docker-compose -f user-compose.yml up -d
	@echo "Waiting for users service..."
	@timeout 10 bash -c 'until ./grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for users..."; done'
	@echo "✓ Users service is healthy!"

achievement: users
	cd Achievements && docker-compose -f achieve-compose.yml up -d
	@echo "Waiting for achievement service..."
	@timeout 10 bash -c 'until ./grpcurl -plaintext localhost:50053 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for achievement..."; done'
	@echo "✓ Achievement service is healthy!"


vacancy:
	cd Vacancy && docker-compose -f vacancy-compose.yml up -d
	@echo "Waiting for vacancy service..."
	@timeout 10 bash -c 'until ./grpcurl -plaintext localhost:50054 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for vacancy..."; done'
	@echo "✓ Vacancy service is healthy!"

company:
	cd Company && docker-compose -f company-compose.yml up -d
	@echo "Waiting for company service..."
	@timeout 10 bash -c 'until ./grpcurl -plaintext localhost:50055 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for company..."; done'
	@echo "✓ Company service is healthy!"

skills:
	cd Skills && docker-compose -f skills-compose.yml up -d
	@echo "Waiting for skills service..."
	@timeout 10 bash -c 'until ./grpcurl -plaintext localhost:50056 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for skills..."; done'
	@echo "✓ Skills service is healthy!"

search: es users vacancy
	cd Search && docker-compose -f search-compose.yml up -d
	@echo "Waiting for search service..."
	@timeout 60 bash -c 'until ./grpcurl -plaintext localhost:50057 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for search..."; done'
	@echo "✓ Search service is healthy!"

microtasks: search
	cd MicroTasks && docker-compose -f microtasks-compose.yml up -d
	@echo "Waiting for microtasks service..."
	@timeout 30 bash -c 'until ./grpcurl -plaintext localhost:50058 grpc.health.v1.Health/Check >/dev/null 2>&1; do sleep 2; echo "Waiting for microtasks..."; done'
	@echo "✓ MicroTasks service is healthy!"

# Observability — Prometheus + Grafana. Сначала поднимаются основные сервисы (make all),
# затем `make obs` подцепляется к той же microservices-net и начинает scrape /metrics.
obs:
	cd devops && docker-compose -f observability-compose.yml up -d
	@echo "✓ Observability stack up"
	@echo "  Prometheus → http://localhost:9090"
	@echo "  Grafana    → http://localhost:3001 (anon Viewer / admin:admin)"

obs-down:
	cd devops && docker-compose -f observability-compose.yml down

# Нагрузочное тестирование через k6. Скрипт прогоняет 3 сценария (search/users/tasks).
# Требует установленный k6 (`brew install k6`) и поднятый стек (`make all`).
# Результаты — в стандартный вывод, summary с p50/p95/p99 и RPS.
loadtest:
	@if ! command -v k6 >/dev/null 2>&1; then \
		echo "k6 not installed. Install with: brew install k6"; exit 1; \
	fi
	k6 run devops/k6/loadtest.js

# Холодная переиндексация PG → ES (вызывается после миграций или для первого старта).
# Reindex RPC создаёт индексы (recreate=true) и перечитывает все профили/вакансии.
reindex:
	@echo "Reindexing all profiles and vacancies into Elasticsearch..."
	./grpcurl -plaintext -d '{"recreate_indices": true}' localhost:50057 search.v1.SearchService/Reindex
	@echo "✓ Reindex done"

gateway: auth vacancy skills search microtasks
	cd API-Gateway && docker-compose -f api-gateway-compose.yml up -d
	@echo "Waiting for gateway service..."
	@timeout 10 bash -c 'until curl -f http://localhost:8000/health >/dev/null 2>&1; do sleep 2; echo "Waiting for gateway..."; done'
	@echo "✓ Gateway service is healthy!"

# Управление
down:
	@echo "Stopping all services..."
	docker-compose -f API-Gateway/api-gateway-compose.yml down
	docker-compose -f Auth/auth-compose.yml down
	docker-compose -f Users/user-compose.yml down
	docker-compose -f Achievements/achivement-compose.yml down
	docker-compose -f devops/haproxy-compose.yml down
	docker-compose -f devops/minio-compose.yml down
	docker-compose -f devops/elasticsearch-compose.yml down
	docker-compose -f Company/company-compose.yml down
	docker-compose -f vacancy-service/vacancy-compose.yml down
	docker-compose -f Skills/skills-compose.yml down
	docker-compose -f Search/search-compose.yml down
	docker-compose -f MicroTasks/microtasks-compose.yml down
	@echo "✓ All services stopped"

logs:
	# Логи конкретного сервиса
	@if [ -z "$(service)" ]; then \
		echo "Usage: make logs service=<service_name>"; \
		echo "Available services: minio, haproxy, API-Gateway, Auth, Users, achievement, company, vacancy"; \
	else \
		docker-compose -f $(service)/docker-compose.yml logs -f; \
	fi

status:
	@echo "=== Service Status ==="
	@for service in devops API-Gateway Auth Users Achievements Company vacancy-service Skills; do \
		if [ -f "$$service/docker-compose.yml" ] || [ -f "$$service/*-compose.yml" ]; then \
			echo "--- $$service ---"; \
			if [ "$$service" = "devops" ]; then \
				docker-compose -f $$service/haproxy-compose.yml ps 2>/dev/null || true; \
				docker-compose -f $$service/minio-compose.yml ps 2>/dev/null || true; \
			else \
				docker-compose -f $$service/*-compose.yml ps 2>/dev/null || true; \
			fi; \
			echo; \
		fi; \
	done

# Дополнительные утилиты
restart: down all
	@echo "✓ All services restarted"

clean: down
	@echo "Cleaning up..."
	docker system prune -f
	@echo "✓ Cleanup completed"

# Установка grpcurl
setup-grpcurl:
	@if [ ! -f "./grpcurl" ]; then \
		echo "Downloading grpcurl..."; \
		wget -q https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_linux_x86_64.tar.gz -O grpcurl.tar.gz; \
		tar -xzf grpcurl.tar.gz; \
		rm grpcurl.tar.gz; \
		chmod +x grpcurl; \
		echo "✓ grpcurl installed"; \
	else \
		echo "✓ grpcurl already installed"; \
	fi

# Предварительная установка зависимостей
deps: setup-grpcurl