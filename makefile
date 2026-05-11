# StudJobs · Makefile
# Все таргеты — phony (HFS+/APFS case-insensitive: каталоги Vacancy/, Company/
# конфликтуют с lowercase-таргетами).
# Хелсчек — curl /health на metrics-порт каждого сервиса (HTTP отвечает только
# когда процесс реально стартовал, в отличие от `nc -z` на macOS Docker Desktop,
# где Docker proxy слушает порт до того, как контейнер готов).
# .env подхватывается через --env-file ../.env, т.к. каждый сервис запускается
# из своего каталога и теряет родительский .env.

.PHONY: all help \
        redis es haproxy minio \
        auth users achievement vacancy company skills search microtasks gateway \
        obs obs-down loadtest reindex \
        down wipe stop start soft-restart logs status restart clean setup-grpcurl deps

ENVFILE := --env-file $(CURDIR)/.env

help:
	@echo "── StudJobs · Makefile ───────────────────────────────────────"
	@echo "Запуск:"
	@echo "  make all           — поднять весь стек (первый раз / после wipe)"
	@echo "  make obs           — Prometheus + Grafana"
	@echo ""
	@echo "Мягкое управление (данные сохраняются):"
	@echo "  make stop          — остановить контейнеры (volumes на месте)"
	@echo "  make start         — поднять обратно остановленные"
	@echo "  make soft-restart  — stop + start (быстрый рестарт без потери БД)"
	@echo ""
	@echo "Жёсткое управление (DESTRUCTIVE — удаляет volumes и данные):"
	@echo "  make down          — снос всего + удаление volumes"
	@echo "  make wipe          — алиас на down"
	@echo "  make restart       — down + all (свежий старт с нуля)"
	@echo ""
	@echo "Диагностика:"
	@echo "  make status                     — список запущенных контейнеров"
	@echo "  make logs service=<container>   — tail логов конкретного контейнера"
	@echo ""
	@echo "Тестирование:"
	@echo "  make loadtest      — k6 нагрузочный прогон"
	@echo "  make reindex       — холодная переиндексация PG → ES"

# Запуск всего в правильном порядке.
# HAProxy сознательно не в зависимостях — на локалке мы ходим в API-Gateway напрямую
# через :8000 без TLS-терминации. Если нужен HAProxy — `make haproxy` отдельно.
all: minio es redis auth users achievement company vacancy skills search microtasks gateway

# Redis для cache-aside в API-Gateway. Должен подняться до gateway, иначе тот стартует
# с no-op кэшом (см. main.go::cacheClient.Ping).
redis:
	cd devops && docker-compose $(ENVFILE) -f redis-compose.yml up -d
	@echo "Waiting for Redis..."
	@i=0; until docker exec studjobs_redis redis-cli ping 2>/dev/null | grep -q PONG; do \
		[ $$i -ge 15 ] && echo "✗ Redis timeout" && exit 1; \
		i=$$((i+1)); sleep 1; \
	done
	@echo "✓ Redis is healthy!"

# Elasticsearch — нужен Search-сервису и индексаторам в Users/Vacancy
es:
	cd devops && docker-compose $(ENVFILE) -f elasticsearch-compose.yml up -d
	@echo "Waiting for Elasticsearch (this can take ~30s on first start)..."
	@i=0; until curl -fs http://localhost:9200/_cluster/health 2>/dev/null | grep -E "(green|yellow)" >/dev/null; do \
		[ $$i -ge 20 ] && echo "✗ Elasticsearch timeout" && exit 1; \
		i=$$((i+1)); echo "  Waiting for Elasticsearch..."; sleep 3; \
	done
	@echo "✓ Elasticsearch is healthy!"

# HAProxy опционален: TLS-терминация для прод-демо. На локалке не нужен.
haproxy:
	cd devops && docker-compose $(ENVFILE) -f haproxy-compose.yml up -d
	@echo "Waiting for HAProxy..."
	@i=0; until nc -z localhost 80 || nc -z localhost 443 || nc -z localhost 8443; do \
		[ $$i -ge 15 ] && echo "✗ HAProxy timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ HAProxy is healthy!"

