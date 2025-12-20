<template>
  <component
    :is="tag"
    class="dropdown-item"
    :class="itemClasses"
    :disabled="disabled"
    @click="handleClick"
  >
    <span v-if="icon" class="dropdown-item-icon">
      <slot name="icon">{{ icon }}</slot>
    </span>
    <span class="dropdown-item-content">
      <slot />
    </span>
    <span v-if="shortcut" class="dropdown-item-shortcut">
      {{ shortcut }}
    </span>
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  disabled?: boolean
  divided?: boolean
  icon?: string
  shortcut?: string
  danger?: boolean
  tag?: string
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  divided: false,
  danger: false,
  tag: 'button',
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const itemClasses = computed(() => {
  const classes = []
  
  if (props.disabled) {
    classes.push('dropdown-item-disabled')
  }
  
  if (props.divided) {
    classes.push('dropdown-item-divided')
  }
  
  if (props.danger) {
    classes.push('dropdown-item-danger')
  }
  
  return classes.join(' ')
})

const handleClick = (event: MouseEvent) => {
  if (props.disabled) {
    event.preventDefault()
    return
  }
  
  emit('click', event)
}
</script>

<style scoped>
.dropdown-item {
  @apply w-full;
  @apply flex items-center gap-2;
  @apply px-4 py-2;
  @apply text-sm text-gray-700;
  @apply text-left;
  @apply hover:bg-gray-100;
  @apply transition-colors duration-150;
  @apply cursor-pointer;
}

.dropdown-item:focus {
  @apply outline-none bg-gray-100;
}

.dropdown-item-disabled {
  @apply opacity-50 cursor-not-allowed;
  @apply hover:bg-transparent;
}

.dropdown-item-divided {
  @apply border-t border-gray-200;
}

.dropdown-item-danger {
  @apply text-red-600;
  @apply hover:bg-red-50;
}

.dropdown-item-icon {
  @apply flex-shrink-0;
  @apply w-4 h-4;
}

.dropdown-item-content {
  @apply flex-1;
}

.dropdown-item-shortcut {
  @apply text-xs text-gray-400;
  @apply ml-auto;
}
</style>
