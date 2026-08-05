// 和后端打交道的接口封装。走相对路径 /api，由 Vite 代理转发到 :8080。

export const buildLoginMessage = (nonce: string) =>
  `Sign in to Web3 Wallet Exchange.\n\nNonce: ${nonce}`

export async function getNonce(address: string): Promise<string> {
  const res = await fetch(`/api/auth/nonce?address=${address}`)
  if (!res.ok) throw new Error('failed to get nonce')
  const data = await res.json()
  return data.nonce
}

export async function login(address: string, signature: string): Promise<string> {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ address, signature }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error ?? 'login failed')
  }
  const data = await res.json()
  return data.token
}

export async function getMe(token: string) {
  const res = await fetch('/api/me', {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error(`unauthorized (${res.status})`)
  return res.json()
}

export async function getWallet(token: string): Promise<string> {
  const res = await fetch('/api/wallet', {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error ?? `get wallet failed (${res.status})`)
  }
  const data = await res.json()
  if (!data.address) throw new Error('wallet response missing address')
  return data.address as string
}

export type BalanceItem = { asset: string; available: string }

export async function getBalances(token: string): Promise<BalanceItem[]> {
  const res = await fetch('/api/balances', {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error ?? `get balances failed (${res.status})`)
  }
  const data = await res.json()
  const rows = data.balances ?? []
  return rows.map((r: Record<string, string>) => ({
    asset: r.asset ?? r.Asset ?? 'ETH',
    available: r.available ?? r.Available ?? '0',
  }))
}

export type DepositItem = {
  id: number
  txHash: string
  fromAddress: string
  toAddress: string
  amount: string
  asset: string
  status: string
  confirmations: number
  blockNumber: number
  createdAt?: string
}

export type DepositListQuery = {
  page?: number
  pageSize?: number
  status?: string
  asset?: string
  txHash?: string
}

export type DepositListResult = {
  items: DepositItem[]
  total: number
  page: number
  pageSize: number
}

function mapDeposit(r: Record<string, unknown>): DepositItem {
  return {
    id: Number(r.ID ?? r.id ?? 0),
    txHash: String(r.tx_hash ?? r.TxHash ?? ''),
    fromAddress: String(r.from_address ?? r.FromAddress ?? ''),
    toAddress: String(r.to_address ?? r.ToAddress ?? ''),
    amount: String(r.amount ?? r.Amount ?? '0'),
    asset: String(r.asset ?? r.Asset ?? 'ETH'),
    status: String(r.status ?? r.Status ?? ''),
    confirmations: Number(r.confirmations ?? r.Confirmations ?? 0),
    blockNumber: Number(r.block_number ?? r.BlockNumber ?? 0),
    createdAt: String(r.CreatedAt ?? r.created_at ?? ''),
  }
}

/**
 * 列表：GET /api/deposits?page&page_size&status&asset&tx_hash
 * 进入列表页 / 点查询 / 翻页时再调（不要在 App 根上常驻狂刷）
 */
export async function getDeposits(
  token: string,
  opts?: DepositListQuery,
): Promise<DepositListResult> {
  const page = opts?.page ?? 1
  const pageSize = opts?.pageSize ?? 10
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  })
  if (opts?.status) params.set('status', opts.status)
  if (opts?.asset) params.set('asset', opts.asset)
  if (opts?.txHash) params.set('tx_hash', opts.txHash)

  const res = await fetch(`/api/deposits?${params}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error ?? `get deposits failed (${res.status})`)
  }
  const data = await res.json()
  // 兼容 list（新）与 deposits（旧）
  const rows = data.list ?? data.deposits ?? []
  return {
    items: rows.map((r: Record<string, unknown>) => mapDeposit(r)),
    total: Number(data.total ?? 0),
    page: Number(data.page ?? page),
    pageSize: Number(data.page_size ?? pageSize),
  }
}

/** 详情：GET /api/deposits/:id —— 点行进入，再查一次 */
export async function getDepositDetail(token: string, id: number): Promise<DepositItem> {
  const res = await fetch(`/api/deposits/${id}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error ?? `get deposit detail failed (${res.status})`)
  }
  const data = await res.json()
  return mapDeposit(data.deposit ?? data)
}

export function formatWeiToEth(wei: string): string {
  try {
    const w = BigInt(wei || '0')
    const base = 10n ** 18n
    const whole = w / base
    const frac = w % base
    if (frac === 0n) return whole.toString()
    const fracStr = frac.toString().padStart(18, '0').replace(/0+$/, '')
    return `${whole}.${fracStr}`
  } catch {
    return wei
  }
}

export function shortAddr(addr?: string) {
  if (!addr) return ''
  if (addr.length < 12) return addr
  return `${addr.slice(0, 6)}…${addr.slice(-4)}`
}
