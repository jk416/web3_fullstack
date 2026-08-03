// 和后端打交道的三个接口封装。走相对路径 /api，由 Vite 代理转发到 :8080（同源，无 CORS）。

// ⚠️ 消息契约：必须和 Go 后端 service.VerifyLogin 里的字符串【一字不差】，
// 否则两端哈希不同、验签必然失败。
export const buildLoginMessage = (nonce: string) =>
  `Sign in to Web3 Wallet Exchange.\n\nNonce: ${nonce}`

// ① 用地址换一个待签的 nonce
export async function getNonce(address: string): Promise<string> {
  const res = await fetch(`/api/auth/nonce?address=${address}`)
  if (!res.ok) throw new Error('failed to get nonce')
  const data = await res.json()
  return data.nonce
}

// ④ 用 {地址, 签名} 换 JWT
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

// 验证受保护接口：带上 JWT 调 /api/me
export async function getMe(token: string) {
  const res = await fetch('/api/me', {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error(`unauthorized (${res.status})`)
  return res.json()
}

// ⑤ 登录后取托管充值地址（JWT 鉴权；无则服务端创建）
// 响应字段与后端一致：{ "address": "0x..." } —— 这是充值地址，不是 MetaMask 登录地址
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
