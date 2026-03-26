import { useEffect } from 'react'
import { useQueryClient, useQuery } from '@tanstack/react-query'
import { natsRequest, natsSubscribe } from './nats'
import type { Section } from 'virtual:remmd/schema'

interface SectionsResponse {
  doc_id: string
  sections: Section[]
  version: number
}

export function useNatsInvalidation() {
  const queryClient = useQueryClient()

  useEffect(() => {
    let unsub: (() => void) | null = null

    natsSubscribe('remmd.doc.>', (data, subject) => {
      const parts = subject.split('.')
      const docId = parts[2]
      console.log('[nats] change event:', subject, data)

      queryClient.invalidateQueries({ queryKey: ['sections', docId] })
      queryClient.invalidateQueries({ queryKey: ['graph'] })
      queryClient.invalidateQueries({ queryKey: ['validation'] })
    }).then(fn => { unsub = fn })

    return () => { unsub?.() }
  }, [queryClient])
}

export function useSections(docId: string) {
  return useQuery({
    queryKey: ['sections', docId],
    queryFn: () => natsRequest<SectionsResponse>(`remmd.q.documents.${docId}.sections`),
    staleTime: Infinity, // Only refetch when NATS invalidates
  })
}

export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: () => natsRequest<{ id: string; title: string; status: string }[]>('remmd.q.documents'),
    staleTime: Infinity,
  })
}
