# Журналирование

LowCoooode регистрирует большинство операций, произошедших в системе, в **журнале действий**.
Пользовательский интерфейс **журнала действий** предоставляет обзор событий, таких как регистрация или вход пользователей, создание записей и рендеринг шаблонов.

!!! tip
    Ошибки, сообщаемые сервером Corredor, также отображаются в журнале действий.


Вы можете использовать журнал действий для отладки и обнаружения подозрительного поведения, поскольку он предоставляет полное представление о произошедшем.

## Интерфейс журнала действий

Пользовательский интерфейс журнала действий находится в веб-приложении [LowCoooode Admin](modules/integrator-guide/pages/troubleshooting/index.md#webapp-admin), в меню:system[журнал действий].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "troubleshooting/logging/action-log.png",
    "alias": "troubleshooting-logging-action-log",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 437,
    "w": 288,
    "h": 37
  }]
}

## Просмотр действий

.Чтобы просмотреть текущие записи журнала действий:
1. Перейдите в меню:system[журнал действий],
1. при необходимости введите параметры фильтрации и нажмите на кнопку btn:[поиск].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "troubleshooting/logging/action-log.png",
    "alias": "troubleshooting-logging-action-log-list",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "w": 1599,
    "h": 1000
  },
  "annotations": []
}

## Просмотр конкретного действия

.Чтобы просмотреть конкретное действие:
1. Перейдите в меню:system[журнал действий],
1. при необходимости введите параметры фильтрации и нажмите на кнопку btn:[поиск],
1. нажмите на действие, которое хотите просмотреть.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "troubleshooting/logging/action-log-inspect.png",
    "alias": "troubleshooting-logging-action-log-inspect",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "w": 1599,
    "h": 1000
  },
  "focus": {
    "x": 563,
    "y": 444,
    "w": 1110,
    "h": 425,
    "padding": 5
  },
  "annotations": []
}

## Серьёзность действий

!!! note
    *DevNote* обратите внимание на уровни серьёзности
    
    * emergency: 'Emergency'
    * alert: 'Alert'
    * critical: 'Critical'
    * error: 'Error'
    * warning: 'Warning'
    * notice: 'Notice'
    * info: 'Info'
    * debug: 'Debug'
