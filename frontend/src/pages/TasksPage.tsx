import { useEffect, useMemo, useState } from 'react'
import type { Task } from '../api/types'
import {
  completeTask,
  createTask,
  deleteTask,
  listTasks,
  setDueDate,
  setPriority,
  uncompleteTask,
  updateTask,
} from '../api/tasks'
import type { TaskFilter } from '../api/tasks'
import { ApiError } from '../api/client'
import TaskForm from '../components/TaskForm'
import { formatDate, priorityLabel, toInputDate } from '../lib/format'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '../components/ui/dialog'
import { Input } from '../components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '../components/ui/tabs'
import { Textarea } from '../components/ui/textarea'
import { Calendar, CheckCircle2, Circle, Pencil, Trash2 } from 'lucide-react'

type RowState = {
  editing: boolean
  title: string
  description: string
}

function defaultRowState(t: Task): RowState {
  return {
    editing: false,
    title: t.title,
    description: t.description,
  }
}

export default function TasksPage() {
  const [filter, setFilter] = useState<TaskFilter>('all')
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [rowState, setRowState] = useState<Record<number, RowState>>({})

  const sortedTasks = useMemo(() => {
    const copy = [...tasks]
    copy.sort((a, b) => a.id - b.id)
    return copy
  }, [tasks])

  const [editId, setEditId] = useState<number | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function refresh(currentFilter: TaskFilter) {
    setLoading(true)
    setError(null)
    try {
      const data = await listTasks(currentFilter)
      setTasks(data)
      setRowState((prev) => {
        const next: Record<number, RowState> = { ...prev }
        for (const t of data) {
          if (!next[t.id]) next[t.id] = defaultRowState(t)
        }
        return next
      })
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Failed to load tasks'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void refresh(filter)
  }, [filter])

  async function handleCreate(input: {
    title: string
    description?: string
    due_date?: string
    priority?: string
  }) {
    setError(null)
    try {
      const t = await createTask(input)
      setTasks((prev) => [t, ...prev])
      setRowState((prev) => ({ ...prev, [t.id]: defaultRowState(t) }))
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Failed to create task'
      setError(msg)
      throw e
    }
  }

  async function handleDelete(id: number) {
    setError(null)
    try {
      await deleteTask(id)
      setTasks((prev) => prev.filter((t) => t.id !== id))
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Failed to delete task'
      setError(msg)
    }
  }

  async function handleToggleComplete(t: Task) {
    setError(null)
    try {
      const updated = t.completed ? await uncompleteTask(t.id) : await completeTask(t.id)
      setTasks((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Failed to update task'
      setError(msg)
    }
  }

  async function handleSaveEdit(id: number) {
    const st = rowState[id]
    if (!st) return

    setError(null)
    try {
      const updated = await updateTask(id, {
        title: st.title.trim(),
        description: st.description.trim(),
      })
      setTasks((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
      setRowState((prev) => ({ ...prev, [id]: { ...prev[id], editing: false } }))
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Failed to update task'
      setError(msg)
    }
  }

  async function handleSetPriority(id: number, priority: string) {
    setError(null)
    try {
      const updated = await setPriority(id, priority)
      setTasks((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Failed to set priority'
      setError(msg)
    }
  }

  async function handleSetDueDate(id: number, due_date: string) {
    setError(null)
    try {
      const updated = await setDueDate(id, due_date)
      setTasks((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Failed to set due date'
      setError(msg)
    }
  }

  function updateRowField(id: number, patch: Partial<RowState>) {
    setRowState((prev) => ({
      ...prev,
      [id]: {
        ...(prev[id] ?? { editing: false, title: '', description: '' }),
        ...patch,
      },
    }))
  }

  const editTask = useMemo(() => {
    if (editId === null) return null
    return sortedTasks.find((t) => t.id === editId) ?? null
  }, [editId, sortedTasks])

  const editState = useMemo(() => {
    if (!editTask) return null
    return rowState[editTask.id] ?? defaultRowState(editTask)
  }, [editTask, rowState])

  function openEdit(t: Task) {
    setRowState((prev) => ({ ...prev, [t.id]: prev[t.id] ?? defaultRowState(t) }))
    setEditId(t.id)
  }

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Tasks</h1>
        <p className="text-sm text-muted-foreground">Manage tasks via your Go API (`/api/v1/*`).</p>
      </div>

      {error ? (
        <Card className="border-destructive/40">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Something went wrong</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
        </Card>
      ) : null}

      <div className="flex flex-col gap-4">
        <Tabs value={filter} onValueChange={(v) => setFilter(v as TaskFilter)}>
          <TabsList>
            <TabsTrigger value="all">All</TabsTrigger>
            <TabsTrigger value="pending">Pending</TabsTrigger>
            <TabsTrigger value="completed">Completed</TabsTrigger>
            <TabsTrigger value="overdue">Overdue</TabsTrigger>
          </TabsList>
        </Tabs>

        <TaskForm onSubmit={handleCreate} />
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="text-sm font-medium">Task list</div>
          <div className="text-xs text-muted-foreground">
            {loading ? 'Loading…' : `${sortedTasks.length} task(s)`}
          </div>
        </div>

        {loading ? (
          <div className="grid gap-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <Card key={i}>
                <CardHeader className="pb-3">
                  <div className="h-4 w-2/3 animate-pulse rounded bg-muted" />
                  <div className="mt-2 h-3 w-1/2 animate-pulse rounded bg-muted" />
                </CardHeader>
                <CardContent>
                  <div className="h-8 w-full animate-pulse rounded bg-muted" />
                </CardContent>
              </Card>
            ))}
          </div>
        ) : sortedTasks.length === 0 ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-base">No tasks</CardTitle>
              <CardDescription>Create your first task above.</CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <div className="grid gap-3">
            {sortedTasks.map((t) => {
              const due = toInputDate(t.due_date)

              return (
                <Card key={t.id} className={t.is_overdue ? 'border-destructive/40' : undefined}>
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <div className="text-sm text-muted-foreground">#{t.id}</div>
                          {t.completed ? (
                            <Badge variant="secondary">Completed</Badge>
                          ) : (
                            <Badge variant="outline">Pending</Badge>
                          )}
                          {t.is_overdue ? <Badge variant="destructive">Overdue</Badge> : null}
                        </div>
                        <CardTitle className={t.completed ? 'text-base line-through opacity-70' : 'text-base'}>
                          {t.title}
                        </CardTitle>
                        {t.description ? <CardDescription>{t.description}</CardDescription> : null}
                      </div>

                      <div className="flex shrink-0 items-center gap-2">
                        <Button variant="outline" size="sm" onClick={() => openEdit(t)}>
                          <Pencil className="h-4 w-4" />
                          Edit
                        </Button>
                        <Button variant={t.completed ? 'secondary' : 'default'} size="sm" onClick={() => void handleToggleComplete(t)}>
                          {t.completed ? <Circle className="h-4 w-4" /> : <CheckCircle2 className="h-4 w-4" />}
                          {t.completed ? 'Uncomplete' : 'Complete'}
                        </Button>
                        <Button variant="destructive" size="sm" onClick={() => setDeleteId(t.id)}>
                          <Trash2 className="h-4 w-4" />
                          Delete
                        </Button>
                      </div>
                    </div>
                  </CardHeader>

                  <CardContent>
                    <div className="grid gap-3 md:grid-cols-3">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <Calendar className="h-4 w-4" />
                          Due date
                        </div>
                        <div className="flex items-center gap-2">
                          <Input
                            type="date"
                            value={due}
                            onChange={(e) => {
                              const v = e.target.value
                              if (v) void handleSetDueDate(t.id, v)
                            }}
                          />
                        </div>
                        <div className="text-xs text-muted-foreground">{formatDate(t.due_date)}</div>
                      </div>

                      <div className="space-y-1">
                        <div className="text-xs text-muted-foreground">Priority</div>
                        <select
                          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                          value={String(t.priority)}
                          onChange={(e) => void handleSetPriority(t.id, e.target.value)}
                        >
                          <option value="1">Highest</option>
                          <option value="2">High</option>
                          <option value="3">Medium</option>
                          <option value="4">Low</option>
                          <option value="5">Lowest</option>
                        </select>
                        <div className="text-xs text-muted-foreground">{priorityLabel(t.priority)}</div>
                      </div>

                      <div className="space-y-1">
                        <div className="text-xs text-muted-foreground">Actions</div>
                        <div className="flex flex-wrap gap-2">
                          <Button variant="outline" size="sm" onClick={() => void refresh(filter)}>
                            Refresh
                          </Button>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        )}
      </div>

      <Dialog open={editId !== null} onOpenChange={(open) => (!open ? setEditId(null) : null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit task</DialogTitle>
            <DialogDescription>Update the title and description.</DialogDescription>
          </DialogHeader>

          {editTask && editState ? (
            <div className="grid gap-3">
              <div className="space-y-1">
                <div className="text-sm font-medium">Title</div>
                <Input
                  value={editState.title}
                  onChange={(e) => updateRowField(editTask.id, { title: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <div className="text-sm font-medium">Description</div>
                <Textarea
                  value={editState.description}
                  onChange={(e) => updateRowField(editTask.id, { description: e.target.value })}
                  rows={4}
                />
              </div>
            </div>
          ) : null}

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditId(null)}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                if (editId !== null) void handleSaveEdit(editId).then(() => setEditId(null))
              }}
            >
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteId !== null} onOpenChange={(open) => (!open ? setDeleteId(null) : null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete task?</DialogTitle>
            <DialogDescription>This action cannot be undone.</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                if (deleteId !== null) void handleDelete(deleteId).then(() => setDeleteId(null))
              }}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
