// k6 нагрузочный сценарий для StudJobs.
// Прогоняем три горячих GET-маршрута через API-Gateway:
//   1. /skills/popular       — статичный кэшируемый запрос (cache-friendly после Phase 3)
//   2. /users?skill_slugs=go — поиск через Elasticsearch (тяжёлый, но кэшируемый)
//   3. /tasks                — список открытых микрозадач (микс PG + ES)
//
// Запуск:
//   make loadtest
// или:
//   BASE_URL=http://localhost:8000 k6 run devops/k6/loadtest.js
//
// Перед прогоном нужен валидный Bearer-токен. Регистрируйся в UI или через curl на
// /api/v1/auth/login и положи в env: TOKEN="ey...".  Пример:
//   TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
//     -H 'Content-Type: application/json' \
//     -d '{"email":"student@smoke.local","password":"secret123","role":"ROLE_STUDENT"}' \
//     | jq -r .token)
//   TOKEN=$TOKEN make loadtest
//
// Результаты сравниваем до/после Redis-кэша (Phase 3) — фиксируем p95, p99, RPS.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';
const TOKEN    = __ENV.TOKEN || '';

// Кастомные метрики, которые k6 запишет в summary.
const skillsLatency = new Trend('skills_popular_ms', true);
const usersLatency  = new Trend('users_search_ms', true);
const tasksLatency  = new Trend('tasks_list_ms', true);
const cacheHitRate  = new Rate('cache_hit_ratio'); // эвристика: <50ms = вероятно cache hit

export const options = {
  // Сценарии нагрузки: ramp-up до 200 VU, держим минуту, потом ramp-down.
  // Для baseline без Redis ожидаем p95 на /users?skill_slugs=go ~150-300ms.
  scenarios: {
    skills: {
      executor: 'ramping-vus',
      exec: 'hitSkillsPopular',
      startVUs: 0,
      stages: [
        { duration: '20s', target: 50 },
        { duration: '40s', target: 200 },
        { duration: '30s', target: 200 },
        { duration: '10s', target: 0 },
      ],
    },
    users: {
      executor: 'ramping-vus',
      exec: 'hitUsersSearch',
      startVUs: 0,
      stages: [
        { duration: '20s', target: 30 },
        { duration: '40s', target: 100 },
        { duration: '30s', target: 100 },
        { duration: '10s', target: 0 },
      ],
    },
    tasks: {
      executor: 'ramping-vus',
      exec: 'hitTasksList',
      startVUs: 0,
      stages: [
        { duration: '20s', target: 30 },
        { duration: '40s', target: 100 },
        { duration: '30s', target: 100 },
        { duration: '10s', target: 0 },
      ],
    },
  },
  thresholds: {
    'http_req_duration{scenario:skills}': ['p(95)<200', 'p(99)<500'],
    'http_req_duration{scenario:users}':  ['p(95)<400', 'p(99)<1000'],
    'http_req_duration{scenario:tasks}':  ['p(95)<400', 'p(99)<1000'],
    'http_req_failed': ['rate<0.05'],
  },
};

const headers = TOKEN
  ? { Authorization: `Bearer ${TOKEN}`, 'Content-Type': 'application/json' }
  : { 'Content-Type': 'application/json' };

export function hitSkillsPopular() {
  const r = http.get(`${BASE_URL}/api/v1/skills/popular?limit=10`, { headers, tags: { scenario: 'skills' } });
  skillsLatency.add(r.timings.duration);
  cacheHitRate.add(r.timings.duration < 30);
  check(r, { 'skills 200': (res) => res.status === 200 });
  sleep(0.1);
}

export function hitUsersSearch() {
  const r = http.get(`${BASE_URL}/api/v1/users?skill_slugs=go&limit=10`, { headers, tags: { scenario: 'users' } });
  usersLatency.add(r.timings.duration);
  cacheHitRate.add(r.timings.duration < 50);
  check(r, { 'users 200': (res) => res.status === 200 });
  sleep(0.2);
}

export function hitTasksList() {
  const r = http.get(`${BASE_URL}/api/v1/tasks?limit=20`, { headers, tags: { scenario: 'tasks' } });
  tasksLatency.add(r.timings.duration);
  cacheHitRate.add(r.timings.duration < 50);
  check(r, { 'tasks 200': (res) => res.status === 200 });
  sleep(0.2);
}

// Сводный repro-блок печатается в конце прогона.
export function handleSummary(data) {
  const ms = (m) => m && m.values ? `${m.values['p(95)']?.toFixed(1)}/${m.values['p(99)']?.toFixed(1)}` : '—';
  const m = data.metrics;
  console.log('\n=== StudJobs load test summary (p95/p99 ms) ===');
  console.log(`skills/popular:  ${ms(m.skills_popular_ms)}`);
  console.log(`users?skill=go:  ${ms(m.users_search_ms)}`);
  console.log(`tasks:           ${ms(m.tasks_list_ms)}`);
  if (m.cache_hit_ratio) {
    console.log(`fast-response ratio (latency<50ms): ${(m.cache_hit_ratio.values.rate * 100).toFixed(1)}%`);
  }
  console.log(`total req:       ${m.http_reqs?.values?.count} (${m.http_reqs?.values?.rate?.toFixed(1)} rps)`);
  console.log(`failures:        ${(m.http_req_failed?.values?.rate * 100).toFixed(2)}%`);
  return { 'stdout': '' }; // подавляем дефолтный длинный summary
}
