const buildVueConfig = require('./vue.config-builder')

module.exports = buildVueConfig({
  appFlavour: 'Reporter',
  appName: 'reporter',
  appLabel: 'LowCoooode Reporter Editor',
  theme: 'corteza-base',
  packageAlias: 'corteza-webapp-reporter',
})
