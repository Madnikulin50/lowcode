<template>
  <div v-if="module">
    <h5 class="mb-3">{{ t('title') }}</h5>
    <p class="text-muted small">{{ t('description') }}</p>

    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ t('type') }}</label>
          <select
            v-model="cfg.type"
            class="form-select"
          >
            <option
              v-for="opt in typeOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </option>
          </select>
        </div>
      </div>
      <div class="col-12 col-lg-6 d-flex align-items-end">
        <div class="mb-3 w-100">
          <button
            class="btn btn-outline-primary btn-sm"
            type="button"
            :disabled="testing"
            @click="testConnection"
          >
            <span
              v-if="testing"
              class="spinner-border spinner-border-sm me-1"
            />
            {{ t('testConnection') }}
          </button>
          <span
            v-if="testResult === 'ok'"
            class="text-success small ms-2"
          >✓ {{ t('testOk') }}</span>
          <span
            v-if="testResult === 'fail'"
            class="text-danger small ms-2"
          >✗ {{ testError || t('testFail') }}</span>
        </div>
      </div>
    </div>

    <template v-if="isHTTPType">
      <div class="row">
        <div class="col-12 col-lg-8">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t(urlKey + '.url') }}</label>
            <input
              v-model="cfg.restUrl"
              class="form-control"
              type="text"
              :placeholder="t(urlKey + '.urlPlaceholder')"
            >
            <div
              v-if="cfg.type === 'elasticsearch'"
              class="form-text"
            >
              {{ t('elasticsearch.urlHint') }}
            </div>
          </div>
        </div>
        <div
          v-if="cfg.type !== 'elasticsearch'"
          class="col-12 col-lg-4"
        >
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('rest.method') }}</label>
            <select
              v-model="cfg.restMethod"
              class="form-select"
            >
              <option value="GET">
                GET
              </option>
              <option value="POST">
                POST
              </option>
              <option value="PUT">
                PUT
              </option>
            </select>
          </div>
        </div>
      </div>

      <div
        v-if="cfg.type === 'elasticsearch'"
        class="row"
      >
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('elasticsearch.index') }}</label>
            <input
              v-model="cfg.esIndex"
              class="form-control"
              type="text"
              :placeholder="t('elasticsearch.indexPlaceholder')"
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('elasticsearch.limitParam') }}</label>
            <input
              v-model="cfg.restLimitParam"
              class="form-control"
              type="text"
              placeholder="size"
            >
          </div>
        </div>
      </div>

      <div
        v-if="cfg.type === 'graphql'"
        class="row"
      >
        <div class="col-12 col-lg-8">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('graphql.query') }}</label>
            <textarea
              v-model="cfg.restBody"
              class="form-control font-monospace"
              rows="4"
              style="font-size: 0.8rem"
              :placeholder="t('graphql.queryPlaceholder')"
            />
          </div>
        </div>
        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('graphql.dataPath') }}</label>
            <input
              v-model="cfg.restDataPath"
              class="form-control"
              type="text"
              :placeholder="t('graphql.dataPathPlaceholder')"
            >
            <div class="form-text">{{ t('graphql.dataPathHint') }}</div>
          </div>
        </div>
      </div>

      <template v-if="cfg.type === 'rest'">
        <div class="row">
          <div class="col-12 col-lg-8">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('rest.dataPath') }}</label>
              <input
                v-model="cfg.restDataPath"
                class="form-control"
                type="text"
                :placeholder="t('rest.dataPathPlaceholder')"
              >
              <div class="form-text">{{ t('rest.dataPathHint') }}</div>
            </div>
          </div>
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('rest.body') }}</label>
              <textarea
                v-model="cfg.restBody"
                class="form-control font-monospace"
                rows="3"
                style="font-size: 0.8rem"
                :placeholder="t('rest.bodyPlaceholder')"
              />
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('rest.limitParam') }}</label>
              <input
                v-model="cfg.restLimitParam"
                class="form-control"
                type="text"
                placeholder="limit"
              >
            </div>
          </div>
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('rest.offsetParam') }}</label>
              <input
                v-model="cfg.restOffsetParam"
                class="form-control"
                type="text"
                placeholder="offset"
              >
            </div>
          </div>
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('rest.pageParam') }}</label>
              <input
                v-model="cfg.restPageParam"
                class="form-control"
                type="text"
                placeholder="page"
              >
            </div>
          </div>
        </div>
      </template>

      <template v-if="cfg.type === 'elasticsearch' || cfg.type === 'graphql'">
        <div class="row">
          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t(cfg.type + '.query') }}</label>
              <textarea
                v-model="cfg.restBody"
                class="form-control font-monospace"
                rows="4"
                style="font-size: 0.8rem"
                :placeholder="t(cfg.type + '.queryPlaceholder')"
              />
              <div class="form-text">{{ t(cfg.type + '.queryHint') }}</div>
            </div>
          </div>
        </div>
      </template>

      <div class="row">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('rest.headers') }}</label>
            <div
              v-for="(_, key) in cfg.restHeaders"
              :key="'hdr-'+key"
              class="input-group mb-1"
            >
              <input
                :value="key"
                class="form-control font-monospace"
                style="font-size: 0.8rem; max-width: 30%"
                :placeholder="t('rest.headerName')"
                @input="onHeaderKey(key, $event.target.value)"
              >
              <input
                :value="cfg.restHeaders[key]"
                class="form-control font-monospace"
                style="font-size: 0.8rem"
                :placeholder="t('rest.headerValue')"
                @input="onHeaderValue(key, $event.target.value)"
              >
              <button
                class="btn btn-outline-danger"
                type="button"
                @click="removeHeader(key)"
              >
                ×
              </button>
            </div>
            <button
              class="btn btn-outline-secondary btn-sm"
              type="button"
              @click="addHeader"
            >
              + {{ t('rest.addHeader') }}
            </button>
          </div>
        </div>
      </div>
    </template>

    <template v-if="cfg.type === 'mongodb'">
      <div class="alert alert-info small">
        {{ $t('connector.notImplemented', { type: 'MongoDB' }) }}
      </div>
      <div class="row">
        <div class="col-12 col-lg-8">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('mongodb.host') }}</label>
            <input
              v-model="cfg.mongoHost"
              class="form-control"
              type="text"
              :placeholder="t('mongodb.hostPlaceholder')"
            >
          </div>
        </div>
        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('mongodb.port') }}</label>
            <input
              v-model.number="cfg.mongoPort"
              class="form-control"
              type="number"
              placeholder="27017"
            >
          </div>
        </div>
      </div>
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('mongodb.database') }}</label>
            <input
              v-model="cfg.mongoDb"
              class="form-control"
              type="text"
              placeholder="mydb"
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('mongodb.collection') }}</label>
            <input
              v-model="cfg.mongoColl"
              class="form-control"
              type="text"
              :placeholder="t('mongodb.collectionPlaceholder')"
            >
          </div>
        </div>
      </div>
      <div class="row">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('mongodb.query') }}</label>
            <textarea
              v-model="cfg.mongoQuery"
              class="form-control font-monospace"
              rows="3"
              style="font-size: 0.8rem"
              :placeholder="t('mongodb.queryPlaceholder')"
            />
            <div class="form-text">{{ t('mongodb.queryHint') }}</div>
          </div>
        </div>
      </div>
    </template>

    <template v-if="cfg.type === 'kafka'">
      <div class="alert alert-info small">
        {{ $t('connector.notImplemented', { type: 'Kafka' }) }}
      </div>
      <div class="row">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kafka.brokers') }}</label>
            <input
              v-model="cfg.kafkaBrokers"
              class="form-control"
              type="text"
              :placeholder="t('kafka.brokersPlaceholder')"
            >
          </div>
        </div>
      </div>
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kafka.topic') }}</label>
            <input
              v-model="cfg.kafkaTopic"
              class="form-control"
              type="text"
              :placeholder="t('kafka.topicPlaceholder')"
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kafka.group') }}</label>
            <input
              v-model="cfg.kafkaGroup"
              class="form-control"
              type="text"
              :placeholder="t('kafka.groupPlaceholder')"
            >
          </div>
        </div>
      </div>
      <p class="text-muted small">{{ t('kafka.hint') }}</p>
    </template>

    <template v-if="cfg.type === 'redis'">
      <div class="alert alert-info small">
        {{ $t('connector.notImplemented', { type: 'Redis' }) }}
      </div>
      <div class="row">
        <div class="col-12 col-lg-8">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('redis.host') }}</label>
            <input
              v-model="cfg.redisHost"
              class="form-control"
              type="text"
              :placeholder="t('redis.hostPlaceholder')"
            >
          </div>
        </div>
        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('redis.port') }}</label>
            <input
              v-model.number="cfg.redisPort"
              class="form-control"
              type="number"
              placeholder="6379"
            >
          </div>
        </div>
      </div>
      <div class="row">
        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('redis.password') }}</label>
            <input
              v-model="cfg.redisPass"
              class="form-control"
              type="password"
            >
          </div>
        </div>
        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('redis.keyPattern') }}</label>
            <input
              v-model="cfg.redisKey"
              class="form-control"
              type="text"
              :placeholder="t('redis.keyPatternPlaceholder')"
            >
            <div class="form-text">{{ t('redis.keyPatternHint') }}</div>
          </div>
        </div>
        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('redis.database') }}</label>
            <input
              v-model.number="cfg.redisDb"
              class="form-control"
              type="number"
              placeholder="0"
            >
          </div>
        </div>
      </div>
      <p class="text-muted small">{{ t('redis.hint') }}</p>
    </template>

    <template v-if="cfg.type === 'grpc'">
      <div class="alert alert-info small">
        {{ $t('connector.notImplemented', { type: 'gRPC' }) }}
      </div>
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('grpc.address') }}</label>
            <input
              v-model="cfg.grpcAddr"
              class="form-control"
              type="text"
              :placeholder="t('grpc.addressPlaceholder')"
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('grpc.method') }}</label>
            <input
              v-model="cfg.grpcMethod"
              class="form-control"
              type="text"
              :placeholder="t('grpc.methodPlaceholder')"
            >
          </div>
        </div>
      </div>
      <div class="row">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('grpc.payload') }}</label>
            <textarea
              v-model="cfg.grpcPayload"
              class="form-control font-monospace"
              rows="3"
              style="font-size: 0.8rem"
              :placeholder="t('grpc.payloadPlaceholder')"
            />
            <div class="form-text">{{ t('grpc.payloadHint') }}</div>
          </div>
        </div>
      </div>
      <p class="text-muted small">{{ t('grpc.hint') }}</p>
    </template>

    <template v-if="cfg.type === 'db'">
      <div class="row">
        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('db.driver') }}</label>
            <select
              v-model="cfg.dbDriver"
              class="form-select"
            >
              <option value="postgres">
                PostgreSQL
              </option>
              <option value="mysql">
                MySQL
              </option>
              <option value="sqlite3">
                SQLite
              </option>
              <option value="sqlserver">
                MS SQL Server
              </option>
            </select>
          </div>
        </div>
        <div class="col-12 col-lg-8">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('db.connectionString') }}</label>
            <input
              v-model="cfg.dbConnectionString"
              class="form-control font-monospace"
              type="text"
              style="font-size: 0.8rem"
              :placeholder="t('db.connectionStringPlaceholder')"
            >
            <div class="form-text">{{ t('db.connectionStringHint') }}</div>
          </div>
        </div>
      </div>
      <div class="row">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('db.query') }}</label>
            <textarea
              v-model="cfg.dbQuery"
              class="form-control font-monospace"
              rows="4"
              style="font-size: 0.8rem"
              :placeholder="t('db.queryPlaceholder')"
            />
            <div class="form-text">{{ t('db.queryHint') }}</div>
          </div>
        </div>
      </div>
    </template>

    <hr class="my-4">

    <div class="d-flex align-items-center mb-2">
      <span class="fw-bold">{{ t('fieldMapping.title') }}</span>
      <button
        class="btn btn-outline-secondary btn-sm ms-2"
        type="button"
        @click="addFieldMapping"
      >
        + {{ t('fieldMapping.add') }}
      </button>
    </div>
    <p class="text-muted small mb-3">{{ t('fieldMapping.hint') }}</p>

    <div
      v-if="cfg.fieldMapping && cfg.fieldMapping.length"
      class="table-responsive"
    >
      <table class="table table-sm table-borderless align-middle">
        <thead>
          <tr class="text-muted small">
            <th>{{ t('fieldMapping.moduleField') }}</th>
            <th>{{ t('fieldMapping.sourceKey') }}</th>
            <th style="width: 30px" />
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(_, idx) in cfg.fieldMapping"
            :key="idx"
          >
            <td>
              <select
                v-model="cfg.fieldMapping[idx].field"
                class="form-select form-select-sm"
              >
                <option value="">
                  -- {{ t('fieldMapping.selectField') }} --
                </option>
                <option
                  v-for="f in module.fields"
                  :key="f.name"
                  :value="f.name"
                >
                  {{ f.name }}{{ f.label ? ' (' + f.label + ')' : '' }}
                </option>
              </select>
            </td>
            <td>
              <input
                v-model="cfg.fieldMapping[idx].source"
                class="form-control form-control-sm font-monospace"
                type="text"
                :placeholder="t('fieldMapping.sourceKeyPlaceholder')"
              >
            </td>
            <td>
              <button
                class="btn btn-outline-danger btn-sm"
                type="button"
                @click="removeFieldMapping(idx)"
              >
                ×
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div
      v-else
      class="text-muted small py-2"
    >
      {{ t('fieldMapping.empty') }}
    </div>
  </div>
