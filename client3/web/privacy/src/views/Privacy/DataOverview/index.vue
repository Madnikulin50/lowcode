<template>
  <div class="container-fluid d-flex flex-column p-3">
    <Teleport to="#topbar-title-target">{{ t('title') }}</Teleport>

    <div v-for="type in dataTypes" :key="type.title" class="row">
      <div class="col-12 col-lg-6">
        <div class="row">
          <div class="col mb-3">
            <div class="card card-hover-popup shadow-sm pointer">
              <div class="row g-0">
                <div class="col-2 align-self-center p-2 text-center">
                  <font-awesome-icon :icon="type.icon" class="text-primary h2 mb-0" />
                </div>
                <div class="col-9">
                  <div class="card-body px-2">
                    <h5 class="card-title">{{ type.title }}</h5>
                    <p class="card-text">{{ type.description }}</p>
                  </div>
                </div>
                <div class="col-1 align-self-center">
                  <font-awesome-icon :icon="['fas', 'chevron-right']" />
                </div>
              </div>

              <a
                v-if="type.href"
                :href="type.href"
                target="_blank"
                class="pointer stretched-link"
              />

              <router-link
                v-else-if="type.to"
                :to="type.to"
                class="pointer stretched-link"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'data-overview' } })

import { useAuth, useNsI18n } from 'corteza-lib/vue/dist'

const t = useNsI18n()
const { auth } = useAuth()

const dataTypes = [
  {
    title: t('data-types.profile-information.title'),
    description: t('data-types.profile-information.description'),
    icon: ['far', 'user'],
    href: auth.cortezaAuthURL,
  },
  {
    title: t('data-types.application-data.title'),
    description: t('data-types.application-data.description'),
    icon: ['fas', 'th-large'],
    to: { name: 'data-overview.application' },
  },
]
</script>