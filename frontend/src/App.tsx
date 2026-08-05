import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import Layout from './Layout'
import HomePage from './pages/HomePage'
import DepositsPage from './pages/DepositsPage'
import DepositDetailPage from './pages/DepositDetailPage'
import './App.css'

/**
 * 路由：
 * /                首页
 * /deposits        列表（分页+条件，进页再查）
 * /deposits/:id    详情（进页再查）
 */
export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route index element={<HomePage />} />
          <Route path="deposits" element={<DepositsPage />} />
          <Route path="deposits/:id" element={<DepositDetailPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
