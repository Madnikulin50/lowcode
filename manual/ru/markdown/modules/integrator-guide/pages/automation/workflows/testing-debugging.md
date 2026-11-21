# Тестирование и отладка рабочих процессов

<a id="testing"></a>
## Тестирование

Тестирование — мощный и необходимый компонент любого проекта.

!!! caution
    Описанный ниже подход к тестированию никак не изолирует ваши данные.
    Любые изменения, выполняемые рабочим процессом, отражаются в системе.


Рассмотрим пример.
Вы хотите реализовать рабочий процесс, который вычисляет стоимость только что созданного лида.
Стоимость лида отправляется администратору по электронной почте.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/manual-test-play.png",
    "alias": "automation-workflows-testing-debugging-manual-test-play.png",
    "w": 978,
    "h": 145
  },
  "view": {},
  "annotations": []
}

Чтобы протестировать рабочий процесс, нажмите на значок запуска {ICON*WORKFLOW*PLAY} в правом верхнем углу триггера.
Запускать можно только сохранённые и корректные рабочие процессы.

Появится всплывающее окно с запросом начальной области видимости.
Начальная область видимости — это данные, передаваемые рабочему процессу для выполнения операции.

!!! note
    Начальная область видимости зависит от типа ресурса и события.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/manual-test-load-resources.png",
    "alias": "automation-workflows-testing-debugging-manual-test-load-resources.png",
    "w": 745,
    "h": 705
  },
  "view": {},
  "annotations": []
}

Вставьте требуемые параметры и нажмите на кнопку btn:[Load and Configure], чтобы задать начальную область видимости.
Любые недостающие параметры автоматически добавляются как пустое значение этого типа.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/manual-test-configure-resources.png",
    "alias": "automation-workflows-testing-debugging-manual-test-configure-resources.png",
    "w": 743,
    "h": 690
  },
  "view": {},
  "annotations": []
}

Последний шаг позволяет изменить переменные начальной области видимости перед тестовым запуском рабочего процесса.
Нажмите на кнопку btn:[Run Workflow], чтобы запустить рабочий процесс, или на кнопку btn:[Back], чтобы вернуться на предыдущий экран.

При ручном запуске рабочего процесса вы увидите зелёную подсветку, указывающую путь выполнения на основе переданных входных данных.
Если какой-либо шаг завершился с ошибкой, он будет помечен красной подсветкой.
Если обработка ошибок не настроена, ручное выполнение также прекращается.

.Пример успешного теста:
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/manual-test-run-ok.png",
    "alias": "automation-workflows-testing-debugging-manual-test-run-ok.png",
    "w": 1012,
    "h": 146
  },
  "view": {},
  "annotations": []
}

.Пример неудачного теста:
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/manual-test-run-nok.png",
    "alias": "automation-workflows-testing-debugging-manual-test-run-nok.png",
    "w": 1397,
    "h": 288
  },
  "view": {},
  "annotations": []
}

В некоторых случаях рабочие процессы могут содержать шаги, которые должны выполняться только в производстве, например списание средств с клиентов, массовые уведомления по электронной почте и тому подобное.
Мы можем имитировать этот путь выполнения, определив в начале флаг `testing`.
Перед шагами, предназначенными «только для производства», следует использовать исключающий шлюз для выполнения особой обработки, например записи сообщения в журнал или использования [debug-step,шага отладки {ICON*WORKFLOW*DEBUG}](#debug-step,шага отладки {ICON*WORKFLOW*DEBUG}).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/manual-test-special.png",
    "alias": "automation-workflows-testing-debugging-manual-test-special.png",
    "w": 1370,
    "h": 399
  },
  "view": {},
  "annotations": []
}

Когда вы запустите рабочий процесс после этого шага, вы должны увидеть что-то вроде примера ниже:

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/manual-test-special-ok.png",
    "alias": "automation-workflows-testing-debugging-manual-test-special-ok.png",
    "w": 1371,
    "h": 438
  },
  "view": {},
  "annotations": []
}

## Отладка

Умение отлаживать конкретный инструмент или программное обеспечение может сэкономить вам много времени при возникновении проблемы.

<a id="debug-step"></a>
### Шаг отладки

Шаг отладки {ICON*WORKFLOW*DEBUG} позволяет просмотреть содержимое области видимости, в которой он находится.

Шаг отладки использует журнал, настроенный сервером, и не виден в самом редакторе рабочих процессов.

.Простой пример:
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/debug-example-simple.png",
    "alias": "automation-workflows-testing-debugging-debug-example-simple.png",
    "w": 829,
    "h": 144
  },
  "view": {},
  "annotations": []
}

### Контекст результатов теста

Когда вы [testing,запускаете ручной тест](#testing,запускаете ручной тест) рабочего процесса, вы можете заметить значок в правом нижнем углу шага.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/debug-stepctx.png",
    "alias": "automation-workflows-testing-debugging-debug-stepctx.png",
    "w": 325,
    "h": 143
  },
  "view": {},
  "annotations": []
}
При наведении курсора на значок вы увидите важную информацию о выполнении шага.

Если навести курсор на успешно выполненный шаг (шаг с зелёной подсветкой), вы увидите продолжительность выполнения.
Знание продолжительности выполнения может помочь выявить узкие места системы.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/debug-stepctx-ok.png",
    "alias": "automation-workflows-testing-debugging-debug-stepctx-ok.png",
    "w": 434,
    "h": 184
  },
  "view": {},
  "annotations": []
}

Если навести курсор на шаг с ошибкой (шаг с красной подсветкой), вы увидите продолжительность выполнения и возникшую ошибку.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/testing-debugging/debug-stepctx-nok.png",
    "alias": "automation-workflows-testing-debugging-debug-stepctx-nok.png",
    "w": 1127,
    "h": 159
  },
  "view": {},
  "annotations": []
}
