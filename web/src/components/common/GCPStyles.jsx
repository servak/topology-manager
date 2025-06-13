import React from 'react'

// GCP風のカラーパレット
export const GCP_COLORS = {
  primary: '#1a73e8',
  primaryHover: '#1557b0',
  secondary: '#5f6368',
  success: '#34a853',
  warning: '#fbbc04',
  error: '#ea4335',
  surface: '#ffffff',
  background: '#f8f9fa',
  border: '#dadce0',
  textPrimary: '#202124',
  textSecondary: '#5f6368',
  textTertiary: '#80868b'
}

// GCP風のフォームコンテナ
export const GCPCard = ({ children, title, subtitle, className = "" }) => (
  <div className={`gcp-card ${className}`}>
    {title && (
      <div className="gcp-card-header">
        <h2 className="gcp-card-title">{title}</h2>
        {subtitle && <p className="gcp-card-subtitle">{subtitle}</p>}
      </div>
    )}
    <div className="gcp-card-content">
      {children}
    </div>
    <style jsx>{`
      .gcp-card {
        background: ${GCP_COLORS.surface};
        border: 1px solid ${GCP_COLORS.border};
        border-radius: 8px;
        box-shadow: 0 1px 2px 0 rgba(60,64,67,0.3), 0 1px 3px 1px rgba(60,64,67,0.15);
        margin-bottom: 24px;
        overflow: hidden;
      }

      .gcp-card-header {
        padding: 24px 24px 0 24px;
        border-bottom: none;
      }

      .gcp-card-title {
        font-size: 1.375rem;
        font-weight: 400;
        color: ${GCP_COLORS.textPrimary};
        margin: 0 0 4px 0;
        line-height: 1.75rem;
      }

      .gcp-card-subtitle {
        font-size: 0.875rem;
        color: ${GCP_COLORS.textSecondary};
        margin: 0 0 16px 0;
        line-height: 1.25rem;
      }

      .gcp-card-content {
        padding: 24px;
      }
    `}</style>
  </div>
)

// GCP風のフォームフィールド
export const GCPFormField = ({ label, children, required = false, helperText, error }) => (
  <div className="gcp-form-field">
    <label className="gcp-form-label">
      {label}
      {required && <span className="gcp-required">*</span>}
    </label>
    {children}
    {helperText && <div className="gcp-helper-text">{helperText}</div>}
    {error && <div className="gcp-error-text">{error}</div>}
    <style jsx>{`
      .gcp-form-field {
        margin-bottom: 24px;
      }

      .gcp-form-label {
        display: block;
        font-size: 0.875rem;
        font-weight: 500;
        color: ${GCP_COLORS.textPrimary};
        margin-bottom: 8px;
        line-height: 1.25rem;
      }

      .gcp-required {
        color: ${GCP_COLORS.error};
        margin-left: 4px;
      }

      .gcp-helper-text {
        font-size: 0.75rem;
        color: ${GCP_COLORS.textSecondary};
        margin-top: 4px;
        line-height: 1rem;
      }

      .gcp-error-text {
        font-size: 0.75rem;
        color: ${GCP_COLORS.error};
        margin-top: 4px;
        line-height: 1rem;
      }
    `}</style>
  </div>
)

