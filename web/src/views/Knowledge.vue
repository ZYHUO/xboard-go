<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api'

interface KnowledgeItem {
  id: number
  title: string
  body: string
  category: string
  language: string
}

const categories = ref<string[]>([])
const items = ref<KnowledgeItem[]>([])
const selectedCategory = ref('')
const selectedItem = ref<KnowledgeItem | null>(null)
const loading = ref(false)

const filteredItems = computed(() => {
  if (!selectedCategory.value) return items.value
  return items.value.filter(item => item.category === selectedCategory.value)
})

const fetchCategories = async () => {
  try {
    const res = await api.get('/knowledge/categories')
    categories.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchItems = async () => {
  loading.value = true
  try {
    const res = await api.get('/knowledge', {
      params: { category: selectedCategory.value }
    })
    items.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const selectCategory = (category: string) => {
  selectedCategory.value = category
  selectedItem.value = null
  fetchItems()
}

const selectItem = (item: KnowledgeItem) => {
  selectedItem.value = item
}

const goBack = () => {
  selectedItem.value = null
}

onMounted(() => {
  fetchCategories()
  fetchItems()
})
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-semibold text-gray-800">帮助文档</h1>

    <div class="flex gap-6">
      <!-- 分类侧边栏 -->
      <div class="w-48 flex-shrink-0">
        <div class="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
          <h3 class="text-sm font-medium text-gray-500 mb-3">分类</h3>
          <ul class="space-y-1">
            <li>
              <button
                @click="selectCategory('')"
                :class="selectedCategory === '' ? 'bg-primary-50 text-primary-600' : 'text-gray-600 hover:bg-gray-50'"
                class="w-full text-left px-3 py-2 rounded-lg text-sm transition"
              >
                全部
              </button>
            </li>
            <li v-for="cat in categories" :key="cat">
              <button
                @click="selectCategory(cat)"
                :class="selectedCategory === cat ? 'bg-primary-50 text-primary-600' : 'text-gray-600 hover:bg-gray-50'"
                class="w-full text-left px-3 py-2 rounded-lg text-sm transition"
              >
                {{ cat }}
              </button>
            </li>
          </ul>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="flex-1">
        <!-- 文章详情 -->
        <div v-if="selectedItem" class="bg-white rounded-2xl p-6 shadow-sm border border-gray-100">
          <button
            @click="goBack"
            class="flex items-center gap-2 text-gray-500 hover:text-gray-700 mb-4"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            返回列表
          </button>
          
          <h2 class="text-xl font-semibold text-gray-800 mb-4">{{ selectedItem.title }}</h2>
          <div class="prose prose-sm max-w-none" v-html="selectedItem.body"></div>
        </div>

        <!-- 文章列表 -->
        <div v-else class="space-y-3">
          <div v-if="loading" class="text-center py-12 text-gray-500">
            加载中...
          </div>
          
          <div v-else-if="filteredItems.length === 0" class="text-center py-12 text-gray-500">
            暂无文档
          </div>

          <div
            v-else
            v-for="item in filteredItems"
            :key="item.id"
            @click="selectItem(item)"
            class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100 cursor-pointer hover:border-primary-200 hover:shadow-md transition"
          >
            <div class="flex items-center justify-between">
              <div>
                <h3 class="font-medium text-gray-800">{{ item.title }}</h3>
                <span class="text-xs text-gray-400 mt-1">{{ item.category }}</span>
              </div>
              <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
