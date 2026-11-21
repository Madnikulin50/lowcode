# Настройка хранения записей

Переработанное хранение записей LowCoooode теперь позволяет вам **настраивать способ хранения записей** «под капотом».
Настройка хранения записей позволяет **легко интегрировать внешние данные**, как если бы они были созданы через LowCoooode, **выгружать** наборы данных в **выделенные базы данных**, а большие наборы данных — в **специализированные центры данных**.

!!! note
    Чтобы хранить записи в другом подключении, вам сначала нужно будет [создать его](modules/integrator-guide/pages/compose-configuration/dal-connections/index.md).


По умолчанию **LowCoooode хранит** все записи (вместе с их значениями) в **одной таблице**, чего должно быть достаточно для **большинства базовых сценариев использования**.
Если ваш сценарий использования не требует хранения записей в выделенном хранилище, вы можете пропустить этот раздел.

!!! important
    Предлагаемые LowCoooode изменения схемы могут быть не самыми оптимальными и могут меняться в будущих версиях.
    Если вы предпочитаете иметь полный контроль над происходящим, лучше вносить необходимые изменения в базу данных вручную.


Когда вы настраиваете хранение записей, в случаях, когда используемая база данных не может хранить записи для этого модуля (например, таблица или колонка не существует), LowCoooode предлагает набор изменений, которые вы можете внести для правильного хранения данных.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/schema-alterations-base.png",
    "alias": "compose-configuration-dal-module-schema-alterations-base",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 340,
    "y": 0,
    "h": 490,
    "w": 1230
  },
  "focus": {
    "x": 400,
    "y": 40,
    "h": 400,
    "w": 1120
  },
  "annotations": []
}

Вы можете устранить эти изменения непосредственно через LowCoooode, используя предложенное изменение, нажав на кнопку btn:[Resolve] или btn:[Automatically Resolve].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/schema-alterations-base.png",
    "alias": "compose-configuration-dal-module-schema-alterations-apply",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 340,
    "y": 0,
    "h": 490,
    "w": 1230
  },
  "focus": {
    "x": 400,
    "y": 40,
    "h": 400,
    "w": 1120
  },
  "annotations": [{
    "x": 1375,
    "y": 147,
    "h": 27,
    "w": 65
  }, {
    "x": 1353,
    "y": 405,
    "h": 30,
    "w": 160
  }]
}

В качестве альтернативы вы можете внести изменения непосредственно в базу данных и отклонить изменение, нажав на кнопку btn:[Dismiss].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/schema-alterations-base.png",
    "alias": "compose-configuration-dal-module-schema-alterations-dismiss",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 340,
    "y": 0,
    "h": 490,
    "w": 1230
  },
  "focus": {
    "x": 400,
    "y": 40,
    "h": 400,
    "w": 1120
  },
  "annotations": [{
    "x": 1447,
    "y": 147,
    "h": 27,
    "w": 65
  }]
}

## Конфигурация

Параметры конфигурации хранения записей доступны на **странице редактирования модуля** во вкладке btn:[data store].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-base",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 352,
    "y": 0,
    "w": 1568,
    "h": 790
  },
  "annotations": [{
    "x": 617,
    "y": 187,
    "w": 120,
    "h": 44
  }]
}

### Подключение

Базовые параметры подключения определяют, **какое подключение** LowCoooode должно использовать при хранении записей модуля и **где в подключении** должны храниться данные — например, какая таблица или коллекция.

!!! caution
    Смена подключения приведёт к сбросу конфигурации уровня чувствительности на модуле и полях, если уровень чувствительности несовместим с новым подключением.
    // @todo add notes on module-level privacy
    // Refer to [privacy settings](modules/integrator-guide/pages/compose-configuration/compose-configuration/privacy.md) to learn more about sensitivity levels.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-connection",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 352,
    "y": 0,
    "w": 1568,
    "h": 790
  },
  "focus": {
    "padding": "xs",
    "x": 537,
    "y": 249,
    "w": 1200,
    "h": 240
  },
  "annotations": []
}

