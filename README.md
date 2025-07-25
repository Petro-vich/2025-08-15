# Архиватор

Этот сервис позволяет пользователю отправлять ссылки на открытые файлы (например, .pdf или .jpeg), скачивать их, упаковывать в zip-архив и скачивать архив одним файлом.

## Возможности
- Создание задачи на архивирование файлов по ссылкам
- Добавление до 3 ссылок в одну задачу
- Фильтрация по расширениям файлов (.pdf, .jpeg)
- Одновременно не более 3 активных задач
- Получение статуса задачи
- Скачивание готового архива

## Быстрый старт

1. **Сборка и запуск**

```sh
make run
```

Сервер стартует на порту 8080 (или указанном в configs/config.toml).

2. **Примеры работы с API**

### 1. Создать задачу
```sh
curl -X POST http://localhost:8080/tasks
```
Ответ: `{ "id": "<task-uuid>" }`

### 2. Добавить ссылку на файл
```sh
curl -X POST -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/file1.pdf"}' \
  http://localhost:8080/tasks/<task-uuid>/files
```

### 3. Проверить статус задачи
```sh
curl http://localhost:8080/tasks/<task-uuid>
```

### 4. Скачать архив
```sh
curl -O -J http://localhost:8080/tasks/<task-uuid>/archive
```

## Ограничения
- В задаче не более 3 файлов
- Разрешены только .pdf и .jpeg (можно изменить в configs/config.toml)
- Одновременно не более 3 активных задач
- Если файл недоступен, архив создаётся только из доступных файлов, статус задачи будет "error"

## Тесты

Для запуска тестов:
```sh
make test
```

## Структура проекта
- `cmd/` — точка входа (main.go)
- `internal/taskmanager/` — бизнес-логика задач
- `internal/archiver/` — упаковка файлов в архив
- `internal/downloader/` — скачивание файлов
- `configs/` — конфигурация

---
