import axios from 'axios';
import { ArrowLeft, Save } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import './Admin.css';

const ProductForm = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEdit = !!id;

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    price: '',
    stock: '',
  });
  const [loading, setLoading] = useState(isEdit);
  const [error, setError] = useState('');

  useEffect(() => {
    if (isEdit) {
      fetchProduct();
    }
  }, [id]);

  const fetchProduct = async () => {
    try {
      const response = await axios.get(`http://localhost:8080/api/admin/products/${id}`, {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
      });
      const product = response.data;
      setFormData({
        name: product.name,
        description: product.description || '',
        price: product.price,
        stock: product.stock,
      });
    } catch (error) {
      console.error('Error fetching product:', error);
      setError('Failed to load product data.');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: name === 'price' || name === 'stock' ? parseFloat(value) || 0 : value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    try {
      const config = {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
      };

      if (isEdit) {
        await axios.put(`http://localhost:8080/api/admin/products/${id}`, formData, config);
      } else {
        await axios.post('http://localhost:8080/api/admin/products', formData, config);
      }

      navigate('/admin/product');
    } catch (error) {
      console.error('Error saving product:', error);
      setError(error.response?.data?.error || 'Failed to save product.');
    }
  };

  if (loading) return <div className="admin-main">Loading...</div>;

  return (
    <div>
      <div className="admin-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <Link to="/admin/product" className="btn-icon">
            <ArrowLeft size={20} />
          </Link>
          <h1 className="admin-title">{isEdit ? 'Edit Product' : 'Create Product'}</h1>
        </div>
      </div>

      <div className="form-card">
        {error && <div style={{ color: '#dc2626', marginBottom: '1rem', fontSize: '0.875rem' }}>{error}</div>}
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">Product Name</label>
            <input
              type="text"
              name="name"
              className="form-input"
              value={formData.name}
              onChange={handleChange}
              required
              placeholder="Enter product name"
            />
          </div>

          <div className="form-group">
            <label className="form-label">Description</label>
            <textarea
              name="description"
              className="form-input"
              rows="4"
              value={formData.description}
              onChange={handleChange}
              placeholder="Enter product description"
            />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
            <div className="form-group">
              <label className="form-label">Price ($)</label>
              <input
                type="number"
                name="price"
                step="0.01"
                className="form-input"
                value={formData.price}
                onChange={handleChange}
                required
                placeholder="0.00"
              />
            </div>

            <div className="form-group">
              <label className="form-label">Stock Quantity</label>
              <input
                type="number"
                name="stock"
                className="form-input"
                value={formData.stock}
                onChange={handleChange}
                required
                placeholder="0"
              />
            </div>
          </div>

          <div className="form-actions">
            <Link to="/admin/product" className="btn-secondary">
              Cancel
            </Link>
            <button type="submit" className="btn-primary">
              <Save size={18} />
              {isEdit ? 'Update Product' : 'Create Product'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default ProductForm;
