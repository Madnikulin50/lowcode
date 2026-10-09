<a id="2021-9-x"></a>
# `2021.9`

С LowCoooode `2021.9` мы продолжаем улучшать доступность, внедряя интернационализацию пользовательского интерфейса, а также пользовательские конфигурации.
Кроме того, мы улучшили общий дизайн и пользовательский опыт, усовершенствовали средства аутентификации и управления доступом, расширили существующий набор функций и добавили несколько новых возможностей.

**Интернационализация**

С помощью [интернационализации](2021.9@integrator-guide/personalization/i18n.md) мы добавляем поддержку перевода пользовательских интерфейсов, а также форматирования даты, времени и чисел с учётом локали.
LowCoooode позволяет полностью изменять встроенные переводы, а также определять переводы для дополнительных языков или любых пользовательских модификаций веб-приложений.
Вы также можете переводить большинство пользовательских конфигураций (таких как пространства имён и поля модулей), что позволяет настраивать ваши приложения Low Code для доступности.

!!! important
    Чтобы включить переводы для пользовательской конфигурации, необходимо установить переменную `.env` `LOCALE_RESOURCE_TRANSLATIONS_ENABLED=true`.


**UI/UX**

Развивая релиз `2021.3`, мы продолжаем улучшать дизайн пользовательского интерфейса и общий опыт взаимодействия с веб-приложениями LowCoooode.
Наиболее заметные изменения: переработана навигация по страницам, улучшена согласованность между различными веб-приложениями и добавлена более детальная фильтрация списков записей.

**Контроль доступа**

