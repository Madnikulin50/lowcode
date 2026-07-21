import { capitalize, uniq } from 'lodash'
import { h } from 'vue'

import AutomationBase from './AutomationBase.vue'
import AutomationConfigurator from './AutomationConfigurator.vue'
import CalendarBase from './CalendarBase.vue'
import CalendarConfigurator from './CalendarConfigurator/index.vue'
import ChartBase from './ChartBase.vue'
import ChartConfigurator from './ChartConfigurator.vue'
import CommentBase from './Comment/Base.vue'
import CommentConfigurator from './Comment/Configurator.vue'
import ContentBase from './ContentBase.vue'
import ContentConfigurator from './ContentConfigurator.vue'
import FileBase from './FileBase.vue'
import FileConfigurator from './FileConfigurator.vue'
import GeometryBase from './GeometryBase.vue'
import GeometryConfigurator from './GeometryConfigurator/index.vue'
import IFrameBase from './IFrameBase.vue'
import IFrameConfigurator from './IFrameConfigurator.vue'
import MetricBase from './MetricBase.vue'
import MetricConfigurator from './MetricConfigurator/index.vue'
import NavigationBase from './Navigation/Base.vue'
import NavigationConfigurator from './Navigation/Configurator.vue'
import ProgressBase from './ProgressBase.vue'
import ProgressConfigurator from './ProgressConfigurator.vue'
import RecordBase from './RecordBase.vue'
import RecordConfigurator from './RecordConfigurator.vue'
import RecordEditor from './RecordEditor.vue'
import RecordListBase from './RecordListBase.vue'
import RecordListConfigurator from './RecordListConfigurator.vue'
import RecordOrganizerBase from './RecordOrganizerBase.vue'
import RecordOrganizerConfigurator from './RecordOrganizerConfigurator.vue'
import RecordRevisionsBase from './RecordRevisionsBase.vue'
import RecordRevisionsConfigurator from './RecordRevisionsConfigurator.vue'
import ReportBase from './Report/Base.vue'
import ReportConfigurator from './Report/Configurator.vue'
import SocialFeedBase from './SocialFeedBase.vue'
import SocialFeedConfigurator from './SocialFeedConfigurator.vue'
import TabsBase from './TabsBase.vue'
import TabsConfigurator from './TabsConfigurator.vue'

const Registry = {
  AutomationBase,
  AutomationConfigurator,
  CalendarBase,
  CalendarConfigurator,
  ChartBase,
  ChartConfigurator,
  ContentBase,
  ContentConfigurator,
  FileBase,
  FileConfigurator,
  IFrameBase,
  IFrameConfigurator,
  RecordBase,
  RecordConfigurator,
  RecordEditor,
  RecordListBase,
  RecordListConfigurator,
  RecordRevisionsBase,
  RecordRevisionsConfigurator,
  RecordOrganizerBase,
  RecordOrganizerConfigurator,
  ReportBase,
  ReportConfigurator,
  SocialFeedBase,
  SocialFeedConfigurator,
  MetricBase,
  MetricConfigurator,
  CommentBase,
  CommentConfigurator,
  ProgressBase,
  ProgressConfigurator,
  GeometryBase,
  GeometryConfigurator,
  TabsBase,
  TabsConfigurator,
  NavigationConfigurator,
  NavigationBase,
}

const defaultMode = 'Base'

function GetComponent({ block, mode = defaultMode }) {
  if (!block) {
    throw new Error('block prop missing')
  }

  const { kind } = block
  for (mode of uniq([capitalize(mode), defaultMode])) {
    if (mode === 'Editor' && block.options.referenceField && block.options.referenceModuleID) {
      mode = 'Base'
    }

    const cmpName = kind + mode
    if (Object.hasOwnProperty.call(Registry, cmpName)) {
      return Registry[cmpName]
    }
  }

  throw new Error('unknown block kind: ' + kind)
}

function PageBlock(props, { slots, attrs }) {
  const component = GetComponent(props)
  return h(component, { ...props, ...attrs }, slots)
}

export { Registry, GetComponent }
export default PageBlock
