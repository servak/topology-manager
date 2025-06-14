import React from 'react'

function DetailPanel({ selectedObject, onClose, onNavigateToDevice, onShowNeighbors, showNeighbors, onCloseNeighbors, onNeighborClick }) {
  if (!selectedObject) {
    return (
      <div className="detail-panel">
        <div className="detail-panel-header">
          <h3>Object Details</h3>
        </div>
        <div className="detail-panel-content">
          <p className="no-selection">Select a node or edge to view details</p>
        </div>
      </div>
    )
  }

  const { type, data } = selectedObject

  return (
    <div className="detail-panel">
      <div className="detail-panel-header">
        <h3>{type === 'node' ? 'Device Details' : 'Link Details'}</h3>
        <button className="close-button" onClick={onClose}>×</button>
      </div>
      
      <div className="detail-panel-content">
        {type === 'node' ? (
          <NodeDetails 
            data={data} 
            onNavigateToDevice={onNavigateToDevice}
            onShowNeighbors={onShowNeighbors}
            showNeighbors={showNeighbors}
            onCloseNeighbors={onCloseNeighbors}
            onNeighborClick={onNeighborClick}
          />
        ) : (
          <EdgeDetails data={data} />
        )}
      </div>
    </div>
  )
}

function NodeDetails({ data, onNavigateToDevice, onShowNeighbors, showNeighbors, onCloseNeighbors, onNeighborClick }) {
  return (
    <div className="node-details">
      <div className="detail-section">
        <h4>Basic Information</h4>
        <div className="detail-row">
          <span className="label">Name:</span>
          <span className="value">{data.label}</span>
        </div>
        <div className="detail-row">
          <span className="label">Type:</span>
          <span className="value device-type">{data.type}</span>
        </div>
        <div className="detail-row">
          <span className="label">Layer:</span>
          <span className="value">{data.layer}</span>
        </div>
        <div className="detail-row">
          <span className="label">Status:</span>
          <span className={`value status ${data.status}`}>{data.status}</span>
        </div>
        {data.isRoot && (
          <div className="detail-row">
            <span className="label">Role:</span>
            <span className="value root-device">ROOT DEVICE</span>
          </div>
        )}
      </div>

      {data.hardware && (
        <div className="detail-section">
          <h4>Hardware</h4>
          <div className="detail-row">
            <span className="label">Model:</span>
            <span className="value">{data.hardware}</span>
          </div>
        </div>
      )}

      <div className="detail-section">
        <h4>Network Information</h4>
        <div className="detail-row">
          <span className="label">Device ID:</span>
          <span className="value monospace">{data.id}</span>
        </div>
        {data.location && (
          <div className="detail-row">
            <span className="label">Location:</span>
            <span className="value">{data.location}</span>
          </div>
        )}
      </div>

      <div className="detail-section">
        <h4>Actions</h4>
        <button 
          className="action-button primary"
          onClick={() => {
            if (onNavigateToDevice) {
              onNavigateToDevice(data.id)
            }
          }}
          disabled={!onNavigateToDevice}
        >
          View as Root
        </button>
        <button 
          className="action-button secondary"
          onClick={() => {
            if (onShowNeighbors) {
              onShowNeighbors(data.id)
            }
          }}
          disabled={!onShowNeighbors}
        >
          Show Neighbors
        </button>
      </div>

      {/* 隣接デバイス表示 */}
      {showNeighbors && showNeighbors.deviceId === data.id && (
        <div className="neighbors-display">
          <h5>
            隣接デバイス ({showNeighbors.neighbors.length}台)
            <button className="neighbors-close" onClick={onCloseNeighbors}>×</button>
          </h5>
          <div className="neighbors-list">
            {showNeighbors.neighbors.map((neighborId, index) => (
              <div 
                key={index}
                className="neighbor-item"
                onClick={() => onNeighborClick(neighborId)}
                title={`${neighborId} をルートとして表示`}
              >
                {neighborId}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function EdgeDetails({ data }) {
  return (
    <div className="edge-details">
      <div className="detail-section">
        <h4>Connection Information</h4>
        <div className="detail-row">
          <span className="label">Source:</span>
          <span className="value monospace">{data.source}</span>
        </div>
        <div className="detail-row">
          <span className="label">Target:</span>
          <span className="value monospace">{data.target}</span>
        </div>
        <div className="detail-row">
          <span className="label">Status:</span>
          <span className={`value status ${data.status}`}>{data.status}</span>
        </div>
      </div>

      <div className="detail-section">
        <h4>Port Information</h4>
        {data.localPort && (
          <div className="detail-row">
            <span className="label">Source Port:</span>
            <span className="value monospace">{data.localPort}</span>
          </div>
        )}
        {data.remotePort && (
          <div className="detail-row">
            <span className="label">Target Port:</span>
            <span className="value monospace">{data.remotePort}</span>
          </div>
        )}
      </div>

      {data.weight && (
        <div className="detail-section">
          <h4>Link Properties</h4>
          <div className="detail-row">
            <span className="label">Weight:</span>
            <span className="value">{data.weight}</span>
          </div>
        </div>
      )}

      <div className="detail-section">
        <h4>Actions</h4>
        <button 
          className="action-button secondary"
          onClick={() => {
            console.log('Show link details:', data.id)
          }}
        >
          View Link Details
        </button>
      </div>
    </div>
  )
}

export default DetailPanel