[Средство контроля доступа](2021.9@integrator-guide/access-control.md) получило обновление, обеспечивающее более тонкий контроль над доступом к вашим данным, например, к конкретным записям или модулям определённых пространств имён.
[Контекстные роли](2021.9@integrator-guide/access-control.md#role-type-ctx) позволяют определять членство в роли и, следовательно, доступ к ресурсам на основе [состояния системы](2021.9@integrator-guide/access-control.md#role-type-ctx) (например, какую запись мы редактируем).

!!! tip
    С помощью контекстных ролей вы можете охватить случаи, когда пользователю разрешён доступ или управление только теми данными, которые созданы или принадлежат ему.


**Аутентификация и безопасность**


Процесс аутентификации теперь стал более гибким благодаря поддержке [SAML](2021.9@administrator-guide/authentication/authentication-providers/index.md#saml) и разделённого потока учётных данных.
Дальнейшие улучшения генератора и обработчика токенов аутентификации повышают стабильность и безопасность системы.

**Шлюз интеграции**

Отвечая на потребности в определении пользовательских API-эндпоинтов, мы расширили механизм sink-маршрутов и представили [Шлюз интеграции](2021.9@integrator-guide/api-gw/index.md).
Улучшенный механизм упрощает процесс определения и управления пользовательскими API-эндпоинтами, а также обработки HTTP-запросов с помощью встроенных функций, рабочих процессов или [пользовательского кода](2021.9@integrator-guide/api-gw/index.md#js-processing).

!!! note
    Устаревший механизм sink-маршрутов всё ещё присутствует и работает как прежде, но мы планируем заменить sink-маршруты.


**Репортёр**

[Репортёр](2021.9@integrator-guide/reporting/index.md) предоставляет специализированное средство для определения и просмотра отчётов на основе данных, которые определяет ваш LowCoooode Low Code.

.[#2021*3-important]#[2021*3-important,Важные примечания к обновлению:](#2021_3-important,Важные примечания к обновлению:)#
[cols="1s,5a"]
|===
| [#2021*3-important-rbac]#[2021*3-important-rbac,Контроль доступа](#2021_3-important-rbac,Контроль доступа)#
|

LowCoooode 2021.9 перерабатывает внутреннюю структуру нашего механизма RBAC.
От процесса оценки доступа до способа кодирования правил.

.При обновлении до `2021.9`:
- LowCoooode автоматически просканирует все существующие RBAC-правила, удалит устаревшие правила (например, для обмена сообщениями), очистит существующие (переименование модуля федерации) и правильно изменит идентификаторы ресурсов.
- LowCoooode автоматически добавит, удалит и переименует роли по мере необходимости на основе переработанной системы:
** будут добавлены роли authenticated, anonymous, super admin,
*** роль `everyone` будет *удалена**,
** все RBAC-правила, принадлежащие устаревшей роли `everyone`, будут перенесены в роль `authenticated`,
** статические ID ролей (1 для everyone и 2 для admin) будут заменены последовательными ID.
Все членства и RBAC-правила будут перенесены на обновлённые ID ролей.
- LowCoooode автоматически добавит следующих системных пользователей:
** LowCoooode Provisioner (`provision@lowcode.local`, `lowcode-provisioner`); член учётной записи super admin, используется для всех действий по предоставлению.
** LowCoooode Service Account (`service@lowcode.local`, `lowcode-service`); член учётной записи super admin, используется для всех сервисных действий, интерфейса CLI.
** LowCoooode Federation (`federation@lowcode.local`, `lowcode-federation`); член учётной записи super admin, используется для всех действий федерации.

|===

:leveloffset: +1


# `2021.9.12`

**Релиз состоялся**: `2022-06-15`

.[#2021*9_12-fixed]#[2021*9*12-fixed,Исправлено:](#2021*9_12-fixed,Исправлено:)#
- Исправлена некорректная валидация учётных данных при регистрации пользователя, из-за которой пользователь создавался даже при неверных учётных данных (в основном пароле).
Исправление перемещает этап проверки пароля на более раннюю стадию процесса регистрации ([`#ff13912`](https://github.com/lowcode/lowcode-server/commit/ff13912)).
- Исправлен некорректный подсчёт записей для сгруппированных отчётов с участием полей с множественными значениями.
Исправление в некоторых случаях опускает дублирующиеся значения ([`7546dbb`](https://github.com/lowcode/lowcode-server/commit/7546dbb)).


# `2021.9.11`

**Релиз состоялся**: `2022-05-31`

.[#2021*9_11-fixed]#[2021*9*11-fixed,Исправлено:](#2021*9_11-fixed,Исправлено:)#
- Исправлено некорректное форматирование PostgreSQL для ISO-шаблонов времени путём добавления обработки граничного случая ([`#034a7f9`](https://github.com/lowcode/lowcode-server/commit/034a7f9)).
- Исправлено некорректное декодирование значений числовых и текстовых настроек путём обработки граничных случаев ([`#ced2daf`](https://github.com/lowcode/lowcode-server/commit/ced2daf)).
- Исправлено некорректное кодирование, декодирование и сохранение в хранилище переводов ресурсов полей модуля, из-за чего поле описания и поле подсказки путались; исправлено неверное сопоставление значений и удалены переводы из БД ([`#0422e5f`](https://github.com/lowcode/lowcode-server/commit/0422e5f), [`#e7ea299`](https://github.com/lowcode/lowcode-server/commit/e7ea299), [`#6c2dd28`](https://github.com/lowcode/lowcode-server/commit/6c2dd28)).
- Исправлена некорректная обработка запросов рабочего процесса, когда у пользователя открыто несколько экземпляров LowCoooode (вкладок) одновременно.
Веб-приложения LowCoooode улучшили коммуникацию о том, какие запросы уже разрешены и где должны отображаться запросы ([`#ef238df`](https://github.com/lowcode/lowcode-vue/commit/ef238df)).
- Исправлены неработающие выражения значений полей модуля, где использовалась ссылка на `old` запись.
Проблема возникала, потому что значение никогда не предоставлялось ([`#993cd22`](https://github.com/lowcode/lowcode-server/commit/993cd22)).


# `2021.9.10`

**Релиз состоялся**: `2022-04-20`

.[#2021*9_10-changed]#[2021*9*10-changed,Изменено:](#2021*9_10-changed,Изменено:)#
- Эндпоинты профилировщика (когда включены) перемещены на `/debug` вместо `/__profiler` для соответствия требованиям pprof ([`2653c3894`](https://github.com/lowcode/lowcode-server/commit/2653c3894)).

.[#2021*9_10-fixed]#[2021*9*10-fixed,Исправлено:](#2021*9_10-fixed,Исправлено:)#
- Исправлена некорректная загрузка статических переводов для пользовательских языков ([`964b71a56`](https://github.com/lowcode/lowcode-server/commit/964b71a56)).
- Исправлены утечки памяти и общие оптимизации производительности для санации содержимого ([`2711b0211`](https://github.com/lowcode/lowcode-server/commit/2711b0211)), загрузки рабочих процессов ([`35c1c0892`](https://github.com/lowcode/lowcode-server/commit/35c1c0892)) и обработки выражений ([`4eccaa826`](https://github.com/lowcode/lowcode-server/commit/4eccaa826)).
- Исправлена загрузка вложений рабочего процесса для всех поддерживаемых типов (`Reader`, `Bytes` и `String`) ([`7448a2d7b`](https://github.com/lowcode/lowcode-server/commit/7448a2d7b)).
- Исправлено дублирование запросов путём предотвращения повторной отправки запросов через веб-сокеты ([`f8a089a15`](https://github.com/lowcode/lowcode-server/commit/)f8a089a15).


# `2021.9.9`

**Релиз состоялся**: `2022-04-05`

.[#{PATCH*V}-added]#[{PATCH*V}-added,Добавлено:](#{PATCH_V}-added,Добавлено:)#
- Добавлена поддержка перевода ресурсов для метки поля `Boolean` ([`11af9dba7`](https://github.com/lowcode/lowcode-server/commit/11af9dba7), [`245c20e24`](https://github.com/lowcode/lowcode-webapp-compose/commit/245c20e24)).

.[#{PATCH*V}-changed]#[{PATCH*V}-changed,Изменено:](#{PATCH_V}-changed,Изменено:)#
- Санация переводов ресурсов ослаблена для поддержки более широкого спектра вариантов форматирования ([`4278e5823`](https://github.com/lowcode/lowcode-server/commit/4278e5823)).

.[#{PATCH*V}-fixed]#[{PATCH*V}-fixed,Исправлено:](#{PATCH_V}-fixed,Исправлено:)*
- Исправлены отсутствующие ссылки на роли при экспорте пространства имён, где поле модуля User определяло предфильтры ролей ([`cb44b6591`](https://github.com/lowcode/lowcode-server/commit/cb44b6591)).
- Исправлена некорректная переидентификация ресурсов при импорте пространства имён из-за игнорирования ресурсов ([`3b235e330`](https://github.com/lowcode/lowcode-server/commit/3b235e330)).
- Исправлен параметр трассировки рабочего процесса ([`3e068026`](https://github.com/lowcode/lowcode-server/commit/3e068026)).
- Исправлена некорректная санация содержимого переводов ресурсов для содержимого RTE ([`50671180d`](https://github.com/lowcode/lowcode-server/commit/50671180d)), атрибутов ссылок ([`f699d4b1e`](https://github.com/lowcode/lowcode-server/commit/f699d4b1e)).
- Исправлена опечатка в codegen actionlog ([`d44d396b3`](https://github.com/lowcode/lowcode-server/commit/d44d396b3)).
- Исправлены отсутствующие роли у имперсонализированного пользователя — роли не были включены в сгенерированный JWT ([`e4ba223da`](https://github.com/lowcode/lowcode-server/commit/e4ba223da)).
- Исправлено некорректное разбиение данных на чанки в итераторе ресурсов, из-за которого база данных выдавала ошибку для больших наборов данных ([`69c95a5a3`](https://github.com/lowcode/lowcode-server/commit/69c95a5a3)).
- Исправлена неработающая работа списка записей, когда на одной странице находятся дублирующиеся поля ([`c13dd1e81`](https://github.com/lowcode/lowcode-webapp-compose/commit/c13dd1e81)).

.[#{PATCH*V}-security]#[{PATCH*V}-security,Безопасность:](#{PATCH_V}-security,Безопасность:)#
- Добавлены дополнительные теги ссылок для повышения безопасности ([`889e2485f`](https://github.com/lowcode/lowcode-webapp-compose/commit/889e2485f)).


# `2021.9.8`

**Релиз состоялся**: `2022-03-07`

.[#{PATCH*V}-changed]#[{PATCH*V}-changed,Изменено:](#{PATCH_V}-changed,Изменено:)#
1. Цвет фона по умолчанию в `MetricConfigurator` изменён на белый ([`49a525a`](https://github.com/lowcode/lowcode-webapp-compose/commit/49a525a)).

.[#{PATCH*V}-fixed]#[{PATCH*V}-fixed,Исправлено:](#{PATCH_V}-fixed,Исправлено:)#
1. Исправлена функция выражения `count` для корректной обработки случая без аргументов при использовании типизированных значений (обычно при вызове через выражения рабочего процесса) ([`5d7d3aa3`](https://github.com/lowcode/lowcode-server/commit/5d7d3aa3)).
1. Исправлено некорректное присвоение ID блочного элемента страницы при создании страницы ([`95065f8e7`](https://github.com/lowcode/lowcode-server/commit/95065f8e7)).
1. Исправлено некорректное присвоение ID блочного элемента страницы при создании страницы ([`95065f8e7`](https://github.com/lowcode/lowcode-server/commit/95065f8e7)).
1. Исправлено управление ролями на внешних провайдерах аутентификации ([`d679a59`](https://github.com/lowcode/lowcode-webapp-compose/commit/d679a59)).
1. Добавлены пользовательские аргументы при выполнении скриптов Corredor через REST-обработчики Compose ([`2741e3577`](https://github.com/lowcode/lowcode-server/commit/2741e3577)).
1. Исправлена некорректная обработка неразрешённых значений (тип any) ([`5ff68c414`](https://github.com/lowcode/lowcode-server/commit/5ff68c414)).
1. Скрыта кнопка обновления данных, когда данные отсутствуют в `MetricConfigurator` ([`8fc446c`](https://github.com/lowcode/lowcode-webapp-compose/commit/8fc446c)).
1. Увеличен z-index для PageBuilder, чтобы панель инструментов с иконками не появлялась под заголовком и описанием блока ([`667a9b0`](https://github.com/lowcode/lowcode-webapp-compose/commit/667a9b0)).
1. Переведены сообщения в компоненте ошибок ([`acd5473`](https://github.com/lowcode/lowcode-webapp-compose/commit/acd5473), [`1e933b2`](https://github.com/lowcode/lowcode-js/commit/1e933b2), [`0af18b1`](https://github.com/lowcode/lowcode-locale/commit/0af18b1)).
1. Напоминания других людей отображаются ([`63b3efe`](https://github.com/lowcode/lowcode-webapp-compose/commit/63b3efe)).
1. Переводы конфигураторов полей и календарей ([`d4a379e`](https://github.com/lowcode/lowcode-webapp-compose/commit/d4a379e), [`5f00980`](https://github.com/lowcode/lowcode-webapp-compose/commit/5f00980)).
1. Относительные значения в диаграммах ([`eb5c4c7`](https://github.com/lowcode/lowcode-js/commit/eb5c4c7), [`025cfd5`](https://github.com/lowcode/lowcode-js/commit/025cfd5)).
1. Исправлено количество записей для экспорта в ExporterModal ([`eb5c4c7`](https://github.com/lowcode/lowcode-webapp-compose/commit/eb5c4c7), [`5f36917`](https://github.com/lowcode/lowcode-webapp-compose/commit/5f36917)).
1. Добавлена недостающая входная строка для пространства имён шлюза интеграции в admin ([`eb5c4c7`](https://github.com/lowcode/lowcode-locale/commit/eb5c4c7), [`64d27e1`](https://github.com/lowcode/lowcode-locale/commit/64d27e1)).
1. Исправлены накопительные вычисления значений воронкообразной диаграммы ([`0fbce36`](https://github.com/lowcode/lowcode-js/commit/0fbce36)).
1. Исправлено некорректное отображение воронкообразной диаграммы репортёра из-за неверного преобразования типа данных ([`c7ca7d1`](https://github.com/lowcode/lowcode-js/commit/c7ca7d1)).

.[#{PATCH*V}-security]#[{PATCH*V}-security,Безопасность:](#{PATCH_V}-security,Безопасность:)#
1. Усилен поток сброса пароля с помощью ограничения скорости и инвалидации существующих токенов ([`d2d0245d5`](https://github.com/lowcode/lowcode-webapp-workflow/commit/d2d0245d5)).
1. Улучшена валидация вложений по размеру и типу для загрузок в compose ([`6f19f00b2`](https://github.com/lowcode/lowcode-server/commit/6f19f00b2)).
1. Санирована обратная ссылка на странице выхода (XSS) ([`8c0a62284`](https://github.com/lowcode/lowcode-server/commit/8c0a62284)).
1. Рефакторинг: удаление всех сеансов пользователя со страницы аутентификации ([`23a8b757ee`](https://github.com/lowcode/lowcode-server/commit/23a8b757ee)).


# `2021.9.7`

**Релиз состоялся**: 2022-02-10

.[#2021*9_7-fixed]#[2021*9*7-fixed,Исправлено:](#2021*9_7-fixed,Исправлено:)#
- Исправлена функция выражения `set` для корректной обработки типов, предоставляемых выполнением рабочего процесса ([`f42d707`](https://github.com/lowcode/lowcode-server/commit/f42d707)).
- Исправлена проблема со строкой области видимости OIDC-провайдера ([`b1572e1`](https://github.com/lowcode/lowcode-server/commit/b1572e1)).
- Исправлены проблемы рабочего процесса с invoker/runner ([`c4d80b88`](https://github.com/lowcode/lowcode-server/commit/c4d80b88)).
- Исправлены переводы ресурсов для блочных элементов страницы контента и автоматизации ([`8a8cf42e`](https://github.com/lowcode/lowcode-server/commit/8a8cf42e)).
- Исправлен несовместимый порядок NULL для разных движков БД, вызывающий проблемы с курсором пагинации ([`2be460ee`](https://github.com/lowcode/lowcode-server/commit/2be460ee)).


# `2021.9.6`

**Релиз состоялся**: 2022-01-26

**Участники**:
Peter Grlica ([GH](https://github.com/petergrlica)),
Denis Arh ([GH](https://github.com/darh)),
Vivek Patel ([GH](https://github.com/vicpatel)),
Tomaž Jerman ([GH](https://github.com/tjerman)),
Katrin Yordanova ([GH](https://github.com/katrinDY)),
Jože Fortun ([GH](https://github.com/fajfa)).

.[#2021*9_6-added]#[2021*9*6-added,Добавлено:](#2021*9_6-added,Добавлено:)#
- Добавлена поддержка направления RTL.
- Добавлена поддержка дополнительных аргументов при вызове скриптов автоматизации через API ([`91eb88d2`](https://github.com/lowcode/lowcode-server/commit/91eb88d2)).
- Добавлена поддержка перевода ресурсов для опций поля выбора ([`c72902a8`](https://github.com/lowcode/lowcode-server/commit/c72902a8)).
- Добавлена поддержка подписанных SAML-запросов ([`5e4486c7`](https://github.com/lowcode/lowcode-server/commit/5e4486c7), [`7ec02f6e`](https://github.com/lowcode/lowcode-webapp-admin/commit/7ec02f6e)).
- Добавлена поддержка HTTP-привязки SAML ([`717cae5c0b`](https://github.com/lowcode/lowcode-server/commit/717cae5c0b), [`5cf0597b40`](https://github.com/lowcode/lowcode-webapp-admin/commit/5cf0597b40)).
- Добавлена опция скрытия кнопки импорта в списке записей ([`cd982a8`](https://github.com/lowcode/lowcode-webapp-compose/commit/cd982a8), [`9c85d0b`](https://github.com/lowcode/lowcode-js/commit/9c85d0b), [`874e0b2`](https://github.com/lowcode/lowcode-locale/commit/874e0b2)).
- Добавлена поддержка версионной переменной `.env` `DB_DSN`, упрощающей разработку ([`c3516dd`](https://github.com/lowcode/lowcode-server/commit/c3516dd)).
- Добавлены функции выражений `find` и `has` ([`86deaea9`](https://github.com/lowcode/lowcode-server/commit/86deaea9)).
- Добавлена функция выражения `sort` ([`be6b572`](https://github.com/lowcode/lowcode-server/commit/be6b572)).
- Добавлена поддержка фильтрации по множественным значениям в списках записей ([`d0213cb`](https://github.com/lowcode/lowcode-webapp-compose/commit/d0213cb)).
- Добавлен индикатор обработки при удалении записей в списке записей ([`6d47c1b`](https://github.com/lowcode/lowcode-webapp-compose/commit/6d47c1b)).


.[#2021*9_6-fixed]#[2021*9*6-fixed,Исправлено:](#2021*9_6-fixed,Исправлено:)#
- Рефакторинг выполнения рабочего процесса и исправлена логика runner/invoker ([`92224360`](https://github.com/lowcode/lowcode-server/commit/92224360)).
- Исправлены некорректные проверки разрешений на загрузку вложений из-за частичных параметров запроса ([#309](https://github.com/lowcode/lowcode-server/pull/309), [#346](https://github.com/lowcode/lowcode-webapp-compose/pull/346)).
- Исправлены некорректные преобразования аргументов QL в зависимости от используемой базы данных ([#308](https://github.com/lowcode/lowcode-server/pull/308)).
- Исправлена неработающая проверка здоровья Docker-контейнера на сервере Corredor ([`dce30ba`](https://github.com/lowcode/lowcode-server-corredor/commit/dce30ba)).
- Изменены локали федерации ([`91094e44`](https://github.com/lowcode/lowcode-webapp-compose/commit/91094e44)).
- Рефакторинг хранилища minio и исправление проблемы с неверным именем бакета ([`23a2446`](https://github.com/lowcode/lowcode-server/commit/23a2446)).
- Исправлена прокрутка, когда добавлен только один блочный элемент страницы ([`88f3e72`](https://github.com/lowcode/lowcode-webapp-compose/commit/88f3e72)).
- Исправлены переводы при сбросе пароля ([`87f08d8`](https://github.com/lowcode/lowcode-server/commit/87f08d8)).
- Исправлено неприменение правил RBAC и рабочих процессов после импорта через импорт пространства имён ([`ff6cadc0`](https://github.com/lowcode/lowcode-server/commit/ff6cadc0)).
- Исправлена некорректная обработка пустых строк перевода ресурсов (resource-translations) ([`4b264798`](https://github.com/lowcode/lowcode-server/commit/4b264798), [`043588f1`](https://github.com/lowcode/lowcode-webapp-compose/commit/043588f1)).
- Исправлен внешний OIDC для потока LowCoooode ([`398242b`](https://github.com/lowcode/lowcode-server/commit/398242b)).
- Исправлен расчёт процентов для всех типов диаграмм ([`2b7c949`](https://github.com/lowcode/lowcode-js/commit/2b7c949)).
- Исправлены переводы системных полей в средстве выбора полей ([`fa80ade`](https://github.com/lowcode/lowcode-vue/commit/fa80ade), [`485187e`](https://github.com/lowcode/lowcode-webapp-compose/commit/485187e)).
- Исправлен импорт пространства имён — дополнительная валидация запроса ([`ca178714b`](https://github.com/lowcode/lowcode-server/commit/ca178714b)), улучшена внутренняя идентификация ресурсов ([`00dd86fb9`](https://github.com/lowcode/lowcode-server/commit/00dd86fb9), [`d247ec678`](https://github.com/lowcode/lowcode-server/commit/d247ec678) и [`8be8be96d`](https://github.com/lowcode/lowcode-server/commit/8be8be96d)).
- Исправлен некорректный разбор содержимого API-запроса, когда `Content-Type` определяет кодировку ([`d0154cc1b`](https://github.com/lowcode/lowcode-server/commit/d0154cc1b)).
- Исправлена некорректная обработка ошибок для компонента загрузки ([`a0a1ac0`](https://github.com/lowcode/lowcode-webapp-compose/commit/a0a1ac0)).
- Исправлена настройка безопасности ролей для внешних провайдеров ([`05a40ca`](https://github.com/lowcode/lowcode-webapp-compose/commit/05a40ca)).
- Исправлена проблема с удалением OIDC-провайдера ([`c078808`](https://github.com/lowcode/lowcode-webapp-compose/commit/c078808)).
- Исправлено сопоставление имени очереди в событии onMessage рабочего процесса ([`79d8a842`](https://github.com/lowcode/lowcode-server/commit/79d8a842)).
- Исправлено отображение событий календаря для событий, охватывающих несколько месяцев ([`4d2f824`](https://github.com/lowcode/lowcode-js/commit/4d2f824)).
- Исправлен фильтр списка записей для числовых полей и полей ID ([`3b421a1`](https://github.com/lowcode/lowcode-webapp-compose/commit/3b421a1)).
- Исправлено переполнение боковой панели селектором пространства имён ([`b9aa41e`](https://github.com/lowcode/lowcode-webapp-compose/commit/b9aa41e)).
- Исправлены опции селектора полей блочного элемента метрики ([`cb69c80`](https://github.com/lowcode/lowcode-webapp-compose/commit/cb69c80)).
- Исправлен граничный случай рендеринга диаграммы, когда холст ещё не существовал ([`1b965f8`](https://github.com/lowcode/lowcode-webapp-compose/commit/1b965f8)).
- Исправлено центрирование метки пустого списка записей ([`f467e0f`](https://github.com/lowcode/lowcode-webapp-compose/commit/f467e0f)).
- Исправлено отсутствие проверки разрешений на удаление во встроенном редакторе записей ([`f74d398`](https://github.com/lowcode/lowcode-webapp-compose/commit/f74d398)).
- Исправлена некорректная обработка напоминаний, из-за которой они иногда не отображались ([`94247f0`](https://github.com/lowcode/lowcode-webapp-compose/commit/94247f0)).
- Исправлено неверное получение членства для закрытых ролей ([`88955eb`](https://github.com/lowcode/lowcode-webapp-admin/commit/88955eb)).
- Исправлено несохранение порядка столбцов для компонента ColumnPicker ([`9659d9f`](https://github.com/lowcode/lowcode-webapp-compose/commit/9659d9f), [`2b7c949`](https://github.com/lowcode/lowcode-js/commit/2b7c949), [`b53bc6b`](https://github.com/lowcode/lowcode-locale/commit/b53bc6b)).


.[#2021*9_6-changed]#[2021*9*6-changed,Изменено:](#2021*9_6-changed,Изменено:)#
- Значение по умолчанию больше не требуется для множественных полей ([`aae3f6f`](https://github.com/lowcode/lowcode-webapp-compose/commit/aae3f6f)).
- Предзаполнение связанных значений во встроенном редакторе записей ([`c457679`](https://github.com/lowcode/lowcode-webapp-compose/commit/c457679)).
- Отключение удаления страницы, если у неё есть дочерние страницы ([`67c3b2c`](https://github.com/lowcode/lowcode-webapp-compose/commit/67c3b2c)).


# `2021.9.5`

.[#2021*9_5-important]#[2021*9*5-important,Важные примечания к обновлению:](#2021*9_5-important,Важные примечания к обновлению:)#
- [Импорт записей](2021.9@devops-guide/cli/record-import.md) через команду CLI импорта требует указания определений полей модуля вместе с определениями импорта записей.
Это ограничение будет снято в будущих патч-релизах.

**Релиз состоялся**: 2021-12-16

**Участники**:
Katrin Yordanova ([GH](https://github.com/katrinDY)),
Vivek Patel ([GH](https://github.com/vicpatel)),
Jože Fortun ([GH](https://github.com/fajfa)),
Denis Arh ([GH](https://github.com/darh)),
Tomaž Jerman ([GH](https://github.com/tjerman)),
Peter Grlica ([GH](https://github.com/petergrlica)).

.[#2021*9_5-added]#[2021*9*5-added,Добавлено:](#2021*9_5-added,Добавлено:)#
- Безопасность ролей (запрещено, разрешено, принудительно) для внешних провайдеров аутентификации ([`91eb88d2`](https://github.com/lowcode/lowcode-server/commit/91eb88d2)).
- Функции рабочего процесса для управления вложениями ([`8a8c7685`](https://github.com/lowcode/lowcode-server/commit/8a8c7685)).
- Поддержка OIDC для lowcode ([`a2091db`](https://github.com/lowcode/lowcode-server/commit/a2091db)).
- Постфильтр JSON-ответа шлюза интеграции и HTTP-запрос в область видимости ([`b0590d2f`](https://github.com/lowcode/lowcode-server/commit/b0590d2f)).
- Функция рабочего процесса для генерации JWT ([`eecf8670`](https://github.com/lowcode/lowcode-server/commit/eecf8670)).
- Поддержка Unix strftime в функциях gval ([`d5001341`](https://github.com/lowcode/lowcode-server/commit/d5001341)).
- Приведение к целому числу в функциях gval ([`3473a267`](https://github.com/lowcode/lowcode-server/commit/3473a267)).
- Добавлена поддержка клонирования разрешений ролей ([`0564fe7`](https://github.com/lowcode/lowcode-server/commit/0564fe7),
[`d06cd41`](https://github.com/lowcode/lowcode-webapp-admin/commit/d06cd41), [`9bd530a`](https://github.com/lowcode/lowcode-js/commit/9bd530a)).
- Добавлена кнопка пользовательской обработки ([`6ec4157`](https://github.com/lowcode/lowcode-vue/commit/6ec4157)).
- Переводы для веб-приложения репортёра ([`6e26d5a`](https://github.com/lowcode/lowcode-webapp-reporter/commit/6e26d5a), [`931011c`](https://github.com/lowcode/lowcode-locale/commit/931011c)).
- Добавлены дополнительные функции для работы со строками в выражениях репортёра ([`a45c914e`](https://github.com/lowcode/lowcode-server/commit/a45c914e)).
- Добавлена поддержка [импорта записей](2021.9@devops-guide/cli/record-import.md) через команду CLI импорта ([`96556f54`](https://github.com/lowcode/lowcode-server/commit/96556f54)).


.[#2021*9_5-fixed]#[2021*9*5-fixed,Исправлено:](#2021*9_5-fixed,Исправлено:)#
- Переводимый контент (модули, имена, страницы, пространства имён) теперь отправляется с текущим языком ([`5bd3bd37`](https://github.com/lowcode/lowcode-webapp-compose/commit/5bd3bd37)).
- Улучшена медленная производительность в некоторых случаях с циклами WF путём увеличения сброса состояния сеанса рабочего процесса с 10 до 1000 ([`486a5752`](https://github.com/lowcode/lowcode-server/commit/486a5752)).
- Предотвращён сбой элементов отображения таблицы при получении пустого фрейма объединения ([`f244b7b2`](https://github.com/lowcode/lowcode-webapp-reporter/commit/f244b7b2)).
- Исправлены некорректные привязки маршрутов отчётов ([`d250827b`](https://github.com/lowcode/lowcode-webapp-reporter/commit/d250827b)).
- Добавлено недостающее определение типа в реестр парсера фильтров отчётов ([`08ef5ab10`](https://github.com/lowcode/lowcode-server/commit/08ef5ab10)).
- Исправлена паника при маршалинге фильтра отчёта в неверном состоянии ([`59ef8da1a`](https://github.com/lowcode/lowcode-server/commit/59ef8da1a)).
- Исправлена некорректная обработка блочных элементов страницы автоматизации, которые определяют кнопки без ссылок ([`047b647af`](https://github.com/lowcode/lowcode-server/commit/047b647af)).
- Предотвращён выход текста за пределы контейнера в EditorToolbox ([`147dcd7`](https://github.com/lowcode/lowcode-webapp-admin/commit/147dcd7)).
- Исправлены некорректные функции манипуляции датами для выражений репортёра ([`46372f55`](https://github.com/lowcode/lowcode-webapp-reporter/commit/46372f55)).
- Исправлена постоянная проблема «выполнить от имени» в рабочем процессе после удаления триггера ([`87f08d8b`](https://github.com/lowcode/lowcode-server/commit/87f08d8b)).
- Исправлены проверки контроля доступа в UI ([`e902382`](https://github.com/lowcode/lowcode-webapp-admin/commit/e902382)).
- Исправлено средство выбора членства в роли пользователя ([`cadb6e1`](https://github.com/lowcode/lowcode-webapp-admin/commit/cadb6e1)).
- Исправлено изменение порядка полей модуля при переходе ко всем записям ([`3525ef8`](https://github.com/lowcode/lowcode-webapp-compose/commit/3525ef8)).
- Исправлена некорректная отрисовка редакторов выбора записей и пользователей ([`09bca49`](https://github.com/lowcode/lowcode-webapp-compose/commit/09bca49)).
- Исправлен неправильный сброс формы опций конфигуратора элемента отображения ([`6b33655`](https://github.com/lowcode/lowcode-webapp-reporter/commit/6b33655)).
- Неработающие переводы при сбросе пароля ([`87f08d8`](https://github.com/lowcode/lowcode-server/commit/87f08d8)).
- Предотвращено преобразование сообщения об ошибке с помощью toLowerCase() ([`32e9325`](https://github.com/lowcode/lowcode-webapp-workflow/commit/32e9325)).

.[#2021*9_5-changed]#[2021*9*5-changed,Изменено:](#2021*9_5-changed,Изменено:)#
- Изменён конфигуратор элемента отображения диаграммы ([`b553f72`](https://github.com/lowcode/lowcode-webapp-reporter/commit/b553f72)).

.[#2021*9_5-removed]#[2021*9*5-removed,Удалено:](#2021*9_5-removed,Удалено:)#
- Удалена неиспользуемая функция агрегации метрик `COUNTD` ([`29e8ab2a`](https://github.com/lowcode/lowcode-js/commit/29e8ab2a), [`58a0a3b7`](https://github.com/lowcode/lowcode-locale/commit/58a0a3b7), [`4633b238`](https://github.com/lowcode/lowcode-webapp-compose/commit/4633b238)).


# `2021.9.4`

**Релиз состоялся**: 2021-11-26

**Участники**:
Katrin Yordanova ([GH](https://github.com/katrinDY)),
Vivek Patel ([GH](https://github.com/vicpatel)),
Jože Fortun ([GH](https://github.com/fajfa)),
Denis Arh ([GH](https://github.com/darh)),
Tomaž Jerman ([GH](https://github.com/tjerman)).

.[#2021*9_4-added]#[2021*9*4-added,Добавлено:](#2021*9_4-added,Добавлено:)#
- Добавлены опции `.env` для управления временем жизни OAuth2-токенов доступа и обновления ([`14450dc4`](https://github.com/lowcode/lowcode-server/commit/14450dc4)).
- Добавлена поддержка базовых ограничений пароля ([`420b5ee1`](https://github.com/lowcode/lowcode-server/commit/420b5ee1), [`984a5e99`](https://github.com/lowcode/lowcode-webapp-admin/commit/984a5e99)).
- Добавлены переводы для webapp-workflow ([`15d12b3`](https://github.com/lowcode/lowcode-webapp-workflow/commit/15d12b3), [`77de17e`](https://github.com/lowcode/lowcode-locale/commit/77de17e)).
- Добавлена кнопка пользовательской обработки ([`6ec4157`](https://github.com/lowcode/lowcode-vue/commit/6ec4157), [`1b67f4b`](https://github.com/lowcode/lowcode-webapp-compose/commit/1b67f4b)).
- Добавлен множественный выбор в запросах опций рабочего процесса ([`661781d`](https://github.com/lowcode/lowcode-vue/commit/661781d)).
- Добавлено предупреждение при изменении путей триггеров рабочего процесса ([`07705d21`](https://github.com/lowcode/lowcode-webapp-workflow/commit/07705d21)).
- Добавлен расширяемый редактор выражений для рабочих процессов ([`680149e`](https://github.com/lowcode/lowcode-webapp-workflow/commit/680149e)).


.[#2021*9_4-fixed]#[2021*9*4-fixed,Исправлено:](#2021*9_4-fixed,Исправлено:)#
- Исправлено некорректное приведение типа фильтра отчёта для ID-подобных значений ([`597484914`](https://github.com/lowcode/lowcode-server/commit/597484914)).
- Исправлена настройка валидации значения по умолчанию для поля записи модуля ([`aced989ae`](https://github.com/lowcode/lowcode-server/commit/aced989ae)).
- Санация логических значений (поле записи) ([`edbbf2f0`](https://github.com/lowcode/lowcode-server/commit/edbbf2f0)).
- Сериализация Uint64 JSON в полезных нагрузках Corredor ([`3241ff4e`](https://github.com/lowcode/lowcode-server/commit/3241ff4e)).
- Исправлена некорректная маркировка точек данных для радиальных диаграмм ([`4cbeb210`](https://github.com/lowcode/lowcode-js/commit/4cbeb210)).
- Исправлена пагинация таблицы отчётов ([`71dc2d2`](https://github.com/lowcode/lowcode-js/commit/71dc2d2)).
- Исправлено открытие справки рабочего процесса при вводе `?` ([`8fd0f15`](https://github.com/lowcode/lowcode-webapp-workflow/commit/8fd0f15)).


.[#2021*9_4-changed]#[2021*9*4-changed,Изменено:](#2021*9_4-changed,Изменено:)#
- Управление членством в роли администратора ([`2df2f48`](https://github.com/lowcode/lowcode-webapp-admin/commit/2df2f48)).


.[#2021*9_4-security]#[2021*9*4-security,Безопасность:](#2021*9_4-security,Безопасность:)#
- Обновлены пакеты Bluemonday и net ([`73c0b312`](https://github.com/lowcode/lowcode-server/commit/73c0b312)).


# `2021.9.3`

**Релиз состоялся**: 2021-11-10

**Участники**:
Denis Arh ([GH](https://github.com/darh)),
Jože Fortun ([GH](https://github.com/fajfa)),
Tomaž Jerman ([GH](https://github.com/tjerman))
Matija Rešek ([GH](https://github.com/resek)

.[#2021*9_3-changed]#[2021*9*3-changed,Изменено:](#2021*9_3-changed,Изменено:)#
- Все токены доступа пользователя теперь удаляются после смены пароля ([`01577191`](https://github.com/lowcode/lowcode-server/commit/01577191)).

.[#2021*9_3-fixed]#[2021*9*3-fixed,Исправлено:](#2021*9_3-fixed,Исправлено:)#
- Исправлена некорректная валидация ссылки на блочный элемент страницы для yaml-кодирования ([`5afc715f`](https://github.com/lowcode/lowcode-server/commit/5afc715f)).
- Исправлена нестабильная проверка RBAC, вызванная более сложной настройкой ролей и правил RBAC ([`a385fe1c`](https://github.com/lowcode/lowcode-server/commit/a385fe1c)).
- Исправлены ошибки обязательных полей с множественными значениями ([`a5e4fb21`](https://github.com/lowcode/lowcode-webapp-compose/commit/a5e4fb21)).
- Исправлена проблема, когда конфигуратор элементов отображения иногда не загружал правильную информацию при переключении между элементами ([`b8121e5`](https://github.com/lowcode/lowcode-webapp-reporter/commit/b8121e5)).

.[#2021*9_3-added]#[2021*9*3-added,Добавлено:](#2021*9_3-added,Добавлено:)#
- Добавлен текст ошибки handle в Admin ([`d903a735`](https://github.com/lowcode/lowcode-webapp-admin/commit/d903a735)).
- Добавлен текст ошибки handle в Reporter ([`963c2161`](https://github.com/lowcode/lowcode-webapp-reporter/commit/963c2161)).
- Добавлен текст ошибки handle в Workflow ([`cb1b42b3`](https://github.com/lowcode/lowcode-webapp-workflow/commit/cb1b42b3)).
- Добавлен переключатель сценариев в Reporter ([`4b6b52b3f`](https://github.com/lowcode/lowcode-server/commit/4b6b52b3f) [`b8121e5`](https://github.com/lowcode/lowcode-webapp-reporter/commit/b8121e5), [`79e13cb`](https://github.com/lowcode/lowcode-js/commit/79e13cb)).


# `2021.9.2`

.[#2021*9_2-important]#[2021*9*2-important,Важные примечания к обновлению:](#2021*9_2-important,Важные примечания к обновлению:)#
- Если вы хотите включить функции для взаимодействия рабочего процесса с журналом действий, необходимо установить переменную `.env` `ACTIONLOG*ENABLE*WORKFLOW_FUNCTIONS`.

**Релиз состоялся**: 2021-11-04

**Участники**:
Denis Arh ([GH](https://github.com/darh)),
Jože Fortun ([GH](https://github.com/fajfa)),
Tomaž Jerman ([GH](https://github.com/tjerman)),
Katrin Yordanova ([GH](https://github.com/katrinDY))

.[#2021*9_2-added]#[2021*9*2-added,Добавлено:](#2021*9_2-added,Добавлено:)#
- Добавлен пользовательский интерфейс для настройки SMTP-параметров; такая конфигурация не требует перезапуска сервера ([`0b69d1a2`](https://github.com/lowcode/lowcode-server/commit/0b69d1a2), [`20a85d8`](https://github.com/lowcode/lowcode-webapp-admin/commit/20a85d8)).
- Добавлена поддержка взаимодействия рабочего процесса с журналом действий (поиск, создание) ([`1014f53a`](https://github.com/lowcode/lowcode-server/commit/1014f53a)).
- Добавлена поддержка серверных плагинов ([`614d2b30`](https://github.com/lowcode/lowcode-server/commit/614d2b30)).
- Добавлено примечание об импорте рабочего процесса ([`9d98170`](https://github.com/lowcode/lowcode-webapp-workflow/commit/9d98170)).


.[#2021*9_2-changed]#[2021*9*2-changed,Изменено:](#2021*9_2-changed,Изменено:)*
- Поля модулей Compose больше не принимают зарезервированные системные имена — `recordID`, `ownedBy`, `createdBy`, `createdAt`, `updatedBy`, `updatedAt`, `deletedBy` и `deletedAt` ([`20757e58`](https://github.com/lowcode/lowcode-server/commit/20757e58), [`20a85d8`](https://github.com/lowcode/lowcode-webapp-admin/commit/20a85d8)).
- Переводы ресурсов больше не используют базовый язык в качестве запасного варианта при отсутствии перевода ([`4cd54a58`](https://github.com/lowcode/lowcode-server/commit/4cd54a58)).
- Веб-приложение Compose теперь отправляет HTTP-заголовки `Content-Language` и `Accept-Language` ([`f8427346`](https://github.com/lowcode/lowcode-webapp-compose/commit/f8427346)).
- Исходный код сервера теперь собирается с флагами -trimpath и без -mod=readonly ([`0b02535c`](https://github.com/lowcode/lowcode-server/commit/0b02535c)).
- Экспорт пространства имён больше не сохраняет ссылки на логотипы/иконки ([`dab413ece`](https://github.com/lowcode/lowcode-server/commit/dab413ece)).
- Начальный редизайн сайта документации ([14550adf](https://github.com/lowcode/lowcode-docs/commit/14550adf)).


.[#2021*9_2-fixed]#[2021*9*2-fixed,Исправлено:](#2021*9_2-fixed,Исправлено:)#
- Исправлена логика инициализации рабочего процесса на уровне загрузки, которая приводила к сбою сервера, если включённый рабочий процесс определял неверную конфигурацию триггера ([`415982c8`](https://github.com/lowcode/lowcode-server/commit/415982c8)).
- Исправлено сохранение рабочего процесса, когда боковая панель конфигурации была открыта ([`6d8796e`](https://github.com/lowcode/lowcode-webapp-workflow/commit/6d8796e)).
- Исправлено некорректное представление меток `Checkbox` Low Code для значений false ([`0330e31`](https://github.com/lowcode/lowcode-webapp-compose/commit/0330e31), [`aef1a14`](https://github.com/lowcode/lowcode-js/commit/aef1a14)).
- Исправлено экранирование акцентов и HTML в переведённых строках ([`556ffc5e`](https://github.com/lowcode/lowcode-server/commit/556ffc5e)).
- Исправлены проблемы перевода ресурсов для текущего языка, акцентов и экранированного HTML ({webapp*compose*commit_base}05178c2b[`05178c2b`]).
- Страницы скрываются, если родительская страница отмечена как невидимая ([`957a9de2`](https://github.com/lowcode/lowcode-webapp-compose/commit/957a9de2)).
- Исправлена неработающая настройка разрешений из панели администрирования Low Code для модуля для полей и записей ([`8ae2a48d`](https://github.com/lowcode/lowcode-webapp-compose/commit/8ae2a48d)).
- Исправлено присваивание значений expr в RenderOptions через селекторы ([`445f0ed5`](https://github.com/lowcode/lowcode-server/commit/556ffc5e)).
- Исправлен сбой миграции RBAC-правил из-за дублированных правил ([`e8bc6141`](https://github.com/lowcode/lowcode-server/commit/e8bc6141)).
- Проверка здоровья Docker-контейнера ([`9d7cf23c`](https://github.com/lowcode/lowcode-server/commit/9d7cf23c)).
- Исправлена метка ложного значения в compose ([`6da6989`](https://github.com/lowcode/lowcode-js/commit/6da6989)).
- Исправлена тень боковой панели ([`4a02d90`](https://github.com/lowcode/lowcode-vue/commit/4a02d90)).
- Исправлено изменение порядка столбцов таблицы репортёра ([`6b25473`](https://github.com/lowcode/lowcode-webapp-reporter/commit/6b25473)).
- Исправлена ошибка Low Code, когда скрипты автоматизации не загружены ([`bb94645`](https://github.com/lowcode/lowcode-webapp-compose/commit/bb94645)).
- Отключена кнопка загрузки в редакторе диаграмм ([`6912fcd`](https://github.com/lowcode/lowcode-webapp-compose/commit/6912fcd)).
- Исправлена некорректная обработка предфильтра в редакторе диаграмм ([`aeceb35`](https://github.com/lowcode/lowcode-webapp-compose/commit/aeceb35)).
- Исправлен некорректный экспорт страницы Low Code с ненастроенными или частично настроенными блочными элементами ([`032566d9`](https://github.com/lowcode/lowcode-server/commit/032566d9)).


# `2021.9.1`

**Релиз состоялся**: 2021-10-18

**Участники**:
Peter Grlica ([GH](https://github.com/petergrlica)),
Denis Arh ([GH](https://github.com/darh)),
Katrin Yordanova ([GH](https://github.com/katrinDY)),
Jože Fortun ([GH](https://github.com/fajfa)),
Vivek Patel ([GH](https://github.com/vicpatel)),
Matija Rešek ([GH](https://github.com/resek)),
Mario Burazer ([GH](https://github.com/MarioBur))

.[#2021*9_1-added]#[2021*9*1-added,Добавлено:](#2021*9_1-added,Добавлено:)#
- Добавлен текст ошибки handle/slug в Compose ([`c7f543ec`](https://github.com/lowcode/lowcode-webapp-compose/commit/c7f543ec)).
- Расширен экспорт записей с помощью фильтра в Compose ([`1f5d2abf`](https://github.com/lowcode/lowcode-webapp-compose/commit/1f5d2abf)).
- Добавлена подсказка для эндпоинта шлюза интеграции в Admin ([`d897ba3d`](https://github.com/lowcode/lowcode-webapp-admin/commit/d897ba3d)).
- Добавлены поля сортировки на сервере для шлюза интеграции для поддержки UI ([`c388f8`](https://github.com/lowcode/lowcode-server/commit/c388f8)).
- Добавлен тип выражения Bytes ([]byte) ([`614237`](https://github.com/lowcode/lowcode-server/commit/614237)).

.[#2021*9_1-changed]#[2021*9*1-changed,Изменено:](#2021*9_1-changed,Изменено:)#
- Улучшен выбор цветовой схемы в Compose ([`211227ba`](https://github.com/lowcode/lowcode-webapp-compose/commit/211227ba)).
- Предпросмотр шаблонов Admin открывается в новой вкладке ([`88f05df2`](https://github.com/lowcode/lowcode-webapp-admin/commit/88f05df2)).
- Рефакторинг шины сообщений для соответствия архитектуре уровней rbac, сервисов и пакетов ([`54b716`](https://github.com/lowcode/lowcode-server/commit/54b716)).
- Улучшена обработка фильтров шлюза интеграции ([`c6e3d0e9`](https://github.com/lowcode/lowcode-webapp-admin/commit/c6e3d0e9)).

.[#2021*9_1-fixed]#[2021*9*1-fixed,Исправлено:](#2021*9_1-fixed,Исправлено:)#
- Исправлены стили кнопок календаря в Compose.
- Исправлена кнопка «Назад» в просмотрщике записей.
- Исправлено отсутствие отражения настроек администрирования Compose в Compose ([`bf9e7064`](https://github.com/lowcode/lowcode-webapp-compose/commit/bf9e7064)).
- Исправлено неотображение сообщения об ошибке для неподдерживаемых MIME-типов в Compose ([`8561dca6`](https://github.com/lowcode/lowcode-webapp-compose/commit/8561dca6)).
- Исправлена обработка запроса при экспорте записей в Compose ([`78e6d296`](https://github.com/lowcode/lowcode-webapp-compose/commit/78e6d296)).
- Исправлена проблема дублирования шагов рабочего процесса на сервере ([`e2e751`](https://github.com/lowcode/lowcode-server/commit/e2e751)).
- Исправлено сопоставление уникальных ограничений для ресурсов на сервере ([`59ffe7`](https://github.com/lowcode/lowcode-server/commit/59ffe7)).
- Исправлено: отсутствие корневых SSL-сертификатов образа LowCoooode server, вызывавшее проблемы с исходящими HTTP и SMTP-запросами ([`8b008545`](https://github.com/lowcode/lowcode-server/commit/8b008545)).
- Исправлен неверный z-index для компонентов фильтра списка записей ([`6171af5b`](https://github.com/lowcode/lowcode-webapp-compose/commit/6171af5b)).


# `2021.9.0`

**Релиз состоялся**: 2021-10-11

**Участники**:
Tomaž Jerman ([GH](https://github.com/tjerman)),
Peter Grlica ([GH](https://github.com/petergrlica)),
Mia Arh ([GH](https://github.com/zmija)),
Denis Arh ([GH](https://github.com/darh)),
sgg-adraynrion ([GH](https://github.com/sgg-adraynrion)),
Katrin Yordanova ([GH](https://github.com/katrinDY)),
Jože Fortun ([GH](https://github.com/fajfa)),
Vivek Patel ([GH](https://github.com/vicpatel)),
Matija Rešek ([GH](https://github.com/resek)),
Mario Burazer ([GH](https://github.com/MarioBur)),
Bill Ewanick ([GH](https://github.com/billewanick))

.[#2021*9_0-added]#[2021*9*0-added,Добавлено:](#2021*9_0-added,Добавлено:)#
- Добавлена поддержка интернационализации веб-приложений LowCoooode ([#237](https://github.com/lowcode/lowcode-server/pull/237), [`31132570`](https://github.com/lowcode/lowcode-server/commit/31132570), [`e4eb28b8`](https://github.com/lowcode/lowcode-webapp-compose/commit/e4eb28b8), [`c3ff0ae1`](https://github.com/lowcode/lowcode-webapp-one/commit/c3ff0ae1)), а также некоторых пользовательских ресурсов (модулей Low Code, пространств имён и страниц) ([`46a7d94d`](https://github.com/lowcode/lowcode-server/commit/46a7d94d)).
Также включено форматирование чисел и даты/времени с учётом локали ([`da9a450f`](https://github.com/lowcode/lowcode-js/commit/da9a450f)).
- Добавлено системное средство для определения и обработки пользовательских API-эндпоинтов ([#232](https://github.com/lowcode/lowcode-server/pull/232), [`652cc074`](https://github.com/lowcode/lowcode-webapp-admin/commit/)652cc074).
Средство позволяет легко определять новые API-эндпоинты для вебхуков или пользовательских интеграций, необходимых для ваших бизнес-процессов.
Средство обеспечивает тесную интеграцию с [Workflows](modules/integrator-guide/pages/automation/workflows/index.md) для обработки запросов ([#245](https://github.com/lowcode/lowcode-server/pull/245)).
- Добавлено специализированное средство для создания, управления и запуска отчётов ([02b3e833](https://github.com/lowcode/lowcode-server/commit/02b3e833)).
Средство отчётности определяет выделенный пользовательский интерфейс ([lowcode-webapp-reporter](https://github.com/lowcode/lowcode-webapp-reporter)).
- Расширен набор функций Low Code:
** добавлена фильтрация на основе ролей для полей модуля пользователя ([`da181c30`](https://github.com/lowcode/lowcode-webapp-compose/commit/da181c30)),
** добавлена расширенная фильтрация списка записей с использованием фильтров по конкретным полям ([`5e7e8ce5`](https://github.com/lowcode/lowcode-webapp-compose/commit/5e7e8ce5)),
** добавлен блочный элемент «Комментарий» ([`1032399f`](https://github.com/lowcode/lowcode-webapp-compose/commit/1032399f)) и общие улучшения UI/UX для упрощения навигации,
** добавлены настраиваемые описания и подсказки для полей модуля,
** добавлен дополнительный параметр `namespaceID` при поиске по пространствам имён ([`21a3c5e6`](https://github.com/lowcode/lowcode-server/commit/21a3c5e6)).
- Добавлен генератор фиктивных данных, который можно использовать для создания тестовых записей и пользователей ([#216](https://github.com/lowcode/lowcode-server/pull/216)).
Генератор данных вызывается через CLI @todo CLI ref.
- Добавлена поддержка дублирования, импорта и экспорта целого пространства имён Low Code непосредственно из интерфейса Low Code ([`000664ef`](https://github.com/lowcode/lowcode-server/commit/000664ef), [`533b534f`](https://github.com/lowcode/lowcode-server/commit/533b534f)).
- Расширен набор функций аутентификации; добавлена возможность ручного отзыва сеансов аутентификации ([#254](https://github.com/lowcode/lowcode-server/pull/254), [#210](https://github.com/lowcode/lowcode-server/pull/210), [`1cb2e64d`](https://github.com/lowcode/lowcode-server/commit/1cb2e64d)), улучшены команды CLI для пользователей с дополнительными опциями ([`bed63c4f`](https://github.com/lowcode/lowcode-server/commit/bed63c4f), [`e4cd1f5b`](https://github.com/lowcode/lowcode-server/commit/e4cd1f5b)), а также добавлены `client_credentials` и имперсонализация пользователей ([`b245726c`](https://github.com/lowcode/lowcode-server/commit/b245726c), [`25e4d11f`](https://github.com/lowcode/lowcode-server/commit/25e4d11f)).
Пользовательский интерфейс клиента аутентификации теперь предоставляет серию примеров cURL для взаимодействия с клиентами аутентификации ([`16ae4c22`](https://github.com/lowcode/lowcode-webapp-admin/commit/16ae4c22)).
- Добавлена поддержка SAML-провайдеров аутентификации ([#188](https://github.com/lowcode/lowcode-server/pull/188), [`aedb2aef`](https://github.com/lowcode/lowcode-server/commit/aedb2aef), [`670b1609`](https://github.com/lowcode/lowcode-server/commit/670b1609)).
- Добавлены операции контроля доступа `*.search` RBAC для всех ресурсов [`92d2de86`](https://github.com/lowcode/lowcode-server/commit/92d2de86), [`f630a3d9`](https://github.com/lowcode/lowcode-server/commit/f630a3d9), [`0a388382`](https://github.com/lowcode/lowcode-server/commit/0a388382).
- Добавлена поддержка автоматизации, которая запускается до или после приостановки пользователя ([`13fc9d26`](https://github.com/lowcode/lowcode-server/commit/13fc9d26)).
- Расширен набор функций Workflow:
** добавлены учётные данные инициатора и исполнителя в начальную область видимости ([`806dbfaa`](https://github.com/lowcode/lowcode-server/commit/806dbfaa)),
** улучшена валидация триггеров на основе конфигурации рабочего процесса ([`f40f7982`](https://github.com/lowcode/lowcode-server/commit/f40f7982)),
** добавлены функции для взаимодействия с механизмом RBAC ([`89ae50db`](https://github.com/lowcode/lowcode-server/commit/89ae50db)),
** улучшен пользовательский интерфейс для отображения ошибок конфигурации и отладки (триггеры теперь также показывают ошибки),
** добавлен индикатор выполнения пробного запуска рабочего процесса.
- Расширен набор функций движка выражений:
** улучшена поддержка регулярных выражений ([`767f86f0`](https://github.com/lowcode/lowcode-server/commit/767f86f0)),
** улучшена работа с KV-подобными структурами ([`14b3f079`](https://github.com/lowcode/lowcode-server/commit/14b3f079), [`044d02bb`](https://github.com/lowcode/lowcode-server/commit/044d02bb)).
- Улучшены процесс установки и конфигурации системы, а также общая стабильность ([`5a67ecf7`](https://github.com/lowcode/lowcode-server/commit/5a67ecf7), [`a94e39b3`](https://github.com/lowcode/lowcode-server/commit/a94e39b3), [`a229d0ec`](https://github.com/lowcode/lowcode-server/commit/a229d0ec)):
** добавлена опция ограничения количества пользователей ([`1b3a811c`](https://github.com/lowcode/lowcode-server/commit/1b3a811c)),
** добавлена поддержка конфигурации файла `.env` из произвольного расположения через параметр командной строки `--env-file` ({SERVER*COMMIT*BASE6496027a[6496027a]}).
- Подготовлена инфраструктура хранилища для поддержки cockroachDB ([`109e23fc`](https://github.com/lowcode/lowcode-server/commit/109e23fc)).


.[#2021*9_0-changed]#[2021*9*0-changed,Изменено:](#2021*9_0-changed,Изменено:)#
- Пользовательский интерфейс веб-приложений LowCoooode был изменён для повышения согласованности, доступности ([`58aa46ee`](https://github.com/lowcode/lowcode-server/commit/58aa46ee), [`89ad4311`](https://github.com/lowcode/lowcode-server/commit/89ad4311)) и удобства использования.
Наиболее заметные изменения:
** навигация перемещена под левую боковую панель,
** верхняя панель определяет ярлыки для наиболее распространённых операций, связанных с просматриваемой страницей,
** средство выбора полей модуля было полностью переработано ([`8364da10`](https://github.com/lowcode/lowcode-server/commit/8364da10)).
- Изменён предпросмотр поля «Файл»: теперь он показывает последнюю загруженную вложенную картинку, когда выбрана опция «Одиночное изображение» ([`2d593af0`](https://github.com/lowcode/lowcode-webapp-compose/commit/2d593af0)).
- Переработан механизм контроля доступа RBAC, что обеспечило большую гибкость с правилами для конкретных ресурсов, контекстными ролями ([`2f2c055e`}) и улучшенным логированием (https://github.com/lowcode/lowcode-server/commit/922f4c31[`922f4c31`](https://github.com/lowcode/lowcode-server/commit/2f2c055e)).
LowCoooode теперь определяет ряд системных пользователей и ролей, которые используются для системных задач, таких как предоставление и федерация.
- Кнопки настройки правил RBAC для модуля Low Code, поля модуля и записи теперь находятся в одном выпадающем списке.
- Веб-приложение репортёра добавлено в список веб-приложений по умолчанию ([`e6950812`](https://github.com/lowcode/lowcode-server/commit/e6950812)).
- Изменено поведение отложенных триггеров рабочего процесса: теперь они игнорируют и пропускают пустые значения ограничений ([`8d9a3d54`](https://github.com/lowcode/lowcode-server/commit/8d9a3d54)).
- Обновлён zap logger до v1.19 ([`e48ffb2e`](https://github.com/lowcode/lowcode-server/commit/e48ffb2e)).
- Изменено системное логирование:
** ошибки заменены на предупреждения для проблем OAuth2 во время выполнения ([`0cb91793`](https://github.com/lowcode/lowcode-server/commit/0cb91793)),
** изменена трассировка стека логирования и добавлена поддержка управления уровнем глубины с помощью переменной `.env` `LOG*STACKTRACE*LEVEL` ([`28e1774c`](https://github.com/lowcode/lowcode-server/commit/28e1774c)).
- Настройки `PROVISION*SETTINGS*` перемещены в YAML-файл предоставления ([`2d78ae42`](https://github.com/lowcode/lowcode-server/commit/2d78ae42)).
- Базовый образ заменён на deb/ubuntu из-за несовместимости библиотек ([`00ba60e5`](https://github.com/lowcode/lowcode-server/commit/00ba60e5)).


.[#2021*9_0-removed]#[2021*9*0-removed,Удалено:](#2021*9_0-removed,Удалено:)#
- Удалены `PROVISION*SETTINGS*` в пользу YAML-файла предоставления ([`2d78ae42`](https://github.com/lowcode/lowcode-server/commit/2d78ae42)).
- Удалён параметр `query` из эндпоинта фильтрации списка записей ([`10e8b77d`](https://github.com/lowcode/lowcode-server/commit/10e8b77d)).
- Удалены Google карты из списка предоставляемых приложений ([`d6f24605`](https://github.com/lowcode/lowcode-server/commit/d6f24605)).
- Удалены устаревшие настройки для боковой панели пространства имён и LowCoooode One ([`b459bd35`](https://github.com/lowcode/lowcode-server/commit/b459bd35)).
- Удалены вкладки и панели в LowCoooode One.


.[#2021*9_0-fixed]#[2021*9*0-fixed,Исправлено:](#2021*9_0-fixed,Исправлено:)#
- Исправлены неработающие ссылки в README ([`7974ca65`](https://github.com/lowcode/lowcode-server/commit/7974ca65)).
- Исправлено несоответствие имени JSON-пропа `grant-validGrant` у клиента аутентификации ([`40ddb9db`](https://github.com/lowcode/lowcode-server/commit/40ddb9db)).
- Исправлены ошибки загрузки вложений при загрузке пустого вложения или ico-файла ([`f5532acf`](https://github.com/lowcode/lowcode-server/commit/f5532acf)).
- Удалён ненужный контент из проверки содержимого обслуживаемого веб-приложения ([`3638ecac`](https://github.com/lowcode/lowcode-server/commit/3638ecac)).
- Исправлена ошибка монтирования, когда веб-приложения отключены ([`63dbe702`](https://github.com/lowcode/lowcode-server/commit/63dbe702)).
- Исключены удалённые напоминания из эндпоинта API списка напоминаний ([`9f74d5c0`](https://github.com/lowcode/lowcode-server/commit/9f74d5c0)).
- Предотвращены дублирующиеся значения в полях множественного выбора.
- Исправлена ошибка дублирования задач в календарях ([`2e322054`](https://github.com/lowcode/lowcode-js/commit/2e322054)).
- Исправлен регистронезависимый поиск пространств имён ([`5ce9572d`](https://github.com/lowcode/lowcode-webapp-compose/commit/5ce9572d)).
- Исправлено некорректное приведение типов в actionlog, которое приводило к неработающим сообщениям лога, когда фронтендный стек технологий не мог разобрать значения ([`5ac8790b`](https://github.com/lowcode/lowcode-server/commit/5ac8790b), [`d1ccbc3e`](https://github.com/lowcode/lowcode-server/commit/d1ccbc3e)).
- Исправлено некорректное сообщение об ошибке, если пользователю не разрешён поиск по пространствам имён ([`7cf6c18d`](https://github.com/lowcode/lowcode-server/commit/7cf6c18d)).
- Исправлены отсутствующие уведомления в веб-приложениях.
- Исправлены опечатки в сообщениях об ошибках Envoy ([`0a241fab`](https://github.com/lowcode/lowcode-server/commit/0a241fab)).
- Исправлено исчезновение уведомлений при смене текущей страницы.
- Исправлено ClaimsToIdentify для возврата идентификатора со всеми аутентифицированными ролями ([`67d7882b`](https://github.com/lowcode/lowcode-server/commit/67d7882b)).
- Добавлены отсутствующие свойства контроля доступа в ответы ресурсов ([`774354d6`](https://github.com/lowcode/lowcode-server/commit/774354d6)).
- Добавлены отсутствующие проверки контроля доступа для напоминаний ([`03344782`](https://github.com/lowcode/lowcode-server/commit/03344782)).
- Исправлено некорректное отображение разрешений в веб-приложении администрирования, если у пользователя недостаточно прав.
- Исправлено некорректное представление состояния сеанса автоматизации для сеансов с запросами ([`234d3795`](https://github.com/lowcode/lowcode-server/commit/234d3795)).
- Исправлено приведение параметров и возвращаемых значений функций выражений для строковых функций.
- Добавлены отсутствующие обёртки ответов синхронизации структуры федерации ([`8ee91eb7`](https://github.com/lowcode/lowcode-server/commit/8ee91eb7)).
- Повышена общая стабильность системы.


.[#2021*9_0-security]#[2021*9*0-security,Безопасность:](#2021*9_0-security,Безопасность:)#
- Команды CLI теперь используют системного пользователя при выполнении команд ([`dca5757f`](https://github.com/lowcode/lowcode-server/commit/dca5757f)).
- Контроль доступа при импорте/предоставлении перемещён из Envoy в вызывающий сервис ([`a2b964c5`](https://github.com/lowcode/lowcode-server/commit/a2b964c5)).


.[#2021*9_0-development]#[2021*9*0-development,Разработка:](#2021*9_0-development,Разработка:)#
- Определено надлежащее средство для тестирования логики обработки шлюза интеграции ([`6ceadf40`](https://github.com/lowcode/lowcode-server/commit/6ceadf40)).
- Логика codegen для функций хранилища теперь может определять импорты, специфичные для них ([`b95e878c`](https://github.com/lowcode/lowcode-server/commit/b95e878c)).
- Конвейеры сборки и интеграции перенесены на Github Actions.
- Удалена вводящая в заблуждение директория федерации `etc/` ([`d4505482`](https://github.com/lowcode/lowcode-server/commit/d4505482)).
- Удалён давно устаревший storybook ([`76270476`](https://github.com/lowcode/lowcode-webapp-compose/commit/76270476)).
- Реализована функция C3 и применена к веб-приложениям ([`a318b380`](https://github.com/lowcode/lowcode-webapp-compose/commit/a318b380), [`4c5e2c24`](https://github.com/lowcode/lowcode-vue/commit/4c5e2c24)).


:leveloffset: -1
