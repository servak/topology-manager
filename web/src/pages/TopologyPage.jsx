import React, { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import HierarchicalTopology from '../components/HierarchicalTopology'
import CytoscapeTopology from '../components/CytoscapeTopology'

function TopologyPage() {
  const [topology, setTopology] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [selectedDevice, setSelectedDevice] = useState('')
  const [viewMode, setViewMode] = useState('cytoscape') // 'cytoscape' or 'hierarchical'
  const [searchParams, setSearchParams] = useSearchParams()

  // URLパラメータからデバイスIDを読み込み
  useEffect(() => {
    const deviceFromUrl = searchParams.get('device')
    if (deviceFromUrl) {
      setSelectedDevice(deviceFromUrl)
      loadHierarchicalTopology(deviceFromUrl)
    }
  }, [searchParams])

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
    setSearchParams({ device: deviceId })
  }

  const handleDeviceChange = (e) => {
    setSelectedDevice(e.target.value)
  }

  const handleLoadTopology = () => {
    if (selectedDevice.trim()) {
      loadHierarchicalTopology(selectedDevice.trim())
      setSearchParams({ device: selectedDevice.trim() })
    }
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && selectedDevice.trim()) {
      handleLoadTopology()
    }
  }

  return (
    <div className="topology-view">
      <div className="topology-section">
        <div className="topology-controls">
          <input
            type="text"
            placeholder="デバイスIDを入力してトポロジーを表示"
            value={selectedDevice}
            onChange={handleDeviceChange}
            onKeyPress={handleKeyPress}
          />
          <button 
            onClick={handleLoadTopology}
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
              <div className="view-mode-selector">
                <button 
                  className={viewMode === 'cytoscape' ? 'active' : ''}
                  onClick={() => setViewMode('cytoscape')}
                >
                  🌐 ネットワークグラフ
                </button>
                <button 
                  className={viewMode === 'hierarchical' ? 'active' : ''}
                  onClick={() => setViewMode('hierarchical')}
                >
                  📋 階層リスト
                </button>
              </div>
              <div className="topology-stats">
                <span>Nodes: {topology.stats?.total_nodes || topology.nodes?.length || 0}</span>
                <span>Edges: {topology.stats?.total_edges || topology.edges?.length || 0}</span>
                <span>Root: {topology.root_device}</span>
                <span>Depth: {topology.depth}</span>
              </div>
            </div>
            
            {viewMode === 'cytoscape' ? (
              <CytoscapeTopology 
                topology={topology} 
                onDeviceSelect={handleNavigateToDevice}
                selectedDevice={selectedDevice}
              />
            ) : (
              <HierarchicalTopology 
                topology={topology} 
                onDeviceSelect={handleNavigateToDevice}
                selectedDevice={selectedDevice}
              />
            )}
          </div>
        )}

        {!topology && !loading && !error && (
          <div className="welcome-message">
            <h3>🏗️ ネットワーク階層トポロジー</h3>
            <p>デバイスIDを入力して、階層表示でネットワーク構造を確認できます。</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default TopologyPage