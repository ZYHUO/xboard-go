<template>
  <div class="table-wrapper" :class="wrapperClasses">
    <table class="table" :class="tableClasses">
      <thead class="table-header">
        <tr>
          <th
            v-if="selectable"
            class="table-header-cell table-checkbox-cell"
          >
            <input
              type="checkbox"
              :checked="allSelected"
              :indeterminate="someSelected"
              @change="toggleSelectAll"
              class="table-checkbox"
            />
          </th>
          <th
            v-for="column in columns"
            :key="column.key"
            :class="getHeaderClass(column)"
            @click="handleSort(column)"
          >
            <div class="table-header-content">
              <span>{{ column.label }}</span>
              <span v-if="column.sortable" class="sort-icon" :class="getSortIconClass(column)">
                <svg v-if="sortKey === column.key" class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path v-if="sortOrder === 'asc'" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" />
                  <path v-else d="M14.707 12.707a1 1 0 01-1.414 0L10 9.414l-3.293 3.293a1 1 0 01-1.414-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 010 1.414z" />
                </svg>
              </span>
            </div>
          </th>
        </tr>
      </thead>
      <tbody class="table-body">
        <tr
          v-for="(row, index) in sortedData"
          :key="getRowKey(row, index)"
          :class="getRowClass(row, index)"
          @click="handleRowClick(row, index)"
        >
          <td v-if="selectable" class="table-cell table-checkbox-cell">
            <input
              type="checkbox"
              :checked="isRowSelected(row)"
              @change="toggleRowSelection(row)"
              @click.stop
              class="table-checkbox"
            />
          </td>
          <td
            v-for="column in columns"
            :key="column.key"
            :class="getCellClass(column)"
          >
            <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]" :index="index">
              {{ row[column.key] }}
            </slot>
          </td>
        </tr>
        <tr v-if="sortedData.length === 0" class="table-empty-row">
          <td :colspan="columns.length + (selectable ? 1 : 0)" class="table-empty-cell">
            <slot name="empty">
              <div class="table-empty-content">
                <p class="text-gray-500">{{ emptyText }}</p>
              </div>
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Column {
  key: string
  label: string
  sortable?: boolean
  align?: 'left' | 'center' | 'right'
  width?: string
}

interface Props {
  columns: Column[]
  data: Record<string, any>[]
  striped?: boolean
  bordered?: boolean
  hoverable?: boolean
  selectable?: boolean
  rowKey?: string
  emptyText?: string
}

const props = withDefaults(defineProps<Props>(), {
  striped: true,
  bordered: true,
  hoverable: true,
  selectable: false,
  rowKey: 'id',
  emptyText: '暂无数据',
})

const emit = defineEmits<{
  'row-click': [row: Record<string, any>, index: number]
  'selection-change': [selectedRows: Record<string, any>[]]
}>()

const sortKey = ref<string>('')
const sortOrder = ref<'asc' | 'desc'>('asc')
const selectedRows = ref<Set<any>>(new Set())

const sortedData = computed(() => {
  if (!sortKey.value) return props.data
  
  return [...props.data].sort((a, b) => {
    const aVal = a[sortKey.value]
    const bVal = b[sortKey.value]
    
    if (aVal === bVal) return 0
    
    const comparison = aVal > bVal ? 1 : -1
    return sortOrder.value === 'asc' ? comparison : -comparison
  })
})

const allSelected = computed(() => {
  return props.data.length > 0 && selectedRows.value.size === props.data.length
})

const someSelected = computed(() => {
  return selectedRows.value.size > 0 && selectedRows.value.size < props.data.length
})

const wrapperClasses = computed(() => {
  const classes = []
  if (props.bordered) classes.push('table-wrapper-bordered')
  return classes.join(' ')
})

const tableClasses = computed(() => {
  const classes = []
  if (props.striped) classes.push('table-striped')
  if (props.hoverable) classes.push('table-hoverable')
  return classes.join(' ')
})

const getRowKey = (row: Record<string, any>, index: number) => {
  return props.rowKey && row[props.rowKey] ? row[props.rowKey] : index
}

