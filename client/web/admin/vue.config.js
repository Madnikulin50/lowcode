const buildVueConfig = require('./vue.config-builder')

module.exports = buildVueConfig({
  appFlavour: 'Admin Area',
  appName: 'admin',
  appLabel: 'Corteza Admin',
  theme: 'basic',
  packageAlias: 'corteza-webapp-admin',
})
