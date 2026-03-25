import { useMemo } from 'react'

interface DiffHighlightProps {
  oldText: string
  newText: string
}

export function DiffHighlight({ oldText, newText }: DiffHighlightProps) {
  const diff = useMemo(() => computeWordDiff(oldText, newText), [oldText, newText])

  return (
    <div className="diff-container">
      {diff.map((part, i) => {
        if (part.type === 'equal') return <span key={i}>{part.text}</span>
        if (part.type === 'insert') return <span key={i} className="diff-insert">{part.text}</span>
        if (part.type === 'delete') return <span key={i} className="diff-delete">{part.text}</span>
        return null
      })}
    </div>
  )
}

interface DiffPart { type: 'equal' | 'insert' | 'delete'; text: string }

// Simple word-level diff using common prefix/suffix approach
// POC-sufficient; production would use a proper LCS algorithm
function computeWordDiff(oldText: string, newText: string): DiffPart[] {
  const oldWords = oldText.split(/(\s+)/)
  const newWords = newText.split(/(\s+)/)

  const result: DiffPart[] = []

  let i = 0, j = 0
  // Common prefix
  while (i < oldWords.length && j < newWords.length && oldWords[i] === newWords[j]) {
    result.push({ type: 'equal', text: oldWords[i] })
    i++; j++
  }

  // Find common suffix
  let oi = oldWords.length - 1, nj = newWords.length - 1
  const suffix: DiffPart[] = []
  while (oi > i && nj > j && oldWords[oi] === newWords[nj]) {
    suffix.unshift({ type: 'equal', text: oldWords[oi] })
    oi--; nj--
  }

  // Middle: everything between prefix and suffix
  if (i <= oi) {
    result.push({ type: 'delete', text: oldWords.slice(i, oi + 1).join('') })
  }
  if (j <= nj) {
    result.push({ type: 'insert', text: newWords.slice(j, nj + 1).join('') })
  }

  result.push(...suffix)
  return result
}
