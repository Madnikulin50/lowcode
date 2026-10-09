const buildVueConfig = require('./vue.config-builder')

module.exports = buildVueConfig({
  appFlavour: 'Admin Area',
  appName: 'admin',
  appLabel: 'Lowcooode Admin',
  theme: 'basic',
  packageAlias: 'corteza-webapp-admin',
})
