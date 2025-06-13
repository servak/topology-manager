import React, { useState, useEffect } from 'react'
import TopologyGraph from './components/TopologyGraph'
import DeviceSelector from './components/DeviceSelector'
import DetailPanel from './components/DetailPanel'
import ReachabilityAnalysis from './components/ReachabilityAnalysis'
import PathAnalysis from './components/PathAnalysis'
import './App.css'

function App() {
  const [topology, setTopology] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [selectedDevice, setSelectedDevice] = useState('')
  const [depth, setDepth] = useState(3)
  const [selectedObject, setSelectedObject] = useState(null)
  const [activeTab, setActiveTab] = useState('topology')

  const fetchTopology = async (hostname, explorationDepth = 3) => {
    if (!hostname) return

    setLoading(true)
    setError(null)

    try {
      // 新しいAPI形式: /api/topology/{deviceId}?depth=N
      const response = await fetch(`/api/topology/${encodeURIComponent(hostname)}?depth=${explorationDepth}`)
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      setTopology(data)
    } catch (err) {
      setError(err.message)
      console.error('Failed to fetch topology:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleDeviceSearch = (hostname) => {
    setSelectedDevice(hostname)
    setSelectedObject(null) // 新しいトポロジー読み込み時は選択解除
    fetchTopology(hostname, depth)
  }

  const handleDepthChange = (newDepth) => {
    setDepth(newDepth)
    setSelectedObject(null) // 深度変更時は選択解除
    if (selectedDevice) {
      fetchTopology(selectedDevice, newDepth)
    }
  }

  const handleObjectSelect = (object) => {
    setSelectedObject(object)
  }

  const handleObjectDeselect = () => {
    setSelectedObject(null)
  }

  const renderTabContent = () => {
    switch (activeTab) {
      case 'topology':
        return (
          <div className="main-content">
            <div className="topology-section">
              {error && (
                <div className="error-message">
                  <h3>Error</h3>
                  <p>{error}</p>
                </div>
              )}
              
              {loading && (
                <div className="loading-message">
                  <p>Loading topology...</p>
                </div>
              )}

              {topology && !loading && (
                <div className="topology-container">
                  <div className="topology-stats">
                    <span>Nodes: {topology.stats.total_nodes}</span>
                    <span>Edges: {topology.stats.total_edges}</span>
                    <span>Root: {topology.root_device}</span>
                    <span>Depth: {topology.depth}</span>
                  </div>
                  <TopologyGraph 
                    topology={topology} 
                    onObjectSelect={handleObjectSelect}
                  />
                </div>
              )}

              {!topology && !loading && !error && (
                <div className="welcome-message">
                  <h2>Welcome to Network Topology Manager</h2>
                  <p>Enter a device ID above to visualize the network topology.</p>
                </div>
              )}
            </div>

            <div className="detail-section">
              <DetailPanel 
                selectedObject={selectedObject} 
                onClose={handleObjectDeselect}
              />
            </div>
          </div>
        )
      case 'reachability':
        return <ReachabilityAnalysis />
      case 'path':
        return <PathAnalysis />
      default:
        return null
    }
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>🌐 Network Topology Manager</h1>
        
        <nav className="app-nav">
          <button 
            className={`nav-tab ${activeTab === 'topology' ? 'active' : ''}`}
            onClick={() => setActiveTab('topology')}
          >
            🗺️ トポロジー可視化
          </button>
          <button 
            className={`nav-tab ${activeTab === 'reachability' ? 'active' : ''}`}
            onClick={() => setActiveTab('reachability')}
          >
            🔍 到達可能性分析
          </button>
          <button 
            className={`nav-tab ${activeTab === 'path' ? 'active' : ''}`}
            onClick={() => setActiveTab('path')}
          >
            🛤️ 最短パス分析
          </button>
        </nav>
        
        {activeTab === 'topology' && (
          <DeviceSelector
            onDeviceSelect={handleDeviceSearch}
            selectedDevice={selectedDevice}
            depth={depth}
            onDepthChange={handleDepthChange}
            loading={loading}
          />
        )}
      </header>

      <main className="app-main">
        {renderTabContent()}
      </main>
    </div>
  )
}

export default App