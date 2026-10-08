// Additive patch for useSSEStream.ts - copy into applyEvent switch
// No read of the file needed; just apply this 6-line addition.

export const additiveCases = `
    case 'relay_step':
      useStreamStore.getState().appendRelayStep(chatId, (event as any).data)
      break
    case 'relay_failed':
      useStreamStore.getState().setRelayFailed(chatId, (event as any).data)
      break
`

// In streamStore.ts add:
// relayFailed?: { runId: string; failedAtIndex: number; reason: string; retryUntil: string }
// setRelayFailed(chatId, payload) { set state relayFailed = payload }
// And ensure ExpertResponse red card: mode REFUSE + reason "provider unavailable" renders same row
