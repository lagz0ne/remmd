import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import {
  CheckCircleIcon,
  WarningCircleIcon,
  ArrowRightIcon,
  ArrowLeftIcon,
  ChatCircleIcon,
} from '@phosphor-icons/react'
import { useZoomLevel } from './use-zoom-level'
import { useSections } from '../hooks'
import { stateColor, validationBg, validationBorder, validationText } from '../theme/colors'
import type { DocNodeData } from './use-graph-data'

/* ── Derived data extracted once from DocNodeData ── */

interface NodeMetrics {
  errors: { rule: string; message: string }[]
  passing: number
  total: number
  outgoing: number
  incoming: number
  threadCount: number
  portWorstState: string | undefined
  hasErrors: boolean
  allPass: boolean
  valKind: 'error' | 'pass'
  borderColor: string
}

function deriveMetrics(data: DocNodeData): NodeMetrics {
  const errors = (data.validationErrors as { rule: string; message: string }[]) || []
  const passing = (data.validationPassing as number) || 0
  const total = (data.validationTotal as number) || 0
  const outgoing = (data.outgoing as number) || 0
  const incoming = (data.incoming as number) || 0
  const threadCount = (data.threadCount as number) || 0
  const portWorstState = data.portWorstState as string | undefined
  const hasErrors = errors.length > 0
  return {
    errors,
    passing,
    total,
    outgoing,
    incoming,
    threadCount,
    portWorstState,
    hasErrors,
    allPass: total > 0 && !hasErrors,
    valKind: hasErrors ? 'error' : 'pass',
    borderColor: data.worstState ? stateColor[data.worstState] || '#d4d4d8' : '#d4d4d8',
  }
}

/* ── Shared floating indicators ── */

function ValidationIndicator({ m }: { m: NodeMetrics }) {
  if (m.total <= 0) return null
  return (
    <div
      className="absolute flex items-center gap-1"
      style={{
        top: -14,
        left: -6,
        background: validationBg[m.valKind],
        border: `1px solid ${validationBorder[m.valKind]}`,
        borderRadius: 8,
        padding: '4px 8px',
        boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
      }}
    >
      {m.allPass ? (
        <CheckCircleIcon size={11} weight="light" style={{ color: validationText.pass }} />
      ) : (
        <WarningCircleIcon size={11} weight="light" style={{ color: validationText.error }} />
      )}
      <span className="text-[9px] font-medium" style={{ color: validationText[m.valKind] }}>
        {m.passing}/{m.total}
      </span>
    </div>
  )
}

function PortSummary({ m }: { m: NodeMetrics }) {
  if (m.outgoing <= 0 && m.incoming <= 0 && m.threadCount <= 0) return null
  return (
    <div className="absolute flex items-center gap-1" style={{ bottom: -14, right: -6 }}>
      {m.outgoing > 0 && (
        <PortPill
          icon={<ArrowRightIcon size={10} weight="light" />}
          count={m.outgoing}
          portWorstState={m.portWorstState}
        />
      )}
      {m.incoming > 0 && (
        <PortPill
          icon={<ArrowLeftIcon size={10} weight="light" />}
          count={m.incoming}
          portWorstState={m.portWorstState}
        />
      )}
      {m.threadCount > 0 && (
        <PortPill
          icon={<ChatCircleIcon size={10} weight="light" />}
          count={m.threadCount}
          portWorstState={m.portWorstState}
        />
      )}
    </div>
  )
}

/* ── Shared header: type label + owner tag + title + brief ── */

function NodeHeader({ data, truncateTitle }: { data: DocNodeData; truncateTitle?: boolean }) {
  return (
    <>
      <div className="flex items-center gap-2">
        <span className="text-[9px] text-zinc-400 font-normal">
          {(data.playbookType as string) || data.status}
        </span>
        {(data.owner as string) && (
          <span className="ml-auto text-[8px] bg-zinc-100 text-zinc-500 px-1.5 py-0.5 rounded-full">
            {data.owner as string}
          </span>
        )}
      </div>
      <div className={`text-[14px] font-semibold text-zinc-900${truncateTitle ? ' truncate' : ''}`}>
        {data.title}
      </div>
      {(data.brief as string) && (
        <div className="text-[10px] text-zinc-400 truncate mt-1">{data.brief as string}</div>
      )}
    </>
  )
}

/* ── Entry point ── */

function PlaybookNodeInner({ data, id, selected }: NodeProps & { data: DocNodeData }) {
  const zoom = useZoomLevel()

  if (zoom === 'far') return <FarView data={data} selected={selected} />
  if (zoom === 'medium') return <MediumView data={data} selected={selected} />
  return <CloseView data={data} docId={id} selected={selected} />
}

/* ── Far View ── zoom < 0.35 ── */

