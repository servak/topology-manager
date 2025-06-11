import React, { useState, useEffect } from 'react'
import TopologyGraph from './components/TopologyGraph'
import DeviceSelector from './components/DeviceSelector'
import './App.css'

function App() {
  const [topology, setTopology] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [selectedDevice, setSelectedDevice] = useState('')
  const [depth, setDepth] = useState(3)

  const fetchTopology = async (hostname, explorationDepth = 3) => {
    if (!hostname) return

    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`/api/topology?hostname=${encodeURIComponent(hostname)}&depth=${explorationDepth}`)
      
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
    fetchTopology(hostname, depth)
  }

  const handleDepthChange = (newDepth) => {
    setDepth(newDepth)
    if (selectedDevice) {
      fetchTopology(selectedDevice, newDepth)
    }
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>Network Topology Manager</h1>
        <DeviceSelector
          onDeviceSelect={handleDeviceSearch}
          selectedDevice={selectedDevice}
          depth={depth}
          onDepthChange={handleDepthChange}
          loading={loading}
        />
      </header>

      <main className="app-main">
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
            <TopologyGraph topology={topology} />
          </div>
        )}

        {!topology && !loading && !error && (
          <div className="welcome-message">
            <h2>Welcome to Network Topology Manager</h2>
            <p>Enter a device hostname above to visualize the network topology.</p>
          </div>
        )}
      </main>
    </div>
  )
}

export default App