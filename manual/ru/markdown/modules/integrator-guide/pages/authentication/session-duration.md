# Определение длительности сессии

LowCoooode позволяет вам определить, как долго токены доступа считаются действительными.

.Три переменные `.env`, управляющие сессией аутентификации:
[cols="2s,5a"]
|===
| [#auth-sesion-auth*oauth2*access*token*lifetime]#[auth-sesion-auth*oauth2*access*token*lifetime,AUTH*OAUTH2*ACCESS*TOKEN*LIFETIME](#auth-sesion-auth*oauth2*access*token*lifetime,AUTH*OAUTH2*ACCESS*TOKEN*LIFETIME)#
|
Переменная `AUTH*OAUTH2*ACCESS*TOKEN*LIFETIME` в `.env` позволяет вам определить, как долго действует токен доступа.

Токен доступа представляет собой учётные данные, которые позволяют пользователям получать доступ к защищённым ресурсам, таким как записи, пользователи и рабочие процессы.

.Пример ограничения срока действия токена доступа двумя минутами:
[source,.env]
AUTH*OAUTH2*ACCESS*TOKEN*LIFETIME=2m

| [#auth-sesion-auth*oauth2*refresh*token*lifetime]#[auth-sesion-auth*oauth2*refresh*token*lifetime,AUTH*OAUTH2*REFRESH*TOKEN*LIFETIME](#auth-sesion-auth*oauth2*refresh*token*lifetime,AUTH*OAUTH2*REFRESH*TOKEN*LIFETIME)#
|
Переменная `AUTH*OAUTH2*REFRESH*TOKEN*LIFETIME` в `.env` позволяет вам определить, как долго должен действовать токен обновления.

Токен обновления предоставляет механизм, который генерирует новый токен доступа, когда старый истекает, избавляя от необходимости повторно аутентифицировать пользователя.

.Пример ограничения срока действия токена обновления двумя минутами:
[source,.env]
AUTH*OAUTH2*REFRESH*TOKEN*LIFETIME=2m

| [#auth-sesion-auth*session*lifetime]#[auth-sesion-auth*session*lifetime,AUTH*SESSION*LIFETIME](#auth-sesion-auth*session*lifetime,AUTH*SESSION*LIFETIME)#
|
Переменная `AUTH*SESSION*LIFETIME` в `.env` позволяет вам определить, как долго должна действовать сессия аутентификации.

Сессия аутентификации создаётся, когда пользователь вводит свои учётные данные на **странице входа LowCoooode**.
Сессия аутентификации не зависит от токенов доступа.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/auth-login.png",
    "alias": "auth-login-reg",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "sm",
    "x": 672,
    "y": 546,
    "h": 40,
    "w": 577
  }]
}

.Пример ограничения сессии аутентификации двумя минутами:
[source,.env]
AUTH*SESSION*LIFETIME=2m

| [#auth-sesion-auth*session*perm*lifetime]#[auth-sesion-auth*session*perm*lifetime,AUTH*SESSION*PERM*LIFETIME](#auth-sesion-auth*session*perm*lifetime,AUTH*SESSION*PERM_LIFETIME)#
|
Переменная `AUTH*SESSION*PERM_LIFETIME` в `.env` позволяет вам определить, как долго должна действовать сессия аутентификации при использовании кнопки btn:[login and remember me].

Сессия аутентификации создаётся, когда пользователь вводит свои учётные данные на **странице входа LowCoooode**.
Сессия аутентификации не зависит от токенов доступа.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/auth-login.png",
    "alias": "auth-login-perm",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "sm",
    "x": 672,
    "y": 497,
    "h": 42,
    "w": 577
  }]
}

.Пример ограничения сессии аутентификации двумя минутами:
[source,.env]
AUTH*SESSION*PERM_LIFETIME=2m

|===

## Выход неактивных пользователей через две минуты

Если вы хотите выполнять выход неактивных пользователей, вам нужно задать все три упомянутые выше переменные `.env`.

[source,.env]
AUTH*OAUTH2*ACCESS*TOKEN*LIFETIME=2m
AUTH*SESSION*LIFETIME=2m
AUTH*OAUTH2*REFRESH*TOKEN*LIFETIME=2m

!!! important
    Момент, когда пользователь считается неактивным, определяется браузером пользователя.
    Обычно это происходит, когда он закрывает вкладку/окно или когда его компьютер переходит в режим ожидания.
