declare module 'virtual:remmd/schema' {
  export const SECTION_TYPES: readonly string[]
  export const CONTENT_TYPES: readonly string[]
  export const LINK_STATES: readonly string[]
  export const RELATIONSHIP_TYPES: readonly string[]
  export const INTERVENTION_LEVELS: readonly string[]

  export type SectionType = string
  export type ContentType = 'native' | 'external'
  export type LinkState = 'pending' | 'aligned' | 'stale' | 'broken' | 'archived'
  export type RelationshipType = string
  export type InterventionLevel = string

  export interface Section {
    id: string
    ref: string
    title: string
    content: string
    content_hash: string
    content_type: ContentType
    source_url?: string
    link_state?: LinkState
  }
}

declare module 'virtual:section/*' {
  interface SectionData {
    id: string
    ref: string
    title: string
    content: string
    content_hash: string
    content_type: 'native' | 'external'
    source_url?: string
    link_state?: string
  }
  const data: SectionData
  export default data
}
