# Руководство по стилю

Рабочие процессы основаны на стандарте BPMN 2.0 и наследуют его базовые правила.
Мы определяем несколько дополнительных правил, помогающих сохранять согласованность рабочих процессов.

## Следуйте принципу KISS

Принцип «Keep It Simple, Stupid» (не усложняй) означает, что ваши рабочие процессы не должны быть излишне сложными.
Излишнее усложнение обычно снижает наглядность и добавляет дополнительные точки отказа.

Например, вместо использования шлюза для установки переменной `Boolean` в нужное значение напишите выражение, вычисляющее желаемое значение.

.На скриншоте показана излишне усложнённая версия.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/simple-gateway.png",
    "alias": "automation-workflows-styleguide-simple-gateway.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 164,
    "y": 286,
    "w": 691,
    "h": 430
  },
  "annotations": []
}

.На скриншоте показана упрощённая версия.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/simple-expression.png",
    "alias": "automation-workflows-styleguide-simple-expression.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 1030,
    "y": 1,
    "w": 889,
    "h": 600
  },

  "annotations": []
}

!!! caution
    Большинство операций можно преобразовать в выражения, но ценой наглядности.
    Если упрощение снижает наглядность, от него следует отказаться.


!!! tip
    Более крупные сложные операции можно заменить [скриптами автоматизации](modules/integrator-guide/pages/automation/workflows/automation/workflows/automation-scripts.md).


## Избегайте пересечения соединителей

При работе с крупными рабочими процессами пересекающиеся соединители могут вводить читателя в заблуждение и сбивать с толку.
Шаги рабочего процесса определяют множество опорных точек, которые помогают этого избежать.

.Пример пересекающихся соединителей:
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/connectors-nok.png",
    "alias": "automation-workflows-styleguide-connectors-nok.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 78,
    "y": 156,
    "w": 1295,
    "h": 389
  },
  "annotations": []
}

.Пример исправленных, непересекающихся соединителей:
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/connectors-ok.png",
    "alias": "automation-workflows-styleguide-connectors-ok.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 78,
    "y": 155,
    "w": 1295,
    "h": 303
  },
  "annotations": []
}

!!! tip
    Соединитель определяет ряд опорных точек, которые можно использовать для регулировки линии соединителя.
    Опорные точки становятся доступными, когда вы нажимаете на соединитель.
    
    [annotation,role="data-zoomable"]
    ----
    {
      "image": {
        "rel": "automation/workflows/styleguide/connector-anchors.png",
        "alias": "automation-workflows-styleguide-connector-anchors.png",
        "w": 1920,
        "h": 1080
      },
      "view": {
        "x": 188,
        "y": 290,
        "w": 996,
        "h": 373
      },
      "annotations": []
    }
    ----
    


## Подписи шагов

Все шаги рабочего процесса должны иметь короткую подпись, отражающую суть шага.
Например, выражение, вычисляющее комиссионные с продаж, лучше всего подписать как «Расчёт комиссионных с продаж».

Подпись **триггера** должна указывать ресурс и тип триггера.
Это позволяет читателю понять, когда срабатывает триггер, не читая конфигурацию.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/label-trigger.png",
    "alias": "automation-workflows-styleguide-label-trigger.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 325,
    "y": 335,
    "w": 629,
    "h": 360
  },
  "annotations": []
}

Подпись **шлюза** должна указывать вопрос, определяющий выход.
Подписи выходных путей должны указывать ответ на вопрос соответствующего шлюза.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/label-gateway.png",
    "alias": "automation-workflows-styleguide-label-gateway.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 192,
    "y": 228,
    "w": 725,
    "h": 517
  },
  "annotations": []
}

Подпись **функции** должна содержать глагол и объект действия функции.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/label-function.png",
    "alias": "automation-workflows-styleguide-label-function.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 193,
    "y": 280,
    "w": 361,
    "h": 206
  },
  "annotations": []
}

## Согласованные макеты

Первоначальная интеграция обычно состоит из нескольких рабочих процессов, которые в будущем можно расширять и дополнять.
Согласованный макет повышает согласованность и целостность.

.Вот несколько советов:
- поток должен идти сверху слева вниз направо,
- группируйте связанные шаги,
- выравнивайте шаги по сетке,
- сохраняйте симметричность макета, так как её обычно проще понять.

## Используйте дорожки (swimlanes) для организации рабочих процессов

Для организации рабочих процессов вы можете использовать визуальный элемент **дорожка** {ICON*WORKFLOW*SWIMLANE}.

В небольших рабочих процессах, выполняющих простые одиночные операции (например, уведомление администратора о регистрации клиента или журналирование изменений), дорожки не имеют особого смысла и могут даже снизить наглядность.

Однако для крупных проектов дорожки могут стать упорядочивающим дополнением, поскольку помогают визуально группировать мелкие шаги в более крупные операции и таким образом позволяют понять, что делает конкретная часть рабочего процесса.

.На скриншоте показан рефакторинг рабочего процесса с дорожками для группировки операций.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "automation/workflows/styleguide/swimlane.png",
    "alias": "automation-workflows-styleguide-swimlane.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 100,
    "y": 86,
    "w": 1026,
    "h": 945
  },
  "annotations": []
}
