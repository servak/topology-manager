import React, { useState, useEffect } from 'react'
import TopologyGraph from './components/TopologyGraph'
import HierarchicalTopology from './components/HierarchicalTopology'
import DeviceSelector from './components/DeviceSelector'
import DetailPanel from './components/DetailPanel'
import ReachabilityAnalysis from './components/ReachabilityAnalysis'
import PathAnalysis from './components/PathAnalysis'
import DeviceClassification from './components/DeviceClassification'
import DeviceClassificationBoard from './components/DeviceClassificationBoard'
import './App.css'

function App() {
  const [topology, setTopology] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [selectedDevice, setSelectedDevice] = useState('')
  const [depth, setDepth] = useState(3)
  const [selectedObject, setSelectedObject] = useState(null)
  const [activeTab, setActiveTab] = useState('topology')
  const [showNeighbors, setShowNeighbors] = useState(null) // 隣接デバイス表示用
  const [groupingOptions, setGroupingOptions] = useState({
    enabled: true,  // デフォルトでグルーピングを有効化
    groupByPrefix: true,
    groupByType: false,
    groupByDepth: false,
    minGroupSize: 3,
    maxGroupDepth: 2,
    prefixMinLen: 3
  })
  const [expandedDevices, setExpandedDevices] = useState(new Set()) // 展開済みデバイスを追跡
  const [viewMode, setViewMode] = useState('graph') // 'graph' or 'hierarchy'

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
      
      // トポロジー読み込み完了後、rootデバイスを自動選択
      if (data && data.nodes) {
        const rootNode = data.nodes.find(node => node.is_root === true)
        if (rootNode) {
          console.log('Auto-selecting root device:', rootNode.id)
          setSelectedObject({
            type: 'node',
            data: {
              id: rootNode.id,
              label: rootNode.name,
              type: rootNode.type,
              hardware: rootNode.hardware,
              status: rootNode.status,
              layer: rootNode.layer,
              isRoot: rootNode.is_root
            }
          })
        }
      }
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

  // Actions用のコールバック関数
  const handleNavigateToDevice = (deviceId) => {
    console.log('Navigating to device:', deviceId)
    setSelectedDevice(deviceId)
    setSelectedObject(null) // 選択解除
    fetchTopology(deviceId, depth)
    
    // URLも更新
    const newUrl = `${window.location.pathname}?device=${encodeURIComponent(deviceId)}&depth=${depth}`
    window.history.pushState({}, '', newUrl)
  }

  const handleShowNeighbors = (deviceId) => {
    console.log('Showing neighbors for device:', deviceId)
    if (!topology) return

    // 隣接デバイスのIDを取得
    const neighborIds = new Set()
    topology.edges.forEach(edge => {
      if (edge.source === deviceId) {
        neighborIds.add(edge.target)
      } else if (edge.target === deviceId) {
        neighborIds.add(edge.source)
      }
    })

    console.log('Neighbor devices:', Array.from(neighborIds))
    
    // 隣接デバイス情報を設定して表示
    setShowNeighbors({
      deviceId: deviceId,
      neighbors: Array.from(neighborIds)
    })
  }

  const handleCloseNeighbors = () => {
    setShowNeighbors(null)
  }

  const handleNeighborClick = (neighborId) => {
    // 隣接デバイスをクリックした時の処理
    handleNavigateToDevice(neighborId)
    setShowNeighbors(null) // 隣接デバイス表示を閉じる
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
      const group = topology.groups?.find(g => g.id === groupData.id)
      if (!group) {
        console.error('Group not found:', groupData.id)
        return
      }

      console.log('Expanding group with devices:', group.device_ids)

      // 展開済みデバイスに追加
      const newExpandedDevices = new Set(expandedDevices)
      group.device_ids.forEach(deviceId => newExpandedDevices.add(deviceId))
      setExpandedDevices(newExpandedDevices)

      // 各グループ内デバイスから新しいトポロジーを取得
      const expandedNodesMap = new Map()
      const expandedEdgesMap = new Map()

      for (const deviceId of group.device_ids) {
        const params = new URLSearchParams({
          depth: '2',
          enable_grouping: groupingOptions.enabled.toString(),
          min_group_size: Math.max(groupingOptions.minGroupSize + 2, 5).toString(), // より厳しい条件
          max_group_depth: Math.max(groupingOptions.maxGroupDepth + 1, 3).toString(), // より深いレベルでグループ化
          group_by_prefix: groupingOptions.groupByPrefix.toString(),
          group_by_type: groupingOptions.groupByType.toString(),
          group_by_depth: groupingOptions.groupByDepth.toString(),
          prefix_min_len: groupingOptions.prefixMinLen.toString()
        })

        const response = await fetch(`/api/topology/${encodeURIComponent(deviceId)}/expand?${params}`)
        
        if (!response.ok) {
          console.warn(`Failed to expand from device ${deviceId}: ${response.status}`)
          continue
        }

        const data = await response.json()
        
        // 新しいノードをマップに追加（重複を避ける）
        data.nodes.forEach(node => {
          expandedNodesMap.set(node.id, node)
        })
        
        // 新しいエッジをマップに追加（重複を避ける）
        data.edges.forEach(edge => {
          expandedEdgesMap.set(edge.id, edge)
        })
      }

      // 既存のトポロジーを更新
      const updatedTopology = { ...topology }

      // グループノードを削除
      updatedTopology.nodes = updatedTopology.nodes.filter(node => node.id !== group.id)
      
      // グループに接続されたエッジを削除
      updatedTopology.edges = updatedTopology.edges.filter(edge => 
        edge.source !== group.id && edge.target !== group.id
      )
      
      // グループ情報を削除
      updatedTopology.groups = (updatedTopology.groups || []).filter(g => g.id !== group.id)

      // 新しいノードとエッジを追加（既存のものと重複しないように）
      expandedNodesMap.forEach((node, nodeId) => {
        if (!updatedTopology.nodes.find(n => n.id === nodeId)) {
          updatedTopology.nodes.push(node)
        }
      })
      
      expandedEdgesMap.forEach((edge, edgeId) => {
        if (!updatedTopology.edges.find(e => e.id === edgeId)) {
          updatedTopology.edges.push(edge)
        }
      })

      // 統計情報を更新
      updatedTopology.stats.total_nodes = updatedTopology.nodes.length
      updatedTopology.stats.total_edges = updatedTopology.edges.length
      updatedTopology.stats.total_groups = (updatedTopology.groups || []).length

      setTopology(updatedTopology)
      console.log('Group expanded successfully. Total nodes:', updatedTopology.nodes.length, 'Total edges:', updatedTopology.edges.length)
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
                  <div className="topology-header">
                    <div className="topology-stats">
                      <span>Nodes: {topology.stats.total_nodes}</span>
                      <span>Edges: {topology.stats.total_edges}</span>
                      {topology.stats?.total_groups > 0 && <span>Groups: {topology.stats.total_groups}</span>}
                      <span>Root: {topology.root_device}</span>
                      <span>Depth: {topology.depth}</span>
                    </div>
                    <div className="view-mode-toggle">
                      <button 
                        className={`view-mode-btn ${viewMode === 'graph' ? 'active' : ''}`}
                        onClick={() => setViewMode('graph')}
                      >
                        🗺️ グラフ表示
                      </button>
                      <button 
                        className={`view-mode-btn ${viewMode === 'hierarchy' ? 'active' : ''}`}
                        onClick={() => setViewMode('hierarchy')}
                      >
                        🏗️ 階層表示
                      </button>
                    </div>
                  </div>
                  
                  {viewMode === 'graph' ? (
                    <TopologyGraph 
                      topology={topology} 
                      onObjectSelect={handleObjectSelect}
                      onGroupExpand={handleGroupExpand}
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
                  <h2>Welcome to Network Topology Manager</h2>
                  <p>Enter a device ID above to visualize the network topology.</p>
                </div>
              )}
            </div>

            <div className="detail-section">
              <DetailPanel 
                selectedObject={selectedObject} 
                onClose={handleObjectDeselect}
                onNavigateToDevice={handleNavigateToDevice}
                onShowNeighbors={handleShowNeighbors}
                showNeighbors={showNeighbors}
                onCloseNeighbors={handleCloseNeighbors}
                onNeighborClick={handleNeighborClick}
              />
            </div>
          </div>
        )
      case 'reachability':
        return <ReachabilityAnalysis />
      case 'path':
        return <PathAnalysis />
      case 'classification':
        return <DeviceClassificationBoard />
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
          <button 
            className={`nav-tab ${activeTab === 'classification' ? 'active' : ''}`}
            onClick={() => setActiveTab('classification')}
          >
            🏷️ デバイス分類
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