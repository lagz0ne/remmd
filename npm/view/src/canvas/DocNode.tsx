import { memo, useMemo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { marked } from 'marked'

interface DocNodeData extends Record<string, unknown> {
  title: string
  brief: string
  playbookType: string
  status: string
}

marked.setOptions({ breaks: true, gfm: true })

function renderBrief(raw: string): string {
  const unescaped = raw.replace(/\\n/g, '\n')
  return marked.parse(unescaped) as string
}

function DocNodeInner({ data }: NodeProps & { data: DocNodeData }) {
  const typeName = (data.playbookType as string) || ''
  const title = (data.title as string) || 'Untitled'
  const brief = (data.brief as string) || ''
  const html = useMemo(() => brief ? renderBrief(brief) : '', [brief])

  return (
    <div
      style={{
        background: '#fff',
        border: '1px solid #d4d4d8',
        borderRadius: 6,
        padding: '8px 12px',
        width: 220,
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

      {html && (
        <div style={{ position: 'relative', marginTop: 6, maxHeight: 80, overflow: 'hidden' }}>
          <div
            className="doc-node-md"
            dangerouslySetInnerHTML={{ __html: html }}
          />
          <div style={{
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: 20,
            background: 'linear-gradient(transparent, white)',
          }} />
        </div>
      )}

      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}

export const DocNode = memo(DocNodeInner)
