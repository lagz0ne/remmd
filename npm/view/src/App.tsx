import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useNatsInvalidation, useSections, useDocuments } from './hooks'
import { Section } from './components/Section'

const queryClient = new QueryClient()

function DocumentList() {
  useNatsInvalidation()
  const { data: docs, isLoading } = useDocuments()
  const [selectedDocId, setSelectedDocId] = useState<string | null>(null)

  if (isLoading) return <div className="loading">Loading documents...</div>
  if (!docs || docs.length === 0) return <div className="loading">No documents. Create one with: remmd doc create "Title" --content "# Section"</div>

  const activeDocId = selectedDocId || docs[0].id

  return (
    <div className="doc-view">
      <header>
        <h1>remmd view</h1>
        <div className="meta">
          {docs.length > 1 && (
            <select value={activeDocId} onChange={e => setSelectedDocId(e.target.value)}>
              {docs.map(d => <option key={d.id} value={d.id}>{d.title}</option>)}
            </select>
          )}
          {docs.length === 1 && <span className="doc-title">{docs[0].title}</span>}
        </div>
      </header>
      <DocumentView docId={activeDocId} />
    </div>
  )
}

function DocumentView({ docId }: { docId: string }) {
  const { data, isLoading, error } = useSections(docId)

  if (isLoading) return <div className="loading">Loading sections...</div>
  if (error) return <div className="error">Error: {String(error)}</div>
  if (!data) return null

  return (
    <div className="sections">
      {data.sections.map((s: any) => (
        <Section key={s.id} ref_={s.ref} data={s} />
      ))}
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <DocumentList />
    </QueryClientProvider>
  )
}
