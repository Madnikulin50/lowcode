# Аутентификация

Плагин `/src/plugins/auth.ts` обрабатывает логику аутентификации OAuth2.

.Плагин отвечает за:
- идентификацию сервера аутентификации,
- получение токена аутентификации,
- обновление токена аутентификации.

## Регистрация плагина

.Плагин аутентификации регистрируется следующим образом:
```js
```
import { plugins } from '@lowcode/lowcode-vue'


Vue.use(plugins.Auth(), { app: 'compose' })


## Идентификация сервера аутентификации

Вы можете определить сервер аутентификации явно или неявно.

Сервер определяется явно, когда предоставлен параметр `window.LowCoooodeAuth`.
Обычно он предоставляется в файле `public/config.js`.

Сервер определяется неявно, когда параметр `window.LowCoooodeAuth` **не** предоставлен.
Когда параметр `window.LowCoooodeAuth` не предоставлен, используется параметр `window.LowCoooodeAPI`.
Обычно он предоставляется в файле `public/config.js`.

.Определение сервера аутентификации выглядит следующим образом:
- Эндпоинт `LowCoooodeAPI` явно задан и заканчивается на `/api`; суффикс `/api` заменяется на `/auth` и неявно используется как `LowCoooodeAuth`.
- Эндпоинт `LowCoooodeAPI` явно задан и не заканчивается на `/api`; к значению `LowCoooodeAPI` добавляется суффикс `/auth` и неявно используется как `LowCoooodeAuth`.

.Примеры конфигурации:
|===
| Описание | LowCoooodeAPI | LowCoooodeAuth

| Задание обоих эндпоинтов явно
| `window.LowCoooodeAPI = 'your-lowcode-instance.tld'`
| `window.LowCoooodeAuth = 'your-lowcode-instance.tld/auth'`

| Задание только LowCoooodeAPI
| `window.LowCoooodeAPI = 'your-lowcode-instance.tld/custom'`
| `window.LowCoooodeAuth = 'your-lowcode-instance.tld/custom/auth'`

| Задание LowCoooodeAPI с автоматической установкой
| `window.LowCoooodeAPI = 'your-lowcode-instance.tld/api'`
| `window.LowCoooodeAuth = 'your-lowcode-instance.tld/auth'`
|===
