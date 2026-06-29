export function cleanBackendMessage(message: string) {
  return message.replace(/^invalid input:\s*/i, '').trim()
}
