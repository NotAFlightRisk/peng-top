export interface CPU {
  total: number
  user: number
  system: number
  iowait: number
  steal: number
  cores: number[]
  load?: [number, number, number]
  mhz?: number
  quota?: number
}

export interface Mem {
  total: number
  used: number
  available: number
  cached: number
  buffers: number
  free: number
  swapTotal: number
  swapUsed: number
  limit?: number
}

export interface Mount {
  path: string
  device: string
  fs: string
  total: number
  used: number
  free: number
  inodes?: number
}

export interface DiskIO {
  name: string
  read: number
  write: number
  readOps: number
  writeOps: number
  busy: number
  readTotal: number
  writeTotal: number
}

export interface NIC {
  name: string
  rx: number
  tx: number
  rxTotal: number
  txTotal: number
  errors: number
  drops: number
  addrs?: string[]
  mac?: string
  mtu?: number
  up: boolean
  virtual: boolean
}

export interface Sensor {
  name: string
  temp: number
  high?: number
  crit?: number
}

export interface Proc {
  pid: number
  ppid: number
  name: string
  user: string
  cpu: number
  mem: number
  rss: number
  threads: number
  state: string
  cmd?: string
  started: number
  kernel?: boolean
}

export interface Snapshot {
  t: number
  cpu?: CPU
  mem?: Mem
  disks?: { mounts?: Mount[]; io?: DiskIO[] }
  net?: NIC[]
  sensors?: Sensor[]
  procs?: { total: number; running: number; threads: number; list: Proc[] }
}

export type Mode = 'host' | 'container' | 'host-via-container'

export interface Info {
  hostname: string
  os: string
  platform: string
  release: string
  kernel: string
  arch: string
  virt?: string
  bootTime: number
  cpuModel: string
  cores: number
  threads: number
  maxMhz?: number
  mode: Mode
  version: string
  interval: number
  issues: Record<string, string>
}
