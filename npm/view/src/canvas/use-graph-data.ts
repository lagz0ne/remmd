import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { Node, Edge } from '@xyflow/react'
import { natsRequest } from '../nats'
import { usePlaybook, useValidation } from '../hooks/use-playbook'
import type { PlaybookResponse, ValidationResponse } from '../hooks/use-playbook'

interface GraphResponse {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

interface GraphNode {
  id: string
  title: string
  status: string
  source: string
  doc_type: string
  brief: string
  section_count: number
}

interface GraphEdge {
  id: string
  source_doc_id: string
  target_doc_id: string
  state: string
  relationship_type: string
  left_section_ids: string[]
  right_section_ids: string[]
  is_relation?: boolean
}

export interface DocNodeData extends Record<string, unknown> {
  title: string
  status: string
  source: string
  worstState: string
  sectionCount: number
  edgeCounts: Record<string, number>
  playbookType: string
  owner: string
  brief: string
  validationPassing: number
  validationTotal: number
  validationErrors: { rule: string; message: string }[]
  outgoing: number
  incoming: number
  threadCount: number
  portWorstState: string
}

export interface BundleEdgeData extends Record<string, unknown> {
  links: GraphEdge[]
  worstState: string
  count: number
  edgeType: string
  threadCount: number
  isStructural: boolean
}

export function useGraphData() {
  const { data: raw, ...graphRest } = useQuery({
    queryKey: ['graph'],
    queryFn: () => natsRequest<GraphResponse>('remmd.q.graph'),
    staleTime: Infinity,
  })

  const { data: pb } = usePlaybook()
  const { data: validation } = useValidation()

  const { nodes, edges } = useMemo(() => {
    if (!raw) return { nodes: [], edges: [] }
    return transformGraph(raw, pb, validation)
  }, [raw, pb, validation])

  return { nodes, edges, ...graphRest }
}

export function transformGraph(
  raw: GraphResponse,
  pb?: PlaybookResponse | null,
  validation?: ValidationResponse | null,
): {
  nodes: Node<DocNodeData>[]
  edges: Edge<BundleEdgeData>[]
} {
  const statePriority: Record<string, number> = {
    aligned: 1,
    pending: 2,
    stale: 3,
    broken: 4,
  }

  const docEdgeCounts = new Map<string, Record<string, number>>()
  const docOutgoing = new Map<string, number>()
  const docIncoming = new Map<string, number>()

  for (const e of raw.edges) {
    for (const docId of [e.source_doc_id, e.target_doc_id]) {
      if (!docEdgeCounts.has(docId)) docEdgeCounts.set(docId, {})
      const counts = docEdgeCounts.get(docId)!
      counts[e.state] = (counts[e.state] || 0) + 1
    }
    docOutgoing.set(e.source_doc_id, (docOutgoing.get(e.source_doc_id) || 0) + 1)
    docIncoming.set(e.target_doc_id, (docIncoming.get(e.target_doc_id) || 0) + 1)
  }

  const nodes: Node<DocNodeData>[] = raw.nodes.map((n) => {
    const edgeCounts = docEdgeCounts.get(n.id) || {}
    const worstState = Object.keys(edgeCounts).reduce(
      (worst, state) =>
        (statePriority[state] || 0) > (statePriority[worst] || 0) ? state : worst,
      '',
    )

    const playbookType = n.doc_type || ''

    const nodeErrors = (validation?.diagnostics ?? [])
      .filter((d) => d.node_id === n.id)
      .map((d) => ({ rule: d.rule, message: d.message }))
    const nodeType = n.doc_type || ''
    const typeRules = pb?.types?.find(t => t.name === nodeType)?.rules?.length ?? 0
    const globalRules = pb?.rules?.length ?? 0
    const validationTotal = typeRules + globalRules
    const validationPassing = Math.max(0, validationTotal - nodeErrors.length)

    return {
      id: n.id,
      type: 'document',
      position: { x: (Math.random() - 0.5) * 800, y: (Math.random() - 0.5) * 600 },
      data: {
        title: n.title,
        status: n.status,
        source: n.source,
        worstState,
        sectionCount: n.section_count || 0,
        edgeCounts,
        playbookType,
        owner: '',
        brief: n.brief || '',
        validationPassing,
        validationTotal,
        validationErrors: nodeErrors,
        outgoing: docOutgoing.get(n.id) || 0,
        incoming: docIncoming.get(n.id) || 0,
        threadCount: 0,
        portWorstState: worstState,
      },
    }
  })

  const bundles = new Map<string, GraphEdge[]>()
  for (const e of raw.edges) {
    const key = `${e.source_doc_id}::${e.target_doc_id}`
    if (!bundles.has(key)) bundles.set(key, [])
    bundles.get(key)!.push(e)
  }

  const edges: Edge<BundleEdgeData>[] = Array.from(bundles.entries()).map(
    ([key, links]) => {
      const worstState = links.reduce(
        (worst, link) =>
          (statePriority[link.state] || 0) > (statePriority[worst] || 0)
            ? link.state
            : worst,
        '',
      )

      const first = links[0]
      const edgeType = first.relationship_type || ''
      const isStructural = first.is_relation || ['contains', 'parent-of'].includes(edgeType)

      return {
        id: `bundle-${key}`,
        source: first.source_doc_id,
        target: first.target_doc_id,
        type: 'bundled',
        data: {
          links,
          worstState,
          count: links.length,
          edgeType,
          threadCount: 0,
          isStructural,
        },
      }
    },
  )

  return { nodes, edges }
}
