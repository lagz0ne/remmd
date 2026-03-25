import { wsconnect, type NatsConnection } from '@nats-io/nats-core'

let nc: NatsConnection | null = null
let connecting = false

/** Get or create singleton NATS connection via WebSocket. */
export async function getNatsConnection(): Promise<NatsConnection> {
  if (nc) return nc
  if (connecting) {
    // Wait for in-flight connection
    return new Promise((resolve) => {
      const check = setInterval(() => {
        if (nc) { clearInterval(check); resolve(nc) }
      }, 50)
    })
  }

  connecting = true
  // Connect directly to embedded NATS WebSocket port.
  // NATS WS runs on :4313 — separate from Go HTTP (:4312) and Vite (:5173).
  const natsPort = import.meta.env.VITE_NATS_WS_PORT || '4313'
  const wsUrl = `ws://${window.location.hostname}:${natsPort}`
  nc = await wsconnect({ servers: wsUrl })
  connecting = false
  console.log('[nats] connected via', wsUrl)
  return nc
}

/** NATS request-reply: send request, get JSON response. */
export async function natsRequest<T>(subject: string, data?: unknown): Promise<T> {
  const conn = await getNatsConnection()
  const payload = data ? new TextEncoder().encode(JSON.stringify(data)) : undefined
  const reply = await conn.request(subject, payload, { timeout: 5000 })
  return JSON.parse(new TextDecoder().decode(reply.data)) as T
}

/** Subscribe to NATS subject, call handler on each message. Returns unsubscribe fn. */
export async function natsSubscribe(
  subject: string,
  handler: (data: unknown, subject: string) => void,
): Promise<() => void> {
  const conn = await getNatsConnection()
  const sub = conn.subscribe(subject)

  // Process messages in background
  void (async () => {
    for await (const msg of sub) {
      const data = msg.data.length > 0
        ? JSON.parse(new TextDecoder().decode(msg.data))
        : null
      handler(data, msg.subject)
    }
  })()

  return () => {
    sub.unsubscribe()
  }
}
