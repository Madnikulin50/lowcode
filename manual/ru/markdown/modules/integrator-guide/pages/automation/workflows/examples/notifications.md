# Уведомления
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true


LowCoooode предоставляет гибкую систему уведомлений, позволяющую отправлять уведомления для любого события, поддерживаемого триггерами рабочих процессов.

Это даёт вам полный контроль над:

- Когда отправляются уведомления
- Кто их получает
- Какую информацию они содержат

!!! important
    Убедитесь, что у пользователя, вызывающего рабочий процесс, есть разрешение на назначение уведомлений пользователям (меню:Admin Area[System > Permissions > `Allow notification assignment])


## Включение уведомлений

Сначала перейдите в админ-область (Admin Area) и откройте меню:User Interface[Settings], затем прокрутите вниз до раздела Topbar.

Убедитесь, что параметр `Hide notifications` не отмечен.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-settings.png",
    "alias": "automation-workflows-examples-notifications-notifications-settings.png",
    "w": 1258,
    "h": 657
  },
  "view": {},
  "annotations": []
}

Уведомления будут отображаться в верхней панели (topbar) приложения.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-topbar.png",
    "alias": "automation-workflows-examples-notifications-notifications-topbar.png",
    "w": 154,
    "h": 52
  },
  "view": {},
  "annotations": []
}

## Отправка простого уведомления

Чтобы отправить простое уведомление, вы можете использовать функцию `Send simple notification` внутри рабочего процесса.
Функция отправит уведомление получателю с указанными заголовком и описанием.

Она имеет следующие параметры:

- `recipient` — получатель уведомления.
- `title` — заголовок уведомления.
- `description` — описание уведомления.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-simple-workflow.png",
    "alias": "automation-workflows-examples-notifications-notifications-simple-workflow.png",
    "w": 1918,
    "h": 757
  },
  "view": {},
  "annotations": []
}

После выполнения рабочего процесса уведомление появится в `Notification Sidebar` (боковой панели уведомлений), доступ к которой осуществляется из верхней панели.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-simple.png",
    "alias": "automation-workflows-examples-notifications-notifications-simple.png",
    "w": 399,
    "h": 414
  },
  "view": {},
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}send-simple-notification.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Send simple notification**:
*** *recipient**:
****** **value type**: constant
****** **value**: `test-user`
*** *title**:
****** **value type**: constant
****** **value**: `Simple notification`
*** *description**:
****** **value type**: constant
****** **value**: `This is a simple notification`
3. **(3) Done**:
******

## Отправка уведомления о записи

Чтобы отправить уведомление о записи, вы можете использовать функцию `Send record notification` внутри рабочего процесса.
Функция отправит уведомление получателю с указанными заголовком и описанием.
Если на уведомление нажать, оно откроет запись в указанном режиме (модальное окно, новая вкладка, текущая вкладка).

Она имеет следующие параметры:

- `recipient` — получатель уведомления.
- `title` — заголовок уведомления.
- `description` — описание уведомления.
- `namespace` — пространство имён записи.
- `module` — модуль записи.
- `record` — запись.
- `openMode` — режим, в котором будет открыта запись (модальное окно, новая вкладка, текущая вкладка).
- `edit` — если true, запись будет открыта в режиме редактирования.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-record-workflow.png",
    "alias": "automation-workflows-examples-notifications-notifications-record-workflow.png",
    "w": 1918,
    "h": 873
  },
  "view": {},
  "annotations": []
}

После выполнения рабочего процесса уведомление появится в `Notification Sidebar` (боковой панели уведомлений), доступ к которой осуществляется из верхней панели.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/notifications/notifications-record.png",
    "alias": "automation-workflows-examples-notifications-notifications-record.png",
    "w": 397,
    "h": 373
  },
  "view": {},
  "annotations": []
}

Исходный код [примера рабочего процесса]({attachment-path}send-record-notification.json).

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onManual`
*** *enabled**: checked
2. **(2) Send record notification**:
*** *recipient**:
****** **value type**: constant
****** **value**: `test-user`
*** *title**:
****** **value type**: constant
****** **value**: `Record notification`
*** *description**:
****** **value type**: constant
****** **value**: `This is a record notification`
*** *namespace**:
****** **value type**: constant
****** **value**: `test-namespace`
*** *module**:
****** **value type**: constant
****** **value**: `test-module`
*** *record**:
****** **value type**: constant
****** **value**: `123`
*** *openMode**:
****** **value type**: constant
****** **value**: `modal`
*** *edit**:
****** **value type**: constant
****** **value**: `false`
3. **(3) Done**:
******
