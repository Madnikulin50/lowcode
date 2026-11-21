# Устранение неполадок

Если у вас возникают проблемы при настройке или использовании LowCoooode, вы можете воспользоваться следующими инструментами и базами знаний, чтобы решить вашу проблему.

LowCoooode вводит [Web Console](modules/devops-guide/pages/troubleshooting/troubleshooting/web-console.md), которая предоставляет простой способ доступа к важной информации, такой как журналы сервера.
Веб-консоль можно использовать для попытки определить, что вызывает проблему.

Если вам не удаётся решить проблему, вы можете обратиться к сообществу на [нашем форуме](https://forum.lowcode.org/).

## Порты недоступны

При запуске различных сервисов на вашей машине часто бывает, что порты уже заняты.
Вы увидите что-то вроде этого:

```
```
Cannot start service server: Ports are not available: listen tcp 127.0.0.1:18080: bind: address already in use

Если вы видите эту ошибку, вы можете изменить номер порта на число в диапазоне от `1024` до `65535`.
Вы также можете заменить значение `services.server.ports` в `docker-compose.yaml` на `["80"]`, и тогда Docker сам выберет доступный порт.

<a id="ws-nginx-connection-fail"></a>
## Не удаётся установить WebSocket-подключение с Nginx

Если WebSocket-подключение не устанавливается, возможно, вам нужно включить WebSocket-проксирование в Nginx.

!!! note
    Подробные инструкции и дополнительные примеры вы найдёте в https://nginx.org/en/docs/http/websocket.html[документации Nginx].


Внутри вашего файла `nginx.conf` (по умолчанию он находится в каталоге `/etc/nginx`) добавьте следующие строки в секцию конфигурации `server`;

```
```
location /api/websocket {
  proxy_pass http://server:80;
  proxy*http*version 1.1;
  proxy*set*header Upgrade $http_upgrade;
  proxy*set*header Connection "Upgrade";
  proxy*set*header Host $host;
}

!!! caution
    Обязательно скорректируйте location, если вы определили пользовательскую конфигурацию, влияющую на базовый путь.


## Подключение к серверу Corredor

В журналах серверного контейнера (`docker-compose logs -f server`) вы можете увидеть одну или несколько ошибок connection refused:

```
```
{"level":"error","ts":1608125024.4714684,"logger":"corredor","caller":"corredor/service.go:427","msg":"could not load corredor server scripts","error":"rpc error: code = DeadlineExceeded desc = latest balancer error: connection error: desc = \"transport: Error while dialing dial tcp 172.23.0.2:80: connect: connection refused\"","stacktrace":"github.com/lowcode/lowcode-server/pkg/corredor.(*service).loadServerScripts\n\t/drone/src/pkg/corredor/service.go:427"}

Если при запуске сервера возникает пара ошибок, это нормально.
Иногда серверу Corredor требуется больше времени для запуска, и сервер LowCoooode ещё не может к нему подключиться.

Если проблема сохраняется и вы видите состояние Corredor как healthy, проверьте изменения, которые вы могли внести в конфигурацию.

## Сетевой прокси объявлен как внешний

```
```
ERROR: Network proxy declared as external, but could not be found. Please create the network manually using `docker network create proxy` and try again.

Убедитесь, что ваш сервис nginx-proxy запущен до запуска LowCoooode.

```bash
```
$ docker network create proxy

## Пустой экран на `/auth` с ошибкой «state does not match» в консоли браузера

Ошибка `state does not match` обычно возникает в среде разработки или там, где сервер LowCoooode часто перезапускается.

Состояние аутентификации, которое передаётся через поток аутентификации пользователя, не является постоянным и будет потеряно при перезапуске сервера.
Если ошибка возникает и сохраняется в производственной среде, проверьте ваши настройки `.env` `AUTH*SESSION**` (если вы их изменяли).

## Застревание на экране входа после ввода действительных учётных данных

Обратите внимание, что некоторые развёртывания (например, с Docker) могут не иметь всей конфигурации для корректной автоматической настройки.
Проверьте, какие куки отправляются в заголовках ответа при входе.

*Обратите внимание, что в приведённых ниже примерах значение сессии сокращено.*

### Сломанная конфигурация: пользователь застревает на экране входа

```
```
set-Cookie:
    session=MTYzODQ...tCLvO_DHhw==;
    Path=/auth;
    Domain=e3a47cb50c17; <1>
    Expires=Sun, 27 Nov 2022 10:13:08 GMT;
    Max-Age=31104000;
    HttpOnly
<1> Набор случайных символов там, где домен или имя хоста должны представлять ID Docker-контейнера (и имя хоста).
Такая настройка требует установки переменной `DOMAIN` (например, `DOMAIN=localhost:8080`).
Не забудьте пересоздать контейнер сервера после изменения.

### Пример допустимого куки

```
```
set-Cookie: <1>
    session=MTYzODQ...tCLvO_DHhw==;
    Path=/auth;
    Expires=Sun, 27 Nov 2022 12:58:01 GMT;
    Max-Age=31104000;
    HttpOnly
<1> Когда вы указываете опцию `DOMAIN` так, как указано в примере выше, обратите внимание, что флаг «Domain» у куки отсутствует.
LowCoooode удаляет его, когда вы используете порт в домене.

### Допустимый пример куки на защищённом хосте

```
```
set-cookie:
    session=MTYzODQ...tCLvO_DHhw==;
    Path=/auth;
    Domain=lowcode.example.org;
    Expires=Fri, 03 Dec 2021 12:58:46 GMT;
    Max-Age=86400;
    HttpOnly;
    Secure <1>
<1> Обратите внимание на флаг куки «Secure».
Этот куки отправляется обратно на сервер только в том случае, если сервер находится на защищённом домене (HTTPS).
Если вы использовали опции `HTTP*SSL*TERMINATED` или `LETSENCRYPT_HOST`, LowCoooode предполагает, что он обслуживается на защищённом домене.

## Дальнейшее устранение неполадок

Если у вас продолжают возникать проблемы с LowCoooode, мы призываем вас связаться с другими пользователями на нашем [community-сервере](https://latest.lowcode.org).
Скорее всего, вы найдёте кого-то, кто сможет вам помочь.
Вы также можете открыть issue в нашем репозитории [lowcode/lowcode-server](https://github.com/lowcode/lowcode-server) на GitHub.

## SMTP не работает

Вы можете использовать команду `lowcode-server auth test-notifications`, чтобы убедиться, что ваша конфигурация SMTP работает корректно.
Команда отправляет тестовое письмо на указанный адрес электронной почты.

```bash
```
lowcode-server auth test-notifications your-email@example.tld

В зависимости от вашего провайдера, доставка письма может занять разное время.
Если вы не видите письмо, обязательно проверьте папку со спамом и журналы сервера.

.Ошибки подключения и аутентификации выглядят так:
```
```
could not send email: dial tcp [::1]:25: connect: connection refused

## Изменения настроек клиента аутентификации по умолчанию не применились

Перезапуск сервера требуется, когда администратор изменяет настройки клиента аутентификации по умолчанию, такие как секрет клиента или redirect URIs, поскольку эти изменения не применяются немедленно.

## Discovery не запускается

Если ваш Discovery не работает, попробуйте перезапустить сервис `discovery`.
Существует известная проблема с порядком выполнения и отчётами о проверке работоспособности, из-за которой сервис может у вас не работать.

## Ошибка `client does not support authorization_code flow` при попытке войти

Эта ошибка возникает, потому что LowCoooode ожидает `authorization*code`, но ваш клиент аутентификации может использовать `client*credentials`.
Если у вас всё ещё есть доступ к LowCoooode, перейдите в Admin и обновите клиент вашего webapp LowCoooode по умолчанию, чтобы он использовал `authorization_code` (по умолчанию хендл клиента — `lowcode-webapp`).

Если у вас нет доступа к LowCoooode, выполните следующий SQL-запрос:

```sql
```
UPDATE auth*clients SET valid*grant = 'authorization_code' WHERE id = your-auth-client-id-here};

Возможно, вам потребуется перезапустить LowCoooode, чтобы изменения вступили в силу.