</template>

<script setup>
import { computed, ref, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })
const t = (key) => $t('connector.' + key, { default: key })

const { $ComposeAPI } = getCurrentInstance().appContext.config.globalProperties

const testing = ref(false)
const testResult = ref(null)
const testError = ref('')

const props = defineProps({
  module: {
    type: Object,
    required: true,
  },
})

const typeOptions = [
  { value: 'rest', labelKey: 'connector.types.rest' },
  { value: 'graphql', labelKey: 'connector.types.graphql' },
  { value: 'elasticsearch', labelKey: 'connector.types.elasticsearch' },
  { value: 'mongodb', labelKey: 'connector.types.mongodb' },
  { value: 'kafka', labelKey: 'connector.types.kafka' },
  { value: 'redis', labelKey: 'connector.types.redis' },
  { value: 'grpc', labelKey: 'connector.types.grpc' },
  { value: 'db', labelKey: 'connector.types.db' },
].map(o => ({ ...o, label: $t(o.labelKey, { default: o.value }) }))

const httpTypes = ['rest', 'graphql', 'elasticsearch']

const cfg = computed(() => {
  if (!props.module.config.connector) {
    props.module.config.connector = {
      type: 'rest',
      restMethod: 'GET',
      restHeaders: {},
      fieldMapping: [],
    }
  }
  return props.module.config.connector
})

