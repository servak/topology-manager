import React, { useState, useEffect } from 'react'
import './DeviceClassification.css'

const HIERARCHY_LAYERS = [
  { id: 0, name: 'Internet Gateway', color: '#e74c3c', description: 'External internet connection point' },
  { id: 1, name: 'Firewall', color: '#e67e22', description: 'Security appliances' },
  { id: 2, name: 'Core Router', color: '#f39c12', description: 'Core network routing' },
  { id: 3, name: 'Distribution', color: '#3498db', description: 'Distribution layer switches' },
  { id: 4, name: 'Access', color: '#2ecc71', description: 'Access layer switches' },
  { id: 5, name: 'Server', color: '#95a5a6', description: 'End devices and servers' }
]

function DeviceClassification() {
  const [unclassifiedDevices, setUnclassifiedDevices] = useState([])
  const [classificationRules, setClassificationRules] = useState([])
  const [suggestions, setSuggestions] = useState([])
  const [hierarchyLayers, setHierarchyLayers] = useState(HIERARCHY_LAYERS)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [selectedDevices, setSelectedDevices] = useState(new Set())
  const [draggedDevice, setDraggedDevice] = useState(null)
  const [editingLayers, setEditingLayers] = useState(false)
  const [editingLayer, setEditingLayer] = useState(null)
  const [classificationStats, setClassificationStats] = useState(null)

  useEffect(() => {
    loadUnclassifiedDevices()
    loadClassificationRules()
    loadSuggestions()
    loadHierarchyLayers()
    loadClassificationStats()
  }, [])

  const loadUnclassifiedDevices = async () => {
    try {
      const response = await fetch('/api/v1/classification/devices/unclassified')
      if (!response.ok) throw new Error('Failed to load unclassified devices')
      const data = await response.json()
      setUnclassifiedDevices(data.devices || [])
    } catch (err) {
      setError(err.message)
    }
  }

  const loadClassificationRules = async () => {
    try {
      const response = await fetch('/api/v1/classification/rules')
      if (!response.ok) throw new Error('Failed to load classification rules')
      const data = await response.json()
      setClassificationRules(data.rules || [])
    } catch (err) {
      setError(err.message)
    }
  }

  const loadSuggestions = async () => {
    try {
      const response = await fetch('/api/v1/classification/suggestions')
      if (!response.ok) throw new Error('Failed to load suggestions')
      const data = await response.json()
      setSuggestions(data.suggestions || [])
    } catch (err) {
      setError(err.message)
    }
  }

  const loadHierarchyLayers = async () => {
    try {
      const response = await fetch('/api/v1/classification/layers')
      if (!response.ok) throw new Error('Failed to load hierarchy layers')
      const data = await response.json()
      setHierarchyLayers(data.layers || HIERARCHY_LAYERS)
    } catch (err) {
      // フォールバックとしてデフォルト階層を使用
      setHierarchyLayers(HIERARCHY_LAYERS)
      console.warn('Failed to load custom layers, using defaults:', err.message)
    }
  }

  const loadClassificationStats = async () => {
    try {
      const response = await fetch('/api/v1/classification/stats')
      if (!response.ok) throw new Error('Failed to load classification statistics')
      const data = await response.json()
      setClassificationStats(data)
    } catch (err) {
      console.warn('Failed to load classification statistics:', err.message)
    }
  }

  const saveHierarchyLayer = async (layer) => {
    setLoading(true)
    try {
      const url = layer.id !== undefined 
        ? `/api/v1/classification/layers/${layer.id}` 
        : '/api/v1/classification/layers'
      
      const method = layer.id !== undefined ? 'PUT' : 'POST'
      
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(layer)
      })
      
      if (!response.ok) throw new Error('Failed to save layer')
      
      await loadHierarchyLayers()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const deleteHierarchyLayer = async (layerId) => {
    setLoading(true)
    try {
      const response = await fetch(`/api/v1/classification/layers/${layerId}`, {
        method: 'DELETE'
      })
      
      if (!response.ok) throw new Error('Failed to delete layer')
      
      await loadHierarchyLayers()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const classifyDevice = async (deviceId, layer, deviceType) => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/classification/devices', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ device_id: deviceId, layer, device_type: deviceType })
      })
      
      if (!response.ok) throw new Error('Failed to classify device')
      
      // 未分類デバイス一覧を更新
      await loadUnclassifiedDevices()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const createRule = async (ruleData) => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/classification/rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(ruleData)
      })
      
      if (!response.ok) throw new Error('Failed to create rule')
      
      await loadClassificationRules()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const applyRules = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/classification/rules/apply', {
        method: 'POST'
      })
      
      if (!response.ok) throw new Error('Failed to apply rules')
      
      await loadUnclassifiedDevices()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const generateSuggestions = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/classification/suggestions/generate', {
        method: 'POST'
      })
      
      if (!response.ok) throw new Error('Failed to generate suggestions')
      
      await loadSuggestions()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleSuggestion = async (suggestionId, action) => {
    setLoading(true)
    try {
      const response = await fetch(`/api/v1/classification/suggestions/${suggestionId}/action`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action })
      })
      
      if (!response.ok) throw new Error(`Failed to ${action} suggestion`)
      
      await loadSuggestions()
      await loadClassificationRules()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleDragStart = (e, device) => {
    setDraggedDevice(device)
    e.dataTransfer.effectAllowed = 'move'
  }

  const handleDragOver = (e) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }

  const handleDrop = async (e, layer) => {
    e.preventDefault()
    if (!draggedDevice) return

    // レイヤー情報から適切なdevice_typeを推定
    const deviceType = getDeviceTypeFromLayer(layer.id)
    await classifyDevice(draggedDevice.id, layer.id, deviceType)
    setDraggedDevice(null)
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

  const handleDeviceSelect = (deviceId) => {
    const newSelected = new Set(selectedDevices)
    if (newSelected.has(deviceId)) {
      newSelected.delete(deviceId)
    } else {
      newSelected.add(deviceId)
    }
    setSelectedDevices(newSelected)
  }

  const handleLayerEdit = (layer) => {
    setEditingLayer({ ...layer })
  }

  const handleLayerSave = async () => {
    if (!editingLayer) return
    
    // バリデーション
    if (!editingLayer.name || !editingLayer.name.trim()) {
      setError('階層名は必須です')
      return
    }
    
    if (!editingLayer.color || !editingLayer.color.match(/^#[0-9A-Fa-f]{6}$/)) {
      setError('有効なカラーコードを入力してください')
      return
    }
    
    try {
      await saveHierarchyLayer(editingLayer)
      setEditingLayer(null)
      setError(null)
    } catch (err) {
      // エラーは saveHierarchyLayer 内で設定されるため、ここでは何もしない
    }
  }

  const handleLayerCancel = () => {
    setEditingLayer(null)
  }

  const handleLayerDelete = async (layerId) => {
    const layer = hierarchyLayers.find(l => l.id === layerId)
    const layerName = layer ? layer.name : `ID:${layerId}`
    
    if (window.confirm(`階層「${layerName}」を削除しますか？\n\n注意: この階層に分類されているデバイスは未分類に戻ります。`)) {
      try {
        await deleteHierarchyLayer(layerId)
        // 削除成功後、未分類デバイス一覧も更新
        await loadUnclassifiedDevices()
      } catch (err) {
        // エラーは deleteHierarchyLayer 内で設定される
      }
    }
  }

  const handleAddLayer = () => {
    const maxId = Math.max(...hierarchyLayers.map(l => l.id), -1)
    const maxOrder = Math.max(...hierarchyLayers.map(l => l.order || l.id), -1)
    setEditingLayer({
      name: '新しい階層',
      color: '#95a5a6',
      description: '新しい階層の説明',
      order: maxOrder + 1
    })
  }

  const moveLayer = async (fromIndex, toIndex) => {
    if (fromIndex === toIndex) return
    
    const newLayers = [...hierarchyLayers]
    const [movedLayer] = newLayers.splice(fromIndex, 1)
    newLayers.splice(toIndex, 0, movedLayer)
    
    // order を更新
    const updatedLayers = newLayers.map((layer, index) => ({
      ...layer,
      order: index
    }))
    
    setHierarchyLayers(updatedLayers)
    
    // サーバーに保存（バッチ更新）
    try {
      setLoading(true)
      for (const layer of updatedLayers) {
        await saveHierarchyLayer(layer)
      }
    } catch (err) {
      setError('階層の並び替えに失敗しました: ' + err.message)
      // エラーの場合は元に戻す
      await loadHierarchyLayers()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="device-classification">
      <div className="classification-header">
        <div className="header-left">
          <h2>🏷️ デバイス分類管理</h2>
          {classificationStats && (
            <div className="classification-stats">
              <span>総デバイス数: {classificationStats.total_devices || 0}</span>
              <span>分類済み: {classificationStats.classified_devices || 0}</span>
              <span>未分類: {classificationStats.unclassified_devices || 0}</span>
              <span>適用可能ルール: {classificationStats.active_rules || 0}</span>
            </div>
          )}
        </div>
        <div className="classification-actions">
          <button onClick={applyRules} disabled={loading} className="btn btn-primary">
            ルール適用
          </button>
          <button onClick={generateSuggestions} disabled={loading} className="btn btn-secondary">
            ルール提案生成
          </button>
          <button 
            onClick={() => setEditingLayers(!editingLayers)} 
            className={`btn ${editingLayers ? 'btn-warning' : 'btn-secondary'}`}
          >
            {editingLayers ? '📝 編集完了' : '⚙️ 階層編集'}
          </button>
        </div>
      </div>

      {error && (
        <div className="error-message">
          <p>❌ {error}</p>
          <button onClick={() => setError(null)} className="btn btn-small">
            閉じる
          </button>
        </div>
      )}

      {loading && (
        <div className="loading-message">
          <p>🔄 処理中...</p>
        </div>
      )}

      <div className="classification-content">
        {/* 未分類デバイス一覧 */}
        <div className="unclassified-section">
          <h3>未分類デバイス ({unclassifiedDevices.length})</h3>
          <div className="device-list">
            {unclassifiedDevices.map(device => (
              <div
                key={device.id}
                className={`device-item ${selectedDevices.has(device.id) ? 'selected' : ''}`}
                draggable
                onDragStart={(e) => handleDragStart(e, device)}
                onClick={() => handleDeviceSelect(device.id)}
              >
                <div className="device-info">
                  <span className="device-id">{device.id}</span>
                  <span className="device-type">{device.type}</span>
                  <span className="device-hardware">{device.hardware}</span>
                </div>
                <div className="device-drag-handle">⋮⋮</div>
              </div>
            ))}
          </div>
        </div>

        {/* 階層レイヤー */}
        <div className="layers-section">
          <div className="layers-header">
            <h3>ネットワーク階層</h3>
            {editingLayers && (
              <button onClick={handleAddLayer} className="btn btn-primary btn-small">
                ➕ 階層追加
              </button>
            )}
          </div>
          <div className="hierarchy-layers">
            {hierarchyLayers.sort((a, b) => a.order - b.order).map((layer, index) => (
              <div
                key={layer.id}
                className={`layer-item ${editingLayers ? 'editing-mode' : ''}`}
                style={{ '--layer-color': layer.color }}
                onDragOver={!editingLayers ? handleDragOver : undefined}
                onDrop={!editingLayers ? (e) => handleDrop(e, layer) : undefined}
              >
                <div className="layer-header">
                  <span className="layer-name">Layer {layer.id}: {layer.name}</span>
                  <div className="layer-color" style={{ backgroundColor: layer.color }}></div>
                  {editingLayers && (
                    <div className="layer-actions">
                      <button 
                        onClick={() => handleLayerEdit(layer)} 
                        className="btn btn-small btn-secondary"
                        title="編集"
                        disabled={loading}
                      >
                        ✏️
                      </button>
                      <button 
                        onClick={() => handleLayerDelete(layer.id)} 
                        className="btn btn-small btn-danger"
                        title="削除"
                        disabled={loading || hierarchyLayers.length <= 1}
                      >
                        🗑️
                      </button>
                      {index > 0 && (
                        <button 
                          onClick={() => moveLayer(index, index - 1)} 
                          className="btn btn-small btn-secondary"
                          title="上に移動"
                          disabled={loading}
                        >
                          ⬆️
                        </button>
                      )}
                      {index < hierarchyLayers.length - 1 && (
                        <button 
                          onClick={() => moveLayer(index, index + 1)} 
                          className="btn btn-small btn-secondary"
                          title="下に移動"
                          disabled={loading}
                        >
                          ⬇️
                        </button>
                      )}
                    </div>
                  )}
                </div>
                <p className="layer-description">{layer.description}</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* 分類ルール一覧 */}
      <div className="rules-section">
        <h3>分類ルール ({classificationRules.length})</h3>
        <div className="rules-list">
          {classificationRules.map(rule => (
            <div key={rule.id} className="rule-item">
              <div className="rule-info">
                <span className="rule-name">{rule.name}</span>
                <span className="rule-condition">
                  {rule.field} {rule.operator} "{rule.value}"
                </span>
                <span className="rule-target">
                  → Layer {rule.layer} ({rule.device_type})
                </span>
              </div>
              <div className="rule-status">
                <span className={`status ${rule.is_active ? 'active' : 'inactive'}`}>
                  {rule.is_active ? '有効' : '無効'}
                </span>
                <span className="priority">優先度: {rule.priority}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ルール提案 */}
      {suggestions.length > 0 && (
        <div className="suggestions-section">
          <h3>ルール提案 ({suggestions.length})</h3>
          <div className="suggestions-list">
            {suggestions.map(suggestion => (
              <div key={suggestion.id} className="suggestion-item">
                <div className="suggestion-header">
                  <span className="suggestion-name">{suggestion.rule.name}</span>
                  <span className="confidence">信頼度: {Math.round(suggestion.confidence * 100)}%</span>
                </div>
                <div className="suggestion-rule">
                  {suggestion.rule.field} {suggestion.rule.operator} "{suggestion.rule.value}"
                  → Layer {suggestion.rule.layer} ({suggestion.rule.device_type})
                </div>
                <div className="suggestion-devices">
                  適用対象: {suggestion.affected_devices?.length || 0}台
                  {suggestion.based_on_devices?.length > 0 && (
                    <span className="based-on">
                      (based on: {suggestion.based_on_devices.slice(0, 3).join(', ')}
                      {suggestion.based_on_devices.length > 3 && '...'})
                    </span>
                  )}
                </div>
                <div className="suggestion-actions">
                  <button
                    onClick={() => handleSuggestion(suggestion.id, 'accept')}
                    className="btn btn-primary btn-small"
                    disabled={loading}
                  >
                    ✓ 承認
                  </button>
                  <button
                    onClick={() => handleSuggestion(suggestion.id, 'reject')}
                    className="btn btn-secondary btn-small"
                    disabled={loading}
                  >
                    ✗ 拒否
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 階層編集モーダル */}
      {editingLayer && (
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-header">
              <h3>{editingLayer.id !== undefined ? '階層編集' : '階層追加'}</h3>
              <button onClick={handleLayerCancel} className="close-button">×</button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>階層ID</label>
                <input
                  type="number"
                  value={editingLayer.id !== undefined ? editingLayer.id : ''}
                  onChange={(e) => setEditingLayer({
                    ...editingLayer,
                    id: parseInt(e.target.value) || 0
                  })}
                  disabled={editingLayer.id !== undefined && editingLayer.id >= 0}
                  className="form-input"
                  min="0"
                  max="10"
                />
              </div>
              <div className="form-group">
                <label>階層名 *</label>
                <input
                  type="text"
                  value={editingLayer.name || ''}
                  onChange={(e) => setEditingLayer({
                    ...editingLayer,
                    name: e.target.value
                  })}
                  className="form-input"
                  placeholder="例: Core Router"
                  required
                  maxLength="50"
                />
              </div>
              <div className="form-group">
                <label>説明</label>
                <textarea
                  value={editingLayer.description || ''}
                  onChange={(e) => setEditingLayer({
                    ...editingLayer,
                    description: e.target.value
                  })}
                  className="form-input"
                  placeholder="例: コアネットワークルーティング"
                  rows="3"
                  maxLength="200"
                />
              </div>
              <div className="form-group">
                <label>カラー *</label>
                <div className="color-picker">
                  <input
                    type="color"
                    value={editingLayer.color || '#95a5a6'}
                    onChange={(e) => setEditingLayer({
                      ...editingLayer,
                      color: e.target.value
                    })}
                    className="color-input"
                  />
                  <input
                    type="text"
                    value={editingLayer.color || '#95a5a6'}
                    onChange={(e) => {
                      const color = e.target.value
                      if (color.match(/^#[0-9A-Fa-f]{0,6}$/)) {
                        setEditingLayer({
                          ...editingLayer,
                          color: color
                        })
                      }
                    }}
                    className="form-input color-text"
                    placeholder="#95a5a6"
                    pattern="^#[0-9A-Fa-f]{6}$"
                    maxLength="7"
                  />
                </div>
              </div>
              <div className="color-presets">
                <span>プリセット:</span>
                {[
                  '#e74c3c', '#e67e22', '#f39c12', '#f1c40f',
                  '#2ecc71', '#27ae60', '#3498db', '#2980b9', 
                  '#9b59b6', '#8e44ad', '#95a5a6', '#34495e',
                  '#16a085', '#d35400', '#c0392b', '#7f8c8d'
                ].map(color => (
                  <button
                    key={color}
                    className="color-preset"
                    style={{ backgroundColor: color }}
                    onClick={() => setEditingLayer({ ...editingLayer, color })}
                    title={color}
                  />
                ))}
              </div>
            </div>
            <div className="modal-footer">
              <button onClick={handleLayerCancel} className="btn btn-secondary">
                キャンセル
              </button>
              <button 
                onClick={handleLayerSave} 
                className="btn btn-primary" 
                disabled={loading || !editingLayer?.name?.trim() || !editingLayer?.color?.match(/^#[0-9A-Fa-f]{6}$/)}
              >
                {loading ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default DeviceClassification