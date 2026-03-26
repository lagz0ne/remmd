import { memo, useState } from 'react'
import {
  getBezierPath,
  EdgeLabelRenderer,
  type EdgeProps,
} from '@xyflow/react'
import { CheckCircleIcon, ChatCircleIcon } from '@phosphor-icons/react'
import type { BundleEdgeData } from './use-graph-data'
import { stateColor } from '../theme/colors'

const STRUCTURAL_TYPES = new Set(['contains', 'parent-of'])

function abbreviateType(type: string): string {
  if (type === 'contains') return 'cnt'
  return type
}

function BundledEdgeInner({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  selected,
}: EdgeProps & { data: BundleEdgeData }) {
  const [hovered, setHovered] = useState(false)

  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
  })

  const isStructural = data.isStructural || STRUCTURAL_TYPES.has(data.edgeType)
  const color = isStructural ? '#a1a1aa' : (stateColor[data.worstState] || '#a1a1aa')
  const baseOpacity = selected ? 1 : hovered ? 0.9 : 0.6
  const strokeWidth = Math.min(2 + data.count, 6)

  const typeName = data.count > 1
    ? `${data.count}\u00d7${abbreviateType(data.edgeType || 'link')}`
    : abbreviateType(data.edgeType || 'link')

  const threads = data.threadCount || 0

  return (
    <>
      <defs>
        <marker
          id={`arrow-${id}`}
          markerWidth="8"
          markerHeight="6"
          refX="7"
          refY="3"
          orient="auto"
        >
          <path d="M 0 0 L 8 3 L 0 6 Z" fill={color} opacity="0.6" />
        </marker>
      </defs>
      <path
        id={id}
        d={edgePath}
        stroke={color}
        strokeWidth={strokeWidth}
        strokeOpacity={baseOpacity}
        strokeDasharray={isStructural ? '4 3' : undefined}
        fill="none"
        markerEnd={`url(#arrow-${id})`}
        style={{ transition: 'stroke-opacity 0.2s ease' }}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
      />
      <path
        d={edgePath}
        stroke="transparent"
        strokeWidth={strokeWidth + 12}
        fill="none"
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
      />
      <EdgeLabelRenderer>
        <div
          className="absolute pointer-events-auto cursor-pointer flex items-stretch"
          style={{
            transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
          }}
          onMouseEnter={() => setHovered(true)}
          onMouseLeave={() => setHovered(false)}
        >
          <div
            className="flex items-center gap-1 px-1.5 py-0.5 text-[9px] font-medium"
            style={{
              background: '#fff',
              border: `1px solid ${color}`,
              borderRadius: threads > 0 || data.worstState === 'aligned'
                ? '4px 0 0 4px'
                : '4px',
              color,
            }}
          >
            {typeName}
          </div>

          {threads > 0 ? (
            <div
              className="flex items-center gap-0.5 px-1.5 py-0.5 text-[9px] font-medium"
              style={{
                background: `${color}18`,
                border: `1px solid ${color}`,
                borderLeft: 'none',
                borderRadius: '0 4px 4px 0',
                color,
              }}
            >
              <ChatCircleIcon size={9} weight="light" />
              {threads}
            </div>
          ) : data.worstState === 'aligned' ? (
            <div
              className="flex items-center px-1 py-0.5"
              style={{
                background: `${color}18`,
                border: `1px solid ${color}`,
                borderLeft: 'none',
                borderRadius: '0 4px 4px 0',
                color,
              }}
            >
              <CheckCircleIcon size={9} weight="light" />
            </div>
          ) : null}
        </div>
      </EdgeLabelRenderer>
    </>
  )
}

export const BundledEdge = memo(BundledEdgeInner)