[cols="1s,5a"]
|===
| [#config-connection-connection]#[config-connection-connection,Database connection](#config-connection-connection,Database connection)#
|
**Подключение к базе данных** определяет, какое подключение LowCoooode должно использовать при обработке записей модуля.
По умолчанию LowCoooode использует основное подключение.

| [#config-connection-ident]#[config-connection-ident,Collection or database table name](#config-connection-ident,Collection or database table name)#
|
**Имя коллекции или таблицы базы данных** определяет, где внутри подключения хранятся записи модуля.

!!! important
    Если идентификатор опущен, используется идентификатор, определённый в подключении.
    Если подключение не предоставляет значение по умолчанию, используется системное значение по умолчанию `compose_record`.


В некоторых сценариях LowCoooode может создать соответствующие структуры, если они ещё не существуют в подключении (например, новую таблицу в базе данных RDBMS).

!!! note
    *DevNote* расширить примечания о том, когда таблицы создаются автоматически.


.Вы можете использовать следующие плейсхолдеры:
- [#ident-placeholder-module]#[ident-placeholder-module,`{\{module}}`](#ident-placeholder-module,`{\{module}}`)#: плейсхолдер `{\{module}}` заменяется на handle модуля или ID, если handle не определён.
- [#ident-placeholder-namespace]#[ident-placeholder-namespace,`{\{namespace}}`](#ident-placeholder-namespace,`{\{namespace}}`)#: плейсхолдер `{\{namespace}}` заменяется на slug пространства имён или ID, если slug не определён.

|===

### Сопоставление полей модуля

Параметры сопоставления и кодирования полей модуля определяют, какие поля модуля должны храниться и как они должны храниться.
Обратитесь к разделу [encoding-codec,кодеки кодирования полей](#encoding-codec,кодеки кодирования полей) за подробностями о доступных кодеках кодирования значений.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-fields",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 352,
    "y": 300,
    "w": 1568,
    "h": 580
  },
  "focus": {
    "x": 537,
    "y": 500,
    "w": 1200,
    "h": 210
  },
  "annotations": []
}

[cols="1s,5a"]
|===
| [#config-module-fields-map]#[config-module-fields-map,Map module field](#config-module-fields-map,Map module field)#
|
Флажок **сопоставления поля модуля** позволяет выбрать, должен ли LowCoooode хранить значения полей модуля или нет.
Если флажок не отмечен, LowCoooode отбросит значения.

Как правило, вы захотите хранить все поля модуля.
Если поле больше не нужно, его следует удалить из определения модуля.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-fields-use",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 487,
    "y": 450,
    "w": 1300,
    "h": 310
  },
  "focus": {
    "x": 537,
    "y": 500,
    "w": 1200,
    "h": 210
  },
  "annotations": [{
    "x": 567,
    "y": 542,
    "w": 77,
    "h": 19
  }]
}

| [#config-module-fields-codec]#[config-module-fields-codec,Codec](#config-module-fields-codec,Codec)#
|
**Кодек кодирования** определяет, какой кодек LowCoooode должен использовать при работе со значениями полей.
Обратитесь к разделу [encoding-codec,кодеки кодирования полей](#encoding-codec,кодеки кодирования полей) за подробностями.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-fields-codec",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 487,
    "y": 450,
    "w": 1300,
    "h": 310
  },
  "focus": {
    "x": 537,
    "y": 500,
    "w": 1200,
    "h": 210
  },
  "annotations": [{
    "x": 859,
    "y": 534,
    "w": 261,
    "h": 33
  }]
}

| [#config-module-fields-codec-conf]#[config-module-fields-codec-conf,Codec configuration](#config-module-fields-codec-conf,Codec configuration)#
|
**Конфигурация кодека** предоставляет необходимые параметры конфигурации выбранному кодека кодирования.
Обратитесь к разделу [encoding-codec,кодеки кодирования полей](#encoding-codec,кодеки кодирования полей) за подробностями.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-fields-ident",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 487,
    "y": 450,
    "w": 1300,
    "h": 310
  },
  "focus": {
    "x": 537,
    "y": 500,
    "w": 1200,
    "h": 210
  },
  "annotations": [{
    "x": 1152,
    "y": 534,
    "w": 552,
    "h": 33
  }]
}

|===

### Системные поля

Параметры **сопоставления и кодирования системных полей модуля** позволяют указать, какие поля модуля должны храниться и как они должны храниться.
Обратитесь к разделу [encoding-codec,кодеки кодирования полей](#encoding-codec,кодеки кодирования полей) за подробностями о доступных кодеках кодирования значений.

!!! tip
    Системные поля в большинстве своём предоставляют метаданные о записи.
    При необходимости вы можете исключить отдельные системные поля из хранения.
    Исключение системных полей может пригодиться при подключении к существующим базам данных, структура которых не определяет все поддерживаемые системные поля.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-system.png",
    "alias": "compose-configuration-dal-module-config-system",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 352,
    "y": 270,
    "w": 1568,
    "h": 850
  },
  "focus": {
    "padding": "xs",
    "x": 537,
    "y": 350,
    "w": 1200,
    "h": 590
  },
  "annotations": []
}

[cols="1s,5a"]
|===
| [#config-module-fields-map]#[config-module-fields-map,Map module field](#config-module-fields-map,Map module field)#
|
Флажок **сопоставления поля модуля** позволяет выбрать, должен ли LowCoooode хранить значения полей модуля или нет.
Если флажок не отмечен, LowCoooode отбросит значения.

Как правило, вы захотите хранить все системные поля модуля, но вы можете исключить те, которые не поддерживаются вашей собственной схемой.

!!! caution
    Исключение системных полей, таких как `id` и метки времени, может привести к непредвиденному поведению.
    Мы рекомендуем сохранять все системные поля.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-fields-use",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 487,
    "y": 450,
    "w": 1300,
    "h": 310
  },
  "focus": {
    "x": 537,
    "y": 500,
    "w": 1200,
    "h": 210
  },
  "annotations": [{
    "x": 567,
    "y": 542,
    "w": 77,
    "h": 19
  }]
}

