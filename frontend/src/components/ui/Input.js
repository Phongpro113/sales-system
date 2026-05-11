import './Input.css';

const Input = ({ type = 'text', ...props }) => (
  <input className="ui-input" type={type} {...props} />
);

export const Textarea = (props) => (
  <textarea className="ui-input ui-textarea" {...props} />
);

export default Input;
