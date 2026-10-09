# Отправка писем

Чтобы отправить письмо получателю или набору получателей, вызовите эндпоинт `POST $ComposeAPI/notification/email/send`.

!!! note
    Обратитесь к справке по API, чтобы найти все доступные параметры.


!!! important
    Убедитесь, что вы правильно настроили ваше окружение с учётными данными SMTP.


## Пример запроса

```bash
```
curl "$ComposeAPI/notification/email" \
  -H "Authorization: Bearer $JWT" \
  --data "{
    \"to\": [\"$USER_EMAIL\"],
    \"subject\": \"Test CURL email\",
    \"content\": { \"html\": \"<div>Test Content</div>\" }
  }"

## Пример ответа

```json
```
{
  "response": true
}
