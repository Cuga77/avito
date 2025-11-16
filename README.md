# Сервис назначения ревьюеров #

[](https://go.dev/)
[](https://www.postgresql.org/)
[](https://www.docker.com/)

Этот сервис представляет собой Go-микросервис, реализующий API для управления Pull Request'ами, командами и пользователями, а также для автоматического назначения ревьюеров. Разработан в рамках тестового задания для стажёра Backend (Avito, осенняя волна 2025).
---

## 1\. Быстрый запуск

Сервис, база данных (PostgreSQL) и мигратор полностью управляются через `docker-compose`.

**1. Собрать и запустить все сервисы (api, postgres, migrator) в фоновом режиме:**

```bash
make run
```

**2. Проверить, что API доступен:**

```bash
curl -s http://localhost:8081/health | jq .
```

```json
{
  "status": "ok",
  "service": "pr-reviewer"
}
```

**3. Остановить и очистить окружение:**

```bash
make down
```

Сервис будет доступен по адресу `http://localhost:8081`.

-----

## 2\. Тестирование

Проект включает E2E (End-to-End) тесты для проверки всей бизнес-логики и `k6` для нагрузочного тестирования.

### E2E Тесты

E2E-тесты запускаются в изолированном окружении (`docker-compose.e2e.yml`) и проверяют полный цикл API (создание, мерж, асинхронную деактивацию).

```bash
make test-e2e
```

**Результат E2E тестов:**

```
Запуск E2E тестов...
=== RUN   TestFullLifecycle
=== RUN   TestFullLifecycle/Create_Team
=== RUN   TestFullLifecycle/Create_PR
=== RUN   TestFullLifecycle/Merge_PR
=== RUN   TestFullLifecycle/Reassign_Merged_PR_Fails
--- PASS: TestFullLifecycle (0.13s)
    --- PASS: TestFullLifecycle/Create_Team (0.04s)
    --- PASS: TestFullLifecycle/Create_PR (0.06s)
    --- PASS: TestFullLifecycle/Merge_PR (0.02s)
    --- PASS: TestFullLifecycle/Reassign_Merged_PR_Fails (0.00s)
=== RUN   TestBatchDeactivate
    api_test.go:189: Waiting 15 seconds for background worker to process task (TeamID: 11)...
    api_test.go:194: Checking status for user e2e_async_user_a_1763245780408040106...
    api_test.go:210: E2E test SUCCESS: User 'e2e_async_user_a_1763245780408040106' was correctly deactivated.
--- PASS: TestBatchDeactivate (15.07s)
PASS
ok      avito/tests/e2e 15.200s
```

### Нагрузочное тестирование (k6)

Скрипт `tests/k6/load_test.js` выполняет стресс-тест, имитирующий полный жизненный цикл API под высокой нагрузкой (до 100 одновременных пользователей).

```bash
# (Перед запуском убедитесь, что основной сервис запущен: make run)
make test-load-stress
```

**Результаты нагрузочного тестирования (Stress Test, 100 VUs):**

Сервис успешно прошел нагрузочное тестирование, многократно превысив требуемые SLI.

  * **SLI (Задание):** 5 RPS, p(95) \< 300ms
  * **Результат (Факт):** **\~141 RPS**, **p(95) = 82.74ms**

Все `checks` (проверки бизнес-логики) прошли со 100% успехом.

```
  █ THRESHOLDS

    http_req_duration
    ✓ 'p(95)<500' p(95)=82.74ms
    ✓ 'p(99)<1000' p(99)=145.14ms

  █ TOTAL RESULTS

    checks_total.......: 86652   283.218173/s
    checks_succeeded...: 100.00% 86652 out of 86652
    checks_failed......: 0.00%   0 out of 86652

    ✓ health check status is 200
    ✓ create team status is 201
    ✓ create team returns team_name
    ✓ get team status is 200
    ✓ get team returns members
    ✓ create PR status is 201
    ✓ create PR has reviewers
    ✓ create PR author not in reviewers
    ✓ get user reviews status is 200
    ✓ get user reviews contains PR
    ✓ merge PR status is 200
    ✓ merge PR sets status to MERGED
    ✓ merge PR again status is 200
    ✓ merge PR idempotent
    ✓ reassign after merge should fail

    CUSTOM
    critical_errors................: 0.00%  0 out of 0

    HTTP
    http_req_duration..............: avg=29ms    min=281.56µs med=7.03ms  max=718.76ms p(90)=64.82ms p(95)=82.74ms
    http_req_failed................: 11.11%
    http_reqs......................: 43326   141.609086/s
```

-----

## 3\. Архитектурные решения (Допущения)

Во время разработки были приняты следующие решения, основанные на требованиях и фидбэке:

### СУБД (PostgreSQL)

Выбрана PostgreSQL (вместо in-memory) для надежности и использования транзакций.

### Чистая архитектура

Проект разделен на 3 слоя:

  * `internal/handler`: Отвечает за парсинг JSON, валидацию DTO и форматирование HTTP-ответов.
  * `internal/service`: Содержит всю бизнес-логику, управляет транзакциями и координирует репозитории.
  * `internal/repository`: Отвечает только за выполнение SQL-запросов.

### Атомарность

Все операции, изменяющие несколько таблиц (создание команды + юзеров, создание PR + ревьюеров, переназначение), обернуты в транзакции (`db.WithTransaction`) на уровне *сервиса*, а не репозитория.

### Асинхронная деактивация

(Доп. задание) Метод `POST /users/batchDeactivate` отвечает \< 100 мс (возвращает `HTTP 202 Accepted`), создавая задачу в таблице `batch_deactivate_tasks`. Отдельный фоновый `TaskWorker` (запущенный в `main.go`) опрашивает эту таблицу, блокирует задачи (`FOR UPDATE SKIP LOCKED`) и безопасно выполняет деактивацию и переназначение PR.

### Конфигурация

Вся конфигурация (порт, БД) загружается из `ENV`.

### Линтер

(Доп. задание) Проект настроен на использование `golangci-lint` (конфигурация в `.golangci.yml`, не показан) для обеспечения единого стиля кода.
