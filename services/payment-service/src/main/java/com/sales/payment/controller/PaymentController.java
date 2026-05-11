package com.sales.payment.controller;

import com.sales.payment.dto.CreatePaymentRequest;
import com.sales.payment.dto.PaymentResponse;
import com.sales.payment.gateway.momo.MomoIpnRequest;
import com.sales.payment.service.PaymentService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/api/payments")
@RequiredArgsConstructor
public class PaymentController {

    private final PaymentService paymentService;

    // Called by frontend after order is created
    @PostMapping
    public ResponseEntity<PaymentResponse> createPayment(
            @RequestHeader("X-User-ID") Long userId,
            @Valid @RequestBody CreatePaymentRequest request
    ) {
        PaymentResponse response = paymentService.createPayment(userId, request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/order/{orderId}")
    public ResponseEntity<List<PaymentResponse>> getByOrder(@PathVariable Long orderId) {
        return ResponseEntity.ok(paymentService.getPaymentsByOrder(orderId));
    }

    @GetMapping("/{id}")
    public ResponseEntity<PaymentResponse> getPayment(@PathVariable Long id) {
        return ResponseEntity.ok(paymentService.getPayment(id));
    }

    // MoMo IPN callback — called by MoMo server after payment
    @PostMapping("/momo/ipn")
    public ResponseEntity<Map<String, String>> momoIpn(@RequestBody MomoIpnRequest ipn) {
        log.info("Received MoMo IPN: orderId={}, resultCode={}", ipn.getOrderId(), ipn.getResultCode());
        paymentService.handleMomoIpn(ipn);
        return ResponseEntity.ok(Map.of("message", "ok"));
    }

    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "healthy", "service", "payment-service"));
    }
}
