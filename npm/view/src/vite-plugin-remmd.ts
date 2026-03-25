import type { Plugin, ViteDevServer } from 'vite'
import type { NatsConnection, Subscription } from '@nats-io/nats-core'

interface SectionData {
  id: string
  ref: string
  title: string
  content: string
  content_hash: string
  content_type: 'native' | 'external'
  source_url?: string
  link_state?: string
}

/**
 * Vite plugin for remmd — virtual modules + NATS-driven HMR.
 *
 * Two virtual module namespaces:
 *   1. 'virtual:remmd/schema' — static domain constants
 *   2. 'virtual:section/<ref>' — dynamic section content from DB via NATS
 *
 * In dev mode, subscribes to NATS change events and triggers HMR with
 * diff data so the browser can update sections in-place.
 */
export function remmdPlugin(): Plugin {
  const SCHEMA_ID = 'virtual:remmd/schema'
  const SECTION_PREFIX = 'virtual:section/'

  const MAX_CACHE_SIZE = 500
  const sectionCache = new Map<string, SectionData>()

  // NATS connection state (server-side, in Vite's Node process)
  let nc: NatsConnection | null = null
  let sub: Subscription | null = null

  const schemaSource = `
export const SECTION_TYPES = ["heading", "list-item", "checklist", "table-row", "code-block"];
export const CONTENT_TYPES = ["native", "external"];
export const LINK_STATES = ["pending", "aligned", "stale", "broken", "archived"];
export const RELATIONSHIP_TYPES = ["agrees_with", "implements", "tests", "evidences"];
export const INTERVENTION_LEVELS = ["watch", "notify", "urgent", "blocking"];
`

  /** Server-side NATS request-reply. Runs in Node.js, not the browser. */
  async function natsRequest(subject: string): Promise<SectionData> {
    if (!nc) throw new Error('[remmd] NATS not connected')
    const reply = await nc.request(subject, undefined, { timeout: 5000 })
    return JSON.parse(new TextDecoder().decode(reply.data)) as SectionData
  }

  /**
   * Connect to embedded NATS WS and subscribe to document change events.
   * On each change: invalidate the virtual module + send HMR event with diff.
   */
  async function setupNatsSubscription(vite: ViteDevServer, port: string) {
    try {
      // Dynamic import — @nats-io/nats-core is ESM
      const { wsconnect } = await import('@nats-io/nats-core')
      const wsUrl = `ws://127.0.0.1:${port}`
      nc = await wsconnect({ servers: wsUrl, name: 'vite-plugin-remmd' })
      console.log('[remmd] NATS connected via', wsUrl)

      // Subscribe to all document change events
      sub = nc.subscribe('remmd.doc.>')

      // Process change events in background
      void (async () => {
        for await (const msg of sub!) {
          // Subject format: remmd.doc.<docId>.section.<ref>
          const parts = msg.subject.split('.')
          if (parts.length < 5 || parts[3] !== 'section') continue
          const docId = parts[2]
          const ref = parts[4]

          const oldData = sectionCache.get(ref)

          // Fetch fresh section content
          let newData: SectionData
          try {
            newData = await natsRequest(`remmd.q.section.${ref}`)
          } catch (err) {
            console.error(`[remmd] Failed to fetch updated section ${ref}:`, err)
            continue
          }

          sectionCache.set(ref, newData)
          if (sectionCache.size > MAX_CACHE_SIZE) {
            const oldest = sectionCache.keys().next().value!
            sectionCache.delete(oldest)
          }

          // Invalidate the virtual module so next import gets fresh content
          const moduleId = '\0' + SECTION_PREFIX + ref
          const mod = vite.moduleGraph.getModuleById(moduleId)
          if (mod) {
            vite.moduleGraph.invalidateModule(mod)
          }

          // Send HMR event to browser with diff data
          vite.hot.send({
            type: 'custom',
            event: 'remmd:section-update',
            data: {
              ref,
              docId,
              oldContent: oldData?.content ?? '',
              newContent: newData.content,
              oldHash: oldData?.content_hash ?? '',
              newHash: newData.content_hash,
              timestamp: Date.now(),
            },
          })

          console.log(`[remmd] HMR: section ${ref} updated (${oldData?.content_hash?.slice(0, 8) ?? 'new'} -> ${newData.content_hash.slice(0, 8)})`)
        }
      })()
    } catch (err) {
      // NATS not available — degrade gracefully, sections load on demand
      console.warn('[remmd] NATS subscription failed (sections will load on-demand):', err)
    }
  }

  return {
    name: 'vite-plugin-remmd',

    resolveId(id: string) {
      if (id === SCHEMA_ID) return '\0' + SCHEMA_ID
      if (id.startsWith(SECTION_PREFIX)) return '\0' + id
    },

    async load(id: string) {
      // Static schema constants
      if (id === '\0' + SCHEMA_ID) {
        return schemaSource
      }

      // Dynamic section content via NATS request-reply
      if (id.startsWith('\0' + SECTION_PREFIX)) {
        const ref = id.slice(('\0' + SECTION_PREFIX).length)

        try {
          const data = await natsRequest(`remmd.q.section.${ref}`)
          sectionCache.set(ref, data)
          return `export default ${JSON.stringify(data)}`
        } catch (err) {
          console.error(`[remmd] Failed to load section ${ref}:`, err)
          // Return error module so the import doesn't break the build
          return `export default ${JSON.stringify({ error: `Failed to load section ${ref}`, ref })}`
        }
      }
    },

    configureServer(server: ViteDevServer) {
      const natsPort = process.env.VITE_NATS_WS_PORT || '4313'
      setupNatsSubscription(server, natsPort)

      // Cleanup on server close
      return () => {
        if (sub) {
          sub.unsubscribe()
          sub = null
        }
        if (nc) {
          nc.drain().catch(() => {})
          nc = null
        }
      }
    },
  }
}
