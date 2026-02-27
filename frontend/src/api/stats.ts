import { api } from './client'
import type { TaskStats } from './types'

export function getStats(): Promise<TaskStats> {
  return api<TaskStats>('/api/v1/stats')
}
