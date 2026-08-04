# Загрузка файлов с помощью шлюза интеграции (Integration Gateway)
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true

Шлюзы интеграции в сочетании с Workflow позволяют пользователям обрабатывать загрузку файлов с помощью `multipart/form-data`.

.В этом примере мы:
- создадим рабочий процесс, который извлекает загруженный файл,
- создадим эндпоинт шлюза интеграции.

## Создание рабочего процесса

Чтобы создать рабочий процесс, сначала откройте Workflow и создайте новый рабочий процесс.
Основной рабочий процесс требует двух частей, а в примере добавлены несколько дополнительных шагов для прикрепления файла к записи.

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/workflow.png)

.Рабочий процесс требует этих двух шагов:
- **System / on manual** триггер: этот триггер позволяет шлюзу интеграции выполнять конкретный рабочий процесс.
- **Read file from integration gateway** функция:: эта функция извлекает файл из предоставленного HTTP-запроса:
** Аргумент `request` получает HTTP-запрос; при использовании со шлюзами интеграции запрос предоставляется в переменной `request`.
** Аргумент `name` задаёт имя поля `multipart/form-data`,
** Выходной аргумент `file` предоставляет `Reader` файла.

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/step-config.png)

`Reader` содержит содержимое файла и может использоваться по вашему усмотрению.

## Создание нового эндпоинта шлюза интеграции

Чтобы создать новый эндпоинт, перейдите в веб-приложение Admin и нажмите на пункт меню «Integration Gateway», затем нажмите на кнопку btn:[New Route].

Укажите эндпоинт (в примере для загрузки файлов используется `/fup`).
Выберите метод `POST` и обязательно отметьте шлюз как включённый.
Нажмите на кнопку btn:[Submit], чтобы инициализировать эндпоинт.

В новом разделе «filter list» используйте следующие параметры:
- **Prefilters** (предварительные фильтры):
** При желании выберите и включите «Profiler» (это помогает убедиться, что эндпоинт обрабатывается).

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/prefilter.png)

- **Processing** (обработка):
** Выберите и настройте «Workflow processor» (используйте рабочий процесс, подготовленный на предыдущем шаге).

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/processing.png)

- **Postfiltering** (постфильтрация):
** Добавьте «default JSON response», чтобы мы могли видеть статусы ответов и возможные ошибки

![role="data-zoomable"](automation/workflows/examples/gateway-file-upload/postfilter.png)

## Тестирование

Для тестирования вы можете отправить HTTP-запрос `POST` на эндпоинт your-lowcode-instance.tld/api/gateway/YOUR*ENDPOINT*HERE.
Вы можете использовать следующий шаблон cURL:

```shell
```
curl -v -X POST http://localhost:18080/api/gateway/fup \
  -F "file=@/path/to/your/file.txt"

В случае успеха ответ должен выглядеть следующим образом:

```txt
```
- Host localhost:18091 was resolved.
- IPv6: ::1
- IPv4: 127.0.0.1
-   Trying [::1]:18091...
- connect to ::1 port 18091 from ::1 port 58275 failed: Connection refused
-   Trying 127.0.0.1:18091...
- Connected to localhost (127.0.0.1) port 18091
> POST /api/gateway/fup HTTP/1.1
> Host: localhost:18091
> User-Agent: curl/8.7.1
> Accept: */*
> Content-Length: 206
> Content-Type: multipart/form-data; boundary=------------------------Caf19XAT8xI5uYCDVXRkRz
> 
- upload completely sent off: 206 bytes
< HTTP/1.1 202 Accepted
< Content-Type: application/json
< Vary: Origin
< Vary: Origin
< X-Request-Id: 4b9179be911f/FwhKqZfCsa-000622
< Date: Tue, 22 Jul 2025 12:47:20 GMT
< Content-Length: 2
< 
- Connection #0 to host localhost left intact
{}%