# MinIO — S3-совместимое хранилище для файлов ачивок (Achievements-сервис ходит через minio:9000).
minio:
	cd devops && docker-compose $(ENVFILE) -f minio-compose.yml up -d
	@echo "Waiting for MinIO..."
	@i=0; until curl -f http://localhost:9000/minio/health/live >/dev/null 2>&1; do \
		[ $$i -ge 15 ] && echo "✗ MinIO timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ MinIO is healthy!"

# Микросервисы.
# Хелсчек — /health на metrics-порту (9092..9099). Это HTTP-endpoint, который
# `metrics.ServeMetrics(...)` поднимает синхронно при старте main(). Если он отвечает —
# процесс реально живёт; в отличие от `nc -z` Docker Desktop'а, который врёт.

auth:
	cd Auth && docker-compose $(ENVFILE) -f auth-compose.yml up -d
	@echo "Waiting for auth service..."
	@i=0; until curl -fs http://localhost:9092/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ Auth timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Auth service is healthy!"

users: auth
	cd Users && docker-compose $(ENVFILE) -f user-compose.yml up -d
	@echo "Waiting for users service..."
	@i=0; until curl -fs http://localhost:9093/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ Users timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Users service is healthy!"

achievement: users
	cd Achievements && docker-compose $(ENVFILE) -f achieve-compose.yml up -d
	@echo "Waiting for achievement service..."
	@i=0; until curl -fs http://localhost:9094/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ Achievement timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Achievement service is healthy!"

vacancy:
	cd Vacancy && docker-compose $(ENVFILE) -f vacancy-compose.yml up -d
	@echo "Waiting for vacancy service..."
	@i=0; until curl -fs http://localhost:9095/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ Vacancy timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Vacancy service is healthy!"

company:
	cd Company && docker-compose $(ENVFILE) -f company-compose.yml up -d
	@echo "Waiting for company service..."
	@i=0; until curl -fs http://localhost:9096/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ Company timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Company service is healthy!"

skills:
	cd Skills && docker-compose $(ENVFILE) -f skills-compose.yml up -d
	@echo "Waiting for skills service..."
	@i=0; until curl -fs http://localhost:9097/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ Skills timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Skills service is healthy!"

search: es users vacancy
	cd Search && docker-compose $(ENVFILE) -f search-compose.yml up -d
	@echo "Waiting for search service..."
	@i=0; until curl -fs http://localhost:9098/health >/dev/null 2>&1; do \
		[ $$i -ge 60 ] && echo "✗ Search timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Search service is healthy!"

microtasks: search
	cd MicroTasks && docker-compose $(ENVFILE) -f microtasks-compose.yml up -d
	@echo "Waiting for microtasks service..."
	@i=0; until curl -fs http://localhost:9099/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ MicroTasks timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ MicroTasks service is healthy!"

gateway: auth vacancy skills search microtasks redis company achievement
	cd API-Gateway && docker-compose $(ENVFILE) -f api-gateway-compose.yml up -d
	@echo "Waiting for gateway service..."
	@i=0; until curl -fs http://localhost:9091/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "✗ Gateway timeout" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ Gateway service is healthy! (Fiber на :8000, metrics на :9091)"

# Observability — Prometheus + Grafana. Сначала поднимаются основные сервисы (make all),
# затем `make obs` подцепляется к той же microservices-net и начинает scrape /metrics.
obs:
	cd devops && docker-compose $(ENVFILE) -f observability-compose.yml up -d
	@echo "✓ Observability stack up"
	@echo "  Prometheus → http://localhost:9090"
	@echo "  Grafana    → http://localhost:3001 (anon Viewer / admin:admin)"

obs-down:
	cd devops && docker-compose -f observability-compose.yml down

# Нагрузочное тестирование через k6.
loadtest:
	@if ! command -v k6 >/dev/null 2>&1; then \
		echo "k6 not installed. Install with: brew install k6"; exit 1; \
	fi
	k6 run devops/k6/loadtest.js

