# REST API

LowCoooode использует стандартные HTTP-серверы и обработчики Go с [chi](https://github.com/go-chi/chi) для обработки маршрутизации запросов.

## Обработка запросов

.Схема потока запроса:
[plantuml,api-request-life-cycle,svg]
@startuml
skinparam ParticipantPadding 20
skinparam BoxPadding 200
skinparam SequenceArrowThickness 2
skinparam RoundCorner 10

Server     -> Handler: HTTP Request
activate Handler
Handler    -> Controller: Request params
activate Controller
Controller -> Handler: Response payload
deactivate Controller
Handler    -> Server: HTTP Response
deactivate Handler

@enduml

.Общая схема HTTP-запроса:
1. обработчики запросов преобразуют запрос в соответствующую структуру,
1. сервис инициализирует параметры действия (используемые для журнала действий),
1. загружаются запрошенные ресурсы и проверяются разрешения на доступ,
1. выполняется соответствующая операция,
1. обработчик запроса подготавливает ответ, чаще всего в формате JSON.

## Эндпоинты API

Эндпоинты определяются в файлах `*/rest.yaml` (`/(compose|system|federation|automation)/rest.yaml`).
Инструмент генерации кода использует эти файлы `rest.yaml` для генерации шаблонной логики обработки запросов и ответов.

Файлы `rest.yaml` используются для генерации шаблонного кода обработки форматирования запросов и ответов (`rest/request/**.go` и `rest/handlers/**.go`).
Вы можете запустить инструмент генерации кода командой CLI `make codegen`.

Обработчики читают и нормализуют данные запроса в структуры **request** и передают их контроллерам.
Контроллеры — это функции, которые обрабатывают запрос, маршрутизируют его во внутренний сервис и форматируют итоговый вывод.

!!! note
    Контроллеры изначально разрабатывались для решения операций CRUD.
    Дальнейшее развитие платформы LowCoooode переросло это, поэтому вы можете заметить некоторые необычные паттерны (например, возврат функций `http.HandlerFunc` из обработчиков)


## Документация API

Документация API предоставляется Swagger с использованием спецификации Open API.
Формат Open API генерируется с помощью инструмента [openapi3-converter](https://github.com/lowcode/openapi3-converter/tree/develop).

Инструмент конвертера генерирует серию файлов `\{resource\}.yaml` в каталоге `/swagger`.
Если `lowcode-server` находится в том же каталоге, что и инструмент конвертера, файлы перемещаются в каталоги `lowcode-server/docs` и `/swagger`.

Чтобы сгенерировать статическую HTML-документацию API, обслуживаемую `lowcode-server`, выполните команду `make docs` в репозитории `lowcode-server`.
Статические HTML-файлы генерируются в каталоге `lowcode-server/docs/`.

### Расширение инструмента

Весь процесс конвертации выполняется в `openapi3-converter/tools/convert.js`.

### Добавление новых определений

Добавление нового определения добавляет новую запись в массив `const namespaces = [...]`.
Пересоберите определения Open API командой `yarn convert:yaml`.

.Пример нового ресурса foo:
```js
```
const namespaces = [
  {
    path: `${path}/system/rest.yaml`,
    namespace: 'system',
    className: 'System',
  },
  {
    path: `${path}/compose/rest.yaml`,
    namespace: 'compose',
    className: 'Compose',
  },
  {
    path: `${path}/federation/rest.yaml`,
    namespace: 'federation',
    className: 'Federation',
  },
  {
    path: `${path}/foo/rest.yaml`,
    namespace: 'foo',
    className: 'Foo',
  },
]

### Добавление новых эндпоинтов

Инструмент использует файлы `rest.yaml`, определённые в репозитории `lowcode-server`.
Добавьте новый эндпоинт в соответствующий файл `rest.yaml` и выполните команду `yarn convert:yaml`, чтобы пересобрать определения.

## Планы на будущее

- Удаление собственных определений API и замена их на OpenAPI 3.
- Улучшенная поддержка обновлений и частичных обновлений (PUT/PATCH).
