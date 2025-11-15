# Сервис назначения ревьюеров для Pull Request'ов

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

Микросервис для автоматического назначения ревьюеров на Pull Request'ы с управлением командами и пользователями. Разработан в рамках тестового задания для стажёра Backend (Avito, осенняя волна 2025).

## Содержание

- [Описание задачи](#описание-задачи)
- [Функциональные требования](#функциональные-требования)
- [Технические требования](#технические-требования)
- [Планируемая архитектура](#планируемая-архитектура)

---

## Описание задачи

Необходимо реализовать микросервис, который:
- Автоматически назначает ревьюеров на Pull Request из команды автора
- Позволяет переназначать ревьюверов
- Управляет командами и пользователями
- Запрещает изменения после merge PR

### Основные сущности:

**Пользователь (User)**
- Уникальный идентификатор (user_id)
- Имя (username)
- Принадлежность к команде
- Флаг активности (is_active)

**Команда (Team)**
- Уникальное имя (team_name)
- Список участников

**Pull Request (PR)**
- Уникальный идентификатор (pull_request_id)
- Название
- Автор
- Статус (OPEN | MERGED)
- Список назначенных ревьюверов (до 2)

---

## Функциональные требования

### Правила назначения ревьюверов:

1. **При создании PR:**
   - Автоматически назначаются до 2 активных ревьюверов из команды автора
   - Автор не может быть назначен ревьювером своего PR
   - Учитываются только пользователи с `is_active = true`
   - Если доступных кандидатов меньше 2, назначается доступное количество (0/1)

2. **Переназначение ревьювера:**
   - Заменяет конкретного ревьювера на случайного активного участника из **команды заменяемого** ревьювера
   - Доступно только для PR со статусом OPEN

3. **Merge PR:**
   - Меняет статус PR на MERGED
   - Операция должна быть идемпотентной (повторный вызов не приводит к ошибке)
   - После merge изменение списка ревьюверов запрещено

4. **Управление пользователями:**
   - Деактивация пользователя (is_active = false) исключает его из назначения на новые PR
   - Получение списка PR, где пользователь назначен ревьювером

---

## Технические требования

### Производительность:
- **RPS:** 5 (с запасом до 20)
- **SLI времени ответа:** p99 < 300 мс
- **SLI успешности:** 99.9%

### Объёмы данных:
- До 20 команд
- До 200 пользователей

### Развёртывание:
- Сервис должен подниматься командой `docker-compose up`
- Миграции БД должны применяться автоматически
- Сервис должен быть доступен на порту **8080**

### Стек технологий:
- **Язык:** Go 1.25+
- **БД:** PostgreSQL 18
- **Контейнеризация:** Docker + Docker Compose

---

## Планируемая архитектура

Проект будет разделён на три основных слоя:

```
┌─────────────────────────────────────────────┐
│        HTTP Handlers        │
│  - Парсинг запросов                         │
│  - Валидация входных данных                 │
│  - Формирование ответов                     │
│  - Обработка HTTP ошибок                    │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│          Service Layer  │
│  - Бизнес-правила назначения ревьюверов     │
│  - Валидация бизнес-ограничений             │
│  - Координация операций                     │
│  - Управление транзакциями                  │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│       Repository Layer       │
│  - CRUD операции с БД                       │
│  - SQL запросы                              │
│  - Маппинг данных                           │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│            PostgreSQL Database              │
└─────────────────────────────────────────────┘
```

### Схема базы данных:

```sql
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE pr_statuses (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

INSERT INTO pr_statuses (id, name) VALUES (1, 'OPEN'), (2, 'MERGED');

CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    team_id INT NOT NULL REFERENCES teams(id) 
);

CREATE TABLE pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE,
    author_id VARCHAR(255) NOT NULL REFERENCES users(id),
    
    status_id SMALLINT NOT NULL REFERENCES pr_statuses(id) DEFAULT 1
);

CREATE TABLE pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status_id ON pull_requests(status_id);
CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);
```

### API Endpoints:

**Teams:**
- `POST /team/add` - Создать команду
- `GET /team/get?team_name=...` - Получить команду

**Users:**
- `POST /users/setIsActive` - Установить флаг активности
- `GET /users/getReview?user_id=...` - Получить PR пользователя

**Pull Requests:**
- `POST /pullRequest/create` - Создать PR
- `POST /pullRequest/merge` - Merge PR
- `POST /pullRequest/reassign` - Переназначить ревьювера

---

Этот проект создан в образовательных целях в рамках тестового задания для стажёра Backend (Avito, осенняя волна 2025).
