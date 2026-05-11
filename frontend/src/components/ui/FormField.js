import './FormField.css';

const FormField = ({ label, error, required, children }) => (
  <div className={`form-group${error ? ' has-error' : ''}`}>
    {label && (
      <label>
        {label}
        {required && <span className="required-mark"> *</span>}
      </label>
    )}
    {children}
    {error && <span className="field-error">{error}</span>}
  </div>
);

export default FormField;
