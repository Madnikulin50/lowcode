# Sink-маршруты

LowCoooode позволяет обнаруживать входящие HTTP-запросы с помощью скриптов автоматизации — sink-маршрутов.
Sink-маршруты позволяют реализовывать собственные API-эндпоинты для добавления поддержки таких вещей, как вебхуки.

!!! note
    Здесь мы рассматриваем только настройку sink-маршрута.
    Обратитесь к [Integrator Guide](modules/integrator-guide/pages/index.md) за подробностями об их использовании.


## Генерация сигнатуры

**Сигнатура sink** используется для авторизации входящих HTTP-запросов к маршруту `/sink`.

!!! important
    Сигнатуры sink следует рассматривать как пароли.


Сигнатура sink генерируется с помощью команды CLI [sink signature](modules/devops-guide/pages/references/cli-reference.md#sink-signature).

Например, `{CLI*CMD*PREFIX} sink signature` возвращает такой вывод (сигнатура будет другой):

```
```
/system/sink?__sign=187...3D
Sink request constraints:
 - signature should be part of query-string
 - body size is not limited

!!! important
    Обратитесь к справочнику команд CLI за подробностями о доступных опциях (`{CLI_CMD_PREFIX} sink signature -h`).
    Приведённая выше команда создаёт неограниченную сигнатуру (любой запрос считается допустимым).


Сигнатуру можно найти в конце параметра запроса `__sign`.

!!! important
    Эту сигнатуру следует хранить в безопасном месте, так как она требуется для аутентификации любого запроса к sink-маршруту.


Каждый запрос к sink-маршруту должен указывать сгенерированную выше сигнатуру sink.

!!! caution
    Некоторые сервисы (при реализации потока OAuth2) могут не разрешать использование параметров запроса.
    См. <<sign-in-path>>, чтобы обойти это ограничение.


## Использование sink-маршрута

Когда вам нужно использовать sink-маршрут (например, вы хотите определить вебхук), все входящие запросы должны указывать сигнатуру sink, сгенерированную выше.
Либо в параметре запроса, либо в пути (в зависимости от предоставленных параметров CLI).

Например:
```bash
```
curl '$BASE*API*URL/system/sink?*_sign=$SINK*SIGNATURE' \
  --data-binary '{
    "a": "b"
  }';

!!! important
    Обратите внимание, что при использовании пользовательского базового URL API (см.: [HTTP_API_BASE_URL](modules/devops-guide/pages/configuration/server.md#_http_api_base_url)) или если webapp встроен в сервер (см.: [HTTP_WEBAPP_ENABLED](modules/devops-guide/pages/configuration/server.md#_http_webapp_enabled)), URL sink будет иметь префикс (например, по умолчанию при включённом webapp: `$BASE_API_URL/api/system/sink?__sign=$SINK_SIGNATURE`).


<a id="sign-in-path"></a>
## Не можете использовать параметры запроса?

Если вы не можете использовать параметр запроса для аутентификации запроса, добавьте аргумент `--signature-in-path` к команде CLI [sink signature](modules/devops-guide/pages/references/cli-reference.md#sink-signature).

Например, `{CLI*CMD*PREFIX} sink signature --signature-in-path` возвращает такой вывод (сигнатура будет другой):

```
```
/system/sink/ext*mautic/lead/*_sign=7a9...==
Sink request constraints:
 - signature should be part of path
 - body size is not limited

!!! caution
    Сигнатура «в пути» *не может использоваться* в параметре запроса и наоборот.
