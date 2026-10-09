<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.actionlog.list.title')" />
    <c-actionlog-list />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.actionlog', keyPrefix: 'list' } })
import CActionlogList from '../../../components/Actionlog/CActionlogList.vue'
</script>
