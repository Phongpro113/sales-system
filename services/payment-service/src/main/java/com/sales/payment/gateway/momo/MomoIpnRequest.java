package com.sales.payment.gateway.momo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class MomoIpnRequest {

    @JsonProperty("partnerCode")
    private String partnerCode;

    @JsonProperty("orderId")
    private String orderId;

    @JsonProperty("requestId")
    private String requestId;

    @JsonProperty("amount")
    private long amount;

    @JsonProperty("orderInfo")
    private String orderInfo;

    @JsonProperty("orderType")
    private String orderType;

    @JsonProperty("transId")
    private long transId;

    @JsonProperty("resultCode")
    private int resultCode;

    @JsonProperty("message")
    private String message;

    @JsonProperty("payType")
    private String payType;

    @JsonProperty("responseTime")
    private long responseTime;

    @JsonProperty("extraData")
    private String extraData;

    @JsonProperty("signature")
    private String signature;

    public boolean isSuccess() {
        return resultCode == 0;
    }
}
