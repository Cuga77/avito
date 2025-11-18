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
curl http://localhost:8080/health | jq .
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

Сервис будет доступен по адресу `http://localhost:8080`.

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

    critical_errors
    ✓ 'rate<0.05' rate=0.00%

    http_req_duration
    ✓ 'p(95)<500' p(95)=74.9ms
    ✓ 'p(99)<2000' p(99)=115.19ms


  █ TOTAL RESULTS

    checks_total.......: 151907  504.541339/s
    checks_succeeded...: 100.00% 151907 out of 151907
    checks_failed......: 0.00%   0 out of 151907

    ✓ health check status is 200
    ✓ create team status is 201
    ✓ get team status is 200
    ✓ create PR status is 201
    ✓ get user reviews status is 200
    ✓ merge PR status is 200
    ✓ merge PR again status is 200

    CUSTOM
    critical_errors................: 0.00%  0 out of 0

    HTTP
    http_req_duration..............: avg=25.63ms min=289.36µs med=13.01ms max=643.46ms p(90)=59.9ms p(95)=74.9ms
      { expected_response:true }...: avg=25.63ms min=289.36µs med=13.01ms max=643.46ms p(90)=59.9ms p(95)=74.9ms
    http_req_failed................: 0.00%  0 out of 151907
    http_reqs......................: 151907 504.541339/s

    EXECUTION
    iteration_duration.............: avg=1.38s   min=1.22s    med=1.38s   max=1.98s    p(90)=1.44s  p(95)=1.46s
    iterations.....................: 21701  72.077334/s
    vus............................: 1      min=1           max=100
    vus_max........................: 100    min=100         max=100

    NETWORK
    data_received..................: 61 MB  201 kB/s
    data_sent......................: 33 MB  108 kB/s




running (5m01.1s), 000/100 VUs, 21701 complete and 0 interrupted iterations
default ✓ [ 100% ] 100 VUs  5m0s
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
