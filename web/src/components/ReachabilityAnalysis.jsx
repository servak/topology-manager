import React, { useState } from 'react'
import { FormContainer, FormGrid, FormGroup, FormInput, FormSelect, FormButton } from './common/FormStyles'

function ReachabilityAnalysis() {
  const [deviceId, setDeviceId] = useState('')
  const [algorithm, setAlgorithm] = useState('bfs')
  const [maxHops, setMaxHops] = useState(3)
  const [results, setResults] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!deviceId.trim()) return

    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`/api/devices/${encodeURIComponent(deviceId)}/reachable?algorithm=${algorithm}&max_hops=${maxHops}`)
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      setResults(data)
    } catch (err) {
      setError(err.message)
      console.error('Failed to fetch reachable devices:', err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="reachability-analysis">
      <div className="analysis-header">
        <h2>🔍 到達可能性分析</h2>
        <p>指定したデバイスから到達可能なデバイスを検索します</p>
      </div>

      <FormContainer onSubmit={handleSearch}>
        <FormGrid columns={4}>
          <FormGroup label="起点デバイス" htmlFor="deviceId">
            <FormInput
              id="deviceId"
              type="text"
              value={deviceId}
              onChange={(e) => setDeviceId(e.target.value)}
              placeholder="例: access-001, core-001, dist-001"
              required
            />
          </FormGroup>

          <FormGroup label="探索アルゴリズム" htmlFor="algorithm">
            <FormSelect 
              id="algorithm"
              value={algorithm} 
              onChange={(e) => setAlgorithm(e.target.value)}
            >
              <option value="bfs">BFS (幅優先探索)</option>
              <option value="dfs">DFS (深度優先探索)</option>
            </FormSelect>
          </FormGroup>

          <FormGroup label="最大ホップ数" htmlFor="maxHops">
            <FormInput
              id="maxHops"
              type="number"
              value={maxHops}
              onChange={(e) => setMaxHops(parseInt(e.target.value))}
              min="1"
              max="10"
            />
          </FormGroup>

          <FormGroup label=" " htmlFor="search">
            <FormButton type="submit" disabled={loading}>
              {loading ? '🔄 検索中...' : '🔍 検索開始'}
            </FormButton>
          </FormGroup>
        </FormGrid>
      </FormContainer>

      {error && (
        <div className="error-message">
          <div className="error-icon">⚠️</div>
          <div className="error-content">
            <h3>エラーが発生しました</h3>
            <p>{error}</p>
          </div>
        </div>
      )}

      {results && (
        <div className="results-section">
          <div className="results-header">
            <h3>📊 検索結果</h3>
            <div className="results-stats">
              <div className="stat-item">
                <span className="stat-label">アルゴリズム</span>
                <span className="stat-value">{results.algorithm.toUpperCase()}</span>
              </div>
              <div className="stat-item">
                <span className="stat-label">最大ホップ数</span>
                <span className="stat-value">{results.max_hops}</span>
              </div>
              <div className="stat-item">
                <span className="stat-label">発見デバイス数</span>
                <span className="stat-value highlight">{results.count}</span>
              </div>
            </div>
          </div>

          <div className="device-table-container">
            <table className="device-table">
              <thead>
                <tr>
                  <th>デバイス名</th>
                  <th>タイプ</th>
                  <th>ハードウェア</th>
                  <th>レイヤー</th>
                  <th>ステータス</th>
                  <th>場所</th>
                </tr>
              </thead>
              <tbody>
                {results.devices.map((device, index) => (
                  <tr key={device.id} className="device-row">
                    <td className="device-name-cell">
                      <div className="device-name-wrapper">
                        <span className="device-number">#{index + 1}</span>
                        <span className="device-name">{device.name}</span>
                      </div>
                    </td>
                    <td>
                      <span className={`device-type-badge ${device.type}`}>
                        {device.type}
                      </span>
                    </td>
                    <td className="hardware-cell">{device.hardware}</td>
                    <td className="layer-cell">
                      <span className="layer-badge">{device.layer}</span>
                    </td>
                    <td>
                      <span className={`status-badge ${device.status}`}>
                        {device.status}
                      </span>
                    </td>
                    <td className="location-cell">{device.location || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <style jsx>{`
        .reachability-analysis {
          padding: 24px;
          max-width: 1200px;
          margin: 0 auto;
        }

        .analysis-header {
          margin-bottom: 32px;
          text-align: center;
        }

        .analysis-header h2 {
          color: #2c3e50;
          margin-bottom: 8px;
          font-size: 2rem;
        }

        .analysis-header p {
          color: #666;
          font-size: 1.1rem;
        }

        .error-message {
          background: #fee;
          border: 2px solid #e74c3c;
          border-radius: 12px;
          padding: 20px;
          margin-bottom: 24px;
          display: flex;
          align-items: center;
          gap: 16px;
        }

        .error-icon {
          font-size: 2rem;
        }

        .error-content h3 {
          color: #c0392b;
          margin: 0 0 8px 0;
        }

        .error-content p {
          color: #e74c3c;
          margin: 0;
        }

        .results-section {
          background: white;
          border-radius: 12px;
          padding: 24px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .results-header {
          margin-bottom: 24px;
        }

        .results-header h3 {
          color: #2c3e50;
          margin-bottom: 16px;
        }

        .results-stats {
          display: flex;
          gap: 24px;
          flex-wrap: wrap;
        }

        .stat-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 12px 20px;
          background: #f8f9fa;
          border-radius: 8px;
          min-width: 120px;
        }

        .stat-label {
          font-size: 0.85rem;
          color: #666;
          margin-bottom: 4px;
        }

        .stat-value {
          font-size: 1.2rem;
          font-weight: 700;
          color: #2c3e50;
        }

        .stat-value.highlight {
          color: #e67e22;
        }

        .device-table-container {
          background: white;
          border-radius: 12px;
          overflow: hidden;
          box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .device-table {
          width: 100%;
          border-collapse: collapse;
          font-size: 0.9rem;
        }

        .device-table th {
          background: #f8f9fa;
          padding: 16px 12px;
          text-align: left;
          font-weight: 600;
          color: #2c3e50;
          border-bottom: 2px solid #e9ecef;
          font-size: 0.85rem;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .device-table td {
          padding: 12px;
          border-bottom: 1px solid #f1f3f4;
          vertical-align: middle;
        }

        .device-row:hover {
          background: #f8f9fa;
        }

        .device-name-cell {
          min-width: 180px;
        }

        .device-name-wrapper {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .device-number {
          background: #3498db;
          color: white;
          border-radius: 50%;
          width: 28px;
          height: 28px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-weight: 700;
          font-size: 0.75rem;
        }

        .device-name {
          font-weight: 700;
          color: #2c3e50;
        }

        .device-type-badge {
          padding: 4px 12px;
          border-radius: 20px;
          font-size: 0.75rem;
          font-weight: 600;
          text-transform: uppercase;
          white-space: nowrap;
        }

        .device-type-badge.core {
          background: #e74c3c;
          color: white;
        }

        .device-type-badge.distribution {
          background: #f39c12;
          color: white;
        }

        .device-type-badge.access {
          background: #27ae60;
          color: white;
        }

        .device-type-badge.server {
          background: #9b59b6;
          color: white;
        }

        .hardware-cell {
          max-width: 200px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          color: #666;
        }

        .ip-cell {
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 0.85rem;
          color: #2c3e50;
        }

        .layer-cell {
          text-align: center;
        }

        .layer-badge {
          background: #ecf0f1;
          color: #2c3e50;
          padding: 4px 8px;
          border-radius: 16px;
          font-size: 0.75rem;
          font-weight: 600;
          text-transform: uppercase;
        }

        .location-cell {
          color: #666;
          font-size: 0.85rem;
        }

        .status-badge {
          padding: 2px 8px;
          border-radius: 12px;
          font-size: 0.8rem;
          font-weight: 600;
          text-transform: uppercase;
        }

        .status-badge.active {
          background: #d5f4e6;
          color: #27ae60;
        }

        .status-badge.inactive {
          background: #fadbd8;
          color: #e74c3c;
        }

        .status-badge.unknown {
          background: #fef9e7;
          color: #f39c12;
        }
      `}</style>
    </div>
  )
}

export default ReachabilityAnalysis