import { useState, useEffect } from 'react'
import { DiffHighlight } from './DiffHighlight'

interface SectionProps {
  ref_: string  // @a1, @ext:jira/ENG-1234, etc.
  data: {
    id: string; ref: string; title: string; content: string;
    content_hash: string; content_type: 'native' | 'external';
    source_url?: string; link_state?: string;
  }
}

export function Section({ ref_, data }: SectionProps) {
  const [diff, setDiff] = useState<{ oldContent: string; newContent: string } | null>(null)
  const stateClass = `state-${data.link_state || 'none'}`

  // Listen for HMR section updates
  useEffect(() => {
    if (import.meta.hot) {
      const handler = (update: any) => {
        if (update.ref === ref_) {
          setDiff({ oldContent: update.oldContent, newContent: update.newContent })
          // Clear diff highlight after 3 seconds
          setTimeout(() => setDiff(null), 3000)
        }
      }
      import.meta.hot.on('remmd:section-update', handler)
      return () => { import.meta.hot?.off?.('remmd:section-update', handler) }
    }
  }, [ref_])

  const copyRef = async () => {
    await navigator.clipboard.writeText(ref_)
  }

  return (
    <div className={`section-card ${stateClass} ${diff ? 'has-diff' : ''}`} onClick={copyRef} title={`Click to copy ${ref_}`}>
      <div className="section-header">
        <span className="ref">{ref_}</span>
        <span className="title">{data.title}</span>
        {data.link_state && <span className={`badge ${stateClass}`}>{data.link_state}</span>}
        {data.content_type === 'external' && data.source_url && (
          <a className="source-link" href={data.source_url} target="_blank" rel="noopener" onClick={e => e.stopPropagation()}>source</a>
        )}
      </div>
      <div className="section-content">
        {diff ? (
          <DiffHighlight oldText={diff.oldContent} newText={diff.newContent} />
        ) : data.content_type === 'native' ? (
          <p>{data.content}</p>
        ) : (
          <p className="external-note">External content — view at <a href={data.source_url} target="_blank" rel="noopener">{data.source_url}</a></p>
        )}
      </div>
      <div className="section-footer">
        <span className="hash">{data.content_hash}</span>
        <span className="type">{data.content_type}</span>
      </div>
    </div>
  )
}
