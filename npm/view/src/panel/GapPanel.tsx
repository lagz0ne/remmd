import { useState } from 'react'
import {
  CheckCircleIcon,
  WarningCircleIcon,
  CaretDownIcon,
  CaretUpIcon,
  XIcon,
} from '@phosphor-icons/react'
import { validationBg, validationBorder, validationText } from '../theme/colors'

interface GapPanelProps {
  docId: string
  docTitle: string
  playbookType: string
  owner: string
  validationErrors: { rule: string; message: string }[]
  validationPassing: number
  validationTotal: number
  onClose: () => void
}

const STORAGE_KEY = 'remmd-passing-collapsed'

function readCollapsed(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) !== 'false'
  } catch {
    return true
  }
}

export function GapPanel({
  docTitle,
  playbookType,
  owner,
  validationErrors,
  validationPassing,
  validationTotal,
  onClose,
}: GapPanelProps) {
  const [passingCollapsed, setPassingCollapsed] = useState(readCollapsed)

  const togglePassing = () => {
    const next = !passingCollapsed
    setPassingCollapsed(next)
    try {
      localStorage.setItem(STORAGE_KEY, String(next))
    } catch {
      /* noop */
    }
  }

  const errorCount = validationErrors.length
  const hasErrors = errorCount > 0

  return (
    <div className="flex-1 flex flex-col overflow-hidden min-w-0">
      <div className="flex-1 overflow-y-auto px-6 py-4">
        <div className="mb-4">
          <div className="flex items-center gap-2">
            <span className="text-[9px] text-zinc-400 uppercase tracking-wide">
              {playbookType}
            </span>
            {hasErrors ? (
              <span
                className="text-[9px] px-1.5 py-0.5 rounded-full font-medium"
                style={{
                  background: validationBg.error,
                  color: validationText.error,
                  border: `1px solid ${validationBorder.error}`,
                }}
              >
                {errorCount} issue{errorCount !== 1 ? 's' : ''}
              </span>
            ) : (
              <span
                className="text-[9px] px-1.5 py-0.5 rounded-full font-medium"
                style={{
                  background: validationBg.pass,
                  color: validationText.pass,
                  border: `1px solid ${validationBorder.pass}`,
                }}
              >
                all passing
              </span>
            )}
          </div>

          <div className="flex items-center justify-between mt-1">
            <h2 className="text-[14px] font-semibold text-zinc-900">
              {docTitle}
            </h2>
            <button
              onClick={onClose}
              className="text-zinc-400 hover:text-zinc-700 p-0.5"
            >
              <XIcon size={14} weight="light" />
            </button>
          </div>

          <div className="text-[10px] text-zinc-400 mt-0.5">
            owner: {owner}
          </div>
        </div>

        {hasErrors && (
          <div className="mb-4">
            <div className="text-[10px] font-semibold text-zinc-600 mb-2">
              Needs attention
            </div>
            <div
              className="rounded-lg p-3 space-y-2"
              style={{ background: validationBg.error }}
            >
              {validationErrors.map((err, i) => (
                <div
                  key={`${err.rule}-${i}`}
                  className="bg-white rounded-md p-2.5"
                  style={{ border: `1px solid ${validationBorder.error}` }}
                >
                  <div className="flex items-center gap-1.5">
                    <WarningCircleIcon
                      size={12}
                      weight="light"
                      style={{ color: validationText.error }}
                      className="shrink-0"
                    />
                    <span className="text-[11px] font-semibold text-zinc-800">
                      {err.rule}
                    </span>
                  </div>
                  <div className="text-[10px] text-zinc-500 mt-1 ml-[18px]">
                    {err.message}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {validationPassing > 0 && (
          <div>
            <button
              onClick={togglePassing}
              className="flex items-center gap-1.5 text-[10px] text-zinc-500 hover:text-zinc-700 w-full text-left py-1"
            >
              <CheckCircleIcon
                size={12}
                weight="light"
                style={{ color: validationText.pass }}
              />
              <span>
                {validationPassing} passing
              </span>
              {passingCollapsed ? (
                <CaretDownIcon size={10} weight="light" className="ml-auto" />
              ) : (
                <CaretUpIcon size={10} weight="light" className="ml-auto" />
              )}
            </button>

            {!passingCollapsed && (
              <div className="text-[10px] text-zinc-400 mt-1 pl-1">
                {validationPassing} rules passing
              </div>
            )}
          </div>
        )}

        <div className="mt-4 pt-3 border-t border-zinc-100 text-[9px] text-zinc-400">
          {validationPassing}/{validationTotal} checks passing
        </div>
      </div>
    </div>
  )
}
