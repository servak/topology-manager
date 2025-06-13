import React from 'react'

// ダッシュボード風のカラーパレット
export const DASHBOARD_COLORS = {
  primary: '#6366f1',
  primaryHover: '#4f46e5',
  secondary: '#ec4899',
  success: '#10b981',
  warning: '#f59e0b',
  error: '#ef4444',
  info: '#3b82f6',
  surface: '#ffffff',
  background: '#f1f5f9',
  backgroundDark: '#0f172a',
  cardBackground: '#ffffff',
  border: '#e2e8f0',
  borderHover: '#cbd5e1',
  textPrimary: '#1e293b',
  textSecondary: '#64748b',
  textMuted: '#94a3b8'
}

// ダッシュボード風のカード
export const DashboardCard = ({ children, title, subtitle, icon, className = "", gradient = false }) => (
  <div className={`dashboard-card ${gradient ? 'gradient' : ''} ${className}`}>
    {title && (
      <div className="dashboard-card-header">
        {icon && <div className="dashboard-card-icon">{icon}</div>}
        <div className="dashboard-card-title-wrapper">
          <h3 className="dashboard-card-title">{title}</h3>
          {subtitle && <p className="dashboard-card-subtitle">{subtitle}</p>}
        </div>
      </div>
    )}
    <div className="dashboard-card-content">
      {children}
    </div>
    <style jsx>{`
      .dashboard-card {
        background: ${DASHBOARD_COLORS.cardBackground};
        border-radius: 16px;
        padding: 24px;
        box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
        transition: all 0.3s ease;
        border: 1px solid ${DASHBOARD_COLORS.border};
        margin-bottom: 24px;
        position: relative;
        overflow: hidden;
      }

      .dashboard-card:hover {
        transform: translateY(-4px);
        box-shadow: 0 10px 25px 0 rgba(0, 0, 0, 0.1), 0 4px 6px 0 rgba(0, 0, 0, 0.05);
        border-color: ${DASHBOARD_COLORS.borderHover};
      }

      .dashboard-card.gradient {
        background: linear-gradient(135deg, ${DASHBOARD_COLORS.primary} 0%, ${DASHBOARD_COLORS.secondary} 100%);
        color: white;
      }

      .dashboard-card.gradient::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%);
        pointer-events: none;
      }

      .dashboard-card-header {
        display: flex;
        align-items: flex-start;
        gap: 16px;
        margin-bottom: 20px;
      }

      .dashboard-card-icon {
        font-size: 2rem;
        display: flex;
        align-items: center;
        justify-content: center;
        width: 48px;
        height: 48px;
        background: linear-gradient(135deg, ${DASHBOARD_COLORS.primary}, ${DASHBOARD_COLORS.secondary});
        border-radius: 12px;
        color: white;
        flex-shrink: 0;
      }

      .dashboard-card.gradient .dashboard-card-icon {
        background: rgba(255, 255, 255, 0.2);
        backdrop-filter: blur(10px);
      }

      .dashboard-card-title-wrapper {
        flex: 1;
      }

      .dashboard-card-title {
        font-size: 1.25rem;
        font-weight: 600;
        color: ${DASHBOARD_COLORS.textPrimary};
        margin: 0 0 4px 0;
        line-height: 1.5;
      }

      .dashboard-card.gradient .dashboard-card-title {
        color: white;
      }

      .dashboard-card-subtitle {
        font-size: 0.875rem;
        color: ${DASHBOARD_COLORS.textSecondary};
        margin: 0;
        line-height: 1.4;
      }

      .dashboard-card.gradient .dashboard-card-subtitle {
        color: rgba(255, 255, 255, 0.8);
      }

      .dashboard-card-content {
        position: relative;
        z-index: 1;
      }
    `}</style>
  </div>
)

