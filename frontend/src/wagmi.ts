import { http, createConfig } from 'wagmi'
import { mainnet, sepolia } from 'wagmi/chains'
import { injected } from 'wagmi/connectors'

// wagmi 全局配置：支持哪些链、用哪些连接器、怎么连节点。
// 这一关只做"签名登录"，还不读链上数据，但 wagmi 要求先声明链和 transport。
// injected() = 浏览器注入式钱包（MetaMask 就是这类）。
export const config = createConfig({
  chains: [sepolia, mainnet],
  connectors: [injected()],
  transports: {
    [sepolia.id]: http(),
    [mainnet.id]: http(),
  },
})
