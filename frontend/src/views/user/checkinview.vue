<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
    <div class="max-w-3xl mx-auto space-y-6">
      <!-- 签到区域 -->
      <section class="bg-white dark:bg-gray-800 rounded-2xl shadow p-6">
        <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">每日签到</h2>
        <div v-if="status.checked_in" class="text-center py-4">
          <div class="text-green-500 text-6xl mb-3">&#10003;</div>
          <p class="text-gray-600 dark:text-gray-300 text-lg">今日已签到</p>
          <p class="text-gray-400 dark:text-gray-500 mt-2">
            获得 <span class="text-yellow-500 font-bold">{{ status.today_points }}</span> 积分
          </p>
          <p class="text-gray-400 dark:text-gray-500">
            连续签到 <span class="font-bold">{{ status.consecutive_days }}</span> 天 | 总积分 <span class="font-bold">{{ status.total_points }}</span>
          </p>
        </div>
        <div v-else class="text-center py-4">
          <button
            @click="doCheckIn"
            :disabled="loading"
            class="px-8 py-3 bg-yellow-500 hover:bg-yellow-600 disabled:bg-gray-400 text-white font-bold rounded-xl text-lg transition"
          >
            {{ loading ? '签到中...' : '签到领积分' }}
          </button>
          <p class="text-gray-400 dark:text-gray-500 mt-3">
            连续签到 <span class="font-bold">{{ status.consecutive_days }}</span> 天 | 总积分 <span class="font-bold">{{ status.total_points }}</span>
          </p>
          <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">连续签到3/7/30天有额外加成</p>
        </div>
      </section>

      <!-- 抽奖区域 -->
      <section class="bg-white dark:bg-gray-800 rounded-2xl shadow p-6">
        <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">幸运转盘</h2>
        <p class="text-gray-500 dark:text-gray-400 mb-4">
          单次抽奖消耗 <span class="font-bold text-yellow-500">{{ config.lottery_cost }}</span> 积分 |
          今日已抽 <span class="font-bold">{{ config.today_draws }}</span>/{{ config.daily_max_draws }} 次
        </p>

        <!-- 奖品展示 -->
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3 mb-6" v-if="prizes.length > 0">
          <div
            v-for="prize in prizes"
            :key="prize.id"
            class="border border-gray-200 dark:border-gray-700 rounded-xl p-3 text-center"
          >
            <div class="text-2xl mb-1">{{ prize.icon || '🎁' }}</div>
            <div class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ prize.name }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ prize.prize_type === 'balance' ? '余额' : prize.prize_type === 'points' ? '积分' : '-' }}</div>
          </div>
        </div>

        <div class="text-center">
          <button
            @click="drawLottery"
            :disabled="drawLoading || !canDraw"
            class="px-8 py-3 bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 disabled:from-gray-400 disabled:to-gray-400 text-white font-bold rounded-xl text-lg transition"
          >
            {{ drawLoading ? '抽奖中...' : '抽奖' }}
          </button>
          <p v-if="!canDraw && !drawLoading" class="text-sm text-gray-400 dark:text-gray-500 mt-2">
            {{ config.today_draws >= config.daily_max_draws ? '今日次数已用完' : status.total_points < config.lottery_cost ? '积分不足' : '' }}
          </p>
        </div>

        <!-- 中奖弹窗 -->
        <div v-if="showResult" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" @click.self="showResult = false">
          <div class="bg-white dark:bg-gray-800 rounded-2xl p-8 max-w-sm w-full mx-4 text-center shadow-xl">
            <div class="text-6xl mb-4">{{ lastResult.prize_type === 'none' ? '😅' : '🎉' }}</div>
            <h3 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
              {{ lastResult.prize_type === 'none' ? '谢谢参与' : '恭喜中奖！' }}
            </h3>
            <p class="text-gray-600 dark:text-gray-300 text-lg mb-6">
              {{ lastResult.prize_name }}
            </p>
            <button
              @click="showResult = false"
              class="w-full py-3 bg-yellow-500 hover:bg-yellow-600 text-white font-bold rounded-xl transition"
            >
              知道了
            </button>
          </div>
        </div>
      </section>

      <!-- 抽奖历史 -->
      <section class="bg-white dark:bg-gray-800 rounded-2xl shadow p-6">
        <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">抽奖记录</h2>
        <div v-if="history.length === 0" class="text-center text-gray-400 dark:text-gray-500 py-4">
          暂无抽奖记录
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="item in history"
            :key="item.id"
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg"
          >
            <div>
              <div class="font-medium text-gray-800 dark:text-gray-200">{{ item.prize_name }}</div>
              <div class="text-xs text-gray-400 dark:text-gray-500">{{ item.created_at }}</div>
            </div>
            <span
              :class="[
                'px-2 py-1 rounded text-xs font-medium',
                item.claimed ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
              ]"
            >
              {{ item.claimed ? '已领取' : '待领取' }}
            </span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '@/api'

