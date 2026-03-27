import { useCallback } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  ReactFlow,
  ReactFlowProvider,
  MiniMap,
  Controls,
  Background,
  BackgroundVariant,
  type NodeMouseHandler,
  type EdgeMouseHandler,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { ArrowClockwiseIcon } from '@phosphor-icons/react'

import { useNatsInvalidation } from './hooks'
import { useGraphData } from './canvas/use-graph-data'
import { useForceLayout } from './canvas/use-layout'
import { PlaybookNode } from './canvas/PlaybookNode'
import { GhostNode } from './canvas/GhostNode'
import { BundledEdge } from './canvas/BundledEdge'
import { Legend } from './canvas/Legend'
import { EmptyState } from './canvas/EmptyState'
import { PanelShell } from './panel/PanelShell'
import { GapPanel } from './panel/GapPanel'
import { ThreadPanel } from './panel/ThreadPanel'
import { usePanelState } from './panel/use-panel-state'
import { usePlaybook } from './hooks/use-playbook'

const queryClient = new QueryClient()

const nodeTypes = { document: PlaybookNode, ghost: GhostNode }
const edgeTypes = { bundled: BundledEdge }

function Canvas() {
  useNatsInvalidation()
  const { nodes, edges, isLoading } = useGraphData()
  const { onNodeDragStart, onNodeDrag, onNodeDragStop, resetLayout } = useForceLayout(nodes, edges)
  const panel = usePanelState()
  const { data: pb } = usePlaybook()
  const playbookTypes = pb?.types ? pb.types.map(t => t.name) : []

  const onNodeClick: NodeMouseHandler = useCallback(
    (_, node) => {
      panel.selectNode(node.id)
    },
    [panel.selectNode],
  )

  const onEdgeClick: EdgeMouseHandler = useCallback(
    (_, edge) => {
      const sourceDocId = edge.source
      panel.selectEdge(edge.id, sourceDocId)
    },
    [panel.selectEdge],
  )

  const onPaneClick = useCallback(() => {
    panel.close()
  }, [panel.close])

  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center text-sm text-zinc-400">
        Loading graph...
      </div>
    )
  }

  if (nodes.length === 0) {
    return (
      <div className="h-screen w-screen relative">
        <EmptyState />
      </div>
    )
  }

  const selectedNode = nodes.find((n) => n.id === panel.selectedNodeId)
  const selectedEdge = edges.find((e) => e.id === panel.selectedEdgeId)
  const connectedDocId =
    selectedEdge && panel.selectedNodeId
      ? selectedEdge.source === panel.selectedNodeId
        ? selectedEdge.target
        : selectedEdge.source
      : null
  const connectedNode = connectedDocId ? nodes.find((n) => n.id === connectedDocId) : null

  return (
    <div className="h-screen w-screen">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        onNodeClick={onNodeClick}
        onEdgeClick={onEdgeClick}
        onPaneClick={onPaneClick}
        onNodeDragStart={onNodeDragStart}
        onNodeDrag={onNodeDrag}
        onNodeDragStop={onNodeDragStop}
        panOnScroll
        panOnDrag
        proOptions={{ hideAttribution: true }}
        minZoom={0.1}
        maxZoom={2}
        fitView
        fitViewOptions={{ padding: 0.3 }}
      >
        <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#e4e4e7" />
        <MiniMap
          position="bottom-left"
          pannable
          zoomable
          className="!bg-white/80 !border-zinc-200"
        />
        <Controls
          position="bottom-left"
          showInteractive={false}
          className="!border-zinc-200 !bg-white/90 !shadow-sm"
          style={{ marginBottom: 140 }}
        />
        <div className="absolute bottom-2 left-14 z-10" style={{ marginBottom: 140 }}>
          <button
            onClick={resetLayout}
            className="flex items-center gap-1 px-2 py-1.5 text-xs text-zinc-500 bg-white/90 border border-zinc-200 rounded shadow-sm hover:bg-zinc-50 hover:text-zinc-700 transition-colors"
            title="Re-layout nodes"
          >
            <ArrowClockwiseIcon size={14} />
            Re-layout
          </button>
        </div>
        <Legend types={playbookTypes} />

        <PanelShell mode={panel.mode} columns={panel.columns} onClose={panel.close}>
          {selectedNode && (panel.mode === 'doc' || panel.mode === 'edge') && (
            <GapPanel
              key={selectedNode.id}
              docId={selectedNode.id}
              docTitle={selectedNode.data.title}
              playbookType={selectedNode.data.playbookType || ''}
              owner={selectedNode.data.owner || ''}
              validationErrors={selectedNode.data.validationErrors || []}
              validationPassing={selectedNode.data.validationPassing || 0}
              validationTotal={selectedNode.data.validationTotal || 0}
            />
          )}

          {panel.mode === 'edge' && selectedEdge?.data && (
            <>
              <div className="w-px bg-zinc-200 shrink-0" />
              <ThreadPanel
                edgeType={selectedEdge.data.edgeType || ''}
                state={selectedEdge.data.worstState || ''}
                sourceTitle={selectedNode?.data.title || ''}
                targetTitle={connectedNode?.data.title || ''}
                links={selectedEdge.data.links.map(l => ({
                  id: l.id,
                  state: l.state,
                  relationship_type: l.relationship_type,
                }))}
              />
            </>
          )}
        </PanelShell>
      </ReactFlow>
      <div className="fixed bottom-1 right-1 text-[8px] text-zinc-300 font-mono pointer-events-none select-none z-50">
        {__BUILD_VERSION__}
      </div>
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ReactFlowProvider>
        <Canvas />
      </ReactFlowProvider>
    </QueryClientProvider>
  )
}
