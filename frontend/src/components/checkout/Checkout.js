import axios from 'axios';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../../contexts/CartContext';
import FormField from '../ui/FormField';
import Input, { Textarea } from '../ui/Input';
import './Checkout.css';

const PAYMENT_METHODS = [
  { value: 'cod', label: 'Cash on Delivery', icon: '/images/icon/cash.svg' },
  { value: 'bank_transfer', label: 'Bank Transfer', icon: '/images/icon/bank-transfer.svg' },
  { value: 'momo', label: 'MoMo', icon: '/images/icon/momo.svg' },
];

const Checkout = () => {
  const { cart, getTotal, clearCart } = useCart();
  const navigate = useNavigate();

  const [form, setForm] = useState({
    fullName: '',
    phone: '',
    address: '',
    city: '',
    postalCode: '',
    notes: '',
    paymentMethod: 'cod',
  });
  const [errors, setErrors] = useState({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');

  const validate = () => {
    const next = {};
    if (!form.fullName.trim()) next.fullName = 'Full name is required';
    if (!form.phone.trim()) next.phone = 'Phone number is required';
    else if (!/^\+?[\d\s\-]{7,15}$/.test(form.phone.trim())) next.phone = 'Invalid phone number';
    if (!form.address.trim()) next.address = 'Address is required';
    if (!form.city.trim()) next.city = 'City is required';
    return next;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setForm(prev => ({ ...prev, [name]: value }));
    if (errors[name]) setErrors(prev => ({ ...prev, [name]: '' }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitError('');
    if (cart.length === 0) return;

    const validationErrors = validate();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setSubmitting(true);
    try {
      const shippingAddress = [form.fullName, form.phone, form.address, form.city, form.postalCode]
        .filter(Boolean)
        .join(', ');

      const res = await axios.post('/api/orders', {
        items: cart.map(item => ({ product_id: item.id, quantity: item.quantity })),
        payment_method: form.paymentMethod,
        shipping_address: shippingAddress,
        notes: form.notes,
      });

      const orderId = res.data.id;
      const total = getTotal();
      clearCart();

      if (form.paymentMethod === 'momo') {
        const payRes = await axios.post('/api/payments', {
          orderId,
          amount: Math.round(total),
          method: 'MOMO',
        });
        navigate('/checkout/momo', {
          state: { orderId, amount: Math.round(total), qrCodeUrl: payRes.data.qrCodeUrl, payUrl: payRes.data.payUrl },
        });
      } else {
        navigate('/orders');
      }
    } catch (err) {
      console.error('Order error:', err.response?.status, JSON.stringify(err.response?.data));
      const data = err.response?.data;
      setSubmitError(
        typeof data === 'string' ? data.trim()
        : data?.message || data?.error || 'Đặt hàng thất bại. Vui lòng thử lại.'
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form className="checkout" onSubmit={handleSubmit} noValidate>
      {/* Left — shipping + payment */}
      <div className="checkout-form-section">
        <h2>Shipping Information</h2>

        <div className="form-row">
          <FormField label="Full Name" error={errors.fullName} required>
            <Input name="fullName" value={form.fullName} onChange={handleChange} placeholder="Nguyen Van A" />
          </FormField>

          <FormField label="Phone Number" error={errors.phone} required>
            <Input type="tel" name="phone" value={form.phone} onChange={handleChange} placeholder="0912 345 678" />
          </FormField>
        </div>

        <FormField label="Street Address" error={errors.address} required>
          <Input name="address" value={form.address} onChange={handleChange} placeholder="123 Le Loi Street, District 1" />
        </FormField>

        <div className="form-row">
          <FormField label="City" error={errors.city} required>
            <Input name="city" value={form.city} onChange={handleChange} placeholder="Ho Chi Minh City" />
          </FormField>

          <FormField label="Postal Code">
            <Input name="postalCode" value={form.postalCode} onChange={handleChange} placeholder="700000" />
          </FormField>
        </div>

        <FormField label="Order Notes">
          <Textarea name="notes" value={form.notes} onChange={handleChange} placeholder="Delivery instructions, gate code, etc." />
        </FormField>

        <hr className="section-divider" />

        <h2>Payment Method</h2>
        <div className="payment-options">
          {PAYMENT_METHODS.map(({ value, label, icon }) => (
            <label key={value} className={`payment-option${form.paymentMethod === value ? ' selected' : ''}`}>
              <input type="radio" name="paymentMethod" value={value} checked={form.paymentMethod === value} onChange={handleChange} />
              <img src={icon} alt={label} className="payment-icon" />
              <span>{label}</span>
            </label>
          ))}
        </div>
      </div>

      {/* Right — order summary */}
      <div className="checkout-summary">
        <h3>Order Summary</h3>
        <div className="summary-items">
          {cart.map(item => (
            <div key={item.id} className="summary-item">
              <img src={item.image_url || '/placeholder.png'} alt={item.name} />
              <div className="summary-item-info">
                <div className="summary-item-name">{item.name}</div>
                <div className="summary-item-qty">x{item.quantity}</div>
              </div>
              <div className="summary-item-price">${(item.price * item.quantity).toFixed(2)}</div>
            </div>
          ))}
        </div>
        <hr className="summary-divider" />
        <div className="summary-total-row">
          <span>Total</span>
          <span>${getTotal().toFixed(2)}</span>
        </div>

        <button type="submit" className="place-order-btn" disabled={submitting}>
          {submitting ? 'Placing Order...' : 'Place Order'}
        </button>

        {submitError && <div className="submit-error">{submitError}</div>}
      </div>
    </form>
  );
};

export default Checkout;
