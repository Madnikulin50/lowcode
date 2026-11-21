# Работа с электронной почтой
:attachment-path: ../../_attachments/automation/workflows/

!!! important
    Для отправки электронных писем необходимо настроить LowCoooode с вашим SMTP-провайдером.


## Отправка писем напрямую

Используйте шаг функции **email** {ICON*WORKFLOW*FUNCTION}, чтобы немедленно отправить электронное письмо.

Если вы вызываете функцию email без указания аргумента `from`, используется переменная `SMTP_FROM` из файла `.env`.

.Пример рабочего процесса, настроенного на отправку электронного письма:
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/email-send/email-direct-ex-1.png",
    "alias": "automation-workflows-email-send-email-direct-ex-1.png",
    "w": 642,
    "h": 240
  },
  "view": {},
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}email*direct*send.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Send Email**:
*** *type**: `Email`
*** *arguments**:
**** *subject**
****** **value type**: constant
****** **value**: `Hello, world!`
**** *replyTo**
****** **value type**: constant
****** **value**: `replyto@test.tld`
**** *from**
****** **value type**: constant
****** **value**: `from@test.tld`
**** *to**
****** **value type**: constant
****** **value**: `to@test.tld`
**** *cc**
****** **value type**: constant
****** **value**: `cc@test.tld`
**** *html**
****** **value type**: constant
****** **value**: `br>Hello, world!</br>`
**** *plain**
****** **value type**: constant
****** **value**: `Hello, world!`
******

.Итоговое электронное письмо:
```
```
Cc: cc@test.tld
Content-Type: multipart/alternative; boundary=4305315520b0c6018a31f71ea361d14aba596d49cb6041dcf323dbe83440
Date: Sun, 29 Aug 2021 11:16:06 +0200
From: from@test.tld
MIME-Version: 1.0
Message-ID: omDfvBmYViB5Jo2Y-MESerSmRIi0Z0a7wt9z*LaX*wk=@mailhog.example
Received: from localhost by mailhog.example (MailHog)
          id omDfvBmYViB5Jo2Y-MESerSmRIi0Z0a7wt9z*LaX*wk=@mailhog.example; Sun, 29 Aug 2021 09:16:06 +0000
ReplyTo: replyto@test.tld
Return-Path: <from@test.tld>
Subject: Hello, world!
To: to@test.tld

--4305315520b0c6018a31f71ea361d14aba596d49cb6041dcf323dbe83440
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain; charset=UTF-8

Hello, world!
--4305315520b0c6018a31f71ea361d14aba596d49cb6041dcf323dbe83440
Content-Transfer-Encoding: quoted-printable
Content-Type: text/html; charset=UTF-8

<br>Hello, world!</br>
--4305315520b0c6018a31f71ea361d14aba596d49cb6041dcf323dbe83440--

## Конструктор писем

Шаг функции **email builder** (конструктор писем) {ICON*WORKFLOW*FUNCTION} позволяет собрать электронное письмо без его отправки.

Конструктор писем позволяет дополнительно обогатить письмо такими параметрами, как другие получатели.

.Пример рабочего процесса, созданного для формирования письма с помощью конструктора писем.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/email-send/email-builder-ex-1.png",
    "alias": "automation-workflows-email-send-email-builder-ex-1.png",
    "w": 561,
    "h": 642
  },
  "view": {},
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}email*builder*send.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked

2. **(2) Prepare Email**:
*** *type**: `Email`
*** *arguments**:
**** *subject**
****** **value type**: constant
****** **value**: `Hello, world!`
**** *html**
****** **value type**: constant
****** **value**: `br>Hello, world!</br>`
**** *plain**
****** **value type**: constant
****** **value**: `Hello, world!`
*** *results**:
**** *message target**: `msg`

3. **(3) Add recipient**:
*** *type**: `Email add address`
*** *arguments**:
**** *message**
****** **value type**: expression
****** **value**: `msg`
**** *type**
****** **value type**: constant
****** **value**: `To`
**** *address**
****** **value type**: constant
****** **value**: `testko@test.tld`
*** *name 
****** **value type**: constant
****** **value**: `testko`

4. **(4) Add recipient**:
*** *type**: `Email add address`
*** *arguments**:
**** *message**
****** **value type**: expression
****** **value**: `msg`
**** *type**
****** **value type**: constant
****** **value**: `To`
**** *address**
****** **value type**: constant
****** **value**: `testko2@test.tld`
*** *name 
****** **value type**: constant
****** **value**: `testko2`

5. **(5) Send message**:
*** *type**: `Email sender`
*** *arguments**:
**** *message**
****** **value type**: expression
****** **value**: `msg`
******