interface CheckInStatus {
  checked_in: boolean
  consecutive_days: number
  total_points: number
  today_points: number
  enabled: boolean
}

interface LotteryConfig {
  config: any
  prizes: Array<{
    id: number
    name: string
    prize_type: string
    amount: number
    weight: number
    remaining_stock: number
    icon: string
    status: string
  }>
  today_draws: number
  daily_max_draws: number
  lottery_cost: number
}

interface LotteryResult {
  id: number
  prize_id: number
  prize_name: string
  prize_type: string
  amount: number
  cost_points: number
  claimed: boolean
  created_at: string
}

const status = ref<CheckInStatus>({
  checked_in: false,
  consecutive_days: 0,
  total_points: 0,
  today_points: 0,
  enabled: true,
})

const config = ref({
  lottery_cost: 10,
  today_draws: 0,
  daily_max_draws: 5,
})

const prizes = ref<LotteryConfig['prizes']>([])
const history = ref<LotteryResult[]>([])
const loading = ref(false)
const drawLoading = ref(false)
const showResult = ref(false)
const lastResult = ref<LotteryResult>({
  id: 0,
  prize_id: 0,
  prize_name: '',
  prize_type: 'none',
  amount: 0,
  cost_points: 0,
  claimed: false,
  created_at: '',
})

const canDraw = computed(() => {
  return status.value.total_points >= config.value.lottery_cost &&
    config.value.today_draws < config.value.daily_max_draws
})

async function fetchStatus() {
  try {
    const { data } = await api.get('/user/checkin/status')
    status.value = data
  } catch (e) {
    console.error('Failed to fetch checkin status', e)
  }
}

async function fetchConfig() {
  try {
    const { data } = await api.get('/lottery/config')
    config.value = data
    prizes.value = data.prizes || []
  } catch (e) {
    console.error('Failed to fetch lottery config', e)
  }
}

async function fetchHistory() {
  try {
    const { data } = await api.get('/lottery/history', { params: { page: 1, page_size: 20 } })
    history.value = data.records || []
  } catch (e) {
    console.error('Failed to fetch lottery history', e)
  }
}

async function doCheckIn() {
  loading.value = true
  try {
    const { data } = await api.post('/user/checkin')
    status.value = data
  } catch (e: any) {
    alert(e?.response?.data?.message || '签到失败')
  } finally {
    loading.value = false
  }
}

async function drawLottery() {
  drawLoading.value = true
  try {
    const { data } = await api.post('/lottery/draw')
    lastResult.value = data
    showResult.value = true
    await fetchStatus()
    await fetchConfig()
    await fetchHistory()
  } catch (e: any) {
    alert(e?.response?.data?.message || '抽奖失败')
  } finally {
    drawLoading.value = false
  }
}

onMounted(() => {
  fetchStatus()
  fetchConfig()
  fetchHistory()
})
</script>