# Холодная переиндексация PG → ES (вызывается после миграций или для первого старта).
# Требует grpcurl. На macOS можно поставить через `brew install grpcurl`.
reindex:
	@if ! command -v grpcurl >/dev/null 2>&1 && [ ! -f "./grpcurl" ]; then \
		echo "grpcurl не найден. Установи через brew install grpcurl или make setup-grpcurl"; exit 1; \
	fi
	@echo "Reindexing all profiles and vacancies into Elasticsearch..."
	@if command -v grpcurl >/dev/null 2>&1; then \
		grpcurl -plaintext -d '{"recreate_indices": true}' localhost:50057 search.v1.SearchService/Reindex; \
	else \
		./grpcurl -plaintext -d '{"recreate_indices": true}' localhost:50057 search.v1.SearchService/Reindex; \
	fi
	@echo "✓ Reindex done"

# Управление
# DESTRUCTIVE: down останавливает контейнеры И удаляет volumes (Postgres, ES, Redis,
# MinIO). Все зарегистрированные юзеры/ачивки/вакансии пропадут. Это сознательный выбор
# для dev-режима: при перезапуске стек гарантированно стартует с чистой инициализацией
# и Postgres получает свежий DB_PASS.
#
# ВАЖНО: `cd <dir> && docker-compose ...` — project name = имя каталога. Если запускать
# `docker-compose -f <dir>/compose.yml` из корня, проект будет `hh_for_students` (имя
# корня), и docker не найдёт ранее созданные volumes (они в проектах `auth`, `users`...).
down:
	@echo "⚠ Stopping all services and DELETING volumes (data will be lost)..."
	-cd devops && docker-compose -f redis-compose.yml down -v 2>/dev/null
	-cd devops && docker-compose -f elasticsearch-compose.yml down -v 2>/dev/null
	-cd devops && docker-compose -f minio-compose.yml down -v 2>/dev/null
	-cd devops && docker-compose -f haproxy-compose.yml down -v 2>/dev/null
	-cd devops && docker-compose -f observability-compose.yml down -v 2>/dev/null
	-cd API-Gateway && docker-compose -f api-gateway-compose.yml down -v 2>/dev/null
	-cd Auth && docker-compose -f auth-compose.yml down -v 2>/dev/null
	-cd Users && docker-compose -f user-compose.yml down -v 2>/dev/null
	-cd Achievements && docker-compose -f achieve-compose.yml down -v 2>/dev/null
	-cd Company && docker-compose -f company-compose.yml down -v 2>/dev/null
	-cd Vacancy && docker-compose -f vacancy-compose.yml down -v 2>/dev/null
	-cd Skills && docker-compose -f skills-compose.yml down -v 2>/dev/null
	-cd Search && docker-compose -f search-compose.yml down -v 2>/dev/null
	-cd MicroTasks && docker-compose -f microtasks-compose.yml down -v 2>/dev/null
	@echo "✓ All services stopped and volumes wiped"

# wipe — alias для down (явное название для тех, кто привык).
wipe: down

logs:
	@if [ -z "$(service)" ]; then \
		echo "Usage: make logs service=<container_name>"; \
		echo "  Например: make logs service=api-gateway-api-gateway-1"; \
	else \
		docker logs $(service) -f --tail 100; \
	fi

status:
	@docker ps --filter "name=studjobs\|auth-\|users-\|achievements-\|vacancy-\|company-\|skills-\|search-\|microtasks-\|api-gateway-" \
		--format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

restart: down all
	@echo "✓ All services restarted (свежие volumes, свежая инициализация)"

# ── Soft-управление (без удаления данных) ───────────────────────────────────
# stop останавливает контейнеры, но НЕ удаляет их и volumes. Можно потом start.
# start поднимает обратно — Postgres, ES, Redis с накопленными данными.
# soft-restart = stop + start — мгновенный in-place рестарт всех сервисов
#                                без потери юзеров/ачивок/вакансий.
# Когда использовать что:
#   make soft-restart — контейнер залип, хочешь освежить процесс без потери БД
#   make restart      — нужна чистая инициализация (после смены DB_PASS, миграций)
#   make all          — первый старт или после wipe

