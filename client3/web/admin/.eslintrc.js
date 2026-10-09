module.exports = {
  extends: [
    'plugin:vue/essential', // or 'plugin:vue/recommended'
    '@vue/standard'         // This must match the package installed
    ],
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@babel/eslint-parser',
    requireConfigFile: false,
    ecmaVersion: 2020,
    sourceType: 'module',
  },
}
