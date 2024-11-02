# Event Schedule Management Service (cor-events-scheduler)

Микросервис для управления расписаниями мероприятий с автоматическим распределением блоков и выступлений.

## Содержание
- [Возможности](#возможности)
- [Технологии](#технологии)
- [Структура проекта](#структура-проекта)
- [Установка и запуск](#установка-и-запуск)
- [API Документация](#api-документация)
- [Конфигурация](#конфигурация)
- [Мониторинг](#мониторинг)
- [Примеры использования](#примеры-использования)

## Возможности

- CRUD операции для расписаний мероприятий
- Автоматическое распределение блоков и выступлений
- Поддержка вложенной структуры (блоки и элементы)
- Валидация времени и длительности
- Автоматический расчет времени начала блоков
- Метрики Prometheus
- Структурированное логирование
- REST API
- Поддержка Docker и Kubernetes

## Технологии

- Go 1.21+
- Gin Web Framework
- GORM (PostgreSQL)
- Prometheus
- Zap Logger
- Docker & Kubernetes
- Viper (конфигурация)

## Структура проекта

```
cor-events-scheduler/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── models/
│   │   └── repositories/
│   ├── handlers/
│   ├── services/
│   └── infrastructure/
│       ├── db/
│       └── metrics/
├── pkg/
│   └── utils/
├── api/
│   └── swagger/
├── deployments/
│   └── kubernetes/
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── .env
```

## Установка и запуск

### Предварительные требования

- Go 1.21 или выше
- PostgreSQL
- Docker (опционально)
- Kubernetes (опционально)

### Локальная установка

1. Клонирование репозитория:
```bash
git clone <repository-url>
cd cor-events-scheduler
```

2. Установка зависимостей:
```bash
go mod download
```

3. Создание конфигурационного файла:
```bash
cat > config/config.yaml << EOF
server:
  address: ""
  port: "8282"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  name: "scheduler"
EOF
```

4. Запуск:
```bash
go run cmd/app/main.go
```

### Docker

1. Сборка образа:
```bash
docker build -t cor-events-scheduler .
```

2. Запуск с Docker Compose:
```bash
docker-compose up -d
```

### Kubernetes

1. Создание секрета:
```bash
kubectl create secret generic cor-events-scheduler-secrets \
  --from-literal=db_user=postgres \
  --from-literal=db_password=your-password
```

2. Применение конфигурации:
```bash
kubectl apply -f deployments/kubernetes/
```

## API Документация

### Endpoints

#### Расписания

##### Создание расписания
```http
POST /api/v1/schedules
```

Пример запроса:
```json
{
    "name": "Фестиваль Аниме 2024",
    "description": "Ежегодный фестиваль аниме и косплея",
    "start_date": "2024-04-01T10:00:00Z",
    "end_date": "2024-04-01T20:00:00Z",
    "blocks": [
        {
            "name": "Открытие фестиваля",
            "type": "opening",
            "duration": 30,
            "description": "Торжественное открытие",
            "order": 1,
            "items": [
                {
                    "name": "Приветственное слово",
                    "type": "speech",
                    "duration": 10,
                    "performer": "Организатор",
                    "requirements": "Микрофон"
                }
            ]
        }
    ]
}
```

##### Получение расписания
```http
GET /api/v1/schedules/{id}
```

##### Обновление расписания
```http
PUT /api/v1/schedules/{id}
```

##### Удаление расписания
```http
DELETE /api/v1/schedules/{id}
```

##### Список расписаний
```http
GET /api/v1/schedules?page=1&page_size=10
```

##### Автоматическое распределение выступлений
```http
POST /api/v1/schedules/arrange
```

Пример запроса:
```json
{
    "schedule_id": 1,
    "items": [
        {
            "type": "cosplay_performance",
            "name": "Выступление",
            "description": "Косплей персонажа",
            "duration": 10,
            "performer": "Участник",
            "requirements": "Музыка"
        }
    ]
}
```

### Типы данных

#### Schedule (Расписание)
```go
type Schedule struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    StartDate   time.Time `json:"start_date"`
    EndDate     time.Time `json:"end_date"`
    Blocks      []Block   `json:"blocks"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### Block (Блок)
```go
type Block struct {
    ID          uint        `json:"id"`
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    StartTime   time.Time   `json:"start_time"`
    Duration    int         `json:"duration"`
    Description string      `json:"description"`
    Order       int         `json:"order"`
    Items       []BlockItem `json:"items"`
}
```

#### BlockItem (Элемент блока)
```go
type BlockItem struct {
    ID           uint   `json:"id"`
    Name         string `json:"name"`
    Type         string `json:"type"`
    Description  string `json:"description"`
    Duration     int    `json:"duration"`
    Order        int    `json:"order"`
    Performer    string `json:"performer"`
    Requirements string `json:"requirements"`
}
```

## Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| SERVER_ADDRESS | Адрес сервера | "" |
| SERVER_PORT | Порт сервера | "8282" |
| DB_HOST | Хост БД | "localhost" |
| DB_PORT | Порт БД | "5432" |
| DB_USER | Пользователь БД | "postgres" |
| DB_PASSWORD | Пароль БД | "postgres" |
| DB_NAME | Имя БД | "scheduler" |

### Конфигурационный файл (config.yaml)
```yaml
server:
  address: ""
  port: "8282"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  name: "scheduler"
```

## Мониторинг

### Метрики Prometheus

Доступны по адресу `/metrics`

Основные метрики:
- `schedule_operations_total` - количество операций с расписаниями
- `schedule_operation_duration_seconds` - длительность операций
- `active_schedules` - количество активных расписаний

### Логирование

Структурированные логи в формате JSON с использованием Zap Logger.

Уровни логирования:
- INFO - стандартные операции
- ERROR - ошибки
- DEBUG - отладочная информация

## Примеры использования

### Создание расписания с блоками

```bash
curl -X POST http://localhost:8282/api/v1/schedules \
-H "Content-Type: application/json" \
-d '{
    "name": "Фестиваль Аниме 2024",
    "description": "Ежегодный фестиваль аниме и косплея",
    "start_date": "2024-04-01T10:00:00Z",
    "end_date": "2024-04-01T20:00:00Z",
    "blocks": [
        {
            "name": "Открытие фестиваля",
            "type": "opening",
            "duration": 30,
            "description": "Торжественное открытие",
            "order": 1,
            "items": [
                {
                    "name": "Приветственное слово",
                    "type": "speech",
                    "duration": 10,
                    "performer": "Организатор",
                    "requirements": "Микрофон"
                }
            ]
        }
    ]
}'
```

### Добавление новых выступлений

```bash
curl -X POST http://localhost:8282/api/v1/schedules/arrange \
-H "Content-Type: application/json" \
-d '{
    "schedule_id": 1,
    "items": [
        {
            "type": "cosplay_performance",
            "name": "Выступление 3",
            "description": "Косплей персонажа из Attack on Titan",
            "duration": 10,
            "performer": "Участник 3",
            "requirements": "Музыка: track3.mp3"
        }
    ]
}'
```

### Получение расписания

```bash
curl -X GET http://localhost:8282/api/v1/schedules/1 | jq .
```

## Разработка

### Добавление нового типа блока

1. Добавьте новый тип в модели
2. Обновите валидацию в сервисном слое
3. Обновите логику arrange для поддержки нового типа

### Добавление новых метрик

1. Определите метрику в `internal/infrastructure/metrics/prometheus.go`
2. Добавьте сбор метрики в соответствующих обработчиках

## Лицензия

[MIT](LICENSE)