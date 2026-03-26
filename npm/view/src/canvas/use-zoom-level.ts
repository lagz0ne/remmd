import { useStore } from '@xyflow/react'

export type ZoomLevel = 'far' | 'medium' | 'close'

const FAR_THRESHOLD = 0.5
const CLOSE_THRESHOLD = 0.75

export function useZoomLevel(): ZoomLevel {
  return useStore((s) => {
    const zoom = s.transform[2]
    if (zoom < FAR_THRESHOLD) return 'far'
    if (zoom > CLOSE_THRESHOLD) return 'close'
    return 'medium'
  })
}

export function deriveZoomLevel(zoom: number): ZoomLevel {
  if (zoom < FAR_THRESHOLD) return 'far'
  if (zoom > CLOSE_THRESHOLD) return 'close'
  return 'medium'
}
