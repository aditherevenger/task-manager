export function toInputDate(iso: string | undefined): string {
  if (!iso) return ''
  if (iso.startsWith('0001-01-01')) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function formatDate(iso: string | undefined): string {
  const input = toInputDate(iso)
  return input || '-'
}

export function priorityLabel(p: number): string {
  switch (p) {
    case 1:
      return 'Highest'
    case 2:
      return 'High'
    case 3:
      return 'Medium'
    case 4:
      return 'Low'
    case 5:
      return 'Lowest'
    default:
      return String(p)
  }
}
