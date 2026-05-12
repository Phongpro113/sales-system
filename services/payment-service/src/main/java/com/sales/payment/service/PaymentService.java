package com.sales.payment.service;

import com.sales.payment.dto.CreatePaymentRequest;
import com.sales.payment.dto.PaymentResponse;
import com.sales.payment.entity.Payment;
import com.sales.payment.enums.PaymentMethod;
import com.sales.payment.enums.PaymentStatus;
import com.sales.payment.exception.PaymentException;
import com.sales.payment.gateway.momo.MomoGateway;
import com.sales.payment.gateway.momo.MomoIpnRequest;
import com.sales.payment.gateway.momo.MomoResponse;
import com.sales.payment.repository.PaymentRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class PaymentService {

    private final PaymentRepository paymentRepository;
    private final MomoGateway momoGateway;

    @Transactional
    public PaymentResponse createPayment(Long userId, CreatePaymentRequest req) {
        Payment payment = Payment.builder()
                .orderId(req.getOrderId())
                .userId(userId)
                .amount(req.getAmount())
                .method(req.getMethod())
                .status(PaymentStatus.PENDING)
                .build();

        if (req.getMethod() == PaymentMethod.MOMO) {
            payment = initiateMomoPayment(payment);
        } else {
            // COD and bank transfer are confirmed on delivery / manual verification
            payment.setStatus(PaymentStatus.PENDING);
        }

        return new PaymentResponse(paymentRepository.save(payment));
    }

    public List<PaymentResponse> getPaymentsByOrder(Long orderId) {
        return paymentRepository.findByOrderId(orderId)
                .stream()
                .map(PaymentResponse::new)
                .toList();
    }

    public PaymentResponse getPayment(Long id) {
        return paymentRepository.findById(id)
                .map(PaymentResponse::new)
                .orElseThrow(() -> new PaymentException("Payment not found: " + id));
    }

    @Transactional
    public void handleMomoIpn(MomoIpnRequest ipn) {
        if (!momoGateway.verifyIpnSignature(ipn)) {
            log.warn("Invalid MoMo IPN signature for orderId={}", ipn.getOrderId());
            throw new PaymentException("Invalid IPN signature");
        }

        Payment payment = paymentRepository.findByGatewayRef(ipn.getRequestId())
                .orElseThrow(() -> new PaymentException("Payment not found for requestId: " + ipn.getRequestId()));

        if (ipn.isSuccess()) {
            payment.setStatus(PaymentStatus.SUCCESS);
            payment.setTransactionId(String.valueOf(ipn.getTransId()));
            log.info("MoMo payment SUCCESS for orderId={}, transId={}", payment.getOrderId(), ipn.getTransId());
        } else {
            payment.setStatus(PaymentStatus.FAILED);
            payment.setFailureReason(ipn.getMessage());
            log.warn("MoMo payment FAILED for orderId={}: {}", payment.getOrderId(), ipn.getMessage());
        }

        paymentRepository.save(payment);
    }

    private Payment initiateMomoPayment(Payment payment) {
        String orderInfo = "Thanh toan don hang #" + payment.getOrderId();
        long amountVnd = payment.getAmount().longValue();

        MomoResponse momoResponse = momoGateway.createPayment(
                payment.getOrderId(), amountVnd, orderInfo
        );

        payment.setGatewayRef(momoResponse.getRequestId());
        payment.setPayUrl(momoResponse.getPayUrl());
        payment.setQrCodeUrl(momoResponse.getQrCodeUrl());
        return payment;
    }
}
