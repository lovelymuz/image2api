<script setup>
// App shell is layout-driven: each top-level route renders its own layout
// (PublicLayout for /, /user; AdminLayout for /admin/*). The login modal is
// mounted here so it can overlay any page instead of being a separate route.
import { onMounted } from 'vue'
import LoginModal from './components/LoginModal.vue'
import { auth, refreshMe, openRegister } from './auth'

// An invite link (/?ref=CODE) should drop a guest straight into registration
// with the code attached. Logged-in users just ignore the ref.
onMounted(async () => {
  const code = new URLSearchParams(location.search).get('ref')
  if (!code) return
  if (!auth.ready) await refreshMe()
  if (!auth.token || !auth.user) openRegister(code)
})
</script>

<template>
  <router-view />
  <LoginModal />
</template>
