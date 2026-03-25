import { useEffect } from 'react'
import { useQueryClient, useQuery } from '@tanstack/react-query'
import { natsRequest, natsSubscribe } from './nats'
import type { Section } from 'virtual:remmd/schema'

interface SectionsResponse {
  doc_id: string
  sections: Section[]
  version: number
}

/**
 * Hook: subscribe to NATS change events and invalidate React Query caches.
 * All data flows through NATS — this is the realtime bridge.
 */
export function useNatsInvalidation() {
  const queryClient = useQueryClient()

  useEffect(() => {
    let unsub: (() => void) | null = null

    natsSubscribe('remmd.doc.>', (data, subject) => {
      // Subject: remmd.doc.<docId>.section.<ref>
      const parts = subject.split('.')
      const docId = parts[2]
      console.log('[nats] change event:', subject, data)

      // Invalidate the sections query for this document
      queryClient.invalidateQueries({ queryKey: ['sections', docId] })
    }).then(fn => { unsub = fn })

    return () => { unsub?.() }
  }, [queryClient])
}

/**
 * Hook: fetch sections via NATS request-reply.
 * React Query manages caching, stale-while-revalidate, and deduplication.
 */
export function useSections(docId: string) {
  return useQuery({
    queryKey: ['sections', docId],
    queryFn: () => natsRequest<SectionsResponse>(`remmd.q.documents.${docId}.sections`),
    staleTime: Infinity, // Only refetch when NATS invalidates
  })
}

/**
 * Hook: fetch document list via NATS request-reply.
 */
export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: () => natsRequest<{ id: string; title: string; status: string }[]>('remmd.q.documents'),
    staleTime: Infinity,
  })
}
