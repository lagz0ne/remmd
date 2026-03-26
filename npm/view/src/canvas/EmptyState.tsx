export function EmptyState() {
  return (
    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
      <div className="text-center max-w-[400px]">
        <div className="text-sm font-semibold text-zinc-800">No documents yet</div>
        <div className="text-[11px] text-zinc-400 mt-2 leading-relaxed">
          Import a playbook to define your graph structure,
          then create documents that follow it.
        </div>
        <div className="mt-3 space-y-1.5">
          <code className="block text-[10px] bg-zinc-100 px-3 py-1.5 rounded text-zinc-600 font-mono">
            remmd playbook import c3.playbook.yaml
          </code>
          <code className="block text-[10px] bg-zinc-100 px-3 py-1.5 rounded text-zinc-600 font-mono">
            remmd doc create "My First Doc"
          </code>
        </div>
      </div>
    </div>
  )
}
