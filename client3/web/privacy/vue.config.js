const buildVueConfig = require('./vue.config-builder')

module.exports = buildVueConfig({
  appFlavour: 'Privacy',
  appName: 'privacy',
  appLabel: 'Lowcooode Privacy',
  theme: 'corteza-base',
  packageAlias: 'corteza-webapp-privacy',
})
