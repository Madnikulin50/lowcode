# Фильтр списка записей

Фильтр списка записей предоставляет набор компонентов, которые позволяют реализовать гибкую фильтрацию записей.

Обратитесь к [реализации фильтра списка записей](https://github.com/lowcode/lowcode/blob/{PAGE-VERSION}.x/client/web/compose/src/components/PageBlocks/RecordListBase.vue) за примером использования.

!!! note
    В будущих релизах компонент может быть обобщён для различных приложений.


## Обзор структуры

### `components/Common/RecordListFilter.vue`

Файл `components/Common/RecordListFilter.vue` определяет основные компоненты функции фильтра списка записей.

.Пример использования:
```js
```
<record-list-filter
    :selectedField="field.moduleField" <1>
    :namespace="namespace" <2>
    :module="recordListModule" <3>
    :recordListFilter="recordListFilter" <4>
    @filter="onFilter"
/>
<1> свойство `selectedField` определяет поле по умолчанию, которое должно использоваться при задании фильтров.
<2> свойство `namespace` определяет объект пространства имён, на основе которого мы задаём фильтр.
<3> свойство `module` определяет объект модуля, на основе которого мы задаём фильтр.
<4> свойство `recordListFilter` определяет фильтр, который вы хотите отобразить в данном компоненте.

### `/src/lib/record-filter.js`

Файл `/src/lib/record-filter.js` определяет логику преобразования выходных данных компонента `RecordListFilter` в запрос, который можно использовать с сервером LowCoooode.

.Пример использования:
```js
```
import { queryToFilter } from 'lowcode-webapp-compose/src/lib/record-filter'
