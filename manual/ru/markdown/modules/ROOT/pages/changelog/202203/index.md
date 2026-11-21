<a id="2022-3-x"></a>
# `2022.3`

С LowCoooode `2022.3` мы продолжаем улучшать стабильность системы, безопасность и производительность, а также внедряем новые улучшения качества жизни при установке и разработке.
Кроме того, мы представляем новый LowCoooode Discovery и профилировщик шлюза интеграции.

[**LowCoooode Discovery**](2022.3@integrator-guide/discovery/index.md)

[Discovery](2022.3@integrator-guide/discovery/index.md) предоставляет мощную поисковую систему для взаимодействия с вашими данными с интуитивно понятным пользовательским интерфейсом для поиска и, в некоторых случаях, визуализации данных, таких как географические метаданные.

!!! important
    Чтобы включить LowCoooode Discovery, вам необходимо настроить ваш экземпляр.
    Обратитесь к [документации по установке](2022.3@devops-guide/discovery/index.md) за инструкциями.


**Улучшения UX конфигурации LowCoooode**

Чтобы упростить процесс настройки приложения LowCoooode Low Code, мы внедрили несколько улучшений качества жизни: подсказки предпросмотра конфигурации шагов рабочего процесса, дополнительные всплывающие подсказки и метки, предложения меток шагов рабочего процесса и стратегию удаления страниц с дочерними страницами.

Кроме того, мы представляем [профилировщик маршрутов](2022.3@integrator-guide/api-gw/profiler.md) шлюза интеграции, который поможет вам тестировать производительность системы, автоматизировать, устранять неполадки и оптимизировать.

**Улучшения UX развёртывания LowCoooode**

Чтобы упростить начальные шаги по устранению неполадок при развёртывании LowCoooode, мы представляем [Web Console](2022.3@devops-guide/troubleshooting/web-console.md).
Веб-консоль позволяет быстро получить доступ к системным логам, проверкам здоровья и другим параметрам, необходимым для отладки неисправного экземпляра.

!!! tip
    Веб-консоль наиболее полезна в сценариях, когда у вас может не быть (или вы не хотите давать) прямой доступ к вашим серверам.


:leveloffset: +1


# `2022.3.6`

**Дата выпуска**: `2022.11.08`

