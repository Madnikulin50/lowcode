# Examples

:leveloffset: +1


# Индексация только указанного пространства имён
:page-noindex: true

[Видимость данных](modules/integrator-guide/pages/discovery/examples/discovery/index.md#data-visibility) контролируется стандартным механизмом контроля доступа.
Discovery индексирует всё, к чему у него в данный момент есть доступ.

Чтобы управлять видимостью данных, настройте правила контроля доступа для роли, назначенной [пользователю-индексатору Discovery](modules/integrator-guide/pages/discovery/examples/discovery/index.md#indexer-user).

В этом примере мы настраиваем LowCoooode Discovery на индексацию только записей из пространства имён Discovery.

Перейдите на страницу управления пространствами имён и нажмите на кнопку **Разрешения**, чтобы открыть глобальные правила контроля доступа к пространствам имён.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/examples/limit-namespace/namespace-manage.png",
    "alias": "discovery-examples-limit-namespace-namespace-manage",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Выберите роль **Authenticated** и установите правило `Читать любое пространство имён` в значение `Запретить`.
В качестве альтернативы вы можете установить то же правило на `Запретить` для роли индексатора Discovery.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/examples/limit-namespace/namespace-rbac-global.png",
    "alias": "discovery-examples-limit-namespace-namespace-rbac-global",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 365,
    "y": 21,
    "w": 1258,
    "h": 500
  },
  "annotations": []
}

Затем нажмите на кнопку **Редактировать** для пространства имён, к которому вы хотите разрешить доступ, и нажмите на кнопку **Разрешения**, чтобы открыть правила контроля доступа для данного пространства имён.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/examples/limit-namespace/namespace-edit.png",
    "alias": "discovery-examples-limit-namespace-namespace-edit",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 378,
    "y": 62,
    "w": 1186,
    "h": 755
  },
  "annotations": []
}

Выберите роль **Discoverer** (роль, назначенная пользователю-индексатору Discovery) и установите правило `Читать пространство имён` в значение `Разрешить`.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/examples/limit-namespace/namespace-rbac.png",
    "alias": "discovery-examples-limit-namespace-namespace-rbac",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 365,
    "y": 21,
    "w": 1258,
    "h": 500
  },
  "annotations": []
}

!!! important
    Индексатор Discovery должен завершить переиндексацию, чтобы изменения вступили в силу.


После того как индексатор Discovery переиндексирует ваш экземпляр, результаты поиска будут ограничены тем пространством имён Discovery, доступ к которому разрешён.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/examples/limit-namespace/search-results-limited.png",
    "alias": "discovery-examples-limit-namespace-search-results-limited",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}


:leveloffset: -1
