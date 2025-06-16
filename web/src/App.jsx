import React, { useState } from 'react'
import { BrowserRouter as Router, Routes, Route, Link, useLocation } from 'react-router-dom'
import ClassificationPage from './pages/ClassificationPage'
import HierarchyPage from './pages/HierarchyPage'
import TopologyPage from './pages/TopologyPage'
import SearchPage from './pages/SearchPage'
import StatusPage from './pages/StatusPage'
import './App.css'

function AppContent() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const location = useLocation()

  const navItems = [
    {
      section: "📊 管理メニュー",
      items: [
        { path: '/classification', icon: '🏷️', label: 'デバイス分類', title: 'デバイス分類管理' },
        { path: '/hierarchy', icon: '📋', label: '階層一覧', title: 'デバイス階層一覧' }
      ]
    },
    {
      section: "🗺️ トポロジー",
      items: [
        { path: '/topology', icon: '🌐', label: '階層表示', title: '階層トポロジー表示' },
        { path: '/search', icon: '🔍', label: 'デバイス検索', title: 'デバイス検索' }
      ]
    },
    {
      section: "⚙️ システム",
      items: [
        { path: '/status', icon: '💚', label: 'システム状態', title: 'システム状態' }
      ]
    }
  ]

  // 現在のページタイトルを取得
  const getCurrentPageTitle = () => {
    for (const section of navItems) {
      for (const item of section.items) {
        if (item.path === location.pathname) {
          return item.title
        }
      }
    }
    return 'Network Topology Manager'
  }

  return (
    <div className="app">
      <aside className={`sidebar ${sidebarCollapsed ? 'collapsed' : ''}`}>
        <div className="sidebar-header">
          <div className="app-logo">
            <span className="logo-icon">🌐</span>
            <span className="logo-text">Network Topology Manager</span>
          </div>
        </div>
        <nav className="sidebar-nav">
          {navItems.map((section, sectionIndex) => (
            <div key={sectionIndex} className="nav-section">
              <h3>{section.section}</h3>
              {section.items.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`nav-item ${location.pathname === item.path ? 'active' : ''}`}
                >
                  <span className="nav-icon">{item.icon}</span>
                  <span className="nav-label">{item.label}</span>
                </Link>
              ))}
            </div>
          ))}
        </nav>
      </aside>

      <div className="main-layout">
        <header className="page-header">
          <button 
            className="sidebar-toggle"
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
          >
            ☰
          </button>
          <h1 className="page-title">{getCurrentPageTitle()}</h1>
        </header>

        <main className="page-content">
          <Routes>
            <Route path="/" element={<ClassificationPage />} />
            <Route path="/classification" element={<ClassificationPage />} />
            <Route path="/hierarchy" element={<HierarchyPage />} />
            <Route path="/topology" element={<TopologyPage />} />
            <Route path="/search" element={<SearchPage />} />
            <Route path="/status" element={<StatusPage />} />
          </Routes>
        </main>
      </div>
    </div>
  )
}

function App() {
  return (
    <Router>
      <AppContent />
    </Router>
  )
}

export default App