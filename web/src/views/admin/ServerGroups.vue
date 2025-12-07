<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">用户组管理</h1>
      <button @click="showCreateModal = true" class="btn btn-primary">
        添加用户组
      </button>
    </div>

    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>名称</th>
                <th>创建时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="group in groups" :key="group.id">
                <td>{{ group.id }}</td>
                <td>{{ group.name }}</td>
                <td>{{ formatDate(group.created_at) }}</td>
                <td>
                  <button @click="editGroup(group)" class="btn btn-sm btn-ghost">编辑</button>
                  <button @click="deleteGroup(group.id)" class="btn btn-sm btn-ghost text-error">删除</button>
                </td>
              </tr>
              <tr v-if="groups.length === 0">
                <td colspan="4" class="text-center text-gray-500">暂无数据</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 创建/编辑弹窗 -->
    <div v-if="showCreateModal" class="modal modal-open">
      <div class="modal-box">
        <h3 class="font-bold text-lg">{{ editingGroup ? '编辑用户组' : '添加用户组' }}</h3>
        <div class="form-control mt-4">
          <label class="label">
            <span class="label-text">名称</span>
          </label>
          <input v-model="form.name" type="text" class="input input-bordered" placeholder="请输入用户组名称" />
        </div>
        <div class="modal-action">
          <button @click="closeModal" class="btn">取消</button>
          <button @click="saveGroup" class="btn btn-primary" :disabled="!form.name">保存</button>
        </div>
      </div>
    </div>

    <div class="mt-6 p-4 bg-base-200 rounded-lg">
      <h3 class="font-bold mb-2">使用说明</h3>
      <ul class="list-disc list-inside text-sm text-gray-600 space-y-1">
        <li>用户组用于控制用户可以访问哪些节点</li>
        <li>在「套餐管理」中为套餐设置用户组，用户购买后自动加入该组</li>
        <li>在「节点管理」中为节点设置允许的用户组</li>
        <li>用户组为空的节点表示所有用户都可以使用</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

interface ServerGroup {
  id: number
  name: string
  created_at: number
}

const groups = ref<ServerGroup[]>([])
const showCreateModal = ref(false)
const editingGroup = ref<ServerGroup | null>(null)
const form = ref({ name: '' })

const fetchGroups = async () => {
  try {
    const res = await api.get('/admin/server_groups')
    groups.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const editGroup = (group: ServerGroup) => {
  editingGroup.value = group
  form.value.name = group.name
  showCreateModal.value = true
}

const closeModal = () => {
  showCreateModal.value = false
  editingGroup.value = null
  form.value.name = ''
}

const saveGroup = async () => {
  try {
    if (editingGroup.value) {
      await api.put(`/admin/server_group/${editingGroup.value.id}`, form.value)
    } else {
      await api.post('/admin/server_group', form.value)
    }
    closeModal()
    fetchGroups()
  } catch (e) {
    console.error(e)
  }
}

const deleteGroup = async (id: number) => {
  if (!confirm('确定删除该用户组？')) return
  try {
    await api.delete(`/admin/server_group/${id}`)
    fetchGroups()
  } catch (e) {
    console.error(e)
  }
}

const formatDate = (ts: number) => {
  return new Date(ts * 1000).toLocaleString()
}

onMounted(fetchGroups)
</script>
