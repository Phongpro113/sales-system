import axios from 'axios';
import { ArrowLeft, Save } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { Toast } from 'primereact/toast';
import { Image } from 'primereact/image';
import '../Admin.css';

const ProductForm = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEdit = !!id;
  const toast = React.useRef(null);

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    price: '',
    stock: '',
    image_url: '',
    sku: '',
  });
  const [selectedFile, setSelectedFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [loading, setLoading] = useState(isEdit);
  const [error, setError] = useState('');

  useEffect(() => {
    if (isEdit) {
      fetchProduct();
    }
  }, [id]);

  const fetchProduct = async () => {
    try {
      const response = await axios.get(`/api/admin/products/${id}`, {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
      });
      const product = response.data;
      setFormData({
        name: product.name,
        description: product.description || '',
        price: product.price,
        stock: product.stock,
        image_url: product.image_url || '',
        sku: product.sku || '',
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

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setSelectedFile(file);
      setPreviewUrl(URL.createObjectURL(file));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const data = new FormData();
    data.append('name', formData.name);
    data.append('description', formData.description);
    data.append('price', formData.price);
    data.append('stock', formData.stock);
    data.append('sku', formData.sku);
    
    if (selectedFile) {
      data.append('image', selectedFile);
    }

    try {
      const config = {
        headers: { 
          'Content-Type': 'multipart/form-data',
          Authorization: `Bearer ${localStorage.getItem('token')}` 
        }
      };

      if (isEdit) {
        await axios.put(`/api/admin/products/${id}`, data, config);
      } else {
        await axios.post(`/api/admin/products`, data, config);
      }

      toast.current.show({ severity: 'success', summary: 'Success', detail: 'Product saved successfully' });
      navigate('/admin/product');
    } catch (error) {
      console.error('Error saving product:', error);
      setError(error.response?.data?.error || 'Failed to save product.');
    }
  };

  if (loading) return <div className="admin-main">Loading...</div>;

  return (
    <div>
      <Toast ref={toast} />
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
            <label className="form-label">SKU</label>
            <input
              type="text"
              name="sku"
              className="form-input"
              value={formData.sku}
              onChange={handleChange}
              required
              placeholder="e.g. PROD-001"
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

          <div className="form-group">
            <label className="form-label">Product Image</label>
            <div className="image-upload-container" style={{ marginBottom: '1.5rem' }}>
              {(previewUrl || (formData.image_url && formData.image_url.includes('.'))) && (
                <div style={{ marginBottom: '1rem' }}>
                  <Image 
                    src={previewUrl || formData.image_url} 
                    alt="Product" 
                    width="150" 
                    preview 
                  />
                </div>
              )}
              <input 
                type="file" 
                accept="image/*" 
                onChange={handleFileChange}
                className="form-input"
                style={{ padding: '0.5rem' }}
              />
              <p style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.5rem' }}>
                Selected file: {selectedFile ? selectedFile.name : 'No file chosen'}
              </p>
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
