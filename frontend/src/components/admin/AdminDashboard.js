import { DollarSign, Package, ShoppingCart, Users } from 'lucide-react';
import React from 'react';
import './Admin.css';

const AdminDashboard = () => {
  const stats = [
    { label: 'Total Revenue', value: '$12,450', icon: <DollarSign size={24} color="#10b981" />, trend: '+12.5%' },
    { label: 'Total Orders', value: '156', icon: <ShoppingCart size={24} color="#3b82f6" />, trend: '+8.2%' },
    { label: 'Total Products', value: '42', icon: <Package size={24} color="#f59e0b" />, trend: '+3 new' },
    { label: 'Active Users', value: '892', icon: <Users size={24} color="#8b5cf6" />, trend: '+5.4%' },
  ];

  return (
    <div>
      <div className="admin-header">
        <h1 className="admin-title">Dashboard Overview</h1>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.5rem', marginBottom: '2rem' }}>
        {stats.map((stat, i) => (
          <div key={i} className="form-card" style={{ padding: '1.5rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
              <div style={{ padding: '0.75rem', borderRadius: '0.75rem', backgroundColor: '#f8fafc' }}>
                {stat.icon}
              </div>
              <span style={{ fontSize: '0.875rem', fontWeight: 600, color: '#10b981' }}>{stat.trend}</span>
            </div>
            <h3 style={{ fontSize: '0.875rem', color: '#64748b', margin: '0 0 0.25rem 0' }}>{stat.label}</h3>
            <p style={{ fontSize: '1.5rem', fontWeight: 700, color: '#0f172a', margin: 0 }}>{stat.value}</p>
          </div>
        ))}
      </div>

      <div className="form-card" style={{ maxWidth: 'none' }}>
        <h3 style={{ marginBottom: '1.5rem' }}>Recent Activity</h3>
        <p style={{ color: '#64748b' }}>No recent activity to display.</p>
      </div>
    </div>
  );
};

export default AdminDashboard;
