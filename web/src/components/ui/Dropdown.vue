<template>
  <div class="dropdown" ref="dropdownRef">
    <!-- Trigger -->
    <div
      class="dropdown-trigger"
      @click="handleTriggerClick"
      @mouseenter="handleMouseEnter"
      @mouseleave="handleMouseLeave"
    >
      <slot name="trigger">
        <button class="dropdown-button">
          <span>{{ triggerText }}</span>
          <svg
            class="dropdown-arrow"
            :class="{ 'dropdown-arrow-open': isOpen }"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </slot>
    </div>

    <!-- Dropdown Menu -->
    <Transition :name="transition">
      <div
        v-if="isOpen"
        class="dropdown-menu"
        :class="menuClasses"
        :style="menuStyle"
        v-bind="elevationListeners"
      >
        <slot />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useElevation } from '../../composables/useElevation'

type Trigger = 'click' | 'hover'
type Placement = 'bottom-start' | 'bottom-end' | 'top-start' | 'top-end' | 'left' | 'right'

interface Props {
  modelValue?: boolean
  trigger?: Trigger
  placement?: Placement
  disabled?: boolean
  triggerText?: string
  transition?: string
}

const props = withDefaults(defineProps<Props>(), {
  trigger: 'click',
  placement: 'bottom-start',
  disabled: false,
  triggerText: 'Dropdown',
  transition: 'dropdown',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  open: []
  close: []
}>()

const dropdownRef = ref<HTMLElement>()
const isOpen = ref(props.modelValue || false)
const hoverTimer = ref<number>()

// 使用 elevation 系统
const elevation = useElevation(3, { interactive: false })
const elevationListeners = elevation.listeners

const menuClasses = computed(() => {
  const classes = ['dropdown-menu-base']
  classes.push(`dropdown-menu-${props.placement}`)
  classes.push(elevation.className.value)
  return classes.join(' ')
})

const menuStyle = computed(() => ({
  ...elevation.style.value,
}))

const handleTriggerClick = () => {
  if (props.disabled) return
  if (props.trigger !== 'click') return
  
  toggleDropdown()
}

const handleMouseEnter = () => {
  if (props.disabled) return
  if (props.trigger !== 'hover') return
  
  clearTimeout(hoverTimer.value)
  openDropdown()
}

const handleMouseLeave = () => {
  if (props.disabled) return
  if (props.trigger !== 'hover') return
  
  hoverTimer.value = window.setTimeout(() => {
    closeDropdown()
  }, 200)
}

const toggleDropdown = () => {
  if (isOpen.value) {
    closeDropdown()
  } else {
    openDropdown()
  }
}

const openDropdown = () => {
  isOpen.value = true
  emit('update:modelValue', true)
  emit('open')
}

const closeDropdown = () => {
  isOpen.value = false
  emit('update:modelValue', false)
  emit('close')
}

// 点击外部关闭
const handleClickOutside = (e: MouseEvent) => {
  if (!dropdownRef.value) return
  if (!isOpen.value) return
  
  const target = e.target as Node
  if (!dropdownRef.value.contains(target)) {
    closeDropdown()
  }
}

// 键盘事件
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && isOpen.value) {
    closeDropdown()
  }
}

// 监听 modelValue 变化
watch(() => props.modelValue, (value) => {
  if (value !== undefined) {
    isOpen.value = value
  }
})

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
  clearTimeout(hoverTimer.value)
})
</script>

<style scoped>
.dropdown {
  @apply relative inline-block;
}

.dropdown-trigger {
  @apply inline-block;
}

.dropdown-button {
  @apply inline-flex items-center gap-2;
  @apply px-4 py-2;
  @apply bg-white border border-gray-300;
  @apply rounded-md;
  @apply text-sm font-medium text-gray-700;
  @apply hover:bg-gray-50;
  @apply focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2;
  @apply transition-colors duration-150;
}

.dropdown-arrow {
  @apply w-4 h-4;
  @apply transition-transform duration-200;
}

.dropdown-arrow-open {
  @apply rotate-180;
}

.dropdown-menu {
  @apply absolute z-dropdown;
  @apply min-w-[160px];
  @apply bg-white;
  @apply border border-gray-200;
  @apply rounded-md;
  @apply py-1;
}

.dropdown-menu-base {
  @apply mt-2;
}

/* Placement */
.dropdown-menu-bottom-start {
  @apply top-full left-0;
}

.dropdown-menu-bottom-end {
  @apply top-full right-0;
}

.dropdown-menu-top-start {
  @apply bottom-full left-0 mb-2;
}

.dropdown-menu-top-end {
  @apply bottom-full right-0 mb-2;
}

.dropdown-menu-left {
  @apply right-full top-0 mr-2;
}

.dropdown-menu-right {
  @apply left-full top-0 ml-2;
}

/* Transitions */
.dropdown-enter-active,
.dropdown-leave-active {
  @apply transition-all duration-150;
}

.dropdown-enter-from,
.dropdown-leave-to {
  @apply opacity-0 scale-95;
}
</style>