function FarView({ data, selected }: { data: DocNodeData; selected?: boolean }) {
  const hasErrors = (data.validationErrors as { rule: string; message: string }[])?.length > 0
  const isOrphan = (data.outgoing as number) === 0 && (data.incoming as number) === 0
  const borderColor = data.worstState ? stateColor[data.worstState] || '#d4d4d8' : '#d4d4d8'

  return (
    <div
      className={`rounded-lg px-2 py-1.5 bg-white text-center transition-shadow ${
        selected ? 'shadow-md ring-2 ring-blue-400' : ''
      }`}
      style={{
        border: `1.5px ${isOrphan ? 'dashed' : 'solid'} ${borderColor}`,
        background: hasErrors ? '#fef2f2' : 'white',
        minWidth: 50,
      }}
    >
      <Handle type="target" position={Position.Left} className="opacity-0" />
      <span className="text-[8px] font-semibold text-zinc-600 truncate block max-w-[80px]">
        {data.title}
      </span>
      <Handle type="source" position={Position.Right} className="opacity-0" />
    </div>
  )
}

/* ── Medium View ── 0.35-0.85 ── */

function MediumView({ data, selected }: { data: DocNodeData; selected?: boolean }) {
  const m = deriveMetrics(data)

  return (
    <div
      className={`relative rounded-xl bg-white transition-shadow hover:shadow-[0_4px_12px_rgba(0,0,0,0.06)] cursor-pointer ${
        selected ? 'shadow-md ring-2 ring-blue-400' : ''
      }`}
      style={{
        border: `1.5px solid ${m.borderColor}`,
        borderRadius: 12,
        padding: '16px 20px',
        minWidth: 170,
      }}
    >
      <Handle type="target" position={Position.Left} className="opacity-0" />
      <NodeHeader data={data} truncateTitle />
      <ValidationIndicator m={m} />
      <PortSummary m={m} />
      <Handle type="source" position={Position.Right} className="opacity-0" />
    </div>
  )
}

/* ── Close View ── zoom > 0.85 ── */

function CloseView({
  data,
  docId,
  selected,
}: {
  data: DocNodeData
  docId: string
  selected?: boolean
}) {
  const { data: sectionData } = useSections(docId)
  const sections = sectionData?.sections || []
  const m = deriveMetrics(data)

  return (
    <div
      className={`relative rounded-xl bg-white transition-shadow hover:shadow-[0_4px_12px_rgba(0,0,0,0.06)] cursor-pointer ${
        selected ? 'shadow-md ring-2 ring-blue-400' : ''
      }`}
      style={{
        border: `1.5px solid ${m.borderColor}`,
        borderRadius: 12,
        padding: '16px 20px',
        minWidth: 220,
        maxWidth: 320,
      }}
    >
      <Handle type="target" position={Position.Left} className="opacity-0" />
      <NodeHeader data={data} />

      {sections.length > 0 && (
        <div
          className="relative border-t border-zinc-100 pt-2 mt-2 space-y-1 overflow-hidden"
          style={{ maxHeight: 140 }}
        >
          {sections
            .slice(0, 6)
            .map(
              (s: {
                id: string
                ref: string
                title: string
                link_state?: string
                missing?: boolean
              }) => (
                <div
                  key={s.id}
                  className="flex items-center gap-1.5 text-xs text-zinc-600"
                >
                  <span
                    className="w-0.5 h-4 rounded-full shrink-0"
                    style={{
                      background: s.link_state
                        ? stateColor[s.link_state] || '#e4e4e7'
                        : '#e4e4e7',
                    }}
                  />
                  {(s as { missing?: boolean }).missing ? (
                    <span className="italic text-zinc-400">
                      {s.title || s.ref}{' '}
                      <span className="text-[9px] text-zinc-300">missing</span>
                    </span>
                  ) : (
                    <span className="truncate">{s.title || s.ref}</span>
                  )}
                </div>
              ),
            )}
          {sections.length > 6 && (
            <div className="text-[10px] text-zinc-400">+{sections.length - 6} more</div>
          )}
          <div
            className="absolute bottom-0 left-0 right-0 h-6 pointer-events-none"
            style={{ background: 'linear-gradient(transparent, white)' }}
          />
        </div>
      )}

      {m.hasErrors && (
        <div
          className="mt-2 text-[10px] space-y-0.5"
          style={{
            background: '#fef2f2',
            border: '1px solid #fecaca',
            borderRadius: 4,
            padding: '6px 8px',
          }}
        >
          {m.errors.map((e, i) => (
            <div key={i} style={{ color: '#dc2626' }}>
              {e.message}
            </div>
          ))}
        </div>
      )}

      <ValidationIndicator m={m} />
      <PortSummary m={m} />
      <Handle type="source" position={Position.Right} className="opacity-0" />
    </div>
  )
}

/* ── Port Pill helper ── */

function PortPill({
  icon,
  count,
  portWorstState,
}: {
  icon: React.ReactNode
  count: number
  portWorstState?: string
}) {
  const borderColor = portWorstState ? stateColor[portWorstState] || '#e4e4e7' : '#e4e4e7'
  return (
    <div
      className="flex items-center gap-0.5 text-[9px] text-zinc-500"
      style={{
        background: 'white',
        border: `1px solid ${borderColor}`,
        borderRadius: 8,
        padding: '4px 7px',
      }}
    >
      {icon}
      {count}
    </div>
  )
}

export const PlaybookNode = memo(PlaybookNodeInner)
