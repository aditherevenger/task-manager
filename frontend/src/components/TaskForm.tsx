import { useMemo, useState } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Input } from './ui/input'
import { Textarea } from './ui/textarea'

type Props = {
  onSubmit: (input: {
    title: string
    description?: string
    due_date?: string
    priority?: string
  }) => Promise<void>
}

const priorityOptions = ['highest', 'high', 'medium', 'low', 'lowest'] as const

export default function TaskForm({ onSubmit }: Props) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [dueDate, setDueDate] = useState('')
  const [priority, setPriority] = useState<string>('medium')

  const [submitting, setSubmitting] = useState(false)
  const canSubmit = useMemo(() => title.trim().length > 0 && !submitting, [title, submitting])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!canSubmit) return

    setSubmitting(true)
    try {
      await onSubmit({
        title: title.trim(),
        description: description.trim() || undefined,
        due_date: dueDate || undefined,
        priority: priority || undefined,
      })
      setTitle('')
      setDescription('')
      setDueDate('')
      setPriority('medium')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle>Create task</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-3 md:grid-cols-3">
            <div className="space-y-1 md:col-span-2">
              <div className="text-sm font-medium">Title</div>
              <Input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g. Finish frontend"
              />
            </div>

            <div className="space-y-1">
              <div className="text-sm font-medium">Due date</div>
              <Input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
            </div>

            <div className="space-y-1">
              <div className="text-sm font-medium">Priority</div>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={priority}
                onChange={(e) => setPriority(e.target.value)}
              >
                {priorityOptions.map((p) => (
                  <option key={p} value={p}>
                    {p}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-1 md:col-span-3">
              <div className="text-sm font-medium">Description</div>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="optional"
                rows={3}
              />
            </div>
          </div>

          <div className="flex justify-end">
            <Button type="submit" disabled={!canSubmit}>
              {submitting ? 'Creating…' : 'Create'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
