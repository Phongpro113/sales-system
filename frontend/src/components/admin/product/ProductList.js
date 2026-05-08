import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table';
import axios from 'axios';
import { Edit, Plus, Trash2 } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import '../Admin.css';

const ProductList = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchProducts();
  }, []);

  const fetchProducts = async () => {
    try {
      const response = await axios.get('/api/admin/products', {
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
      });
      setData(response.data.products || []);
    } catch (error) {
      console.error('Error fetching products:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    if (window.confirm('Are you sure you want to delete this product?')) {
      try {
        await axios.delete(`/api/admin/products/${id}`, {
          headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
        });
        fetchProducts();
      } catch (error) {
        console.error('Error deleting product:', error);
      }
    }
  };

  const columns = [
    {
      accessorKey: 'id',
      header: 'ID',
    },
    {
      accessorKey: 'image_url',
      header: 'Image',
      cell: (info) => {
        const url = info.getValue();
        // Strict check: must be a string, length > 10, and contain a dot
        const isValidUrl = typeof url === 'string' && url.length > 10 && url.includes('.');

        if (!isValidUrl) {
          return (
            <div style={{ width: '40px', height: '40px', background: '#f1f5f9', borderRadius: '4px', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '10px', color: '#94a3b8' }}>
              No img
            </div>
          );
        }

        return (
          <img
            src={url}
            alt="Product"
            style={{ width: '40px', height: '40px', objectFit: 'cover', borderRadius: '4px' }}
            onError={(e) => {
              e.target.style.display = 'none';
            }}
          />
        );
      },
    },
    {
      accessorKey: 'name',
      header: 'Product Name',
      cell: (info) => <span style={{ fontWeight: 600 }}>{info.getValue()}</span>,
    },
    {
      accessorKey: 'sku',
      header: 'SKU',
      cell: (info) => <code style={{ fontSize: '0.875rem' }}>{info.getValue()}</code>,
    },
    {
      accessorKey: 'price',
      header: 'Price',
      cell: (info) => `$${parseFloat(info.getValue()).toLocaleString()}`,
    },
    {
      accessorKey: 'stock',
      header: 'Stock',
      cell: (info) => (
        <span className={`badge ${info.getValue() > 0 ? 'badge-success' : ''}`}>
          {info.getValue()} in stock
        </span>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: (info) => (
        <div className="action-buttons">
          <Link to={`/admin/product/${info.row.original.id}/edit`} className="btn-icon edit" title="Edit">
            <Edit size={16} />
          </Link>
          <button onClick={() => handleDelete(info.row.original.id)} className="btn-icon delete" title="Delete">
            <Trash2 size={16} />
          </button>
        </div>
      ),
    },
  ];

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div>
      <div className="admin-header">
        <h1 className="admin-title">Products</h1>
        <Link to="/admin/product/create" className="btn-primary">
          <Plus size={20} />
          Add Product
        </Link>
      </div>

      <div className="table-container">
        {loading ? (
          <div style={{ padding: '2rem', textAlign: 'center' }}>Loading products...</div>
        ) : (
          <table>
            <thead>
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {table.getRowModel().rows.map((row) => (
                <tr key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default ProductList;
