import React, { useState } from 'react'
import { FormContainer, FormGrid, FormGroup, FormInput, FormSelect, FormButton } from './common/FormStyles'

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
    <FormContainer onSubmit={handleSubmit}>
      <FormGrid columns={3}>
        <FormGroup label="デバイスID" htmlFor="deviceId">
          <FormInput
            id="deviceId"
            type="text"
            placeholder="例: core-001, access-019"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            required
          />
        </FormGroup>
        
        <FormGroup label="探索深度" htmlFor="depth">
          <FormSelect
            id="depth"
            value={depth}
            onChange={handleDepthChange}
          >
            <option value={1}>1ホップ</option>
            <option value={2}>2ホップ</option>
            <option value={3}>3ホップ</option>
            <option value={4}>4ホップ</option>
            <option value={5}>5ホップ</option>
          </FormSelect>
        </FormGroup>
        
        <FormGroup label=" " htmlFor="submit">
          <FormButton
            type="submit"
            disabled={loading || !inputValue.trim()}
            variant="success"
          >
            {loading ? '🔄 読み込み中...' : '🗺️ 可視化'}
          </FormButton>
        </FormGroup>
      </FormGrid>
    </FormContainer>
  )
}

export default DeviceSelector