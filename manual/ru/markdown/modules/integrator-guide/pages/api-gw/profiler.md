# Route Profiler

**Integration gateway route profiler** упрощает процесс разработки и устранения неполадок, позволяя детально просматривать отдельные запросы.
Помимо разработки, профайлер можно использовать для тестирования производительности системы, чтобы помочь определить потенциальные узкие места или получить общее представление о возможностях системы.

Профайлер спроектирован таким образом, что его можно использовать как инструмент разработки и отладки или как панель для просмотра входящих запросов.

!!! important
    Результаты профилирования и агрегированная статистика не сохраняются в хранилище, поэтому любой перезапуск сервера приведет к перезапуску самого профайлера и, как следствие, к удалению ранее собранных запросов шлюза.


!!! tip
    Вы можете использовать эти тесты, чтобы определить, потребуется ли вам применять какую-либо стратегию масштабирования системы.


## Включение профайлера

.Есть два способа использования профайлера Integration Gateway:
- глобально включенный ([`APIGW*PROFILER*GLOBAL`](modules/devops-guide/pages/references/configuration/server.md#*apigw*profiler_global) переменная `.env`)
- включенный конкретно для маршрута (с помощью profiler prefilter)

!!! important
    Сам профайлер включен по умолчанию, но глобальная опция отключена, поэтому на странице профайлера не будет запросов, кроме тех, которые уже существуют в Integration Gateway и имеют добавленный и включенный profiler prefilter.


<a id="profiler-add"></a>
## Настройка профайлера маршрута

Чтобы профайлер включил маршрут Integration Gateway, вы должны явно включить его, настроив **profiler** prefilter.
Чтобы добавить profiler prefilter к эндпоинту Integration Gateway, нажмите на кнопку btn:[add filter] и выберите "profiler" из выпадающего списка.

!!! caution
    Prefilters применяются в том же порядке, в котором они определены, то есть профайлер применяется после всех других prefilters, определенных до него.
    Сортировка фильтров (с помощью перетаскивания) отсутствует, поэтому пока единственный способ — удалить другие фильтры и создать их в нужном порядке.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/prefilter-select.png",
    "alias": "api-gw-profiler-prefilter-select",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 289,
    "y": 1,
    "w": 1630,
    "h": 1078
  },
  "annotations": [{
    "kind": "box-note",
    "padding": "sm",
    "x": 621,
    "y": 754,
    "w": 94,
    "h": 27
  }, {
    "kind": "box-note",
    "padding": "sm",
    "x": 621,
    "y": 851,
    "w": 161,
    "h": 11
  }]
}

В модальном окне конфигурации убедитесь, что опция enabled отмечена, и нажмите кнопку btn:[save & close].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/prefilter-cofigure.png",
    "alias": "api-gw-profiler-prefilter-cofigure",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 600,
    "y": 26,
    "w": 720,
    "h": 160
  }
}

Когда profiler prefilter добавлен, вы увидите новый **список запросов** в нижней части страницы.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/preflter-added.png",
    "alias": "api-gw-profiler-preflter-added",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 598,
    "y": 578,
    "w": 1000,
    "h": 490
  },
  "annotations": [{
    "kind": "box",
    "x": 598,
    "y": 885,
    "w": 1000,
    "h": 182
  }]
}

<a id="profile-specific"></a>
## Профилирование конкретных маршрутов

Чтобы профилировать конкретный маршрут, откройте его и прокрутите страницу вниз.
Список запросов показывает все HTTP-запросы для данного маршрута Integration Gateway, которые произошли **после** добавления profiler prefilter.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/route-profiler.png",
    "alias": "api-gw-profiler-route-profiler",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 598,
    "y": 418,
    "w": 1000,
    "h": 648
  }
}

Чтобы просмотреть детали конкретного запроса, нажмите на иконку btn:[edit] в конце записи запроса.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/route-profiler.png",
    "alias": "api-gw-profiler-route-profiler-details-open",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 599,
    "y": 419,
    "w": 997,
    "h": 647
  },
  "annotations": [{
    "kind": "box-note",
    "x": 1572,
    "y": 583,
    "w": 13,
    "h": 13
  }]
}

В самом верху страницы вы видите **общую информацию**, такую как ID запроса, HTTP-метод, код статуса ответа и длительность.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/request-details-top.png",
    "alias": "api-gw-profiler-request-details-top-general",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 599,
    "y": 73,
    "w": 998,
    "h": 448
  }
}

Далее идет список **заголовков запроса**, предоставленных с HTTP-запросом.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/request-details-bottom.png",
    "alias": "api-gw-profiler-request-details-bottom-headers",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 599,
    "y": 410,
    "w": 999,
    "h": 230
  }
}

И наконец, сырая **полезная нагрузка запроса**, предоставленная с HTTP-запросом.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/request-details-bottom.png",
    "alias": "api-gw-profiler-request-details-bottom-body",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 598,
    "y": 653,
    "w": 999,
    "h": 414
  }
}

## Профилирование системы

Чтобы увидеть общую производительность системы, мы предлагаем общесистемное представление профайлера.
Перейдите в menu:system[integration gateway] и нажмите на кнопку btn:[profiler] в правом верхнем углу.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/route-list.png",
    "alias": "api-gw-profiler-route-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
   "focus": {
    "x": 605,
    "y": 293,
    "w": 999,
    "h": 774
  },
  "annotations": [{
    "kind": "box-note",
    "x": 735,
    "y": 361,
    "w": 79,
    "h": 37
  }]
}

Появляется новый экран, показывающий все зарегистрированные маршруты Integration Gateway (отображаются только маршруты с **profiler prefilter**) вместе с их статистикой, такой как среднее время запросов и размер полезной нагрузки.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/profiler-general.png",
    "alias": "api-gw-profiler-profiler-general",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 289,
    "y": 1,
    "w": 1630,
    "h": 520
  }
}

Чтобы просмотреть запросы для конкретного маршрута, нажмите на иконку btn:[edit] рядом с маршрутом Integration Gateway или следуйте инструкциям, [profile-specific,приведенным выше](#profile-specific,приведенным выше).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/profiler-general.png",
    "alias": "api-gw-profiler-profiler-general-see-specific",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 289,
    "y": 1,
    "w": 1630,
    "h": 520
  },
    "annotations": [{
    "x": 1578,
    "y": 274,
    "w": 13,
    "h": 13
  }]
}

Когда вы просматриваете конкретный маршрут Integration Gateway, отображается список всех зарегистрированных запросов.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "api-gw/profiler/profiler-general-route.png",
    "alias": "api-gw-profiler-profiler-general-route",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 320,
    "y": 0,
    "h": 830,
    "w": 1600
  }
}
