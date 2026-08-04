# Аннотации к изображениям

По возможности аннотации к изображениям следует делать **программно**, чтобы уменьшить объём усилий, необходимых для приведения изображений к единому виду, соответствующему нашим фирменным правилам.

!!! important
    Эти инструкции следует использовать по возможности, но могут быть и исключения.
    В случаях, когда это невозможно, вы можете использовать внешние программы, если вы следуете руководству по стилю и фирменным правилам.


!!! note
    *DevNote* добавить ссылки на руководство по стилю и фирменные правила.


## Создание скриншотов

Скриншоты можно делать в любом удобном вам браузере.
Мы предлагаем использовать Firefox или браузер на основе Chromium.

<a id="setup"></a>
### Настройка

**Определите новое устройство** для создания скриншотов.
Мы можем определить устройство с определёнными параметрами экрана, чтобы скриншоты выглядели как можно более единообразно.

.В инструментах разработчика:
1. нажмите на кнопку btn:[переключить панель устройств],
1. нажмите на выпадающий список с перечнем устройств,
1. нажмите на пункт «Изменить», btn:[добавить собственное устройство].

[cols="1s,5a"]
|===
| Имя
|
Имя не имеет значения; я (автор) использую «LowCoooode screenshots».

| Ширина
|
Ширина **должна быть** `1920` пикселей.
Любое изменение размера должно выполняться плагином аннотаций.

| Высота
|
Высота **должна быть** `1080` пикселей.
Любое изменение размера должно выполняться плагином аннотаций.

| Плотность пикселей устройства
|
Плотность пикселей устройства должна быть установлена на `1`.

| Тип пользовательского агента
|
Тип пользовательского агента должен быть установлен в значение «desktop».
Пользователи **Firefox** могут использовать следующее значение.

```
```
Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0

|===

**Добавьте расширение-линейку**, чтобы упростить определение положения элементов на странице в пикселях.
Я (автор) использую [Better Ruler](https://chromewebstore.google.com/detail/better-ruler/ilcnadaaninblgbekoaihdhoiecaflie), так как, на момент написания, это лучшее и самое простое расширение из имеющихся.

Если вы не хотите или не можете найти подходящее расширение для вашего браузера, вы можете определять положение вручную, если вы следуете фирменным правилам и действуете последовательно.

### Скриншоты

.Чтобы сделать скриншот:
1. Подготовьте состояние страницы, содержащее всю информацию, которую вы хотите запечатлеть (например, откройте модальные окна и заполните поля ввода).
1. Откройте инструменты разработчика и нажмите на кнопку btn:[переключить панель устройств].
1. Нажмите на выпадающий список с перечнем устройств и выберите устройство, подготовленное вами в [setup,разделе настройки](#setup,разделе настройки).

.Скриншот показывает расположение кнопки создания скриншота в Firefox.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/screenshot-firefox.png",
    "w": 977,
    "h": 557
  },
  "view": {},
  "annotations": [{
    "kind": "box-note",
    "x": 335,
    "y": 25,
    "w": 0,
    "h": 0
  }]
}

.Скриншот показывает расположение кнопки создания скриншота в браузере на основе Chromium (Brave).
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/screenshot-brave.png",
    "w": 711,
    "h": 723
  },
  "view": {},
  "annotations": [{
    "kind": "box-note",
    "x": 215,
    "y": 185,
    "w": 220,
    "h": 20
  }]
}

## Аннотирование

