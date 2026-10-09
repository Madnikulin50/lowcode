# Журнал изменений

Страница журнала изменений — это место, где мы собираем все изменения для данного релиза.
Все релизы находятся на одной странице в обратном хронологическом порядке (новые сверху).

Основная структура основана на [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## Файл индекса

Файл индекса обобщает основные релизы, на которых мы сосредоточились, такие как DAL в релизе `2022.9`.
Файл индекса находится в папке, обозначающей версию, например `202209/index.adoc`.

.Файл индекса имеет следующую структуру:
```adoc
```
\include::ROOT:partial$variables.adoc[]

<a id="yyyy-mm-x"></a>
# `yyyy.mm`


:leveloffset: +1


:leveloffset: -1


## Патч-версии

Патч-версии описывают изменения, представленные в данном патче версии, такие как дополнения, вывод из эксплуатации и изменения.

Теперь мы используем инструмент для журнала изменений, чтобы он взял на себя основную работу.
Патч-релизы описываются в `JSON` файлах внутри директории `src/modules/ROOT/pages/changelog` (в папке версии).
В качестве примера, патч `02` для версии `2022.9` будет находиться в файле `src/modules/ROOT/pages/changelog/202209/02.json`.

.Файл JSON патч-релиза имеет следующую структуру:
```json
```
{
  "meta": {
    "releasedOn": "YYYY.MM.DD"
  },

  "added": [],
  "changed": [],
  "deprecated": [],
  "removed": [],
  "fixed": [],
  "security": [],
  "development": []
}

.Простой пример с единственной записью журнала изменений:
```json
```
{
  "meta": {
    "releasedOn": "2022.10.20"
  },

  "added": [{
    "what": ["..."],
    "why": ["..."],
    "refs": ["..."]
  }],
  "changed": []
}

!!! note
    Вы можете свободно опускать ненужные элементы в JSON.


## Записи журнала изменений

### `added`

`added` описывает дополнения к LowCoooode, такие как новые функции или добавленные элементы интерфейса.

[cols="1s,5a"]
|===
| [#entry-added-what]#[entry-added-what,what](#entry-added-what,what)#
|
Опишите, что представляет собой новое дополнение.
В качестве примера: `"what": "gauge, funnel, and pie charts to the Low Code charts"`.

Значение `what` должно предоставить читателю достаточно информации, чтобы понять, что было добавлено и куда оно было добавлено.

| [#entry-added-why]#[entry-added-why,why *необязательно*](#entry-added-why,why *необязательно*)#
|
Опишите обоснование нового дополнения.
В качестве примера: `"why": "to cover use-cases which require these charts"`.

Значение `why` должно дать некоторое представление о предпосылке дополнения.

|===

.Шаблон:
```json
```
{
  "what": ["..."],
  "why": ["..."],
  "refs": []
}

### `changed`

`changed` описывает изменения существующих функций LowCoooode.

[cols="1s,5a"]
|===
| [#entry-changed-what]#[entry-changed-what,what](#entry-changed-what,what)#
|
Опишите, что было изменено.
В качестве примера: `"what": "workflow initialization and preprocessing pipeline"`.

Значение `what` должно предоставить читателю достаточно информации, чтобы понять, что было изменено.

| [#entry-changed-why]#[entry-changed-why,why *необязательно*](#entry-changed-why,why *необязательно*)#
|
Опишите обоснование изменения.
В качестве примера: `"why": "to increase performance and optimize memory usage"`.

Значение `why` должно дать некоторое представление о предпосылке дополнения.

| [#entry-changed-how]#[entry-changed-how,how *необязательно*](#entry-changed-how,how *необязательно*)#
|
Опишите, как изменилась *вещь*.
В качестве примера: `"how": "by pre-processing workflows and initializing them in LowCoooode boot procedure"`.

Значение `how` должно предоставить читателю достаточно информации, чтобы понять новое состояние и возможные последствия.

|===

.Шаблон:
```json
```
{
  "what": ["..."],
  "why": ["..."],
  "how": ["..."],
  "refs": []
}

### `deprecated`

`deprecated` описывает функции, ожидающие удаления.
Устаревшие функции всё ещё могут использоваться, но на них больше не следует полагаться, так как мы можем удалить их в будущих релизах.

[cols="1s,5a"]
|===
| [#entry-deprecated-what]#[entry-deprecated-what,what](#entry-deprecated-what,what)#
|
Опишите, что было выведено из эксплуатации.
В качестве примера: `"what": "automation scripts and the Corredor server"`.

Значение `what` должно предоставить читателю достаточно информации, чтобы понять, что было выведено из эксплуатации.

| [#entry-deprecated-why]#[entry-deprecated-why,why *необязательно*](#entry-deprecated-why,why *необязательно*)#
|
Опишите обоснование вывода из эксплуатации.
В качестве примера: `"why": "because it's a legacy system and the NG Corredor server will replace it"`.

Значение `why` должно дать некоторое представление об обосновании вывода из эксплуатации.

|===

.Шаблон:
```json
```
{
  "what": ["..."],
  "why": ["..."],
  "refs": []
}

### `removed`

`removed` описывает функции, которые больше недоступны в LowCoooode.

[cols="1s,5a"]
|===
| [#entry-removed-what]#[entry-removed-what,what](#entry-removed-what,what)#
|
Опишите, что было удалено.
В качестве примера: `"what": "automation scripts and the Corredor server"`.

Значение `what` должно предоставить читателю достаточно информации, чтобы понять, что было удалено.

| [#entry-removed-why]#[entry-removed-why,why *необязательно*](#entry-removed-why,why *необязательно*)#
|
Опишите обоснование удаления.
В качестве примера: `"why": "because the NG Corredor server replaced it"`.

Значение `why` должно дать некоторое представление об обосновании вывода из эксплуатации.

|===

.Шаблон:
```json
```
{
  "what": ["..."],
  "why": ["..."],
  "refs": []
}

### `fixed`

`fixed` описывает изменения существующих функций LowCoooode.

!!! important
    Если исправление ошибки меняет исходный ход работы функции, упомяните об этом также в разделе `changed`.


[cols="1s,5a"]
|===
| [#entry-fixed-what]#[entry-fixed-what,what](#entry-fixed-what,what)#
|
Опишите, что было исправлено.
В качестве примера: `"what": "funnel chart labels rendered over the page block when there were too many available options"`.

Значение `what` должно предоставить читателю достаточно информации, чтобы понять, что было исправлено.
Уделите больше внимания тому, в чём заключалась проблема с точки зрения пользовательского интерфейса, а не логики.

| [#entry-fixed-how]#[entry-fixed-how,how *необязательно*](#entry-fixed-how,how *необязательно*)#
|
Опишите, как была исправлена *вещь*.
В качестве примера: `"how": "by moving overflowing elements into /dev/null"`.

Значение `how` должно предоставить читателю достаточно информации о том, что делает патч.

|===

.Шаблон:
```json
```
{
  "what": ["..."],
  "how": ["..."],
  "refs": []
}

### `security`

`security` описывает исправления, связанные с безопасностью.

!!! important
    Если исправление ошибки меняет исходный ход работы функции, упомяните об этом также в разделе `changed`.


[cols="1s,5a"]
|===
| [#entry-security-what]#[entry-security-what,what](#entry-security-what,what)#
|
Опишите, в чём заключалась уязвимость безопасности.
В качестве примера: `"what": "Fixed the stored XSS attack via xyz"`.

Значение `what` должно предоставить читателю достаточно информации, чтобы понять проблему и её причину.
Уделите больше внимания тому, в чём заключалась проблема с точки зрения пользовательского интерфейса, а не логики.

| [#entry-security-how]#[entry-security-how,how *необязательно*](#entry-security-how,how *необязательно*)#
|
Опишите, как была исправлена *вещь*.
В качестве примера: `"how": "by removing xyz completely"`.

Значение `how` должно предоставить читателю достаточно информации о том, что делает патч.

|===

.Шаблон:
```json
```
{
  "what": ["..."],
  "how": ["..."],
  "refs": []
}

### `development`

`development` описывает связанные с разработкой улучшения качества жизни.

[cols="1s,5a"]
|===
| [#entry-development-what]#[entry-development-what,what](#entry-development-what,what)#
|
Опишите, в чём заключалось улучшение качества жизни для разработки.
В качестве примера: `"what": "Added C3 to simplify component development in a containerized environment"`.

Значение `what` должно предоставить читателю достаточно информации, чтобы понять, что было добавлено и почему это хорошо.

|===

.Шаблон:
```json
```
{
  "what": ["..."],
  "refs": []
}

## Добавление ссылок

### Коммиты GitHub

Чтобы сослаться на коммит GitHub, просто добавьте полный URL в массив `refs`.

.Пример ссылки на коммит GitHub:
```json
```
{
  "refs": ["https://github.com/lowcode/lowcode-server/commit/fd6465d0f95d78210401f0d0be9c16e1290341af"]
}

## Советы по улучшению журналов изменений

- избегайте двусмысленных формулировок, таких как `Fixed Low Code chart rendering` — в чём заключалась проблема? Как она была исправлена? Та ли это проблема, которая затронула меня?
- Старайтесь формулировать записи журнала изменений так, чтобы они описывали, с чем столкнулся пользователь или что пользователь мог бы сделать с этим.
В качестве примера, `"Fixed chart legend overflowing onto the chart when a lot of options are presented"` вместо `Fixed chart legend positioning"`.
