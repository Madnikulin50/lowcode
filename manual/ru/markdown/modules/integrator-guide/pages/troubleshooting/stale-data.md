# Устаревшие данные

Ошибка `stale data` возникает, когда LowCoooode обнаруживает, что запрос пытался изменить ресурс, который уже был изменён другим запросом.

.Диаграмма иллюстрирует, как возникает ошибка устаревших данных.
[plantuml,envoy-arch,svg,role=sequence]
@startuml
concise "Alice" as alc
concise "Bob" as bb

@0
alc is "Fetching resource R"
bb is "Fetching resource R"

@500
alc is "Editing resource R"
bb is "Editing resource R"

@900
alc is "Updating resource R"

@1600
alc is {hidden}

@1800
bb is "Updating resource R"

@2200
bb is {hidden}


highlight 0 to 500 #Gold;line:DimGrey : \t\tBoth Alice and Bob fetch\nthe same version of resource
highlight 900 to 1600 #Gold;line:DimGrey : Alice caused the resource on the server to change
highlight 1800 to 2200 #Pink;line:DimGrey : Bob tried to update a resource\ndetermined as stale

@enduml

## Пример: массовое обновление записей

В примере используется пакетное создание записей, где мы пытаемся обновить запись `$RECORD_ID` дважды.

```bash
```
curl -X POST "$BASE*URL/api/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  -H 'content-type: application/json' \
  --data-raw "{
    \"records\": [{
      \"set\": [{
        \"recordID\": \"$RECORD_ID\",
        \"moduleID\": \"$MODULE_ID\",
        \"values\": [{ \"name\": \"name\", \"value\": \"Some value\" }],
        \"namespaceID\": \"$NAMESPACE_ID\",
        \"createdAt\": \"2022-02-17T11:44:20Z\",
        \"updatedAt\": \"2022-02-17T17:13:42Z\"
      },
      {
        \"recordID\": \"$RECORD_ID\",
        \"moduleID\": \"$MODULE_ID\",
        \"values\": [{ \"name\": \"name\", \"value\": \"Some OTHER value\" }],
        \"namespaceID\": \"$NAMESPACE_ID\",
        \"createdAt\": \"2022-02-17T11:44:20Z\",
        \"updatedAt\": \"2022-02-17T17:13:42Z\"
      }]
    }]
  }" \
  --compressed

Ошибка `stale data` возникает, потому что первая запись вызывает изменение ресурса на сервере.

Когда обрабатывается вторая запись, текущий ресурс на сервере больше не тот, на который ссылается вторая запись.

Чтобы решить проблему, либо разделите запрос на два, либо определите последнюю версию локально перед отправкой запроса.

## Пример: выполнение workflow

В примере используется шлюз `fork`, где мы сначала получаем запись, а затем пытаемся обновить один и тот же экземпляр дважды.
Поскольку один разрешается раньше другого, второй вызывает ошибку `stale data`.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "troubleshooting/stale-data-workflow.png",
    "alias": "troubleshooting-stale-data-workflow",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 105,
    "y": 284,
    "w": 1200,
    "h": 481
  },
  "annotations": [{
    "kind": "box-danger",
    "padding": "lg",
    "x": 1024,
    "y": 364,
    "w": 200,
    "h": 321
  }]
}

Используйте две ветви fork для обновления значений без сохранения записи, чтобы решить проблему.

Сохраните запись в самом конце, когда все значения будут установлены так, как нужно.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "troubleshooting/stale-data-workflow-fix.png",
    "alias": "troubleshooting-stale-data-workflow-fix",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 235,
    "y": 139,
    "w": 1240,
    "h": 640
  },
  "annotations": [{
    "kind": "box-success",
    "padding": "lg",
    "x": 635,
    "y": 419,
    "w": 760,
    "h": 280
  }]
}
