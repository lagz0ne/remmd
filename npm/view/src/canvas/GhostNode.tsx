import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { PlusIcon } from '@phosphor-icons/react'

export interface GhostNodeData extends Record<string, unknown> {
  typeName: string
  requiredBy: string
}

function GhostNodeInner({ data }: NodeProps & { data: GhostNodeData }) {
  return (
    <div
      className="rounded-xl px-5 py-4 cursor-pointer transition-shadow hover:shadow-[0_4px_12px_rgba(99,102,241,0.08)]"
      style={{
        border: '1.5px dashed #c7d2fe',
        background: '#eef2ff',
        minWidth: 170,
      }}
    >
      <Handle type="target" position={Position.Left} className="opacity-0" />
      <div className="text-[9px] text-indigo-400 font-normal">{data.typeName}</div>
      <div className="flex items-center gap-1 mt-1.5">
        <PlusIcon size={12} weight="light" className="text-indigo-500" />
        <span className="text-[13px] font-medium text-indigo-500">add {data.typeName}</span>
      </div>
      <div className="text-[9px] text-indigo-300 mt-1">required by {data.requiredBy}</div>
      <Handle type="source" position={Position.Right} className="opacity-0" />
    </div>
  )
}

export const GhostNode = memo(GhostNodeInner)
