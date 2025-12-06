<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api'

interface InviteCode {
  id: number
  code: string
  status: boolean
  pv: number
  created_at: number
}

interface InviteStats {
  invited_count: number
  total_commission: number
  commission_balance: number
}

const codes = ref<InviteCode[]>([])
const stats = ref<InviteStats>({
  invited_count: 0,
  total_commission: 0,
  commission_balance: 0
})
const loading = ref(false)
const generating = ref(false)
const withdrawing = ref(false)

const inviteUrl = computed(() => {
  if (codes.value.length === 0) return ''
  return `${window.location.origin}/register?code=${codes.value[0].code}`
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.get('/user/invite')
    codes.value = res.data.data.codes || []
    stats.value = res.data.data.stats || stats.value
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const generateCode = async () => {
  generating.value = true
  try {
    await api.post('/user/invite/generate')
    await fetchData()
  } catch (e) {
    console.error(e)
  } finally {
    generating.value = false
  }
}

const withdraw = async () => {
  if (stats.value.commission_balance <= 0) return
  withdrawing.value = true
  try {
    await api.post('/user/invite/withdraw')
    await fetchData()
  } catch (e) {
    console.error(e)
  } finally {
    withdrawing.value = false
  }
}

const copyUrl = () => {
  navigator.clipboard.writeText(inviteUrl.value)
}

const formatMoney = (amount: number) => {
  return (amount / 100).toFixed(2)
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-semibold text-gray-800">邀请返利</h1>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-white rounded-2xl p-6 shadow-sm border border-gray-100">
        <div class="text-sm text-gray-500 mb-2">邀请人数</div>
        <div class="text-3xl font-bold text-primary-600">{{ stats.invited_count }}</div>
      </div>
      <div class="bg-white rounded-2xl p-6 shadow-sm border border-gray-100">
        <div class="text-sm text-gray-500 mb-2">累计佣金</div>
        <div class="text-3xl font-bold text-green-600">¥{{ formatMoney(stats.total_commission) }}</div>
      </div>
      <div class="bg-white rounded-2xl p-6 shadow-sm border border-gray-100">
        <div class="text-sm text-gray-500 mb-2">可提现佣金</div>
        <div class="text-3xl font-bold text-orange-600">¥{{ formatMoney(stats.commission_balance) }}</div>
        <button
          v-if="stats.commission_balance > 0"
          @click="withdraw"
          :disabled="withdrawing"
          class="mt-3 px-4 py-2 bg-orange-500 text-white rounded-lg text-sm hover:bg-orange-600 transition disabled:opacity-50"
        >
          {{ withdrawing ? '处理中...' : '转入余额' }}
        </button>
      </div>
    </div>

    <!-- 邀请链接 -->
    <div class="bg-white rounded-2xl p-6 shadow-sm border border-gray-100">
      <h2 class="text-lg font-medium text-gray-800 mb-4">邀请链接</h2>
      
      <div v-if="codes.length > 0" class="space-y-4">
        <div class="flex items-center gap-3">
          <input
            type="text"
            :value="inviteUrl"
            readonly
            class="flex-1 px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm"
          />
          <button
            @click="copyUrl"
            class="px-6 py-3 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition"
          >
            复制
          </button>
        </div>
        <p class="text-sm text-gray-500">
          分享此链接给好友，好友注册并购买套餐后，您将获得佣金奖励
        </p>
      </div>

      <div v-else class="text-center py-8">
        <p class="text-gray-500 mb-4">您还没有邀请码</p>
        <button
          @click="generateCode"
          :disabled="generating"
          class="px-6 py-3 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition disabled:opacity-50"
        >
          {{ generating ? '生成中...' : '生成邀请码' }}
        </button>
      </div>
    </div>

    <!-- 邀请码列表 -->
    <div v-if="codes.length > 0" class="bg-white rounded-2xl p-6 shadow-sm border border-gray-100">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-medium text-gray-800">我的邀请码</h2>
        <button
          @click="generateCode"
          :disabled="generating"
          class="px-4 py-2 text-primary-600 hover:bg-primary-50 rounded-lg transition text-sm"
        >
          生成新邀请码
        </button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="text-left text-sm text-gray-500 border-b">
              <th class="pb-3 font-medium">邀请码</th>
              <th class="pb-3 font-medium">状态</th>
              <th class="pb-3 font-medium">访问次数</th>
              <th class="pb-3 font-medium">创建时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="code in codes" :key="code.id" class="border-b border-gray-50">
              <td class="py-4 font-mono text-sm">{{ code.code }}</td>
              <td class="py-4">
                <span
                  :class="code.status ? 'bg-gray-100 text-gray-600' : 'bg-green-100 text-green-600'"
                  class="px-2 py-1 rounded-full text-xs"
                >
                  {{ code.status ? '已使用' : '可用' }}
                </span>
              </td>
              <td class="py-4 text-sm text-gray-600">{{ code.pv }}</td>
              <td class="py-4 text-sm text-gray-500">
                {{ new Date(code.created_at * 1000).toLocaleDateString() }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
