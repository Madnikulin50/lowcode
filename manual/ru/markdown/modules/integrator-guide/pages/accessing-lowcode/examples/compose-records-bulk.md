# Массовые операции с записями Low Code

LowCoooode позволяет создавать, обновлять или обновлять несколько записей одним запросом.

## Массовое создание записей

Массовая полезная нагрузка, в которой записи определяют только свои значения, создаёт новую запись для каждой записи.

```bash
```
curl -X POST "$BASE*URL/api/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  -H 'content-type: application/json' \
  --data-raw '{
    "records": [{
      "set": [
        {"values":[{"name":"f1","value":"a"},{"name":"f1","value":"b"}]},
        {"values":[{"name":"f1","value":"a"},{"name":"f1","value":"b"}]}
      ]
    }]
  }' \
  --compressed

- `$BASE_URL` — это URL вашего инстанса LowCoooode (например, `https://lowcode.mydomain.tld`),
- `$NAMESPACE_ID` — это идентификатор пространства имён, для которого вы хотите создавать записи (например, `257934065070534659`),
- `$MODULE_ID` — это идентификатор модуля, для которого вы хотите создавать записи (например, `257934099950366723`),
- `$JWT` — это токен доступа для авторизации запроса.

## Массовое обновление записей

Массовая полезная нагрузка, в которой записи определяют свой `recordID` и свои значения, обновит существующие записи и их значения.

!!! caution
    LowCoooode удаляет из существующей записи любое значение, опущенное в запросе для обновляемой записи.
    Если вы хотите сохранить его, вы должны указать эту информацию в запросе.


```bash
```
curl -X POST "$BASE*URL/api/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  -H 'content-type: application/json' \
  --data-raw '{
    "records": [{
      "set": [
        {"recordID": "257936153397719043", "values":[{"name":"f1","value":"a edited"},{"name":"f1","value":"b edited"}]},
        {"recordID": "257936153565556739", "values":[{"name":"f1","value":"a edited"},{"name":"f1","value":"b edited 2"}]}
      ]
    }]
}' \
  --compressed

- `$BASE_URL` — это URL вашего инстанса LowCoooode (например, `https://lowcode.mydomain.tld`),
- `$NAMESPACE_ID` — это идентификатор пространства имён, для которого вы хотите обновлять записи (например, `257934065070534659`),
- `$MODULE_ID` — это идентификатор модуля, для которого вы хотите обновлять записи (например, `257934099950366723`),
- `$RECORD*ID**` — это `recordID` записи, которую вы хотите обновить (например, `257936153397719043`),
- `$JWT` — это токен доступа для авторизации запроса.

## Массовое удаление записей

Массовая полезная нагрузка, в которой записи определяют свой `recordID` и временную метку `deletedAt`, удаляет указанные записи и их значения.

```bash
```
curl -X POST "$BASE*URL/api/compose/namespace/$NAMESPACE*ID/module/$MODULE_ID/record/" \
  -H 'accept: application/json, text/plain, */*' \
  -H "authorization: Bearer $JWT" \
  -H 'content-type: application/json' \
  --data-raw '{
    "records": [{
      "set": [
        { "recordID": "$RECORD*ID*0", "deletedAt": "2021-11-15T09:49:22Z" },
        { "recordID": "$RECORD*ID*1", "deletedAt": "2021-11-15T09:49:22Z" },
        { "recordID": "$RECORD*ID*2", "deletedAt": "2021-11-15T09:49:22Z" }
      ]
    }]
}' \
  --compressed

- `$BASE_URL` — это URL вашего инстанса LowCoooode (например, `https://lowcode.mydomain.tld`),
- `$NAMESPACE_ID` — это идентификатор пространства имён, для которого вы хотите удалять записи (например, `257934065070534659`),
- `$MODULE_ID` — это идентификатор модуля, для которого вы хотите удалять записи (например, `257934099950366723`),
- `$RECORD*ID**` — это `recordID` записи, которую вы хотите удалить (например, `257936153397719043`),
- `$JWT` — это токен доступа для авторизации запроса.
