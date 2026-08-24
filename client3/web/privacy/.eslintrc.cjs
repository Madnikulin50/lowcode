module.exports = {
  extends: ['corteza-client'],
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@babel/eslint-parser',
    requireConfigFile: false,
    // babel.config.cjs is leftover webpack/Vue-CLI config, unused by this
    // app's Vite build -- don't let @babel/eslint-parser pick it up.
    babelOptions: {
      configFile: false,
      babelrc: false,
    },
    ecmaVersion: 2020,
    sourceType: 'module',
  },
}
