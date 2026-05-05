import axios from 'axios';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useCart } from '../contexts/CartContext';

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

  return (
    <div>
      <h2>{product.name}</h2>
      <p>{product.description}</p>
      <p>${price.toFixed(2)}</p>
      <button onClick={() => addToCart(product)}>Add to cart</button>
    </div>
  );
};

export default ProductDetail;
