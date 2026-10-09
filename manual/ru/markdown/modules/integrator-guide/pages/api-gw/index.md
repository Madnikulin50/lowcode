# Integration Gateway

Возможность Integration Gateway позволяет вам определять пользовательские эндпоинты, которые могут поддерживать кастомные методы аутентификации, проверку запросов и ограничение скорости.

!!! tip
    Integration Gateway служит альтернативой sink-эндпоинтам.


Эти эндпоинты можно использовать для определения пользовательского эндпоинта для входящих вебхуков, для обработки данных, предоставленных внешней интеграцией, или для проксирования запросов к другому сервису.

Большинство базовых операций могут быть выполнены с помощью встроенной функциональности; более продвинутые операции могут быть реализованы с помощью простых фрагментов кода JavaScript или [воркфлоу](modules/integrator-guide/pages/api-gw/automation/workflows/index.md).

## Определение нового эндпоинта

Чтобы определить новый эндпоинт, перейдите в ваш экземпляр LowCoooode (например http://latest.lowcode.org/) и нажмите на приложение "admin area".

.Скриншот показывает селектор приложений и приложение admin area.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "app-selector.png",
    "alias": "apigw-app-selector-admin",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 669,
    "y": 285,
    "w": 282,
    "h": 230
  },
  "annotations": []
}

В admin area перейдите на подстраницу menu:System[Integration Gateway].
Из admin area вы можете создавать и обновлять ваши эндпоинты, а также управлять контролем доступа к ним.

.Скриншот показывает экран списка маршрутов.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-list.png",
    "alias": "api-gw-route-list",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "w": 1600,
    "h": 450
  },
  "annotations": []
}

Когда вы нажимаете на кнопку btn:[new], появляется новый экран, где нужно указать базовые параметры вашего эндпоинта.

.Скриншот показывает пользовательский интерфейс создания нового эндпоинта.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-create.png",
    "alias": "api-gw-route-create",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "w": 1600,
    "h": 450
  },
  "annotations": []
}

- **Endpoint** определяет путь для эндпоинта,
- **method** определяет HTTP-метод для эндпоинта.

Эндпоинт будет доступен по адресу `$BASE*URL/api/gateway$YOUR*ENDPOINT`.
Например, эндпоинт `/test-endpoint` для экземпляра `https://latest.lowcode.org` доступен как `https://latest.lowcode.org/api/gateway/test-endpoint`.

!!! tip
    Вы можете указать разные HTTP-методы на одном маршруте, но их нужно указывать как разные эндпоинты, т.е. `*GET* /customer` и `*PUT* /customer` как два разных эндпоинта Integration Gateway.


Когда вы отправляете форму, дополнительный раздел "filter list" открывается под базовыми параметрами.
Эти фильтры позволяют проверять и обрабатывать запросы, а также определять ответ.

Последующие разделы подробно рассматривают конкретные фильтры; что они делают и как их следует использовать.

.Скриншот показывает пользовательский интерфейс для прикрепления фильтров к эндпоинту.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-filters.png",
    "alias": "api-gw-route-filters",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "w": 1600,
    "h": 1080
  },
  "focus": {
    "x": 586,
    "y": 598,
    "w": 1070,
    "h": 295
  },
  "annotations": []
}

## Проверка запросов

Проверка запросов выполняется с помощью **prefilters**.
Prefilters позволяют вам проверить запрос и определить, должен ли данный эндпоинт его обрабатывать.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-filters.png",
    "alias": "api-gw-route-filters-prefilter",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "w": 1600,
    "h": 1080
  },
  "focus": {
    "x": 586,
    "y": 598,
    "w": 1070,
    "h": 295
  },
  "annotations": [{
    "kind": "box-note",
    "padding": "sm",
    "x": 586,
    "y": 666,
    "w": 110,
    "h": 20
  }]
}

!!! note
    В настоящее время невозможно использовать встроенный механизм аутентификации для проверки подлинности запросов.
    Эта возможность будет добавлена в будущем.


