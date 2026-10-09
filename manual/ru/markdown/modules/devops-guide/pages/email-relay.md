# Реле электронной почты

LowCoooode позволяет обнаруживать входящие письма и реагировать на них с помощью скриптов автоматизации.
Вы можете реализовать автоматические ответы, создавать записи на основе содержимого письма и пересылать письмо своему руководителю.

!!! note
    Здесь рассматривается только настройка реле электронной почты.
    Обратитесь к [руководству разработчика low-code платформы](modules/integrator-guide/pages/index.md) за подробностями об использовании этой функции.


!!! note
    *DevNote* добавьте небольшую наглядную диаграмму, описывающую это.


.Схема потока:
1. Postfix (или аналогичный) пересылает письмо на эндпоинт sink API,
1. сервис sink LowCoooode обрабатывает полезную нагрузку, чтобы подготовить событие скрипта автоматизации,
1. событие отправляется на шину событий, выполняя все скрипты автоматизации, соответствующие ограничениям события.

## Настройка LowCoooode

Внутренне реле электронной почты используют sink-маршруты, поэтому должна быть сгенерирована сигнатура sink-маршрута.
См. [Sink Route](modules/devops-guide/pages/sink-route.md) о том, как это настроить.

!!! important
    Обязательно укажите `--content-type email`.


## Настройка Postfix

!!! important
    Обязательно измените параметры в соответствии с вашей средой.


.Отредактируйте `/etc/postfix/main.cf`:
```text
```
virtual*alias*maps = pcre:/etc/postfix/virtual_alias

.Добавьте виртуальный алиас в `/etc/postfix/virtual_alias`:
```text
```
# Catch-all for lowcode.domain.tld and redirect it to lowcode_sink mailbox
/.+@lowcode\.domain\.tld$/ lowcode_sink

.Обновите карту/файл БД виртуальных алиасов и перезапустите postfix
```bash
```
postmap /etc/postfix/virtual_alias
postfix reload

.Добавьте запись в `/etc/aliases`
[source,bash,subs="attributes"]
lowcode*sink: "| curl --data-binary @- 'https://api.{API*DOMAIN}/system/sink?content-type=email&expires=&method=POST&origin=postfix&__sign=187...3D'"

!!! note
    Вышеперечисленное пересылает любое письмо для конкретного почтового ящика в команду curl, которая затем отправляет это исходное письмо на эндпоинт sink API LowCoooode.


.Обновите алиасы
```bash
```
newaliases

## Проверка изменений Postfix

!!! note
    Мы рекомендуем использовать другую машину — ту, на которой не запущен postfix.


Вы можете проверить, всё ли работает правильно, либо отправив письмо на настроенный адрес, либо с помощью простой команды CLI:

```bash
```
# Make sure to change `test@lowcode.domain.tld`.
echo "hello lowcode"|mail -s 'hello' test@lowcode.domain.tld

Это создаёт новую запись в вашем журнале почты (обычно `/var/log/mail.log`) для тестового письма, а также журнал, который выглядит примерно так:

```text
```
postfix/smtpd[23155]: connect from some-host.tld[xxx.xxx.xxx.xxx]
postfix/smtpd[23155]: 277AF5C1B78: client=some-host.tld[xxx.xxx.xxx.xxx]
postfix/cleanup[23159]: 277AF5C1B78: message-id=<b808218e-ce41-6cbf-cb4f-be2b4cf8f776@crust.tech>
postfix/qmgr[14490]: 277AF5C1B78: from=<sender@some-host.tld>, size=1476, nrcpt=1 (queue active)
postfix/smtpd[23155]: disconnect from some-host.tld[xxx.xxx.xxx.xxx] ehlo=2 starttls=1 mail=1 rcpt=1 data=1 quit=1 commands=7
postfix/local[23160]: 277AF5C1B78: to=<lowcode*sink@my-server>, orig*to=<demo@lowcode.domain.tld>, relay=local, delay=0.67, delays=0.03/0.01/0/0.62, dsn=2.0.0, status=sent (delivered to command:  curl --data-binary @- 'https://api.your-lowcode-instance.tld/system/sink?content-type=email&expires=&method=POST&origin=postfix&__sign=187...3D')
postfix/qmgr[14490]: 277AF5C1B78: removed

!!! tip
    Если при отправке письма ничего не происходит, возможно, проблема в межсетевом экране и заблокированных портах.


## Проверка LowCoooode

Чтобы проверить, правильно ли настроены ваша сигнатура sink и скрипт автоматизации, вы можете использовать следующую команду:

[source,bash,subs=attributes]
echo "
From: &lt;sender@lowcode.org&gt;
To: &lt;test@lowcode.domain.tld&gt;
Subject: hello
Message-ID: &lt;1234@local.machine.example&gt;

Ola LowCoooode!
" | curl -i --data-binary @- "https://api.{API*DOMAIN}/system/sink?content-type=email&expires=&method=POST&origin=postfix&*_sign=187...3D"

Если эта команда не возвращает ответ `200 OK`, это означает, что что-то настроено неправильно.
Обратитесь к системным журналам, чтобы увидеть, где именно находится проблема.
