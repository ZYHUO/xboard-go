<template>
  <div 
    :class="cardClasses"
    :style="elevationStyle"
    v-bind="elevationListeners"
  >
    <div v-if="$slots.header || title" class="card-header">
      <slot name="header">
        <h3 v-if="title" class="card-title">{{ title }}</h3>
      </slot>
    </div>
    <div class="card-body" :class="bodyClasses">
      <slot />
    </div>
    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCardElevation } from '../../composables/useElevation'

interface Props {
  variant?: 'default' | 'outlined' | 'filled' | 'elevated'
  padding?: 'none' | 'sm' | 'md' | 'lg'
  interactive?: boolean
  title?: string
  hoverable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  padding: 'md',
  interactive: false,
  hoverable: false,
})

// 使用层次系统（仅当 interactive 或 hoverable 时）
const shouldUseElevation = computed(() => props.interactive || props.hoverable)
const elevation = shouldUseElevation.value ? useCardElevation() : null

const elevationStyle = computed(() => {
  if (!shouldUseElevation.value || !elevation) return {}
  return elevation.style.value
})

const elevationListeners = computed(() => {
  if (!shouldUseElevation.value || !elevation) return {}
  return elevation.listeners
})

const cardClasses = computed(() => {
  const classes = ['card']
  
  // Variant styles
  classes.push(`card-${props.variant}`)
  
  // Interactive/Hoverable
  if (props.interactive || props.hoverable) {
    classes.push('card-interactive')
  }
  
  return classes.join(' ')
})

const bodyClasses = computed(() => {
  // Padding styles
  const paddingClasses = {
    none: 'p-0',
    sm: 'p-3',
    md: 'p-4',
    lg: 'p-6',
  }
  return paddingClasses[props.padding]
})
</script>

<style scoped>
.card {
  @apply bg-white rounded-md;
  @apply transition-all duration-200 ease-in-out;
}

/* Default variant - subtle border */
.card-default {
  @apply border border-gray-200;
}

/* Outlined variant - prominent border */
.card-outlined {
  @apply border-2 border-gray-300;
}

/* Filled variant - background color */
.card-filled {
  @apply bg-gray-50 border border-gray-200;
}

/* Elevated variant - with shadow */
.card-elevated {
  @apply border border-gray-100;
}

/* Interactive cards */
.card-interactive {
  @apply cursor-pointer;
}

.card-interactive:hover {
  @apply border-gray-400;
}

.card-interactive:active {
  @apply scale-[0.98];
}

/* Card header */
.card-header {
  @apply px-4 py-3 border-b border-gray-200;
}

.card-title {
  @apply text-lg font-semibold text-gray-900;
}

/* Card body */
.card-body {
  @apply flex-1;
}

/* Card footer */
.card-footer {
  @apply px-4 py-3 border-t border-gray-200;
}

/* Focus state for interactive cards */
.card-interactive:focus-within {
  @apply outline-none ring-2 ring-gray-400 ring-offset-2;
}
</style>
