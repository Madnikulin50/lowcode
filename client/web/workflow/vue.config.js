const buildVueConfig = require('./vue.config-builder')

module.exports = buildVueConfig({
  appFlavour: 'Workflows',
  appName: 'workflow',
  appLabel: 'Lowcooode Workflows',
  theme: 'corteza-base',
  packageAlias: 'corteza-workflow',
})
