<template>
  <Teleport to="body">
    <Transition name="modal" @after-leave="handleAfterLeave">
      <div
        v-if="modelValue"
        class="modal-overlay"
        :class="overlayClasses"
        @click="handleOverlayClick"
      >
        <div
          class="modal-container"
          :class="containerClasses"
          :style="modalStyle"
          @click.stop
        >
          <!-- Header -->
          <div v-if="$slots.header || title" class="modal-header">
            <slot name="header">
              <h3 class="modal-title">{{ title }}</h3>
            </slot>
            <button
              v-if="closable"
              class="modal-close"
              @click="handleClose"
              aria-label="关闭"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="modal-body" :class="bodyClasses">
            <slot />
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer || showFooter" class="modal-footer">
            <slot name="footer">
              <button
                v-if="showCancel"
                class="btn btn-secondary"
                @click="handleCancel"
              >
                {{ cancelText }}
              </button>
              <button
                v-if="showConfirm"
                class="btn btn-primary"
                :disabled="confirmLoading"
                @click="handleConfirm"
              >
                <span v-if="confirmLoading" class="btn-loading">
                  <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </span>
                {{ confirmText }}
              </button>
            </slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch, onMounted, onUnmounted } from 'vue'

interface Props {
  modelValue: boolean
  title?: string
  width?: string
  closable?: boolean
  maskClosable?: boolean
  showFooter?: boolean
  showCancel?: boolean
  showConfirm?: boolean
  cancelText?: string
  confirmText?: string
  confirmLoading?: boolean
  centered?: boolean
  fullscreen?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  closable: true,
  maskClosable: true,
  showFooter: true,
  showCancel: true,
  showConfirm: true,
  cancelText: '取消',
  confirmText: '确定',
  confirmLoading: false,
  centered: false,
  fullscreen: false,
  width: '520px',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  cancel: []
  confirm: []
}>()

const overlayClasses = computed(() => {
  const classes = []
  if (props.centered) classes.push('modal-overlay-centered')
  return classes.join(' ')
})

const containerClasses = computed(() => {
  const classes = []
  if (props.fullscreen) classes.push('modal-container-fullscreen')
  return classes.join(' ')
})

const bodyClasses = computed(() => {
  const classes = []
  if (!props.showFooter && !props.$slots.footer) classes.push('modal-body-no-footer')
  return classes.join(' ')
})

const modalStyle = computed(() => {
  if (props.fullscreen) return {}
  return {
    width: props.width,
    maxWidth: '90vw',
  }
})

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleCancel = () => {
  emit('cancel')
  handleClose()
}

const handleConfirm = () => {
  emit('confirm')
}

const handleOverlayClick = () => {
  if (props.maskClosable) {
    handleClose()
  }
}

const handleAfterLeave = () => {
  // 清理工作
}

// 键盘事件处理
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && props.modelValue && props.closable) {
    handleClose()
  }
}

// 防止背景滚动
watch(() => props.modelValue, (value) => {
  if (value) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = ''
})
</script>

<style scoped>
.modal-overlay {
  @apply fixed inset-0 z-modal;
  @apply bg-black bg-opacity-50;
  @apply flex items-start justify-center;
  @apply overflow-y-auto;
  @apply p-4;
}

.modal-overlay-centered {
  @apply items-center;
}

.modal-container {
  @apply relative;
  @apply bg-white rounded-md;
  @apply shadow-2xl;
  @apply my-8;
  @apply w-full;
}

.modal-container-fullscreen {
  @apply w-full h-full m-0 rounded-none;
}

.modal-header {
  @apply flex items-center justify-between;
  @apply px-6 py-4;
  @apply border-b border-gray-200;
}

.modal-title {
  @apply text-lg font-semibold text-gray-900;
}

.modal-close {
  @apply p-1 -mr-1;
  @apply text-gray-400 hover:text-gray-600;
  @apply rounded-md hover:bg-gray-100;
  @apply transition-colors duration-150;
}

.modal-body {
  @apply px-6 py-4;
  @apply text-gray-900;
}

.modal-body-no-footer {
  @apply pb-6;
}

.modal-footer {
  @apply flex items-center justify-end gap-3;
  @apply px-6 py-4;
  @apply border-t border-gray-200;
}

/* 动画 */
.modal-enter-active,
.modal-leave-active {
  @apply transition-opacity duration-200;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
  @apply transition-all duration-200;
}

.modal-enter-from,
.modal-leave-to {
  @apply opacity-0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  @apply scale-95 opacity-0;
}

/* 按钮样式 */
.btn {
  @apply inline-flex items-center justify-center gap-2;
  @apply px-4 py-2 text-sm font-medium;
  @apply rounded-md;
  @apply transition-all duration-150;
  @apply focus:outline-none focus:ring-2 focus:ring-offset-2;
}

.btn-primary {
  @apply bg-black text-white;
  @apply hover:bg-gray-800;
  @apply focus:ring-gray-500;
}

.btn-secondary {
  @apply bg-white text-gray-700;
  @apply border border-gray-300;
  @apply hover:bg-gray-50;
  @apply focus:ring-gray-400;
}

.btn-loading {
  @apply inline-flex items-center;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}
</style>
