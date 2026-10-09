# Подключения DAL

Подключения DAL (также называемые **подключениями**) определяют место, где LowCoooode может хранить и получать данные.
Подключение описывает физическое местоположение хранилища данных, его свойства (такие как кодирование и восстановление) и предоставляет параметры подключения.

!!! note
    .В настоящее время поддерживаются подключения только к следующим СУБД:
    * MySQL,
    * PostgreSQL,
    * SQLite.


Вы можете определить столько подключений, сколько необходимо, чтобы ваш бизнес соответствовал требованиям защиты данных, выносить отдельные наборы данных в специализированные базы данных и легко интегрировать внешние данные.

!!! tip
    Вы можете подключаться к базам данных с уже существующими данными, что исключает необходимость их переноса в LowCoooode.


Интерфейс управления подключениями доступен в LowCoooode Admin в разделе menu:system[connections]

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/list.png",
    "alias": "dal-connections-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 21,
    "y": 575,
    "h": 11,
    "w": 171
  }]
}

## Типы подключений

[cols="1s,5a"]
|===
| [#connection-type-primary]#[connection-type-primary,Основное подключение](#connection-type-primary,Основное подключение)#
|
Основное подключение — это подключение по умолчанию, которое LowCoooode использует для хранения записей; оно совпадает с `DB_DSN`, определённым в вашем файле `.env`.
Основное подключение определяется при развёртывании нового инстанса LowCoooode или обновлении со старой версии.

!!! important
    Основные подключения лишь частично поддерживают изменение и блокируют любые изменения, которые затронут их идентификацию или базовое подключение к хранилищу.


| [#connection-type-external]#[connection-type-external,Внешнее подключение](#connection-type-external,Внешнее подключение)#
|
Администратор добавляет внешние подключения, которые можно использовать в качестве альтернативы основному подключению.

!!! caution
    Изменение параметров подключения может привести к потере данных, если LowCoooode больше не сможет подключиться к исходному хранилищу.


|===

<a id="configuration"></a>
## Конфигурация подключения

### Базовые настройки

Базовые настройки дают общее представление о подключении: его идентификация, физическое местоположение, владелец и уровни чувствительности данных.
Уровни чувствительности данных рассматриваются в разделах о защите данных.

!!! note
    *DevNote* link to the data privacy section


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/config-basic.png",
    "alias": "dal-connections-config-basic",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 560,
    "y": 81,
    "w": 1107,
    "h": 486
  },
  "view": {
    "x": 352,
    "y": 0,
    "w": 1568,
    "h": 1080
  },
  "annotations": []
}

