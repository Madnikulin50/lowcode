# Серверный скрипт

```js
```
export default {
  label: "label goes here",
  description: "description goes here",

  triggers({ on }) {
    return on('manual')
      .for('compose')
      .uiProp('app', 'compose')
  },


  async exec(args, ctx) {
    console.log('Hello World!');
  },
}
