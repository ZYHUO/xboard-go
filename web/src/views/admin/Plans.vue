<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface Plan {
  id: number
  name: string
  group_id: number | null
  transfer_enable: number
  speed_limit: number | null
  device_limit: number | null
  prices: Record<string, number>
  show: boolean
  sell: boolean
  content: string
  sort: number
}

interface ServerGroup {
  id: number
  name: string
}

const plans = ref<Plan[]>([])
const serverGroups = ref<ServerGroup[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingPlan = ref<Partial<Plan> | null>(null)

const periodLabels: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  half_yearly: '半年付',
  yearly: '年付',
  two_yearly: '两年付',
  three_yearly: '三年付',
  onetime: '一次性',
  reset: '流量重置',
}

const formatBytes = (gb: number) => {
  if (gb >= 1024) return `${(gb / 1024).toFixed(0)} TB`
  return `${gb} GB`
}

const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

const getLowestPrice = (plan: Plan) => {
  if (!plan.prices) return 0
  const prices = Object.values(plan.prices).filter(p => p > 0)
  return prices.length > 0 ? Math.min(...prices) : 0
}

const fetchPlans = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/plans')
    plans.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchServerGroups = async () => {
  try {
    const res = await api.get('/api/v2/admin/server_groups')
    serverGroups.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const openCreateModal = () => {
  editingPlan.value = {
    name: '',
    group_id: null,
    transfer_enable: 100,
    speed_limit: null,
    device_limit: null,
    prices: { monthly: 0, quarterly: 0, half_yearly: 0, yearly: 0 },
    show: true,
    sell: true,
    content: '',
    sort: 0,
  }
  showModal.value = true
}

const openEditModal = (plan: Plan) => {
  editingPlan.value = { ...plan }
  showModal.value = true
}

const savePlan = async () => {
  if (!editingPlan.value) return
  try {
    if (editingPlan.value.id) {
      await api.put(`/api/v2/admin/plan/${editingPlan.value.id}`, editingPlan.value)
    } else {
      await api.post('/api/v2/admin/plan', editingPlan.value)
    }
    showModal.value = false
    fetchPlans()
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  }
}

const deletePlan = async (plan: Plan) => {
  if (!confirm(`确定要删除套餐 "${plan.name}" 吗？`)) return
  try {
    await api.delete(`/api/v2/admin/plan/${plan.id}`)
    fetchPlans()
  } catch (e: any) {
    alert(e.response?.data?.error || '删除失败')
  }
}

onMounted(() => {
  fetchPlans()
  fetchServerGroups()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">套餐管理</h1>
        <p class="text-gray-500 mt-1">管理订阅套餐</p>
      </div>
      <button @click="openCreateModal" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition">
        添加套餐
      </button>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <table v-else class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">流量</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">限速</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">起步价</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="plan in plans" :key="plan.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 font-medium text-gray-900">{{ plan.name }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ formatBytes(plan.transfer_enable) }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">
              {{ plan.speed_limit ? `${plan.speed_limit} Mbps` : '不限速' }}
            </td>
            <td class="px-6 py-4 text-sm font-medium text-primary-600">
              {{ formatPrice(getLowestPrice(plan)) }}
            </td>
            <td class="px-6 py-4">
              <span :class="['px-2 py-1 rounded-full text-xs', plan.show && plan.sell ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600']">
                {{ plan.show && plan.sell ? '销售中' : '已下架' }}
              </span>
            </td>
            <td class="px-6 py-4 text-right space-x-2">
              <button @click="openEditModal(plan)" class="text-primary-600 hover:text-primary-700 text-sm">编辑</button>
              <button @click="deletePlan(plan)" class="text-red-600 hover:text-red-700 text-sm">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto">
          <h3 class="text-lg font-bold mb-4">{{ editingPlan?.id ? '编辑套餐' : '添加套餐' }}</h3>
          
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">名称</label>
                <input v-model="editingPlan!.name" type="text" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">用户组</label>
                <select v-model="editingPlan!.group_id" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                  <option :value="null">不限制</option>
                  <option v-for="group in serverGroups" :key="group.id" :value="group.id">{{ group.name }}</option>
                </select>
              </div>
            </div>
            
            <div class="grid grid-cols-3 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">流量 (GB)</label>
                <input v-model.number="editingPlan!.transfer_enable" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">限速 (Mbps)</label>
                <input v-model.number="editingPlan!.speed_limit" type="number" placeholder="不限速" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">设备数</label>
                <input v-model.number="editingPlan!.device_limit" type="number" placeholder="不限制" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">价格设置 (分)</label>
              <div class="grid grid-cols-4 gap-3">
                <div v-for="(label, key) in periodLabels" :key="key">
                  <label class="block text-xs text-gray-500 mb-1">{{ label }}</label>
                  <input v-model.number="editingPlan!.prices![key]" type="number" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm" />
                </div>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">描述</label>
              <textarea v-model="editingPlan!.content" rows="3" class="w-full px-4 py-2 border border-gray-200 rounded-xl"></textarea>
            </div>

            <div class="flex items-center gap-4">
              <label class="flex items-center gap-2">
                <input v-model="editingPlan!.show" type="checkbox" class="rounded" />
                <span class="text-sm">显示</span>
              </label>
              <label class="flex items-center gap-2">
                <input v-model="editingPlan!.sell" type="checkbox" class="rounded" />
                <span class="text-sm">销售</span>
              </label>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showModal = false" class="flex-1 px-4 py-2 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50">取消</button>
            <button @click="savePlan" class="flex-1 px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
