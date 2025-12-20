<template>
  <button
    :class="buttonClasses"
    :type="type"
    :disabled="disabled || loading"
    :style="elevationStyle"
    v-bind="elevationListeners"
    @click="handleClick"
  >
    <span v-if="loading" class="btn-loading">
      <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </span>
    <span v-if="icon && iconPosition === 'left'" class="btn-icon btn-icon-left">
      <slot name="icon">{{ icon }}</slot>
    </span>
    <span class="btn-content">
      <slot />
    </span>
    <span v-if="icon && iconPosition === 'right'" class="btn-icon btn-icon-right">
      <slot name="icon">{{ icon }}</slot>
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useButtonElevation } from '../../composables/useElevation'

interface Props {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  block?: boolean
  icon?: string
  iconPosition?: 'left' | 'right'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  type: 'button',
  disabled: false,
  loading: false,
  block: false,
  iconPosition: 'left',
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

// 使用层次系统
const elevation = useButtonElevation()
const elevationStyle = computed(() => elevation.style.value)
const elevationListeners = elevation.listeners

const buttonClasses = computed(() => {
  const classes = ['btn']
  
  // Variant styles
  classes.push(`btn-${props.variant}`)
  
  // Size styles
  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm min-h-[32px]',
    md: 'px-4 py-2.5 text-base min-h-[40px]',
    lg: 'px-6 py-3 text-lg min-h-[48px]',
  }
  classes.push(sizeClasses[props.size])
  
  // Block style
  if (props.block) {
    classes.push('w-full')
  }
  
  // Disabled/Loading style
  if (props.disabled || props.loading) {
    classes.push('opacity-50 cursor-not-allowed')
  }
  
  // Loading state
  if (props.loading) {
    classes.push('btn-loading-state')
  }
  
  return classes.join(' ')
})

const handleClick = (event: MouseEvent) => {
  if (!props.disabled && !props.loading) {
    emit('click', event)
  }
}
</script>

<style scoped>
.btn {
  @apply inline-flex items-center justify-center gap-2;
  @apply font-medium rounded-md;
  @apply transition-all duration-200 ease-in-out;
  @apply focus:outline-none focus:ring-2 focus:ring-offset-2;
  @apply disabled:pointer-events-none;
}

/* Primary variant - Modern depth */
.btn-primary {
  @apply bg-black text-white;
  @apply hover:bg-gray-800;
  @apply active:bg-gray-900;
  @apply focus:ring-gray-500;
}

/* Secondary variant */
.btn-secondary {
  @apply bg-gray-100 text-gray-900;
  @apply hover:bg-gray-200;
  @apply active:bg-gray-300;
  @apply focus:ring-gray-400;
}

/* Outline variant */
.btn-outline {
  @apply bg-transparent text-gray-900 border-2 border-gray-300;
  @apply hover:bg-gray-50 hover:border-gray-400;
  @apply active:bg-gray-100;
  @apply focus:ring-gray-400;
}

/* Ghost variant */
.btn-ghost {
  @apply bg-transparent text-gray-700;
  @apply hover:bg-gray-100;
  @apply active:bg-gray-200;
  @apply focus:ring-gray-400;
}

/* Danger variant */
.btn-danger {
  @apply bg-red-600 text-white;
  @apply hover:bg-red-700;
  @apply active:bg-red-800;
  @apply focus:ring-red-500;
}

/* Loading state */
.btn-loading {
  @apply inline-flex items-center;
}

.btn-loading-state {
  @apply pointer-events-none;
}

/* Icon styles */
.btn-icon {
  @apply inline-flex items-center;
}

.btn-icon-left {
  @apply -ml-1;
}

.btn-icon-right {
  @apply -mr-1;
}

/* Animation */
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}
</style>