.Итоговое электронное письмо:
```
```
Content-Type: multipart/alternative; boundary=122730f4fe6b767154618c2972886463a224ea74df15f5c96662da9e6a70
Date: Sun, 29 Aug 2021 11:51:33 +0200
From: LowCoooode <info@local.lowcode.org>
MIME-Version: 1.0
Message-ID: m7mFhiH28t1F8HCaTB7g3CKMKH4qE_N4J6uqLDktytE=@mailhog.example
Received: from localhost by mailhog.example (MailHog)
          id m7mFhiH28t1F8HCaTB7g3CKMKH4qE_N4J6uqLDktytE=@mailhog.example; Sun, 29 Aug 2021 09:51:33 +0000
Return-Path: <info@local.lowcode.org>
Subject: Hello, world!
To: "testko" <testko@test.tld>, "testko2" <testko2@test.tld>

--122730f4fe6b767154618c2972886463a224ea74df15f5c96662da9e6a70
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain; charset=UTF-8

Hello, world!
--122730f4fe6b767154618c2972886463a224ea74df15f5c96662da9e6a70
Content-Transfer-Encoding: quoted-printable
Content-Type: text/html; charset=UTF-8

<b>Hello, world!</b>
--122730f4fe6b767154618c2972886463a224ea74df15f5c96662da9e6a70--

### Методы конструктора писем

[cols="2s,5a"]
|===
| [#email-builder-fnc-message]#[email-builder-fnc-message,Конструктор писем](#email-builder-fnc-message,Конструктор писем)#
|
Функция возвращает собранное электронное письмо, которое вы можете впоследствии расширить с помощью следующих функций.

| [#email-builder-fnc-sendMessage]#[email-builder-fnc-sendMessage,Отправка письма](#email-builder-fnc-sendMessage,Отправка письма)#
|
Функция отправляет заданное электронное письмо.

| [#email-builder-fnc-setSubject]#[email-builder-fnc-setSubject,Тема письма](#email-builder-fnc-setSubject,Тема письма)#
|
Функция задаёт тему письма.

| [#email-builder-fnc-setHeaders]#[email-builder-fnc-setHeaders,Заголовки письма](#email-builder-fnc-setHeaders,Заголовки письма)#
|
Функция перезаписывает заголовки, изначально заданные функцией [email-builder-fnc-message,конструктора писем](#email-builder-fnc-message,конструктора писем).

| [#email-builder-fnc-setHeader]#[email-builder-fnc-setHeader,Заголовок письма](#email-builder-fnc-setHeader,Заголовок письма)#
|
Функция перезаписывает заголовок, изначально заданный функцией [email-builder-fnc-message,конструктора писем](#email-builder-fnc-message,конструктора писем).

| [#email-builder-fnc-setAddress]#[email-builder-fnc-setAddress,Задать адрес](#email-builder-fnc-setAddress,Задать адрес)#
|
Функция перезаписывает адреса электронной почты указанного типа получателя предоставленным адресом.

.Доступные типы получателей:
- `To`
- `Cc`
- `ReplyTo`
- `From`

| [#email-builder-fnc-addAddress]#[email-builder-fnc-addAddress,Добавить адрес](#email-builder-fnc-addAddress,Добавить адрес)#
|
Функция добавляет предоставленный адрес электронной почты в список указанного типа получателей.

.Доступные типы получателей:
- `To`
- `Cc`
- `ReplyTo`
- `From`

| [#email-builder-fnc-attach]#[email-builder-fnc-attach,Вложение в письмо](#email-builder-fnc-attach,Вложение в письмо)#
|
Функция добавляет вложение в электронное письмо.

| [#email-builder-fnc-embed]#[email-builder-fnc-embed,Встроенное вложение](#email-builder-fnc-embed,Встроенное вложение)#
|
Функция встраивает вложение в электронное письмо.

|===

## Форматы адресов

.При передаче адресов электронной почты поддерживаются следующие форматы:
`test@mail.tld name`::
  Имя может следовать за адресом электронной почты.
  См. примеры ниже:
  - `test@mail.tld Jane Doe`
  - `test@mail.tld Jane`
  - `test@mail.tld`

`{"test@mail.tld": "name"}`::
  Значение `KV`, где ключ — адрес электронной почты, а значение — имя.
  Примеры:
  - `{"test@mail.tld": "Jane Doe"}`
  - `{"test@mail.tld": "Jane Doe", "test2@mail.tld": "John"}`

## Отправка вложений

Для отправки вложений обратитесь к [примеру рабочего процесса](modules/integrator-guide/pages/automation/workflows/automation/workflows/examples/email-attachment.md).
