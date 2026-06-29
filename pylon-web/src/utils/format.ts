export function getInitial(name: string) {
  const trimmedName = name.trim()
  return trimmedName[0]?.toUpperCase() ?? '?'
}
