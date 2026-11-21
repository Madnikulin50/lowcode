# Разбор запроса шлюза интеграции (Integration Gateway)
:attachment-path: ../../../_attachments/automation/workflows/
:page-noindex: true

В этом примере мы разберём запрос, отправленный на [Api Gw](modules/integrator-guide/pages/automation/workflows/examples/api-gw/index.md)

.На скриншоте показан рабочий процесс, который можно использовать для обработки запроса.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parse-gateway-payload.png",
    "alias": "automation-workflows-examples-parse-gateway-payload.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 93,
    "y": 183,
    "w": 1006,
    "h": 144
  },
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}request_process.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) System; onManual**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Parse request payload**:
*** *type**: `Process arbitrary data in jsenv`
*** *arguments**:
**** *scope**:
****** **type**: `Any`
****** **value type**: expression
****** **value**: `payload`
**** *source**: refer below
*** *results**:
**** *resultAny**: `parsedPayload`
3. **(3) Debug state**
4. **(7) Done**
******

## Конфигурация шлюза интеграции

.Ниже приведены параметры конфигурации соответствующего шлюза интеграции:
[cols="1s,5a"]
|===
| Endpoint
| `/examples/payload`

| Prefilter
| Header: `Token == "SOME*SECRET*TOKEN"`

| Processing
| Workflow processer: `On request payload notify user`

| Postfilter
| `Default JSON Response`
|===

## Запрос cURL

.Ниже приведён пример запроса cURL, вызывающего шлюз интеграции и рабочий процесс:
```bash
```
curl -X POST "$LOWCODE_BASE/api/gateway/examples/payload" \
  -H "Content-Type: application/json" \
  -H 'Token: SOME*SECRET*TOKEN' \
  -d '{"name":"Peter","id":123}';

## Функция разбора JSEnv

.Ниже приведён извлечённый исходный код, который следует использовать в шаге функции JSEnv:
```js
```
var inputData;

try {
    inputData = JSON.parse(input)
} catch (e) {
    throw new Error('could not parse input payload: ' + e);
}

if (!inputData.name) {
    throw new Error('could not parse input payload, name missing');
}

return {
    id: inputData.id,
    name: inputData.name
}
