import { useEffect, useMemo, useState } from 'react'
import { ApiError } from '../api/client'
import { getStats } from '../api/stats'
import type { TaskStats } from '../api/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'

export default function StatsPage() {
  const [stats, setStats] = useState<TaskStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const entries = useMemo(() => {
    if (!stats) return [] as Array<[string, number]>
    return Object.entries(stats)
  }, [stats])

  useEffect(() => {
    async function run() {
      setLoading(true)
      setError(null)
      try {
        const s = await getStats()
        setStats(s)
      } catch (e) {
        const msg = e instanceof ApiError ? e.message : 'Failed to load stats'
        setError(msg)
      } finally {
        setLoading(false)
      }
    }

    void run()
  }, [])

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">Stats</h1>
        <p className="text-sm text-muted-foreground">From `GET /api/v1/stats`.</p>
      </div>

      {error ? (
        <Card className="border-destructive/40">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Something went wrong</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
        </Card>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Overview</CardTitle>
          <CardDescription>High-level summary of your tasks.</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="h-20 animate-pulse rounded-lg bg-muted" />
              ))}
            </div>
          ) : entries.length === 0 ? (
            <div className="text-sm text-muted-foreground">No stats.</div>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {entries.map(([k, v]) => (
                <div key={k} className="rounded-xl border bg-card p-4">
                  <div className="text-3xl font-semibold leading-none tracking-tight">{v}</div>
                  <div className="mt-2 text-xs text-muted-foreground">{k}</div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
