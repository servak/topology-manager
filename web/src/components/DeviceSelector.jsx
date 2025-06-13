import React, { useState, useEffect, useRef } from 'react'
import { FormContainer, FormGrid, FormGroup, FormInput, FormSelect, FormButton } from './common/FormStyles'

function DeviceSelector({ onDeviceSelect, selectedDevice, depth, onDepthChange, loading, groupingOptions, onGroupingChange }) {
  const [inputValue, setInputValue] = useState(selectedDevice)
  const [searchResults, setSearchResults] = useState([])
  const [isSearching, setIsSearching] = useState(false)
  const [showDropdown, setShowDropdown] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const searchTimeoutRef = useRef()
  const dropdownRef = useRef()

  // selectedDeviceが変更されたときにinputValueを更新
  useEffect(() => {
    setInputValue(selectedDevice)
  }, [selectedDevice])

  // デバイス検索のAPI呼び出し
  const searchDevices = async (query) => {
    if (!query || query.length < 2) {
      setSearchResults([])
      setShowDropdown(false)
      return
    }

    setIsSearching(true)
    try {
      const response = await fetch(`/api/devices/search?q=${encodeURIComponent(query)}&limit=10`)
      if (response.ok) {
        const data = await response.json()
        setSearchResults(data.devices || [])
        setShowDropdown(data.devices && data.devices.length > 0)
        setSelectedIndex(-1)
      }
    } catch (error) {
      console.error('Failed to search devices:', error)
      setSearchResults([])
      setShowDropdown(false)
    } finally {
      setIsSearching(false)
    }
  }

  // 入力値変更時のハンドラー
  const handleInputChange = (e) => {
    const value = e.target.value
    setInputValue(value)

    // 既存のタイマーをクリア
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current)
    }

    // 300ms後に検索実行（デバウンス）
    searchTimeoutRef.current = setTimeout(() => {
      searchDevices(value)
    }, 300)
  }

  // キーボード操作のハンドラー
  const handleKeyDown = (e) => {
    if (!showDropdown) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(prev => 
          prev < searchResults.length - 1 ? prev + 1 : prev
        )
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(prev => prev > 0 ? prev - 1 : -1)
        break
      case 'Enter':
        e.preventDefault()
        if (selectedIndex >= 0 && selectedIndex < searchResults.length) {
          selectDevice(searchResults[selectedIndex])
        } else {
          handleSubmit(e)
        }
        break
      case 'Escape':
        setShowDropdown(false)
        setSelectedIndex(-1)
        break
    }
  }

  // デバイス選択のハンドラー
  const selectDevice = (device) => {
    setInputValue(device.id)
    setShowDropdown(false)
    setSelectedIndex(-1)
    setSearchResults([])
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    setShowDropdown(false)
    if (inputValue.trim()) {
      onDeviceSelect(inputValue.trim())
    }
  }

  const handleDepthChange = (e) => {
    onDepthChange(parseInt(e.target.value))
  }

  // ドロップダウン外クリック時の処理
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setShowDropdown(false)
        setSelectedIndex(-1)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleGroupingToggle = (e) => {
    if (onGroupingChange) {
      onGroupingChange({
        ...groupingOptions,
        enabled: e.target.checked
      })
    }
  }

  const handleGroupingOptionChange = (option, value) => {
    if (onGroupingChange) {
      onGroupingChange({
        ...groupingOptions,
        [option]: value
      })
    }
  }

  return (
    <FormContainer onSubmit={handleSubmit}>
      <FormGrid columns={5}>
        <FormGroup label="デバイスID" htmlFor="deviceId">
          <div className="autocomplete-container" ref={dropdownRef} style={{ position: 'relative' }}>
            <FormInput
              id="deviceId"
              type="text"
              placeholder="例: core-001, access-019"
              value={inputValue}
              onChange={handleInputChange}
              onKeyDown={handleKeyDown}
              onFocus={() => inputValue.length >= 2 && searchResults.length > 0 && setShowDropdown(true)}
              required
            />
            {isSearching && (
              <div style={{
                position: 'absolute',
                right: '8px',
                top: '50%',
                transform: 'translateY(-50%)',
                fontSize: '12px',
                color: '#666'
              }}>
                🔄
              </div>
            )}
            {showDropdown && searchResults.length > 0 && (
              <div style={{
                position: 'absolute',
                top: '100%',
                left: 0,
                right: 0,
                backgroundColor: 'white',
                border: '1px solid #ddd',
                borderRadius: '4px',
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                maxHeight: '200px',
                overflowY: 'auto',
                zIndex: 1000
              }}>
                {searchResults.map((device, index) => (
                  <div
                    key={device.id}
                    onClick={() => selectDevice(device)}
                    style={{
                      padding: '8px 12px',
                      cursor: 'pointer',
                      backgroundColor: index === selectedIndex ? '#f0f0f0' : 'white',
                      borderBottom: index < searchResults.length - 1 ? '1px solid #eee' : 'none'
                    }}
                    onMouseEnter={() => setSelectedIndex(index)}
                  >
                    <div style={{ fontWeight: 'bold' }}>{device.id}</div>
                    {device.hardware && (
                      <div style={{ fontSize: '11px', color: '#999' }}>{device.hardware}</div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
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

        <FormGroup label="スマートグルーピング" htmlFor="grouping">
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <label style={{ display: 'flex', alignItems: 'center', fontSize: '14px' }}>
              <input
                type="checkbox"
                checked={groupingOptions?.enabled || false}
                onChange={handleGroupingToggle}
                style={{ marginRight: '8px' }}
              />
              大規模トポロジー対応
            </label>
            {groupingOptions?.enabled ? (
              <div style={{ fontSize: '12px', color: '#666' }}>
                <label style={{ display: 'flex', alignItems: 'center', marginBottom: '4px' }}>
                  <input
                    type="checkbox"
                    checked={groupingOptions?.groupByPrefix || false}
                    onChange={(e) => handleGroupingOptionChange('groupByPrefix', e.target.checked)}
                    style={{ marginRight: '6px' }}
                  />
                  共通プレフィックス
                </label>
                <label style={{ display: 'flex', alignItems: 'center' }}>
                  <input
                    type="checkbox"
                    checked={groupingOptions?.groupByType || false}
                    onChange={(e) => handleGroupingOptionChange('groupByType', e.target.checked)}
                    style={{ marginRight: '6px' }}
                  />
                  デバイスタイプ
                </label>
              </div>
            ) : (
              <div style={{ fontSize: '11px', color: '#e74c3c', marginTop: '4px' }}>
                ⚠️ 大規模ネットワークでは表示が重くなる可能性があります
              </div>
            )}
          </div>
        </FormGroup>

        <FormGroup label="表示最適化" htmlFor="groupSettings">
          {groupingOptions?.enabled && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
              <label style={{ fontSize: '12px', color: '#666' }}>
                集約閾値:
                <FormSelect
                  value={groupingOptions?.minGroupSize || 3}
                  onChange={(e) => handleGroupingOptionChange('minGroupSize', parseInt(e.target.value))}
                  style={{ marginLeft: '4px', fontSize: '12px', padding: '2px' }}
                >
                  <option value={2}>2</option>
                  <option value={3}>3</option>
                  <option value={4}>4</option>
                  <option value={5}>5</option>
                </FormSelect>
              </label>
              <label style={{ fontSize: '12px', color: '#666' }}>
                集約開始深度:
                <FormSelect
                  value={groupingOptions?.maxGroupDepth || 2}
                  onChange={(e) => handleGroupingOptionChange('maxGroupDepth', parseInt(e.target.value))}
                  style={{ marginLeft: '4px', fontSize: '12px', padding: '2px' }}
                >
                  <option value={1}>1</option>
                  <option value={2}>2</option>
                  <option value={3}>3</option>
                </FormSelect>
              </label>
            </div>
          )}
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