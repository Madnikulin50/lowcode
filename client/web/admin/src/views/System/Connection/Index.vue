<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.connections.list.title')" />
    <div class="d-flex flex-column h-100">
      <c-connection-primary class="mb-4" />
      <c-external-connection-list class="flex-fill" />
    </div>
  </div>
</template>
<script setup>
import CConnectionPrimary from '../../../components/Connection/CConnectionPrimary.vue'
import CExternalConnectionList from '../../../components/Connection/CExternalConnectionList.vue'
</script>
