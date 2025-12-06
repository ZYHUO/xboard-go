<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

interface Notice {
  id: number
  title: string
  content: string
  show: boolean
  sort: number | null
  created_at: number
}

const notices = ref<Notice[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingNotice = ref<Partial<Notice>>({})

const fetchNotices = async () => {
  loading.value = true
  try {
    const res = await api.get('/admin/notices')
    notices.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openModal = (notice?: Notice) => {
  if (notice) {
    editingNotice.value = { ...notice }
  } else {
    editingNotice.value = { show: true, sort: 0 }
  }
  showModal.value = true
}

const saveNotice = async () => {
  try {
    if (editingNotice.value.id) {
      await api.put(`/admin/notice/${editingNotice.value.id}`, editingNotice.value)
    } else {
      await api.post('/admin/notice', editingNotice.value)
    }
    showModal.value = false
    fetchNotices()
  } catch (e) {
    console.error(e)
  }
}

const deleteNotice = async (id: number) => {
  if (!confirm('确定删除此公告？')) return
  try {
    await api.delete(`/admin/notice/${id}`)
    fetchNotices()
  } catch (e) {
    console.error(e)
  }
}

const formatDate = (ts: number) => new Date(ts * 1000).toLocaleString()

onMounted(fetchNotices)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-semibold text-gray-800">公告管理</h1>
      <button
        @click="openModal()"
        class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition"
      >
        添加公告
      </button>
    </div>

    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
      <table class="w-full">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">标题</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">排序</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">状态</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">创建时间</th>
            <th class="px-6 py-4 text-left text-sm font-medium text-gray-500">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="notice in notices" :key="notice.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 font-medium">{{ notice.title }}</td>
            <td class="px-6 py-4 text-gray-500">{{ notice.sort || 0 }}</td>
            <td class="px-6 py-4">
              <span
                :class="notice.show ? 'bg-green-100 text-green-600' : 'bg-gray-100 text-gray-600'"
                class="px-2 py-1 rounded-full text-xs"
              >
                {{ notice.show ? '显示' : '隐藏' }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(notice.created_at) }}</td>
            <td class="px-6 py-4">
              <button @click="openModal(notice)" class="text-primary-600 hover:text-primary-700 mr-3">编辑</button>
              <button @click="deleteNotice(notice.id)" class="text-red-600 hover:text-red-700">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="bg-white rounded-2xl p-6 w-full max-w-lg">
        <h2 class="text-lg font-semibold mb-4">{{ editingNotice.id ? '编辑' : '添加' }}公告</h2>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-gray-600 mb-1">标题</label>
            <input v-model="editingNotice.title" type="text" class="w-full px-4 py-2 border rounded-xl" />
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">内容</label>
            <textarea v-model="editingNotice.content" rows="6" class="w-full px-4 py-2 border rounded-xl"></textarea>
          </div>
          <div class="flex gap-4">
            <div class="flex-1">
              <label class="block text-sm text-gray-600 mb-1">排序</label>
              <input v-model.number="editingNotice.sort" type="number" class="w-full px-4 py-2 border rounded-xl" />
            </div>
            <div class="flex items-center gap-2 pt-6">
              <input v-model="editingNotice.show" type="checkbox" id="show" class="rounded" />
              <label for="show" class="text-sm text-gray-600">显示</label>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-3 mt-6">
          <button @click="showModal = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-xl">取消</button>
          <button @click="saveNotice" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>
