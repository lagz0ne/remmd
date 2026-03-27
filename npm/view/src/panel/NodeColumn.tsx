import { useSections } from '../hooks'
import { marked } from 'marked'
import { useMemo } from 'react'

marked.setOptions({ breaks: true, gfm: true })

function unescape(s: string): string {
  return s.replace(/\\n/g, '\n')
}

interface NodeColumnProps {
  docId: string
  title: string
  typeName: string
  onClose?: () => void
  header?: React.ReactNode
  footer?: React.ReactNode
}

export function NodeColumn({ docId, title, typeName, onClose, header, footer }: NodeColumnProps) {
  const { data } = useSections(docId)
  const sections = data?.sections ?? []

  const html = useMemo(() => {
    if (sections.length === 0) return ''
    const md = sections.map(s => s.content || '').join('\n\n')
    return marked.parse(unescape(md)) as string
  }, [sections])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', minWidth: 0 }}>
      {header}
      <div style={{
        padding: '8px 14px',
        borderBottom: '1px solid #f4f4f5',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
      }}>
        <div>
          {typeName && <div style={{ fontSize: 9, color: '#a1a1aa', letterSpacing: '0.03em' }}>{typeName}</div>}
          <div style={{ fontSize: 14, fontWeight: 600, color: '#18181b', marginTop: 1 }}>{title}</div>
        </div>
        {onClose && (
          <span onClick={onClose} style={{ fontSize: 11, color: '#a1a1aa', cursor: 'pointer', padding: '2px 4px' }}>✕</span>
        )}
      </div>
      <div
        className="doc-node-md"
        style={{
          flex: 1,
          padding: '10px 14px',
          overflowY: 'auto',
          fontSize: 11,
          color: '#52525b',
          lineHeight: 1.6,
        }}
        dangerouslySetInnerHTML={{ __html: html }}
      />
      {footer}
    </div>
  )
}
