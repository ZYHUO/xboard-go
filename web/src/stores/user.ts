import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api'

export interface User {
  id: number
  email: string
  uuid: string
  token: string
  balance: number
  plan_id: number | null
  group_id: number | null
  transfer_enable: number
  u: number
  d: number
  expired_at: number | null
  is_admin: boolean
  is_staff: boolean
  created_at: number
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.is_admin ?? false)

  const usedTraffic = computed(() => (user.value?.u ?? 0) + (user.value?.d ?? 0))
  const totalTraffic = computed(() => user.value?.transfer_enable ?? 0)
  const trafficPercent = computed(() => {
    if (totalTraffic.value === 0) return 0
    return Math.min(100, Math.round((usedTraffic.value / totalTraffic.value) * 100))
  })

  async function login(email: string, password: string) {
    const res = await api.post('/api/v1/guest/login', { email, password })
    token.value = res.data.data.token
    localStorage.setItem('token', token.value!)
    await fetchUser()
    return res.data
  }

  async function register(email: string, password: string, inviteCode?: string) {
    const res = await api.post('/api/v1/guest/register', { 
      email, 
      password,
      invite_code: inviteCode 
    })
    token.value = res.data.data.token
    localStorage.setItem('token', token.value!)
    await fetchUser()
    return res.data
  }

  async function fetchUser() {
    if (!token.value) return
    try {
      const res = await api.get('/api/v1/user/info')
      user.value = res.data.data
    } catch (e) {
      logout()
    }
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }

  // 初始化时获取用户信息
  if (token.value) {
    fetchUser()
  }

  return {
    user,
    token,
    isLoggedIn,
    isAdmin,
    usedTraffic,
    totalTraffic,
    trafficPercent,
    login,
    register,
    fetchUser,
    logout,
  }
})
