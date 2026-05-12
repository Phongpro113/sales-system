import { useLocation, useNavigate } from 'react-router-dom';
import MomoPayment from './MomoPayment';

const MomoPaymentPage = () => {
  const { state } = useLocation();
  const navigate = useNavigate();

  if (!state?.orderId) {
    navigate('/orders');
    return null;
  }

  return <MomoPayment orderId={state.orderId} amount={state.amount} qrCodeUrl={state.qrCodeUrl} payUrl={state.payUrl} />;
};

export default MomoPaymentPage;
