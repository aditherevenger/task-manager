import { api } from './client'
import type { Task } from './types'

export type TaskFilter = 'all' | 'pending' | 'completed' | 'overdue'

function filterToQuery(filter: TaskFilter): string {
  switch (filter) {
    case 'completed':
      return '?completed=true'
    case 'pending':
      return '?completed=false'
    case 'overdue':
      return '?overdue=true'
    default:
      return ''
  }
}

export function listTasks(filter: TaskFilter): Promise<Task[]> {
  return api<Task[]>(`/api/v1/tasks${filterToQuery(filter)}`)
}

export function getTask(id: number): Promise<Task> {
  return api<Task>(`/api/v1/tasks/${id}`)
}

export function createTask(input: {
  title: string
  description?: string
  due_date?: string
  priority?: string
}): Promise<Task> {
  return api<Task>('/api/v1/tasks', { method: 'POST', json: input })
}

export function updateTask(
  id: number,
  input: { title?: string; description?: string },
): Promise<Task> {
  return api<Task>(`/api/v1/tasks/${id}`, { method: 'PUT', json: input })
}

export function deleteTask(id: number): Promise<void> {
  return api<void>(`/api/v1/tasks/${id}`, { method: 'DELETE' })
}

export function completeTask(id: number): Promise<Task> {
  return api<Task>(`/api/v1/tasks/${id}/complete`, { method: 'PATCH' })
}

export function uncompleteTask(id: number): Promise<Task> {
  return api<Task>(`/api/v1/tasks/${id}/uncomplete`, { method: 'PATCH' })
}

export function setDueDate(id: number, due_date: string): Promise<Task> {
  return api<Task>(`/api/v1/tasks/${id}/due-date`, {
    method: 'PATCH',
    json: { due_date },
  })
}

export function setPriority(id: number, priority: string): Promise<Task> {
  return api<Task>(`/api/v1/tasks/${id}/priority`, {
    method: 'PATCH',
    json: { priority },
  })
}
