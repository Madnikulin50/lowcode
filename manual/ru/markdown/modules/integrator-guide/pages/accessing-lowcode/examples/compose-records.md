# Работа с записями Low Code

## Получение списка записей

Чтобы выполнить поиск по записям конкретного модуля, используйте эндпоинт `GET $BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/`.

Обратитесь к [документации](modules/integrator-guide/pages/accessing-lowcode/examples/accessing-lowcode/index.md#api-endpoints) за информацией о получении полного справочника.

```bash
```
curl -X GET "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  --compressed;

```json
```
{
  "response": {
    "filter": {
      "moduleID": "$MODULE_ID",
      "namespaceID": "$NAMESPACE_ID",
      "query": "",
      "deleted": 0,
      "sort": "id"
    },
    "set": [
      {}
    ]
  }
}

## Фильтрация записей

Чтобы выполнить поиск по записям конкретного модуля, используйте эндпоинт `GET $BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/` с параметром запроса `query`.

Обратитесь к [документации](modules/integrator-guide/pages/accessing-lowcode/examples/accessing-lowcode/index.md#api-endpoints) за информацией о получении полного справочника.

```bash
```
curl -X GET "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/?query=f2='value+2'" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  --compressed;

```json
```
{
  "response": {
    "filter": {
      "moduleID": "$MODULE_ID",
      "namespaceID": "$NAMESPACE_ID",
      "query": "f2='value 2'",
      "deleted": 0,
      "sort": "id"
    },
    "set": [
      {}
    ]
  }
}

## Фильтрация записей с отсутствующими значениями

Значение записи может находиться в двух состояниях — **существующем** и **несуществующем**.
Когда значение **не существует**, оно равно `NULL`.

Например; поле-флажок может быть `true`, `false` или `NULL`, поэтому распознаются три значения вместо двух.

.В следующих примерах используются три записи:
- `{ "name": "a", "good": true }`
- `{ "name": "b", "good": false }`
- `{ "name": "c", "good": NULL }`

### Поиск записей со значением `NULL`

```bash
```
curl "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/?query=good+IS+NULL" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  --compressed;

```json
```
{
    "response": {
        "filter": {
            "moduleID": "$MODULE_ID",
            "namespaceID": "$NAMESPACE_ID",
            "query": "good IS NULL",
            "deleted": 0,
            "sort": "id"
        },
        "set": [
            {
                "recordID": "$RECORD_ID",
                "moduleID": "$MODULE_ID",
                "values": [
                    {
                        "name": "name",
                        "value": "c"
                    }
                ],
                "namespaceID": "$NAMESPACE_ID"
            }
        ]
    }
}

### Поиск записей со значением `NULL` или `false`

```bash
```
curl "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/?query=good+IS+NULL+OR+good=false" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  --compressed | pp_json;

```json
```
{
    "response": {
        "filter": {
            "moduleID": "$MODULE_ID",
            "namespaceID": "$NAMESPACE_ID",
            "query": "good IS NULL OR good=false",
            "deleted": 0,
            "sort": "id"
        },
        "set": [
            {
                "recordID": "$RECORD_ID",
                "moduleID": "$MODULE_ID",
                "values": [
                    {
                        "name": "good"
                    },
                    {
                        "name": "name",
                        "value": "b"
                    }
                ],
                "namespaceID": "$NAMESPACE_ID"
            },
            {
                "recordID": "$RECORD_ID",
                "moduleID": "$MODULE_ID",
                "values": [
                    {
                        "name": "name",
                        "value": "c"
                    }
                ],
                "namespaceID": "$NAMESPACE_ID"
            }
        ]
    }
}

## Обновление записей

Чтобы обновить конкретную запись конкретного модуля, используйте эндпоинт `POST $BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE*ID/record/$RECORD*ID`.

Эндпоинт обновления записи устанавливает значения записи ровно в то, что было указано в запросе.
Если вы хотите обновить только конкретные поля, сначала получите исходную запись, измените нужные значения, затем обновите запись.

Поля, которые вы опускаете в запросе, удаляются из записи (санация значений, автоматизация и другие системные процессы всё ещё могут задавать значение).

### Пример опускания существующих значений

.Следующий запрос задаёт второе поле и сбрасывает первое (уже заполненное):
```bash
```
curl -X POST "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE*ID/record/$RECORD*ID" \
  -H 'Accept: application/json, text/plain, */*' \
  -H "Authorization: Bearer $JWT" \
  -H 'Content-Type: application/json' \
  --data-raw '{ "values": [
    { "name": "f2", "value": "value 2" }
  ]}
