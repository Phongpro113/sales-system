package com.sales.payment.gateway.momo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class MomoResponse {

    @JsonProperty("partnerCode")
    private String partnerCode;

    @JsonProperty("requestId")
    private String requestId;

    @JsonProperty("orderId")
    private String orderId;

    @JsonProperty("amount")
    private long amount;

    @JsonProperty("responseTime")
    private long responseTime;

    @JsonProperty("message")
    private String message;

    @JsonProperty("resultCode")
    private int resultCode;

    @JsonProperty("payUrl")
    private String payUrl;

    @JsonProperty("deeplink")
    private String deeplink;

    @JsonProperty("qrCodeUrl")
    private String qrCodeUrl;

    public boolean isSuccess() {
        return resultCode == 0;
    }
}
