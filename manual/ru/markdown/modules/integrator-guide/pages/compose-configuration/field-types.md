# Справочник типов полей

!!! important
    Тип поля и тип рабочего процесса могут не соответствовать одному и тому же типу.
    
    *DevNote* добавить некоторый дополнительный контекст об этом.


<a id="field-type-checkbox"></a>
## Checkbox (Y/N)

Поле `Checkbox` хранит значение true/false (истина/ложь) (логическое значение).

Поле `Checkbox` следует использовать, когда вы хотите хранить истинное/ложное значение.
Например, была ли оплачена подписка или подписан ли контакт на вашу рассылку.

Поле типа `Checkbox` при просмотре отображается как подпись.
При редактировании оно отображается как флажок.

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-checkbox-label-true]#[field-type-checkbox-label-true,True label](#field-type-checkbox-label-true,True label)#
|
Эта опция позволяет указать подпись отображения истинных значений.

| [#field-type-checkbox-label-false]#[field-type-checkbox-label-false,False label](#field-type-checkbox-label-false,False label)#
|
Эта опция позволяет указать подпись отображения ложных значений.

|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-checkbox-validation-format]#[field-type-checkbox-validation-format,Format](#field-type-checkbox-validation-format,Format)#
|
.Формат значения для истинного значения должен быть одним из следующих:
- `true`
- `t`
- `yes`
- `y`
- `1`

Любое неистинное значение считается ложным.
|===

<a id="field-type-datetime"></a>
## Дата и время

Тип поля `Date and time` хранит временное значение (метку времени).

Тип поля `Date and time` следует использовать, когда вы хотите хранить какую-либо временную информацию: дату, время или и то, и другое.
Например, время, когда был конвертирован лид.

Поле типа `Date and time` при просмотре отображается как форматированная строка.
При редактировании оно отображается как поле ввода даты, поле ввода времени или поле ввода даты-времени (в зависимости от конфигурации).

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-datetime-do]#[field-type-datetime-do,Date only](#field-type-datetime-do,Date only)#
|
Удаляет часть времени из метки времени и оставляет только дату.
При редактировании поля отображается только поле ввода даты.

