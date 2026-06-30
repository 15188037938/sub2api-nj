<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">系统更新</h1>

      <!-- Version Info Card -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">版本信息</h2>
            <button
              type="button"
              class="btn btn-primary btn-sm"
              :disabled="checking"
              @click="checkUpdate"
            >
              <Icon
                name="refresh"
                size="sm"
                :class="checking ? 'animate-spin' : ''"
                class="mr-1 inline-block"
              />
              {{ checking ? '检查中...' : '检查更新' }}
            </button>
          </div>
        </div>
        <div class="space-y-4 p-6">
          <!-- Loading State -->
          <div v-if="loading" class="flex items-center justify-center py-8">
            <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
          </div>

          <!-- Data State -->
          <div v-else class="space-y-4">
            <div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800">
              <dl class="space-y-3">
                <div class="flex items-center justify-between">
                  <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">当前版本</dt>
                  <dd class="font-mono text-sm text-gray-900 dark:text-white">
                    {{ status?.current_commit || '--' }}
                  </dd>
                </div>
                <div class="flex items-center justify-between">
                  <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">构建时间</dt>
                  <dd class="text-sm text-gray-900 dark:text-white">
                    {{ status?.build_time || '--' }}
                  </dd>
                </div>
              </dl>
            </div>
          </div>
        </div>
      </div>

      <!-- Update Status Alert -->
      <div v-if="!loading && status" class="space-y-4">
        <!-- Update Available -->
        <div
          v-if="status.update_available"
          class="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20"
        >
          <div class="flex items-start">
            <Icon
              name="info"
              size="md"
              class="mt-0.5 flex-shrink-0 text-green-500"
            />
            <div class="ml-3 flex-1">
              <p class="text-sm font-medium text-green-800 dark:text-green-300">
                有可用更新
              </p>
              <div class="mt-2 space-y-1">
                <div class="flex items-center gap-2 text-sm">
                  <span class="text-gray-500 dark:text-gray-400">最新版本:</span>
                  <code class="font-mono text-green-700 dark:text-green-300">
                    {{ status.latest_commit }}
                  </code>
                </div>
                <div class="flex items-center gap-2 text-sm">
                  <span class="text-gray-500 dark:text-gray-400">当前版本:</span>
                  <code class="font-mono text-red-600 dark:text-red-400">
                    {{ status.current_commit }}
                  </code>
                </div>
              </div>
              <div class="mt-4">
                <button
                  type="button"
                  class="btn btn-warning"
                  :disabled="updating"
                  @click="doUpdate"
                >
                  <Icon
                    name="download"
                    size="sm"
                    :class="updating ? 'animate-spin' : ''"
                    class="mr-1 inline-block"
                  />
                  {{ updating ? '更新中...' : '立即更新' }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Up to Date -->
        <div
          v-else
          class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-gray-700 dark:bg-gray-800"
        >
          <div class="flex items-start">
            <Icon
              name="check"
              size="md"
              class="mt-0.5 flex-shrink-0 text-gray-400"
            />
            <p class="ml-3 text-sm text-gray-600 dark:text-gray-400">
              当前已是最新版本
            </p>
          </div>
        </div>
      </div>

      <!-- Update Progress -->
      <div
        v-if="updating"
        class="rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-900/20"
      >
        <div class="flex items-center space-x-3">
          <div class="h-5 w-5 animate-spin rounded-full border-b-2 border-blue-600"></div>
          <p class="text-sm text-blue-700 dark:text-blue-300">
            正在执行更新，请耐心等待...
          </p>
        </div>
      </div>

      <!-- Error Message -->
      <div
        v-if="error"
        class="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
      >
        <div class="flex items-start">
          <Icon
            name="exclamationTriangle"
            size="md"
            class="mt-0.5 flex-shrink-0 text-red-500"
          />
          <p class="ml-3 text-sm text-red-700 dark:text-red-300">{{ error }}</p>
        </div>
      </div>

      <!-- Update Result -->
      <div
        v-if="resultMessage"
        class="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20"
      >
        <div class="flex items-start">
          <Icon
            name="checkCircle"
            size="md"
            class="mt-0.5 flex-shrink-0 text-green-500"
          />
          <p class="ml-3 text-sm text-green-700 dark:text-green-300">
            {{ resultMessage }}
          </p>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUpdateStatus, applyUpdate, type UpdateStatus } from '@/api/update'

const loading = ref(true)
const checking = ref(false)
const updating = ref(false)
const error = ref('')
const resultMessage = ref('')
const status = ref<UpdateStatus | null>(null)

async function fetchStatus() {
  loading.value = true
  error.value = ''
  try {
    status.value = await getUpdateStatus()
  } catch (err: any) {
    error.value = err?.response?.data?.message || err?.message || '获取更新状态失败'
  } finally {
    loading.value = false
  }
}

async function checkUpdate() {
  checking.value = true
  error.value = ''
  resultMessage.value = ''
  try {
    status.value = await getUpdateStatus()
  } catch (err: any) {
    error.value = err?.response?.data?.message || err?.message || '检查更新失败'
  } finally {
    checking.value = false
  }
}

async function doUpdate() {
  updating.value = true
  error.value = ''
  resultMessage.value = ''
  try {
    const result = await applyUpdate()
    resultMessage.value = result.message || '更新请求已提交'
  } catch (err: any) {
    error.value = err?.response?.data?.message || err?.message || '更新失败'
  } finally {
    updating.value = false
  }
}

onMounted(() => {
  fetchStatus()
})
</script>
