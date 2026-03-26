import { ChatCircleIcon, ArrowClockwiseIcon } from '@phosphor-icons/react'
import { stateColor } from '../theme/colors'

interface ThreadPanelProps {
  edgeType: string
  state: string
  sourceTitle: string
  targetTitle: string
  links: Array<{
    id: string
    state: string
    relationship_type: string
  }>
}

export function ThreadPanel({
  edgeType,
  state,
  sourceTitle,
  targetTitle,
  links,
}: ThreadPanelProps) {
  return (
    <div className="flex-1 flex flex-col overflow-hidden min-w-0">
      <div className="flex-1 overflow-y-auto px-6 py-4">
        <div className="mb-4">
          <div className="flex items-center gap-2">
            <span
              className="text-[9px] px-1.5 py-0.5 rounded-full font-medium text-white"
              style={{ background: stateColor[state] || '#a1a1aa' }}
            >
              {state}
            </span>
            <span className="text-[10px] text-zinc-500 font-mono">
              {edgeType}
            </span>
          </div>

          <div className="text-[14px] font-semibold text-zinc-900 mt-1.5">
            {sourceTitle}
            <span className="text-zinc-400 mx-1.5">&rarr;</span>
            {targetTitle}
          </div>

          <div className="text-[10px] text-zinc-400 mt-0.5">
            Changed recently &middot; awaiting review
          </div>
        </div>

        {state === 'stale' && (
          <div
            className="rounded-md px-3 py-2.5 mb-4 flex items-center gap-2"
            style={{
              background: '#fffbeb',
              border: '1px solid #fef3c7',
            }}
          >
            <ArrowClockwiseIcon
              size={13}
              weight="light"
              className="text-amber-500 shrink-0"
            />
            <span className="text-[10px] text-amber-700">
              Content changed since last alignment
            </span>
          </div>
        )}

        {links.length > 0 && (
          <div className="mb-4">
            <div className="text-[10px] font-semibold text-zinc-600 mb-2">
              Links &middot; {links.length}
            </div>
            <div className="space-y-1">
              {links.map((link) => (
                <div
                  key={link.id}
                  className="flex items-center gap-2 text-[10px] px-2 py-1.5 rounded border"
                  style={{
                    borderColor: stateColor[link.state] || '#e4e4e7',
                  }}
                >
                  <span
                    className="w-1.5 h-1.5 rounded-full shrink-0"
                    style={{
                      background: stateColor[link.state] || '#a1a1aa',
                    }}
                  />
                  <span className="text-zinc-700 font-mono">
                    {link.relationship_type}
                  </span>
                  <span className="text-zinc-400 ml-auto">{link.state}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        <div>
          <div className="text-[10px] font-semibold text-zinc-600 mb-2 flex items-center gap-1.5">
            <ChatCircleIcon size={12} weight="light" className="text-zinc-400" />
            Thread &middot; 0 entries
          </div>
          <div className="text-[10px] text-zinc-400 italic py-4 text-center">
            No thread entries yet. Start a conversation.
          </div>
        </div>
      </div>

      <div className="border-t border-zinc-100 px-6 py-3 flex items-center gap-2">
        <button className="text-[10px] font-medium bg-zinc-900 text-white px-3 py-1.5 rounded hover:bg-zinc-800 transition-colors">
          Reaffirm
        </button>
        <button className="text-[10px] font-medium bg-white border border-zinc-200 text-zinc-700 px-3 py-1.5 rounded hover:border-zinc-300 transition-colors">
          Comment
        </button>
        <button className="text-[10px] font-medium bg-white border border-red-200 text-red-600 px-3 py-1.5 rounded hover:border-red-300 transition-colors ml-auto">
          Withdraw
        </button>
      </div>
    </div>
  )
}
