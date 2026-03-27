import { useState, useRef, useCallback } from 'react'
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query'
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
import { computeAutoLayout, LAYOUT_NODE_WIDTH, LAYOUT_NODE_HEIGHT } from './canvas/use-auto-layout'
import { DocNode } from './canvas/DocNode'
import { BundledEdge } from './canvas/BundledEdge'
import { NodePanel, type StackEntry } from './panel/NodePanel'
import type { EdgeItem } from './panel/EdgeFooter'
import { natsRequest } from './nats'

interface NodePosition {
  node_id: string
  x: number
  y: number
  width: number
  height: number
}

const queryClient = new QueryClient()
const nodeTypes = { document: DocNode }
const edgeTypes = { bundled: BundledEdge }

function Canvas() {
  useNatsInvalidation()
  const { nodes: graphNodes, edges: graphEdges, isLoading } = useGraphData()
  const { data: savedPositions } = useQuery({
    queryKey: ['positions'],
    queryFn: () => natsRequest<NodePosition[]>('remmd.q.positions'),
    staleTime: Infinity,
  })
  const [nodes, setNodes, onNodesChange] = useNodesState([] as Node[])
  const [edges, setEdges, onEdgesChange] = useEdgesState([] as Edge[])
  const [initialized, setInitialized] = useState(false)
  const [panelStack, setPanelStack] = useState<StackEntry[]>([])
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const savePositions = useCallback((currentNodes: Node[]) => {
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current)
    saveTimerRef.current = setTimeout(() => {
      const positions: NodePosition[] = currentNodes.map(n => ({
        node_id: n.id,
        x: n.position.x,
        y: n.position.y,
        width: (n.measured?.width ?? n.width) || LAYOUT_NODE_WIDTH,
        height: (n.measured?.height ?? n.height) || LAYOUT_NODE_HEIGHT,
      }))
      natsRequest('remmd.c.positions', positions).catch(err =>
        console.warn('[positions] save failed:', err),
      )
    }, 1000)
  }, [])

  const onNodeDragStop = useCallback(
    (_: any, _node: Node, draggedNodes: Node[]) => {
      // React Flow gives us the dragged nodes with updated positions,
      // but we save ALL node positions from current state
      void draggedNodes // used by React Flow callback signature
      setNodes(current => {
        savePositions(current)
        return current
      })
    },
    [savePositions, setNodes],
  )

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

  // Wait for both graph data and saved positions before initializing.
  // savedPositions can be undefined (loading) or an array (loaded, possibly empty).
  if (!initialized && graphNodes.length > 0 && savedPositions !== undefined) {
    const posArr = Array.isArray(savedPositions) ? savedPositions : []
    const posMap = new Map(posArr.map(p => [p.node_id, p]))

    const hasPositioned = graphNodes.some(n => posMap.has(n.id))
    if (hasPositioned) {
      // Nodes with saved positions: apply directly
      // Nodes without saved positions: auto-layout only those, then offset
      const withPos: Node[] = []
      const withoutPos: Node[] = []
      for (const n of graphNodes) {
        if (posMap.has(n.id)) {
          const p = posMap.get(n.id)!
          withPos.push({ ...n, position: { x: p.x, y: p.y } })
        } else {
          withoutPos.push(n)
        }
      }

      if (withoutPos.length > 0) {
        // Layout only new nodes, then shift them to avoid overlap
        const autoLaid = computeAutoLayout(withoutPos, graphEdges)
        // Find bounding box of positioned nodes to place new ones nearby
        let maxX = -Infinity
        for (const n of withPos) {
          maxX = Math.max(maxX, n.position.x + LAYOUT_NODE_WIDTH)
        }
        const offsetX = maxX + 100
        for (const n of autoLaid) {
          n.position.x += offsetX
        }
        setNodes([...withPos, ...autoLaid])
      } else {
        setNodes(withPos)
      }
    } else {
      // No saved positions at all — full auto-layout
      setNodes(computeAutoLayout(graphNodes, graphEdges))
    }
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
        edgeTypes={edgeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onNodeDragStop={onNodeDragStop}
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
