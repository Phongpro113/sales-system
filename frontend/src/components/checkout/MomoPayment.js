import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './MomoPayment.css';

const MomoPayment = ({ orderId, amount, qrCodeUrl, payUrl }) => {
  const navigate = useNavigate();
  const [copied, setCopied] = useState('');

  const copy = (text, field) => {
    navigator.clipboard.writeText(text);
    setCopied(field);
    setTimeout(() => setCopied(''), 2000);
  };

  return (
    <div className="momo-payment">
      <div className="momo-header">
        <img src="/images/icon/momo.svg" alt="MoMo" className="momo-logo" />
        <h2>Thanh toán MoMo</h2>
      </div>

      <div className="momo-body">
        {/* QR chính thức từ MoMo server */}
        <div className="momo-qr-wrap">
          {qrCodeUrl ? (
            <img src={qrCodeUrl} alt="MoMo QR Code" className="momo-qr-image" />
          ) : (
            <p className="momo-qr-hint">Không tải được mã QR</p>
          )}
          <p className="momo-qr-hint">Mở app MoMo → Quét mã</p>
        </div>

        {payUrl && (
          <>
            <div className="momo-divider"><span>hoặc mở trực tiếp</span></div>
            <div className="momo-info">
              <div className="momo-row">
                <span className="momo-label">Link thanh toán</span>
                <div className="momo-value-wrap">
                  <a href={payUrl} className="momo-value momo-link" target="_blank" rel="noreferrer">
                    Mở trang MoMo
                  </a>
                  <button className="copy-btn" onClick={() => copy(payUrl, 'payUrl')}>
                    {copied === 'payUrl' ? '✓ Đã chép' : 'Sao chép'}
                  </button>
                </div>
              </div>
            </div>
          </>
        )}

        <div className="momo-info" style={{ marginTop: '12px' }}>
          <div className="momo-row">
            <span className="momo-label">Số tiền</span>
            <span className="momo-value momo-amount">
              {Number(amount).toLocaleString('vi-VN')}₫
            </span>
          </div>
          <div className="momo-row">
            <span className="momo-label">Mã đơn hàng</span>
            <span className="momo-value">#{orderId}</span>
          </div>
        </div>

        <p className="momo-note">
          ⚠️ Vui lòng hoàn tất thanh toán để đơn hàng được xác nhận tự động.
        </p>
      </div>

      <div className="momo-footer">
        <button className="momo-done-btn" onClick={() => navigate('/orders')}>
          Đã thanh toán xong
        </button>
        <button className="momo-cancel-btn" onClick={() => navigate(-1)}>
          Quay lại
        </button>
      </div>
    </div>
  );
};

export default MomoPayment;
