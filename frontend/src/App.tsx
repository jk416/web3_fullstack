import {useEffect, useState} from 'react'

// App 是一个组件：一个"返回 UI 的函数"。

// 类比后端：把它想成一个纯函数——输入 state/props，输出该渲染成什么样。
function App() {
    // useState：组件的"记忆"。status 是当前值，setStatus 是改它的唯一入口。
    // 一旦 setStatus 被调用，React 会重新执行 App() 并刷新界面。
    // 类比：像一个被监听的字段，赋值会自动触发"重绘"。
    const [status, setStatus] = useState<string>('loading...')
    const [msg, setMsg] = useState<string>('')

    function checkHealth() {
        fetch('/api/health')
            .then((res) => res.json())
            .then((data) => {
                setStatus(data.status)
                setMsg('✔ 检查成功')
                setTimeout(() => setMsg(''), 2000)
            })
            .catch(() => {
                setStatus('backend unreachable')
                setMsg('✘ 后端不可达')
                setTimeout(() => setMsg(''), 2000)
            })
    }

    // useEffect(fn, [])：空依赖数组 = 只在组件首次挂载后跑一次。
    // 类比：像生命周期回调 onMounted / @PostConstruct——初始化时做副作用（这里是发请求）。
    useEffect(() => {
        checkHealth()
    }, [])

    return (
        <div style={{fontFamily: 'sans-serif', padding: 40}}>
            <h1>Web3 Wallet Exchange</h1>
            <p>
                Backend health: <strong>{status}</strong>
            </p>
            <button onClick={checkHealth}>重新检查</button>
            {msg && <p style={{color: msg.startsWith('✔') ? 'green' : 'red'}}>{msg}</p>}
        </div>
    )
}

export default App
