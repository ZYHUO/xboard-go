<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

interface Coupon {
  id: number
  code: string
  name: string
  type: number
  value: number
  show: boolean
  limit_use: number | null
  started_at: number
  ended_at: number
}

const coupons = ref<Coupon[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingCoupon = ref<Partial<Coupon>>({})

const fetchCoupons = async () => {
  loading.value = true
  try {
    const res = await api.get('/admin/coupons')
    coupons.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openModal = (coupon?: Coupon) => {
  if (coupon) {
    editingCoupon.value = { ...coupon }
  } else {
    editingCoupon.value = {
      type: 1,
      value: 0,
      show: true,
      started_at: Math.floor(Date.now() / 1000),
      ended_at: Math.floor(Date.now() / 1000) + 30 * 86400
    }
  }
  showModal.value = true
}

const saveCoupon = async () => {
  try {
    if (editingCoupon.value.id) {
      await api.put(`/admin/coupon/${editingCoupon.value.id}`, editingCoupon.value)
    } else {
      await api.post('/admin/coupon', editingCoupon.value)
    }
    showModal.value = false
    fetchCoupons()
  } catch (e) {
    console.error(e)
  }
}

const deleteCoupon = async (id: number) => {
  if (!confirm('确定删除此优惠券？')) return
  try {
    await api.delete(`/admin/coupon/${id}`)
    fetchCoupons()
  } catch (e) {
    console.error(e)
  }
}

const formatMoney = (amount: number) => (amount / 100).toFixed(2)
const formatDate = (ts: number) => new Date(ts * 1000).toLocaleDateString()

onMounted(fetchCoupons)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-semibold text-gray-800">优惠券管理</h1>
      <button
        @click="openModal()"
        class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition"
      >
        添加优惠券
      </button>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
      <table class="w-full">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">优惠码</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">名称</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">类型</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">优惠值</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">有效期</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">状态</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="coupon in coupons" :key="coupon.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 font-mono text-sm">{{ coupon.code }}</td>
            <td class="px-6 py-4">{{ coupon.name }}</td>
            <td class="px-6 py-4">
              <span :class="coupon.type === 1 ? 'text-green-600' : 'text-blue-600'">
                {{ coupon.type === 1 ? '固定金额' : '百分比' }}
              </span>
            </td>
            <td class="px-6 py-4">
              {{ coupon.type === 1 ? `¥${formatMoney(coupon.value)}` : `${coupon.value}%` }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">
              {{ formatDate(coupon.started_at) }} - {{ formatDate(coupon.ended_at) }}
            </td>
            <td class="px-6 py-4">
              <span
                :class="coupon.show ? 'bg-green-100 text-green-600' : 'bg-gray-100 text-gray-600'"
                class="px-2 py-1 rounded-full text-xs"
              >
                {{ coupon.show ? '显示' : '隐藏' }}
              </span>
            </td>
            <td class="px-6 py-4">
              <button @click="openModal(coupon)" class="text-primary-600 hover:text-primary-700 mr-3">编辑</button>
              <button @click="deleteCoupon(coupon.id)" class="text-red-600 hover:text-red-700">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="bg-white rounded-2xl p-6 w-full max-w-md">
        <h2 class="text-lg font-semibold mb-4">{{ editingCoupon.id ? '编辑' : '添加' }}优惠券</h2>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-gray-600 mb-1">优惠码</label>
            <input v-model="editingCoupon.code" type="text" class="w-full px-4 py-2 border rounded-xl" />
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">名称</label>
            <input v-model="editingCoupon.name" type="text" class="w-full px-4 py-2 border rounded-xl" />
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">类型</label>
            <select v-model="editingCoupon.type" class="w-full px-4 py-2 border rounded-xl">
              <option :value="1">固定金额</option>
              <option :value="2">百分比折扣</option>
            </select>
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">优惠值（分/百分比）</label>
            <input v-model.number="editingCoupon.value" type="number" class="w-full px-4 py-2 border rounded-xl" />
          </div>
        </div>

        <div class="flex justify-end gap-3 mt-6">
          <button @click="showModal = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-xl">取消</button>
          <button @click="saveCoupon" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>