| [#config-module-fields-codec]#[config-module-fields-codec,Codec](#config-module-fields-codec,Codec)#
|
**Кодек кодирования** определяет, какой кодек LowCoooode должен использовать при работе со значениями полей.
Обратитесь к разделу [encoding-codec,кодеки кодирования полей](#encoding-codec,кодеки кодирования полей) за подробностями.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-fields-codec",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 487,
    "y": 450,
    "w": 1300,
    "h": 310
  },
  "focus": {
    "x": 537,
    "y": 500,
    "w": 1200,
    "h": 210
  },
  "annotations": [{
    "x": 859,
    "y": 534,
    "w": 261,
    "h": 33
  }]
}

| [#config-module-fields-codec-conf]#[config-module-fields-codec-conf,Codec configuration](#config-module-fields-codec-conf,Codec configuration)#
|
**Конфигурация кодека** предоставляет необходимые параметры конфигурации выбранному кодека кодирования.
Обратитесь к разделу [encoding-codec,кодеки кодирования полей](#encoding-codec,кодеки кодирования полей) за подробностями.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/config-base.png",
    "alias": "compose-configuration-dal-module-config-fields-ident",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 487,
    "y": 450,
    "w": 1300,
    "h": 310
  },
  "focus": {
    "x": 537,
    "y": 500,
    "w": 1200,
    "h": 210
  },
  "annotations": [{
    "x": 1152,
    "y": 534,
    "w": 552,
    "h": 33
  }]
}

|===

<a id="encoding-codec"></a>
## Кодеки кодирования полей

Кодек кодирования полей определяет, как значение записи для соответствующего поля должно храниться и извлекаться в дальнейшем.
Доступные кодеки кодирования перечислены в таблице ниже с объяснением, конфигурацией и примерами использования.

[cols="1s,5a"]
|===
| [#field-codec-json]#[field-codec-json,JSON](#field-codec-json,JSON)#
|
Кодек поля JSON кодирует значение записи в объект JSON.
Расположение объекта JSON (относительно хранилища подключения, например колонки таблицы) определяется идентификатором, предоставленным вместе с определением кодека кодирования.

!!! note
    Эта стратегия кодирования используется по умолчанию для *пользовательских полей модуля*.


Кодек кодирования полезен, когда вы хотите извлекать значения из объектов JSON.

```json
```
!!! important
    .В настоящее время LowCoooode поддерживает только следующую структуру JSON (будущие версии планируют расширить гибкость):
    {
      "identifier": [<1>
        ...<2>
      ]
    }
    ----
    <1> Идентификатор значения должен совпадать с именем поля модуля.
    <2> Для записи хранится одно или несколько значений.
    Значения предоставляются в виде массива, чтобы поле могло переходить от однозначного к мультизначному.


| [#field-codec-alias]#[field-codec-alias,Alias](#field-codec-alias,Alias)#
|
Кодек поля alias кодирует значение записи как самостоятельное значение под пользовательским идентификатором.
Расположение (относительно хранилища подключения, например колонки таблицы) определяется идентификатором, предоставленным вместе с определением кодека кодирования.

Кодек кодирования полезен, когда вы извлекаете значения из выделенного расположения, которое может иметь другой идентификатор, чем поле модуля, например колонка таблицы или атрибут коллекции.

| [#field-codec-column]#[field-codec-column,Column](#field-codec-column,Column)#
|
Кодек поля column кодирует значение записи как самостоятельное значение, используя тот же идентификатор, что и поле модуля.
Расположение (относительно хранилища подключения, например колонки таблицы) определяется именем поля модуля.

!!! note
    Эта стратегия кодирования используется для *системных полей модуля*.


Кодек кодирования полезен, когда вы хотите извлекать значения из выделенного расположения с тем же идентификатором, что и поле модуля, например колонка таблицы или атрибут коллекции.

|===

## Устранение неполадок

## Проблемы модуля

Если модуль настроен неправильно или базовое подключение сообщает о каких-либо ошибках, на экране редактирования модуля отображается список выявленных проблем.
При наличии каких-либо проблем отображается новая вкладка btn:[issues].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/dal-module/troubleshooting-issues.png",
    "alias": "compose-configuration-dal-module-troubleshooting-issues",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 352,
    "y": 0,
    "w": 1568,
    "h": 1080
  },
  "annotations": [{
    "x": 1189,
    "y": 186,
    "w": 110,
    "h": 43
  }]
}
