import { Navigate, Route, BrowserRouter as Router, Routes } from 'react-router-dom';
import './App.css';
import Cart from './components/Cart';
import Login from './components/Login';
import Navbar from './components/Navbar';
import Orders from './components/Orders';
import PrivateRoute from './components/PrivateRoute';
import ProductDetail from './components/ProductDetail';
import ProductList from './components/ProductList';
import Register from './components/Register';
import { AuthProvider } from './contexts/AuthContext';
import { CartProvider } from './contexts/CartContext';
import AdminRoute from './components/AdminRoute';
import AdminLayout from './components/admin/AdminLayout';
import AdminDashboard from './components/admin/AdminDashboard';
import AdminProductList from './components/admin/ProductList';
import ProductForm from './components/admin/ProductForm';

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