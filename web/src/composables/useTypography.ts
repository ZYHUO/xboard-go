import { computed, type ComputedRef } from 'vue'

export type FontSize = 'xs' | 'sm' | 'base' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl' | '5xl'
export type FontWeight = 'normal' | 'medium' | 'semibold' | 'bold'
export type LineHeight = 'tight' | 'normal' | 'relaxed' | 'loose'
export type TextAlign = 'left' | 'center' | 'right' | 'justify'

export interface TypographyConfig {
  size?: FontSize
  weight?: FontWeight
  lineHeight?: LineHeight
  align?: TextAlign
  color?: string
}

// 字体大小映射
const fontSizeMap: Record<FontSize, string> = {
  xs: '0.75rem',    // 12px
  sm: '0.875rem',   // 14px
  base: '1rem',     // 16px
  lg: '1.125rem',   // 18px
  xl: '1.25rem',    // 20px
  '2xl': '1.5rem',  // 24px
  '3xl': '1.875rem', // 30px
  '4xl': '2.25rem',  // 36px
  '5xl': '3rem',     // 48px
}

// 字重映射
const fontWeightMap: Record<FontWeight, number> = {
  normal: 400,
  medium: 500,
  semibold: 600,
  bold: 700,
}

// 行高映射
const lineHeightMap: Record<LineHeight, number> = {
  tight: 1.25,
  normal: 1.5,
  relaxed: 1.75,
  loose: 2,
}

// 预定义的排版样式
export const typographyPresets = {
  h1: { size: '5xl' as FontSize, weight: 'bold' as FontWeight, lineHeight: 'tight' as LineHeight },
  h2: { size: '4xl' as FontSize, weight: 'bold' as FontWeight, lineHeight: 'tight' as LineHeight },
  h3: { size: '3xl' as FontSize, weight: 'semibold' as FontWeight, lineHeight: 'tight' as LineHeight },
  h4: { size: '2xl' as FontSize, weight: 'semibold' as FontWeight, lineHeight: 'normal' as LineHeight },
  h5: { size: 'xl' as FontSize, weight: 'semibold' as FontWeight, lineHeight: 'normal' as LineHeight },
  h6: { size: 'lg' as FontSize, weight: 'semibold' as FontWeight, lineHeight: 'normal' as LineHeight },
  body: { size: 'base' as FontSize, weight: 'normal' as FontWeight, lineHeight: 'normal' as LineHeight },
  bodyLarge: { size: 'lg' as FontSize, weight: 'normal' as FontWeight, lineHeight: 'relaxed' as LineHeight },
  bodySmall: { size: 'sm' as FontSize, weight: 'normal' as FontWeight, lineHeight: 'normal' as LineHeight },
  caption: { size: 'xs' as FontSize, weight: 'normal' as FontWeight, lineHeight: 'normal' as LineHeight },
  button: { size: 'base' as FontSize, weight: 'medium' as FontWeight, lineHeight: 'tight' as LineHeight },
  label: { size: 'sm' as FontSize, weight: 'medium' as FontWeight, lineHeight: 'normal' as LineHeight },
}

export function useTypography(config: TypographyConfig = {}) {
  const fontSize = computed(() => {
    return config.size ? fontSizeMap[config.size] : fontSizeMap.base
  })
  
  const fontWeight = computed(() => {
    return config.weight ? fontWeightMap[config.weight] : fontWeightMap.normal
  })
  
  const lineHeight = computed(() => {
    return config.lineHeight ? lineHeightMap[config.lineHeight] : lineHeightMap.normal
  })
  
  const textAlign = computed(() => {
    return config.align || 'left'
  })
  
  const color = computed(() => {
    return config.color || 'var(--text-primary)'
  })
  
  // 生成 CSS 样式对象
  const style = computed(() => ({
    fontSize: fontSize.value,
    fontWeight: fontWeight.value,
    lineHeight: lineHeight.value,
    textAlign: textAlign.value,
    color: color.value,
  }))
  
  // 生成 CSS 类名
  const className = computed(() => {
    const classes: string[] = []
    
    if (config.size) classes.push(`text-${config.size}`)
    if (config.weight) classes.push(`font-${config.weight}`)
    if (config.lineHeight) classes.push(`leading-${config.lineHeight}`)
    if (config.align) classes.push(`text-${config.align}`)
    
    return classes.join(' ')
  })
  
  return {
    fontSize,
    fontWeight,
    lineHeight,
    textAlign,
    color,
    style,
    className,
  }
}

// 预设排版样式的辅助函数
export function useHeading(level: 1 | 2 | 3 | 4 | 5 | 6) {
  const presetKey = `h${level}` as keyof typeof typographyPresets
  return useTypography(typographyPresets[presetKey])
}

export function useBodyText(variant: 'default' | 'large' | 'small' = 'default') {
  const presetMap = {
    default: typographyPresets.body,
    large: typographyPresets.bodyLarge,
    small: typographyPresets.bodySmall,
  }
  return useTypography(presetMap[variant])
}

export function useCaption() {
  return useTypography(typographyPresets.caption)
}

export function useButtonText() {
  return useTypography(typographyPresets.button)
}

export function useLabelText() {
  return useTypography(typographyPresets.label)
}

// 计算文本对比度
export function getTextContrast(backgroundColor: string): 'light' | 'dark' {
  // 简化的对比度计算
  // 实际应用中应该使用更精确的算法
  const hex = backgroundColor.replace('#', '')
  const r = parseInt(hex.substr(0, 2), 16)
  const g = parseInt(hex.substr(2, 2), 16)
  const b = parseInt(hex.substr(4, 2), 16)
  
  // 计算相对亮度
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255
  
  return luminance > 0.5 ? 'dark' : 'light'
}

// 响应式字体大小
export function useResponsiveFontSize(
  mobile: FontSize,
  tablet: FontSize,
  desktop: FontSize
): ComputedRef<string> {
  return computed(() => {
    // 在实际应用中，这里应该根据屏幕尺寸返回不同的值
    // 这里简化为返回桌面尺寸
    return fontSizeMap[desktop]
  })
}

// 文本截断
export interface TruncateOptions {
  lines?: number
  suffix?: string
}

export function useTruncate(options: TruncateOptions = {}) {
  const lines = options.lines || 1
  const suffix = options.suffix || '...'
  
  const style = computed(() => {
    if (lines === 1) {
      return {
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap' as const,
      }
    } else {
      return {
        display: '-webkit-box',
        WebkitLineClamp: lines,
        WebkitBoxOrient: 'vertical' as const,
        overflow: 'hidden',
      }
    }
  })
  
  return { style, suffix }
}