.Список доступных prefilters:
[cols="1s,5a"]
|===
| [#filters-prefilter-queryparam]#[filters-prefilter-queryparam,Query parameters](#filters-prefilter-queryparam,Query parameters)#
|
Query parameters prefilter позволяет вам определить условие, которое проверяет, может ли запрос быть обработан этим эндпоинтом на основе параметров запроса.

Параметры запроса передаются в механизм вычисления выражений в том виде, в котором они были предоставлены.
Например, параметры запроса `?param1=value1&param2=value2` будут доступны как `param1` и `param2` в выражении.

Обратитесь к [справочнику по выражениям](modules/integrator-guide/pages/api-gw/expr/index.md) для получения подробной информации о написании таких выражений.

Следующий пример проверяет, содержит ли данный HTTP-запрос параметр запроса `token` со значением `"super-secret-value"`.

.Определение эндпоинта:
- endpoint: `/test-query`
- method: `GET`
- выражение параметров запроса: `token == "super-secret-value"`

.Пример HTTP-запроса, соответствующего фильтру:
```bash
```
curl -X GET $BASE_URL/api/test-query?token=super-secret-value

.Пример HTTP-запроса, не соответствующего фильтру:
```bash
```
curl -X GET $BASE_URL/api/test-query?token=some-other-value-i-guessed

!!! caution
    Многословные и многозначные параметры запроса в настоящее время не поддерживаются.


| [#filters-prefilter-header]#[filters-prefilter-header,Header](#filters-prefilter-header,Header)#
|
Header prefilter позволяет вам определить условие, которое проверяет, может ли запрос быть обработан этим эндпоинтом на основе заголовков запроса.

Все системные заголовки передаются в механизм вычисления выражений в исходном виде.
Пользовательские заголовки автоматически преобразуются в формат `snake-case` + `PascalCase`.
Например, `test-header` становится `Test-Header`, а `test` становится `Test`.

Следующий заголовок `test: value` будет доступен как `Test` в выражении.

Обратитесь к [справочнику по выражениям](modules/integrator-guide/pages/api-gw/expr/index.md) для получения подробной информации о написании таких выражений.

Следующий пример проверяет, содержит ли данный HTTP-запрос заголовок `Token` со значением `"super-secret-value"`.

.Определение эндпоинта:
- endpoint: `/test-query`
- method: `GET`
- выражение заголовка: `Token == "super-secret-value"`

.Пример HTTP-запроса, соответствующего фильтру:
```bash
```
curl -X GET $BASE_URL/api/test-query \
  -H 'Token: super-secret-value'

.Пример HTTP-запроса, не соответствующего фильтру:
```bash
```
curl -X GET $BASE_URL/api/test-query \
  -H 'Token: some-other-value-i-guessed'

!!! caution
    Многословные и многозначные заголовки в настоящее время не поддерживаются.


| [#filters-prefilter-profiler]#[filters-prefilter-profiler,Profiler](#filters-prefilter-profiler,Profiler)#
|
Profiler prefilter позволяет конкретному эндпоинту собирать полезную статистику о входящем маршруте, такую как заголовки, тело запроса, длина содержимого и URI запроса.

Prefilter может быть добавлен после любого `Query parameters` или `Header` prefilter, чтобы запрос проходил через все проверки аутентификации, которые могут существовать на маршруте.

Подробнее о добавлении profiler prefilter читайте на [странице profiler](modules/integrator-guide/pages/api-gw/profiler.md#profiler-add).

|===

## Обработка запросов

Обработка запросов выполняется с помощью **processers**.
Processer определяет основную бизнес-логику, которую выполняет эндпоинт.
LowCoooode позволяет обрабатывать произвольные полезные нагрузки, такие как структурированный JSON или бинарное вложение.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-filters.png",
    "alias": "api-gw-route-filters-processer",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "w": 1600,
    "h": 1080
  },
  "focus": {
    "x": 586,
    "y": 598,
    "w": 1070,
    "h": 295
  },
  "annotations": [{
    "kind": "box-note",
    "padding": "sm",
    "x": 699,
    "y": 666,
    "w": 110,
    "h": 20
  }]
}

.Список доступных processers:
[cols="1s,5a"]
|===
| [#filters-proc-wfexec]#[filters-proc-wfexec,Workflow processer](#filters-proc-wfexec,Workflow processer)#
|
Workflow processer позволяет привязать [воркфлоу](modules/integrator-guide/pages/api-gw/automation/workflows/index.md) к эндпоинту.
Обратитесь к странице [workflow processing](modules/integrator-guide/pages/api-gw/api-gw/workflow-processing.md) для получения подробной информации о взаимодействии.

| [#filters-proc-payloadproc]#[filters-proc-payloadproc,Payload processer](#filters-proc-payloadproc,Payload processer)#
|
Payload processer позволяет обрабатывать полезную нагрузку запроса с помощью JavaScript.

Обратитесь к разделу [Javascript Processing](modules/integrator-guide/pages/api-gw/api-gw/javascript-processing.md) для получения более подробной информации.

|===

## Подготовка ответа

Подготовка ответа выполняется с помощью **postfilters**.
Postfilters позволяют завершить жизненный цикл запроса в зависимости от результата обработки.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-filters.png",
    "alias": "api-gw-route-filters-postfilter",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "w": 1600,
    "h": 1080
  },
  "focus": {
    "x": 586,
    "y": 598,
    "w": 1070,
    "h": 295
  },
  "annotations": [{
    "kind": "box-note",
    "padding": "sm",
    "x": 817,
    "y": 666,
    "w": 110,
    "h": 20
  }]
}

.Список доступных postfilters:
[cols="1s,5a"]
|===
| [#filters-postfilter-redirect]#[filters-postfilter-redirect,Redirection](#filters-postfilter-redirect,Redirection)#
|
Redirect postfilter обогащает полезную нагрузку ответа необходимыми HTTP-заголовками перенаправления.

| [#filters-postfilter-json]#[filters-postfilter-json,Default JSON response](#filters-postfilter-json,Default JSON response)#
|
JSON response postfilter обогащает заголовки ответа типом содержимого `application/json` и JSON-кодирует результаты обработки.

|===

## Отладка

### Системные логи

Включите отладочное логирование в вашем файле `.env`.
Обратитесь к [руководству DevOps](modules/devops-guide/pages/references/configuration/server.md) для получения подробной информации.

### Просмотр логов Docker

По умолчанию логи LowCoooode доступны через логи Docker.
Чтобы получить доступ к этим логам, сначала перейдите в директорию, где находится ваш экземпляр LowCoooode.

.Пример перехода в директорию экземпляра LowCoooode:
```bash
```
cd /opt/deploy/{LOWCODE_INSTANCE}/

Вы можете использовать команду `docker-compose logs server` для доступа к логам, выводимым `server`.

!!! tip
    Обратитесь к документации `docker-compose logs` для получения информации о доступных флагах.
    Использование `docker-compose logs -f --tail=20 server` будет следить за логами (новые логи будут добавляться внизу) и ограничит вывод последними 20 записями.


.Усеченный пример выполнения команды логирования:
```bash
```
docker-compose logs -f --tail=20 server

server_1  | 12:53:14.862        DEBUG   rbac    rbac/service.go:102     allow delete for lowcode::compose:record/245030892240891907/245030893465497603/246932114543603715       {"bypass": [], "context": [], "common": [245030892072923139], "authenticated": [245030891334791171], "anonymous": [], "identity": 250804535289769987, "indexed": 63, "rules": 420}

Вы также можете использовать `grep` или фильтрацию логов для отображения только определенных логов.
Обратитесь к [руководству DevOps](modules/devops-guide/pages/troubleshooting/logging.md) для получения подробной информации о логировании.

.Пример использования `grep` для отображения только отладочных логов:
```bash
```
user@server:/opt/deploy/{LOWCODE_INSTANCE}/$ docker-compose logs -f --tail=20 server | grep DEBUG
server_1  | 13:52:29.636        DEBUG   rbac    rbac/service.go:102     allow triggers.manage for lowcode::automation:workflow/248229091554160643
Last updated 2021-09-27 18:01:45 +0200
Some content has been disabled in this document
