# Работа с записями
:attachment-path: ../../../_attachments/automation/workflows/examples/record/
:page-noindex: true

В этом разделе приведены некоторые советы и приёмы, которые можно использовать при работе с записями.

## Проверка существования

Если вы хотите выполнить какую-либо задачу в зависимости от наличия записей, вы можете использовать любой из следующих подходов.

Оба подхода допустимы, и нет никакой разницы в том, какой из них использовать.
Решайте исходя из своих предпочтений/контекста.

### Подход A

При поиске записей отметьте параметр `incTotal` и присвойте значение результата `total` переменной.

Внутри шлюза проверьте, больше ли значение `total` нуля.

.На скриншоте показан базовый пример проверки существования.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/existence-a.png",
    "alias": "automation-workflows-examples-record-existence-a",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 200,
    "y": 77,
    "w": 611,
    "h": 537
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}existence-a.json).

### Подход B

Внутри шлюза проверьте, больше ли значение `count(items)` нуля.

.На скриншоте показан базовый пример проверки существования.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/existence-b.png",
    "alias": "automation-workflows-examples-record-existence-b",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 201,
    "y": 77,
    "w": 609,
    "h": 535
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}existence-b.json).

## Создание или обновление

При создании записи нужно вызвать функцию `compose record create`, а при обновлении записи — функцию `compose record update`.

!!! note
    Только выделенная часть выполняет проверку создания/обновления; остальное — стандартный код для приведения в нужное состояние.


Если вам нужно вызвать ту или иную функцию на лету, можно использовать следующий подход.
Вы можете использовать `record.recordID != "0"`, чтобы определить, нужно ли обновлять запись — значением `recordID` по умолчанию является `"0"`.

.На скриншоте показан базовый пример проверки существования.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/create-update.png",
    "alias": "automation-workflows-examples-record-create-update",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 386,
    "y": 84,
    "w": 684,
    "h": 899
  },
  "annotations": [{
    "kind": "box-note",
    "x": 424,
    "y": 766,
    "w": 614,
    "h": 184
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}create-update.json).

## Удаление значения

Чтобы удалить какое-либо значение записи, используйте шаг выражения, чтобы задать соответствующему значению пустой `Any`.

.На скриншоте показан базовый пример удаления значений записи.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/remove-value.png",
    "alias": "automation-workflows-examples-record-remove-value",
    "w": 1920,
    "h": 1080
  },
 "view": {
    "x": 892,
    "y": 1,
    "w": 1027,
    "h": 405
  },
  "annotations": [{
    "kind": "box-note",
    "x": 1476,
    "y": 303,
    "w": 430,
    "h": 15
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}remove-value.json).

## Обработка отсутствующих значений

Чтобы использовать значение по умолчанию в случае отсутствия значения записи, нужно использовать оператор `??`.

Например, выражение `a ?? b` вернёт `a`, если оно существует, или `b`, если его нет.

!!! note
    В приведённом ниже примере в качестве значения по умолчанию используется переменная.
    Вы можете использовать константу, например `"something string"` или `42`.


.На скриншоте показан базовый пример использования значений по умолчанию.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/record/missing-value-default.png",
    "alias": "automation-workflows-examples-record-missing-value-default",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 892,
    "y": 1,
    "w": 1027,
    "h": 840
  },
  "annotations": [{
    "kind": "box-note",
    "x": 1476,
    "y": 304,
    "w": 431,
    "h": 494
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}missing-value-default.json).
