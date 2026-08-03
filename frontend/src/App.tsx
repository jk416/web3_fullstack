import { useEffect, useMemo, useState } from 'react'
import { useAccount, useConnect, useDisconnect, useSignMessage } from 'wagmi'
import { buildLoginMessage, getNonce, login, getMe, getWallet } from './api'
import './App.css'

function shortAddr(addr?: string) {
  if (!addr) return ''
  return `${addr.slice(0, 6)}…${addr.slice(-4)}`
}

function App() {
  // useAccount：当前连接的钱包信息。address = MetaMask 登录地址（身份），不是托管充值地址。
  const { address, isConnected } = useAccount()
  const { connect, connectors, isPending: isConnecting } = useConnect()
  const { disconnect } = useDisconnect()
  // useSignMessage：私钥全程留在 MetaMask，前端只拿 signature。
  const { signMessageAsync } = useSignMessage()

  const [token, setToken] = useState(() => localStorage.getItem('token') ?? '')
  const [me, setMe] = useState('')
  const [depositAddress, setDepositAddress] = useState('')
  const [status, setStatus] = useState('')
  const [statusKind, setStatusKind] = useState<'idle' | 'ok' | 'error'>('idle')
  const [busy, setBusy] = useState(false)

  const qrUrl = useMemo(() => {
    if (!depositAddress) return ''
    // 演示用第三方 QR 图；生产可换本地库。data = 充值地址字符串。
    return `https://api.qrserver.com/v1/create-qr-code/?size=140x140&data=${encodeURIComponent(depositAddress)}`
  }, [depositAddress])

  // 刷新后若还有 JWT，自动拉一次充值地址（失败则清 token）
  useEffect(() => {
    if (!token) return
    let cancelled = false
    ;(async () => {
      try {
        const addr = await getWallet(token)
        if (!cancelled) {
          setDepositAddress(addr)
          setStatusKind('ok')
          setStatus('session restored · deposit address loaded')
        }
      } catch {
        if (!cancelled) {
          localStorage.removeItem('token')
          setToken('')
          setDepositAddress('')
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [token])

  async function handleLogin() {
    if (!address) return
    setBusy(true)
    try {
      setStatusKind('idle')
      setStatus('requesting nonce…')
      const nonce = await getNonce(address) // ① 用登录地址拿 nonce
      setStatus('please sign in MetaMask…')
      const message = buildLoginMessage(nonce) // ② 与后端消息契约一致
      const signature = await signMessageAsync({ message }) // ③ 钱包签名
      setStatus('logging in…')
      const jwt = await login(address, signature) // ④ 换 JWT
      setToken(jwt)
      localStorage.setItem('token', jwt)

      setStatus('fetching deposit address…')
      const deposit = await getWallet(jwt) // ⑤ 托管充值地址（可能首次创建）
      setDepositAddress(deposit)
      setStatusKind('ok')
      setStatus('logged in · deposit wallet ready')
    } catch (e) {
      setStatusKind('error')
      setStatus('login failed: ' + (e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  async function handleMe() {
    try {
      const data = await getMe(token)
      setMe(JSON.stringify(data))
    } catch (e) {
      setMe('error: ' + (e as Error).message)
    }
  }

  async function handleRefreshWallet() {
    if (!token) return
    setBusy(true)
    try {
      const deposit = await getWallet(token)
      setDepositAddress(deposit)
      setStatusKind('ok')
      setStatus('deposit address refreshed (same address if already created)')
    } catch (e) {
      setStatusKind('error')
      setStatus('get wallet failed: ' + (e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  async function copyDeposit() {
    if (!depositAddress) return
    try {
      await navigator.clipboard.writeText(depositAddress)
      setStatusKind('ok')
      setStatus('deposit address copied')
    } catch {
      setStatusKind('error')
      setStatus('copy failed — select and copy manually')
    }
  }

  function handleDisconnect() {
    disconnect()
    localStorage.removeItem('token')
    setToken('')
    setDepositAddress('')
    setMe('')
    setStatus('')
    setStatusKind('idle')
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          <div className="brand-mark" aria-hidden />
          <div>
            <h1>Web3 Wallet Exchange</h1>
            <p>Custodial deposit · SIWE login demo</p>
          </div>
        </div>
        <div className="pill">
          <span className={`pill-dot ${isConnected ? 'on' : ''}`} />
          {isConnected ? `Wallet ${shortAddr(address)}` : 'Wallet disconnected'}
        </div>
      </header>

      <section className="hero">
        <h2>Sign in, then get your deposit address</h2>
        <p>
          MetaMask proves who you are. The system then issues a <strong>separate</strong> custodial
          address for deposits — private key stays encrypted on the server.
        </p>
      </section>

      <div className="grid two">
        {/* 左：连接 + SIWE 登录 */}
        <section className="card">
          <h3>1. Identity (MetaMask)</h3>
          <p className="sub">Login address lives in <code>users.wallet_address</code>.</p>

          {!isConnected ? (
            <div className="actions">
              <button
                className="btn-primary"
                onClick={() => connect({ connector: connectors[0] })}
                disabled={isConnecting}
              >
                {isConnecting ? 'Connecting…' : 'Connect MetaMask'}
              </button>
            </div>
          ) : (
            <>
              <div className="field">
                <span className="field-label">Connected login address</span>
                <div className="mono-box">
                  <code>{address}</code>
                </div>
              </div>

              <div className="actions">
                <button className="btn-primary" onClick={handleLogin} disabled={busy}>
                  {token ? 'Sign in again' : 'Sign in with Ethereum'}
                </button>
                <button className="btn-ghost" onClick={handleDisconnect}>
                  Disconnect
                </button>
              </div>

              {token && (
                <>
                  <div className="field">
                    <span className="field-label">JWT (truncated)</span>
                    <div className="mono-box">
                      <code>{token.slice(0, 36)}…</code>
                    </div>
                  </div>
                  <div className="actions">
                    <button className="btn-secondary" onClick={handleMe} disabled={busy}>
                      Call /api/me
                    </button>
                    <button className="btn-secondary" onClick={handleRefreshWallet} disabled={busy}>
                      Refresh /api/wallet
                    </button>
                  </div>
                  {me && (
                    <p className="hint">
                      /api/me → <code>{me}</code>
                    </p>
                  )}
                </>
              )}
            </>
          )}

          {status && (
            <div className={`status ${statusKind === 'error' ? 'error' : statusKind === 'ok' ? 'ok' : ''}`}>
              {status}
            </div>
          )}
        </section>

        {/* 右：托管充值地址 */}
        <section className="card">
          <h3>2. Custodial deposit</h3>
          <p className="sub">
            From <code>GET /api/wallet</code> → <code>wallets.address</code>. Send test ETH here later
            (stage 3 scanner).
          </p>

          {!token ? (
            <p className="hint">
              Sign in first. The backend will <strong>get-or-create</strong> one deposit wallet per
              user.
            </p>
          ) : !depositAddress ? (
            <p className="hint">Loading deposit address…</p>
          ) : (
            <div className="deposit">
              <div className="qr-wrap">
                {qrUrl ? (
                  <img src={qrUrl} alt="Deposit address QR" />
                ) : (
                  <span style={{ color: '#666', fontSize: 12 }}>QR</span>
                )}
              </div>
              <div>
                <div className="badge-row">
                  <span className="badge green">deposit only</span>
                  <span className="badge">not your MetaMask address</span>
                </div>
                <span className="field-label">Your deposit address</span>
                <div className="mono-box">
                  <code>{depositAddress}</code>
                  <button type="button" className="copy-btn" onClick={copyDeposit}>
                    Copy
                  </button>
                </div>
                <p className="hint">
                  Calling wallet again is idempotent: same <code>user_id</code> → same address.
                </p>
              </div>
            </div>
          )}
        </section>
      </div>

      <p className="footer-note">
        Demo UI for learning · backend is the hiring story · Sepolia deposit comes in stage 3
      </p>
    </div>
  )
}

export default App
