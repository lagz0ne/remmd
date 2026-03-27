import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'

interface DocNodeData extends Record<string, unknown> {
  title: string
  brief: string
  playbookType: string
  status: string
}

function DocNodeInner({ data }: NodeProps & { data: DocNodeData }) {
  const typeName = (data.playbookType as string) || ''
  const title = (data.title as string) || 'Untitled'
  const brief = (data.brief as string) || ''

  return (
    <div
      style={{
        background: '#fff',
        border: '1px solid #d4d4d8',
        borderRadius: 6,
        padding: '8px 12px',
        width: 200,
        fontFamily: 'system-ui, -apple-system, sans-serif',
        cursor: 'grab',
      }}
    >
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />

      {typeName && (
        <div style={{
          fontSize: 9,
          color: '#a1a1aa',
          letterSpacing: '0.03em',
          marginBottom: 2,
        }}>
          {typeName}
        </div>
      )}

      <div style={{
        fontSize: 13,
        fontWeight: 600,
        color: '#18181b',
        lineHeight: 1.3,
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
      }}>
        {title}
      </div>

      {brief && (
        <div style={{
          fontSize: 10,
          color: '#71717a',
          lineHeight: 1.4,
          marginTop: 4,
          overflow: 'hidden',
          display: '-webkit-box',
          WebkitLineClamp: 2,
          WebkitBoxOrient: 'vertical' as const,
        }}>
          {brief}
        </div>
      )}

      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}

export const DocNode = memo(DocNodeInner)
