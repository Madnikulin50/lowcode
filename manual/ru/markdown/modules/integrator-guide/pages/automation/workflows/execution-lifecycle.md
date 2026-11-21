# Жизненный цикл выполнения

**@todo**

Сессия автоматизации описывает автоматизацию, которая выполнена или всё ещё выполняется.
Сессии можно использовать для отладки неработающей автоматизации и для выявления проблем бесконечного цикла, препятствующих нормальной работе системы.

.Чтобы вывести список текущих сессий:
1. перейдите в menu:system[sessions],
1. при необходимости введите параметры фильтрации (список обновляется автоматически).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/execution-lifecycle/sessions-list.png",
    "alias": "automation-workflows-execution-lifecycle-sessions-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 717,
    "w": 288,
    "h": 37
  }]
}

## Состояния сессий

!!! note
    *DevNote*: обратите внимание на состояния
    
    * completed: Completed
    * excluded: Without
    * exclusive: Only
    * failed: Failed
    * inProgress: completed sessions
    * inclusive: Including
    * prompted: Prompted
    * sessions: sessions
    * started: Started
    * suspended: Suspended
