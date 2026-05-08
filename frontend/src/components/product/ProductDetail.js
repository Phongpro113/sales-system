import axios from 'axios';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useCart } from '../../contexts/CartContext';
import './ProductDetail.css';

const ProductDetail = () => {
  const { id } = useParams();
  const { addToCart } = useCart();
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchProduct = async () => {
      try {
        const response = await axios.get(`/api/products/${id}`);
        setProduct(response.data?.product || response.data);
      } catch (err) {
        setError('Failed to fetch product');
      } finally {
        setLoading(false);
      }
    };
    fetchProduct();
  }, [id]);

  if (loading) return <div className="loading">Loading product...</div>;
  if (error) return <div className="error">{error}</div>;
  if (!product) return <div className="empty">Product not found</div>;

  const price = Number(product.price || 0);
  const stock = Number(product.stock || 0);

  return (
    <div className="product-detail">
      <h2>{product.name}</h2>
      <img src={product.image_url || '/placeholder.png'} alt={product.name} />
      <p className="detail-description">{product.description}</p>
      <p className="detail-price">Price: ${price.toFixed(2)}</p>
      <p className="detail-stock">Stock: {stock}</p>
      <div className="detail-actions">
        <button className="btn-secondary" onClick={() => addToCart(product)}>Add to cart</button>
        <button className="btn-primary" onClick={() => addToCart(product)}>Buy now</button>
      </div>
    </div>
  );
};

export default ProductDetail;
