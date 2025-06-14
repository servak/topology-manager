import React, { useState, useEffect } from 'react'
import HierarchicalTopology from './components/HierarchicalTopology'
import DeviceClassificationBoard from './components/DeviceClassificationBoard'
import './App.css'

function App() {
  const [topology, setTopology] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [selectedDevice, setSelectedDevice] = useState('')
  const [activeTab, setActiveTab] = useState('classification')

  // 階層トポロジー取得
  const loadHierarchicalTopology = async (rootDevice = null, depth = 5) => {
    if (!rootDevice) return
    
    setLoading(true)
    setError(null)
    
    try {
      const params = new URLSearchParams({
        depth: depth.toString()
      })
      
      const response = await fetch(`/api/v1/topology/visual/${encodeURIComponent(rootDevice)}?${params}`)
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      
      const data = await response.json()
      setTopology(data)
    } catch (err) {
      console.error('Failed to load topology:', err)
      setError(`トポロジーの読み込みに失敗しました: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  // デバイス選択時の処理
  const handleNavigateToDevice = (deviceId) => {
    setSelectedDevice(deviceId)
    loadHierarchicalTopology(deviceId)
    
    // URL更新
    const url = new URL(window.location)
    url.searchParams.set('device', deviceId)
    window.history.pushState({}, '', url)
  }

  // URLパラメータからデバイスIDを読み込み
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const deviceFromUrl = params.get('device')
    if (deviceFromUrl) {
      setSelectedDevice(deviceFromUrl)
      loadHierarchicalTopology(deviceFromUrl)
    }
  }, [])

  const renderTabContent = () => {
    switch (activeTab) {
      case 'topology':
        return (
          <div className="topology-section">
            <div className="topology-controls">
              <input
                type="text"
                placeholder="デバイスIDを入力してトポロジーを表示"
                value={selectedDevice}
                onChange={(e) => setSelectedDevice(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter' && selectedDevice.trim()) {
                    loadHierarchicalTopology(selectedDevice.trim())
                  }
                }}
              />
              <button 
                onClick={() => selectedDevice.trim() && loadHierarchicalTopology(selectedDevice.trim())}
                disabled={loading || !selectedDevice.trim()}
              >
                {loading ? '読み込み中...' : 'トポロジー表示'}
              </button>
            </div>

            {error && (
              <div className="error-message">
                {error}
              </div>
            )}

            {topology && !loading && (
              <div className="topology-container">
                <div className="topology-header">
                  <div className="topology-stats">
                    <span>Nodes: {topology.stats.total_nodes}</span>
                    <span>Edges: {topology.stats.total_edges}</span>
                    <span>Root: {topology.root_device}</span>
                    <span>Depth: {topology.depth}</span>
                  </div>
                </div>
                
                <HierarchicalTopology 
                  topology={topology} 
                  onDeviceSelect={handleNavigateToDevice}
                  selectedDevice={selectedDevice}
                />
              </div>
            )}

            {!topology && !loading && !error && (
              <div className="welcome-message">
                <h3>🏗️ ネットワーク階層トポロジー</h3>
                <p>デバイスIDを入力して、階層表示でネットワーク構造を確認できます。</p>
              </div>
            )}
          </div>
        )
      case 'classification':
        return <DeviceClassificationBoard />
      default:
        return null
    }
  }

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-content">
          <h1>🌐 Network Topology Manager</h1>
          <p>ネットワーク機器の階層分類・管理システム</p>
        </div>
      </header>

      <main className="app-main">
        <nav className="nav-tabs">
          <button 
            className={`nav-tab ${activeTab === 'classification' ? 'active' : ''}`}
            onClick={() => setActiveTab('classification')}
          >
            🏷️ デバイス分類管理
          </button>
          <button 
            className={`nav-tab ${activeTab === 'topology' ? 'active' : ''}`}
            onClick={() => setActiveTab('topology')}
          >
            🗺️ 階層トポロジー
          </button>
        </nav>
        
        <div className="tab-content">
          {renderTabContent()}
        </div>
      </main>
    </div>
  )
}

export default App