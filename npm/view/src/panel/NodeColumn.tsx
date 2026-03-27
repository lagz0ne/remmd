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
    const parts: string[] = []
    let inTable = false
    for (const s of sections) {
      const sTitle = (s as any).title || ''
      const sContent = (s as any).content || ''
      if (sTitle === title) continue

      const isTableRow = sContent.trimStart().startsWith('|') || sTitle.startsWith('|')
      if (isTableRow) {
        const row = sContent.trim()
        if (row) {
          if (inTable) {
            parts[parts.length - 1] += '\n' + row
          } else {
            // Start new table: infer column count and add header
            const cols = (row.match(/\|/g) || []).length - 1
            const header = '| ' + Array.from({length: cols}, (_, i) => `Col ${i+1}`).join(' | ') + ' |'
            const sep = '| ' + Array.from({length: cols}, () => '---').join(' | ') + ' |'
            parts.push(header + '\n' + sep + '\n' + row)
            inTable = true
          }
        }
        continue
      }

      inTable = false
      if (sTitle) parts.push(`## ${sTitle}`)
      const body = sContent.startsWith(sTitle) ? sContent.slice(sTitle.length).replace(/^\n+/, '') : sContent
      if (body) parts.push(body)
    }
    return marked.parse(unescape(parts.join('\n\n'))) as string
  }, [sections, title])

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
        className="panel-md"
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