const getHeaderClass = (column: Column) => {
  const classes = ['table-header-cell']
  
  if (column.sortable) {
    classes.push('table-header-sortable')
  }
  
  if (column.align) {
    classes.push(`text-${column.align}`)
  }
  
  return classes.join(' ')
}

const getCellClass = (column: Column) => {
  const classes = ['table-cell']
  
  if (column.align) {
    classes.push(`text-${column.align}`)
  }
  
  return classes.join(' ')
}

const getRowClass = (row: Record<string, any>, index: number) => {
  const classes = ['table-row']
  
  if (isRowSelected(row)) {
    classes.push('table-row-selected')
  }
  
  return classes.join(' ')
}

const getSortIconClass = (column: Column) => {
  const classes = []
  
  if (sortKey.value === column.key) {
    classes.push('sort-icon-active')
  }
  
  return classes.join(' ')
}

const handleSort = (column: Column) => {
  if (!column.sortable) return
  
  if (sortKey.value === column.key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = column.key
    sortOrder.value = 'asc'
  }
}

const handleRowClick = (row: Record<string, any>, index: number) => {
  emit('row-click', row, index)
}

const isRowSelected = (row: Record<string, any>) => {
  const key = getRowKey(row, 0)
  return selectedRows.value.has(key)
}

const toggleRowSelection = (row: Record<string, any>) => {
  const key = getRowKey(row, 0)
  
  if (selectedRows.value.has(key)) {
    selectedRows.value.delete(key)
  } else {
    selectedRows.value.add(key)
  }
  
  emitSelectionChange()
}

const toggleSelectAll = () => {
  if (allSelected.value) {
    selectedRows.value.clear()
  } else {
    props.data.forEach((row, index) => {
      selectedRows.value.add(getRowKey(row, index))
    })
  }
  
  emitSelectionChange()
}

const emitSelectionChange = () => {
  const selected = props.data.filter((row, index) => 
    selectedRows.value.has(getRowKey(row, index))
  )
  emit('selection-change', selected)
}
</script>

<style scoped>
.table-wrapper {
  @apply w-full overflow-x-auto rounded-md;
}

.table-wrapper-bordered {
  @apply border border-gray-200;
}

.table {
  @apply w-full border-collapse;
}

/* 表头样式 */
.table-header {
  @apply bg-gray-50;
}

.table-header-cell {
  @apply px-4 py-3 text-left text-sm font-semibold text-gray-900;
  @apply border-b-2 border-gray-200;
  @apply transition-colors duration-150;
}

.table-header-sortable {
  @apply cursor-pointer select-none;
  @apply hover:bg-gray-100;
}

.table-header-content {
  @apply flex items-center gap-2;
}

.sort-icon {
  @apply text-gray-400 transition-colors duration-150;
}

.sort-icon-active {
  @apply text-gray-900;
}

/* 表体样式 */
.table-body {
  @apply bg-white;
}

.table-row {
  @apply transition-colors duration-150;
}

.table-striped .table-row:nth-child(even) {
  @apply bg-gray-50;
}

.table-hoverable .table-row:hover {
  @apply bg-gray-100 cursor-pointer;
}

.table-row-selected {
  @apply bg-blue-50;
}

.table-hoverable .table-row-selected:hover {
  @apply bg-blue-100;
}

.table-cell {
  @apply px-4 py-3 text-sm text-gray-900;
  @apply border-b border-gray-200;
}

/* 复选框单元格 */
.table-checkbox-cell {
  @apply w-12 text-center;
}

.table-checkbox {
  @apply w-4 h-4 rounded border-gray-300;
  @apply text-black focus:ring-2 focus:ring-black focus:ring-offset-0;
  @apply cursor-pointer transition-colors duration-150;
}

/* 空状态 */
.table-empty-row {
  @apply bg-white;
}

.table-empty-cell {
  @apply px-4 py-8 text-center;
}

.table-empty-content {
  @apply flex flex-col items-center justify-center gap-2;
}
</style>
