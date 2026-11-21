<template>
  <div class="d-flex flex-column vh-100">
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark px-3">
      <router-link to="/" class="navbar-brand d-flex align-items-center gap-2">
        <i class="bi bi-hdd-network fs-4"></i>
        CMDB Discovery Agent
      </router-link>
      <div class="navbar-nav">
        <router-link to="/" class="nav-link" active-class="active">
          <i class="bi bi-speedometer2"></i> Dashboard
        </router-link>
        <router-link to="/devices" class="nav-link" active-class="active">
          <i class="bi bi-devices"></i> Devices
        </router-link>
      </div>
    </nav>
    <div class="flex-grow-1 overflow-auto p-3 bg-light">
      <router-view />
    </div>
  </div>
</template>
