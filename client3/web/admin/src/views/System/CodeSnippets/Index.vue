<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('system.code-snippets.list.title')" />
    <c-code-snippets-list />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.code-snippets', keyPrefix: 'editor' } })
import CCodeSnippetsList from '../../../components/CodeSnippets/CCodeSnippetsList.vue'
</script>
