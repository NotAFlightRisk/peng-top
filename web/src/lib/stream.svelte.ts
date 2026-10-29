import { log } from './log'
import type { Info, Snapshot } from './types'

type Status = 'connecting' | 'live' | 'paused' | 'lost'

const keep = 300

class Stream {
  info = $state.raw<Info | null>(null)
  latest = $state.raw<Snapshot | null>(null)
  history = $state.raw<Snapshot[]>([])
  status = $state<Status>('connecting')
  #source: EventSource | null = null

  start() {
    this.#connect()
    this.refreshInfo()
    setInterval(() => this.refreshInfo(), 30_000)
  }

  toggle() {
    if (this.status === 'paused') {
      this.#connect()
    } else {
      this.#source?.close()
      this.status = 'paused'
    }
  }

  async refreshInfo() {
    try {
      const res = await fetch('api/info', { cache: 'no-store' })
      if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
      this.info = await res.json()
    } catch (err) {
      log('warn', "couldn't fetch the system info", err)
    }
  }

  #connect() {
    this.#source?.close()
    this.status = 'connecting'
    const source = new EventSource('api/stream')
    this.#source = source
    source.addEventListener('history', (e) => {
      this.history = (parse<Snapshot[]>(e.data) ?? []).slice(-keep)
      this.latest = this.history.at(-1) ?? null
      this.status = 'live'
      log('debug', `connected, ${this.history.length} readings of history`)
    })
    source.onmessage = (e) => {
      const reading = parse<Snapshot>(e.data)
      if (!reading) return
      this.latest = reading
      this.history = [...this.history.slice(-(keep - 1)), { ...reading, procs: undefined }]
      this.status = 'live'
    }
    source.onerror = () => {
      if (this.status !== 'paused') this.status = 'lost'
      log('warn', 'lost the stream, it retries by itself')
    }
  }
}

function parse<T>(data: string): T | null {
  try {
    return JSON.parse(data) as T
  } catch (err) {
    log('error', 'got a reading that is not JSON', err)
    return null
  }
}

export const stream = new Stream()
