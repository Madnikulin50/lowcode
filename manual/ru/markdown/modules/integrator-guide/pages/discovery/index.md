# LowCoooode Discovery

LowCoooode Discovery предоставляет мощный поисковый движок для взаимодействия с вашими данными.
LowCoooode Discovery определяет интуитивно понятный интерфейс для поиска и, в некоторых случаях, визуализации данных, таких как географические метаданные.

Обратитесь к [menu:Руководство DevOps[LowCoooode Discovery](modules/devops-guide/pages/discovery/index.md)] чтобы узнать, как настроить системы.

Чтобы получить доступ к LowCoooode Discovery, перейдите в ваш экземпляр LowCoooode и нажмите на приложение Discovery.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "app-selector-discovery.png",
    "alias": "discovery-app-selector",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 1095,
    "y": 477,
    "w": 278,
    "h": 228
  },
  "annotations": []
}

В приложении LowCoooode Discovery вы видите интуитивно понятный пользовательский интерфейс, где **общие фильтры** доступны в левой части экрана, **строка поиска** — в верхней части, а **результаты поиска** — сразу под строкой поиска.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/search-results.png",
    "alias": "discovery-search-results",
    "w": 1920,
    "h": 1080
  },
  "annotations": []
}

## Расширенные запросы

Чтобы найти то, что вам нужно, примените общий фильтр.
Он находится в левой части экрана.
Нажмите на флажок рядом с фильтром, чтобы применить его к результатам поиска.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/search-results.png",
    "alias": "discovery-search-filtering",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 0,
    "y": 64,
    "w": 312,
    "h": 320
  }
}

Вы можете ввести поисковый запрос в поле ввода в верхней части экрана, чтобы отфильтровать оставшиеся результаты поиска.
Поиск старается найти наилучшие результаты на основе доступных данных и введённого вами поискового термина.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/search-results.png",
    "alias": "discovery-search-querying",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 325,
    "y": 0,
    "w": 1600,
    "h": 1080
  }
}

### Визуализация географического местоположения

Если результат поиска определяет географическое местоположение, его можно визуализировать на карте.
Чтобы визуализировать результаты поиска на карте, нажмите на **{ICON*DISCOVERY*MAP} Показать карту** в правом верхнем углу.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/search-results.png",
    "alias": "discovery-search-map",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 1677,
    "y": 104,
    "w": 133,
    "h": 45
  },
  "annotations": []
}

Результаты поиска, которые можно отобразить на карте, представлены в виде маркеров на карте; результаты поиска, которые невозможно отобразить, исключаются с карты, но остаются в результатах поиска.
Нажмите на маркер, чтобы выделить соответствующий результат поиска в списке результатов, или наведите курсор на результат, чтобы увидеть местоположение на карте.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/search-results-mapped.png",
    "alias": "discovery-search-results-mapped",
    "w": 1920,
    "h": 1080
  },
  "view": {}
}

<a id="index-configuration"></a>
## Конфигурация индексации

<a id="data-visibility"></a>
### Видимость данных

**Видимость данных** определяет **что индексируется** и контролируется стандартным **механизмом контроля доступа**.
Discovery индексирует всё, к чему у него в данный момент есть доступ.

Чтобы управлять видимостью данных, настройте **правила контроля доступа** для роли, которую вы определили в [конфигурации Discovery](modules/devops-guide/pages/examples/deploy-online/multi-discovery-pgsql.md).

Обратитесь к [примеру](modules/integrator-guide/pages/discovery/examples/limit-namespace.md), где индексатор ограничен показом записей только из указанного пространства имён.

### Настройки Discovery

!!! important
    В настоящее время настройки Discovery применяются к доступу индексатора к определённым полям.
    Для управления целыми строками используйте <<data-visibility,видимость данных>>.


Нажмите на кнопку **{ICON*DISCOVERY*ICON} Настройки Discovery** на странице конфигурации модуля, чтобы указать, какие поля индексировать.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/module-discovery.png",
    "alias": "discovery-module-discovery",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 414,
    "y": 60,
    "w": 1409,
    "h": 810
  },
  "focus": {
    "x": 564,
    "y": 84,
    "w": 221,
    "h": 64
  },
  "annotations": []
}
Выбранные поля будут проиндексированы и доступны для поиска.


!!! important
    Если поля не выбраны, индексатор проиндексирует все поля, к которым у него есть доступ, в соответствии с Видимостью данных.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "discovery/discovery-settings.png",
    "alias": "discovery-discovery-settings",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 450,
    "y": 60,
    "w": 950,
    "h": 600
  },
  "focus": {
    "x": 544,
    "y": 18,
    "w": 816,
    "h": 888
  },
  "annotations": []
}
