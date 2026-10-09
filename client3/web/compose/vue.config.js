const buildVueConfig = require('./vue.config-builder')

module.exports = buildVueConfig({
  appFlavour: 'Namespaces',
  appName: 'compose',
  appLabel: 'Lowcooode Compose',
  theme: 'corteza-base',
  packageAlias: 'corteza-webapp-compose',
})
