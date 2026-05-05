import axios from 'axios';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../contexts/CartContext';
import './ProductList.css';

const ProductList = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { addToCart } = useCart();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await axios.get('/api/products');
        const items = Array.isArray(response.data)
          ? response.data
          : Array.isArray(response.data?.products)
            ? response.data.products
            : [];
        setProducts(items);
      } catch (err) {
        setError('Failed to fetch products');
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  if (loading) {
    return <div className="loading">Loading products...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  if (products.length === 0) {
    return <div className="empty">No products available</div>;
  }

  return (
    <div className="product-list">
      <h2>Products</h2>
      <div className="products-grid">
        {products.map((product) => {
          const price = Number(product.price || 0);
          return (
            <div key={product.id} className="product-card">
              <img src={product.image_url || '/placeholder.png'} alt={product.name} />
              <h3>{product.name}</h3>
              <p className="description">{product.description}</p>
              <p className="price">${price.toFixed(2)}</p>
              <div className="actions">
                <button onClick={() => navigate(`/products/${product.id}`)}>Details</button>
                <button onClick={() => addToCart(product)}>Add to cart</button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ProductList;