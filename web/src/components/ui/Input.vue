<template>
  <div class="input-wrapper" :class="wrapperClasses">
    <!-- 标准标签 -->
    <label 
      v-if="label && !floatingLabel" 
      :for="inputId" 
      class="input-label"
    >
      {{ label }}
      <span v-if="required" class="text-red-600">*</span>
    </label>
    
    <!-- 输入框容器 -->
    <div class="input-container" :class="containerClasses">
      <!-- 前缀图标 -->
      <span v-if="$slots.prefix || prefixIcon" class="input-prefix">
        <slot name="prefix">{{ prefixIcon }}</slot>
      </span>
      
      <!-- 输入框 -->
      <input
        :id="inputId"
        :type="type"
        :value="modelValue"
        :placeholder="floatingLabel ? ' ' : placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :class="inputClasses"
        @input="handleInput"
        @blur="handleBlur"
        @focus="handleFocus"
      />
      
      <!-- 浮动标签 -->
      <label 
        v-if="label && floatingLabel" 
        :for="inputId" 
        class="input-label-floating"
        :class="floatingLabelClasses"
      >
        {{ label }}
        <span v-if="required" class="text-red-600">*</span>
      </label>
      
      <!-- 后缀图标 -->
      <span v-if="$slots.suffix || suffixIcon" class="input-suffix">
        <slot name="suffix">{{ suffixIcon }}</slot>
      </span>
      
      <!-- 清除按钮 -->
      <button
        v-if="clearable && modelValue"
        type="button"
        class="input-clear"
        @click="handleClear"
      >
        ×
      </button>
    </div>
    
    <!-- 错误/提示信息 -->
    <p v-if="error" class="input-message input-error">{{ error }}</p>
    <p v-else-if="hint" class="input-message input-hint">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  modelValue?: string | number
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'search'
  label?: string
  placeholder?: string
  error?: string
  hint?: string
  disabled?: boolean
  readonly?: boolean
  required?: boolean
  clearable?: boolean
  floatingLabel?: boolean
  prefixIcon?: string
  suffixIcon?: string
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  readonly: false,
  required: false,
  clearable: false,
  floatingLabel: false,
  size: 'md',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  blur: [event: FocusEvent]
  focus: [event: FocusEvent]
  clear: []
}>()

const inputId = ref(`input-${Math.random().toString(36).substr(2, 9)}`)
const isFocused = ref(false)

const hasValue = computed(() => {
  return props.modelValue !== undefined && props.modelValue !== null && props.modelValue !== ''
})

const wrapperClasses = computed(() => {
  const classes = []
  if (props.disabled) classes.push('input-wrapper-disabled')
  return classes.join(' ')
})

const containerClasses = computed(() => {
  const classes = ['input-container-base']
  
  // 状态样式
  if (props.error) {
    classes.push('input-container-error')
  } else if (isFocused.value) {
    classes.push('input-container-focused')
  }
  
  if (props.disabled) {
    classes.push('input-container-disabled')
  }
  
  return classes.join(' ')
})

const inputClasses = computed(() => {
  const classes = ['input']
  
  // 尺寸样式
  const sizeClasses = {
    sm: 'px-3 py-2 text-sm min-h-[36px]',
    md: 'px-4 py-2.5 text-base min-h-[40px]',
    lg: 'px-5 py-3 text-lg min-h-[48px]',
  }
  classes.push(sizeClasses[props.size])
  
  // 浮动标签时的额外 padding
  if (props.floatingLabel) {
    classes.push('pt-6 pb-2')
  }
  
  // 前缀/后缀图标时的 padding 调整
  if (props.prefixIcon || props.$slots.prefix) {
    classes.push('pl-10')
  }
  if (props.suffixIcon || props.$slots.suffix || props.clearable) {
    classes.push('pr-10')
  }
  
  return classes.join(' ')
})

const floatingLabelClasses = computed(() => {
  const classes = []
  
  if (isFocused.value || hasValue.value) {
    classes.push('input-label-floating-active')
  }
  
  return classes.join(' ')
})

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

const handleBlur = (event: FocusEvent) => {
  isFocused.value = false
  emit('blur', event)
}

const handleFocus = (event: FocusEvent) => {
  isFocused.value = true
  emit('focus', event)
}

const handleClear = () => {
  emit('update:modelValue', '')
  emit('clear')
}
</script>

<style scoped>
.input-wrapper {
  @apply w-full;
}

.input-wrapper-disabled {
  @apply opacity-60;
}

/* 标准标签 */
.input-label {
  @apply block text-sm font-medium text-gray-700 mb-2;
}

/* 输入框容器 */
.input-container {
  @apply relative;
}

.input-container-base {
  @apply rounded-md border-2 border-gray-300;
  @apply transition-all duration-200 ease-in-out;
}

.input-container-focused {
  @apply border-black ring-2 ring-black ring-opacity-20;
}

.input-container-error {
  @apply border-red-600 ring-2 ring-red-600 ring-opacity-20;
}

.input-container-disabled {
  @apply bg-gray-100 cursor-not-allowed;
}

/* 输入框 */
.input {
  @apply w-full bg-transparent;
  @apply text-gray-900 placeholder-gray-400;
  @apply focus:outline-none;
  @apply disabled:cursor-not-allowed;
  @apply transition-all duration-200 ease-in-out;
}

/* 浮动标签 */
.input-label-floating {
  @apply absolute left-4 top-1/2 -translate-y-1/2;
  @apply text-gray-500 pointer-events-none;
  @apply transition-all duration-200 ease-in-out;
  @apply origin-left;
}

.input-label-floating-active {
  @apply top-3 translate-y-0 text-xs;
  @apply text-gray-700;
}

/* 前缀/后缀图标 */
.input-prefix,
.input-suffix {
  @apply absolute top-1/2 -translate-y-1/2;
  @apply text-gray-400;
  @apply pointer-events-none;
}

.input-prefix {
  @apply left-3;
}

.input-suffix {
  @apply right-3;
}

/* 清除按钮 */
.input-clear {
  @apply absolute right-3 top-1/2 -translate-y-1/2;
  @apply w-5 h-5 flex items-center justify-center;
  @apply text-gray-400 hover:text-gray-600;
  @apply rounded-full hover:bg-gray-100;
  @apply transition-colors duration-150;
  @apply cursor-pointer;
}

/* 消息文本 */
.input-message {
  @apply mt-2 text-sm;
}

.input-error {
  @apply text-red-600;
}

.input-hint {
  @apply text-gray-500;
}
</style>
