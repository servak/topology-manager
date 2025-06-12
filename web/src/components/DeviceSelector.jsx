import React, { useState } from 'react'

function DeviceSelector({ onDeviceSelect, selectedDevice, depth, onDepthChange, loading }) {
  const [inputValue, setInputValue] = useState(selectedDevice)

  const handleSubmit = (e) => {
    e.preventDefault()
    if (inputValue.trim()) {
      onDeviceSelect(inputValue.trim())
    }
  }

  const handleDepthChange = (e) => {
    onDepthChange(parseInt(e.target.value))
  }

  return (
    <form className="device-selector" onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Enter device ID (e.g., core-001, access-019)"
        value={inputValue}
        onChange={(e) => setInputValue(e.target.value)}
        disabled={loading}
      />
      
      <select value={depth} onChange={handleDepthChange} disabled={loading}>
        <option value={1}>Depth: 1</option>
        <option value={2}>Depth: 2</option>
        <option value={3}>Depth: 3</option>
        <option value={4}>Depth: 4</option>
        <option value={5}>Depth: 5</option>
      </select>
      
      <button type="submit" disabled={loading || !inputValue.trim()}>
        {loading ? 'Loading...' : 'Visualize'}
      </button>
    </form>
  )
}

export default DeviceSelector