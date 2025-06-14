import React, { useState, useEffect } from 'react'
import './DeviceClassificationBoard.css'

const HIERARCHY_LAYERS = [
  { id: 0, name: 'Internet Gateway', color: '#e74c3c', description: 'External internet connection point' },
  { id: 1, name: 'Firewall', color: '#e67e22', description: 'Security appliances' },
  { id: 2, name: 'Core Router', color: '#f39c12', description: 'Core network routing' },
  { id: 3, name: 'Distribution', color: '#3498db', description: 'Distribution layer switches' },
  { id: 4, name: 'Access', color: '#2ecc71', description: 'Access layer switches' },
  { id: 5, name: 'Server', color: '#95a5a6', description: 'End devices and servers' }
]

function DeviceClassificationBoard() {
  const [unclassifiedDevices, setUnclassifiedDevices] = useState([])
  const [classifiedDevices, setClassifiedDevices] = useState({}) // { layerId: [devices] }
  const [hierarchyLayers, setHierarchyLayers] = useState(HIERARCHY_LAYERS)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [draggedDevice, setDraggedDevice] = useState(null)
  const [dragOverLayer, setDragOverLayer] = useState(null)
  const [successMessage, setSuccessMessage] = useState(null)
  const [classificationRules, setClassificationRules] = useState([])
  const [showRuleManager, setShowRuleManager] = useState(false)
  const [editingRule, setEditingRule] = useState(null)
  const [selectedLayer, setSelectedLayer] = useState(null) // 選択された階層のサイドバー表示用
  const [showLayerManager, setShowLayerManager] = useState(false) // 階層管理表示用
  const [editingLayer, setEditingLayer] = useState(null) // 編集中の階層
  const [pagination, setPagination] = useState({ // ページネーション情報
    limit: 100,
    offset: 0,
    total: 0,
    currentPage: 1
  })

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      await Promise.all([
        loadUnclassifiedDevices(),
        loadClassifiedDevices(),
        loadHierarchyLayers(),
        loadClassificationRules()
      ])
    } catch (err) {
      setError('データの読み込みに失敗しました')
    } finally {
      setLoading(false)
    }
  }

  const loadUnclassifiedDevices = async (limit = 100, offset = 0) => {
    try {
      const response = await fetch(`/api/v1/classification/devices/unclassified?limit=${limit}&offset=${offset}`)
      if (!response.ok) throw new Error('Failed to load unclassified devices')
      const data = await response.json()
      setUnclassifiedDevices(data.devices || [])
      
      // ページネーション情報を更新
      if (data.total !== undefined) {
        setPagination(prev => ({
          ...prev,
          total: data.total,
          offset: data.offset || 0,
          limit: data.limit || 100,
          currentPage: Math.floor((data.offset || 0) / (data.limit || 100)) + 1
        }))
      }
    } catch (err) {
      console.error('Failed to load unclassified devices:', err)
    }
  }

  // ページネーション機能
  const goToPage = (page) => {
    const newOffset = (page - 1) * pagination.limit
    loadUnclassifiedDevices(pagination.limit, newOffset)
  }

  const nextPage = () => {
    if (pagination.currentPage * pagination.limit < pagination.total) {
      goToPage(pagination.currentPage + 1)
    }
  }

  const prevPage = () => {
    if (pagination.currentPage > 1) {
      goToPage(pagination.currentPage - 1)
    }
  }

  const loadClassifiedDevices = async () => {
    try {
      const response = await fetch('/api/v1/classification/devices/classified')
      if (!response.ok) throw new Error('Failed to load classified devices')
      const data = await response.json()
      
      // Group devices by layer
      const grouped = {}
      hierarchyLayers.forEach(layer => {
        grouped[layer.id] = []
      })
      
      data.classifications?.forEach(classification => {
        if (grouped[classification.layer]) {
          grouped[classification.layer].push({
            ...classification,
            device_id: classification.device_id
          })
        }
      })
      
      setClassifiedDevices(grouped)
    } catch (err) {
      console.error('Failed to load classified devices:', err)
      // Initialize empty groups
      const grouped = {}
      hierarchyLayers.forEach(layer => {
        grouped[layer.id] = []
      })
      setClassifiedDevices(grouped)
    }
  }

  const loadHierarchyLayers = async () => {
    try {
      const response = await fetch('/api/v1/classification/layers')
      if (!response.ok) throw new Error('Failed to load hierarchy layers')
      const data = await response.json()
      setHierarchyLayers(data.layers || HIERARCHY_LAYERS)
    } catch (err) {
      console.warn('Using default layers:', err.message)
      setHierarchyLayers(HIERARCHY_LAYERS)
    }
  }

  const saveHierarchyLayer = async (layer) => {
    try {
      const url = layer.id ? `/api/v1/classification/layers/${layer.id}` : '/api/v1/classification/layers'
      const method = layer.id ? 'PUT' : 'POST'
      
      // APIが期待するフィールドのみを送信
      const requestBody = {
        name: layer.name,
        description: layer.description,
        order: layer.order,
        color: layer.color
      }
      
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestBody)
      })
      
      if (!response.ok) {
        const errorData = await response.json()
        console.error('API Error:', errorData)
        throw new Error(errorData.detail || 'Failed to save hierarchy layer')
      }
      
      setSuccessMessage(layer.id ? '階層を更新しました' : '階層を作成しました')
      setTimeout(() => setSuccessMessage(null), 3000)
      
      await loadHierarchyLayers()
      setEditingLayer(null)
    } catch (err) {
      setError(err.message)
    }
  }

  const deleteHierarchyLayer = async (layerId) => {
    if (!window.confirm('この階層を削除しますか？関連する分類も削除される可能性があります。')) return
    
    try {
      const response = await fetch(`/api/v1/classification/layers/${layerId}`, {
        method: 'DELETE'
      })
      
      if (!response.ok) throw new Error('Failed to delete hierarchy layer')
      
      setSuccessMessage('階層を削除しました')
      setTimeout(() => setSuccessMessage(null), 3000)
      
      await loadHierarchyLayers()
    } catch (err) {
      setError(err.message)
    }
  }

  const handleCreateLayer = () => {
    setEditingLayer({
      name: '',
      description: '',
      order: hierarchyLayers.length,
      color: '#3498db'
    })
  }

  const loadClassificationRules = async () => {
    try {
      const response = await fetch('/api/v1/classification/rules')
      if (!response.ok) throw new Error('Failed to load classification rules')
      const data = await response.json()
      setClassificationRules(data.rules || [])
    } catch (err) {
      console.error('Failed to load classification rules:', err)
    }
  }

  const classifyDevice = async (deviceId, layer, deviceType) => {
    try {
      const response = await fetch('/api/v1/classification/devices', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ device_id: deviceId, layer, device_type: deviceType })
      })
      
      if (!response.ok) throw new Error('Failed to classify device')
      
      setSuccessMessage(`デバイス ${deviceId} を ${getLayerName(layer)} に分類しました`)
      setTimeout(() => setSuccessMessage(null), 3000)
      
      // Reload data to reflect changes
      await loadData()
    } catch (err) {
      setError(err.message)
    }
  }

  const getLayerName = (layerId) => {
    const layer = hierarchyLayers.find(l => l.id === layerId)
    return layer ? layer.name : `Layer ${layerId}`
  }

  const getDeviceTypeFromLayer = (layerId) => {
    const typeMap = {
      0: 'gateway',
      1: 'firewall', 
      2: 'router',
      3: 'switch',
      4: 'switch',
      5: 'server'
    }
    return typeMap[layerId] || 'unknown'
  }

  const handleDragStart = (e, device) => {
    setDraggedDevice(device)
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', device.id)
  }

  const handleDragOver = (e, layerId) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverLayer(layerId)
  }

  const handleDragLeave = (e) => {
    // Only clear if leaving the layer completely
    if (!e.currentTarget.contains(e.relatedTarget)) {
      setDragOverLayer(null)
    }
  }

  const handleDrop = async (e, layerId) => {
    e.preventDefault()
    setDragOverLayer(null)
    
    if (!draggedDevice) return

    const deviceType = getDeviceTypeFromLayer(layerId)
    await classifyDevice(draggedDevice.id, layerId, deviceType)
    setDraggedDevice(null)
  }

  const handleUnclassifyDevice = async (deviceId) => {
    try {
      const response = await fetch(`/api/v1/classification/devices/${deviceId}`, {
        method: 'DELETE'
      })
      
      if (!response.ok) throw new Error('Failed to unclassify device')
      
      setSuccessMessage(`デバイス ${deviceId} の分類を解除しました`)
      setTimeout(() => setSuccessMessage(null), 3000)
      
      await loadData()
    } catch (err) {
      setError(err.message)
    }
  }

  const applyClassificationRules = async () => {
    try {
      setLoading(true)
      const response = await fetch('/api/v1/classification/rules/apply', {
        method: 'POST'
      })
      
      if (!response.ok) throw new Error('Failed to apply classification rules')
      
      setSuccessMessage('分類ルールを適用しました')
      setTimeout(() => setSuccessMessage(null), 3000)
      
      await loadData()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const saveClassificationRule = async (rule) => {
    try {
      const url = rule.id ? `/api/v1/classification/rules/${rule.id}` : '/api/v1/classification/rules'
      const method = rule.id ? 'PUT' : 'POST'
      
      // APIが期待するフィールドのみを抽出
      const requestBody = {
        name: rule.name,
        description: rule.description,
        logic: rule.logic,
        conditions: rule.conditions,
        layer: rule.layer,
        device_type: rule.device_type,
        priority: rule.priority,
        is_active: rule.is_active
      }
      
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestBody)
      })
      
      if (!response.ok) throw new Error('Failed to save classification rule')
      
      setSuccessMessage(rule.id ? 'ルールを更新しました' : 'ルールを作成しました')
      setTimeout(() => setSuccessMessage(null), 3000)
      
      await loadClassificationRules()
      setEditingRule(null)
    } catch (err) {
      setError(err.message)
    }
  }

  const deleteClassificationRule = async (ruleId) => {
    if (!window.confirm('このルールを削除しますか？')) return
    
    try {
      const response = await fetch(`/api/v1/classification/rules/${ruleId}`, {
        method: 'DELETE'
      })
      
      if (!response.ok) throw new Error('Failed to delete classification rule')
      
      setSuccessMessage('ルールを削除しました')
      setTimeout(() => setSuccessMessage(null), 3000)
      
      await loadClassificationRules()
    } catch (err) {
      setError(err.message)
    }
  }

  const handleCreateRule = () => {
    setEditingRule({
      name: '',
      description: '',
      conditions: [
        {
          field: 'type',
          operator: 'equals',
          value: ''
        }
      ],
      logic: 'AND', // AND または OR
      layer: 4,
      device_type: 'switch',
      priority: 10,
      is_active: true
    })
  }

  const getDeviceIcon = (device) => {
    switch (device.type) {
      case 'server': return '🖥️'
      case 'router': return '🌐'
      case 'switch': return '🔀'
      case 'firewall': return '🛡️'
      case 'access': return '🔀'
      case 'core': return '🌐'
      case 'distribution': return '🔀'
      default: return '📱'
    }
  }

  if (loading) {
    return (
      <div className="classification-board loading">
        <div className="loading-spinner">
          <div className="spinner"></div>
          <p>データを読み込んでいます...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="classification-board">
      <div className="board-header">
        <h2>🏷️ デバイス分類ボード</h2>
        <div className="board-stats">
          <span className="stat-item">
            <span className="stat-label">未分類:</span>
            <span className="stat-value">{pagination.total}</span>
          </span>
          <span className="stat-item">
            <span className="stat-label">分類済み:</span>
            <span className="stat-value">
              {Object.values(classifiedDevices).reduce((sum, devices) => sum + devices.length, 0)}
            </span>
          </span>
          <span className="stat-item">
            <span className="stat-label">アクティブルール:</span>
            <span className="stat-value">{classificationRules.filter(r => r.is_active).length}</span>
          </span>
        </div>
        <div className="board-actions">
          <button 
            onClick={() => setShowLayerManager(!showLayerManager)} 
            className={`btn ${showLayerManager ? 'btn-warning' : 'btn-secondary'}`}
          >
            {showLayerManager ? '🏗️ 階層管理を閉じる' : '🏗️ 階層管理'}
          </button>
          <button 
            onClick={() => setShowRuleManager(!showRuleManager)} 
            className={`btn ${showRuleManager ? 'btn-warning' : 'btn-secondary'}`}
          >
            {showRuleManager ? '📝 ルール管理を閉じる' : '⚙️ ルール管理'}
          </button>
          <button 
            onClick={applyClassificationRules} 
            disabled={loading || classificationRules.filter(r => r.is_active).length === 0}
            className="btn btn-primary"
          >
            🤖 自動分類実行
          </button>
        </div>
      </div>

      {error && (
        <div className="alert alert-error">
          ❌ {error}
          <button onClick={() => setError(null)} className="alert-close">×</button>
        </div>
      )}

      {successMessage && (
        <div className="alert alert-success">
          ✅ {successMessage}
        </div>
      )}

      {/* ルール管理セクション */}
      {showRuleManager && (
        <div className="rule-manager">
          <div className="rule-manager-header">
            <h3>📋 分類ルール管理</h3>
            <button onClick={handleCreateRule} className="btn btn-primary">
              ➕ 新しいルール作成
            </button>
          </div>
          
          <div className="rules-table-container">
            {classificationRules.length > 0 ? (
              <table className="rules-table">
                <thead>
                  <tr>
                    <th>ルール名</th>
                    <th>説明</th>
                    <th>条件</th>
                    <th>分類先</th>
                    <th>優先度</th>
                    <th>ステータス</th>
                    <th>アクション</th>
                  </tr>
                </thead>
                <tbody>
                  {classificationRules.map(rule => (
                    <tr key={rule.id} className={`rule-row ${rule.is_active ? 'active' : 'inactive'}`}>
                      <td className="rule-name-cell">
                        <strong>{rule.name}</strong>
                      </td>
                      <td className="rule-description-cell">
                        {rule.description}
                      </td>
                      <td className="rule-condition-cell">
                        {rule.conditions ? (
                          <div className="condition-display">
                            {rule.conditions.map((condition, index) => (
                              <div key={index} className="condition-item">
                                {index > 0 && <span className="logic-text">{rule.logic}</span>}
                                <code>
                                  {condition.field} {condition.operator} "{condition.value}"
                                </code>
                              </div>
                            ))}
                          </div>
                        ) : (
                          <code>
                            {rule.field} {rule.operator} "{rule.value}"
                          </code>
                        )}
                      </td>
                      <td className="rule-target-cell">
                        <span className="layer-badge" style={{ backgroundColor: `var(--layer-${rule.layer}-color, #3498db)` }}>
                          Layer {rule.layer}
                        </span>
                        <span className="device-type-text">{rule.device_type}</span>
                      </td>
                      <td className="rule-priority-cell">
                        <span className="priority-badge">{rule.priority}</span>
                      </td>
                      <td className="rule-status-cell">
                        <span className={`status-badge ${rule.is_active ? 'active' : 'inactive'}`}>
                          {rule.is_active ? '有効' : '無効'}
                        </span>
                      </td>
                      <td className="rule-actions-cell">
                        <div className="rule-actions">
                          <button 
                            onClick={() => setEditingRule(rule)} 
                            className="btn btn-small btn-secondary"
                            title="編集"
                          >
                            ✏️
                          </button>
                          <button 
                            onClick={() => deleteClassificationRule(rule.id)} 
                            className="btn btn-small btn-danger"
                            title="削除"
                          >
                            🗑️
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="empty-rules">
                <p>📝 まだルールが作成されていません</p>
                <p>「新しいルール作成」ボタンでルールを追加してください</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* 階層管理セクション */}
      {showLayerManager && (
        <div className="layer-manager">
          <div className="layer-manager-header">
            <h3>🏗️ 階層レイヤー管理</h3>
            <button onClick={handleCreateLayer} className="btn btn-primary">
              ➕ 新しい階層作成
            </button>
          </div>
          
          <div className="layers-table-container">
            {hierarchyLayers.length > 0 ? (
              <table className="layers-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>階層名</th>
                    <th>説明</th>
                    <th>順序</th>
                    <th>色</th>
                    <th>デバイス数</th>
                    <th>アクション</th>
                  </tr>
                </thead>
                <tbody>
                  {hierarchyLayers.map(layer => (
                    <tr key={layer.id} className="layer-row">
                      <td className="layer-id-cell">
                        <strong>{layer.id}</strong>
                      </td>
                      <td className="layer-name-cell">
                        <strong>{layer.name}</strong>
                      </td>
                      <td className="layer-description-cell">
                        {layer.description}
                      </td>
                      <td className="layer-order-cell">
                        <span className="order-badge">{layer.order}</span>
                      </td>
                      <td className="layer-color-cell">
                        <div className="color-preview" style={{ backgroundColor: layer.color }}>
                          {layer.color}
                        </div>
                      </td>
                      <td className="layer-device-count-cell">
                        <span className="device-count-badge">
                          {classifiedDevices[layer.id]?.length || 0}台
                        </span>
                      </td>
                      <td className="layer-actions-cell">
                        <div className="layer-actions">
                          <button 
                            onClick={() => setEditingLayer(layer)} 
                            className="btn btn-small btn-secondary"
                            title="編集"
                          >
                            ✏️
                          </button>
                          <button 
                            onClick={() => deleteHierarchyLayer(layer.id)} 
                            className="btn btn-small btn-danger"
                            title="削除"
                            disabled={classifiedDevices[layer.id]?.length > 0}
                          >
                            🗑️
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="empty-layers">
                <p>🏗️ まだ階層が作成されていません</p>
                <p>「新しい階層作成」ボタンで階層を追加してください</p>
              </div>
            )}
          </div>
        </div>
      )}

      <div className="classification-layout">
        {/* 未分類デバイス一覧 */}
        <div className="unclassified-section">
          <div className="section-header">
            <h3>📦 未分類デバイス ({unclassifiedDevices.length}件表示 / 総{pagination.total}件)</h3>
            <p className="section-description">デバイスを右の階層にドラッグ&ドロップして分類してください</p>
          </div>
          <div className="device-pool">
            {unclassifiedDevices.map(device => (
              <div
                key={device.id}
                className="device-card unclassified"
                draggable
                onDragStart={(e) => handleDragStart(e, device)}
              >
                <div className="device-icon">{getDeviceIcon(device)}</div>
                <div className="device-info">
                  <div className="device-id">{device.id}</div>
                  <div className="device-type">{device.type}</div>
                  <div className="device-hardware">{device.hardware}</div>
                </div>
                <div className="drag-handle">⋮⋮</div>
              </div>
            ))}
            {unclassifiedDevices.length === 0 && pagination.total === 0 && (
              <div className="empty-state">
                <p>🎉 すべてのデバイスが分類されています！</p>
              </div>
            )}
          </div>
          
          {/* ページネーション */}
          {pagination.total > pagination.limit && (
            <div className="pagination-controls">
              <button 
                onClick={prevPage} 
                disabled={pagination.currentPage === 1}
                className="btn btn-small btn-secondary"
              >
                ← 前へ
              </button>
              <span className="pagination-info">
                {pagination.currentPage} / {Math.ceil(pagination.total / pagination.limit)} ページ
              </span>
              <button 
                onClick={nextPage} 
                disabled={pagination.currentPage * pagination.limit >= pagination.total}
                className="btn btn-small btn-secondary"
              >
                次へ →
              </button>
            </div>
          )}
        </div>

        {/* 階層レイヤー */}
        <div className="layers-section">
          <div className="layers-grid">
            {hierarchyLayers.sort((a, b) => a.order - b.order).map(layer => (
              <div
                key={layer.id}
                className={`layer-column ${dragOverLayer === layer.id ? 'drag-over' : ''} ${selectedLayer?.id === layer.id ? 'selected' : ''}`}
                style={{ '--layer-color': layer.color }}
                onDragOver={(e) => handleDragOver(e, layer.id)}
                onDragLeave={handleDragLeave}
                onDrop={(e) => handleDrop(e, layer.id)}
              >
                <div className="layer-header">
                  <div className="layer-indicator" style={{ backgroundColor: layer.color }}></div>
                  <div className="layer-info">
                    <h4 className="layer-name">Layer {layer.id}: {layer.name}</h4>
                    <p className="layer-description">{layer.description}</p>
                  </div>
                  <div className="device-count">
                    {classifiedDevices[layer.id]?.length || 0}台
                  </div>
                </div>
                
                <div className="layer-content">
                  <div className="layer-view-button">
                    <button 
                      className="btn btn-secondary view-devices-btn"
                      onClick={() => setSelectedLayer(selectedLayer?.id === layer.id ? null : { ...layer, devices: classifiedDevices[layer.id] || [] })}
                    >
                      {selectedLayer?.id === layer.id ? '📋 一覧を閉じる' : '📋 デバイス一覧'}
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ルール編集モーダル */}
      {editingRule && (
        <div className="modal-overlay">
          <div className="modal-content rule-modal">
            <div className="modal-header">
              <h3>{editingRule.id ? 'ルール編集' : 'ルール作成'}</h3>
              <button onClick={() => setEditingRule(null)} className="close-button">×</button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>ルール名 *</label>
                <input
                  type="text"
                  value={editingRule.name}
                  onChange={(e) => setEditingRule({ ...editingRule, name: e.target.value })}
                  className="form-input"
                  placeholder="例: Aristaアクセススイッチ自動分類"
                />
              </div>
              
              <div className="form-group">
                <label>説明</label>
                <textarea
                  value={editingRule.description}
                  onChange={(e) => setEditingRule({ ...editingRule, description: e.target.value })}
                  className="form-input"
                  placeholder="このルールの説明を入力してください"
                  rows="3"
                />
              </div>
              
              <div className="form-group">
                <div className="conditions-header">
                  <label>分類条件 *</label>
                  <div className="conditions-controls">
                    <select
                      value={editingRule.logic}
                      onChange={(e) => setEditingRule({ ...editingRule, logic: e.target.value })}
                      className="form-input logic-select"
                    >
                      <option value="AND">AND（すべての条件）</option>
                      <option value="OR">OR（いずれかの条件）</option>
                    </select>
                    <button
                      type="button"
                      onClick={() => {
                        const newConditions = [...editingRule.conditions, { field: 'type', operator: 'equals', value: '' }]
                        setEditingRule({ ...editingRule, conditions: newConditions })
                      }}
                      className="btn btn-small btn-secondary"
                    >
                      ➕ 条件追加
                    </button>
                  </div>
                </div>
                
                <div className="conditions-list">
                  {editingRule.conditions?.map((condition, index) => (
                    <div key={index} className="condition-row">
                      <div className="condition-index">
                        {index > 0 && <span className="logic-operator">{editingRule.logic}</span>}
                      </div>
                      <div className="condition-fields">
                        <select
                          value={condition.field}
                          onChange={(e) => {
                            const newConditions = [...editingRule.conditions]
                            newConditions[index] = { ...condition, field: e.target.value }
                            setEditingRule({ ...editingRule, conditions: newConditions })
                          }}
                          className="form-input"
                        >
                          <option value="type">デバイスタイプ</option>
                          <option value="hardware">ハードウェア</option>
                          <option value="name">デバイス名</option>
                          <option value="instance">インスタンス</option>
                        </select>
                        
                        <select
                          value={condition.operator}
                          onChange={(e) => {
                            const newConditions = [...editingRule.conditions]
                            newConditions[index] = { ...condition, operator: e.target.value }
                            setEditingRule({ ...editingRule, conditions: newConditions })
                          }}
                          className="form-input"
                        >
                          <option value="equals">完全一致</option>
                          <option value="contains">含む</option>
                          <option value="starts_with">で始まる</option>
                          <option value="ends_with">で終わる</option>
                          <option value="regex">正規表現</option>
                        </select>
                        
                        <input
                          type="text"
                          value={condition.value}
                          onChange={(e) => {
                            const newConditions = [...editingRule.conditions]
                            newConditions[index] = { ...condition, value: e.target.value }
                            setEditingRule({ ...editingRule, conditions: newConditions })
                          }}
                          className="form-input"
                          placeholder="値を入力"
                        />
                        
                        {editingRule.conditions.length > 1 && (
                          <button
                            type="button"
                            onClick={() => {
                              const newConditions = editingRule.conditions.filter((_, i) => i !== index)
                              setEditingRule({ ...editingRule, conditions: newConditions })
                            }}
                            className="btn btn-small btn-danger remove-condition-btn"
                            title="条件を削除"
                          >
                            🗑️
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
              
              <div className="form-row">
                <div className="form-group">
                  <label>分類先階層 *</label>
                  <select
                    value={editingRule.layer}
                    onChange={(e) => setEditingRule({ ...editingRule, layer: parseInt(e.target.value) })}
                    className="form-input"
                  >
                    {hierarchyLayers.map(layer => (
                      <option key={layer.id} value={layer.id}>
                        Layer {layer.id}: {layer.name}
                      </option>
                    ))}
                  </select>
                </div>
                
                <div className="form-group">
                  <label>デバイスタイプ *</label>
                  <select
                    value={editingRule.device_type}
                    onChange={(e) => setEditingRule({ ...editingRule, device_type: e.target.value })}
                    className="form-input"
                  >
                    <option value="switch">Switch</option>
                    <option value="router">Router</option>
                    <option value="server">Server</option>
                    <option value="firewall">Firewall</option>
                    <option value="gateway">Gateway</option>
                    <option value="access_point">Access Point</option>
                  </select>
                </div>
              </div>
              
              <div className="form-row">
                <div className="form-group">
                  <label>優先度</label>
                  <input
                    type="number"
                    value={editingRule.priority}
                    onChange={(e) => setEditingRule({ ...editingRule, priority: parseInt(e.target.value) || 0 })}
                    className="form-input"
                    min="0"
                    max="100"
                  />
                  <small>数値が大きいほど優先度が高くなります</small>
                </div>
                
                <div className="form-group">
                  <label>ステータス</label>
                  <div className="checkbox-group">
                    <label className="checkbox-label">
                      <input
                        type="checkbox"
                        checked={editingRule.is_active}
                        onChange={(e) => setEditingRule({ ...editingRule, is_active: e.target.checked })}
                      />
                      ルールを有効にする
                    </label>
                  </div>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button onClick={() => setEditingRule(null)} className="btn btn-secondary">
                キャンセル
              </button>
              <button 
                onClick={() => saveClassificationRule(editingRule)} 
                className="btn btn-primary"
                disabled={!editingRule.name || !editingRule.conditions?.some(c => c.field && c.operator && c.value)}
              >
                {editingRule.id ? '更新' : '作成'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 階層編集モーダル */}
      {editingLayer && (
        <div className="modal-overlay">
          <div className="modal-content layer-modal">
            <div className="modal-header">
              <h3>{editingLayer.id ? '階層編集' : '階層作成'}</h3>
              <button onClick={() => setEditingLayer(null)} className="close-button">×</button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>階層名 *</label>
                <input
                  type="text"
                  value={editingLayer.name}
                  onChange={(e) => setEditingLayer({ ...editingLayer, name: e.target.value })}
                  className="form-input"
                  placeholder="例: Core Router"
                />
              </div>
              
              <div className="form-group">
                <label>説明</label>
                <textarea
                  value={editingLayer.description}
                  onChange={(e) => setEditingLayer({ ...editingLayer, description: e.target.value })}
                  className="form-input"
                  placeholder="この階層の説明を入力してください"
                  rows="3"
                />
              </div>
              
              <div className="form-row">
                <div className="form-group">
                  <label>表示順序 *</label>
                  <input
                    type="number"
                    value={editingLayer.order}
                    onChange={(e) => setEditingLayer({ ...editingLayer, order: parseInt(e.target.value) || 0 })}
                    className="form-input"
                    min="0"
                    max="100"
                  />
                  <small>数値が小さいほど上に表示されます</small>
                </div>
                
                <div className="form-group">
                  <label>表示色 *</label>
                  <div className="color-input-group">
                    <input
                      type="color"
                      value={editingLayer.color}
                      onChange={(e) => setEditingLayer({ ...editingLayer, color: e.target.value })}
                      className="form-input color-input"
                    />
                    <input
                      type="text"
                      value={editingLayer.color}
                      onChange={(e) => setEditingLayer({ ...editingLayer, color: e.target.value })}
                      className="form-input color-text"
                      placeholder="#3498db"
                    />
                  </div>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button onClick={() => setEditingLayer(null)} className="btn btn-secondary">
                キャンセル
              </button>
              <button 
                onClick={() => saveHierarchyLayer(editingLayer)} 
                className="btn btn-primary"
                disabled={!editingLayer.name || !editingLayer.color}
              >
                {editingLayer.id ? '更新' : '作成'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 階層デバイス一覧サイドバー */}
      {selectedLayer && (
        <div className="layer-sidebar">
          <div className="sidebar-header">
            <div className="sidebar-title">
              <div className="layer-indicator" style={{ backgroundColor: selectedLayer.color }}></div>
              <h3>Layer {selectedLayer.id}: {selectedLayer.name}</h3>
            </div>
            <button 
              className="close-sidebar-btn"
              onClick={() => setSelectedLayer(null)}
              title="サイドバーを閉じる"
            >
              ×
            </button>
          </div>
          
          <div className="sidebar-content">
            <div className="sidebar-stats">
              <span className="stat-item">
                <span className="stat-label">分類済みデバイス:</span>
                <span className="stat-value">{selectedLayer.devices.length}台</span>
              </span>
            </div>
            
            <div className="sidebar-device-list">
              {selectedLayer.devices.map(classification => (
                <div key={classification.device_id} className="sidebar-device-card">
                  <div className="device-icon">{getDeviceIcon({ type: classification.device_type })}</div>
                  <div className="device-info">
                    <div className="device-id">{classification.device_id}</div>
                    <div className="device-type">{classification.device_type}</div>
                    <div className="device-meta">
                      {classification.is_manual ? '手動' : '自動'}分類
                    </div>
                  </div>
                  <button
                    className="unclassify-btn"
                    onClick={() => handleUnclassifyDevice(classification.device_id)}
                    title="分類を解除"
                  >
                    ×
                  </button>
                </div>
              ))}
              
              {selectedLayer.devices.length === 0 && (
                <div className="empty-sidebar">
                  <p>📋 この階層にはまだデバイスが分類されていません</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
      
      {selectedLayer && <div className="sidebar-overlay" onClick={() => setSelectedLayer(null)}></div>}
    </div>
  )
}

export default DeviceClassificationBoard