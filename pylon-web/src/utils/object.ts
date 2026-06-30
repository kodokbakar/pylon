export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

export function valueField(record: Record<string, unknown>, ...keys: string[]) {
  for (const key of keys) {
    if (key in record) {
      return record[key]
    }
  }

  return undefined
}

export function stringField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)
  return typeof value === 'string' ? value.trim() : ''
}

export function booleanField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)
  return typeof value === 'boolean' ? value : false
}

export function dateField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)

  if (typeof value === 'string' && value.trim()) {
    return value.trim()
  }

  if (isRecord(value)) {
    const seconds = numberField(value, 'seconds')
    const nanos = numberField(value, 'nanos')
    if (seconds > 0) {
      return new Date(seconds * 1000 + Math.floor(nanos / 1_000_000)).toISOString()
    }
  }

  return new Date().toISOString()
}

export function recordField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)
  return isRecord(value) ? value : {}
}

function numberField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)
  if (typeof value === 'number') {
    return value
  }

  if (typeof value === 'string') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : 0
  }

  return 0
}
