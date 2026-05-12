package com.sales.payment.dto;

import com.sales.payment.entity.Payment;
import com.sales.payment.enums.PaymentMethod;
import com.sales.payment.enums.PaymentStatus;
import lombok.Getter;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Getter
public class PaymentResponse {

    private final Long id;
    private final Long orderId;
    private final BigDecimal amount;
    private final PaymentMethod method;
    private final PaymentStatus status;
    private final String payUrl;
    private final String qrCodeUrl;
    private final String transactionId;
    private final LocalDateTime createdAt;

    public PaymentResponse(Payment payment) {
        this.id = payment.getId();
        this.orderId = payment.getOrderId();
        this.amount = payment.getAmount();
        this.method = payment.getMethod();
        this.status = payment.getStatus();
        this.payUrl = payment.getPayUrl();
        this.qrCodeUrl = payment.getQrCodeUrl();
        this.transactionId = payment.getTransactionId();
        this.createdAt = payment.getCreatedAt();
    }
}
