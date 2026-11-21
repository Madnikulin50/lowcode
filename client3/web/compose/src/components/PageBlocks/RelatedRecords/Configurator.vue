<template>
  <div>
    <h5>{{ $t('relatedRecords.label') }}</h5>

    <div
      v-if="!pageModuleID"
      class="alert alert-warning"
    >
      {{ $t('relatedRecords.recordPageOnly') }}
    </div>

    <template v-else>
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('relatedRecords.add.label') }}</label>
        <div class="input-group">
          <select
            v-model="candidateKey"
            class="form-select"
            :disabled="!candidates.length"
          >
            <option value="">
              {{ candidates.length ? $t('relatedRecords.add.placeholder') : $t('relatedRecords.add.none') }}
            </option>
            <option
              v-for="c in candidates"
              :key="c.key"
              :value="c.key"
            >
              {{ c.text }}
            </option>
          </select>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="!candidateKey"
            @click="addRelation"
          >
            {{ $t('relatedRecords.add.button') }}
          </button>
        </div>
        <small class="text-muted">{{ $t('relatedRecords.add.footnote') }}</small>
      </div>

      <div
        v-for="(relation, i) in options.relations"
        :key="`${relation.moduleID}:${relation.refField}`"
        class="card mb-3"
      >
        <div class="card-header d-flex align-items-center gap-2 py-2">
          <div class="flex-fill text-truncate">
            <span class="fw-semibold">{{ relationText(relation) }}</span>
            <span
              v-if="!isUsable(relation)"
              class="badge text-bg-danger ms-2"
            >{{ $t('relatedRecords.missing') }}</span>
          </div>
          <div class="btn-group btn-group-sm">
            <button
              type="button"
              class="btn btn-outline-secondary"
              :disabled="i === 0"
              :title="$t('relatedRecords.moveUp')"
              @click="move(i, -1)"
            >
              <font-awesome-icon :icon="['fas', 'arrow-up']" />
            </button>
            <button
              type="button"
              class="btn btn-outline-secondary"
              :disabled="i === options.relations.length - 1"
              :title="$t('relatedRecords.moveDown')"
              @click="move(i, 1)"
            >
              <font-awesome-icon :icon="['fas', 'arrow-down']" />
            </button>
          </div>
          <c-input-confirm
            show-icon
            :tooltip="$t('relatedRecords.remove')"
            @confirmed="options.relations.splice(i, 1)"
          />
        </div>

        <div class="card-body">
          <div class="row g-3">
            <div class="col-12 col-lg-8">
              <label class="form-label text-primary">{{ $t('relatedRecords.title.label') }}</label>
              <input
                v-model="relation.title"
                class="form-control"
                :placeholder="moduleOf(relation) ? moduleOf(relation).name : ''"
              >
            </div>

            <div class="col-12 col-lg-4">
              <label class="form-label text-primary">{{ $t('relatedRecords.perPage') }}</label>
              <input
                v-model.number="relation.perPage"
                type="number"
                min="1"
                max="100"
                class="form-control"
              >
            </div>

            <div class="col-12 col-lg-6">
              <div class="form-check form-switch">
                <input
                  :id="`rr-exp-${i}`"
                  v-model="relation.expanded"
                  class="form-check-input"
                  type="checkbox"
                >
                <label
                  class="form-check-label"
                  :for="`rr-exp-${i}`"
                >{{ $t('relatedRecords.expanded') }}</label>
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="form-check form-switch">
                <input
                  :id="`rr-add-${i}`"
                  v-model="relation.hideAddButton"
                  class="form-check-input"
                  type="checkbox"
                >
                <label
                  class="form-check-label"
                  :for="`rr-add-${i}`"
                >{{ $t('relatedRecords.hideAddButton') }}</label>
              </div>
            </div>

            <div
              v-if="moduleOf(relation)"
              class="col-12"
            >
              <label class="form-label text-primary">{{ $t('relatedRecords.columns') }}</label>
              <FieldPicker
                v-model:fields="relation.fields"
                :module="moduleOf(relation)"
                style="height: 40vh;"
              />
              <small class="text-muted">{{ $t('relatedRecords.columnsFootnote') }}</small>
            </div>
          </div>
        </div>
      </div>

      <hr>

      <div class="form-check form-switch mb-2">
        <input
          id="rr-counts"
          v-model="options.showCounts"
          class="form-check-input"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="rr-counts"
        >{{ $t('relatedRecords.showCounts') }}</label>
      </div>

      <div class="form-check form-switch mb-3">
        <input
          id="rr-hide-empty"
          v-model="options.hideEmptySections"
          class="form-check-input"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="rr-hide-empty"
        >{{ $t('relatedRecords.hideEmpty') }}</label>
        <div class="form-text">
          {{ $t('relatedRecords.hideEmptyFootnote') }}
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import { useStore } from '../../../store'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'
import { findReferencingFields, relationKey, hasRelation } from 'corteza-webapp-compose/src/lib/related-records'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  page: { type: Object, required: false, default: undefined },
})

const store = useStore()
const options = computed(() => props.block.options)
const candidateKey = ref('')

// the module of the record page the block is on
const pageModuleID = computed(() => props.module?.moduleID || props.page?.moduleID || '')

const fieldText = ({ label, name }) => label ? `${label} (${name})` : name

// the fields, in any module, that refer to the page's module and are not shown yet
const candidates = computed(() => {
  return findReferencingFields(store.module.set, pageModuleID.value)
    .filter(({ module, field }) => !hasRelation(options.value.relations, module.moduleID, field.name))
    .map(({ module, field }) => ({
      key: relationKey({ moduleID: module.moduleID, refField: field.name }),
      moduleID: module.moduleID,
      refField: field.name,
      text: `${module.name || module.handle} · ${fieldText(field)}`,
    }))
})

function moduleOf (relation) {
  return store.module.getByID(relation.moduleID)
}

function isUsable (relation) {
  const m = moduleOf(relation)
  return !!m && m.fields.some(f => f.name === relation.refField && f.kind === 'Record')
}

function relationText (relation) {
  const m = moduleOf(relation)
  if (!m) return `${relation.moduleID} · ${relation.refField}`
  const f = m.fields.find(f => f.name === relation.refField)
  return `${m.name || m.handle} · ${f ? fieldText(f) : relation.refField}`
}

function addRelation () {
  const c = candidates.value.find(c => c.key === candidateKey.value)
  if (!c) return

  const { relations } = compose.PageBlockMaker({
    kind: 'RelatedRecords',
    options: { relations: [{ moduleID: c.moduleID, refField: c.refField }] },
  }).options
  options.value.relations.push(relations[0])
  candidateKey.value = ''
}

function move (i, by) {
  const list = options.value.relations
  const j = i + by
  if (j < 0 || j >= list.length) return
  list.splice(j, 0, list.splice(i, 1)[0])
}
</script>