.Пример аннотации:
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "workflow/accessing-workflows-1.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 818,
    "y": 516,
    "w": 288,
    "h": 250
  }]
}
```

Параметр `image` описывает базовое изображение, которое вы хотите аннотировать.
Изображение **должно** определять размер исходного изображения (в будущем это может быть улучшено).
Параметр `rel` подчиняется тем же правилам, что и [обычный путь к изображению](https://docs.antora.org/antora/2.0/asciidoc/insert-image/).

!!! caution
    Если вы планируете использовать одно и то же изображение несколько раз с разными аннотациями, убедитесь, что вы определили параметр `alias` для `image`.
    
    .В качестве примера:
    ```
    [annotation,role="data-zoomable"]
    ----
    {
      "image": {
        "rel": "documentation/example-blank.png",
        "alias": "example-blank-box",
        "w": 516,
        "h": 353
      },
      "view": {},
      "annotations": [{
        "kind": "box",
        "x": 50,
        "y": 50,
        "w": 403,
        "h": 240
      }]
    }
    ----
    
    [annotation,role="data-zoomable"]
    ----
    {
      "image": {
        "rel": "documentation/example-blank.png",
        "alias": "example-blank-box-danger",
        "w": 516,
        "h": 353
      },
      "view": {},
      "annotations": [{
        "kind": "box-danger",
        "x": 50,
        "y": 50,
        "w": 403,
        "h": 240
      }]
    }
    ----
    ```


Параметр `view` определяет, какая часть исходного изображения должна быть показана.
Это позволяет вам смещать область просмотра и обрезать изображение.

Параметр `annotations` позволяет вам определить массив предопределённых объектов аннотаций.

### Аннотации-прямоугольники

!!! note
    *DevNote* будет добавлено по мере изменения требований и установления фирменных правил.


.Список предопределённых аннотаций-прямоугольников:
[cols="2s,5a,5a"]
|===
| [#annotation-box-note]#[annotation-box-note,Заметка](#annotation-box-note,Заметка)#
|
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_note",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-note",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}

```
|
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_note",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-note",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}


| [#annotation-box-success]#[annotation-box-success,Успех](#annotation-box-success,Успех)#
|
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_success",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-success",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-success",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-success",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-success",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}

```
|
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_success",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-success",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-success",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-success",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-success",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}


| [#annotation-box-danger]#[annotation-box-danger,Опасность](#annotation-box-danger,Опасность)#
|
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_danger",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-danger",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-danger",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-danger",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-danger",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}

```
|
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_danger",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-danger",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-danger",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-danger",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-danger",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}


|===

### Изменение отступов аннотаций

Используйте свойство `padding`, чтобы изменить отступ вашей аннотации.
Доступные значения: `xs`, `sm`, `md` и `lg`; по умолчанию используется значение `md`.

```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_padding",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-note",
    "padding": "xs",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "padding": "sm",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "padding": "md",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "padding": "lg",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}

```
|
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_padding",
    "w": 1000,
    "h": 1000
  },
  "view": {},
  "annotations": [{
    "kind": "box-note",
    "padding": "xs",
    "x": 111,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "padding": "sm",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "padding": "md",
    "x": 111,
    "y": 703,
    "w": 279,
    "h": 95
  }, {
    "kind": "box-note",
    "padding": "lg",
    "x": 611,
    "y": 703,
    "w": 279,
    "h": 95
  }]
}


### Обрезка изображений

Чтобы обрезать изображение, настройте свойство `view`.

.Свойство `view` имеет следующие параметры:
[cols="2s,5a"]
|===
| x
| Горизонтальное смещение от верхнего левого угла исходного изображения.

| y
| Вертикальное смещение от верхнего левого угла исходного изображения.

| w
| Ширина результирующего изображения (область, которую вы хотите показать).

| h
| Высота результирующего изображения (область, которую вы хотите показать).
|===

В качестве примера, давайте удалим шапку и навигацию со следующего изображения.

![role="data-zoomable"](documentation/example-crop-base.png)