// GCP風のインプット
export const GCPInput = ({ 
  type = "text", 
  value, 
  onChange, 
  placeholder, 
  disabled = false,
  error = false,
  ...props 
}) => (
  <>
    <input
      type={type}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      disabled={disabled}
      className={`gcp-input ${error ? 'error' : ''} ${disabled ? 'disabled' : ''}`}
      {...props}
    />
    <style jsx>{`
      .gcp-input {
        width: 100%;
        padding: 12px 16px;
        font-size: 0.875rem;
        line-height: 1.25rem;
        color: ${GCP_COLORS.textPrimary};
        background: ${GCP_COLORS.surface};
        border: 1px solid ${GCP_COLORS.border};
        border-radius: 4px;
        transition: border-color 0.2s, box-shadow 0.2s;
        outline: none;
      }

      .gcp-input:focus {
        border-color: ${GCP_COLORS.primary};
        box-shadow: 0 0 0 1px ${GCP_COLORS.primary};
      }

      .gcp-input:hover:not(:disabled) {
        border-color: ${GCP_COLORS.textSecondary};
      }

      .gcp-input.error {
        border-color: ${GCP_COLORS.error};
      }

      .gcp-input.error:focus {
        border-color: ${GCP_COLORS.error};
        box-shadow: 0 0 0 1px ${GCP_COLORS.error};
      }

      .gcp-input.disabled {
        background: ${GCP_COLORS.background};
        color: ${GCP_COLORS.textTertiary};
        cursor: not-allowed;
      }
    `}</style>
  </>
)

// GCP風のセレクト
export const GCPSelect = ({ value, onChange, children, disabled = false, error = false, ...props }) => (
  <>
    <select
      value={value}
      onChange={onChange}
      disabled={disabled}
      className={`gcp-select ${error ? 'error' : ''} ${disabled ? 'disabled' : ''}`}
      {...props}
    >
      {children}
    </select>
    <style jsx>{`
      .gcp-select {
        width: 100%;
        padding: 12px 16px;
        font-size: 0.875rem;
        line-height: 1.25rem;
        color: ${GCP_COLORS.textPrimary};
        background: ${GCP_COLORS.surface};
        border: 1px solid ${GCP_COLORS.border};
        border-radius: 4px;
        transition: border-color 0.2s, box-shadow 0.2s;
        outline: none;
        cursor: pointer;
        appearance: none;
        background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
        background-position: right 12px center;
        background-repeat: no-repeat;
        background-size: 16px;
        padding-right: 40px;
      }

      .gcp-select:focus {
        border-color: ${GCP_COLORS.primary};
        box-shadow: 0 0 0 1px ${GCP_COLORS.primary};
      }

      .gcp-select:hover:not(:disabled) {
        border-color: ${GCP_COLORS.textSecondary};
      }

      .gcp-select.error {
        border-color: ${GCP_COLORS.error};
      }

      .gcp-select.error:focus {
        border-color: ${GCP_COLORS.error};
        box-shadow: 0 0 0 1px ${GCP_COLORS.error};
      }

      .gcp-select.disabled {
        background: ${GCP_COLORS.background};
        color: ${GCP_COLORS.textTertiary};
        cursor: not-allowed;
      }
    `}</style>
  </>
)