[cols="1s,5a"]
|===
| [#connection-config-basic-name]#[connection-config-basic-name,Название](#connection-config-basic-name,Название)#
|
Название задаёт понятную пользователю подпись для этого подключения.
Название используется при отображении подключений в пользовательских интерфейсах, например на карте в консоли защиты данных.

| [#connection-config-basic-handle]#[connection-config-basic-handle,Идентификатор (handle)](#connection-config-basic-handle,Идентификатор (handle))#
|
Идентификатор (handle) задаёт понятный пользователю идентификатор, используемый системой.
В некоторых случаях handle можно использовать вместо ID для однозначной идентификации подключения.

На handle распространяются те же ограничения, что и на любой другой handle.

1. начинаться с символа,
1. содержать не более 64 символов,
1. содержать только символы, цифры, _ (подчёркивание), - (дефис) или . (точка),
1. заканчиваться символом или цифрой.


| [#connection-config-basic-location-name]#[connection-config-basic-location-name,Название местоположения](#connection-config-basic-location-name,Название местоположения)#
|
Название местоположения указывает физическое местонахождение данных.
Название местоположения может быть названием дата-центра, региона или континента.

| [#connection-config-basic-location-coordinates]#[connection-config-basic-location-coordinates,Координаты местоположения](#connection-config-basic-location-coordinates,Координаты местоположения)#
|
Координаты местоположения указывают точное местонахождение данных.
При указании координат подключение отображается на карте в консоли защиты данных.

| [#connection-config-basic-ownership]#[connection-config-basic-ownership,Владение](#connection-config-basic-ownership,Владение)#
|
Владение определяет, кому принадлежат данные или дата-центр, где хранятся данные.
Указание владельца может пригодиться, если вы передаёте управление базой данных внешним поставщикам услуг.

| [#connection-config-basic-max-sensitivity-level]#[connection-config-basic-max-sensitivity-level,Максимально допустимый уровень чувствительности](#connection-config-basic-max-sensitivity-level,Максимально допустимый уровень чувствительности)#
|
Максимально допустимый уровень чувствительности задаёт верхний предел того, что можно хранить в данном подключении с точки зрения защиты данных.
Ни один модуль и ни одно поле, использующие это подключение, не могут превышать верхний предел.

Можно определить несколько подключений с разными уровнями чувствительности, чтобы лучше соответствовать требованиям защиты данных по всему миру.

|===

### Свойства подключения

Свойства подключения предоставляют дополнительную информацию о базе данных или дата-центре.
Свойства подключения в первую очередь касаются шифрования и восстановления данных.

!!! caution
    Свойства подключения описывают базовое хранилище, к которому мы подключаемся.
    Эти свойства *не* настраивают LowCoooode на выполнение каких-либо действий из описанных в свойствах.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/config-properties.png",
    "alias": "dal-connections-config-properties",
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
    "x": 560,
    "y": 190,
    "w": 1107,
    "h": 763
  },
  "annotations": []
}

[cols="1s,5a"]
|===
| [#connection-config-properties-encrypt-rest]#[connection-config-basic-encrypt-rest,Шифрование данных при хранении включено](#connection-config-basic-encrypt-rest,Шифрование данных при хранении включено)#
|
Шифрование при хранении означает, что данные шифруются при хранении в дата-центре.
Шифрование данных при хранении помогает снизить ущерб в случае взлома.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/config-properties.png",
    "alias": "dal-connections-config-properties-connection-config-properties-encrypt-rest",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 440,
    "y": 92,
    "h": 925,
    "w": 1397
  },
  "focus": {
    "x": 560,
    "y": 189,
    "w": 1106,
    "h": 763
  },
  "annotations": [{
    "x": 575,
    "y": 268,
    "h": 130,
    "w": 1076
  }]
}

| [#connection-config-properties-protected-rest]#[connection-config-basic-protected-rest,Защита данных при хранении обеспечена](#connection-config-basic-protected-rest,Защита данных при хранении обеспечена)#
|
Защита при хранении означает, что данные физически защищены от несанкционированного доступа при хранении в дата-центре.
Защита данных помогает снизить вероятность утечки данных.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/config-properties.png",
    "alias": "dal-connections-config-properties-connection-config-properties-protected-rest",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 440,
    "y": 92,
    "h": 925,
    "w": 1397
  },
  "focus": {
    "x": 559,
    "y": 189,
    "w": 1108,
    "h": 764
  },
  "annotations": [{
    "x": 576,
    "y": 420,
    "h": 129,
    "w": 1075
  }]
}

| [#connection-config-properties-encrypt-transit]#[connection-config-basic-encrypt-transit,Шифрование данных при передаче включено](#connection-config-basic-encrypt-transit,Шифрование данных при передаче включено)#
|
Шифрование при передаче означает, что данные шифруются при транспортировке между системами.
Шифрование данных при передаче помогает снизить риски атак с перехватом трафика.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/config-properties.png",
    "alias": "dal-connections-config-properties-connection-config-properties-encrypt-transit",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 440,
    "y": 92,
    "h": 925,
    "w": 1397
  },
  "focus": {
    "x": 560,
    "y": 190,
    "w": 1107,
    "h": 761
  },
  "annotations": [{
    "x": 575,
    "y": 570,
    "h": 131,
    "w": 1076
  }]
}

| [#connection-config-properties-data-backup-rest]#[connection-config-basic-data-backup-rest,Резервное копирование и восстановление данных при хранении реализованы](#connection-config-basic-data-backup-rest,Резервное копирование и восстановление данных при хранении реализованы)#
|
Резервное копирование и восстановление при хранении означает, что данные копируются и могут быть восстановлены в случае сбоев и катастроф.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/config-properties.png",
    "alias": "dal-connections-config-properties-connection-config-properties-data-backup-restore",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 440,
    "y": 92,
    "h": 925,
    "w": 1397
  },
  "focus": {
    "x": 560,
    "y": 190,
    "w": 1107,
    "h": 761
  },
  "annotations": [{
    "x": 576,
    "y": 723,
    "h": 128,
    "w": 1075
  }]
}

