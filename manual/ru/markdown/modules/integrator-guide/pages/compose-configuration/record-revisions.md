# Отслеживание версий записей

LowCoooode предоставляет функцию версий записей, которая позволяет настраивать модули для отслеживания изменений записей путём введения нового счётчика версий и предоставления истории изменений.
LowCoooode ведёт счётчик версий для каждой записи (начиная с 1), который увеличивается при каждом обновлении записи, и для каждой версии сохраняются только изменения значений записи.

Версии должны быть включены для конкретного модуля, и их можно просматривать внутри блока страницы версий записей.
При включении все изменения записей сохраняются в базе данных с автором и меткой времени.

!!! note
    Версия добавляется только при обновлении записи, поэтому у новых записей нет версий, что уменьшает количество сохраняемых версий.


<a id="configure-module"></a>
## Включение отслеживания версий

Чтобы включить отслеживание версий для конкретных записей, нам нужно включить опцию в модуле Low Code.
В дополнение к конфигурации модуля вам нужно будет [configure-record-page,настроить страницу записи](#configure-record-page,настроить страницу записи) для отображения истории изменений.

!!! important
    Версии записей хранятся в том же подключении, что и сам модуль.


Сначала перейдите в пространство имён Low Code и войдите в область администратора (вы также можете включить версии записей для совершенно новых модулей).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/ns-home.png",
    "alias": "compose-configuration-record-revisions-ns-home",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 17,
    "y": 135,
    "w": 318,
    "h": 35
  }]
}

В списке модулей нажмите на тот, для которого вы хотите включить версии записей.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-modules.png",
    "alias": "compose-configuration-record-revisions-admin-modules",
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
    "kind": "box",
    "x": 525,
    "y": 222,
    "h": 65,
    "w": 1220
  }]
}

Нажмите на вкладку record revisions и установите флажок **enable record revisions checkbox**.
При желании введите опцию идентификатора.

!!! note
    Опция идентификатора позволяет указать таблицу базы данных или коллекцию, которую LowCoooode должен использовать для регистрации версий.
    Используется определённое системой расположение, которое должно удовлетворять большинство сценариев использования.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-module-edit-revisions-enable.png",
    "alias": "compose-configuration-record-revisions-admin-module-edit-revisions-enable",
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
    "kind": "box",
    "x": 547,
    "y": 255,
    "w": 213,
    "h": 19
  }]
}

После завершения нажмите на кнопку btn:[save], чтобы сохранить изменения и включить отслеживание версий записей.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-module-edit-revisions-enable.png",
    "alias": "compose-configuration-record-revisions-admin-module-edit-revisions-enable-save",
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
    "kind": "box",
    "x": 1821,
    "y": 1015,
    "w": 81,
    "h": 47
  }]
}

<a id="configure-record-page"></a>
## Отображение истории изменений

Чтобы отобразить историю изменений, нам нужно добавить [блок страницы **версий записей**](modules/integrator-guide/pages/compose-configuration/compose-configuration/page-blocks.md#page-block-record-revisions) на страницу записи.
Вы можете пропустить этот шаг, если хотите только отслеживать историю изменений, но не отображать её.

!!! important
    Поскольку они тесно связаны, блок страницы версий записей может появляться только на странице записи.


Перейдите на страницу записи непосредственно из самого модуля или из списка страниц (на скриншоте ниже показано, как перейти на страницу записи с экрана редактирования модуля).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-module-edit-open-record-page.png",
    "alias": "compose-configuration-record-revisions-admin-module-edit-open-record-page",
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
    "kind": "box",
    "x": 1563,
    "y": 175,
    "w": 133,
    "h": 18
  }]
}

Затем нажмите на кнопку btn:[+ add block] на панели инструментов, чтобы открыть модальное окно выбора блока страницы.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-record-page-base.png",
    "alias": "compose-configuration-record-revisions-admin-record-page-base",
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
    "kind": "box",
    "x": 1028,
    "y": 1013,
    "w": 144,
    "h": 48
  }]
}

