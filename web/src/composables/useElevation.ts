import { ref, computed, type Ref } from 'vue'

export type ElevationLevel = 0 | 1 | 2 | 3 | 4 | 5 | 6

export interface ElevationConfig {
  level: ElevationLevel
  interactive?: boolean
  hoverLevel?: ElevationLevel
  activeLevel?: ElevationLevel
}

// 阴影定义 - 创建视觉深度
const elevationShadows: Record<ElevationLevel, string> = {
  0: 'none',
  1: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  2: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
  3: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  4: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
  5: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
  6: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
}

// Z-index 层次定义
export const zIndexLevels = {
  base: 0,
  dropdown: 1000,
  sticky: 1020,
  fixed: 1030,
  modalBackdrop: 1040,
  modal: 1050,
  popover: 1060,
  tooltip: 1070,
  notification: 1080,
}

export function useElevation(initialLevel: ElevationLevel = 0, config?: Partial<ElevationConfig>) {
  const currentLevel = ref<ElevationLevel>(initialLevel)
  const isHovered = ref(false)
  const isActive = ref(false)
  
  const interactive = config?.interactive ?? false
  const hoverLevel = config?.hoverLevel ?? (initialLevel + 1 as ElevationLevel)
  const activeLevel = config?.activeLevel ?? initialLevel
  
  // 计算当前应该显示的层次
  const effectiveLevel = computed<ElevationLevel>(() => {
    if (!interactive) return currentLevel.value
    
    if (isActive.value) return activeLevel
    if (isHovered.value) return hoverLevel
    return currentLevel.value
  })
  
  // 获取当前阴影样式
  const shadow = computed(() => elevationShadows[effectiveLevel.value])
  
  // 获取 CSS 样式对象
  const style = computed(() => ({
    boxShadow: shadow.value,
    transition: interactive ? 'box-shadow 200ms ease' : undefined,
  }))
  
  // 获取 CSS 类名
  const className = computed(() => {
    const classes = [`elevation-${effectiveLevel.value}`]
    if (interactive) classes.push('elevation-interactive')
    return classes.join(' ')
  })
  
  // 设置层次级别
  const setLevel = (level: ElevationLevel) => {
    currentLevel.value = level
  }
  
  // 增加层次
  const raise = (amount: number = 1) => {
    const newLevel = Math.min(6, currentLevel.value + amount) as ElevationLevel
    currentLevel.value = newLevel
  }
  
  // 降低层次
  const lower = (amount: number = 1) => {
    const newLevel = Math.max(0, currentLevel.value - amount) as ElevationLevel
    currentLevel.value = newLevel
  }
  
  // 交互事件处理
  const onMouseEnter = () => {
    if (interactive) isHovered.value = true
  }
  
  const onMouseLeave = () => {
    if (interactive) isHovered.value = false
  }
  
  const onMouseDown = () => {
    if (interactive) isActive.value = true
  }
  
  const onMouseUp = () => {
    if (interactive) isActive.value = false
  }
  
  // 事件监听器对象
  const listeners = interactive ? {
    onMouseenter: onMouseEnter,
    onMouseleave: onMouseLeave,
    onMousedown: onMouseDown,
    onMouseup: onMouseUp,
  } : {}
  
  return {
    // 状态
    currentLevel: computed(() => currentLevel.value),
    effectiveLevel,
    isHovered: computed(() => isHovered.value),
    isActive: computed(() => isActive.value),
    
    // 样式
    shadow,
    style,
    className,
    
    // 方法
    setLevel,
    raise,
    lower,
    
    // 事件监听器
    listeners,
    
    // Z-index 辅助
    zIndex: zIndexLevels,
  }
}

// 预定义的组件层次级别
export const componentElevations = {
  card: 1,
  cardHover: 2,
  button: 0,
  buttonHover: 1,
  input: 0,
  inputFocus: 1,
  dropdown: 3,
  modal: 5,
  tooltip: 6,
  notification: 4,
} as const

// 创建交互式层次
export function useInteractiveElevation(
  baseLevel: ElevationLevel = 1,
  hoverLevel: ElevationLevel = 2,
  activeLevel: ElevationLevel = 1
) {
  return useElevation(baseLevel, {
    interactive: true,
    hoverLevel,
    activeLevel,
  })
}

// 为卡片创建层次
export function useCardElevation() {
  return useInteractiveElevation(
    componentElevations.card as ElevationLevel,
    componentElevations.cardHover as ElevationLevel,
    componentElevations.card as ElevationLevel
  )
}

// 为按钮创建层次
export function useButtonElevation() {
  return useInteractiveElevation(
    componentElevations.button as ElevationLevel,
    componentElevations.buttonHover as ElevationLevel,
    componentElevations.button as ElevationLevel
  )
}
