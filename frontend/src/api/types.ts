export type Task = {
  id: number
  title: string
  description: string
  completed: boolean
  created_at: string
  completed_at: string
  due_date: string
  priority: number
  is_overdue: boolean
}

export type TaskStats = Record<string, number>