// ダッシュボード風のメトリクスカード
export const MetricCard = ({ title, value, change, changeType, icon, color = DASHBOARD_COLORS.primary }) => (
  <div className="metric-card">
    <div className="metric-header">
      <div className="metric-icon" style={{ background: color }}>
        {icon}
      </div>
      <div className="metric-info">
        <h4 className="metric-title">{title}</h4>
        <div className="metric-value">{value}</div>
      </div>
    </div>
    {change && (
      <div className={`metric-change ${changeType}`}>
        <span className="metric-change-icon">
          {changeType === 'increase' ? '↗️' : changeType === 'decrease' ? '↘️' : '➡️'}
        </span>
        <span className="metric-change-text">{change}</span>
      </div>
    )}
    <style jsx>{`
      .metric-card {
        background: ${DASHBOARD_COLORS.cardBackground};
        border-radius: 12px;
        padding: 20px;
        border: 1px solid ${DASHBOARD_COLORS.border};
        transition: all 0.3s ease;
        position: relative;
        overflow: hidden;
      }

      .metric-card:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 20px 0 rgba(0, 0, 0, 0.1);
        border-color: ${DASHBOARD_COLORS.borderHover};
      }

      .metric-header {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 12px;
      }

      .metric-icon {
        width: 40px;
        height: 40px;
        border-radius: 10px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-size: 1.2rem;
      }

      .metric-info {
        flex: 1;
      }

      .metric-title {
        font-size: 0.875rem;
        color: ${DASHBOARD_COLORS.textSecondary};
        margin: 0 0 4px 0;
        font-weight: 500;
      }

      .metric-value {
        font-size: 1.75rem;
        font-weight: 700;
        color: ${DASHBOARD_COLORS.textPrimary};
        line-height: 1;
      }

      .metric-change {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 0.75rem;
        font-weight: 500;
      }

      .metric-change.increase {
        color: ${DASHBOARD_COLORS.success};
      }

      .metric-change.decrease {
        color: ${DASHBOARD_COLORS.error};
      }

      .metric-change.neutral {
        color: ${DASHBOARD_COLORS.textMuted};
      }
    `}</style>
  </div>
)

