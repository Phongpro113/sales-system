package com.sales.payment.gateway.momo;

import com.sales.payment.config.MomoProperties;
import com.sales.payment.exception.PaymentException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.*;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.util.HexFormat;
import java.util.UUID;

@Slf4j
@Component
@RequiredArgsConstructor
public class MomoGateway {

    private final MomoProperties props;
    private final RestTemplate restTemplate;

    public MomoResponse createPayment(long orderId, long amountVnd, String orderInfo) {
        String requestId = UUID.randomUUID().toString();
        String momoOrderId = "ORDER-" + orderId + "-" + System.currentTimeMillis();
        String extraData = "";

        String rawSignature = buildRawSignature(requestId, momoOrderId, amountVnd, orderInfo, extraData);
        String signature = hmacSha256(rawSignature, props.getSecretKey());

        MomoRequest request = MomoRequest.builder()
                .partnerCode(props.getPartnerCode())
                .accessKey(props.getAccessKey())
                .requestId(requestId)
                .amount(amountVnd)
                .orderId(momoOrderId)
                .orderInfo(orderInfo)
                .redirectUrl(props.getRedirectUrl())
                .ipnUrl(props.getIpnUrl())
                .extraData(extraData)
                .requestType(props.getRequestType())
                .signature(signature)
                .lang("vi")
                .build();

        log.info("Sending MoMo payment request for orderId={}", orderId);

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        ResponseEntity<MomoResponse> response = restTemplate.exchange(
                props.getEndpoint(),
                HttpMethod.POST,
                new HttpEntity<>(request, headers),
                MomoResponse.class
        );

        MomoResponse body = response.getBody();
        if (body == null) {
            throw new PaymentException("Empty response from MoMo");
        }
        if (!body.isSuccess()) {
            log.warn("MoMo payment failed: code={}, message={}", body.getResultCode(), body.getMessage());
            throw new PaymentException("MoMo error " + body.getResultCode() + ": " + body.getMessage());
        }

        log.info("MoMo payment created: payUrl={}", body.getPayUrl());
        return body;
    }

    public boolean verifyIpnSignature(MomoIpnRequest ipn) {
        String rawSignature = "accessKey=" + props.getAccessKey()
                + "&amount=" + ipn.getAmount()
                + "&extraData=" + ipn.getExtraData()
                + "&message=" + ipn.getMessage()
                + "&orderId=" + ipn.getOrderId()
                + "&orderInfo=" + ipn.getOrderInfo()
                + "&orderType=" + ipn.getOrderType()
                + "&partnerCode=" + ipn.getPartnerCode()
                + "&payType=" + ipn.getPayType()
                + "&requestId=" + ipn.getRequestId()
                + "&responseTime=" + ipn.getResponseTime()
                + "&resultCode=" + ipn.getResultCode()
                + "&transId=" + ipn.getTransId();

        String expected = hmacSha256(rawSignature, props.getSecretKey());
        return expected.equals(ipn.getSignature());
    }

    private String buildRawSignature(String requestId, String orderId, long amount, String orderInfo, String extraData) {
        return "accessKey=" + props.getAccessKey()
                + "&amount=" + amount
                + "&extraData=" + extraData
                + "&ipnUrl=" + props.getIpnUrl()
                + "&orderId=" + orderId
                + "&orderInfo=" + orderInfo
                + "&partnerCode=" + props.getPartnerCode()
                + "&redirectUrl=" + props.getRedirectUrl()
                + "&requestId=" + requestId
                + "&requestType=" + props.getRequestType();
    }

    private String hmacSha256(String data, String key) {
        try {
            Mac mac = Mac.getInstance("HmacSHA256");
            mac.init(new SecretKeySpec(key.getBytes(StandardCharsets.UTF_8), "HmacSHA256"));
            byte[] hash = mac.doFinal(data.getBytes(StandardCharsets.UTF_8));
            return HexFormat.of().formatHex(hash);
        } catch (Exception e) {
            throw new PaymentException("Failed to compute HMAC signature", e);
        }
    }
}
