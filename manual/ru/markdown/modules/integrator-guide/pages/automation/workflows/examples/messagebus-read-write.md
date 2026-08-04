# Очередь Messagebus
:attachment-path: ../../../_attachments/automation/workflows/
:page-noindex: true

Механизм очередей Messagebus дополняет рабочие процессы таким образом, что любые асинхронные вычисления могут быть вынесены в другой рабочий процесс или процесс, расширяемый рабочим процессом.
Больше информации по описанным темам можно найти в [Workflows](modules/integrator-guide/pages/automation/workflows/examples/automation/workflows/index.md), у шлюза интеграции — в [Profiler](modules/integrator-guide/pages/automation/workflows/examples/api-gw/profiler.md) и в подсистеме [Messagebus](modules/developer-guide/pages/lowcode-server/messagebus.md).

.В этом примере мы:
- создадим очередь
- создадим эндпоинт
- запишем в очередь через рабочий процесс
- прочитаем из очереди через рабочий процесс
** отправим полезную нагрузку на эндпоинт
- посмотрим полезную нагрузку эндпоинта в профилировщике

## Создание новой очереди

!!! important
    `Consumer` здесь должен быть `Eventbus`, поскольку именно так внутренний механизм предоставляет данные полезной нагрузки подсистемам LowCoooode, включая рабочие процессы.


.На скриншоте показано создание очереди в интерфейсе.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-create.png",
    "alias": "automation-workflows-examples-queue-create.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

## Запись в очередь

Это рабочий процесс, который мы будем использовать только для ручного запуска записи в очередь.

!!! note
    Вы можете использовать любые кнопки автоматизации, чтобы запустить этот рабочий процесс из интерфейса.
    Обычно это делается после создания записи, когда новый recordID отправляется в очередь.


Полезная нагрузка определяется как `Array`, который затем преобразуется в строку
```json
```
toJSON([
  {"key":"foo", "value": "bar"},
  {"key":"bar", "value": "baz"}
])

.Скриншот рабочего процесса добавления в очередь.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-wf-add.png",
    "alias": "automation-workflows-examples-queue-wf-add.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 164,
    "y": 111,
    "w": 251,
    "h": 287
  },
    "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}queue_add.json).

## Чтение из очереди

Здесь мы разбираем полезную нагрузку очереди и создаём пользовательский HTTP-запрос во внешнюю систему, т.е. Kafka (в нашем примере — эндпоинт, который мы создадим только для этой цели).

.Пример чтения данных из очереди.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-wf-read.png",
    "alias": "automation-workflows-examples-queue-wf-read.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 128,
    "y": 75,
    "w": 251,
    "h": 720
  },
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}queue_read.json).

## Создание эндпоинта имитации сервиса

Теперь нам нужно запустить рабочий процесс `Queue Write`, и запрос будет отправлен на наш локальный сервер на эндпоинт `/api/gateway/example_kafka`.

Результирующий запрос можно увидеть в профилировщике шлюза интеграции (если он включён, подробнее см. [Profiler](modules/integrator-guide/pages/automation/workflows/examples/api-gw/profiler.md)).

.Добавление нового эндпоинта в шлюз интеграции.
![role="data-zoomable"](automation/workflows/examples/queue-endpoint.png)

## Просмотр запроса в профилировщике

Этот просмотр приведён только для примера — это способ использования профилировщика для отладки отправляемых полезных нагрузок в производственные/внешние сервисы.

.Просмотр заголовков запроса и других метаданных.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-profiler-1.png",
    "alias": "automation-workflows-examples-queue-profiler-1.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

.Просмотр полезной нагрузки запроса.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/queue-profiler-2.png",
    "alias": "automation-workflows-examples-queue-profiler-2.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}
