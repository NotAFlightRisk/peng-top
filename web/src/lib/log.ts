type Level = 'debug' | 'info' | 'warn' | 'error'

const styles: Record<Level, string> = {
  debug: 'color:#c5a3ff',
  info: 'color:#78dce8',
  warn: 'color:#ffd866',
  error: 'color:#ff6188;font-weight:bold',
}

const debugOn = () => {
  try {
    return localStorage.getItem('peng-top:debug') === '1'
  } catch {
    return false
  }
}

// the only door to the console, debug shows when peng-top:debug is set
export function log(level: Level, message: string, detail?: unknown) {
  if (level === 'debug' && !debugOn()) return
  const time = new Date().toLocaleTimeString()
  const args: unknown[] = [`%c${level.toUpperCase()}%c ${time} ${message}`, styles[level], '']
  if (detail !== undefined) args.push(detail)
  console[level === 'debug' ? 'log' : level](...args)
}
