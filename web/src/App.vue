<script setup lang="ts">
import AppLayout from '@/layouts/AppLayout.vue'
import { useUIStore } from '@/stores/ui'

const ui = useUIStore()
</script>

<template>
  <AppLayout />
</template>

<style>
@import 'tailwindcss/base';
@import 'tailwindcss/components';
@import 'tailwindcss/utilities';

:root {
  --vt-c-white: #ffffff;
  --vt-c-white-soft: #f8f8f8;
  --vt-c-white-mute: #f2f2f2;

  --vt-c-black: #213547;
  --vt-c-black-soft: #222222;
  --vt-c-black-mute: #313131;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Roboto', system-ui, -apple-system, sans-serif;
}
</style>
