import { useQuery } from '@tanstack/react-query'
import { natsRequest } from '../nats'

interface PlaybookField {
  name: string
  type: string
  required: boolean
  target?: string
  targets?: string[]
  values?: string[]
}

interface PlaybookSection {
  name: string
  required: boolean
}

interface PlaybookRule {
  name: string
  description?: string
  severity: string
  expr: string
}

interface PlaybookType {
  name: string
  description?: string
  fields: PlaybookField[]
  sections: PlaybookSection[]
  rules: PlaybookRule[]
}

interface PlaybookEdge {
  name: string
  notation: string
}

export interface PlaybookResponse {
  types: PlaybookType[]
  edges: PlaybookEdge[]
  rules: PlaybookRule[]
}

interface ValidationDiag {
  rule: string
  node_id: string
  node_type: string
  severity: string
  message: string
}

interface ValidationResponse {
  diagnostics: ValidationDiag[]
}

export function usePlaybook() {
  return useQuery({
    queryKey: ['playbook'],
    queryFn: () => natsRequest<PlaybookResponse>('remmd.q.playbook'),
    staleTime: Infinity,
    retry: false,
  })
}

export function useValidation() {
  return useQuery({
    queryKey: ['validation'],
    queryFn: () => natsRequest<ValidationResponse>('remmd.q.validate'),
    staleTime: 30_000,
    retry: false,
  })
}

export type { PlaybookType, ValidationDiag, ValidationResponse }