|===

### Параметры подключения

Параметры подключения предоставляют параметры, необходимые DAL LowCoooode для подключения к базовому хранилищу.

Параметры подключения определяют, как LowCoooode должен взаимодействовать с базовым хранилищем.
Параметры подключения также определяют некоторые значения по умолчанию для внутренних механизмов драйвера хранилища, например идентификатор таблицы или коллекции по умолчанию.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "dal-connections/config-connection.png",
    "alias": "dal-connections-config-connection",
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
    "y": 493,
    "w": 1108,
    "h": 573
  },
  "annotations": []
}

[cols="1s,5a"]
|===
| [#connection-config-conn-default-container]#[connection-config-conn-default-container,Название таблицы или контейнера по умолчанию](#connection-config-conn-default-container,Название таблицы или контейнера по умолчанию)#
|
Название таблицы или контейнера по умолчанию указывает, где LowCoooode должен хранить записи внутри подключения.
В контексте СУБД этот идентификатор указывает, какую таблицу должен использовать LowCoooode.

Поле ввода идентификатора поддерживает подстановочные значения, которые помогают сократить необходимую конфигурацию для крупных инстансов.

.Вы можете использовать следующие плейсхолдеры:
- [#ident-placeholder-module]#[ident-placeholder-module,`{\{module}}`](#ident-placeholder-module,`{\{module}}`)#: плейсхолдер `{\{module}}` заменяется на handle модуля или ID, если handle не определён.
- [#ident-placeholder-namespace]#[ident-placeholder-namespace,`{\{namespace}}`](#ident-placeholder-namespace,`{\{namespace}}`)#: плейсхолдер `{\{namespace}}` заменяется на slug пространства имён или ID, если slug не определён.


Например, идентификатор `compose*records*{\{namespace}}*{\{module}}` может дать идентификатор `compose*records*crm*lead`.

| [#connection-config-conn-type]#[connection-config-conn-type,Тип подключения и параметров](#connection-config-conn-type,Тип подключения и параметров)#
|
Тип подключения и параметров определяет, как будут передаваться параметры подключения и, соответственно, тип базового драйвера хранилища.
Обратитесь к разделу [drivers,типы параметров подключения](#drivers,типы параметров подключения) за полным обзором доступных вариантов и форматов их ввода.

| [#connection-config-conn-parameters]#[connection-config-conn-parameters,Параметры подключения](#connection-config-conn-parameters,Параметры подключения)#
|
Параметры подключения задают параметры подключения, используемые LowCoooode, и зависят от варианта, выбранного в [connection-config-conn-type,типе параметров подключения](#connection-config-conn-type,типе параметров подключения).
Обратитесь к подразделу [drivers,Типы подключений](#drivers,Типы подключений) за подробностями.

|===

<a id="drivers"></a>
## Типы параметров подключения

[cols="1s,5a"]
|===
| [#connection-driver-dsn]#[connection-driver-dsn,`lowcode::dal:connection:dsn`](#connection-driver-dsn,`lowcode::dal:connection:dsn`)#
|
Тип подключения `lowcode::dal:connection:dsn` означает, что для подключения к базе данных будет использоваться строка DSN.

1. Параметры конфигурации необходимо указать в следующем формате:
```json
```
{
  "dsn": ""
}

.В настоящее время LowCoooode поддерживает следующие базы данных:
- MySQL
- PostgreSQL
- SQLite

|===
