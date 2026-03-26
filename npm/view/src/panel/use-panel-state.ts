import { useState, useCallback } from 'react'

export type PanelMode = 'closed' | 'doc' | 'edge' | 'thread'

interface PanelState {
  mode: PanelMode
  selectedNodeId: string | null
  selectedEdgeId: string | null
  /** Number of visible columns: 1 (doc only) or 2 (doc + edge/connected) */
  columns: 1 | 2
}

export function usePanelState() {
  const [state, setState] = useState<PanelState>({
    mode: 'closed',
    selectedNodeId: null,
    selectedEdgeId: null,
    columns: 1,
  })

  const selectNode = useCallback((nodeId: string) => {
    setState({
      mode: 'doc',
      selectedNodeId: nodeId,
      selectedEdgeId: null,
      columns: 1,
    })
  }, [])

  const selectEdge = useCallback((edgeId: string, sourceDocId: string) => {
    setState((prev) => ({
      mode: 'edge',
      selectedNodeId: prev.selectedNodeId || sourceDocId,
      selectedEdgeId: edgeId,
      columns: 2,
    }))
  }, [])

  const close = useCallback(() => {
    setState({
      mode: 'closed',
      selectedNodeId: null,
      selectedEdgeId: null,
      columns: 1,
    })
  }, [])

  return { ...state, selectNode, selectEdge, close }
}
