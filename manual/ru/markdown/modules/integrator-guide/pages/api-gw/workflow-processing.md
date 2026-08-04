# Обработка с помощью воркфлоу
:attachment-path: ../_attachments/api-gw/workflow-processing

Когда вашему эндпоинту Integration Gateway требуется дополнительная обработка и встроенных processers недостаточно, вы можете привязать воркфлоу к эндпоинту для обработки запросов.

Чтобы привязать воркфлоу к маршруту Integration Gateway, сначала нужно правильно настроить воркфлоу, а затем привязать его к [workflow processer](modules/integrator-guide/pages/api-gw/api-gw/index.md#filters-proc-wfexec) маршрута.

## Настройка воркфлоу

Если вы хотите, чтобы воркфлоу мог выполняться Integration Gateway, воркфлоу должен определить правильный триггер.

.Чтобы настроить ваш воркфлоу:
1. Создайте новый воркфлоу или отредактируйте существующий.
Обратитесь к [документации по воркфлоу](modules/integrator-guide/pages/api-gw/automation/workflows/index.md) для получения подробной информации.
1. Убедитесь, что данный воркфлоу определяет триггер с ресурсом `System` и событием `onManual`.

### Чтение тела запроса через функцию

Способ оборачивания http.Request позволяет использовать его в среде обработки JS (подробнее см. [обработка JavaScript](modules/integrator-guide/pages/api-gw/api-gw/index.md)) или в функции, разработанной специально для получения содержимого тела запроса.

.Скриншот показывает базовое определение воркфлоу через функцию чтения тела запроса.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/wf-request-read-fn.png",
    "alias": "wf-request-read-fn",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Найдите [воркфлоу]({attachment-path}/wf-request-read-fn.json) для этого примера по гиперссылке.

### Чтение тела запроса через JS

.Скриншот показывает базовое определение воркфлоу через JS-функцию.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/wf-request-read-js.png",
    "alias": "wf-request-read-js",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Найдите [воркфлоу]({attachment-path}/wf-request-read-js.json) для этого примера по гиперссылке.

.Исходный код JavaScript может быть таким же простым, как:
```javascript
```
try {
    return JSON.parse(readRequestBody(input));
} catch (e) {
    return 'could not parse request body: ' + e.message
}

.Тело, которое было прочитано, также может быть использовано как `string` и позже разобрано на последующих шагах воркфлоу:
```javascript
```
try {
    return readRequestBody(input);
} catch (e) {
    return 'could not read request body: ' + e.message
}

## Конфигурация маршрута Integration Gateway

.Чтобы привязать воркфлоу к эндпоинту Integration Gateway:
1. Создайте новый эндпоинт Integration Gateway или отредактируйте существующий.
Обратитесь к [документации Integration Gateway](modules/integrator-guide/pages/api-gw/api-gw/index.md) для получения подробной информации.
1. Добавьте [workflow processer](modules/integrator-guide/pages/api-gw/api-gw/index.md#filters-proc-wfexec) к эндпоинту.
1. Выберите воркфлоу, который вы хотите привязать, из выпадающего списка в модальном окне.
1. Подтвердите выбор и нажмите на кнопку btn:[save], чтобы обновить эндпоинт.

.Скриншот показывает конфигурацию маршрута с добавленным workflow processer.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-wf.png",
    "alias": "route-wf",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

.Скриншот показывает конфигурацию workflow processer.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/route-wf-modal.png",
    "alias": "route-wf-modal",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

## Тестирование

Чтобы проверить, правильно ли взаимодействуют маршрут Integration Gateway и воркфлоу, инициируйте HTTP-запрос с помощью Postman или любого другого инструмента, такого как cURL.

Обратитесь к [примерам curl](modules/integrator-guide/pages/api-gw/api-gw/index.md#apigw-proc-js-example) для получения подробной информации.