Найдите и нажмите на блок страницы версий записей в модальном окне выбора блока.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-record-page-add-block-select.png",
    "alias": "compose-configuration-record-revisions-admin-record-page-add-block-select",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "padding": "xs",
    "x": 530,
    "y": 41,
    "w": 860,
    "h": 895
  },
  "annotations": [{
    "kind": "box",
    "x": 554,
    "y": 680,
    "h": 50,
    "w": 250
  }]
}

В модальном окне конфигурации блока страницы нажмите на вкладку record revisions и убедитесь, что конфигурация соответствует вашим потребностям.
Конфигурация позволяет выбрать подмножество полей модуля для отображения в блоке страницы версий записей.

Когда вы удовлетворены, нажмите на кнопку btn:[save], чтобы настроить блок страницы и добавить его на страницу записи.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-record-page-add-block-configure.png",
    "alias": "compose-configuration-record-revisions-admin-record-page-add-block-configure",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "padding": "xs",
    "x": 344,
    "y": 42,
    "w": 1232,
    "h": 584
  },
  "annotations": [{
    "kind": "box",
    "x": 1454,
    "y": 583,
    "w": 115,
    "h": 36
  }]
}

Разместите только что добавленный блок страницы версий записей так, чтобы он подходил вашей странице.

Когда отслеживание версий включено, LowCoooode также предоставляет новое **поле модуля версий записей**, которое указывает номер версии.
Вы можете добавить поле модуля как в **блок записи**, так и в **блок списка записей**.

Когда вы удовлетворены своими изменениями, нажмите на кнопку btn:[save], чтобы сохранить изменения.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/admin-record-page-add-block-place.png",
    "alias": "compose-configuration-record-revisions-admin-record-page-add-block-place",
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
    "y": 68,
    "x": 1394,
    "w": 518,
    "h": 882
  }, {
    "kind": "box",
    "x": 1820,
    "y": 1015,
    "w": 82,
    "h": 48
  }]
}

## Просмотр истории изменений

!!! note
    Чтобы просмотреть историю изменений, вам сначала нужно будет <<configure-module,настроить модуль>> и <<configure-record-page,добавить блок страницы версий записей>> на страницу записи.


Чтобы просмотреть версии записей, нажмите на запись, для которой вы хотите увидеть историю изменений.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/ns-home-with-revisions.png",
    "alias": "compose-configuration-record-revisions-ns-home-with-revisions",
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
    "kind": "box",
    "x": 370,
    "y": 464,
    "w": 1533,
    "h": 51
  }]
}

Вы можете просмотреть всю историю изменений в блоке страницы версий записей.
Каждая запись версии фиксирует метку времени, выполненную операцию и пользователя, выполнившего операцию.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/view-record-revision-overview.png",
    "alias": "compose-configuration-record-revisions-view-record-revision-overview",
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
    "padding": "xs",
    "x": 1410,
    "y": 89,
    "w": 482,
    "h": 886
  },
  "annotations": []
}

Чтобы просмотреть детали версии, нажмите на кнопку btn:[change(s)] в крайнем правом углу версии, которую хотите проверить.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/view-record-revision-overview.png",
    "alias": "compose-configuration-record-revisions-view-record-revision-overview-expand-details",
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
    "padding": "xs",
    "padding": "xs",
    "x": 1410,
    "y": 89,
    "w": 482,
    "h": 886
  },
  "annotations": [{
    "kind": "box",
    "x": 1784,
    "y": 149,
    "w": 88,
    "h": 41
  }]
}

Детальный просмотр выведет список зарегистрированных изменений значений для этой версии.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "compose-configuration/record-revisions/view-record-revision-details.png",
    "alias": "compose-configuration-record-revisions-view-record-revision-details",
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
    "padding": "xs",
    "padding": "xs",
    "x": 1410,
    "y": 89,
    "w": 482,
    "h": 886
  },
  "annotations": [{
    "kind": "box",
    "x": 1468,
    "y": 220,
    "w": 420,
    "h": 255
  }]
}
