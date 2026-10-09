const buildVueConfig = require('./vue.config-builder')

module.exports = buildVueConfig({
  appFlavour: 'One',
  appName: 'one',
  appLabel: 'Lowcooode One',
  theme: 'corteza-base',
  packageAlias: 'corteza-webapp-one',
})
