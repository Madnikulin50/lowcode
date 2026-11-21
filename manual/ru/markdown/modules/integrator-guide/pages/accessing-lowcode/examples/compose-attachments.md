# Работа с вложениями Low Code

!!! note
    Мы опускаем большую часть данных, возвращаемых этими эндпоинтами.
    
    Мы заменяем важные данные переменными, например `$RECORD_ID`.


## Получение подписанного URL для скачивания

.Получите подписанный URL, используя эндпоинт чтения вложения или список вложений:
- чтение вложения: `GET $BASE*URL/compose/namespace/$NAMESPACE*ID/attachment/record/$ATTACHMENT_ID`
- список вложений: `GET $BASE*URL/compose/namespace/$NAMESPACE*ID/attachment/record`

.Структура объекта ответа:
```json
```
{
  "attachmentID": "$ATTACHMENT_ID",
  "ownerID": "$USER_ID",
  "url": "$ATTACHMENT*ORIGINAL*URL",<1>
  "previewUrl": "$ATTACHMENT*PREVIEW*URL",<2>
  "name": "$FILENAME_ORIGINAL",
  "meta": {...},
  "namespaceID": "$NAMESPACE_ID"
}
<1> `url` содержит подписанный URL на вложение.
<2> `previewUrl` содержит подписанный URL на версию вложения для предпросмотра (когда доступна).

### Пример запроса

```bash
```
curl "$BASE*URL/compose/namespace/$NAMESPACE*ID/attachment/record/$ATTACHMENT_ID" \
 -H "Authorization: Bearer $JWT";

### Пример ответа

```json
```
{
  "response": {
    "attachmentID": "$ATTACHMENT_ID",
    "ownerID": "$USER_ID",
    "url": "$ATTACHMENT*ORIGINAL*URL",
    "previewUrl": "$ATTACHMENT*PREVIEW*URL",
    "name": "$FILENAME_ORIGINAL",
    "meta": {...},
    "namespaceID": "$NAMESPACE_ID"
  }
}

## Доступ к вложению

Используйте `$ATTACHMENT*ORIGINAL*URL` для доступа к любому загружаемому вложению.
Вы можете либо скачать вложение вручную, использовать HTTP-клиент (например, Axios), либо отобразить вложение с помощью `<img src="..."` или чего-то подобного.
