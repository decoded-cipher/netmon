<template>
  <Dashboard :is-dark="isDark" @toggle-theme="toggleTheme" />
  <ToastContainer :toasts="toasts" @dismiss="dismissToast" />
</template>

<script setup>
import { ref } from 'vue'
import Dashboard      from './components/Dashboard.vue'
import ToastContainer from './components/ToastContainer.vue'
import { useToasts }  from './composables/useToasts.js'

const { toasts, dismiss: dismissToast } = useToasts()

const isDark = ref(localStorage.getItem('theme') === 'dark')

function applyTheme(dark) {
  document.body.classList.toggle('dark-mode', dark)
}
applyTheme(isDark.value)

function toggleTheme() {
  isDark.value = !isDark.value
  applyTheme(isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>
