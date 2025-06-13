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
  const [groupingOptions, setGroupingOptions] = useState({
    enabled: true,  // デフォルトでグルーピングを有効化
    groupByPrefix: true,
    groupByType: false,
    groupByDepth: false,
    minGroupSize: 3,
    maxGroupDepth: 2,
    prefixMinLen: 3
  })

  // URLパラメータからデバイスIDを読み込み
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search)
    const deviceParam = urlParams.get('device')
    const depthParam = urlParams.get('depth')
    
    if (deviceParam) {
      setSelectedDevice(deviceParam)
      if (depthParam && !isNaN(parseInt(depthParam))) {
        setDepth(parseInt(depthParam))
      }
      // URLパラメータでデバイスが指定されている場合は自動で可視化実行
      fetchTopology(deviceParam, depthParam ? parseInt(depthParam) : depth)
    }
  }, [groupingOptions])

  const fetchTopology = async (hostname, explorationDepth = 3) => {
    if (!hostname) return

    setLoading(true)
    setError(null)

    try {
      // グルーピングパラメータを構築
      const params = new URLSearchParams({
        depth: explorationDepth.toString(),
        enable_grouping: groupingOptions.enabled.toString(),
        min_group_size: groupingOptions.minGroupSize.toString(),
        max_group_depth: groupingOptions.maxGroupDepth.toString(),
        group_by_prefix: groupingOptions.groupByPrefix.toString(),
        group_by_type: groupingOptions.groupByType.toString(),
        group_by_depth: groupingOptions.groupByDepth.toString(),
        prefix_min_len: groupingOptions.prefixMinLen.toString()
      })

      const response = await fetch(`/api/topology/${encodeURIComponent(hostname)}?${params}`)
      
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

  const handleGroupingChange = (newGroupingOptions) => {
    setGroupingOptions(newGroupingOptions)
    // グルーピング設定が変更されたら再取得
    if (selectedDevice) {
      fetchTopology(selectedDevice, depth)
    }
  }

  const handleGroupExpand = async (groupData) => {
    if (!topology || !selectedDevice) return

    setLoading(true)
    try {
      // グループ情報を取得
      const group = topology.groups.find(g => g.id === groupData.id)
      if (!group) {
        console.error('Group not found:', groupData.id)
        return
      }

      const response = await fetch('/api/topology/expand-group', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          group_id: group.id,
          root_device_id: selectedDevice,
          group_device_ids: group.device_ids,
          current_topology: topology,
          grouping_options: {
            enabled: groupingOptions.enabled,
            min_group_size: groupingOptions.minGroupSize,
            max_depth: groupingOptions.maxGroupDepth,
            group_by_prefix: groupingOptions.groupByPrefix,
            group_by_type: groupingOptions.groupByType,
            group_by_depth: groupingOptions.groupByDepth,
            prefix_min_len: groupingOptions.prefixMinLen
          },
          expand_depth: 2
        })
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      console.log('Group expand response:', data)
      console.log('Current topology nodes before update:', topology.nodes.length)
      console.log('Expanded topology nodes:', data.expanded_topology.nodes.length)
      setTopology(data.expanded_topology)
      console.log('Group expanded successfully. New nodes:', data.new_nodes.length, 'New edges:', data.new_edges.length)
    } catch (err) {
      setError(err.message)
      console.error('Failed to expand group:', err)
    } finally {
      setLoading(false)
    }
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
                    {topology.stats.total_groups > 0 && <span>Groups: {topology.stats.total_groups}</span>}
                    <span>Root: {topology.root_device}</span>
                    <span>Depth: {topology.depth}</span>
                  </div>
                  <TopologyGraph 
                    topology={topology} 
                    onObjectSelect={handleObjectSelect}
                    onGroupExpand={handleGroupExpand}
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
            groupingOptions={groupingOptions}
            onGroupingChange={handleGroupingChange}
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