| [#field-type-datetime-to]#[field-type-datetime-to,Time only](#field-type-datetime-to,Time only)#
|
Удаляет часть даты из метки времени и оставляет только время.
При редактировании поля отображается только поле ввода времени.

| [#field-type-datetime-pvo]#[field-type-datetime-pvo,Past values only](#field-type-datetime-pvo,Past values only)#
|
Изменяет валидатор значений так, чтобы разрешать только прошлые значения.
Отображаемый выборщик даты-времени отключает будущие значения.

| [#field-type-datetime-fvo]#[field-type-datetime-fvo,Future values only](#field-type-datetime-fvo,Future values only)#
|
Изменяет валидатор значений так, чтобы разрешать только будущие значения.
Отображаемый выборщик даты-времени также отключает прошлые значения.

| [#field-type-datetime-relative]#[field-type-datetime-relative,Output relative value](#field-type-datetime-relative,Output relative value)#
|
Отображает значение относительно текущего времени.
Например, `10min ago` или `last month`.

| [#field-type-datetime-fmt]#[field-type-datetime-fmt,Output format](#field-type-datetime-fmt,Output format)#
|
Определяет строку формата, используемую при отображении меток времени.
Поле поддерживает все опции форматирования [moment.js](https://momentjs.com/docs/#/displaying/format/).

|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-datetime-validation-format]#[field-type-datetime-validation-format,Format](#field-type-datetime-validation-format,Format)#
|
Поле `Date and time` поддерживает широкий диапазон форматов.
Ниже приведён список поддерживаемых форматов разбора даты-времени.

.Только дата:
** `2006-01-02`
** `02 Jan 06`
** `Monday, 02-Jan-06`
** `Mon, 02 Jan 2006`
** `2006/*1/*2`

.Только время:
** `15:04:05`
** `15:04`
** `15:04:05Z07:00`
** `15:04:05 MST`
** `15:04:05 -0700`
** `15:04 MST`
** `15:04Z07:00`
** `15:04 -0700`
** `3:04PM`

.Дата и время::
** `2006-01-02T15:04:05Z07:00`
** `Mon, 02 Jan 2006 15:04:05 -0700`
** `Mon, 02 Jan 2006 15:04:05 MST`
** `Monday, 02-Jan-06 15:04:05 MST`
** `02 Jan 06 15:04 -0700`
** `02 Jan 06 15:04 MST`
** `Mon Jan 02 15:04:05 -0700 2006`
** `Mon Jan 02 15:04:05 -0700 2006`
** `Mon Jan _2 15:04:05 2006`
** `2006/*1/*2 15:04:05`
** `2006/*1/*2 15:04`

Обычный формат времени — [`ISO 8601`](https://en.wikipedia.org/wiki/ISO_8601) с шаблоном `YYYY-MM-DDTHH:MM:SSZ`.

| [#field-type-datetime-validation-future]#[field-type-datetime-validation-future,Future](#field-type-datetime-validation-future,Future)#
|
Когда поле настроено разрешать только будущие значения, валидатор отклоняет любое значение, которое предшествует метке времени проведения проверки.

| [#field-type-datetime-validation-past]#[field-type-datetime-validation-past,Past](#field-type-datetime-validation-past,Past)#
|
Когда поле настроено разрешать только прошлые значения, валидатор отклоняет любое значение, которое следует за меткой времени проведения проверки.
|===

<a id="field-type-email"></a>
## Email

Тип поля `Email` хранит адрес электронной почты.

Тип поля `Email` следует использовать, когда вы хотите хранить электронную почту, например, основной адрес электронной почты клиента.

Поле типа `Email` при просмотре отображается как подпись или кликабельная ссылка (в зависимости от конфигурации).
При редактировании оно отображается как поле ввода электронной почты.

!!! important
    Внутри системы электронная почта хранится в виде обычного текста независимо от конфигурации.
    Форматирование отображения выполняется во фронтенд-приложении.


.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-email-plain]#[field-type-email-plain,Don't turn email into a link](#field-type-email-plain,Don't turn email into a link)#
|
Отображает электронную почту как обычную строку, а не как кликабельную ссылку.

|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-email-validation-format]#[field-type-email-validation-format,Format](#field-type-email-validation-format,Format)#
|
Поле `Email` принимает любой формат электронной почты, определённый `RFC 5322`.
Пример наиболее распространённого формата: `recipient@mail.tld`.

|===

<a id="field-type-select"></a>
## Select / dropdown

Тип поля `select / dropdown` хранит значение из предопределённого набора опций.

Тип поля `select / dropdown` следует использовать, когда вы хотите заставить пользователей выбрать одно из предопределённых значений.
Например, этап обращения или этап лида.

Поле типа `select / dropdown` при просмотре отображается как подпись.
При редактировании оно отображается как выпадающее поле ввода.

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-select-options]#[field-type-select-options,Options](#field-type-select-options,Options)#
|
Определяет набор доступных опций, которые может иметь поле.
Например, `new`, `in progress`, `closed`.

Опция — это пара значение-подпись, где подпись определяет, как значение отображается пользователю.

!!! important
    Внутри системы поле select представлено *значением*, а не подписью.
    При обращении к полю через автоматизацию или внешнюю интеграцию обязательно используйте значение, а не подпись.


| [#field-type-select-select]#[field-type-select-select,Multiple value input type](#field-type-select-select,Multiple value input type)#
|
Определяет, как представлен мультизначный вариант поля.

!!! note
    *DevNote* перечислить и описать опции?


|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-select-validation-value]#[field-type-select-validation-value,Value](#field-type-select-validation-value,Value)#
|
Поле `Select` допускает только использование значений, определённых в конфигурации поля.

|===

<a id="field-type-number"></a>
## Number

Тип поля `Number` хранит числовое значение.

Тип поля `Number` следует использовать, когда вы хотите хранить любое числовое значение.
Например, стоимость подписки или стоимость лида.

Поле типа `Number` при просмотре отображается как форматированная подпись.
При редактировании оно отображается как поле ввода числа.

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-number-prefix]#[field-type-number-prefix,Prefix](#field-type-number-prefix,Prefix)#
|
Добавляет указанный префикс к числовому значению.
Например, префикс `$` и значение `1000` дадут подпись `$1000`.

| [#field-type-number-suffix]#[field-type-number-suffix,Suffix](#field-type-number-suffix,Suffix)#
|
Добавляет указанный суффикс к числовому значению.
Например, суффикс `USD/h` и значение `1000` дадут подпись `1000USD/h`.

| [#field-type-number-precision]#[field-type-number-precision,Precision](#field-type-number-precision,Precision)#
|
Определяет точность, с которой должно храниться значение.
Например, точность 3 позволяет хранить числа с точностью до трёх десятичных знаков.

!!! important
    Точность ограничена значением 6.


| [#field-type-number-format]#[field-type-number-format,Format](#field-type-number-format,Format)#
|
Определяет строку формата, используемую при отображении чисел.
Поле поддерживает все опции форматирования [numeral.js](https://numeraljs.com/#format).

|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-number-validation-value]#[field-type-number-validation-value,Value](#field-type-number-validation-value,Value)#
|
Поле `Number` допускает только значения, которые могут быть представлены как 64-битное число с плавающей точкой.
Примеры наиболее распространённых представлений включают `1024` и `1024.256`.

|===

<a id="field-type-record"></a>
## Record

Тип поля `Record` хранит **ссылку** на другую запись.

!!! important
    Хранимое значение — это `recordID` связанной записи.
    Если вы хотите получить доступ к значениям связанной записи, вам нужно получить её из REST API.


Тип поля `Record` следует использовать, когда вы хотите определить связь между двумя модулями.
Например, родительскую транзакцию или владельца контрагента.

!!! caution
    Вы можете ссылаться только на записи *в том же пространстве имён*.


Поле типа `Record` при просмотре отображается как кликабельная ссылка на связанную запись.
При редактировании оно отображается как выпадающее поле с поиском.

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-record-module]#[field-type-record-module,Module](#field-type-record-module,Module)#
|
Определяет — к записям какого модуля привязано поле.

| [#field-type-record-label]#[field-type-record-label,Record label field](#field-type-record-label,Record label field)#
|
Определяет, какое поле модуля из связанного модуля должно использоваться при отображении его записей.

| [#field-type-record-prefilter]#[field-type-record-prefilter,Pre-filter records](#field-type-record-prefilter,Pre-filter records)#
|
Определяет предустановленный фильтр, используемый при поиске связанных записей.
Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.

| [#field-type-record-ff]#[field-type-record-ff,Query fields on search](#field-type-record-ff,Query fields on search)#
|
Определяет, какие поля используются для запроса при поиске по связанным записям.

| [#field-type-record-select]#[field-type-record-select,Multiple value input type](#field-type-record-select,Multiple value input type)#
|
Определяет, как представлен мультизначный вариант поля.

!!! note
    *DevNote* перечислить и описать опции?


|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-record-validation-value]#[field-type-record-validation-value,Value](#field-type-record-validation-value,Value)#
|
Поле `Record` допускает только идентификаторы записей (ссылки на другие записи), которые существуют в системе.

|===

<a id="field-type-string"></a>
## String

Тип поля `String` хранит обычное текстовое значение.

Тип поля `String` следует использовать, когда вы хотите хранить текст.
Например, имя контакта или пользовательское соглашение, которое должно быть показано клиенту.

Поле типа `String` при просмотре отображается как текст.
При редактировании оно отображается как поле ввода текста или поле ввода форматированного текста (в зависимости от конфигурации).

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-string-multi]#[field-type-string-multi,Multi-line](#field-type-string-multi,Multi-line)#
|
Изменяет простое поле ввода на многострочную текстовую область.

| [#field-type-string-rte]#[field-type-string-rte,Use rich text editor](#field-type-string-rte,Use rich text editor)#
|
Изменяет простой ввод строки на многострочный редактор форматированного текста.
Значение кодируется как стандартный HTML-документ, поэтому значение можно использовать в приложениях, где HTML принимается.

|===

<a id="field-type-url"></a>
## URL

Тип поля `URL` хранит URL-адрес.

Тип поля `URL` следует использовать, когда вы хотите хранить URL-адрес.
Например, ссылку на домашнюю страницу клиента.

Поле типа `URL` при просмотре отображается как подпись или кликабельная ссылка (в зависимости от конфигурации).
При редактировании оно отображается как поле ввода URL.

!!! important
    Внутри системы URL хранится в виде обычного текста независимо от конфигурации.
    Форматирование отображения выполняется во фронтенд-приложении.


.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-url-trimh]#[field-type-url-trimh,Trim # from the URL](#field-type-url-trimh,Trim # from the URL)#
|
Удаляет фрагмент URL (текст после `#`).

| [#field-type-url-trimq]#[field-type-url-trimq,Trim ? from the URL](#field-type-url-trimq,Trim ? from the URL)#
|
Удаляет параметры запроса URL (текст после `?`).

| [#field-type-url-ssl]#[field-type-url-ssl,only allow SSL (HTTPS) URLs](#field-type-url-ssl,only allow SSL (HTTPS) URLs)#
|
Разрешает только защищённые (HTTPS) URL-адреса.

| [#field-type-url-plain]#[field-type-url-plain,Don't turn URL into a link](#field-type-url-plain,Don't turn URL into a link)#
|
URL-адрес отображается как обычная подпись вместо кликабельной ссылки.

|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-url-validation-format]#[field-type-url-validation-format,Format](#field-type-url-validation-format,Format)#
|
Поле `URL` допускает только действительные строки URL, которые определяют схему и хост.
Шаблон допустимого URL: `[scheme:][//[userinfo@]host][/]path[?query][#fragment]`.

| [#field-type-url-validation-secure]#[field-type-url-validation-secure,Secure](#field-type-url-validation-secure,Secure)#
|
Когда поле настроено разрешать только защищённые URL, схема должна быть равна `https`.
|===

<a id="field-type-user"></a>
## User

Тип поля `User` хранит **ссылку** на системного пользователя.

!!! important
    Хранимое значение — это `userID` связанного пользователя.
    Если вы хотите получить доступ к значениям связанного пользователя, вам нужно получить их из REST API.


Тип поля `User` следует использовать, когда вы хотите определить связь между записью и системным пользователем.
Например, владелец записи или исполнитель задачи.

Поле типа `User` при просмотре отображается как подпись.
При редактировании оно отображается как выпадающее поле с поиском.

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-user-prefilter]#[field-type-user-prefilter,Pre-filter users](#field-type-user-prefilter,Pre-filter users)#
|
Определяет предустановленный фильтр, используемый при поиске связанного пользователя.
Обратитесь к [справочнику QL](modules/integrator-guide/pages/compose-configuration/accessing-lowcode/ql.md) за подробностями.

| [#field-type-user-preselect]#[field-type-user-preselect,Preset with current user](#field-type-user-preselect,Preset with current user)#
|
При установленном флажке текущий пользователь будет по умолчанию заполнять данное поле.

| [#field-type-user-role-filter]#[field-type-user-role-filter,User roles](#field-type-user-role-filter,User roles)#
|
Фильтрует пользователей на основе их членства в ролях.

| [#field-type-user-select]#[field-type-user-select,Multiple value input type](#field-type-user-select,Multiple value input type)#
|
Определяет, как представлен мультизначный вариант поля.

|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-user-validation-value]#[field-type-user-validation-value,Value](#field-type-user-validation-value,Value)#
|
Поле `User` допускает только идентификаторы пользователей (ссылки на пользователей), которые существуют в системе.

|===

<a id="field-type-file-upload"></a>
## File upload

Тип поля `File upload` хранит **ссылку** на загруженное вложение.

!!! important
    Хранимое значение — это `attachmentID` связанного пользователя.
    Если вы хотите получить доступ к значениям связанного файла, вам нужно получить их из REST API.


Тип поля `File upload` следует использовать, когда вы хотите прикрепить документ к записи.
Например, коммерческое предложение клиента или юридический документ.

Поле типа `File upload` при просмотре отображается как кликабельная ссылка.
При редактировании оно отображается как поле загрузки файла с перетаскиванием.

.Дополнительные конфигурации:
[cols="1s,5a"]
|===

| [#field-type-file-vm]#[field-type-file-vm,View mode](#field-type-file-vm,View mode)#
|
Определяет, как файлы представлены при просмотре.

!!! note
    *DevNote* перечислить и описать доступные опции.


|===

.Проверка значения:
[cols="1s,5a"]
|===

| [#field-type-file-validation-value]#[field-type-file-validation-value,Value](#field-type-file-validation-value,Value)#
|
Поле `File` допускает только идентификаторы файлов (ссылки на файлы), которые существуют в системе.

|===

<a id="field-type-location"></a>
## Location

Тип поля `Location` хранит местоположение, которое можно выбрать и отобразить на карте.

Тип поля `Location` следует использовать, когда вы хотите хранить местоположение.
Например, географическое местоположение клиента.

Поле типа `Location` отображается как местоположение на карте.
При редактировании оно отображается как выбираемое местоположение на карте.

Поле `Location` не определяет конфигурационные опции, специфичные для типа.