const isHTTPType = computed(() => httpTypes.includes(cfg.value.type))

const urlKey = computed(() => {
  switch (cfg.value.type) {
  case 'elasticsearch': return 'elasticsearch'
  case 'graphql': return 'graphql'
  default: return 'rest'
  }
})

function testConnection () {
  testing.value = true
  testResult.value = null
  testError.value = ''
  $ComposeAPI.connectorTest({ connector: cfg.value })
    .then((r) => {
      if (r.success) {
        testResult.value = 'ok'
      } else {
        testResult.value = 'fail'
        testError.value = r.error || ''
      }
    })
    .catch((e) => {
      testResult.value = 'fail'
      testError.value = e.message || String(e)
    })
    .finally(() => {
      testing.value = false
    })
}

function addHeader () {
  cfg.value.restHeaders = { ...cfg.value.restHeaders, '': '' }
}

function removeHeader (key) {
  const headers = { ...cfg.value.restHeaders }
  delete headers[key]
  cfg.value.restHeaders = headers
}

function onHeaderKey (oldKey, newKey) {
  const headers = { ...cfg.value.restHeaders }
  const value = headers[oldKey]
  delete headers[oldKey]
  if (newKey) headers[newKey] = value
  cfg.value.restHeaders = headers
}

function onHeaderValue (key, value) {
  cfg.value.restHeaders = { ...cfg.value.restHeaders, [key]: value }
}

function addFieldMapping () {
  cfg.value.fieldMapping = [...(cfg.value.fieldMapping || []), { field: '', source: '' }]
}

function removeFieldMapping (idx) {
  const fm = [...(cfg.value.fieldMapping || [])]
  fm.splice(idx, 1)
  cfg.value.fieldMapping = fm
}
</script>
