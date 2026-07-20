<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import DictionaryTab from './components/DictionaryTab.vue'
import { useCourseDictionaryView } from './composables/useCourseDictionaryView'

const { t } = useI18n()
const {
  activeTab,
  createLabel,
  currentColumns,
  currentFormFields,
  currentState,
  displayItems,
  formError,
  saving,
  activateTab,
  changePage,
  handleCreate,
  handleEdit,
  handleToggle,
  initialize,
  retryActiveTab,
} = useCourseDictionaryView()

onMounted(() => {
  void initialize()
})
</script>

<template>
  <div
    class="course-dict-view"
    data-testid="course-dict-view"
  >
    <h1>{{ t('courses.title') }}</h1>

    <div
      class="tabs"
      role="tablist"
    >
      <button
        role="tab"
        :aria-selected="activeTab === 'domains'"
        :class="{ active: activeTab === 'domains' }"
        data-testid="tab-domains"
        @click="activateTab('domains')"
      >
        {{ t('courses.domains') }}
      </button>
      <button
        role="tab"
        :aria-selected="activeTab === 'tracks'"
        :class="{ active: activeTab === 'tracks' }"
        data-testid="tab-tracks"
        @click="activateTab('tracks')"
      >
        {{ t('courses.tracks') }}
      </button>
      <button
        role="tab"
        :aria-selected="activeTab === 'levels'"
        :class="{ active: activeTab === 'levels' }"
        data-testid="tab-levels"
        @click="activateTab('levels')"
      >
        {{ t('courses.levels') }}
      </button>
      <button
        role="tab"
        :aria-selected="activeTab === 'tags'"
        :class="{ active: activeTab === 'tags' }"
        data-testid="tab-tags"
        @click="activateTab('tags')"
      >
        {{ t('courses.tags') }}
      </button>
    </div>

    <DictionaryTab
      :items="displayItems"
      :columns="currentColumns"
      :form-fields="currentFormFields"
      :loading="currentState.loading"
      :error="currentState.error"
      :page="currentState.page"
      :page-size="currentState.pageSize"
      :total="currentState.total"
      :can-create="!saving && !currentState.loading"
      :saving="saving"
      :form-error="formError"
      :create-label="createLabel"
      :edit-label="createLabel"
      @create="handleCreate"
      @edit="(item, body) => handleEdit(item, body)"
      @toggle="handleToggle"
      @page-change="changePage"
      @retry="retryActiveTab"
    />
  </div>
</template>

<style scoped>
.course-dict-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 1rem;
}

.tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
  border-bottom: 2px solid #eee;
}

.tabs button {
  padding: 0.5rem 1rem;
  border: none;
  border-bottom: 2px solid transparent;
  background-color: transparent;
  cursor: pointer;
}

.tabs button.active {
  border-bottom-color: #0d6efd;
  color: #0d6efd;
  font-weight: 600;
}

h1 {
  margin: 0 0 1rem 0;
}
</style>