// GCP風のボタン
export const GCPButton = ({ 
  variant = 'primary',
  size = 'medium',
  disabled = false,
  loading = false,
  children,
  onClick,
  type = 'button',
  ...props 
}) => {
  const getButtonStyles = () => {
    const baseStyles = {
      padding: size === 'small' ? '8px 16px' : size === 'large' ? '16px 32px' : '12px 24px',
      fontSize: size === 'small' ? '0.75rem' : '0.875rem',
      fontWeight: '500',
      borderRadius: '4px',
      border: 'none',
      cursor: disabled ? 'not-allowed' : 'pointer',
      transition: 'all 0.2s',
      outline: 'none',
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '8px',
      lineHeight: '1.25rem'
    }

    switch (variant) {
      case 'primary':
        return {
          ...baseStyles,
          background: disabled ? GCP_COLORS.textTertiary : GCP_COLORS.primary,
          color: GCP_COLORS.surface,
          boxShadow: disabled ? 'none' : '0 1px 2px 0 rgba(60,64,67,0.3), 0 1px 3px 1px rgba(60,64,67,0.15)'
        }
      case 'secondary':
        return {
          ...baseStyles,
          background: GCP_COLORS.surface,
          color: disabled ? GCP_COLORS.textTertiary : GCP_COLORS.primary,
          border: `1px solid ${disabled ? GCP_COLORS.border : GCP_COLORS.primary}`
        }
      case 'text':
        return {
          ...baseStyles,
          background: 'transparent',
          color: disabled ? GCP_COLORS.textTertiary : GCP_COLORS.primary,
          border: 'none',
          boxShadow: 'none'
        }
      default:
        return baseStyles
    }
  }

  return (
    <>
      <button
        type={type}
        disabled={disabled || loading}
        onClick={onClick}
        className="gcp-button"
        style={getButtonStyles()}
        {...props}
      >
        {loading && <span className="gcp-spinner">⏳</span>}
        {children}
      </button>
      <style jsx>{`
        .gcp-button:hover:not(:disabled) {
          transform: ${variant === 'primary' ? 'translateY(-1px)' : 'none'};
          box-shadow: ${variant === 'primary' ? '0 2px 4px 0 rgba(60,64,67,0.3), 0 1px 5px 1px rgba(60,64,67,0.15)' : 'none'};
          background: ${variant === 'secondary' ? GCP_COLORS.background : variant === 'text' ? 'rgba(26, 115, 232, 0.04)' : ''};
        }

        .gcp-button:active:not(:disabled) {
          transform: translateY(0);
        }

        .gcp-spinner {
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </>
  )
}

// GCP風のテーブル
export const GCPTable = ({ headers, children, className = "" }) => (
  <div className={`gcp-table-container ${className}`}>
    <table className="gcp-table">
      <thead>
        <tr>
          {headers.map((header, index) => (
            <th key={index} className="gcp-table-header">
              {header}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {children}
      </tbody>
    </table>
    <style jsx>{`
      .gcp-table-container {
        background: ${GCP_COLORS.surface};
        border: 1px solid ${GCP_COLORS.border};
        border-radius: 8px;
        overflow: hidden;
      }

      .gcp-table {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.875rem;
      }

      .gcp-table-header {
        background: ${GCP_COLORS.background};
        padding: 16px;
        text-align: left;
        font-weight: 500;
        color: ${GCP_COLORS.textPrimary};
        border-bottom: 1px solid ${GCP_COLORS.border};
        font-size: 0.75rem;
        text-transform: uppercase;
        letter-spacing: 0.5px;
      }

      :global(.gcp-table tbody tr) {
        border-bottom: 1px solid ${GCP_COLORS.border};
        transition: background-color 0.2s;
      }

      :global(.gcp-table tbody tr:hover) {
        background: ${GCP_COLORS.background};
      }

      :global(.gcp-table tbody tr:last-child) {
        border-bottom: none;
      }

      :global(.gcp-table td) {
        padding: 16px;
        vertical-align: middle;
        color: ${GCP_COLORS.textPrimary};
      }
    `}</style>
  </div>
)

// GCP風のチップ/バッジ
export const GCPChip = ({ children, variant = 'default', size = 'medium' }) => {
  const getChipStyles = () => {
    const baseStyles = {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      borderRadius: '16px',
      fontSize: size === 'small' ? '0.6875rem' : '0.75rem',
      fontWeight: '500',
      lineHeight: '1rem',
      padding: size === 'small' ? '2px 8px' : '4px 12px',
      whiteSpace: 'nowrap'
    }

    switch (variant) {
      case 'success':
        return { ...baseStyles, background: '#e8f5e8', color: GCP_COLORS.success }
      case 'warning':
        return { ...baseStyles, background: '#fef7e0', color: '#ea8600' }
      case 'error':
        return { ...baseStyles, background: '#fce8e6', color: GCP_COLORS.error }
      case 'info':
        return { ...baseStyles, background: '#e8f0fe', color: GCP_COLORS.primary }
      default:
        return { ...baseStyles, background: GCP_COLORS.background, color: GCP_COLORS.textSecondary }
    }
  }

  return (
    <span className="gcp-chip" style={getChipStyles()}>
      {children}
    </span>
  )
}

export default {
  GCPCard,
  GCPFormField,
  GCPInput,
  GCPSelect,
  GCPButton,
  GCPTable,
  GCPChip,
  GCP_COLORS
}