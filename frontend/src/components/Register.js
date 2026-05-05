import { useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const Register = () => {
  const { user, register } = useAuth();
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  if (user) return <Navigate to="/products" replace />;

  const onSubmit = async (e) => {
    e.preventDefault();
    const result = await register(email, password, name);
    if (result.success) {
      navigate('/products');
    } else {
      setError(result.error || 'Registration failed');
    }
  };

  return (
    <form onSubmit={onSubmit} style={{ maxWidth: 420, margin: '2rem auto', display: 'grid', gap: 12 }}>
      <h2>Register</h2>
      {error && <div className="error">{error}</div>}
      <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Name" />
      <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" />
      <input value={password} onChange={(e) => setPassword(e.target.value)} type="password" placeholder="Password" />
      <button type="submit">Create account</button>
    </form>
  );
};

export default Register;
