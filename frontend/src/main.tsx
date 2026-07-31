import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { WagmiProvider } from 'wagmi'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { config } from './wagmi.ts'
import App from './App.tsx'

// wagmi 需要两层 Provider 包住整个应用：
// WagmiProvider 提供钱包配置；QueryClientProvider 是 wagmi 内部用来管异步状态的（react-query）。
const queryClient = new QueryClient()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </WagmiProvider>
  </StrictMode>,
)
