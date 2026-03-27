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
import { DocNode } from './canvas/DocNode'
import { NodePanel, type StackEntry } from './panel/NodePanel'
import type { EdgeItem } from './panel/EdgeFooter'

const queryClient = new QueryClient()
const nodeTypes = { document: DocNode }

function Canvas() {
  useNatsInvalidation()
  const { nodes: graphNodes, edges: graphEdges, isLoading } = useGraphData()
  const [nodes, setNodes, onNodesChange] = useNodesState([] as Node[])
  const [edges, setEdges, onEdgesChange] = useEdgesState([] as Edge[])
  const [initialized, setInitialized] = useState(false)
  const [panelStack, setPanelStack] = useState<StackEntry[]>([])

  const onNodeClick = (_: any, node: Node) => {
    setPanelStack([{
      nodeId: node.id,
      title: (node.data as any).title || node.id,
      typeName: (node.data as any).playbookType || '',
    }])
  }

  const onPaneClick = () => setPanelStack([])

  const onEdgeFollow = (edge: EdgeItem) => {
    setPanelStack(prev => [...prev, {
      nodeId: edge.targetId,
      title: edge.targetTitle,
      typeName: '',
      edgeType: edge.edgeType,
      sourceId: prev[prev.length - 1]?.nodeId,
    }])
  }

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
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        proOptions={{ hideAttribution: true }}
        fitView
        fitViewOptions={{ padding: 0.3 }}
      >
        <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#e4e4e7" />
        <Controls position="bottom-left" />
        <MiniMap
          position="bottom-right"
          style={{
            width: 140,
            height: 90,
            border: '1px solid #e4e4e7',
            borderRadius: 6,
            background: '#fafafa',
          }}
          maskColor="rgba(0, 0, 0, 0.05)"
          nodeColor="#d4d4d8"
        />
      </ReactFlow>
      <NodePanel
        stack={panelStack}
        graphEdges={edges}
        graphNodes={nodes}
        onClose={() => setPanelStack([])}
        onEdgeFollow={onEdgeFollow}
        onStackChange={setPanelStack}
      />
      <div className="fixed bottom-1 left-1 text-[8px] text-zinc-300 font-mono pointer-events-none select-none z-50">
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
