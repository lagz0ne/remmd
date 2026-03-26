import { useState } from 'react'
import { useSections } from '../hooks'
import { stateColor } from '../theme/colors'

interface DocPanelProps {
  docId: string
  docTitle: string
  onEdgeSelect?: (edgeId: string) => void
}

export function DocPanel({ docId, docTitle }: DocPanelProps) {
  const { data, isLoading } = useSections(docId)
  const [footerOpen, setFooterOpen] = useState(false)

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center text-sm text-zinc-400">
        Loading...
      </div>
    )
  }

  const sections = data?.sections || []

  return (
    <div className="flex-1 flex flex-col overflow-hidden min-w-0">
      <div className="flex-1 overflow-y-auto px-6 py-4">
        <h2 className="text-lg font-semibold text-zinc-900 mb-4 font-serif">
          {docTitle}
        </h2>

        <div className="space-y-3">
          {sections.map((s) => (
            <SectionCard key={s.id} section={s} />
          ))}
        </div>

        {sections.length === 0 && (
          <p className="text-sm text-zinc-400 italic">No sections</p>
        )}
      </div>

      <div className="border-t border-zinc-100">
        <button
          onClick={() => setFooterOpen(!footerOpen)}
          className="w-full px-6 py-2 text-left text-[11px] text-zinc-400 hover:text-zinc-600 font-mono flex items-center gap-1"
        >
          <span className={`transition-transform ${footerOpen ? 'rotate-90' : ''}`}>
            ▸
          </span>
          metadata
        </button>
        {footerOpen && (
          <div className="px-6 pb-3 text-[11px] text-zinc-500 font-mono space-y-1">
            <div>id: {docId}</div>
            <div>sections: {sections.length}</div>
            <div>
              states:{' '}
              {Object.entries(
                sections.reduce((acc: Record<string, number>, s) => {
                  const st = s.link_state || 'none'
                  acc[st] = (acc[st] || 0) + 1
                  return acc
                }, {}),
              )
                .map(([k, v]) => `${v} ${k}`)
                .join(', ')}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function SectionCard({ section }: { section: { id: string; ref: string; title: string; content: string; content_type: string; link_state?: string; source_url?: string } }) {
  return (
    <div className="border-l-2 pl-3 py-1" style={{ borderColor: section.link_state ? stateColor[section.link_state] || '#e4e4e7' : '#e4e4e7' }}>
      <div className="flex items-center gap-2 mb-1">
        <span className="text-[10px] text-zinc-400 font-mono bg-zinc-50 px-1 rounded">
          {section.ref}
        </span>
        <span className="text-sm font-medium text-zinc-800">{section.title}</span>
        {section.link_state && (
          <span
            className="text-[9px] px-1.5 py-0.5 rounded text-white ml-auto"
            style={{ background: stateColor[section.link_state] }}
          >
            {section.link_state}
          </span>
        )}
      </div>
      <div className="text-sm text-zinc-600 leading-relaxed">
        {section.content_type === 'native' ? (
          <p>{section.content}</p>
        ) : (
          <p className="italic text-zinc-400">
            External content —{' '}
            <a href={section.source_url} className="underline" target="_blank" rel="noopener">
              source
            </a>
          </p>
        )}
      </div>
    </div>
  )
}
