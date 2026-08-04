# Триггеры

## Явный триггер (кнопка)

```js
```
triggers ({ on }) {
  return on('manual')
    .for('compose:record')
    .uiProp('app', 'compose')
},

```js
```
triggers ({ on }) {
  return on('manual')
    .for('compose:record')
    .where('module', 'Module1')
    .uiProp('app', 'compose')
},

## Неявный триггер (автоматический)

```js
```
triggers ({ before }) {
  return before('create', 'update')
      .for('compose:record')
      .where('module', 'Module1')
},

triggers ({ after }) {
  return after('create', 'update')
      .for('compose:record')
      .where('module', 'Module1')
},

triggers ({ after }) {
  return after('create', 'update')
      .where('module', 'Module1')
      .where('record.values.Name', 'John')
},

triggers ({ after }) {
  return after('create', 'update')
      .for('system:user')
      .where('user.email', '^[a-zA-Z0-9]{1,20}@lowcode.com$')
},


<a id="sample-trigger-deferred-interval"></a>
## Отложенный (интервал)

```js
```
triggers ({ on }) {
  return on('interval')
    .every('0 2 * * *')
},

triggers ({ on }) {
  return on('interval')
    .every('00 12 1 */4 *')
},


<a id="sample-trigger-deferred-timestamp"></a>
## Отложенный (метка времени)

```js
```
triggers ({ at }) {
  return at('2020-12-24T21:00:00Z')
},

## По HTTP-запросу

```js
```
triggers ({ on }) {
  return on('request')
    .where('request.path', '/some/path/here')
    .where('request.method', 'POST')
    .for('system:sink'),
},

triggers ({ on }) {
  return on('request')
    .where('request.path', '/some/path/here')
    .where('request.headers.authorization', '')
    .for('system:sink'),
},