.Следующая конфигурация выполняет нужную обрезку:
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/example-crop-base.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 545,
    "y": 100,
    "w": 900,
    "h": 970
  },
  "annotations": []
}
```

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/example-crop-base.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 545,
    "y": 100,
    "w": 900,
    "h": 970
  },
  "annotations": []
}

### Обрезка и аннотирование изображений

В качестве примера, давайте оставим только второй объект следующего изображения (верхний правый квадрат).

![width=500px](documentation/annotation-canvas.png)

.Следующая конфигурация выполняет нужную обрезку:
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_cropped-note",
    "w": 1000,
    "h": 1000
  },
  "view": {
    "x": 500,
    "y": 0,
    "w": 500,
    "h": 500
  },
  "annotations": [{
    "kind": "box-note",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }]
}

```

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-canvas.png",
    "alias": "annotation-canvas_cropped-note",
    "w": 1000,
    "h": 1000
  },
  "view": {
    "x": 500,
    "y": 0,
    "w": 500,
    "h": 500
  },
  "annotations": [{
    "kind": "box-note",
    "x": 611,
    "y": 203,
    "w": 279,
    "h": 95
  }]
}


### Фокусировка изображений

Чтобы задать фокус на изображении, настройте свойство `focus`.

.Свойство `focus` имеет следующие параметры:
[cols="2s,5a"]
|===
| x
| Горизонтальное смещение от верхнего левого угла исходного изображения.

| y
| Вертикальное смещение от верхнего левого угла исходного изображения.

| w
| Ширина области фокуса (не размытой области).

| h
| Высота области фокуса (не размытой области).

| padding
| Не размытая буферная область вокруг области фокуса.
|===

В качестве примера, давайте сфокусируемся на приложении Admin Area в переключателе приложений.

![role="data-zoomable"](documentation/example-focus.png)

.Следующая конфигурация выполняет нужную фокусировку:
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/example-focus.png",
    "alias": "focus-base",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "w": 267,
    "h": 253,
    "x": 542,
    "y": 515,
    "padding": 10
  },
  "annotations": []
}

```

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/example-focus.png",
    "alias": "focus-base",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "w": 267,
    "h": 253,
    "x": 542,
    "y": 515,
    "padding": 10
  },
  "annotations": []
}


### Фокусировка и аннотирование изображений

В качестве примера, давайте сфокусируемся на приложении Admin Area в переключателе приложений с аннотацией.

![role="data-zoomable"](documentation/example-focus.png)

.Следующая конфигурация выполняет нужную фокусировку:
```
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/example-focus.png",
    "alias": "focus-annotated",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "w": 267,
    "h": 253,
    "x": 542,
    "y": 515,
    "padding": 10
  },
  "annotations": [{
    "kind": "box-note",
    "x": 542,
    "y": 515,
    "w": 267,
    "h": 253
  }]
}

```

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/example-focus.png",
    "alias": "focus-annotated",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "w": 267,
    "h": 253,
    "x": 542,
    "y": 515,
    "padding": 10
  },
  "annotations": [{
    "kind": "box-note",
    "x": 542,
    "y": 515,
    "w": 267,
    "h": 253
  }]
}


## Советы и хитрости

### Определение положения элементов

Вы можете использовать расширение-линейку, настроенное в [setup,разделе настройки](#setup,разделе настройки), чтобы упростить определение положения элементов.

.Скриншот показывает использование расширения-линейки для определения положения элементов на экране.
![role="data-zoomable"](documentation/ruler-extension-example.png)

### Упрощение размещения аннотаций

Размещение аннотаций может быть утомительным процессом, если вам нужно выполнить финальные штрихи, такие как выравнивание.

Аннотации отображаются в виде SVG-изображений, что позволяет вам использовать либо специальные программы, такие как [Inkscape](https://inkscape.org/), либо консоль разработчика браузера для корректировки положения аннотаций.

Чтобы использовать **специальную программу**, просто скачайте аннотированное SVG-изображение и откройте его в программе.
Затем вы можете скорректировать параметры конфигурации аннотации на основе правок в вашей программе редактирования SVG.

Чтобы использовать **консоль разработчика**, нажмите правой кнопкой мыши на аннотированное изображение и выберите пункт **открыть изображение в новой вкладке**.
Изучите SVG-изображение и скорректируйте аннотации по своему усмотрению.
Используйте скорректированные параметры для обновления конфигурации аннотации.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "documentation/annotation-edit-browser.png",
    "w": 1920,
    "h": 987
  },
  "view": {
    "h": 500
  },
  "annotations": []
}
