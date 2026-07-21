<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'

const { t } = useI18n()

const props = defineProps({
  report: {
    type: Object,
    required: false,
    default: undefined,
  },
  chart: {
    type: compose.Chart,
    default: () => new compose.Chart(),
  },
  modules: {
    type: Array,
    required: true,
  },
  supportedMetrics: {
    type: Number,
    required: false,
    default: -1,
  },
  dimensionFieldKind: {
    type: Array,
    required: false,
    default: () => ['DateTime', 'Select', 'Number', 'Bool', 'String', 'Record', 'User'],
  },
  usesDimensionsField: {
    type: Boolean,
    default: true,
  },
  unSkippable: {
    type: Boolean,
    required: false,
    default: false,
  },
})

const emit = defineEmits(['update:report'])

const checkboxLabel = {
  on: t('label.yes'),
  off: t('label.no'),
}

const formatOptions = [
  { value: 'custom', text: t('edit.formatting.presetFormats.options.custom') },
  { value: 'accounting', text: t('edit.formatting.presetFormats.options.accounting') },
]

const editReport = computed({
  get: () => props.report,
  set: (v) => emit('update.report', v),
})

const isNew = computed(() => props.chart.chartID === NoID)
</script>
