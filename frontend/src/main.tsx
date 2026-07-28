import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'

// 入口：把 <App/> 挂到 index.html 里的 <div id="root">。
// 类比 Java：像 main() 里把 DispatcherServlet 装到容器上——整个 UI 的启动点。
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
