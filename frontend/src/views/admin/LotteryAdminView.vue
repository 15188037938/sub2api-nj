<template>
  <div class="p-6 space-y-6">
    <h2 class="text-2xl font-bold">签到抽奖管理</h2>

    <!-- 签到配置 -->
    <section class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h3 class="text-lg font-semibold mb-4">签到配置</h3>
      <div v-if="config" class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium mb-1">每日最低积分</label>
          <input v-model.number="configForm.daily_min_points" type="number" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">每日最高积分</label>
          <input v-model.number="configForm.daily_max_points" type="number" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">抽奖消耗积分</label>
          <input v-model.number="configForm.lottery_cost" type="number" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">每日最多抽奖次数</label>
          <input v-model.number="configForm.daily_max_draws" type="number" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
        </div>
        <div class="md:col-span-2">
          <label class="flex items-center gap-2">
            <input v-model="configForm.enabled" type="checkbox" class="w-4 h-4" />
            <span class="text-sm font-medium">启用签到抽奖功能</span>
          </label>
        </div>
        <div class="md:col-span-2">
          <button @click="saveConfig" :disabled="savingConfig" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-lg transition">
            {{ savingConfig ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </div>
    </section>

    <!-- 奖品管理 -->
    <section class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold">奖品管理</h3>
        <button @click="showPrizeForm = true" class="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg transition text-sm">
          添加奖品
        </button>
      </div>

      <!-- 奖品表格 -->
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b dark:border-gray-700">
              <th class="text-left py-2 px-3">名称</th>
              <th class="text-left py-2 px-3">类型</th>
              <th class="text-left py-2 px-3">数值</th>
              <th class="text-left py-2 px-3">权重</th>
              <th class="text-left py-2 px-3">库存</th>
              <th class="text-left py-2 px-3">状态</th>
              <th class="text-left py-2 px-3">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="prize in prizes" :key="prize.id" class="border-b dark:border-gray-700">
              <td class="py-2 px-3">{{ prize.name }}</td>
              <td class="py-2 px-3">
                <span :class="typeBadgeClass(prize.prize_type)">{{ typeLabel(prize.prize_type) }}</span>
              </td>
              <td class="py-2 px-3">{{ prize.amount }}</td>
              <td class="py-2 px-3">{{ prize.weight }}</td>
              <td class="py-2 px-3">{{ prize.remaining_stock === -1 ? '无限' : prize.remaining_stock }}</td>
              <td class="py-2 px-3">
                <span :class="prize.status === 'active' ? 'text-green-600' : 'text-gray-400'">
                  {{ prize.status === 'active' ? '启用' : '禁用' }}
                </span>
              </td>
              <td class="py-2 px-3 space-x-1">
                <button @click="editPrize(prize)" class="text-blue-600 hover:text-blue-800 text-xs">编辑</button>
                <button @click="deletePrize(prize.id)" class="text-red-600 hover:text-red-800 text-xs">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 添加/编辑奖品弹窗 -->
      <div v-if="showPrizeForm" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" @click.self="closePrizeForm">
        <div class="bg-white dark:bg-gray-800 rounded-2xl p-6 max-w-md w-full mx-4 shadow-xl">
          <h4 class="text-lg font-bold mb-4">{{ editingPrize ? '编辑奖品' : '添加奖品' }}</h4>
          <div class="space-y-3">
            <div>
              <label class="block text-sm font-medium mb-1">奖品名称</label>
              <input v-model="prizeForm.name" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" placeholder="如: 1元余额" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">奖品类型</label>
              <select v-model="prizeForm.prize_type" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
                <option value="balance">余额</option>
                <option value="points">积分</option>
                <option value="none">谢谢参与</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">数值</label>
              <input v-model.number="prizeForm.amount" type="number" step="0.01" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">权重（越大越容易中）</label>
              <input v-model.number="prizeForm.weight" type="number" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">总库存（-1=无限）</label>
              <input v-model.number="prizeForm.total_stock" type="number" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">排序</label>
              <input v-model.number="prizeForm.sort_order" type="number" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="closePrizeForm" class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm">取消</button>
            <button @click="savePrize" :disabled="savingPrize" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-lg text-sm">
              {{ savingPrize ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface Prize {
  id: number
  name: string
  prize_type: string
  amount: number
  weight: number
  total_stock: number
  remaining_stock: number
  icon: string
  status: string
  sort_order: number
}

interface CheckInConfig {
  id: number
  daily_min_points: number
  daily_max_points: number
  lottery_cost: number
  daily_max_draws: number
  enabled: boolean
}

const config = ref<CheckInConfig | null>(null)
const configForm = ref({
  daily_min_points: 1,
  daily_max_points: 10,
  lottery_cost: 10,
  daily_max_draws: 5,
  enabled: true,
})

const prizes = ref<Prize[]>([])
const showPrizeForm = ref(false)
const editingPrize = ref<Prize | null>(null)
const prizeForm = ref({
  name: '',
  prize_type: 'balance',
  amount: 0,
  weight: 10,
  total_stock: -1,
  sort_order: 0,
  icon: '',
})

const savingConfig = ref(false)
const savingPrize = ref(false)

function typeLabel(type: string) {
  const map: Record<string, string> = { balance: '余额', points: '积分', none: '谢谢参与' }
  return map[type] || type
}

function typeBadgeClass(type: string) {
  const map: Record<string, string> = {
    balance: 'text-green-600 bg-green-100 dark:bg-green-900/30 px-2 py-0.5 rounded text-xs',
    points: 'text-blue-600 bg-blue-100 dark:bg-blue-900/30 px-2 py-0.5 rounded text-xs',
    none: 'text-gray-500 bg-gray-100 dark:bg-gray-700 px-2 py-0.5 rounded text-xs',
  }
  return map[type] || ''
}

async function fetchConfig() {
  try {
    const { data } = await api.get('/admin/checkin/config')
    config.value = data
    configForm.value = {
      daily_min_points: data.daily_min_points,
      daily_max_points: data.daily_max_points,
      lottery_cost: data.lottery_cost,
      daily_max_draws: data.daily_max_draws,
      enabled: data.enabled,
    }
  } catch (e) {
    console.error('Failed to fetch config', e)
  }
}

async function fetchPrizes() {
  try {
    const { data } = await api.get('/admin/lottery/prizes')
    prizes.value = data || []
  } catch (e) {
    console.error('Failed to fetch prizes', e)
  }
}

async function saveConfig() {
  savingConfig.value = true
  try {
    await api.put('/admin/checkin/config', configForm.value)
    await fetchConfig()
  } catch (e: any) {
    alert(e?.response?.data?.message || '保存失败')
  } finally {
    savingConfig.value = false
  }
}

function editPrize(prize: Prize) {
  editingPrize.value = prize
  prizeForm.value = {
    name: prize.name,
    prize_type: prize.prize_type,
    amount: prize.amount,
    weight: prize.weight,
    total_stock: prize.total_stock,
    sort_order: prize.sort_order,
    icon: prize.icon || '',
  }
  showPrizeForm.value = true
}

function closePrizeForm() {
  showPrizeForm.value = false
  editingPrize.value = null
  prizeForm.value = {
    name: '',
    prize_type: 'balance',
    amount: 0,
    weight: 10,
    total_stock: -1,
    sort_order: 0,
    icon: '',
  }
}

async function savePrize() {
  savingPrize.value = true
  try {
    if (editingPrize.value) {
      await api.put(`/admin/lottery/prizes/${editingPrize.value.id}`, prizeForm.value)
    } else {
      await api.post('/admin/lottery/prizes', prizeForm.value)
    }
    closePrizeForm()
    await fetchPrizes()
  } catch (e: any) {
    alert(e?.response?.data?.message || '保存失败')
  } finally {
    savingPrize.value = false
  }
}

async function deletePrize(id: number) {
  if (!confirm('确定删除这个奖品吗？')) return
  try {
    await api.delete(`/admin/lottery/prizes/${id}`)
    await fetchPrizes()
  } catch (e: any) {
    alert(e?.response?.data?.message || '删除失败')
  }
}

onMounted(() => {
  fetchConfig()
  fetchPrizes()
})
</script>
