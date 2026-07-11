<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { SUPPORTED_LOCALES, type Locale } from '../i18n/config'

const { locale, t } = useI18n()

function changeLocale(newLocale: string): void {
  if (SUPPORTED_LOCALES.includes(newLocale as Locale)) {
    locale.value = newLocale as Locale
  }
}
</script>

<template>
  <div
    class="locale-switcher"
    data-testid="locale-switcher"
  >
    <label for="locale-select">{{ t('common.localeLabel') }}: </label>
    <select
      id="locale-select"
      :value="locale"
      @change="changeLocale(($event.target as HTMLSelectElement).value)"
    >
      <option
        v-for="loc in SUPPORTED_LOCALES"
        :key="loc"
        :value="loc"
      >
        {{ loc }}
      </option>
    </select>
  </div>
</template>

<style scoped>
.locale-switcher {
  margin: 0.5rem 0;
}

select {
  padding: 0.25rem 0.5rem;
  border: 1px solid #ddd;
  border-radius: 0.25rem;
}
</style>
