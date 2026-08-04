# Персонализация экранов аутентификации

Все интерфейсы аутентификации, используемые LowCoooode, могут быть полностью изменены и стилизованы под ваши нужды.
Персонализация выполняется путём определения набора шаблонов и ресурсов (например, изображений и JavaScript-кода).
Это позволяет привести экран аутентификации в соответствие с фирменным стилем компании.

## Настройка

### Включение персонализации

Чтобы включить персонализацию интерфейсов аутентификации, необходимо задать следующие переменные в вашем файле `.env`:

[cols="2s,5a"]
|===
| Переменная | Описание

| [#env-devmode]#[env-devmode,`AUTH*DEVELOPMENT*MODE`](#env-devmode,`AUTH*DEVELOPMENT*MODE`)#
| Переменная `AUTH*DEVELOPMENT*MODE` включает инструменты разработчика, помогающие с персонализацией интерфейса.
Инструменты разработчика включают автоматическую перезагрузку шаблонов и пользовательский интерфейс разработчика.

| [#env-path]#[env-path,`AUTH*ASSETS*PATH`](#env-path,`AUTH*ASSETS*PATH`)#
| Переменная `AUTH*ASSETS*PATH` определяет, откуда система должна читать изменённые интерфейсы аутентификации.
Укажите переменной `AUTH*ASSETS*PATH` каталог, содержащий изменённые ресурсы.
Например; `AUTH*ASSETS*PATH=/opt/deploy/lowcode/auth-assets`.

| [#env-assets]#[env-assets,`HTTP*SERVER*ASSETS*PATH`](#env-assets,`HTTP*SERVER*ASSETS*PATH`)#
| Переменная `HTTP*SERVER*ASSETS_PATH` определяет, откуда система должна читать изменённые файлы иконок и логотипов.
Укажите переменной `HTTP*SERVER*ASSETS_PATH` каталог, содержащий файлы `icon.svg` и `logo.svg`.
Например; `HTTP*SERVER*ASSETS_PATH=/opt/deploy/lowcode/auth-assets`.
|===

!!! note
    Путь для переменных `AUTH_ASSETS_PATH` и `HTTP_SERVER_ASSETS_PATH` может быть абсолютным или относительным к бинарю `lowcode-server`.


### Экспорт текущих ресурсов

Чтобы экспортировать текущие ресурсы аутентификации, выполните команду `auth assets export`.

!!! note
    В качестве альтернативы вы можете скачать ресурсы https://github.com/lowcode/lowcode/tree/{{PAGE-VERSION}.x}/server/auth/assets[здесь].


```
```
Exports embedded assets into the provided path (must exists)

Usage:
  lowcode-server auth assets export [flags]

Flags:
  -h, --help   help for export

Например; предположим, что мы будем использовать каталог `/opt/deploy/lowcode/auth-assets`, выполните:

```
```
lowcode-server auth assets export /opt/deploy/lowcode/auth-assets

!!! note
    Если вы уже определили переменную <<env-path,`AUTH_ASSETS_PATH`>> в `.env`, вы можете опустить путь при выполнении команды.


Результат выполнения выглядит так:

```
```
exporting auth assets to public
directory created
exporting asset /opt/deploy/lowcode/auth-assets/public/background.jpeg: ok
exporting asset /opt/deploy/lowcode/auth-assets/public/logo.png: ok
exporting asset /opt/deploy/lowcode/auth-assets/public/script.js: ok
exporting asset /opt/deploy/lowcode/auth-assets/public/style.css: ok
exporting auth assets to templates
directory created
exporting asset /opt/deploy/lowcode/auth-assets/templates/authorized-clients.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/change-password.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/error-internal.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/inc_footer.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/inc_header.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/inc_nav.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/inc_toasts.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/login.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/logout.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/mfa-totp-disable.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/mfa-totp.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/mfa.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/oauth2-authorize-client.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/password-reset-requested.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/pending-email-confirmation.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/profile.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/request-password-reset.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/reset-password.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/security.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/sessions.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/signup.html.tpl: ok
exporting asset /opt/deploy/lowcode/auth-assets/templates/template-dev.html.tpl: ok

!!! caution
    Повторный запуск команды не перезаписывает существующие файлы.
    Если вы хотите начать с чистого листа, удалите экспортированные ресурсы и экспортируйте их заново.


## Изменение ресурсов аутентификации

.Экспортированные ресурсы определяют следующую структуру файлов:
```
```
/ public <1>
  / ...
/ templates <2>
  / ---
<1> Подкаталог `/public` содержит ресурсы, которые вы можете использовать на экранах аутентификации, такие как изображения и таблицы стилей.
Вы можете свободно удалять, добавлять или изменять все файлы в подкаталоге `/public`.
Публичные файлы обслуживаются в том же формате, в котором они определены здесь, в подкаталоге `/public`.
<2> Подкаталог `/templates` содержит HTML-шаблоны, используемые при отображении экранов аутентификации.
Шаблоны написаны на [синтаксисе шаблонов go](https://golang.org/pkg/html/template/).
Синтаксис такой же, как и в наших [шаблонах документов](modules/integrator-guide/pages/authentication/templates/index.md).

!!! warning
    Файлы шаблонов, экспортированные упомянутой выше командой CLI, *не следует* переименовывать.


Если вы хотите определить переиспользуемые компоненты, вы можете добавить дополнительные файлы шаблонов (например, `inc*header.html.tpl` и `inc*footer.html.tpl`).

### Сценарии изменений

Сценарии можно изменять в файле `templates/scenarios.yaml`, пока LowCoooode работает в режиме разработки.

### Настройка иконки и логотипа

Файлы иконки и логотипа встраиваются в экраны аутентификации.
Чтобы изменить иконку и логотип, вы должны определить переменную [env-assets,`HTTP*SERVER*ASSETS*PATH`](#env-assets,`HTTP*SERVER*ASSETS*PATH`) в `.env`.
Путь должен указывать на каталог, содержащий файлы `icon.svg` и `logo.svg`.

Ресурсы по умолчанию находятся в каталоге `server/assets/src`, или вы можете скачать их [здесь](https://github.com/lowcode/lowcode/tree/{{PAGE-VERSION}.x}/server/assets/src).

!!! important
    Убедитесь, что вы также скопировали файлы `api-404.html`, `api-landing.html`, `logo.png` и `release-background.jpg` в каталог `HTTP_SERVER_ASSETS_PATH`.
    В противном случае эти файлы будут недоступны серверу.


### Стилизация

Таблица стилей по умолчанию подключается с CDN Bootstrap в шаблоне `inc_header.html.tpl`.
Вы можете свободно изменить источник таблицы стилей.

!!! note
    Система не реализует автоматическую предобработку стилей или транспиляцию JavaScript.
    Вы можете предобрабатывать ресурсы собственными инструментами перед их использованием в аутентификации.


### Инструмент разработчика

Когда вы включите переменную [env-devmode,`AUTH*DEVELOPMENT*MODE`](#env-devmode,`AUTH*DEVELOPMENT*MODE`) в `.env`, вы получите доступ к инструменту разработчика по URL `$BASE*API*URL/auth/dev`.
Инструмент разработчика отображает все шаблоны со всеми сценариями.

.Скриншот инструмента разработчика для кастомизации аутентификации.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "personalization/devtool.png",
    "alias": "personalization-devtool.png",
    "w": 1890,
    "h": 985
  },
  "view": {},
  "annotations": []
}

## Развёртывание изменённых ресурсов

Как только вы включите переменную [env-devmode,`AUTH*DEVELOPMENT*MODE`](#env-devmode,`AUTH*DEVELOPMENT*MODE`) в `.env`, изменённые ресурсы будут разворачиваться автоматически.

Отключите переменную после завершения персонализации.

!!! important
    Не изменяйте переменную <<env-path,`AUTH_ASSETS_PATH`>> в `.env`, так как шаблоны читаются напрямую с файловой системы.


При использовании docker убедитесь, что вы либо смонтировали исходные ресурсы как том, либо пересобрали пользовательский образ поверх стандартного.