.[#2022*03*06-changed]#[2022*03*06-changed,Изменено:](#2022*03*06-changed,Изменено:)#
- Изменены ограничения на пароль, которые сохраняли свои значения в виде строк, теперь они сохраняются в виде чисел.
Кроме того, подписи двух полей были перефразированы в соответствии со стилем остальных полей ввода в разделе ограничений на пароль.
Изменение было внесено, потому что ограничения на пароль сохранялись в виде строк; с этим изменением ограничения будут сохраняться в виде чисел.
Изменение внесено путём добавления 'number' к модели ввода ограничений на пароль.
([`1b8e02c`](https://github.com/lowcode/lowcode-locale/commit/1b8e02c), [`1d8eb4b`](https://github.com/lowcode/lowcode-locale/commit/1d8eb4b), [`9fd49e1`](https://github.com/lowcode/lowcode-webapp-admin/commit/9fd49e1), [`53d8619`](https://github.com/lowcode/lowcode-webapp-admin/commit/53d8619), [`651993b`](https://github.com/lowcode/lowcode-server/commit/651993b), [`4c4f1df`](https://github.com/lowcode/lowcode-server/commit/4c4f1df)).

.[#2022*03*06-fixed]#[2022*03*06-fixed,Исправлено:](#2022*03*06-fixed,Исправлено:)#
- Исправлена ситуация, когда значения геометрии не сохранялись при ручном изменении в полях ввода.
Исправление внесено путём корректного определения изменений значения геометрии, поскольку оно использует более сложную структуру, чем остальные поля ввода.
([`1c220ec`](https://github.com/lowcode/lowcode-webapp-compose/commit/1c220ec)).
- Исправлен фильтр экспорта записей, который не интерполировал плейсхолдеры вроде `${userID}`, что приводило к неработающему экспорту.
Исправление внесено путём добавления вычисления в фильтр экспорта записей для корректного вычисления и интерполяции плейсхолдеров.
([`ee123ea`](https://github.com/lowcode/lowcode-webapp-compose/commit/ee123ea)).
- Исправлена ситуация, когда загрузка файлов отклонялась при указании, какие MIME-типы разрешены или нет.
Проблема возникала из-за дополнительных метатегов, предоставляемых файлами или базовыми библиотеками.
Исправление внесено путём переработки того, как проверяются MIME-типы: вместо сравнения меток MIME-типов теперь используется более надёжная библиотека, которая уже использовалась для определения MIME-типов.
([`84e2ff1`](https://github.com/lowcode/lowcode-server/commit/84e2ff1)).
- Исправлено падение сервера Corredor при запуске из-за ошибки nil pointer.
Исправление внесено путём добавления проверки на nil для методов Vars, чтобы они не вызывали непредвиденных ошибок при использовании значения nil.
([`9ad5b36`](https://github.com/lowcode/lowcode-server/commit/9ad5b36)).
- Исправлено зависание на странице регистрации при ошибке.
Исправление внесено путём перенаправления обратно на страницу регистрации с соответствующим сообщением об ошибке.
([`be1e035`](https://github.com/lowcode/lowcode-server/commit/be1e035)).
- Исправлена ситуация, когда переводы ресурсов кнопок автоматизации не сохранялись между обновлениями страницы.
Проблема была вызвана некорректной индексацией, а также отсутствием кода для применения переводов ресурсов к кнопкам выбора списка записей.
Исправление внесено путём изменения индексации переводов ресурсов, которая теперь начинается с `0` вместо `1`, а также добавления недостающего кода для корректного применения переводов ресурсов к кнопкам выбора списка записей.
([`34827c5`](https://github.com/lowcode/lowcode-webapp-compose/commit/34827c5), [`04eb3cd`](https://github.com/lowcode/lowcode-server/commit/04eb3cd)).


# `2022.3.4`

**Релиз состоялся**: `2022-07-27`

.[#2022*3_4-added]#[2022*3*4-added,Добавлено:](#2022*3_4-added,Добавлено:)*
- Добавлена новая опция в список записей, которая принудительно открывает просматриваемые записи в новой вкладке вместо текущего окна.
Функция добавлена для устранения необходимости переходить назад и вперёд, повторно применять фильтры при работе с несколькими записями из списка ([`25e7090`](https://github.com/lowcode/lowcode-locale/commit/25e7090), [`f7191e5`](https://github.com/lowcode/lowcode-webapp-compose/commit/f7191e5), [`98fc48e`](https://github.com/lowcode/lowcode-js/commit/98fc48e)).
- Добавлена новая опция «Ссылка на родителя» для поля модуля `Record`, определяющая, будет ли значение ввода предзаполнено ID родительской записи.
Функция добавлена для случаев, когда не нужно предзаполнять ввод, например, для модуля с полем записи, указывающим на себя ([`059ebf8`](https://github.com/lowcode/lowcode-locale/commit/059ebf8), [`a79a852`](https://github.com/lowcode/lowcode-webapp-compose/commit/a79a852), [`0e9de63`](https://github.com/lowcode/lowcode-js/commit/0e9de63)).
- Добавлена поддержка переводов ресурсов для воронкообразных и шкальных диаграмм Low Code.
Функция позволяет переводить метки шагов воронкообразных и шкальных диаграмм ([`9f61346`](https://github.com/lowcode/lowcode-webapp-compose/commit/9f61346)).
- Добавлен заголовок `Content-Type: text/javascript` к файлу `config.js`, обслуживаемому сервером LowCoooode.
Заголовки были добавлены, потому что при предотвращении [MIME-сниффинга](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics*of*HTTP/MIME*types#mime*sniffing) веб-приложения переставали работать ([`1bdf2b5`](https://github.com/lowcode/lowcode-server/commit/1bdf2b5)).
- Добавлена новая опция `.env` `AUTH*JWT*ALGORITHM`, позволяющая настроить алгоритм для JWT-токенов ([`480e70e`](https://github.com/lowcode/lowcode-server/commit/480e70e), [`8e42126`](https://github.com/lowcode/lowcode-server/commit/8e42126)).

.[#2022*3_4-changed]#[2022*3*4-changed,Изменено:](#2022*3_4-changed,Изменено:)*
- Список записей теперь ожидает завершения фильтрации перед разрешением дальнейшей фильтрации.
Изменение сделано из-за непредвиденных случаев, когда фильтр A срабатывал после фильтра B, и список записей показывал данные от фильтра A вместо фильтра B.
Список записей теперь ожидает завершения ввода запроса перед получением данных и блокировкой ввода ([`dea2d99`](https://github.com/lowcode/lowcode-webapp-compose/commit/dea2d99)).
- Выражения полей модуля теперь очищают значение записи, когда выражение возвращает значение `null`.
Раньше поле показывало нулевое значение для этого типа, например `0` для поля `Number`.
Изменение сделано, чтобы выражения полей могли очищать свои значения, что полезно для некоторых специфических случаев ([`0bef905`](https://github.com/lowcode/lowcode-server/commit/0bef905)).
- Изменена политика сброса пароля: теперь пользователям запрещено повторно использовать старые пароли.
Раньше пользователи могли сбросить текущий пароль на тот же, что использовали ранее, но теперь такое изменение запрещено и приводит к неудаче сброса пароля ([`b44024f`](https://github.com/lowcode/lowcode-server/commit/b44024f)).
- Общие изменения UI/UX: изменены размеры кнопок перевода ресурсов диаграмм ([`255317a`](https://github.com/lowcode/lowcode-webapp-compose/commit/255317a)), удалена консольная ошибка о неопределённом модуле при создании диаграммы ([`482d971`](https://github.com/lowcode/lowcode-webapp-compose/commit/482d971)), исправлен разделитель множественных значений поля записи ([`e7249fb`](https://github.com/lowcode/lowcode-webapp-compose/commit/e7249fb)).

.[#2022*3_4-removed]#[2022*3*4-removed,Удалено:](#2022*3_4-removed,Удалено:)*
- Поля `Boolean` и `Record` удалены из поиска записей LowCoooode Low Code с использованием поля ввода поиска.
Поле `Boolean` удалено, потому что оценка true/false создавала слишком свободные фильтры, приводя к неожиданным результатам.
Поле `Record` удалено, потому что значения ищутся по идентификаторам, что вызывало путаницу в некоторых случаях.
Вы можете использовать фильтры по конкретным полям для этих двух случаев ([`dea2d99`](https://github.com/lowcode/lowcode-webapp-compose/commit/dea2d99)).

!!! important
    API-фильтрация по-прежнему поддерживает тот же набор полей, что и ранее.
    Исключение было специфичным для списков записей Low Code.


- Кнопка btn:[delete] удалена из пользовательского интерфейса создания шаблона в веб-приложении Admin.
Кнопка удалена, поскольку нет необходимости в её присутствии: нельзя удалить ещё не созданный шаблон ([`b72bdd0`](https://github.com/lowcode/lowcode-webapp-admin/commit/b72bdd0)).


.[#2022*3_4-fixed]#[2022*3*4-fixed,Исправлено:](#2022*3_4-fixed,Исправлено:)*
- Исправлено отбрасывание изменений в модальном окне редактора выражений рабочего процесса при сохранении с помощью сочетания клавиш kbd:[cmd+s].
Сочетание kbd:[cmd+s] теперь отключено, когда открыто модальное окно редактора выражений ([`0803d25`](https://github.com/lowcode/lowcode-webapp-workflow/commit/0803d25)).
- Исправлено непропадание уведомления об изменении пути триггера при отмене с помощью сочетания kbd:[cmd+z].
Логика изменения пути триггера теперь корректно обрабатывает сочетания kbd:[cmd+z] и kbd:[cmd+y] ([`a43e106`](https://github.com/lowcode/lowcode-webapp-workflow/commit/a43e106)).
- Исправлена некорректная обработка полей модуля с множественными значениями в таблицах LowCoooode Reporter.
Раньше поле с множественными значениями дублировало строку, но теперь оно отображается как одна строка, где значения используют тот же разделитель множественных значений, что настроен в поле модуля ([`d65a767`](https://github.com/lowcode/lowcode-server/commit/d65a767)).
- Исправлена проблема предоставления правил RBAC, когда миграция правил пыталась создать дублирующуюся запись правила RBAC.
Шаг миграции теперь пропускает любые правила RBAC, которые могут вызвать ошибку дублирования при миграции ([`1f657b3`](https://github.com/lowcode/lowcode-server/commit/1f657b3)).
- Исправлена нарушенная обратная совместимость для функции удаления записей compose рабочего процесса.
Предыдущее исправление ошибки добавило дополнительные параметры поиска, которые были установлены как обязательные, из-за чего старый сценарий, где предоставлялась фактическая запись Compose, перестал работать.
Параметры пространства имён и модуля функции удаления записей compose больше не являются обязательными, но применяются во внутренней логике ([`378d0f2`](https://github.com/lowcode/lowcode-server/commit/378d0f2)).
- Исправлена некорректная валидация ограничений пароля, где поля ввода некорректно проверяли некоторые граничные случаи.
Минимальная длина пароля теперь ограничена минимум 8 символами, а другие свойства больше не допускают отрицательных значений ([`192bc08`](https://github.com/lowcode/lowcode-webapp-admin/commit/192bc08), [`e0f912b`](https://github.com/lowcode/lowcode-locale/commit/e0f912b), [`9b424a6`](https://github.com/lowcode/lowcode-server/commit/9b424a6)).
- Исправлена некорректная адаптивность блочного элемента страницы календаря Low Code, из-за которой он выглядел обрезанным при добавлении.
Ошибка исправлена ручной настройкой элементов календаря при изменениях макета ([`cf66f22`](https://github.com/lowcode/lowcode-webapp-compose/commit/cf66f22)).
- Исправлено неотображение данных в диаграммах Low Code при использовании переменных предфильтра `${record}` или `${recordID}`.
Проблема была вызвана тем, что записи неправильно разрешались и не были доступны вовремя ([`ec33de1`](https://github.com/lowcode/lowcode-webapp-compose/commit/ec33de1)).
- Исправлены опции поля `Record` при использовании вложенного поля `Record`, где данные дублировались.
Исправление теперь корректно разрешает дублированные записи и правильно отображает метки опций ([`b954f59`](https://github.com/lowcode/lowcode-webapp-compose/commit/b954f59)).
- Исправлено отображение списка записей, когда показывается только одна колонка.
Проблема решена путём настройки колонок для «метаданных» (флажок и кнопки действий) ([`2671c5d`](https://github.com/lowcode/lowcode-webapp-compose/commit/2671c5d)).
- Исправлен предпросмотр страницы записи при доступе без записи.
Страница больше не загружается бесконечно и, как правило, предотвращает доступ, если записи нет ([`21c4f04`](https://github.com/lowcode/lowcode-webapp-compose/commit/21c4f04), [`924934e`](https://github.com/lowcode/lowcode-webapp-compose/commit/924934e)).
- Исправлена визуальная ошибка с полями `Record` с множественными значениями, когда последняя выбранная запись отображалась как заполнитель, даже если она была удалена из списка выбранных записей.
Исправление теперь отказывается от использования последней выбранной записи в качестве заполнителя и использует общую метку ([`491bc19`](https://github.com/lowcode/lowcode-webapp-compose/commit/491bc19)).
- Исправлена ошибка отображения для таблиц отчётов, когда объединённый источник данных не предоставлял данных.
Столбец таблицы теперь корректно обрабатывает эти граничные случаи и строит таблицу соответствующим образом ([`953556a`](https://github.com/lowcode/lowcode-webapp-reporter/commit/953556a)).
- Исправлена ошибка при попытке отметить роль как контекстную.
Ошибка была вызвана неверным внутренним состоянием и теперь устранена ([`b14e4e7`](https://github.com/lowcode/lowcode-webapp-admin/commit/b14e4e7)).
- Исправлена неудачная попытка создания индекса для LowCoooode Discovery, вызванная неподдерживаемым типом данных.
Исправление меняет используемый тип данных для случаев, которые приводили к ошибке индекса.
Ошибка была вызвана индексами идентификаторов, такими как ID модуля, ID пространства имён и ID записи ([`25ebf75`](https://github.com/lowcode/lowcode-server/commit/25ebf75)).
- Исправлено некорректное время, выдаваемое полем `Date Time`, когда были выбраны опции времени, прошлого/будущего и относительного вывода.
Логика даты и времени теперь корректно обрабатывает эту комбинацию опций и правильно обрабатывает значение ([`a4bbbab`](https://github.com/lowcode/lowcode-js/commit/a4bbbab)).
- Исправлены некорректные переводы ресурсов для диаграмм Low Code при поиске конкретных диаграмм ([`b8ad97e`](https://github.com/lowcode/lowcode-server/commit/b8ad97e)), полей модуля при поиске конкретных модулей ([`5ccd682`](https://github.com/lowcode/lowcode-server/commit/5ccd682)).
- Исправлена некорректная оценка предфильтра списка записей и диаграммы, из-за которой предфильтр с `${recordID}` не показывал данные при создании новых записей ([`6da48a0`](https://github.com/lowcode/lowcode-webapp-compose/commit/6da48a0)).
- Исправлен неверный счётчик при экспорте записей из списка записей ([`d837e9e`](https://github.com/lowcode/lowcode-webapp-compose/commit/d837e9e)).

.[#2022*3_4-development]#[2022*3*4-development,Разработка:](#2022*3_4-development,Разработка:)*
- Добавлено действие сборки SonarQube GitHub в сервер LowCoooode ([`5999c70`](https://github.com/lowcode/lowcode-server/commit/5999c70)).


# `2022.3.3`

**Релиз состоялся**: `2022-06-15`

.[#2022*3_3-added]#[2022*3*3-added,Добавлено:](#2022*3_3-added,Добавлено:)*
- Добавлена поддержка запросов рабочего процесса в веб-приложениях LowCoooode One и Admin (Low Code уже поддерживает запросы).
Некоторые веб-приложения не отображают определённые типы запросов, такие как «перенаправить на страницу записи» ([`#76e4410`](https://github.com/lowcode/lowcode-vue/commit/76e4410), [`#c74fe70`](https://github.com/lowcode/lowcode-vue/commit/c74fe70), [`#f95d26f`](https://github.com/lowcode/lowcode-js/commit/f95d26f), [`6e835bc`](https://github.com/lowcode/lowcode-webapp-one/commit/6e835bc), [`#c62f9fe`](https://github.com/lowcode/lowcode-webapp-compose/commit/c62f9fe), [`#2909ca3`](https://github.com/lowcode/lowcode-webapp-admin/commit/2909ca3)).
- Добавлен макет печати для поддержки функций браузера «Печать в PDF».
Эта функция помогает создавать более качественные PDF-документы без ненужной UI-навигации и других элементов управления ресурсами ([`#ed8ee8b`](https://github.com/lowcode/lowcode-webapp-compose/commit/ed8ee8b), [`#d980c0b`](https://github.com/lowcode/lowcode-webapp-admin/commit/d980c0b), [`#c71e1f0`](https://github.com/lowcode/lowcode-webapp-reporter/commit/c71e1f0), [`#f005b62`](https://github.com/lowcode/lowcode-webapp-workflow/commit/f005b62), [`d582560`](https://github.com/lowcode/lowcode-webapp-one/commit/d582560)).
- Добавлена поддержка предоставления значений системных настроек JSON при выполнении шага предоставления.
Поддержка добавлена для возможности предоставления YAML-файлов для настройки видимости кнопок панели инструментов записей ([`#05b97ee`](https://github.com/lowcode/lowcode-server/commit/05b97ee)).
Примером может быть `compose.ui.record-toolbar: { "hideBack": true }`.

.[#2022*3_3-changed]#[2022*3*3-changed,Изменено:](#2022*3_3-changed,Изменено:)*
- Улучшен UX пользовательского интерфейса Low Code: добавлены спиннеры для индикации загрузки данных ([`#11c7cf8`](https://github.com/lowcode/lowcode-webapp-compose/commit/11c7cf8)), изменён курсор событий календаря на указатель ([`#75cc346`}), добавлены дополнительные исключения для сортировки и фильтрации списка записей во избежание странного поведения (https://github.com/lowcode/lowcode-js/commit/70c6277[`#70c6277`](https://github.com/lowcode/lowcode-webapp-compose/commit/75cc346), [`#6826d19`](https://github.com/lowcode/lowcode-webapp-compose/commit/6826d19)), добавлена небольшая задержка при обработке полей Record ([`#b6e722d`](https://github.com/lowcode/lowcode-webapp-compose/commit/b6e722d)).
- Шрифт диаграмм изменён на Poppins-Regular для согласованности с остальным дизайном LowCoooode ([`#0349363`](https://github.com/lowcode/lowcode-js/commit/0349363)).
- Изменена логика сеанса аутентификации пользователя и токена доступа, а также обновлены соответствующие переменные `.env` и их описания ([`#7f84994`](https://github.com/lowcode/lowcode-server/commit/7f84994)).
Изменено: сеанс аутентификации и соответствующая cookie истекают при завершении сеанса при использовании «входа», и после указанного времени жизни при использовании «входа и запоминания» ([`#5fd5b93`](https://github.com/lowcode/lowcode-server/commit/5fd5b93)).
Изменено: кнопка «вход и запомнить меня» не отображается, если переменная `.env` `AUTH*SESSION*PERM_LIFETIME` установлена в 0 (отключение функции) ([`#ad53ea9`](https://github.com/lowcode/lowcode-server/commit/ad53ea9)).
- Боковая панель LowCoooode Low Code изменена: ссылка на все пространства имён теперь отображается в верхней части выпадающего списка.
Это изменение позволяет видеть ссылку даже при большом количестве пространств имён ([`#15ca4000`](https://github.com/lowcode/lowcode-webapp-compose/commit/15ca4000)).
- Изменена и стандартизирована валидация значений ввода имени и handle во всех веб-приложениях.
Изменение сделано для обеспечения более согласованного пользовательского опыта в веб-приложениях LowCoooode.
Поля ввода теперь корректно отображают, является ли предоставленное значение допустимым именем/handle или нет ([`#e3bbec7`](https://github.com/lowcode/lowcode-webapp-reporter/commit/e3bbec7), [`#cae73df`](https://github.com/lowcode/lowcode-webapp-workflow/commit/cae73df), [`#1b1d165`](https://github.com/lowcode/lowcode-webapp-admin/commit/1b1d165)).
- Типы полей модуля теперь сортируются по алфавиту при редактировании модуля.
Изменение сделано для обеспечения более согласованного пользовательского опыта при настройке модулей Low Code ([`#8378e73`](https://github.com/lowcode/lowcode-webapp-compose/commit/8378e73)).

.[#2022*3_3-removed]#[2022*3*3-removed,Удалено:](#2022*3_3-removed,Удалено:)*
- Удалена переменная `.env` `AUTH*JWT*EXPIRY`, так как она заменена переменной `.env` `AUTH*OAUTH2*ACCESS*TOKEN*LIFETIME` ([`#e626bcd`](https://github.com/lowcode/lowcode-server/commit/e626bcd)).

.[#2022*3_3-fixed]#[2022*3*3-fixed,Исправлено:](#2022*3_3-fixed,Исправлено:)*
- Исправлен неработающий экспорт модуля Low Code, из-за которого поля `Record` становились ненастраиваемыми после импорта модуля ([`#1219112`](https://github.com/lowcode/lowcode-webapp-compose/commit/1219112)).
- Исправлена неверная миграция правил RBAC, когда правило могло использовать подстановочные знаки между ID ресурсов (недопустимое состояние).
Новый шаг миграции пытается исправить повреждённые правила RBAC ([`563a73c6`](https://github.com/lowcode/lowcode-server/commit/563a73c6)).
- Исправлена некорректная обработка запросов рабочего процесса, когда у пользователя открыто несколько экземпляров LowCoooode (вкладок) одновременно.
Веб-приложения LowCoooode улучшили коммуникацию о том, какие запросы уже разрешены и где должны отображаться запросы ([`#40e5416`](https://github.com/lowcode/lowcode-vue/commit/40e5416)).
- Исправлено некорректное декодирование значений числовых и текстовых настроек путём обработки граничных случаев ([`#ced2daf`](https://github.com/lowcode/lowcode-server/commit/ced2daf)).
- Исправлено отклонение некорректных файлов JSONL при импорте записей: добавлены дополнительные проверки content-type в логику предварительной обработки ([`#f726d3f`](https://github.com/lowcode/lowcode-server/commit/f726d3f)).
- Исправлена некорректная миграция правил контроля доступа к записям из-за неверной пагинации ([`#b6d13d9`](https://github.com/lowcode/lowcode-server/commit/b6d13d9)).
- Исправлены бесконечные предупреждения веб-консоли при отсутствии соединения ([`#d8e57b1`](https://github.com/lowcode/lowcode-server/commit/d8e57b1)).
- Исправлена некорректная генерация узла URI федерации: добавлен базовый URL API ([`#798c31e`](https://github.com/lowcode/lowcode-server/commit/798c31e)).
- Исправлена паника проверки здоровья системы, когда соединение Corredor недоступно ([`#2ff1108`](https://github.com/lowcode/lowcode-server/commit/2ff1108)).
- Исправлены неработающие выражения значений полей модуля, где использовалась ссылка на `old` запись.
Проблема возникала, потому что значение никогда не предоставлялось ([`#7135903`](https://github.com/lowcode/lowcode-server/commit/7135903)).
- Исправлена некорректная валидация учётных данных при регистрации пользователя, из-за которой пользователь создавался даже при неверных учётных данных (в основном пароле).
Исправление перемещает этап проверки пароля на более раннюю стадию процесса регистрации ([`#c1500df`](https://github.com/lowcode/lowcode-server/commit/c1500df)).
- Исправлен некорректный подсчёт записей для сгруппированных отчётов с участием полей с множественными значениями.
Исправление в некоторых случаях опускает дублирующиеся значения ([`#416a7ec`](https://github.com/lowcode/lowcode-server/commit/416a7ec)).
- Исправлен сброс блочных элементов страницы в их позиции по умолчанию при сохранении изменений ([`#efe24c3`](https://github.com/lowcode/lowcode-webapp-compose/commit/efe24c3)).
- Исправлен некорректный разбор содержимого пользовательского интерфейса перевода ресурсов, из-за которого некоторые события вставки очищали содержимое или блокировали редактор ([`#6a195df`](https://github.com/lowcode/lowcode-webapp-compose/commit/6a195df)).
- Исправлено и обеспечено корректное отображение разделителей множественных значений при просмотре значения записи ([`#a211af6`](https://github.com/lowcode/lowcode-webapp-compose/commit/a211af6)).
- Исправлены дублирующиеся переводы ресурсов при экспорте пространства имён Low Code.
Исправление применяет дополнительную предварительную обработку и валидацию перед выводом переводов ресурсов в архив ([`#dcef397`](https://github.com/lowcode/lowcode-webapp-compose/commit/dcef397)).
- Исправлено неотображение сообщений об ошибках пользовательской валидации полей на основе языка пользователя.
Валидатор значений теперь использует возможности i18n для предоставления соответствующих сообщений об ошибках ([`#f9e9433`](https://github.com/lowcode/lowcode-server/commit/f9e9433)).
- Исправлена некорректная процедура миграции исправления правил RBAC, вызванная неограниченной выборкой данных, перегружающей базу данных.
Процедура теперь постепенно выбирает необходимые данные, избегая проблемы ([`#f310442`](https://github.com/lowcode/lowcode-server/commit/f310442)).
- Исправлена невозможность отклонения неудачных запросов рабочего процесса ([`#ff3d0e6`](https://github.com/lowcode/lowcode-vue/commit/ff3d0e6)).

.[#2022*3_3-development]#[2022*3*3-development,Разработка:](#2022*3_3-development,Разработка:)*
- Включение бандла веб-консоли в собираемый lowcode-server ([`#54fffd0`](https://github.com/lowcode/lowcode-server/commit/54fffd0)).
- Определена новая структура утилит ресурсов, которая помогает улучшить производительность операций за счёт более интеллектуальных стратегий индексации.
Структура применена к текущему сервису контроля доступа RBAC, экспорту правил RBAC и сервисам экспорта переводов ресурсов ([`#0466ffe`](https://github.com/lowcode/lowcode-server/commit/0466ffe), [`#e7c1fe1`](https://github.com/lowcode/lowcode-server/commit/e7c1fe1), [`#a68ddf1`](https://github.com/lowcode/lowcode-server/commit/a68ddf1), [`#aef3171`](https://github.com/lowcode/lowcode-server/commit/aef3171)).


# `2022.3.2`

**Релиз состоялся**: `2022-05-18`

.[#2022*3_2-added]#[2022*3*2-added,Добавлено:](#2022*3_2-added,Добавлено:)*
- Добавлена проверка на пустое значение для валидации имени поля модуля ([`d3c33a6`](https://github.com/lowcode/lowcode-webapp-compose/commit/d3c33a6)).
- Добавлены более описательные заголовки вкладок в зависимости от текущей страницы пользователя ([`d6bb352`](https://github.com/lowcode/lowcode-locale/commit/d6bb352), [`a7f36e5`](https://github.com/lowcode/lowcode-webapp-compose/commit/a7f36e5)).
- Добавлено сохранение фильтров списка записей в локальном хранилище браузера ([`52ff728`](https://github.com/lowcode/lowcode-webapp-compose/commit/52ff728), [`bc2dd98`](https://github.com/lowcode/lowcode-locale/commit/bc2dd98)).
Обратите внимание, что сохранённые фильтры доступны только в браузере, где они были сохранены.

.[#2022*3_2-changed]#[2022*3*2-changed,Изменено:](#2022*3_2-changed,Изменено:)*
- Изменены статические переводы для модулей Low Code ([`5efb7c9`](https://github.com/lowcode/lowcode-locale/commit/5efb7c9)) и фильтров шлюза интеграции ([`98f8e4b`](https://github.com/lowcode/lowcode-locale/commit/98f8e4b)).
- Состояния ввода конфигурации ресурсов изменены для большей согласованности ([`2ef02c5`](https://github.com/lowcode/lowcode-webapp-compose/commit/2ef02c5)).
- Шаблон формата фильтра даты и времени в списке записей изменён: теперь игнорируются секунды ([`fc2eab1`](https://github.com/lowcode/lowcode-webapp-compose/commit/fc2eab1), [`07c409e`](https://github.com/lowcode/lowcode-server/commit/07c409e)).
- Определения хранилища журнала активности ресурсов стали более строгими и менее гибкими ([`e7ed1d8`](https://github.com/lowcode/lowcode-server/commit/e7ed1d8)).
- Поля записей модуля теперь показывают понятную метку вложенного поля пользователя вместо ID пользователя ([`0d4d74b`](https://github.com/lowcode/lowcode-webapp-compose/commit/0d4d74b)).
- Конфигурация диаграммы изменена: кнопка «Добавить метрику» скрыта, когда не выбран ни один модуль ([`a3177bc`](https://github.com/lowcode/lowcode-webapp-compose/commit/a3177bc)).

.[#2022*3_2-fixed]#[2022*3*2-fixed,Исправлено:](#2022*3_2-fixed,Исправлено:)*
- Исправлен поиск очереди в веб-приложении Admin ([`98a5d95`](https://github.com/lowcode/lowcode-webapp-admin/commit/98a5d95), [`e6a0b0f`](https://github.com/lowcode/lowcode-server/commit/e6a0b0f)).
- Исправлен поиск маршрута шлюза интеграции в веб-приложении Admin ([`f821dd8`](https://github.com/lowcode/lowcode-webapp-admin/commit/f821dd8), [`5ddddf8`](https://github.com/lowcode/lowcode-server/commit/5ddddf8)).
- Исправлена неработающая миграция переводов ресурсов, когда связанный ресурс был удалён ([`1786eda`](https://github.com/lowcode/lowcode-server/commit/1786eda)).
- Исправлена некорректная обработка ошибок в конфигурации диаграммы ([`d4bf472`](https://github.com/lowcode/lowcode-webapp-compose/commit/d4bf472)) и некорректная обработка состояния при создании новой диаграммы ([`41b0a15`](https://github.com/lowcode/lowcode-webapp-compose/commit/41b0a15)).
- Исправлена некорректная обработка переводов ресурсов для свойств подсказки и описания поля модуля Low Code.
Эти два свойства были перепутаны ([`q7fbaf94`](https://github.com/lowcode/lowcode-server/commit/q7fbaf94), [`1c6b793`](https://github.com/lowcode/lowcode-server/commit/1c6b793)), некорректно удалялись из переводов ресурсов при удалении во всплывающем окне конфигурации поля ([`d4c9243`](https://github.com/lowcode/lowcode-server/commit/d4c9243)) и не отображались в веб-приложении ([`18de72f`](https://github.com/lowcode/lowcode-webapp-compose/commit/18de72f)).
- Исправлена ошибка загрузки диаграмм в случаях некорректного управления состоянием ([`629d306`](https://github.com/lowcode/lowcode-webapp-compose/commit/629d306)).
- Исправлен сброс фильтра списка записей при удалении строк фильтра ([`00b491d`](https://github.com/lowcode/lowcode-webapp-compose/commit/00b491d)).
- Исправлена некорректная загрузка сценариев аутентификации при работе в режиме разработки ([`d280fc7`](https://github.com/lowcode/lowcode-server/commit/d280fc7)).
- Исправлен некорректный срок действия токена обновления.
При выдаче новых токенов обновления не корректировались временные метки истечения ([`e7d3df1`](https://github.com/lowcode/lowcode-server/commit/e7d3df1)).
- Исправлена ошибка загрузки новых страниц записей Low Code из-за отсутствия определений кнопок ([`b135287`](https://github.com/lowcode/lowcode-server/commit/b135287)).
- Исправлено дублирование элементов в списках выбора (в основном полей модуля) ([`9a33368`](https://github.com/lowcode/lowcode-vue/commit/9a33368)).
- Исправлен сбой системы при попытке обмена недействительных учётных данных: теперь пользователь перенаправляется на страницу входа ([`0c49832e`](https://github.com/lowcode/lowcode-server/commit/0c49832e)).
- Исправлена некорректная обработка ошибки соединения `peer-closed`, которая без необходимости засоряла системы отслеживания ошибок ([`ab248fe1`](https://github.com/lowcode/lowcode-server/commit/ab248fe1)).


# `2022.3.1`

**Релиз состоялся**: `2022-04-20`

.[#2022*3_1-added]#[2022*3*1-added,Добавлено:](#2022*3_1-added,Добавлено:)*
- Добавлены удалённые ресурсы в индексацию и поиск Discovery ([`09d69f124`](https://github.com/lowcode/lowcode-server-discovery/commit/09d69f124)).
- Добавлена опция длительности JWT-токена в команду CLI ([`679af2f55`](https://github.com/lowcode/lowcode-server/commit/679af2f55)).
- Добавлены дополнительные опции конфигурации поля `Geometry`, такие как уровень масштабирования по умолчанию ([`a8534ed`](https://github.com/lowcode/lowcode-vue/commit/a8534ed), [`b7ab3e47`](https://github.com/lowcode/lowcode-webapp-compose/commit/b7ab3e47), [`97d1aed3`](https://github.com/lowcode/lowcode-webapp-compose/commit/97d1aed3), [`98cf01b8`](https://github.com/lowcode/lowcode-webapp-compose/commit/98cf01b8)).
- Добавлен индикатор обработки в организатор записей для улучшения UX ([`0e85fbca`](https://github.com/lowcode/lowcode-webapp-compose/commit/0e85fbca)).
- Добавлена поддержка перевода системных полей модулей Low Code ([`623adaf3`](https://github.com/lowcode/lowcode-webapp-compose/commit/623adaf3)).
- Добавлено постоянное хранение фильтров списка записей, позволяющее повторно использовать ранее использованные фильтры ([`e18267fd`](https://github.com/lowcode/lowcode-webapp-compose/commit/e18267fd)).
- Добавлено веб-приложение Discovery в make-команду `make webapp`.
- Добавлена пагинация ответов поиска для улучшения производительности веб-приложений при больших ответах ([`e046f12`](https://github.com/lowcode/lowcode-server-discovery/commit/e046f12), [`7cb2d64`](https://github.com/lowcode/lowcode-webapp-discovery/commit/7cb2d64)).

.[#2022*3_1-changed]#[2022*3*1-changed,Изменено:](#2022*3_1-changed,Изменено:)*
- Общие улучшения UI/UX в LowCoooode Low Code, LowCoooode Admin, LowCoooode One и LowCoooode Workflow.
- Удаление Low Code теперь также удаляет связанную страницу записи ([`0ceade66`](https://github.com/lowcode/lowcode-webapp-compose/commit/0ceade66)).

.[#2022*3_1-fixed]#[2022*3*1-fixed,Исправлено:](#2022*3_1-fixed,Исправлено:)*
- Исправлена загрузка файлов CSV, когда сервер некорректно определял MimeType ([`195c2bb13`](https://github.com/lowcode/lowcode-server/commit/195c2bb13)).
- Исправлено некорректное управление итератором ресурсов рабочего процесса, когда большие наборы данных приводили к сбою из-за неправильной реализации ограничения ресурсов ([`0792c0a17`](https://github.com/lowcode/lowcode-server/commit/0792c0a17)).
- Исправлена некорректная трассировка выполнения рабочего процесса, когда логика игнорировала параметр трассировки ([`41667a7da`](https://github.com/lowcode/lowcode-server/commit/41667a7da)).
- Исправлены ошибки гонки данных для pkg/locale ([`345050990`](https://github.com/lowcode/lowcode-server/commit/345050990)), pkg/corredor healthcheck ([`a797c847b`](https://github.com/lowcode/lowcode-server/commit/a797c847b)), pkg/healtcheck ([`662f5155b`](https://github.com/lowcode/lowcode-server/commit/662f5155b)), WebSockets ([`e8cd7d37d`](https://github.com/lowcode/lowcode-server/commit/e8cd7d37d)) и pkg/scheduler ([`4a75778c1`](https://github.com/lowcode/lowcode-server/commit/4a75778c1)).
- Исправлены функции выражений преобразования времени, выдававшие ошибки при использовании корректных строковых значений времени ([`5b40f7875`](https://github.com/lowcode/lowcode-server/commit/5b40f7875)).
- Исправлены утечки памяти, вызывающие деградацию производительности на WebSockets ([`c64116fc8`](https://github.com/lowcode/lowcode-server/commit/c64116fc8)), санации содержимого ([`9346b5702`](https://github.com/lowcode/lowcode-server/commit/9346b5702)), загрузке рабочих процессов ([`fa614c7ac`](https://github.com/lowcode/lowcode-server/commit/fa614c7ac)) и обработке выражений ([`adee67f6b`](https://github.com/lowcode/lowcode-server/commit/adee67f6b)).
- Исправлена загрузка вложений рабочего процесса для всех поддерживаемых типов (`Reader`, `Bytes` и `String`) ([`6fd2288df`](https://github.com/lowcode/lowcode-server/commit/6fd2288df)).
- Исправлено дублирование запросов путём предотвращения повторной отправки запросов через веб-сокеты ([`9c0026462`](https://github.com/lowcode/lowcode-server/commit/)9c0026462).
- Исправлена подготовка заголовка базовой аутентификации функции HTTP-запроса рабочего процесса ([`2898e1b8c`](https://github.com/lowcode/lowcode-server/commit/2898e1b8c)).
- Исправлена адаптивность `CInputDateTime` ([`b0c6422`](https://github.com/lowcode/lowcode-vue/commit/b0c6422)).
- Исправлено поведение CSidebarNavItems ([`7c7b708`](https://github.com/lowcode/lowcode-vue/commit/7c7b708)).
- Исправлены отсутствующие элементы отображения блочного элемента страницы отчёта ([`1625d1fc`](https://github.com/lowcode/lowcode-webapp-compose/commit/1625d1fc)).
- Исправлена ошибка загрузки Low Code из-за присваивания параметра только для чтения ([`ce5cd504`](https://github.com/lowcode/lowcode-webapp-compose/commit/ce5cd504)).
- Исправлены переводы системных полей в блочных элементах страниц записей ([`fa7e6808`](https://github.com/lowcode/lowcode-webapp-compose/commit/fa7e6808)).
- Исправлено переполнение уведомления списка записей, указывающего на отсутствие записей ([`12b9fc7f`](https://github.com/lowcode/lowcode-webapp-compose/commit/12b9fc7f)).
- Исправлено сохранение встроенного списка записей ([`f54452c7`](https://github.com/lowcode/lowcode-webapp-compose/commit/f54452c7)).
- Исправлено отсутствие системных полей в блочных элементах страниц записей ([`de874a8b`](https://github.com/lowcode/lowcode-webapp-compose/commit/de874a8b)).
- Исправлено переполнение нижнего колонтитула списка записей за пределы блочного элемента ([`9cb2d923`](https://github.com/lowcode/lowcode-webapp-compose/commit/9cb2d923)).
- Исправлена ссылка на документацию по выражениям полей ([`301a1de8`](https://github.com/lowcode/lowcode-webapp-compose/commit/301a1de8)).
- Исправлено переполнение экрана выпадающими списками ([`8e37e2b8`](https://github.com/lowcode/lowcode-webapp-compose/commit/8e37e2b8)).
- Исправлена индексация ленты для удалённых ресурсов ([`f5cfb6c`](https://github.com/lowcode/lowcode-server-discovery/commit/f5cfb6c)).

.[#2022*3_1-development]#[2022*3*1-development,Разработка:](#2022*3_1-development,Разработка:)*
- Улучшено E2E-тестирование: добавлены теги `data-test-id` к определённым UI-компонентам ([`56af630c0`](https://github.com/lowcode/lowcode-server/commit/56af630c0), [`3bb8fe59`](https://github.com/lowcode/lowcode-webapp-compose/commit/3bb8fe59)).


# `2022.3.0`

**Релиз состоялся**: 2022-03-30

.[#2022*3_0-added]#[2022*3*0-added,Добавлено:](#2022*3_0-added,Добавлено:)*
- Добавлены дополнительные инструменты разработки для улучшения пользовательского опыта: предпросмотр конфигурации шага при наведении ([`b017702`](https://github.com/lowcode/lowcode-webapp-workflow/commit/b017702) [`6e2a4be`](https://github.com/lowcode/lowcode-webapp-workflow/commit/6e2a4be) [`51d1f9e`](https://github.com/lowcode/lowcode-webapp-workflow/commit/51d1f9e)), дополнительные справочные метки ([`214c973`](https://github.com/lowcode/lowcode-webapp-workflow/commit/214c973)) и предопределённый выбор HTTP-метода ([`5e59bbab2`](https://github.com/lowcode/lowcode-server/commit/5e59bbab2)).
- Добавлены дополнительные запросы рабочего процесса для перенаправления на страницу создания записи ([`a021552`](https://github.com/lowcode/lowcode-vue/commit/a021552) [`c1fe4ff`](https://github.com/lowcode/lowcode-vue/commit/c1fe4ff)) и для выбора записи ([`aba32e6`](https://github.com/lowcode/lowcode-vue/commit/aba32e6)).
- Добавлена дополнительная поддержка в рабочих процессах для пагинации по записям Low Code ([`#336`](https://github.com/lowcode/lowcode-server/pull/336)).
- Добавлены метки шагов рабочего процесса по умолчанию, если не указана пользовательская метка ([`3854ddd`](https://github.com/lowcode/lowcode-webapp-workflow/commit/3854ddd)).
- Добавлена [веб-консоль](2022.3@devops-guide/troubleshooting/web-console.md) — пользовательский интерфейс для инспекции и устранения проблем на стороне сервера ([`#327`](https://github.com/lowcode/lowcode-server/pull/327), [`7fa5e43d0`](https://github.com/lowcode/lowcode-server/commit/7fa5e43d0), [`f66ab4550`](https://github.com/lowcode/lowcode-server/commit/f66ab4550), [`9eca65595`](https://github.com/lowcode/lowcode-server/commit/9eca65595), [`f291cde93`](https://github.com/lowcode/lowcode-server/commit/f291cde93)).
- Добавлены API-эндпоинты с целевой страницей и страницами 404 ([`a81f35d5a`](https://github.com/lowcode/lowcode-server/commit/a81f35d5a)).
- Добавлена дополнительная конфигурация для настройки страниц ([`b478b3777`](https://github.com/lowcode/lowcode-server/commit/b478b3777)) и общие настройки UI, боковой панели и верхней панели ([`f69751190`](https://github.com/lowcode/lowcode-server/commit/f69751190), [`a2dd9fe5`](https://github.com/lowcode/lowcode-webapp-compose/commit/a2dd9fe5), [`eb583a339`](https://github.com/lowcode/lowcode-server/commit/eb583a339), [`c558ece3b`](https://github.com/lowcode/lowcode-server/commit/c558ece3b), [`2b6026182`](https://github.com/lowcode/lowcode-server/commit/2b6026182), [`bed4f1b`](https://github.com/lowcode/lowcode-webapp-admin/commit/bed4f1b), [`6134d9b`](https://github.com/lowcode/lowcode-webapp-workflow/commit/6134d9b), [`314d1d1`](https://github.com/lowcode/lowcode-webapp-reporter/commit/314d1d1), [`2659085`](https://github.com/lowcode/lowcode-vue/commit/2659085), [`97dea83`](https://github.com/lowcode/lowcode-vue/commit/97dea83), [`e4f2977`](https://github.com/lowcode/lowcode-vue/commit/e4f2977)).
- Добавлена поддержка редактирования страниц записей так же, как и страниц списков ([`8bb77988`](https://github.com/lowcode/lowcode-webapp-compose/commit/8bb77988)).
- Добавлена опция «Открыть в новой вкладке» для ссылок, определённых в редакторах форматированного текста ([`e4fddb3`](https://github.com/lowcode/lowcode-vue/commit/e4fddb3)).
- Добавлены дополнительные возможности импорта/экспорта: импорт и экспорт пользователей в LowCoooode Admin ([`9142c7b5a`](https://github.com/lowcode/lowcode-server/commit/9142c7b5a), [`84f86bb`](https://github.com/lowcode/lowcode-webapp-admin/commit/84f86bb), [`89f4aab`](https://github.com/lowcode/lowcode-webapp-admin/commit/89f4aab), [`2b04d79`](https://github.com/lowcode/lowcode-js/commit/2b04d79)), а также включение контроля доступа и переводов ресурсов в существующий экспорт пространства имён ([`8e679bf97`](https://github.com/lowcode/lowcode-server/commit/8e679bf97)).
- Добавлен [LowCoooode Discovery](2022.3@integrator-guide/discovery/index.md) — поисковая система для ваших данных ([`7bde98697`](https://github.com/lowcode/lowcode-server/commit/7bde98697), [`84f918a1d`](https://github.com/lowcode/lowcode-server/commit/84f918a1d), [`d384db951`](https://github.com/lowcode/lowcode-server/commit/d384db951), [`4e3d5dd00`](https://github.com/lowcode/lowcode-server/commit/4e3d5dd00), [`b6ff0f921`](https://github.com/lowcode/lowcode-server/commit/b6ff0f921), [`ed1122d6`](https://github.com/lowcode/lowcode-webapp-compose/commit/ed1122d6)).
- Добавлена стратегия удаления для страниц Low Code, когда страница имеет вложенные дочерние страницы ([`691481424`](https://github.com/lowcode/lowcode-server/commit/691481424), [`efb45ff8`](https://github.com/lowcode/lowcode-webapp-compose/commit/efb45ff8)).
- Добавлены улучшения поиска и фильтрации ресурсов: поддержка полей только с датой/временем ([`640a90c20`](https://github.com/lowcode/lowcode-server/commit/640a90c20)), а также нечёткий поиск для некоторых ресурсов ([`03a6f7c`](https://github.com/lowcode/lowcode-vue/commit/03a6f7c), [`4f28578e`](https://github.com/lowcode/lowcode-webapp-compose/commit/4f28578e), [`29b71da`](https://github.com/lowcode/lowcode-webapp-workflow/commit/29b71da), [`8500f61`](https://github.com/lowcode/lowcode-vue/commit/8500f61)).
- Добавлена поддержка полей `Record` для запросов на основе вложенного поля `Record`, ограничено **1 уровнем** ([`155e6b12`](https://github.com/lowcode/lowcode-webapp-compose/commit/155e6b12), [`eb4c911`](https://github.com/lowcode/lowcode-js/commit/eb4c911)).
- Добавлена поддержка множественных определений предсортировки в списках записей ([`80f3ad47`](https://github.com/lowcode/lowcode-webapp-compose/commit/80f3ad47)).
- Добавлена поддержка определения пользовательского clientID и области видимости при генерации JWT-токенов из CLI ([`957e70056`](https://github.com/lowcode/lowcode-server/commit/957e70056)).
- Добавлена интеграция между Reporter и Low Code путём внедрения [блочного элемента страницы репортёра](2022.3@integrator-guide/compose-configuration/page-blocks.md#page-block-report) для отображения данных отчётов на страницах Low Code ([`89664b8a`](https://github.com/lowcode/lowcode-webapp-compose/commit/89664b8a), [`7201f70`](https://github.com/lowcode/lowcode-webapp-reporter/commit/7201f70), [`b4667c5`](https://github.com/lowcode/lowcode-js/commit/b4667c5), [`a0b8913`](https://github.com/lowcode/lowcode-vue/commit/a0b8913)).
- Добавлено новое [поле `Geometry`](2022.3@integrator-guide/compose-configuration/field-types.md#field-type-geometry) для хранения геолокации, в первую очередь для использования с Discovery ([`cb3ac6c2`](https://github.com/lowcode/lowcode-webapp-compose/commit/cb3ac6c2), [`3f56b79`](https://github.com/lowcode/lowcode-js/commit/3f56b79), [`2a12fc0`](https://github.com/lowcode/lowcode-js/commit/2a12fc0)).
- Добавлена поддержка скрытия выбранных ролей из списка разрешений ([`5bdc3e9`](https://github.com/lowcode/lowcode-webapp-admin/commit/5bdc3e9)).
- Добавлены дополнительные стили медиа для улучшения вывода PDF ([`908ba673`](https://github.com/lowcode/lowcode-webapp-compose/commit/908ba673)).
- Добавлен профилировщик шлюза интеграции ([`9a7f6f90`](https://github.com/lowcode/lowcode-server/commit/9a7f6f90), [`cbd7ab45`](https://github.com/lowcode/lowcode-webapp-admin/commit/cbd7ab45)).


.[#2022*3_0-changed]#[2022*3*0-changed,Изменено:](#2022*3_0-changed,Изменено:)*
- Шаблоны аутентификации изменены для использования общих ресурсов ([`43ddaf1e5`](https://github.com/lowcode/lowcode-server/commit/43ddaf1e5)).
- Пакеты обновлены до более новых версий, наиболее заметно `jwx goth` и `jwt` ([`6eda39f3a`](https://github.com/lowcode/lowcode-server/commit/6eda39f3a)), а также переход на `go 1.17` ([`2d90fe4e9`](https://github.com/lowcode/lowcode-server/commit/2d90fe4e9)).
- Триггеры рабочего процесса `onTimestamp` изменены для использования нового компонента даты/времени ([`e18230b`](https://github.com/lowcode/lowcode-webapp-workflow/commit/e18230b)).
- Шаг выражения рабочего процесса изменён: кнопка добавления теперь отображается внизу списка, а не вверху ([`b16cb39`](https://github.com/lowcode/lowcode-webapp-workflow/commit/b16cb39)).
- Изменена боковая панель конфигуратора шагов рабочего процесса для использования более быстрых переходов для улучшения пользовательского опыта ([`642d3e2`](https://github.com/lowcode/lowcode-webapp-workflow/commit/642d3e2)).
- Шаги ошибки и задержки рабочего процесса изменены для использования выражений вместо константных значений ([`5754544`](https://github.com/lowcode/lowcode-webapp-workflow/commit/5754544), [`9ad29f7`](https://github.com/lowcode/lowcode-webapp-workflow/commit/9ad29f7)).
- Отчёты изменены для отображения пользователей с удобным идентификатором (имя, email и т.д.) вместо системного ID пользователя ([`506a92d2c`](https://github.com/lowcode/lowcode-server/commit/506a92d2c), [`ac6e7332a`](https://github.com/lowcode/lowcode-server/commit/ac6e7332a)).
- Уровень логирования трассировки стека изменён на «dpanic» ([`3f7755cd0`](https://github.com/lowcode/lowcode-server/commit/3f7755cd0)).
- Позиционирование PageBuilder изменено, чтобы панель инструментов с иконками не появлялась под заголовком и описанием блока ([`667a9b0`](https://github.com/lowcode/lowcode-webapp-compose/commit/667a9b0)).
- Список записей Low Code изменён: определён лимит по умолчанию и максимальный лимит для предотвращения перегрузки сервера ([`9e5fd42d4`](https://github.com/lowcode/lowcode-server/commit/9e5fd42d4)).
- Боковая навигация панели администрирования Low Code изменена для отображения ресурсов в виде дерева для упрощения обзора и доступа ([`14968080`](https://github.com/lowcode/lowcode-webapp-compose/commit/14968080), [`81e1c340`](https://github.com/lowcode/lowcode-webapp-compose/commit/81e1c340), [`8262db5`](https://github.com/lowcode/lowcode-vue/commit/8262db5)).
- Отображение пользователя изменено: используется handle пользователя, а затем ID ([`d1238b1`](https://github.com/lowcode/lowcode-js/commit/d1238b1)).
- Ссылка на выражения Low Code изменена для перенаправления на документацию ([`67834a52`](https://github.com/lowcode/lowcode-webapp-compose/commit/67834a52)).
- Встроенное редактирование списка записей теперь работает на страницах списков, а не только на страницах записей ([`6d070b62`](https://github.com/lowcode/lowcode-webapp-compose/commit/6d070b62)).
- Обработка запросов шлюза интеграции с помощью рабочих процессов автоматизации изменена: добавлены более гибкие интерфейсы для доступа к данным ([`#324`](https://github.com/lowcode/lowcode-server/pull/324), [`460646d45`](https://github.com/lowcode/lowcode-server/commit/460646d45)).
- Общие изменения UI ([`1b33bad3`](https://github.com/lowcode/lowcode-webapp-compose/commit/1b33bad3), [`c6bf400b`](https://github.com/lowcode/lowcode-webapp-compose/commit/c6bf400b), [`31d33048`](https://github.com/lowcode/lowcode-webapp-compose/commit/31d33048), [`1c0ec5ae`](https://github.com/lowcode/lowcode-webapp-compose/commit/1c0ec5ae), [`ab987f3d`](https://github.com/lowcode/lowcode-webapp-compose/commit/ab987f3d), [`feed8f0`](https://github.com/lowcode/lowcode-webapp-admin/commit/feed8f0), [`4f0e67b`](https://github.com/lowcode/lowcode-webapp-admin/commit/4f0e67b)), изменён адаптивный дизайн ([`e7cfa97c`](https://github.com/lowcode/lowcode-webapp-compose/commit/e7cfa97c)), скрытие/отключение кнопок, когда они не должны быть доступны ([`36195451`](https://github.com/lowcode/lowcode-webapp-compose/commit/36195451), [`59c797b8`](https://github.com/lowcode/lowcode-webapp-compose/commit/59c797b8)), изменены визуальные ресурсы ([`f5b09bf5`](https://github.com/lowcode/lowcode-webapp-compose/commit/f5b09bf5), [`5c72c30`](https://github.com/lowcode/lowcode-webapp-workflow/commit/5c72c30)).
- Макет журнала действий изменён для соответствия другим спискам логов ([`f96062e`](https://github.com/lowcode/lowcode-webapp-admin/commit/f96062e)).
- Воронкообразные диаграммы стали более настраиваемыми ([`dd28baec`](https://github.com/lowcode/lowcode-webapp-compose/commit/dd28baec), [`4d0eafb`](https://github.com/lowcode/lowcode-js/commit/4d0eafb)).


.[#2022*3_0-removed]#[2022*3*0-removed,Удалено:](#2022*3_0-removed,Удалено:)#


.[#2022*3_0-fixed]#[2022*3*0-fixed,Исправлено:](#2022*3_0-fixed,Исправлено:)*
- Исправлена потеря данных SQLite, вызванная отключением сеанса во время использования ([`23c7f357f`](https://github.com/lowcode/lowcode-server/commit/23c7f357f)).
- Исправлено неверное время жизни сеанса при регистрации пользователя; было установлено в ноль ([`f53463a32`](https://github.com/lowcode/lowcode-server/commit/f53463a32)).
- Исправлены проблемы с удалением внешнего провайдера аутентификации ([`74b3ddf94`](https://github.com/lowcode/lowcode-server/commit/74b3ddf94)).
- Исправлен некорректный разбор утверждения user-id ([`6c7d89a92`](https://github.com/lowcode/lowcode-server/commit/6c7d89a92)).
- Исправлено обслуживание экспортированных ресурсов аутентификации для режимов, отличных от разработки ([`dfe19c4c3`](https://github.com/lowcode/lowcode-server/commit/dfe19c4c3)).
- Исправлено переполнение меток шагов рабочего процесса ([`e424627`](https://github.com/lowcode/lowcode-webapp-workflow/commit/e424627)).
- Исправлено преждевременное завершение выполнения рабочего процесса, вызванное параллельными шагами задержки ([`4fd0ddfaa`](https://github.com/lowcode/lowcode-server/commit/4fd0ddfaa)).
- Исправлено некорректное управление событиями рабочего процесса путём ручного удаления слушателей событий ([`deb301f`](https://github.com/lowcode/lowcode-webapp-workflow/commit/deb301f)), правильного уничтожения экземпляров ([`a192b4a`](https://github.com/lowcode/lowcode-webapp-workflow/commit/a192b4a)) и более надёжной работы привязки клавиш `ctrl+s` ([`faa25d1`](https://github.com/lowcode/lowcode-webapp-workflow/commit/faa25d1)).
- Исправлена невозможность открытия путей в боковой панели конфигуратора рабочего процесса ([`8e0148f`](https://github.com/lowcode/lowcode-webapp-workflow/commit/8e0148f)).
- Исправлено некорректное кодирование/декодирование типа метаданных actionlog ([`e833796aa`](https://github.com/lowcode/lowcode-server/commit/e833796aa)).
- Исправлена неверная процедура системной конфигурации, из-за которой настройки по умолчанию могли быть пропущены ([`7fd719364`](https://github.com/lowcode/lowcode-server/commit/7fd719364)).
- Исправлена системная паника, вызванная веб-сокетами ([`f76b94e74`](https://github.com/lowcode/lowcode-server/commit/f76b94e74)).
- Исправлены конфликты имён отчётов с системными значениями ([`ab8668955`](https://github.com/lowcode/lowcode-server/commit/ab8668955)) и отсутствующие значения в базе данных PostgreSQL ([`cd15f3eaf`](https://github.com/lowcode/lowcode-server/commit/cd15f3eaf)).
- Исправлено включение удалённых пространств имён в вывод отчётов репортёра ([`908008eba`](https://github.com/lowcode/lowcode-server/commit/908008eba)).
- Исправлено неотображение элементов отображения при переключении между разными отчётами ([`24d163a`](https://github.com/lowcode/lowcode-webapp-reporter/commit/24d163a)).
- Исправлены отсутствующие HTTP-заголовки локали при выполнении запросов из веб-приложения репортёра ([`8215602`](https://github.com/lowcode/lowcode-webapp-reporter/commit/8215602)).
- Исправлено возникновение ошибок в табличном элементе отображения репортёра, если данные недоступны ([`d2b0ad0`](https://github.com/lowcode/lowcode-vue/commit/d2b0ad0)).
- Исправлены отсутствующие переводы ресурсов полей модуля ([`e9dfe8254`](https://github.com/lowcode/lowcode-server/commit/e9dfe8254)).
- Исправлена некорректная обработка пустых переводов ресурсов: значение теперь мягко удаляется вместо пропуска изменения ([`631811929`](https://github.com/lowcode/lowcode-server/commit/631811929)).
- Исправлено некорректное обновление и обработка переводов ресурсов для страниц и модулей Low Code ([`c24d7160`](https://github.com/lowcode/lowcode-webapp-compose/commit/c24d7160)).
- Исправлено количество записей для экспорта в ExporterModal ([`8f5f2c3`](https://github.com/lowcode/lowcode-webapp-compose/commit/8f5f2c3)).
- Исправлено определение направления справа-налево на основе текущего языка пользователя ([`30cc4eb7`](https://github.com/lowcode/lowcode-webapp-compose/commit/30cc4eb7), [`c882464`](https://github.com/lowcode/lowcode-webapp-admin/commit/c882464)).
- Исправлена фильтрация для полей `DateTime` только по времени и дате ([`a755b984`](https://github.com/lowcode/lowcode-webapp-compose/commit/a755b984), [`6ab98c52`](https://github.com/lowcode/lowcode-webapp-compose/commit/6ab98c52), [`a755b984`](https://github.com/lowcode/lowcode-webapp-compose/commit/a755b984)).
- Исправлено удаление кнопок автоматизации для ненастроенных кнопок в конфигураторе блочного элемента списка записей ([`7a7307b6`](https://github.com/lowcode/lowcode-webapp-compose/commit/7a7307b6)).
- Исправлен организатор записей: изменена оценка предфильтра ([`74b047de`](https://github.com/lowcode/lowcode-webapp-compose/commit/74b047de), [`2ea20d86`](https://github.com/lowcode/lowcode-webapp-compose/commit/2ea20d86)) и добавлена дополнительная оценка источника записей ([`5cbefcb8`](https://github.com/lowcode/lowcode-webapp-compose/commit/5cbefcb8)).
- Исправлена фильтрация списка записей: корректная обработка логических значений ([`43a1d3cf`](https://github.com/lowcode/lowcode-webapp-compose/commit/43a1d3cf)), добавлены отсутствующие системные поля ([`67e7f4fa`](https://github.com/lowcode/lowcode-webapp-compose/commit/67e7f4fa), [`7cae2b0b`](https://github.com/lowcode/lowcode-webapp-compose/commit/7cae2b0b)), изменена внутренняя логика ([`f11def50`](https://github.com/lowcode/lowcode-webapp-compose/commit/f11def50), [`751c589b`](https://github.com/lowcode/lowcode-webapp-compose/commit/751c589b)), исправлена некорректная обработка типов полей ([`1082cf73`](https://github.com/lowcode/lowcode-webapp-compose/commit/1082cf73)) и добавлен запасной вариант на оператор равенства ([`b5ee8752`](https://github.com/lowcode/lowcode-webapp-compose/commit/b5ee8752)).
- Исправлено неприменение настроек после их обновления и сохранения ([`7f5eebe`](https://github.com/lowcode/lowcode-webapp-admin/commit/7f5eebe)).
- Исправлены проблемы тайм-аута в функции HTTP-запроса рабочего процесса ([`6620b6ea`](https://github.com/lowcode/lowcode-server/commit/6620b6ea)).
- Исправлено сопоставление имени очереди в событии onMessage рабочего процесса ([`465e8ffe`](https://github.com/lowcode/lowcode-server/commit/465e8ffe)).

.[#2022*3_0-security]#[2022*3*0-security,Безопасность:](#2022*3_0-security,Безопасность:)*
- Заменён `dgrijalva/jwt-go` на `lestrrat-go/jwx`, выполнен рефакторинг реализации обработки JWT ([`59ec77e20`](https://github.com/lowcode/lowcode-server/commit/59ec77e20), [`46675080f`](https://github.com/lowcode/lowcode-server/commit/46675080f)), эмитента токенов (https://github.com/lowcode/lowcode-server/commit/6c3bef075[`6c3bef075`}) и перемещена валидация токенов на более ранние этапы (https://github.com/lowcode/lowcode-server/commit/53dd7cc29[`53dd7cc29`}) для более безопасной, надёжной и настраиваемой реализации.
- Добавлена валидация токенов и декодирование идентификатора для веб-сокетов ([`f9c8066e2`](https://github.com/lowcode/lowcode-server/commit/f9c8066e2)).
- Добавлены отсутствующие роли в JWT-токены имперсонализированного пользователя ([`ab805f007`](https://github.com/lowcode/lowcode-server/commit/ab805f007)).
- Изменены настройки CORS ([`9fe21dd8c`](https://github.com/lowcode/lowcode-server/commit/9fe21dd8c)).
- Добавлена санация меток рабочего процесса для предотвращения потенциального XSS ([`82d8f23`](https://github.com/lowcode/lowcode-webapp-workflow/commit/82d8f23)).


.[#2022*3_0-development]#[2022*3*0-development,Разработка:](#2022*3_0-development,Разработка:)*
- Выполнен рефакторинг и улучшение генератора кода с помощью cuelang ([`725f7e9e2`](https://github.com/lowcode/lowcode-server/commit/725f7e9e2)), старая генерация кода перенесена в новое средство ([`a035e6106`](https://github.com/lowcode/lowcode-server/commit/a035e6106), [`3bddce4d3`](https://github.com/lowcode/lowcode-server/commit/3bddce4d3), [`d103d60a3`](https://github.com/lowcode/lowcode-server/commit/d103d60a3), [`d09b037e8`](https://github.com/lowcode/lowcode-server/commit/d09b037e8)).
- Добавлен генератор `.env.example` ([`80d9b466a`](https://github.com/lowcode/lowcode-server/commit/80d9b466a)).
- Исправлено несоответствие версий `pinio` между `lowcode-js` и `lowcode-vue` ([`a8c46b9`](https://github.com/lowcode/lowcode-js/commit/a8c46b9), [`a8f5e24`](https://github.com/lowcode/lowcode-vue/commit/a8f5e24)).
- Изменён Envoy для корректной обработки поиска email пользователей ([`d841aad13`](https://github.com/lowcode/lowcode-server/commit/d841aad13)), расширено определение ресурсов для упрощения доступа к состоянию ([`da1828642`}), удалены потенциальные циклы импорта (https://github.com/lowcode/lowcode-server/commit/bc1550938[`bc1550938`](https://github.com/lowcode/lowcode-server/commit/da1828642)) и общие улучшения ([`f5a93caa`](https://github.com/lowcode/lowcode-webapp-compose/commit/f5a93caa), [`7bb22d96`](https://github.com/lowcode/lowcode-webapp-compose/commit/7bb22d96)).
- Улучшен запуск HTTP-сервера ([`d74239c73`](https://github.com/lowcode/lowcode-server/commit/d74239c73)).
- Очищен интерфейс хранилища ([`1949782cc`](https://github.com/lowcode/lowcode-server/commit/1949782cc)).
- Добавлены дополнительные идентификаторы для ресурсов отчётов ([`330a332fd`](https://github.com/lowcode/lowcode-server/commit/330a332fd)).
- Определён общий компонент для выбора элементов, таких как поля модуля и столбцы диаграмм ([`c3c2d5d6`](https://github.com/lowcode/lowcode-webapp-compose/commit/c3c2d5d6), [`e4a93e78`](https://github.com/lowcode/lowcode-webapp-compose/commit/e4a93e78), [`0b39fcc8`](https://github.com/lowcode/lowcode-webapp-compose/commit/0b39fcc8), [`02b7cb9`](https://github.com/lowcode/lowcode-webapp-reporter/commit/02b7cb9), [`bd02931`](https://github.com/lowcode/lowcode-vue/commit/bd02931)).
- Определён универсальный переиспользуемый контейнер диаграмм ([`afc6d15`](https://github.com/lowcode/lowcode-vue/commit/afc6d15)).
- Обновлены поля даты с помощью компонента `CDateInput` ([`d65b2c41`](https://github.com/lowcode/lowcode-webapp-compose/commit/d65b2c41), [`594b6ea5`](https://github.com/lowcode/lowcode-webapp-compose/commit/594b6ea5), [`f5690fa6`](https://github.com/lowcode/lowcode-webapp-compose/commit/f5690fa6)).
- Обновлена структура определения запроса рабочего процесса для большей гибкости ([`fe8645c`](https://github.com/lowcode/lowcode-vue/commit/fe8645c), [`180271f`](https://github.com/lowcode/lowcode-vue/commit/180271f)).


:leveloffset: -1
