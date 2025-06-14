import React, { useState, useEffect } from 'react'
import './HierarchicalTopology.css'

const LAYER_COLORS = {
  0: '#e74c3c', // Internet Gateway
  1: '#e67e22', // Firewall
  2: '#f39c12', // Core Router
  3: '#3498db', // Distribution
  4: '#2ecc71', // Access
  5: '#95a5a6', // Server
}

const LAYER_NAMES = {
  0: 'Internet Gateway',
  1: 'Firewall',
  2: 'Core Router', 
  3: 'Distribution',
  4: 'Access',
  5: 'Server',
}

function HierarchicalTopology({ topology, onDeviceSelect, selectedDevice }) {
  const [expandedLayers, setExpandedLayers] = useState(new Set([0, 1, 2])) // デフォルトで上位階層を展開
  const [expandedDevices, setExpandedDevices] = useState(new Set())
  const [hierarchicalData, setHierarchicalData] = useState(null)

  useEffect(() => {
    if (topology) {
      buildHierarchicalStructure(topology)
    }
  }, [topology])

  const buildHierarchicalStructure = (topology) => {
    // ノードを階層別に分類
    const layers = {}
    const nodeMap = {}
    
    topology.nodes.forEach(node => {
      const layer = node.layer || 5 // 未分類は最下層として扱う
      if (!layers[layer]) {
        layers[layer] = []
      }
      layers[layer].push(node)
      nodeMap[node.id] = node
    })

    // エッジから隣接関係を構築
    const adjacency = {}
    topology.edges.forEach(edge => {
      if (!adjacency[edge.source]) {
        adjacency[edge.source] = []
      }
      if (!adjacency[edge.target]) {
        adjacency[edge.target] = []
      }
      adjacency[edge.source].push({
        device: edge.target,
        port: edge.local_port,
        remotePort: edge.remote_port,
        status: edge.status
      })
      adjacency[edge.target].push({
        device: edge.source,
        port: edge.remote_port,
        remotePort: edge.local_port,
        status: edge.status
      })
    })

    // 階層構造の構築
    const hierarchical = {}
    Object.keys(layers).sort((a, b) => parseInt(a) - parseInt(b)).forEach(layer => {
      hierarchical[layer] = {
        name: LAYER_NAMES[layer] || `Layer ${layer}`,
        color: LAYER_COLORS[layer] || '#95a5a6',
        devices: layers[layer].map(device => ({
          ...device,
          connections: adjacency[device.id] || []
        }))
      }
    })

    setHierarchicalData(hierarchical)
  }

  const toggleLayerExpansion = (layerId) => {
    const newExpanded = new Set(expandedLayers)
    if (newExpanded.has(parseInt(layerId))) {
      newExpanded.delete(parseInt(layerId))
    } else {
      newExpanded.add(parseInt(layerId))
    }
    setExpandedLayers(newExpanded)
  }

  const toggleDeviceExpansion = (deviceId) => {
    const newExpanded = new Set(expandedDevices)
    if (newExpanded.has(deviceId)) {
      newExpanded.delete(deviceId)
    } else {
      newExpanded.add(deviceId)
    }
    setExpandedDevices(newExpanded)
  }

  const handleDeviceClick = (device) => {
    if (onDeviceSelect) {
      onDeviceSelect(device.id)
    }
  }

  const handleConnectionClick = (connectionDeviceId) => {
    if (onDeviceSelect) {
      onDeviceSelect(connectionDeviceId)
    }
  }

  const getDeviceIcon = (type) => {
    const icons = {
      'gateway': '🌐',
      'firewall': '🛡️',
      'router': '📡',
      'switch': '🔗',
      'server': '💻',
      'unknown': '❓'
    }
    return icons[type] || icons.unknown
  }

  const getStatusIcon = (status) => {
    return status === 'up' || status === 'active' ? '🟢' : '🔴'
  }

  if (!hierarchicalData) {
    return (
      <div className="hierarchical-topology">
        <div className="no-data">
          <p>階層データがありません</p>
        </div>
      </div>
    )
  }

  return (
    <div className="hierarchical-topology">
      <div className="hierarchy-header">
        <h3>🏗️ 階層トポロジー表示</h3>
        <div className="topology-info">
          <span>階層数: {Object.keys(hierarchicalData).length}</span>
          <span>総デバイス数: {Object.values(hierarchicalData).reduce((sum, layer) => sum + layer.devices.length, 0)}</span>
        </div>
      </div>

      <div className="hierarchy-tree">
        {Object.entries(hierarchicalData).map(([layerId, layer]) => (
          <div key={layerId} className="layer-section">
            <div 
              className="layer-header"
              onClick={() => toggleLayerExpansion(layerId)}
              style={{ '--layer-color': layer.color }}
            >
              <span className={`expand-icon ${expandedLayers.has(parseInt(layerId)) ? 'expanded' : ''}`}>
                ▶
              </span>
              <div className="layer-indicator" style={{ backgroundColor: layer.color }}></div>
              <span className="layer-name">
                Layer {layerId}: {layer.name}
              </span>
              <span className="device-count">({layer.devices.length})</span>
            </div>

            {expandedLayers.has(parseInt(layerId)) && (
              <div className="layer-content">
                <div className="devices-list">
                  {layer.devices.map((device) => (
                    <div key={device.id} className="device-section">
                      <div 
                        className={`device-item ${selectedDevice === device.id ? 'selected' : ''} ${device.is_root ? 'root' : ''}`}
                        onClick={() => handleDeviceClick(device)}
                      >
                        <div className="device-main">
                          <span className="device-icon">{getDeviceIcon(device.type)}</span>
                          <span className="device-name">{device.id}</span>
                          <span className="device-type">{device.type}</span>
                          <span className="device-status">{getStatusIcon(device.status)}</span>
                          {device.hardware && (
                            <span className="device-hardware">{device.hardware}</span>
                          )}
                          {device.is_root && <span className="root-badge">ROOT</span>}
                        </div>
                        
                        {device.connections.length > 0 && (
                          <button
                            className="connections-toggle"
                            onClick={(e) => {
                              e.stopPropagation()
                              toggleDeviceExpansion(device.id)
                            }}
                          >
                            <span className={`expand-icon ${expandedDevices.has(device.id) ? 'expanded' : ''}`}>
                              ▶
                            </span>
                            <span>接続先 ({device.connections.length})</span>
                          </button>
                        )}
                      </div>

                      {expandedDevices.has(device.id) && device.connections.length > 0 && (
                        <div className="connections-list">
                          {device.connections.map((connection, index) => {
                            const connectedDevice = topology.nodes.find(n => n.id === connection.device)
                            return (
                              <div 
                                key={`${device.id}-${connection.device}-${index}`}
                                className="connection-item"
                                onClick={() => handleConnectionClick(connection.device)}
                              >
                                <div className="connection-line">
                                  <span className="port-info">
                                    {connection.port} ↔ {connection.remotePort}
                                  </span>
                                  <span className="connection-status">
                                    {getStatusIcon(connection.status)}
                                  </span>
                                </div>
                                <div className="connected-device">
                                  <span className="device-icon">
                                    {getDeviceIcon(connectedDevice?.type || 'unknown')}
                                  </span>
                                  <span className="device-name">{connection.device}</span>
                                  {connectedDevice && (
                                    <>
                                      <span className="device-layer">
                                        L{connectedDevice.layer}
                                      </span>
                                      <span className="device-type">{connectedDevice.type}</span>
                                    </>
                                  )}
                                </div>
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      <div className="hierarchy-legend">
        <h4>凡例</h4>
        <div className="legend-items">
          <div className="legend-item">
            <span>🟢</span> <span>オンライン</span>
          </div>
          <div className="legend-item">
            <span>🔴</span> <span>オフライン</span>
          </div>
          <div className="legend-item">
            <span className="root-badge-small">ROOT</span> <span>ルートデバイス</span>
          </div>
          <div className="legend-item">
            <span>L0-L5</span> <span>ネットワーク階層</span>
          </div>
        </div>
      </div>
    </div>
  )
}

export default HierarchicalTopology