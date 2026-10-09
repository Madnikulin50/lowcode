# Справочник блоков страницы

!!! note
    Обратитесь к menu:accessing LowCoooode[query language] [за подробностями о фильтрации мультизначных полей](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md#query-mv).


<a id="page-block-automation"></a>
## Автоматизация

Блок страницы автоматизации позволяет отображать [автоматизацию с ручными триггерами](modules/integrator-guide/pages/compose-configuration/automation/index.md#automation-application-general).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-automation.png",
    "alias": "compose-configuration-page-block-reference-config-automation",
    "w": 1920,
    "h": 1080
  },
   "focus": {
    "x": 334,
    "y": 32,
    "w": 1251,
    "h": 976
  },
  "view": {},
  "annotations": []
}

Блок страницы автоматизации позволяет использовать как [скрипты автоматизации](modules/integrator-guide/pages/compose-configuration/automation/automation-scripts/index.md), так и [рабочие процессы](modules/integrator-guide/pages/compose-configuration/automation/workflows/index.md).

.Если вы не видите свои скрипты автоматизации, убедитесь:
- что не было ошибок компиляции,
- что вы обновили страницу после добавления скриптов,
- что триггер определяет ограничение `.uiProp('app', 'compose')`.

.Если вы не видите свои рабочие процессы, убедитесь:
- что в рабочем процессе нет ошибок,
- что рабочий процесс имеет триггер `onManual`,
- что рабочий процесс включён.

<a id="page-block-calendar"></a>
## Календарь

**Блок страницы календаря** позволяет отображать события из разных источников событий в календаре.

![...](compose-configuration/page-block-reference/config-calendar-base.png)
![...](compose-configuration/page-block-reference/config-calendar-es.png)

.Поддерживаемые представления календаря:
- представление по месяцам
- представление по неделям
- представление по дням
- представление по повестке

Блоки страницы календаря поддерживают два источника событий: **напоминания** и **записи**.

Источник событий **напоминания** позволяет отображать ваши напоминания как события календаря.

Источник событий **записи** позволяет отображать запись как событие календаря.
Это позволяет вам иметь модуль для хранения задач, которые затем можно отображать в календаре.

!!! tip
    При использовании записей вы можете определить предустановленный фильтр, который позволяет отображать только определённые записи.
    Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.


!!! tip
    Предустановленный фильтр блока страницы календаря поддерживает интерполяцию значений.


<a id="page-block-chart"></a>
## Диаграмма

**Блок страницы диаграммы** позволяет отображать предварительно настроенную диаграмму на вашей странице.
Обратитесь к [Конфигурации Low Code](modules/integrator-guide/pages/compose-configuration/compose-configuration/index.md#charts) за более подробной информацией.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-chart.png",
    "alias": "compose-configuration-page-block-reference-config-chart",
    "w": 1920,
    "h": 1080
  },
  "view": {},
   "focus": {
    "x": 333,
    "y": 32,
    "w": 1252,
    "h": 306
  },
    "annotations": []
}

<a id="page-block-content"></a>
## Содержимое

**Блок страницы содержимого** позволяет отображать статическое содержимое на странице.
Это может быть важное объявление, инструкция или любая другая информация.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-content.png",
    "alias": "compose-configuration-page-block-reference-config-content",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1252,
    "h": 474
  },
  "annotations": []
}

<a id="page-block-file"></a>
## Файл

**Блок страницы файла** позволяет загружать статические файлы, такие как пользовательское соглашение, график работы или брошюру, на страницу.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-file.png",
    "alias": "compose-configuration-page-block-reference-config-file",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1252,
    "h": 435
  },
  "annotations": []
}

Вложение является статическим и не изменяется в зависимости от текущего контекста.

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#page-block-file-vm]#[page-block-file-vm,View mode](#page-block-file-vm,View mode)#
|
Определяет, как файлы представлены при просмотре.

!!! note
    *DevNote* перечислить и описать доступные опции.

|===

<a id="page-block-iframe"></a>
## IFrame

**Блок страницы iframe** позволяет встраивать внешний сайт в ваше приложение.

!!! important
    Учитывайте любые *ограничения встраивания iframe*.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-iframe.png",
    "alias": "compose-configuration-page-block-reference-config-iframe",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1252,
    "h": 469
  },
  "annotations": []
}

