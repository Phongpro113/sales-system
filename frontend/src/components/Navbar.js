import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useCart } from '../contexts/CartContext';

const Navbar = () => {
  const { user, logout } = useAuth();
  const { cart } = useCart();

  const itemCount = cart.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <nav style={{ background: '#111827', color: '#fff', padding: '0.75rem 1rem' }}>
      <div style={{ maxWidth: 1100, margin: '0 auto', display: 'flex', gap: '1rem', alignItems: 'center' }}>
        <Link to="/products" style={{ color: '#fff', textDecoration: 'none', fontWeight: 700 }}>Sales System</Link>
        <Link to="/products" style={{ color: '#fff' }}>Products</Link>
        <Link to="/cart" style={{ color: '#fff' }}>Cart ({itemCount})</Link>
        <Link to="/orders" style={{ color: '#fff' }}>Orders</Link>
        <div style={{ marginLeft: 'auto' }}>
          {user ? (
            <button onClick={logout}>Logout</button>
          ) : (
            <>
              <Link to="/login" style={{ color: '#fff', marginRight: 12 }}>Login</Link>
              <Link to="/register" style={{ color: '#fff' }}>Register</Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
