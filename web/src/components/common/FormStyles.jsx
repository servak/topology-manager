import React from 'react'

// 共通フォームスタイルコンポーネント
export const FormContainer = ({ children, onSubmit }) => (
  <form className="form-container" onSubmit={onSubmit}>
    {children}
    <style jsx>{`
      .form-container {
        background: white;
        border-radius: 12px;
        padding: 24px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        margin-bottom: 32px;
      }
    `}</style>
  </form>
)

export const FormGrid = ({ children, columns = 4 }) => (
  <div className="form-grid">
    {children}
    <style jsx>{`
      .form-grid {
        display: grid;
        grid-template-columns: repeat(${columns}, 1fr);
        gap: 20px;
        align-items: end;
      }

      @media (max-width: 768px) {
        .form-grid {
          grid-template-columns: 1fr;
        }
      }
    `}</style>
  </div>
)

export const FormGroup = ({ label, children, htmlFor }) => (
  <div className="form-group">
    <label htmlFor={htmlFor} className="form-label">{label}</label>
    {children}
    <style jsx>{`
      .form-group {
        display: flex;
        flex-direction: column;
      }

      .form-label {
        font-weight: 600;
        color: #34495e;
        margin-bottom: 8px;
        font-size: 0.95rem;
      }
    `}</style>
  </div>
)

export const FormInput = ({ id, type = "text", value, onChange, placeholder, required = false, min, max, className = "" }) => (
  <>
    <input
      id={id}
      type={type}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      required={required}
      min={min}
      max={max}
      className={`form-input ${className}`}
    />
    <style jsx>{`
      .form-input {
        padding: 12px;
        border: 2px solid #e0e0e0;
        border-radius: 8px;
        font-size: 1rem;
        transition: border-color 0.2s;
        background: white;
      }

      .form-input:focus {
        outline: none;
        border-color: #3498db;
      }

      .form-input:disabled {
        background: #f8f9fa;
        color: #6c757d;
        cursor: not-allowed;
      }
    `}</style>
  </>
)

export const FormSelect = ({ id, value, onChange, children, className = "" }) => (
  <>
    <select
      id={id}
      value={value}
      onChange={onChange}
      className={`form-select ${className}`}
    >
      {children}
    </select>
    <style jsx>{`
      .form-select {
        padding: 12px;
        border: 2px solid #e0e0e0;
        border-radius: 8px;
        font-size: 1rem;
        transition: border-color 0.2s;
        background: white;
        cursor: pointer;
      }

      .form-select:focus {
        outline: none;
        border-color: #3498db;
      }

      .form-select:disabled {
        background: #f8f9fa;
        color: #6c757d;
        cursor: not-allowed;
      }
    `}</style>
  </>
)

export const FormButton = ({ 
  type = "submit", 
  disabled = false, 
  variant = "primary", 
  children, 
  onClick,
  className = "" 
}) => (
  <>
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      className={`form-button ${variant} ${className}`}
    >
      {children}
    </button>
    <style jsx>{`
      .form-button {
        border: none;
        padding: 12px 24px;
        border-radius: 8px;
        font-size: 1rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s;
      }

      .form-button.primary {
        background: linear-gradient(135deg, #3498db, #2980b9);
        color: white;
      }

      .form-button.primary:hover:not(:disabled) {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
      }

      .form-button.secondary {
        background: linear-gradient(135deg, #9b59b6, #8e44ad);
        color: white;
      }

      .form-button.secondary:hover:not(:disabled) {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(155, 89, 182, 0.3);
      }

      .form-button.success {
        background: linear-gradient(135deg, #27ae60, #229954);
        color: white;
      }

      .form-button.success:hover:not(:disabled) {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(39, 174, 96, 0.3);
      }

      .form-button:disabled {
        background: #bdc3c7;
        cursor: not-allowed;
        transform: none;
        box-shadow: none;
      }
    `}</style>
  </>
)

export default {
  FormContainer,
  FormGrid,
  FormGroup,
  FormInput,
  FormSelect,
  FormButton
}