<a id="page-block-metric"></a>
## Метрика

**Блок страницы метрики** позволяет отображать вычисленное числовое значение на основе данных в вашем пространстве имён.
Обычно это показатель, важный для вашего процесса.
Это может быть такая информация, как общая сумма финансов в вашей воронке продаж или текущее количество открытых контрагентов.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-metric.png",
    "alias": "compose-configuration-page-block-reference-config-metric",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1251,
    "h": 976
  },
  "annotations": []
}

.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-metric-label]#[page-block-metric-label,Label](#page-block-metric-label,Label)#
|
Определяет, что показывает метрика.
Подпись не накладывает никаких ограничений на значение и не является обязательной.

| [#page-block-metric-d-module]#[page-block-metric-d-module,Dimension module](#page-block-metric-d-module,Dimension module)#
|
Определяет, какие записи будут использоваться при вычислении метрики.

| [#page-block-metric-d-filter]#[page-block-metric-d-filter,Dimension filter](#page-block-metric-d-filter,Dimension filter)#
|
Определяет, как фильтровать записи при вычислении метрики.
Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.

!!! tip
    Фильтр поддерживает интерполяцию значений.


| [#page-block-metric-m-field]#[page-block-metric-m-field,Metric field](#page-block-metric-m-field,Metric field)#
|
Определяет, какое числовое поле должно использоваться при вычислении метрики.
Каждый модуль имеет системное поле `count`, которое представляет общее количество записей, соответствующих указанному фильтру.

| [#page-block-metric-m-agg]#[page-block-metric-m-agg,Aggregation operation](#page-block-metric-m-agg,Aggregation operation)#
|
Определяет, как агрегируются значения при вычислении метрики.
Вы можете выбрать одну из функций `count`, `sum`, `max`, `min` или `avg`.

| [#page-block-metric-m-trans]#[page-block-metric-m-trans,Value transformation](#page-block-metric-m-trans,Value transformation)#
|
Позволяет определить, как результат метрики преобразуется перед отображением в блоке страницы.

Выражения преобразования значений записываются в виде простых JavaScript-выражений, которые возвращают одно число на основе двух переменных.

!!! note
    *DevNote* добавить более подробную информацию об этом.


| [#page-block-metric-m-fmt]#[page-block-metric-m-fmt,Format](#page-block-metric-m-fmt,Format)#
|
Определяет строку формата, используемую при отображении метрики.
Поле поддерживает все опции форматирования [numeral.js](https://numeraljs.com/#format).

Формат применяется до префикса и суффикса.

| [#page-block-metric-m-pfx]#[page-block-metric-m-pfx,Prefix](#page-block-metric-m-pfx,Prefix)#
|
Добавляет указанный префикс к результирующей метрике перед её отображением в блоке страницы.
Например, префикс `$` и значение `1000` дадут подпись `$1000`.

| [#page-block-metric-m-sfx]#[page-block-metric-m-sfx,Suffix](#page-block-metric-m-sfx,Suffix)#
|
Добавляет указанный суффикс к результирующей метрике перед её отображением в блоке страницы.
Например, суффикс `USD/h` и значение `1000` дадут подпись `1000USD/h`.

| [#page-block-metric-m-style]#[page-block-metric-m-style,Style](#page-block-metric-m-style,Style)#
|
Позволяет определить визуальные аспекты отображения метрик.

|===

<a id="page-block-record-list"></a>
## Список записей

**Блоки страницы списка записей** позволяют отображать записи в виде таблицы.
Списки записей также предоставляют функции добавления, импорта и экспорта записей.

!!! important
    При использовании поля поиска в списке записей система включает только выбранные поля.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-record-list.png",
    "alias": "compose-configuration-page-block-reference-config-record-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1252,
    "h": 977
  },
  "annotations": []
}

!!! tip
    Вы можете запускать *явные скрипты автоматизации* в заголовке списка записей.
    Нажмите на вкладку "Automation", чтобы *выбрать скрипты автоматизации*.
    
    [annotation,role="data-zoomable"]
    ----
    {
      "image": {
        "rel": "compose-configuration/page-block-reference/config-record-list-automation.png",
        "alias": "compose-configuration-page-block-reference-config-record-list-automation",
        "w": 1920,
        "h": 1080
      },
      "view": {},
      "focus": {
        "x": 333,
        "y": 32,
        "w": 1253,
        "h": 978
      },
      "annotations": []
    }
    ----


.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-rl-module]#[page-block-rl-module,Module](#page-block-rl-module,Module)#
|
Определяет модуль, который список записей будет использовать при взаимодействии с записями.

| [#page-block-rl-module-fields]#[page-block-rl-module-fields,Module fields](#page-block-rl-module-fields,Module fields)#
|
Позволяет определить поля и их порядок при отображении таблицы.

| [#page-block-rl-inline]#[page-block-rl-inline,Allow inline record editing](#page-block-rl-inline,Allow inline record editing)#
|
Позволяет создавать, редактировать и удалять записи непосредственно из списка записей, когда содержащая страница записи находится в режиме редактирования.

!!! important
    Эта опция доступна только для *страниц записей*.


| [#page-block-rl-add-hide]#[page-block-rl-add-hide,Hide add record button](#page-block-rl-add-hide,Hide add record button)#
|
Скрывает кнопку btn:[+ Add record], что предотвращает добавление новых записей из пользовательского интерфейса.

!!! caution
    Это *не* предотвращает создание записей, если доступ осуществляется из другого пользовательского интерфейса или через автоматизацию.


| [#page-block-rl-prefilter]#[page-block-rl-prefilter,Pre-filter records](#page-block-rl-prefilter,Pre-filter records)#
|
Определяет предустановленный фильтр, используемый при поиске и отображении записей в списке записей.
Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.

| [#page-block-rl-search-hide]#[page-block-rl-search-hide,Hide search box](#page-block-rl-search-hide,Hide search box)#
|
Скрывает поле *поиска*, что предотвращает применение дополнительных фильтров к списку записей.

!!! caution
    <<page-block-rl-prefilter,Предустановленный фильтр>> применяется независимо от этой опции.


| [#page-block-rl-presort]#[page-block-rl-presort,Presort records](#page-block-rl-presort,Presort records)#
|
Определяет начальную сортировку, применяемую при отображении списка записей.
Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.

!!! note
    При применении пользовательской сортировки из таблицы предустановленная сортировка перезаписывается.


| [#page-block-rl-sort-hide]#[page-block-rl-sort-hide,Hide sorting](#page-block-rl-sort-hide,Hide sorting)#
|
Скрывает *элементы управления сортировкой* из списка записей, предотвращая изменение начальной сортировки.

!!! caution
    <<page-block-rl-presort,Предустановленная сортировка>> применяется независимо от этой опции.


| [#page-block-rl-limit]#[page-block-rl-limit,Records per page](#page-block-rl-limit,Records per page)#
|
Определяет максимальное количество записей, которое может отображаться на странице.

| [#page-block-rl-paging-hide]#[page-block-rl-paging-hide,Hide paging](#page-block-rl-paging-hide,Hide paging)#
|
Скрывает элементы управления пагинацией из списка записей, что предотвращает переход между разными страницами.

| [#page-block-rl-paging-full]#[page-block-rl-paging-full,Modify paging](#page-block-rl-paging-full,Modify paging)#
|
Изменяет элементы управления предыдущей/следующей страницы, включая список доступных страниц, что упрощает навигацию между страницами.

!!! caution
    Рекомендуется отключать эту опцию, когда модуль содержит большое количество записей.


| [#page-block-rl-count]#[page-block-rl-count,Show total record count](#page-block-rl-count,Show total record count)#
|
Показывает общее количество записей, соответствующих определённым фильтрам.

!!! caution
    Мы рекомендуем отключать эту опцию, когда модуль содержит большое количество записей.


| [#page-block-rl-export]#[page-block-rl-export,Allow records export](#page-block-rl-export,Allow records export)#
|
Включает опцию экспорта записей и определяет кнопку btn:[Export] в заголовке списка записей.

| [#page-block-rl-selection]#[page-block-rl-selection,Allow record selection](#page-block-rl-selection,Allow record selection)#
|
Добавляет ряд флажков в список записей, что позволяет выполнять операции над выбранными записями.

| [#page-block-rl-reminder-hide]#[page-block-rl-reminder-hide,Hide record reminder button](#page-block-rl-reminder-hide,Hide record reminder button)#
|
Скрывает кнопку *создания напоминания*, предотвращая создание напоминаний на основе записей непосредственно из списка записей.

Когда функция включена, напоминание можно создать, нажав на кнопку *напоминания* рядом с записью.

| [#page-block-rl-clone-hide]#[page-block-rl-clone-hide,Hide record clone button](#page-block-rl-clone-hide,Hide record clone button)#
|
Скрывает кнопку *клонирования записи*, предотвращая клонирование записей непосредственно из списка записей.

Когда функция включена, запись можно клонировать, нажав на кнопку *клонирования* рядом с записью.

| [#page-block-rl-edit-hide]#[page-block-rl-edit-hide,Hide record edit button](#page-block-rl-edit-hide,Hide record edit button)#
|
Скрывает кнопку *редактирования записи*, предотвращая открытие редактора записи непосредственно из списка записей.

Когда функция включена, редактор записи можно открыть, нажав на значок *редактирования* рядом с записью.

| [#page-block-rl-view-hide]#[page-block-rl-view-hide,Hide record view button](#page-block-rl-view-hide,Hide record view button)#
|
Скрывает кнопку *просмотра записи*, предотвращая открытие просмотрщика записи непосредственно из списка записей.

Когда функция включена, просмотрщик записи можно открыть, нажав на значок *просмотра* рядом с записью.

|===

<a id="page-block-record-organizer"></a>
## Организатор записей

**Блоки страницы органайзера записей** позволяют определять ряд колонок (*этапов*), в которых могут находиться записи данного модуля.
Записи перемещаются с помощью интерфейса перетаскивания.

Один блок страницы органайзера записей определяет одну колонку.
Добавьте дополнительные блоки страницы органайзера записей, чтобы определить дополнительные колонки.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-ro.png",
    "alias": "compose-configuration-page-block-reference-config-ro",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1252,
    "h": 977
  },
  "annotations": []
}

.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-ro-module]#[page-block-ro-module,Module](#page-block-ro-module,Module)#
|
Определяет модуль, который список записей будет использовать при взаимодействии с записями.

| [#page-block-ro-prefilter]#[page-block-ro-prefilter,Pre-filter records](#page-block-ro-prefilter,Pre-filter records)#
|
Определяет предустановленный фильтр, используемый при поиске и отображении записей в органайзере.
Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.

| [#page-block-ro-label]#[page-block-ro-label,Label field](#page-block-ro-label,Label field)#
|
Определяет, какое поле используется при отображении записи в органайзере записей.

| [#page-block-ro-descr]#[page-block-ro-descr,Description field](#page-block-ro-descr,Description field)#
|
Определяет дополнительный текст, используемый при отображении записи в органайзере записей.

| [#page-block-ro-sort]#[page-block-ro-sort,Record sort field](#page-block-ro-sort,Record sort field)#
|
Определяет, какое поле используется для задания порядка записей.
Когда запись перемещается, все связанные записи изменяют значение полей, чтобы отразить новый порядок.

| [#page-block-ro-key-f]#[page-block-ro-key-f,Key field](#page-block-ro-key-f,Key field)#
|
Определяет, какое поле используется для определения того, в какой колонке находится запись.
Когда запись перемещается, значение указанного поля изменяется, чтобы отразить новое состояние.

| [#page-block-ro-key-v]#[page-block-ro-key-v,Key value](#page-block-ro-key-v,Key value)#
|
Определяет, какое значение должно иметь ключевое поле, чтобы запись отображалась в данной колонке.

|===

<a id="page-block-record-revisions"></a>
## Версии записей

**Блоки страницы версий записей** позволяют отображать историю изменений, отслеживаемую для конкретного модуля.
История изменений выводится в порядке возрастания, с последним изменением внизу.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-record-revisions.png",
    "alias": "compose-configuration-page-block-reference-config-record-revisions",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1252,
    "h": 977
  },
    "annotations": []
}

.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-record-revisions-preload-revisions]#[page-block-record-revisions-preload-revisions,Preload record revisions](#page-block-record-revisions-preload-revisions,Preload record revisions)#
|
Определяет, должен ли блок страницы версий записей отображать историю изменений безусловно или блок страницы должен загружать историю изменений по требованию (когда флажок не отмечен).
Отключение опции может помочь с производительностью при более обширной истории изменений, но в целом её можно включать.

Если версии загружаются по требованию, блок страницы отображает кнопку btn:[show record revisions], которая позволяет пользователю загрузить историю изменений.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-record-revisions-on-demand.png",
    "alias": "compose-configuration-page-block-reference-config-record-revisions-on-demand",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 1380,
    "y": 360,
    "h": 352,
    "w": 543
  },
  "focus": {
    "x": 1410,
    "y": 90,
    "h": 882,
    "w": 473
  },
  "annotations": [{
    "x": 1540,
    "y": 517,
    "h": 32,
    "w": 217
  }]
}

| [#page-block-record-revisions-show-all-changes]#[page-block-record-revisions-show-all-changes,Show changes for all fields](#page-block-record-revisions-show-all-changes,Show changes for all fields)#
|
Определяет, должен ли блок страницы истории изменений показывать все изменённые значения или только их подмножество.
Если флажок не отмечен, таблица ниже позволяет выбрать, какие поля должен отображать блок страницы.

|===

<a id="page-block-record"></a>
## Запись

**Блок страницы записи** позволяет отображать содержимое записей на **странице записи**.
При создании или редактировании записи блоки страницы записи предоставляют способ ввода или изменения значений.

!!! important
    Блоки страницы записи доступны только на страницах записей.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-record.png",
    "alias": "compose-configuration-page-block-reference-config-record",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 334,
    "y": 32,
    "w": 1252,
    "h": 977
  },
  "annotations": []
}

!!! tip
    Вы можете определить несколько блоков страницы записи для одной и той же страницы записи, что позволяет группировать соответствующие вещи вместе.


<a id="page-block-sm"></a>
## Лента социальных медиа

**Блок страницы ленты социальных медиа** позволяет встраивать указанную ленту непосредственно в ваше приложение.

!!! note
    В настоящее время поддерживаются только *ленты Twitter*.


При добавлении на **списковую страницу** указанная лента предоставляется в виде фиксированного URL-адреса.

При добавлении на **страницу записи** указанная лента предоставляется либо в виде фиксированного URL-адреса, либо в виде поля URL.
Использование поля модуля вместо фиксированного значения позволяет ленте социальных медиа меняться в зависимости от контекста.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/config-sf.png",
    "alias": "compose-configuration-page-block-reference-config-sf",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 410,
    "y": 50,
    "w": 1100,
    "h": 340
  },
  "annotations": []
}

<a id="page-block-comment"></a>
## Комментарий

**Блок страницы комментария** позволяет отображать записи в формате, похожем на комментарии.

![...](compose-configuration/page-block-reference/config-comments.png)
![...](compose-configuration/page-block-reference/added-comments.png)


.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-comment-module]#[page-block-comment-module,Module](#page-block-comment-module,Module)#
|
Определяет модуль, который блок страницы будет использовать при отображении списка комментариев.

| [#page-block-comment-prefilter]#[page-block-comment-prefilter,Pre-filter](#page-block-comment-prefilter,Pre-filter)#
|
Определяет предустановленный фильтр, применяемый к записям при отображении списка комментариев.
Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.

| [#page-block-comment-title]#[page-block-comment-title,Title](#page-block-comment-title,Title)#
|
Определяет, какое поле модуля должно использоваться при отображении заголовка комментария.

| [#page-block-comment-content]#[page-block-comment-content,Content](#page-block-comment-content,Content)#
|
Определяет, какое поле модуля должно использоваться при отображении содержимого комментария.

| [#page-block-comment-reference]#[page-block-comment-reference,Reference](#page-block-comment-reference,Reference)#
|
Определяет, какое поле модуля будет использоваться для хранения ссылки на конкретную запись, для которой предназначен комментарий.
Это позволяет отображать конкретные комментарии для конкретных записей.

| [#page-block-comment-sort]#[page-block-comment-sort,Sort](#page-block-comment-sort,Sort)#
|
Определяет направление сортировки комментариев.
Сортировка выполняется на основе времени создания комментария.

|===

<a id="page-block-report"></a>
## Отчёт

**Блок страницы отчёта** позволяет отображать определённые компоненты отчётов внутри вашего приложения LowCoooode Low Code.

![...](compose-configuration/page-block-reference/config-report.png)
![...](compose-configuration/page-block-reference/added-report.png)

.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-report-report]#[page-block-report-report,Report](#page-block-report-report,Report)#
|
Определяет отчёт, который блок страницы будет использовать при отображении данных.

| [#page-block-report-display-element]#[page-block-report-display-element,Display element](#page-block-report-display-element,Display element)#
|
Определяет элемент отображения из выбранного отчёта, который блок страницы будет использовать при отображении данных.

|===

<a id="page-block-map"></a>
## Карта

**Блок страницы карты** позволяет отображать карту с маркерами, линиями и полигонами из разных значений полей Location различных источников записей.

![...](compose-configuration/page-block-reference/config-map.png)
![...](compose-configuration/page-block-reference/added-map.png)

.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-zoom-min]#[page-block-zoom-min,Zoom min](#page-block-zoom-min,Zoom min)#
|
Определяет минимальный уровень масштабирования карты.

| [#page-block-zoom-max]#[page-block-zoom-max,Zoom max](#page-block-zoom-max,Zoom max)#
|
Определяет максимальный уровень масштабирования карты.

| [#page-block-zoom-lockbounds]#[page-block-zoom-lockbounds,Lock bounds](#page-block-zoom-lockbounds,Lock bounds)#
|
Определяет, должна ли карта быть ограничена от панорамирования или увеличения за пределы текущей прямоугольной области карты, заданной её координатами долготы и широты. Это означает, что карта будет отображать только область в пределах указанных координат и не позволит пользователю панорамировать или приближать за эти пределы.

|===

<a id="page-block-navigation"></a>
## Навигация

**Блок страницы навигации** позволяет определять собственную навигацию внутри ваших приложений LowCoooode Low Code. Навигация может быть дополнением к предопределённой или полностью её заменой.

![...](compose-configuration/page-block-reference/config-navigation.png)
![...](compose-configuration/page-block-reference/added-navigation.png)

.Дополнительные конфигурации:
[cols="2s,5a"]
|===
| [#page-block-navigation-appearance]#[page-block-navigation-appearance,Appearance](#page-block-navigation-appearance,Appearance)#
|
Определяет стиль навигации

.Поддерживаемые стили отображения:
- Tabs +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation-tabs.png",
    "alias": "compose-configuration-page-block-reference-added-navigation-tabs",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}

- Pill +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation-pills.png",
    "alias": "compose-configuration-page-block-reference-added-navigation-pills",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}

- Small +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation-small.png",
    "alias": "compose-configuration-page-block-reference-added-navigation-small",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}

| [#page-block-navigation-justify]#[page-block-navigation-justify,Justify](#page-block-navigation-justify,Justify)#
|
Устанавливает одинаковую ширину для всех элементов навигации

.Опции:
- Justify +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation.png",
    "alias": "compose-configuration-page-block-reference-added-navigation.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}

- None +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation-justify-none.png",
    "alias": "compose-configuration-page-block-reference-added-navigation-justify-none",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}


| [#page-block-navigation-alignment]#[page-block-navigation-alignment,Alignment](#page-block-navigation-alignment,Alignment)#
|
Определяет выравнивание текста каждого элемента навигации

.Поддерживаемые выравнивания:
- Left +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation-left.png",
    "alias": "compose-configuration-page-block-reference-added-navigation-left",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}

- Center +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation-center.png",
    "alias": "compose-configuration-page-block-reference-added-navigation-center",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}

- Right +

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/page-block-reference/added-navigation-right.png",
    "alias": "compose-configuration-page-block-reference-added-navigation-right",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "h": 400,
    "w": 1599
  },
  "annotations": []
}

| [#page-block-navigation-item-type]#[page-block-navigation-item-type,Navigation item type](#page-block-navigation-item-type,Navigation item type)#
|
Определяет функциональность элемента навигации

.Поддерживаемые типы элементов навигации:
- Text: этот тип отображает элемент навигации просто как текст без дополнительных функций.
- URL: этот тип элемента навигации имеет опцию URL. При нажатии он перенаправляет на соответствующий URL.
- Compose page: этот тип элемента навигации имеет опцию страницы Compose. При нажатии он перенаправляет на соответствующую страницу Compose.
- Dropdown: этот тип элемента навигации имеет опцию выпадающего списка. При нажатии он отображает выпадающий список с соответствующими элементами.

|===
