# FileExporter / PrintServer

PrintServer формирует документы по шаблонам на DSL печати. Сервис получает опубликованный шаблон, подставляет рабочие данные и возвращает HTML или DOCX. Шаблоны и данные предпросмотра хранятся вне PrintServer; сам сервис использует кэш скомпилированных шаблонов.

Спецификация HTTP API находится в [NewsPrint.yaml](NewsPrint.yaml). Инструкции для работы с кодом и тестами — в [AGENT.md](AGENT.md).

## Требования

- Docker с поддержкой Compose.
- Доступ к внешнему источнику конфигурации, заданному через `CFG_METHOD` и `CFG_PATH`. Основной `docker-compose.yml` его не запускает.
- Файл `.env` в корне проекта для обоих compose-файлов.

Для сборки без Docker нужен Go 1.19 или новее; тестовый Dockerfile использует Go 1.19.

## Быстрый запуск

1. Создайте `.env` из образца:

   ```powershell
   Copy-Item example.env .env
   ```

   В Bash используйте `cp example.env .env`.

2. Настройте в `.env` как минимум `SERVICE_ID`, `CFG_METHOD` и `CFG_PATH` для своего окружения. Значение `CFG_PATH` в образце указывает на конкретный адрес инфраструктуры и может быть недоступно у вас. При `CFG_METHOD=psql` сервис загружает из внешней конфигурации адреса подсистем, публичный ключ JWT и ключ внутренних запросов. Пока конфигурация не получена, PrintServer повторяет попытку и не открывает HTTP-порт.

3. Соберите и запустите сервис:

   ```sh
   docker compose up --build
   ```

   Compose публикует порт `8153` контейнера на `localhost:8153`. Если меняете `HTTP_PORT`, обновите также проброс порта в `docker-compose.yml`.

4. Остановите сервис:

   ```sh
   docker compose down
   ```

У сервиса нет веб-страницы на `/`. HTTP-маршруты начинаются с `http://localhost:8153/api/core/print/`.

## HTTP API

| Метод и путь | Назначение |
| --- | --- |
| `POST /api/core/print/export` | Сформировать HTML или DOCX из опубликованного шаблона. |
| `POST /api/core/print/preview` | Сформировать HTML-предпросмотр переданного шаблона. |
| `POST /api/core/print/svc/validate-template` | Проверить шаблон и данные перед публикацией. Внутренний маршрут. |

Все маршруты сейчас проверяют JWT из `Authorization: Bearer ...` (либо из cookie `Authorization`). Внутренний маршрут дополнительно требует заголовок `SubToken`. Запросы принимают `application/json`; обязательные поля, форматы ответов и коды ошибок перечислены в [NewsPrint.yaml](NewsPrint.yaml).

Для экспорта PrintServer запрашивает опубликованный шаблон у ConfigAppServer, затем передаёт рабочие данные в `/printcontent` соответствующей подсистемы. Для предпросмотра данные-заглушки приходят из ConfigAppServer. При успешной валидации скомпилированный шаблон может быть сохранён в кэше.

## Переменные окружения

| Переменная | Для чего нужна |
| --- | --- |
| `HTTP_PORT` | Порт HTTP-сервера внутри контейнера; по умолчанию `8153`. |
| `SERVICE_ID` | Идентификатор сервиса, обязательный при загрузке конфигурации. |
| `CFG_METHOD`, `CFG_PATH` | Способ и адрес загрузки внешней конфигурации (`psql` или `ini`). |
| `Print_Content_Host`, `Print_Content_Port` | Необязательный отдельный адрес сервиса `/printcontent`; задаются вместе. |
| `PRINT_CACHE_MAIN_MAX_ENTRIES` | Максимальное число записей основного кэша. |
| `PRINT_CACHE_MAIN_MAX_BYTES` | Максимальный объём основного кэша в байтах. |

Остальные параметры и значения по умолчанию смотрите в `PrintServer/config`. `.env` используется контейнерами при запуске

## Тесты через Docker

Тесты находятся в `tests/compiler_test`, `tests/cache_test`, `tests/transformer_test` и `tests/handlers_test`. Они используют тестовые реализации зависимостей и не требуют запущенного PrintServer или ConfigAppServer. `Dockerfile.test` выполняет тесты с `-race`; это предпочтительный способ проверки, особенно на Windows.

Тестовый образ копирует исходники при сборке, поэтому после изменений сначала пересоберите его:

```sh
docker compose -f docker-compose.test.yml build test-backend
```

Полный прогон использует команду из `Dockerfile.test` (`go test -v ./tests/... -count=1 -race -timeout=120s`):

```sh
docker compose -f docker-compose.test.yml run --rm --no-deps test-backend
```

Для проверки одного пакета перед полным прогоном переопределите команду:

```sh
docker compose -f docker-compose.test.yml run --rm --no-deps test-backend go test -v ./tests/compiler_test -count=1 -race -timeout=120s
```

В тестовом compose-файле задан `restart: always`, поэтому для одноразового запуска используется `run`. Локальный запуск `go test ./tests/...` возможен, но на Windows запуск с `-race` может завершиться ошибкой доступа к временным исполняемым файлам.

## Структура проекта

| Путь | Содержимое |
| --- | --- |
| `PrintServer/server` и `PrintServer/handlers` | Маршруты, авторизация, разбор запросов и HTTP-ответы. |
| `PrintServer/transformer` | Сценарии экспорта, предпросмотра и валидации; обращения к другим сервисам. |
| `PrintServer/compiler` и `PrintServer/diagnostic` | Разбор DSL и диагностика ошибок шаблонов. |
| `PrintServer/cache` | LRU-кэш скомпилированных шаблонов. |
| `PrintServer/html2docx` и `PrintServer/docxgo` | Преобразование HTML в DOCX; `docxgo` — отдельный вложенный Go-модуль. |
| `tests` | Тесты PrintServer по областям. |

Корневой `go.mod` описывает модуль `PrintServer` и подключает локальный модуль `PrintServer/docxgo` через `replace`.
