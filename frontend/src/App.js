import { Navigate, Route, BrowserRouter as Router, Routes } from 'react-router-dom';
import './App.css';
import Cart from './components/cart/Cart';
import Checkout from './components/checkout/Checkout';
import Login from './components/auth/Login';
import Navbar from './components/layout/Navbar';
import Orders from './components/order/Orders';
import PrivateRoute from './components/auth/PrivateRoute';
import ProductDetail from './components/product/ProductDetail';
import ProductList from './components/product/ProductList';
import Register from './components/auth/Register';
import { AuthProvider } from './contexts/AuthContext';
import { CartProvider } from './contexts/CartContext';
import AdminRoute from './components/auth/AdminRoute';
import AdminLayout from './components/admin/AdminLayout';
import AdminDashboard from './components/admin/AdminDashboard';
import AdminProductList from './components/admin/product/ProductList';
import ProductForm from './components/admin/product/ProductForm';

function App() {
  return (
    <AuthProvider>
      <CartProvider>
        <Router>
          <div className="App">
            <Navbar />
            <div className="container">
              <Routes>
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />
                <Route path="/products" element={<ProductList />} />
                <Route path="/products/:id" element={<ProductDetail />} />
                <Route path="/cart" element={<PrivateRoute><Cart /></PrivateRoute>} />
                <Route path="/checkout" element={<PrivateRoute><Checkout /></PrivateRoute>} />
                <Route path="/orders" element={<PrivateRoute><Orders /></PrivateRoute>} />
                
                {/* Admin Routes */}
                <Route path="/admin" element={<AdminRoute><AdminLayout /></AdminRoute>}>
                  <Route index element={<AdminDashboard />} />
                  <Route path="product" element={<AdminProductList />} />
                  <Route path="product/create" element={<ProductForm />} />
                  <Route path="product/:id/edit" element={<ProductForm />} />
                </Route>

                <Route path="/" element={<Navigate to="/products" />} />
              </Routes>
            </div>
          </div>
        </Router>
      </CartProvider>
    </AuthProvider>
  );
}

export default App;