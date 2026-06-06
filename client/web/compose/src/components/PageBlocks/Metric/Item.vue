<template>
  <div
    v-if="options.likeRecordList"
    ref="fieldContainer"
  >
    <b-form-group
      :key="`${metric.label}`"
      :label-cols-md="options.horizontalFieldLayoutEnabled && '6'"
      :label-cols-xl="options.horizontalFieldLayoutEnabled && '5'"
      :content-cols-md="options.horizontalFieldLayoutEnabled && '6'"
      :content-cols-xl="options.horizontalFieldLayoutEnabled && '7'"
      :style="fieldWidth"
      class="field-container mb-0"
    >
      <template #label>
        <div
          class="d-flex align-items-center text-primary mb-0"
        >
          <span
            class="d-flex"
            style="margin-top: 0.1rem;"
            :style="genStyle(metric.valueStyle, true)"
          >
            {{ metric.label }}
          </span>
        </div>
      </template>
      <span
        :style="genStyle(metric.valueStyle)"
      >
        <template v-if="metric.prefix">
          {{ metric.prefix }}
        </template>
        {{ displayValue }}
        <template v-if="metric.suffix">
          {{ metric.suffix }}
        </template>
      </span>
    </b-form-group>
  </div>
  <div
    v-else
    :style="genStyle(metric.valueStyle)"
    class="text-center"
    :class="{'h-100': metric.valueStyle.notFitVertical !== true}"
  >
    <!--    This div is here because .svg metrics dont render with print to PDF option-->
    <div
      class="d-none d-print-flex w-100 align-items-center justify-content-center overflow-hidden"
      :style="genStyle(metric.valueStyle)"
    >
      <template v-if="metric.prefix">
        {{ metric.prefix }}
      </template>
      {{ displayValue }}
      <template v-if="metric.suffix">
        {{ metric.suffix }}
      </template>
    </div>

    <template v-if="metric.valueStyle.notFitVertical || metric.valueStyle.notFitHorizontal">
      <div
        class="d-print-flex align-items-center justify-content-center overflow-hidden"
        :style="genStyle(metric.valueStyle)"
      >
        <span
          v-if="metric.showLabel"
          :style="genStyle(metric.valueStyle, true)"
        >
          {{ metric.label }}:&nbsp;
        </span>
        <template v-if="metric.prefix">
          {{ metric.prefix }}
        </template>
        {{ displayValue }}
        <template v-if="metric.suffix">
          {{ metric.suffix }}
        </template>
      </div>
    </template>
    <template v-else>
      <svg
        :viewBox="getVB"
        class="h-100 w-100 d-flex d-print-none"
        width="100%"
        height="100%"
      >
        <text
          ref="metricItem"
          y="50%"
          x="50%"
          text-anchor="middle"
          dominant-baseline="central"
          text-rendering="geometricPrecision"
        >
          <template
            v-if="metric.showLabel"
          >
            {{ metric.label }}:&nbsp;
          </template>
          <template v-if="metric.prefix">
            {{ metric.prefix }}
          </template>
          {{ displayValue }}
          <template v-if="metric.suffix">
            {{ metric.suffix }}
          </template>
        </text>
      </svg>
    </template>
  </div>
</template>

<script>
import { fmt } from 'corteza-lib/js/dist'

export default {
  components: { },
  props: {
    metric: {
      type: Object,
      required: false,
      default: () => ({}),
    },
    options: {
      type: Object,
      required: false,
      default: () => ({}),
    },
    value: {
      type: Object,
      required: false,
      default: () => ({}),
    },
  },

  data () {
    return {
      vvb: ['0', '0', '0', '0'],
    }
  },

  computed: {
    getVB () {
      return this.vvb.join(' ')
    },

    displayValue () {
      return fmt.number(this.value.value)
    },
    fieldWidth () {
      if (this.options.recordFieldLayoutOption !== 'noWrap') {
        return {}
      }

      return { 'min-width': '13rem' }
    },

  },

  watch: {
    metric: {
      handler () {
        this.update()
      },
      immediate: true,
    },
    options: {
      handler () {
        this.update()
      },
      immediate: true,
    },
    value: {
      handler () {
        this.update()
      },
      immediate: true,
    },
    'options.recordFieldLayoutOption': {
      handler (newVal) {
        if (newVal === 'wrap' && this.fieldModule) {
          this.initializeResizeObserver(this.$refs.fieldContainer, this.options.recordFieldLayoutOption)
        } else if (this.resizeObserver) {
          this.resizeObserver.unobserve(this.$refs.fieldContainer)
          this.columnWrapClass = ''
        }
      },
    },
  },

  beforeDestroy () {
    this.setDefaultValues()
  },

  methods: {
    update () {
      this.$nextTick(() => {
        const { width, height } = this.$refs.metricItem.getBBox()
        const tmp = [...this.vvb]
        tmp[2] = parseInt(Math.ceil(width))
        tmp[3] = parseInt(Math.ceil(height))
        this.vvb = tmp
      })
    },

    genStyle (s = {}, forLabel = false) {
      const d = {
        fill: forLabel ? (s.labelColor || s.color) : s.color,
        backgroundColor: s.backgroundColor,
        fontSize: s.fontSize ? s.fontSize + 'px' : undefined,
        color: forLabel ? (s.labelColor || s.color) : s.color,
      }

      for (const v of Object.keys(d)) {
        if (d[v] === undefined) {
          delete d[v]
        }
      }

      return d
    },

    setDefaultValues () {
      this.vvb = []
    },
  },
}
</script>
