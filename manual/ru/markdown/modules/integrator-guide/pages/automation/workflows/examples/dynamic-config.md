# Динамическая конфигурация
:page-noindex: true
:attachment-path: ../../../_attachments/automation/workflows/examples/

При определении автоматизации, которая должна взаимодействовать с внешними системами, или когда вам нужно сделать выполнение рабочего процесса настраиваемым, статические рабочие процессы могут оказаться неудобными.

Вы можете определить модуль `settings`, в котором определите все настраиваемые параметры, необходимые вашей автоматизации.
Это может быть что угодно: от URL-адресов до учётных данных для входа и токенов доступа.

!!! caution
    При хранении токенов доступа и других учётных данных обязательно правильно настройте контроль доступа.


.На скриншоте показан базовый пример модуля `settings`.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/dynamic-config/settings-module.png",
    "alias": "automation-workflows-examples-dynamic-config-settings-module",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Внутри рабочего процесса просто получите запись из модуля `settings` и настройте выполнение, используя её значения.

.На скриншоте показан базовый пример рабочего процесса, использующего модуль `settings`.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/dynamic-config/example-workflow.png",
    "alias": "automation-workflows-examples-dynamic-config-example-workflow",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 358,
    "y": 183,
    "w": 610,
    "h": 395
  },
  "annotations": [{
    "kind": "box-note",
    "x": 679,
    "y": 254,
    "w": 183,
    "h": 74
  }]
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}dynamic-config/example-workflow.json).
