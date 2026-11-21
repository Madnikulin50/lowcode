# Управление подключениями

Пользовательский интерфейс управления подключениями находится в веб-приложении LowCoooode Admin в разделе menu:system[connections].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/management/interface.png",
    "alias": "dal-connections-management-interface",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 20,
    "y": 573,
    "h": 13,
    "w": 172
  }]
}

## Список подключений

.Чтобы вывести список текущих подключений:
1. перейдите в menu:system[connections],
1. при необходимости включите удалённые подключения,
1. список обновится автоматически.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/management/interface.png",
    "alias": "dal-connections-management-interface-filtering",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "view": {
    "x": 352,
    "y": 0,
    "w": 1568,
    "h": 1080
  },
  "focus": {
    "x": 559,
    "y": 400,
    "w": 1108,
    "h": 680
  },
    "annotations": []
}

!!! important
    Список включает только внешние подключения; основное подключение доступно в верхней части списка.
    
    [annotation,role="data-zoomable"]
    ----
    {
      "image": {
        "rel": "dal-connections/management/interface.png",
        "alias": "dal-connections-management-interface-primary",
        "w": 1920,
        "h": 1080
      },
      "view": {
        "x": 352,
        "y": 0,
        "w": 1568,
        "h": 1080
      },
      "focus": {
        "x": 559,
        "y": 80,
        "w": 1108,
        "h": 295
      },
        "annotations": []
    }
    ----


## Создание подключений

.Чтобы создать новое подключение:
1. перейдите в menu:system[connections],
1. нажмите кнопку btn:[add connection] и заполните форму.

!!! note
    За подробностями о форме обратитесь к разделу [конфигурация подключения](modules/integrator-guide/pages/dal-connections/dal-connections/index.md#configuration).


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/management/interface.png",
    "alias": "dal-connections-management-interface-create",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 352,
    "y": 0,
    "w": 1568,
    "h": 1080
  },
  "focus": {
    "x": 559,
    "y": 400,
    "w": 1108,
    "h": 679
  },
  "annotations": [{
    "x": 576,
    "y": 473,
    "h": 41,
    "w": 164
  }]
}

## Редактирование подключений

.Чтобы обновить существующее подключение:
1. перейдите в menu:system[connections],
1. нажмите кнопку btn:[edit] и обновите подключение.

!!! note
    За подробностями о форме обратитесь к разделу [конфигурация подключения](modules/integrator-guide/pages/dal-connections/dal-connections/index.md#configuration).


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/management/interface.png",
    "alias": "dal-connections-management-interface-edit",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 352,
    "y": 0,
    "w": 1568,
    "h": 1080
  },
  "focus": {
    "x": 559,
    "y": 397,
    "w": 1108,
    "h": 682
  },
  "annotations": [{
    "x": 559,
    "y": 623,
    "h": 49,
    "w": 1108
  }]
}

## Удаление подключений

.Чтобы удалить существующее подключение:
1. перейдите в menu:system[connections],
1. нажмите кнопку btn:[edit],
1. нажмите кнопку btn:[delete] и подтвердите действие.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/management/delete.png",
    "alias": "dal-connections-management-delete",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 436,
    "y": 280,
    "w": 1400,
    "h": 801
  },
  "focus": {
    "x": 512,
    "y": 1015,
    "w": 1248,
    "h": 58
  },
  "annotations": [{
    "x": 527,
    "y": 1023,
    "w": 85,
    "h": 36
  }]
}
