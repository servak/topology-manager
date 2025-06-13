import React, { useState } from 'react'
import { GCPCard, GCPFormField, GCPInput, GCPSelect, GCPButton } from './common/GCPStyles'

function PathAnalysis() {
  const [fromDevice, setFromDevice] = useState('')
  const [toDevice, setToDevice] = useState('')
  const [algorithm, setAlgorithm] = useState('dijkstra')
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!fromDevice.trim() || !toDevice.trim()) return

    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`/api/path/${encodeURIComponent(fromDevice)}/${encodeURIComponent(toDevice)}?algorithm=${algorithm}`)
      
      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      setResult(data)
    } catch (err) {
      setError(err.message)
      console.error('Failed to find path:', err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="path-analysis">
      <GCPCard 
        title="最短パス分析"
        subtitle="Dijkstra アルゴリズムを使用して2つのデバイス間の最短パスを検索します"
      >
        <form onSubmit={handleSearch}>
          <div className="form-grid">
            <GCPFormField 
              label="起点デバイス" 
              required
              helperText="パス検索の開始点となるデバイス"
            >
              <GCPInput
                type="text"
                value={fromDevice}
                onChange={(e) => setFromDevice(e.target.value)}
                placeholder="例: access-001"
                disabled={loading}
              />
            </GCPFormField>

            <GCPFormField 
              label="終点デバイス" 
              required
              helperText="パス検索の終了点となるデバイス"
            >
              <GCPInput
                type="text"
                value={toDevice}
                onChange={(e) => setToDevice(e.target.value)}
                placeholder="例: server-050"
                disabled={loading}
              />
            </GCPFormField>

            <GCPFormField 
              label="パス検索アルゴリズム"
              helperText="使用するパス検索アルゴリズムを選択"
            >
              <GCPSelect 
                value={algorithm} 
                onChange={(e) => setAlgorithm(e.target.value)}
                disabled={loading}
              >
                <option value="dijkstra">Dijkstra (最短パス)</option>
                <option value="k_shortest">K-Shortest Path</option>
              </GCPSelect>
            </GCPFormField>
          </div>

          <div className="form-actions">
            <GCPButton 
              type="submit" 
              disabled={loading || !fromDevice.trim() || !toDevice.trim()}
              loading={loading}
              variant="primary"
            >
              {loading ? '検索中...' : '最短パスを検索'}
            </GCPButton>
          </div>
        </form>
      </GCPCard>

      {error && (
        <div className="error-message">
          <div className="error-icon">⚠️</div>
          <div className="error-content">
            <h3>パスが見つかりませんでした</h3>
            <p>{error}</p>
          </div>
        </div>
      )}

      {result && (
        <div className="results-section">
          <div className="results-header">
            <h3>🎯 検索結果</h3>
            <div className="path-stats">
              <div className="stat-item">
                <span className="stat-label">総コスト</span>
                <span className="stat-value highlight">{result.total_cost.toFixed(2)}</span>
              </div>
              <div className="stat-item">
                <span className="stat-label">ホップ数</span>
                <span className="stat-value">{result.hop_count}</span>
              </div>
              <div className="stat-item">
                <span className="stat-label">アルゴリズム</span>
                <span className="stat-value">{algorithm.toUpperCase()}</span>
              </div>
            </div>
          </div>

          <div className="path-visualization">
            <h4>📍 パス経路</h4>
            <div className="path-flow">
              {result.devices.map((device, index) => (
                <React.Fragment key={device.id}>
                  <div className="path-device">
                    <div className={`device-node ${device.type}`}>
                      <div className="device-icon">
                        {device.type === 'core' && '🏢'}
                        {device.type === 'distribution' && '🔄'}
                        {device.type === 'access' && '📡'}
                        {device.type === 'server' && '🖥️'}
                      </div>
                      <div className="device-info">
                        <div className="device-name">{device.name}</div>
                        <div className="device-type">{device.type}</div>
                      </div>
                    </div>
                    
                    {index < result.devices.length - 1 && result.links[index] && (
                      <div className="path-link">
                        <div className="link-arrow">→</div>
                        <div className="link-info">
                          <div className="link-ports">
                            {result.links[index].source_port} ↔ {result.links[index].target_port}
                          </div>
                          <div className="link-weight">
                            重み: {result.links[index].weight}
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                </React.Fragment>
              ))}
            </div>
          </div>

          <div className="path-details">
            <div className="devices-section">
              <h4>🖥️ 経由デバイス ({result.devices.length}個)</h4>
              <div className="device-list">
                {result.devices.map((device, index) => (
                  <div key={device.id} className="device-card">
                    <div className="device-header">
                      <div className="device-position">#{index + 1}</div>
                      <h5 className="device-name">{device.name}</h5>
                      <span className={`device-type-badge ${device.type}`}>
                        {device.type}
                      </span>
                    </div>
                    <div className="device-details">
                      <div className="detail-row">
                        <span className="detail-label">ハードウェア:</span>
                        <span className="detail-value">{device.hardware}</span>
                      </div>
                      <div className="detail-row">
                        <span className="detail-label">IPアドレス:</span>
                        <span className="detail-value">{device.ip_address}</span>
                      </div>
                      <div className="detail-row">
                        <span className="detail-label">レイヤー:</span>
                        <span className="detail-value">{device.layer}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="links-section">
              <h4>🔗 接続リンク ({result.links.length}本)</h4>
              <div className="link-list">
                {result.links.map((link, index) => (
                  <div key={link.id} className="link-card">
                    <div className="link-header">
                      <div className="link-position">#{index + 1}</div>
                      <div className="link-connection">
                        {link.source_id} → {link.target_id}
                      </div>
                      <div className="link-weight-badge">
                        重み: {link.weight}
                      </div>
                    </div>
                    <div className="link-details">
                      <div className="detail-row">
                        <span className="detail-label">送信ポート:</span>
                        <span className="detail-value">{link.source_port}</span>
                      </div>
                      <div className="detail-row">
                        <span className="detail-label">受信ポート:</span>
                        <span className="detail-value">{link.target_port}</span>
                      </div>
                      <div className="detail-row">
                        <span className="detail-label">ステータス:</span>
                        <span className={`status-badge ${link.status}`}>
                          {link.status}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      <style jsx>{`
        .path-analysis {
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
          margin-bottom: 32px;
        }

        .results-header h3 {
          color: #2c3e50;
          margin-bottom: 16px;
        }

        .path-stats {
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
          color: #9b59b6;
        }

        .path-visualization {
          margin: 32px 0;
        }

        .path-visualization h4 {
          color: #2c3e50;
          margin-bottom: 20px;
        }

        .path-flow {
          display: flex;
          flex-direction: column;
          gap: 16px;
          background: #f8f9fa;
          padding: 24px;
          border-radius: 12px;
          border: 2px solid #e9ecef;
        }

        .path-device {
          display: flex;
          flex-direction: column;
          align-items: center;
        }

        .device-node {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 16px 20px;
          background: white;
          border-radius: 12px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.1);
          border: 3px solid;
          min-width: 280px;
        }

        .device-node.core {
          border-color: #e74c3c;
        }

        .device-node.distribution {
          border-color: #f39c12;
        }

        .device-node.access {
          border-color: #27ae60;
        }

        .device-node.server {
          border-color: #9b59b6;
        }

        .device-icon {
          font-size: 2rem;
        }

        .device-info {
          flex: 1;
        }

        .device-name {
          font-weight: 700;
          color: #2c3e50;
          font-size: 1.1rem;
        }

        .device-type {
          color: #666;
          font-size: 0.9rem;
          text-transform: uppercase;
        }

        .path-link {
          display: flex;
          flex-direction: column;
          align-items: center;
          margin: 12px 0;
        }

        .link-arrow {
          font-size: 2rem;
          color: #9b59b6;
          font-weight: bold;
        }

        .link-info {
          background: #e8e3f1;
          padding: 8px 16px;
          border-radius: 8px;
          margin-top: 8px;
          text-align: center;
        }

        .link-ports {
          font-weight: 600;
          color: #8e44ad;
          font-size: 0.9rem;
        }

        .link-weight {
          color: #666;
          font-size: 0.85rem;
        }

        .path-details {
          margin-top: 32px;
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 32px;
        }

        @media (max-width: 768px) {
          .path-details {
            grid-template-columns: 1fr;
          }
        }

        .devices-section h4,
        .links-section h4 {
          color: #2c3e50;
          margin-bottom: 16px;
        }

        .device-list,
        .link-list {
          display: flex;
          flex-direction: column;
          gap: 16px;
        }

        .device-card,
        .link-card {
          background: #f8f9fa;
          border-radius: 12px;
          padding: 16px;
          border: 2px solid #e9ecef;
          transition: all 0.2s;
        }

        .device-card:hover,
        .link-card:hover {
          border-color: #9b59b6;
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }

        .device-header,
        .link-header {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 12px;
        }

        .device-position,
        .link-position {
          background: #9b59b6;
          color: white;
          border-radius: 50%;
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-weight: 700;
          font-size: 0.9rem;
        }

        .device-name {
          font-size: 1.1rem;
          font-weight: 700;
          color: #2c3e50;
          margin: 0;
          flex: 1;
        }

        .link-connection {
          font-size: 1rem;
          font-weight: 700;
          color: #2c3e50;
          flex: 1;
        }

        .device-type-badge {
          padding: 4px 12px;
          border-radius: 20px;
          font-size: 0.85rem;
          font-weight: 600;
          text-transform: uppercase;
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

        .link-weight-badge {
          background: #9b59b6;
          color: white;
          padding: 4px 12px;
          border-radius: 16px;
          font-size: 0.85rem;
          font-weight: 600;
        }

        .device-details,
        .link-details {
          display: flex;
          flex-direction: column;
          gap: 6px;
        }

        .detail-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .detail-label {
          font-size: 0.9rem;
          color: #666;
          font-weight: 500;
        }

        .detail-value {
          font-size: 0.9rem;
          color: #2c3e50;
          font-weight: 600;
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

export default PathAnalysis