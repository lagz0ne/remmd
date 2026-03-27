import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  ReactFlow,
  ReactFlowProvider,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { useNatsInvalidation } from './hooks'
import { useGraphData } from './canvas/use-graph-data'
import { computeAutoLayout } from './canvas/use-auto-layout'

const queryClient = new QueryClient()

function Canvas() {
  useNatsInvalidation()
  const { nodes: graphNodes, edges: graphEdges, isLoading } = useGraphData()
  const [nodes, setNodes, onNodesChange] = useNodesState([] as Node[])
  const [edges, setEdges, onEdgesChange] = useEdgesState([] as Edge[])
  const [initialized, setInitialized] = useState(false)

  if (!initialized && graphNodes.length > 0) {
    setNodes(computeAutoLayout(graphNodes, graphEdges))
    setEdges(graphEdges)
    setInitialized(true)
  }

  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center text-sm text-zinc-400">
        Loading...
      </div>
    )
  }

  if (graphNodes.length === 0 && !isLoading) {
    return (
      <div className="h-screen flex items-center justify-center text-sm text-zinc-400">
        No documents. Run: remmd doc create "Title"
      </div>
    )
  }

  return (
    <div className="h-screen w-screen">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        proOptions={{ hideAttribution: true }}
        fitView
        fitViewOptions={{ padding: 0.3 }}
      >
        <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#e4e4e7" />
        <Controls position="bottom-left" />
        <MiniMap position="bottom-left" style={{ marginBottom: 120 }} />
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
