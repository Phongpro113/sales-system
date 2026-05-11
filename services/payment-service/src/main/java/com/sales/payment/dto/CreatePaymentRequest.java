package com.sales.payment.dto;

import com.sales.payment.enums.PaymentMethod;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Getter;
import lombok.Setter;

import java.math.BigDecimal;

@Getter @Setter
public class CreatePaymentRequest {

    @NotNull
    private Long orderId;

    @NotNull
    @Positive
    private BigDecimal amount;

    @NotNull
    private PaymentMethod method;
}