// ダッシュボード風のボタン
export const DashboardButton = ({ 
  children, 
  variant = 'primary', 
  size = 'medium',
  disabled = false,
  loading = false,
  icon,
  onClick,
  type = 'button',
  className = "",
  ...props 
}) => {
  const getButtonStyles = () => {
    const baseStyles = {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '8px',
      border: 'none',
      borderRadius: '12px',
      fontWeight: '600',
      cursor: disabled ? 'not-allowed' : 'pointer',
      transition: 'all 0.2s ease',
      outline: 'none',
      fontFamily: 'inherit',
      fontSize: size === 'small' ? '0.875rem' : size === 'large' ? '1.125rem' : '1rem',
      padding: size === 'small' ? '8px 16px' : size === 'large' ? '16px 32px' : '12px 24px',
      lineHeight: '1.25'
    }

    if (disabled) {
      return {
        ...baseStyles,
        background: DASHBOARD_COLORS.textMuted,
        color: 'white',
        transform: 'none',
        boxShadow: 'none'
      }
    }

    switch (variant) {
      case 'primary':
        return {
          ...baseStyles,
          background: `linear-gradient(135deg, ${DASHBOARD_COLORS.primary}, ${DASHBOARD_COLORS.primaryHover})`,
          color: 'white',
          boxShadow: `0 4px 12px rgba(99, 102, 241, 0.4)`
        }
      case 'secondary':
        return {
          ...baseStyles,
          background: `linear-gradient(135deg, ${DASHBOARD_COLORS.secondary}, #db2777)`,
          color: 'white',
          boxShadow: `0 4px 12px rgba(236, 72, 153, 0.4)`
        }
      case 'success':
        return {
          ...baseStyles,
          background: `linear-gradient(135deg, ${DASHBOARD_COLORS.success}, #059669)`,
          color: 'white',
          boxShadow: `0 4px 12px rgba(16, 185, 129, 0.4)`
        }
      case 'outline':
        return {
          ...baseStyles,
          background: 'transparent',
          color: DASHBOARD_COLORS.primary,
          border: `2px solid ${DASHBOARD_COLORS.primary}`,
          boxShadow: 'none'
        }
      case 'ghost':
        return {
          ...baseStyles,
          background: 'transparent',
          color: DASHBOARD_COLORS.textPrimary,
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
        className={`dashboard-button ${className}`}
        style={getButtonStyles()}
        {...props}
      >
        {loading && <span className="button-spinner">⏳</span>}
        {icon && !loading && <span className="button-icon">{icon}</span>}
        {children}
      </button>
      <style jsx>{`
        .dashboard-button:hover:not(:disabled) {
          transform: translateY(-2px);
          box-shadow: ${
            variant === 'primary' ? '0 8px 20px rgba(99, 102, 241, 0.6)' :
            variant === 'secondary' ? '0 8px 20px rgba(236, 72, 153, 0.6)' :
            variant === 'success' ? '0 8px 20px rgba(16, 185, 129, 0.6)' :
            variant === 'outline' ? `0 4px 12px rgba(99, 102, 241, 0.2)` :
            '0 4px 12px rgba(0, 0, 0, 0.1)'
          } !important;
        }

        .dashboard-button:active:not(:disabled) {
          transform: translateY(0);
        }

        .button-spinner {
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

// ダッシュボード風のフォーム要素
export const DashboardInput = ({ 
  label, 
  value, 
  onChange, 
  placeholder,
  type = 'text',
  disabled = false,
  error,
  helperText,
  icon,
  className = "",
  ...props 
}) => (
  <div className={`dashboard-input-wrapper ${className}`}>
    {label && <label className="dashboard-input-label">{label}</label>}
    <div className="dashboard-input-container">
      {icon && <div className="dashboard-input-icon">{icon}</div>}
      <input
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        disabled={disabled}
        className={`dashboard-input ${icon ? 'with-icon' : ''} ${error ? 'error' : ''}`}
        {...props}
      />
    </div>
    {helperText && <div className="dashboard-input-helper">{helperText}</div>}
    {error && <div className="dashboard-input-error">{error}</div>}
    <style jsx>{`
      .dashboard-input-wrapper {
        margin-bottom: 20px;
      }

      .dashboard-input-label {
        display: block;
        font-size: 0.875rem;
        font-weight: 600;
        color: ${DASHBOARD_COLORS.textPrimary};
        margin-bottom: 8px;
      }

      .dashboard-input-container {
        position: relative;
      }

      .dashboard-input {
        width: 100%;
        padding: 14px 16px;
        font-size: 1rem;
        color: ${DASHBOARD_COLORS.textPrimary};
        background: ${DASHBOARD_COLORS.cardBackground};
        border: 2px solid ${DASHBOARD_COLORS.border};
        border-radius: 12px;
        transition: all 0.2s ease;
        outline: none;
      }

      .dashboard-input.with-icon {
        padding-left: 48px;
      }

      .dashboard-input:focus {
        border-color: ${DASHBOARD_COLORS.primary};
        box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
      }

      .dashboard-input:hover:not(:disabled) {
        border-color: ${DASHBOARD_COLORS.borderHover};
      }

      .dashboard-input.error {
        border-color: ${DASHBOARD_COLORS.error};
      }

      .dashboard-input.error:focus {
        border-color: ${DASHBOARD_COLORS.error};
        box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
      }

      .dashboard-input:disabled {
        background: ${DASHBOARD_COLORS.background};
        color: ${DASHBOARD_COLORS.textMuted};
        cursor: not-allowed;
      }

      .dashboard-input-icon {
        position: absolute;
        left: 16px;
        top: 50%;
        transform: translateY(-50%);
        color: ${DASHBOARD_COLORS.textSecondary};
        font-size: 1.2rem;
        pointer-events: none;
      }

      .dashboard-input-helper {
        font-size: 0.75rem;
        color: ${DASHBOARD_COLORS.textSecondary};
        margin-top: 4px;
      }

      .dashboard-input-error {
        font-size: 0.75rem;
        color: ${DASHBOARD_COLORS.error};
        margin-top: 4px;
      }
    `}</style>
  </div>
)

// ダッシュボード風のセレクト
export const DashboardSelect = ({ 
  label, 
  value, 
  onChange, 
  children,
  disabled = false,
  error,
  helperText,
  className = "",
  ...props 
}) => (
  <div className={`dashboard-select-wrapper ${className}`}>
    {label && <label className="dashboard-select-label">{label}</label>}
    <select
      value={value}
      onChange={onChange}
      disabled={disabled}
      className={`dashboard-select ${error ? 'error' : ''}`}
      {...props}
    >
      {children}
    </select>
    {helperText && <div className="dashboard-select-helper">{helperText}</div>}
    {error && <div className="dashboard-select-error">{error}</div>}
    <style jsx>{`
      .dashboard-select-wrapper {
        margin-bottom: 20px;
      }

      .dashboard-select-label {
        display: block;
        font-size: 0.875rem;
        font-weight: 600;
        color: ${DASHBOARD_COLORS.textPrimary};
        margin-bottom: 8px;
      }

      .dashboard-select {
        width: 100%;
        padding: 14px 16px;
        font-size: 1rem;
        color: ${DASHBOARD_COLORS.textPrimary};
        background: ${DASHBOARD_COLORS.cardBackground};
        border: 2px solid ${DASHBOARD_COLORS.border};
        border-radius: 12px;
        transition: all 0.2s ease;
        outline: none;
        cursor: pointer;
        appearance: none;
        background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%2364748b' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
        background-position: right 12px center;
        background-repeat: no-repeat;
        background-size: 16px;
        padding-right: 40px;
      }

      .dashboard-select:focus {
        border-color: ${DASHBOARD_COLORS.primary};
        box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
      }

      .dashboard-select:hover:not(:disabled) {
        border-color: ${DASHBOARD_COLORS.borderHover};
      }

      .dashboard-select.error {
        border-color: ${DASHBOARD_COLORS.error};
      }

      .dashboard-select.error:focus {
        border-color: ${DASHBOARD_COLORS.error};
        box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
      }

      .dashboard-select:disabled {
        background: ${DASHBOARD_COLORS.background};
        color: ${DASHBOARD_COLORS.textMuted};
        cursor: not-allowed;
      }

      .dashboard-select-helper {
        font-size: 0.75rem;
        color: ${DASHBOARD_COLORS.textSecondary};
        margin-top: 4px;
      }

      .dashboard-select-error {
        font-size: 0.75rem;
        color: ${DASHBOARD_COLORS.error};
        margin-top: 4px;
      }
    `}</style>
  </div>
)

// ダッシュボード風のバッジ
export const DashboardBadge = ({ children, variant = 'default', size = 'medium' }) => {
  const getVariantStyles = () => {
    switch (variant) {
      case 'success':
        return { background: DASHBOARD_COLORS.success, color: 'white' }
      case 'warning':
        return { background: DASHBOARD_COLORS.warning, color: 'white' }
      case 'error':
        return { background: DASHBOARD_COLORS.error, color: 'white' }
      case 'info':
        return { background: DASHBOARD_COLORS.info, color: 'white' }
      case 'primary':
        return { background: DASHBOARD_COLORS.primary, color: 'white' }
      case 'secondary':
        return { background: DASHBOARD_COLORS.secondary, color: 'white' }
      default:
        return { background: DASHBOARD_COLORS.background, color: DASHBOARD_COLORS.textSecondary }
    }
  }

  return (
    <span 
      className="dashboard-badge" 
      style={{
        ...getVariantStyles(),
        padding: size === 'small' ? '2px 8px' : '4px 12px',
        fontSize: size === 'small' ? '0.6875rem' : '0.75rem'
      }}
    >
      {children}
      <style jsx>{`
        .dashboard-badge {
          display: inline-flex;
          align-items: center;
          justify-content: center;
          border-radius: 20px;
          font-weight: 600;
          line-height: 1;
          white-space: nowrap;
          text-transform: uppercase;
          letter-spacing: 0.025em;
        }
      `}</style>
    </span>
  )
}

export default {
  DashboardCard,
  MetricCard,
  DashboardButton,
  DashboardInput,
  DashboardSelect,
  DashboardBadge,
  DASHBOARD_COLORS
}