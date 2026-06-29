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

export function recordField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)
  return isRecord(value) ? value : {}
}