' \
  --compressed

.Следующий ответ создаётся приведённым выше запросом:
```json
```
{
  "response": {
    "recordID": "$RECORD_ID",
    "moduleID": "$MODULE_ID",
    "values": [
      { "name": "f2", "value": "value 2" }
    ],
    "namespaceID": "$NAMESPACE_ID"
  }
}

### Пример сохранения существующих значений

.Следующий запрос задаёт второе поле и сохраняет первое (уже заполненное):
```bash
```
curl -X POST "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE*ID/record/$RECORD*ID" \
  -H 'Accept: application/json, text/plain, */*' \
  -H "Authorization: Bearer $JWT" \
  -H 'Content-Type: application/json' \
  --data-raw '{ "values": [
    { "name": "f1", "value": "value 1" },
    { "name": "f2", "value": "value 2" }
  ]}
' \
  --compressed

.Следующий ответ создаётся приведённым выше запросом:
```json
```
{
  "response": {
    "recordID": "$RECORD_ID",
    "moduleID": "$MODULE_ID",
    "values": [
      { "name": "f1", "value": "value 1" },
      { "name": "f2", "value": "value 2" }
    ],
    "namespaceID": "$NAMESPACE_ID"
  }
}

## Создание записей

Чтобы выполнить поиск по записям конкретного модуля, используйте эндпоинт `POST $BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/`.

Обратитесь к [документации](modules/integrator-guide/pages/accessing-lowcode/examples/accessing-lowcode/index.md#api-endpoints) за информацией о получении полного справочника.

```bash
```
curl -X POST "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  -H 'Content-Type: application/json' \
  --data-raw '{ "values": [
    { "name": "f1", "value": "value 1" },
    { "name": "f2", "value": "value 2" }
  ]}' \
  --compressed;

```json
```
{
  "response": {
    "recordID": "RECORD_ID",
    "moduleID": "MODULE_ID",
    "values": [
      { "name": "f1", "value": "value 1" },
      { "name": "f2", "value": "value 2" }
    ],
    "namespaceID": "NAMESPACE_ID"
  }
}

## Чтение записей

Чтобы выполнить поиск по записям конкретного модуля, используйте эндпоинт `GET $BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE*ID/record/$RECORD*ID`.

Обратитесь к [документации](modules/integrator-guide/pages/accessing-lowcode/examples/accessing-lowcode/index.md#api-endpoints) за информацией о получении полного справочника.

```bash
```
curl -X GET "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE*ID/record/$RECORD*ID" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  --compressed;

```json
```
{
  "response": {
    "recordID": "RECORD_ID",
    "moduleID": "MODULE_ID",
    "values": [
      { "name": "f1", "value": "value 1" },
      { "name": "f2", "value": "value 2" }
    ],
    "namespaceID": "NAMESPACE_ID"
  }
}


## Удаление записей

Чтобы выполнить поиск по записям конкретного модуля, используйте эндпоинт `DELETE $BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE*ID/record/$RECORD*ID`.

Обратитесь к [документации](modules/integrator-guide/pages/accessing-lowcode/examples/accessing-lowcode/index.md#api-endpoints) за информацией о получении полного справочника.

```bash
```
curl -X DELETE "$BASE*URL/compose/namespace/$NAMESPACE*ID/module/$MODULE*ID/record/$RECORD*ID" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  --compressed;

```json
```
{ "success": { "message": "OK" } }
