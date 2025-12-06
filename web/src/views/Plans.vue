<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api'

interface Plan {
  id: number
  name: string
  transfer_enable: number
  speed_limit: number | null
  prices: Record<string, number>
  content: string | null
}

const router = useRouter()
const plans = ref<Plan[]>([])
const loading = ref(false)
const selectedPlan = ref<Plan | null>(null)
const selectedPeriod = ref('')
const showModal = ref(false)
const ordering = ref(false)

const periods = [
  { key: 'monthly', name: '月付', months: 1 },
  { key: 'quarterly', name: '季付', months: 3 },
  { key: 'half_yearly', name: '半年付', months: 6 },
  { key: 'yearly', name: '年付', months: 12 },
  { key: 'two_yearly', name: '两年付', months: 24 },
  { key: 'three_yearly', name: '三年付', months: 36 },
  { key: 'onetime', name: '一次性', months: -1 },
]

const formatBytes = (gb: number) => {
  if (gb >= 1024) return `${(gb / 1024).toFixed(0)} TB`
  return `${gb} GB`
}

const formatPrice = (cents: number) => {
  return `¥${(cents / 100).toFixed(2)}`
}

const getAvailablePeriods = (plan: Plan) => {
  return periods.filter(p => plan.prices && plan.prices[p.key] > 0)
}

const getLowestPrice = (plan: Plan) => {
  const available = getAvailablePeriods(plan)
  if (available.length === 0) return 0
  return Math.min(...available.map(p => plan.prices[p.key]))
}

const openOrderModal = (plan: Plan) => {
  selectedPlan.value = plan
  const available = getAvailablePeriods(plan)
  selectedPeriod.value = available[0]?.key || ''
  showModal.value = true
}

const createOrder = async () => {
  if (!selectedPlan.value || !selectedPeriod.value) return

  ordering.value = true
  try {
    const res = await api.post('/api/v1/user/order/create', {
      plan_id: selectedPlan.value.id,
      period: selectedPeriod.value,
    })
    showModal.value = false
    router.push('/orders')
  } catch (e: any) {
    alert(e.response?.data?.error || '创建订单失败')
  } finally {
    ordering.value = false
  }
}

const fetchPlans = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/guest/plans')
    plans.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchPlans)
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">选择套餐</h1>
      <p class="text-gray-500 mt-1">选择适合您的订阅计划</p>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500">
      加载中...
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="plan in plans"
        :key="plan.id"
        class="card hover:shadow-lg hover:-translate-y-1 transition-all duration-300"
      >
        <div class="text-center mb-6">
          <h3 class="text-xl font-bold text-gray-900">{{ plan.name }}</h3>
          <div class="mt-4">
            <span class="text-3xl font-bold gradient-text">{{ formatPrice(getLowestPrice(plan)) }}</span>
            <span class="text-gray-500">/起</span>
          </div>
        </div>

        <div class="space-y-3 mb-6">
          <div class="flex items-center gap-2 text-gray-600">
            <span class="text-green-500">✓</span>
            <span>{{ formatBytes(plan.transfer_enable) }} 流量</span>
          </div>
          <div class="flex items-center gap-2 text-gray-600">
            <span class="text-green-500">✓</span>
            <span>{{ plan.speed_limit ? `${plan.speed_limit} Mbps` : '不限速' }}</span>
          </div>
          <div class="flex items-center gap-2 text-gray-600">
            <span class="text-green-500">✓</span>
            <span>全部节点</span>
          </div>
        </div>

        <button
          @click="openOrderModal(plan)"
          class="w-full btn btn-primary"
        >
          立即订购
        </button>
      </div>
    </div>

    <!-- Order Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md p-6 animate-scale-in">
          <h3 class="text-xl font-bold mb-4">确认订单</h3>
          
          <div class="mb-4">
            <p class="text-gray-600">套餐：<span class="font-medium text-gray-900">{{ selectedPlan?.name }}</span></p>
          </div>

          <div class="mb-6">
            <label class="block text-sm font-medium text-gray-700 mb-2">选择周期</label>
            <div class="grid grid-cols-2 gap-2">
              <button
                v-for="period in getAvailablePeriods(selectedPlan!)"
                :key="period.key"
                @click="selectedPeriod = period.key"
                :class="[
                  'p-3 rounded-xl border-2 transition-all',
                  selectedPeriod === period.key
                    ? 'border-primary-500 bg-primary-50'
                    : 'border-surface-200 hover:border-primary-300'
                ]"
              >
                <div class="font-medium">{{ period.name }}</div>
                <div class="text-sm text-gray-500">{{ formatPrice(selectedPlan!.prices[period.key]) }}</div>
              </button>
            </div>
          </div>

          <div class="flex gap-3">
            <button @click="showModal = false" class="flex-1 btn btn-secondary">
              取消
            </button>
            <button @click="createOrder" :disabled="ordering" class="flex-1 btn btn-primary">
              {{ ordering ? '创建中...' : '确认订购' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