stop:
	@echo "Stopping all services (data preserved)..."
	-cd API-Gateway && docker-compose -f api-gateway-compose.yml stop 2>/dev/null
	-cd MicroTasks && docker-compose -f microtasks-compose.yml stop 2>/dev/null
	-cd Search && docker-compose -f search-compose.yml stop 2>/dev/null
	-cd Skills && docker-compose -f skills-compose.yml stop 2>/dev/null
	-cd Vacancy && docker-compose -f vacancy-compose.yml stop 2>/dev/null
	-cd Company && docker-compose -f company-compose.yml stop 2>/dev/null
	-cd Achievements && docker-compose -f achieve-compose.yml stop 2>/dev/null
	-cd Users && docker-compose -f user-compose.yml stop 2>/dev/null
	-cd Auth && docker-compose -f auth-compose.yml stop 2>/dev/null
	-cd devops && docker-compose -f redis-compose.yml stop 2>/dev/null
	-cd devops && docker-compose -f elasticsearch-compose.yml stop 2>/dev/null
	-cd devops && docker-compose -f minio-compose.yml stop 2>/dev/null
	-cd devops && docker-compose -f haproxy-compose.yml stop 2>/dev/null
	-cd devops && docker-compose -f observability-compose.yml stop 2>/dev/null
	@echo "✓ All services stopped (volumes preserved)"

start:
	@echo "Starting all services (data preserved)..."
	-cd devops && docker-compose -f minio-compose.yml start 2>/dev/null
	-cd devops && docker-compose -f elasticsearch-compose.yml start 2>/dev/null
	-cd devops && docker-compose -f redis-compose.yml start 2>/dev/null
	-cd Auth && docker-compose -f auth-compose.yml start 2>/dev/null
	-cd Users && docker-compose -f user-compose.yml start 2>/dev/null
	-cd Achievements && docker-compose -f achieve-compose.yml start 2>/dev/null
	-cd Company && docker-compose -f company-compose.yml start 2>/dev/null
	-cd Vacancy && docker-compose -f vacancy-compose.yml start 2>/dev/null
	-cd Skills && docker-compose -f skills-compose.yml start 2>/dev/null
	-cd Search && docker-compose -f search-compose.yml start 2>/dev/null
	-cd MicroTasks && docker-compose -f microtasks-compose.yml start 2>/dev/null
	-cd API-Gateway && docker-compose -f api-gateway-compose.yml start 2>/dev/null
	-cd devops && docker-compose -f observability-compose.yml start 2>/dev/null
	-cd devops && docker-compose -f haproxy-compose.yml start 2>/dev/null
	@echo "Waiting for gateway to respond..."
	@i=0; until curl -fs http://localhost:9091/health >/dev/null 2>&1; do \
		[ $$i -ge 30 ] && echo "⚠ Gateway не отвечает за 60с — проверь docker ps" && exit 1; \
		i=$$((i+1)); sleep 2; \
	done
	@echo "✓ All services started"

soft-restart: stop start
	@echo "✓ Soft-restart done (данные сохранены)"

clean: down
	@echo "Cleaning up..."
	docker system prune -f
	@echo "✓ Cleanup completed"

# Установка grpcurl (опционально, для reindex).
# На macOS лучше: brew install grpcurl
setup-grpcurl:
	@if [ ! -f "./grpcurl" ]; then \
		echo "Downloading grpcurl..."; \
		if [ "$$(uname -s)" = "Darwin" ]; then \
			ARCH=$$([ "$$(uname -m)" = "arm64" ] && echo "arm64" || echo "x86_64"); \
			wget -q https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_osx_$$ARCH.tar.gz -O grpcurl.tar.gz; \
		else \
			wget -q https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_linux_x86_64.tar.gz -O grpcurl.tar.gz; \
		fi; \
		tar -xzf grpcurl.tar.gz; \
		rm grpcurl.tar.gz; \
		chmod +x grpcurl; \
		echo "✓ grpcurl installed"; \
	else \
		echo "✓ grpcurl already installed"; \
	fi

deps: setup-grpcurl
