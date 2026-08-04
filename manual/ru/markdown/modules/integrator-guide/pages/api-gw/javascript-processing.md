# Обработка запросов с помощью JavaScript

Фильтр обработки API GW поддерживает обработку запроса с помощью произвольного кода JavaScript.

Чтобы обрабатывать ваши запросы с помощью JavaScript, откройте веб-приложение Admin, перейдите в menu:system[integration gateway] и отредактируйте эндпоинт, для которого вы хотите определить код.

Нажмите на вкладку btn:[processing] в списке фильтров и выберите опцию btn:[payload processer].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/javascript-processing/define.png",
    "alias": "api-gw-javascript-processing-define",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "w": 1600,
    "h": 1080
  },
  "focus": {
    "x": 586,
    "y": 598,
    "w": 1070,
    "h": 295
  },
  "annotations": [{
    "kind": "box-note",
    "padding": "sm",
    "x": 699,
    "y": 666,
    "w": 110,
    "h": 20
  }, {
    "kind": "box-note",
    "padding": "sm",
    "y": 781,
    "x": 613,
    "w": 137,
    "h": 16
  }]
}

Когда вы добавляете payload processer, открывается всплывающее окно с редактором.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/javascript-processing/editor.png",
    "alias": "api-gw-javascript-processing-editor",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 578,
    "y": 51,
    "w": 766,
    "h": 253
  },
  "annotations": []
}

Введите фрагмент кода JavaScript, который вы хотите выполнить.
Предоставленный код JavaScript автоматически оборачивается в функцию с сигнатурой:

```ts
```
function (input: Scope): unknown {
}

Например, фрагмент кода `return "Hello, world!"` станет:

```js
```
function (input) {
  return "Hello, world!"
}

Следующий пример фрагмента кода принимает указанные `name` и `surname` наших пользователей и возвращает массив параметров `fullname`.

```js
```
var b = JSON.parse(readRequestBody(input.Get('request')));

return {
  "results":
    b.map(function({ name, surname }) {
        return {
          "fullname": name[0].toUpperCase() + name.substring(1) + " " + surname[0].toUpperCase() + surname.substring(1)
        }
    }),
  "count": b.length
};

Следующий [#apigw-proc-js-example]#[apigw-proc-js-example,пример cURL](#apigw-proc-js-example,пример cURL)# вызывает вышеуказанную функцию JavaScript.

```bash
```
curl -X GET $BASE_URL/api/test-js \
  -H 'Content-Type: application/json' \
  -d '[{"name":"johnny","surname":"mnemonic"},{"name":"johnny","surname":"knoxville"}]';

## Аргументы функции

Фрагмент кода получает один аргумент, `input`, который содержит весь запрос.

.Аргумент имеет следующую сигнатуру:
```ts
```
interface {
  Set: (k: string, v: unknown) => void;
  Get: (k: string) => unknown;
}

.Параметры объекта `input`:
[cols="1m,5a"]
|===
| [#proc-js-input-request]#[proc-js-input-request,request](#proc-js-input-request,request)#
|
Это полный объект HTTP-запроса.
Обратитесь к [документации GO](https://pkg.go.dev/net/http?utm_source=gopls#Request) для получения подробной информации о сигнатуре.

| [#proc-js-input-opts]#[proc-js-input-opts,opts](#proc-js-input-opts,opts)#
|
Конфигурация Integration Gateway.
Обратитесь к [исходному коду](https://github.com/lowcode/lowcode-server/blob/{PAGE-VERSION}.x/pkg/options/options.gen.go#L84) для получения подробной информации.

!!! note
    *DevNote* создать документацию для указанной выше опции


|===

## Результат функции

Результат фрагмента JavaScript передается постфильтрам, которые подготавливают и возвращают HTTP-ответ.

Когда результат является `string`, например `Hello, world!`, результат используется как есть.

Когда результат не является строковым значением, например `{ key: "value" }`, результат JSON-кодируется.

## Справочник встроенных функций

[cols="1m,5m,5a,5a"]
|===
|Функция |Сигнатура |Описание |Пример

| [#proc-js-fnc-ref-readRequestBody]#[proc-js-fnc-ref-readRequestBody,readRequestBody](#proc-js-fnc-ref-readRequestBody,readRequestBody)#
| readRequestBody(input: HttpRequest): string
| Функция возвращает содержимое предоставленного объекта запроса.
| `readRequestBody(input.Get('request'))` возвращает тело запроса в виде строки

|===
