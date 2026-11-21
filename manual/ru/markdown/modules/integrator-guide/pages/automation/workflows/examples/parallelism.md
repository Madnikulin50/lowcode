# Параллелизм
:attachment-path: ../../../_attachments/automation/workflows/examples/
:page-noindex: true

В этом разделе приведены некоторые примеры того, как следует выполнять задачи параллельно.

## Безусловный параллелизм

Безусловный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно независимо от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример безусловного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/unconditional.png",
    "alias": "automation-workflows-examples-parallelism-unconditional",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 244,
    "y": 84,
    "w": 683,
    "h": 539
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/unconditional.json).

## Безусловный параллельный сегмент

Параллельный сегмент — это место, где рабочий процесс переходит от последовательного выполнения к параллельному и обратно к последовательному.

Безусловный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно независимо от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.
Завершите параллельный сегмент с помощью **соединяющего шлюза** (join gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример безусловного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/unconditional-segment.png",
    "alias": "automation-workflows-examples-parallelism-unconditional-segment",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 77,
    "w": 683,
    "h": 897
  },
    "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/unconditional-segment.json).

## Условный параллелизм

Условный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно в зависимости от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример условного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/conditional.png",
    "alias": "automation-workflows-examples-parallelism-conditional",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 76,
    "w": 683,
    "h": 601
  },
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/conditional.json).

## Условный параллельный сегмент

Параллельный сегмент — это место, где рабочий процесс переходит от последовательного выполнения к параллельному и обратно к последовательному.

Условный параллелизм следует использовать, когда две или более ветви выполнения должны выполняться параллельно в зависимости от состояния.

Для этого используйте **разветвляющий шлюз** (fork gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}, где каждая исходящая ветвь определяет одну ветвь параллельного выполнения.
Завершите параллельный сегмент с помощью **соединяющего шлюза** (join gateway) {ICON*WORKFLOW*GATEWAYPARALLEL}.

!!! caution
    Если какая-либо из ветвей определяет *шаг завершения* {ICON_WORKFLOW_TERMINATION}, весь рабочий процесс будет завершён.


.На скриншоте показан базовый пример условного параллельного выполнения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/examples/parallelism/conditional-segment.png",
    "alias": "automation-workflows-examples-parallelism-conditional-segment",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 214,
    "y": 57,
    "w": 898,
    "h": 870
  },
  "annotations": []
}

Исходный код рабочего процесса можно найти [здесь]({attachment-path}parallelism/conditional-segment.json).
