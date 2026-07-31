import { useState } from 'react'
import { useAccount, useConnect, useDisconnect, useSignMessage } from 'wagmi'
import { buildLoginMessage, getNonce, login, getMe } from './api'

function App() {
  // useAccount：当前连接的钱包信息。address 就是【钱包给出的地址】，不是用户打字输入的。
  const { address, isConnected } = useAccount()
  // useConnect：连接钱包；connectors[0] 就是我们在 wagmi.ts 配的 injected(MetaMask)。
  const { connect, connectors, isPending: isConnecting } = useConnect()
  const { disconnect } = useDisconnect()
  // useSignMessage：让钱包对一段消息签名（私钥全程留在 MetaMask，前端拿不到）。
  const { signMessageAsync } = useSignMessage()

  const [token, setToken] = useState('')
  const [me, setMe] = useState('')
  const [status, setStatus] = useState('')

  // 登录核心流程 —— 这几行是这一关要你看懂的重点：
  async function handleLogin() {
    if (!address) return
    try {
      setStatus('requesting nonce...')
      const nonce = await getNonce(address)                 // ① 用地址拿 nonce
      setStatus('please sign in your wallet...')
      const message = buildLoginMessage(nonce)              // ② 组装消息（和后端契约一致）
      const signature = await signMessageAsync({ message }) // ③ 钱包弹窗签名 → 得到 signature
      setStatus('logging in...')
      const jwt = await login(address, signature)           // ④ {地址,签名} 换 JWT
      setToken(jwt)
      localStorage.setItem('token', jwt)                    // 存起来，后续请求带上
      setStatus('logged in ✅')
    } catch (e) {
      setStatus('login failed: ' + (e as Error).message)
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

  return (
    <div style={{ fontFamily: 'sans-serif', padding: 40, lineHeight: 1.8 }}>
      <h1>Web3 Wallet Exchange</h1>

      {!isConnected ? (
        // 没连钱包：只有一个"连接钱包"按钮，点了 MetaMask 会弹窗
        <button onClick={() => connect({ connector: connectors[0] })} disabled={isConnecting}>
          {isConnecting ? 'Connecting…' : 'Connect Wallet'}
        </button>
      ) : (
        <>
          <p>
            Connected: <code>{address}</code>
          </p>
          <button onClick={handleLogin}>Sign in</button>{' '}
          <button onClick={() => disconnect()}>Disconnect</button>
          <p>Status: {status}</p>

          {token && (
            <>
              <p>
                JWT: <code>{token.slice(0, 40)}...</code>
              </p>
              <button onClick={handleMe}>Call /api/me (带 token)</button>
              {me && <p>/api/me → {me}</p>}
            </>
          )}
        </>
      )}
    </div>
  )
}

export default App
