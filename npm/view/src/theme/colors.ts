/** CSS var references for link states. Use in inline styles. */
export const stateColor: Record<string, string> = {
  aligned: 'var(--aligned)',
  stale: 'var(--stale)',
  pending: 'var(--pending)',
  broken: 'var(--broken)',
  archived: 'var(--archived)',
}

export const validationBg: Record<string, string> = {
  pass: '#f0fdf4',
  error: '#fef2f2',
  warning: '#fffbeb',
  neutral: '#fafafa',
}

export const validationBorder: Record<string, string> = {
  pass: '#dcfce7',
  error: '#fecaca',
  warning: '#fef3c7',
  neutral: '#e4e4e7',
}

export const validationText: Record<string, string> = {
  pass: '#16a34a',
  error: '#dc2626',
  warning: '#d97706',
  neutral: '#a1a1aa',
}
