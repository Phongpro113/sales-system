import { Link } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { useCart } from '../../contexts/CartContext';
import './Navbar.css';

const Navbar = () => {
  const { user, logout } = useAuth();
  const { cart } = useCart();

  const itemCount = cart.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
          <Link to="/products" className="navbar-brand">SalesSystem</Link>
          <div className="navbar-links">
            <Link to="/products" className="navbar-link">Products</Link>
            <Link to="/cart" className="navbar-link">
              Cart
              {itemCount > 0 && <span className="cart-badge">{itemCount}</span>}
            </Link>
            <Link to="/orders" className="navbar-link">Orders</Link>
            {user?.role === 'admin' && (
              <Link to="/admin" className="navbar-link" style={{ color: '#3182ce' }}>Admin Panel</Link>
            )}
          </div>
        </div>

        <div className="navbar-actions">
          {user ? (
            <div className="user-info">
              <div className="user-avatar">
                {user.name?.charAt(0).toUpperCase()}
              </div>
              <span>{user.name}</span>
              <button onClick={logout} className="btn-logout">Logout</button>
            </div>
          ) : (
            <>
              <Link to="/login" className="btn-login">Login</Link>
              <Link to="/register" className="btn-register">Sign up</Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
