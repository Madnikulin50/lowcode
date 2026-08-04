# Внешние поставщики аутентификации

Использование внешних поставщиков аутентификации позволяет вашим пользователям использовать внешние сервисы (такие как Google и GitHub) для целей аутентификации.
Внешних поставщиков можно определить в панели LowCoooode Admin, в разделе menu:System[Settings,External authentication providers].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/settings.png",
    "alias": "authentication-external-settings",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "w": 1599,
    "h": 1080
  },
  "focus": {
    "x": 576,
    "y": 518,
    "w": 1088,
    "h": 552
  },
  "annotations": []
}

Чтобы включить внешнего поставщика аутентификации, вы должны зарегистрировать LowCoooode как клиента, используя пользовательский интерфейс поставщика.
После того как вы получите ключ пользователя и секрет, вы можете настроить LowCoooode через панель администрирования.
Внешние поставщики вступают в силу немедленно, без перезапуска сервера.

## Поставщики аутентификации

:leveloffset: +2


# Google
:page-noindex: true

.To enable Google authentication, you need to retrieve your application credentials:
1. Go to [Google Sign-in Guide](https://developers.google.com/identity/sign-in/web/sign-in#before*you*begin) and click on "Configure a project" button.
1. Select an **existing or create a new** project.
1. Set a product name.
1. On "Configure your OAuth client" screen select "Web browser" and paste the URL where your LowCoooode system is running (including `https://`).
1. Copy and paste both **Client ID** and **Client Secret** fields to LowCoooode Admin.


# Facebook
:page-noindex: true

.To enable Facebook authentication, you need to retrieve your application credentials:
1. Go to [Facebook for developers](https://developers.facebook.com/apps/) website, click on **"Add a new app"** or **select an existing app**.
1. On the list of available products search for "Facebook Login" and click on the "Set Up" button.
1. Select "Web" platform and paste the URL where your LowCoooode system is running.
1. Go to "Settings" and then "Basic" in the left sidebar.
1. Copy and paste both **App ID** and **App Secret** fields to LowCoooode Admin; **app ID** maps to **client key**, **app secret** maps to **secret**.


# GitHub
:page-noindex: true

.To enable GitHub authentication, you need to retrieve your application credentials:
1. Go to [GitHub](https://github.com/settings/applications/new) and create a new OAuth application.
1. Copy and paste both **Client ID** and **Client Secret** fields to LowCoooode Admin.


# LinkedIn
:page-noindex: true

.To enable LinkedIn authentication, you need to retrieve your application credentials:
1. Go to [LinkedIn](https://www.linkedin.com/developers/apps/new), fill out the form and click on "Create app".
1. Go to Auth section and copy and paste both the **Client ID** and **Client Secret** fields to LowCoooode Admin.


<a id="saml"></a>
# SAML
:page-noindex: true

Чтобы включить SAML-аутентификацию, необходимо обменяться учётными данными с вашим поставщиком удостоверений (далее IdP).
Процесс настройки включает необязательную генерацию ключей и настройку IdP в панели администрирования.

## Генерация пары закрытого и открытого ключей

!!! caution
    Примеры учётных данных, приведённые в этом документе, *не следует* использовать в ваших инстансах.


!!! note
    Этот шаг необязателен.
    Вы можете использовать существующую пару ключей для настройки SAML.


.Минимальные необходимые команды для генерации закрытого и открытого ключей, используемых для SAML-аутентификации:
[source,shell script]
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 \
  -keyout private.key \
  -out certificate.crt \
  -subj "/CN=lowcode.your-server.com" <1>

openssl rsa -in private.key -out private-rsa.key <2>

cat certificate.crt private-rsa.key <3>
<1> Команда генерирует новую пару ключей.
<2> Команда преобразует закрытый ключ в формат RSA.
<3> Команда выводит соответствующие ключи, которые вы можете скопировать.

!!! important
    Примеры учётных данных, приведённые в этом документе, *не следует* использовать в ваших инстансах.


.Следующий фрагмент показывает пример открытого ключа.
```txt
```
-----BEGIN CERTIFICATE-----
MIICwDCCAagCCQDbTNd4i3X/4zANBgkqhkiG9w0BAQsFADAiMSAwHgYDVQQDDBdj
b3J0ZXphLnlvdXItc2VydmVyLmNvbTAeFw0yMTEwMDIwNzAwNDVaFw0yMjEwMDIw
NzAwNDVaMCIxIDAeBgNVBAMMF2NvcnRlemEueW91ci1zZXJ2ZXIuY29tMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzb6QFbaa0IOpso8ZyhhWW6P2nwDn
N08LTFtHvRl0y0rspz2/PX5DSkUV+4I5Q2La21BZ8WAm28Ttr3BuUVQsnfDzPbSo
FSdgZYXcJXxeuaumwwtxpHpayvZj8zs0hyaie3diEMX7uqpGs/dL0pCTNmaI8nMo
LMjqJnYlQCz/HAUC5wrQHflfbLy8LA6KpCJuuTrFZGaMSIhW74HCYyp+2jTc6G1N
pxwBnwEqMy4RrYi5Mgn3GCPxo2LnSq3SVIurd5KLZb65YWHqAR4dKEncmdvVIbtS
8s9OgluSL6eXL364gXWW0DPs7saJdd8qclOfNI21Z4wr0PMVk9pyxSEvewIDAQAB
MA0GCSqGSIb3DQEBCwUAA4IBAQBnfhceUSfyRPZjrDixTcilyz8eLoWGOAqIsAOQ
Ai/D6/3mGLMOrIzEfhfkx7yHzwz0RnxNSxr6zdMf2vwWv8uCqf4oii51CXV3XllD
JnXVZjxjzuQbbQUaLHESx3qGpkWDrjMCqkLxTtLyQhG4oHT+re3C5sTodofyPc0Z
FiizNUs9CrdTmjUb43BpDyIZT9CXYbq75VHY/UB/ZtKBmD0PS524CTjegQ66BvW4
rL6Rri8GYpcnFNXZXGvJYeJOMT4U8nP5Bqo0bTV7YiRdx4pqVtroFFh304N4q7gl
aDlfJubiocv/fc4BiqCt/5cPiypAmR5mSTN2x5RjLBGqLhBN
-----END CERTIFICATE-----

.Следующий фрагмент показывает пример закрытого RSA-ключа.
```txt
```
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAzb6QFbaa0IOpso8ZyhhWW6P2nwDnN08LTFtHvRl0y0rspz2/
PX5DSkUV+4I5Q2La21BZ8WAm28Ttr3BuUVQsnfDzPbSoFSdgZYXcJXxeuaumwwtx
pHpayvZj8zs0hyaie3diEMX7uqpGs/dL0pCTNmaI8nMoLMjqJnYlQCz/HAUC5wrQ
HflfbLy8LA6KpCJuuTrFZGaMSIhW74HCYyp+2jTc6G1NpxwBnwEqMy4RrYi5Mgn3
GCPxo2LnSq3SVIurd5KLZb65YWHqAR4dKEncmdvVIbtS8s9OgluSL6eXL364gXWW
0DPs7saJdd8qclOfNI21Z4wr0PMVk9pyxSEvewIDAQABAoIBACLk3fUAylLw8Zf6
EyqmZAcY0Nv4wD4uJsFlfc5BggB0jZxzqXqZbnorK2ZDmMin/GxTvV1lrFF7ncAy
dlNOzl5fHjHp8NPdoMi3IjYtWLduuFK+HyEBK2Le1ObMtMTzNX1xKu2jVmU7OdCN
8YsdwAnq6/EKvNaToLqzMPAocX/jv7DDGUJ1r6LWwelV+RW6fbWHV9alnOwOQoPt
4o8r3+ZCLz1YsI+qW9cUBcbos1dkKxY9CdBtdtmFhUb+/tTx9/JRb6W+XJEAgCEM
/xxsu6pKdr9+Rc1YrwgMaMvDRzDQhe6eJM+l84W7axmqm07iuAuOFSE/b8yev/4e
8izw4fECgYEA9G6y/pk9fGxGgE3PRWVqq3/VsVNfNNjOMFUpFlQ3GlA+t0hSC+TU
XpxOufOI6WmV6bj+iVbhQk5iFTmW0+UeS+jHFSr0uzyavag6UJz8EuQfozAonkT2
AXtB9/85MWLo/Eo+CoIhofM+SEKBvcyrx6e+Mn+zEKssWAo5dh9Gd8MCgYEA13so
ZSnT/7vLaWZwCt2rxqHpnmWFwDF9nBWgiyT7ea3dym4DfXMTKIomqTVK9Vzbd9Oy
CmJols2otHtaV96hSi1z9tvfy/6om1k1ms75rFf4GatBjrVMzB/HtE2uGuWsW8Ez
F9BwDU9wX9qs+jCnG0fZm0nzef3lOHZIj3qlJekCgYBbzW/Am4EyR+A6s/6Sy8JC
YyK5FNz/FiZqlLF3x21inpzPbYQTH4B7gC05PbRAJf296FMA9fZoVtQTsKtrLfQx
Al4zHw0HfX2ImbQ9Lpil57PSMHYw6ymR6N8f62VpnQJwLtoaTEGhd5/+t6vOwx4J
QID4qmlwazmeX0ixipGGzQKBgQCq10JXsqoaf9HuZwE+HDIs8gI/S06X6qUkMyFu
MIwRFQBblo290JbH9YBhd5dOoah/gKAQC6XQqo2vSn1+XUyTeyYN+pWdLvKO+FO/
wYnCUpyp/VWkx6lzzV6QXWZEfQQCW1Me9mtgojL+TGoIkrpqrrSgoikf92TdNyqg
VyTIwQKBgGtVa5aJEhau3XUWKUAQbv95QDUNfP9PLs5jHhfflxVkAxlpMc9G4pNV
6mXFSsl8fnvg3HCmSr4rpV5QHUzvdKgbPkMQ8mUEONn/7ad4Xhsxeoz0xS5a8Tub
XhV4F/ngKrnQgxp47BCdNe0vC5Z44OImPpvQBsZ5PueqW+Pr5JHO
-----END RSA PRIVATE KEY-----

## Использование подписанных и зашифрованных запросов

Запрос `AuthNRequest` к IdP может быть подписан и зашифрован предоставленным сертификатом и отправлен через HTTP-Post binding на выбранный IdP.
Обратитесь к дополнительным примерам на веб-странице [samltool](https://www.samltool.com/generic*sso*req.php) за помощью.

Утверждение, возвращаемое от IdP, может быть подписано и зашифровано при условии, что сертификат на стороне поставщика услуг (LowCoooode) включён в метаданные SAML.
LowCoooode поддерживает подписанные сообщения и подписанные и зашифрованные утверждения.
Обратитесь к дополнительным примерам на веб-странице [samltool](https://www.samltool.com/generic*sso*res.php) за помощью.

.Включение подписанных запросов и ответов:
1. Флажок «Sign requests» в разделе «Certificate» должен быть включён (см. [saml-admin-ui,скриншот панели администрирования](#saml-admin-ui,скриншот панели администрирования)).
1. «Signature method» должен быть задан и зависит от типа алгоритма, указанного при создании сертификата (в этом примере использовался SHA256).

Для получения дополнительной информации о AuthNRequest и утверждении в метаданных обратитесь к [saml-developer-notes,заметкам разработчика](#saml-developer-notes,заметкам разработчика).

## Привязки SAML

LowCoooode поддерживает привязки SAML 2.0 **HTTP POST** и **HTTP Redirect**.

.Указание привязки SAML:
Выпадающий список «binding» должен быть выбран с указанием привязки *HTTP POST* или *HTTP Redirect* (см. [saml-admin-ui,скриншот панели администрирования](#saml-admin-ui,скриншот панели администрирования)).

!!! note
    Вариант по умолчанию — *HTTP Redirect*.


Для получения дополнительной информации о привязках в метаданных обратитесь к [saml-developer-notes,заметкам разработчика](#saml-developer-notes,заметкам разработчика).

## Настройка IdP в панели администрирования

Откройте панель администрирования и перейдите в menu:system[settings].

<a id="saml-admin-ui"></a>
.Скриншот показывает настройки SAML в панели администрирования в разделе menu:system[settings].
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/admin-settings.png",
    "alias": "authentication-external-saml-admin-settings.png",
    "w": 725,
    "h": 1253
  },
  "view": {},
  "annotations": []
}

.Заметки к форме настройки SAML:
1. Если вы отключите ваш SAML IdP, все настройки сохранятся в LowCoooode.
Пользователи больше не смогут использовать SAML для входа.
Все активные сессии входа, созданные с помощью SAML, сохраняются; пользователи не выходят из системы.
1. Имя отображается на форме входа (см. [saml-login-ui,скриншот ниже](#saml-login-ui,скриншот ниже)).
1. URL поставщика удостоверений должен указывать на URL сервера, который предоставляет метаданные.
Метаданные — это XML-документ, предоставляющий машиночитаемые инструкции для настройки SAML.
1. Поля «имя», «handle» и «идентификатор» должны быть указаны для того, чтобы пользователь был создан.
Когда пользователь успешно входит через SAML, LowCoooode выполняет поиск по значению из поля идентификатора, которое, как ожидается, будет адресом электронной почты.
Значения из полей имени и handle используются при создании нового пользователя.

.LowCoooode пытается угадать идентификатор среди одного из следующих:
- идентификатор по умолчанию, определённый в настройках
- `emailAddress`
- `urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress`
- `urn:oasis:names:tc:SAML:attribute:subject-id`
- `email`
- `mail`

!!! note
    Приведённые выше значения соответствуют отраслевому стандарту, но вам следует обратиться к документации вашего конкретного IdP, чтобы узнать, что предоставляется в составе данных пользователя.


## Экран входа с включённым SAML

<a id="saml-login-ui"></a>
.Скриншот показывает экран входа, когда включена SAML-аутентификация.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/login-screen.png",
    "alias": "authentication-external-saml-login-screen.png",
    "w": 1766,
    "h": 1370
  },
  "view": {},
  "annotations": []
}

При нажатии на соответствующую кнопку (в данном случае это btn:["Login with Example IdP"]) пользователь входит в поток SAML-аутентификации, в котором сервис (LowCoooode) и поставщики удостоверений (IdP) обмениваются информацией.
После успешного входа на IdP пользователь автоматически перенаправляется обратно в LowCoooode.

<a id="saml-developer-notes"></a>
## Заметки разработчика

При настройке поставщика SAML-аутентификации есть некоторая информация, которую можно прочитать из метаданных SP, генерируемых из конфигурации.

URL метаданных находится по адресу `$BASE_URL/auth/external/saml/metadata`.

.Пример получения метаданных с помощью curl:
[source,shell script]
curl $BASE_URL/auth/external/saml/metadata

.В зависимости от конфигурации можно прочитать следующую информацию:
[cols="1s,5a"]
|===
| [#logout]#[logout,Информация о выходе](#logout,Информация о выходе)#
|
```xml
```
<SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="$BASE*URL/auth/external/saml/slo" ResponseLocation="$BASE*URL/auth/external/saml/slo"></SingleLogoutService>

| [#authnrequest]#[authnrequest,Подпись AuthNRequest включена / отключена](#authnrequest,Подпись AuthNRequest включена / отключена)#
|
```xml
```
<SPSSODescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol" AuthnRequestsSigned="true" WantAssertionsSigned="true">

|===


:leveloffset: -1

<a id="idp-roles"></a>
## Настройка членства пользователя в ролях

Каждый из внешних поставщиков аутентификации поддерживает ограничение и корректировку членства пользователя в ролях при использовании конкретного внешнего поставщика аутентификации.

!!! important
    При совместном использовании с настройками безопасности клиента аутентификации сначала применяются настройки клиента аутентификации, затем настройки внешнего поставщика аутентификации.


Чтобы настроить членство в ролях, нажмите на значок редактирования рядом с внешним поставщиком аутентификации.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/settings.png",
    "alias": "authentication-external-settings-roles",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "w": 1599,
    "h": 1080
  },
  "focus": {
    "x": 576,
    "y": 518,
    "w": 1088,
    "h": 552
  },
  "annotations": [{
    "kind": "box",
    "x": 1610,
    "y": 728,
    "w": 20,
    "h": 20
  }]
}

В нижней части модального окна вы должны увидеть три поля ввода для разрешённых, запрещённых и принудительных ролей.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/roles.png",
    "alias": "authentication-external-roles",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "w": 1599,
    "h": 1080
  },
  "focus": {
    "x": 560,
    "y": 29,
    "w": 800,
    "h": 566
  },
  "annotations": [{
    "kind": "box",
    "x": 575,
    "y": 266,
    "w": 770,
    "h": 225
  }]
}

[cols="1s,5a"]
|===
| [#idp-roles-permitted]#[idp-roles-permitted,Разрешённые роли](#idp-roles-permitted,Разрешённые роли)#
| Список ролей, которые пользователям разрешено сохранять.

| [#idp-roles-prohibited]#[idp-roles-prohibited,Запрещённые роли](#idp-roles-prohibited,Запрещённые роли)#
| Список ролей, которые удаляются у пользователя.

| [#idp-roles-forced]#[idp-roles-forced,Принудительные роли](#idp-roles-forced,Принудительные роли)#
| Список ролей, которые добавляются пользователю при аутентификации с этим внешним поставщиком.
|===
