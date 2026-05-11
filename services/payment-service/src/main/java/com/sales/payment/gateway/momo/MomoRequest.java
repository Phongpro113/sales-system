package com.sales.payment.gateway.momo;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.*;

@Getter @Setter @NoArgsConstructor @AllArgsConstructor @Builder
public class MomoRequest {

    @JsonProperty("partnerCode")
    private String partnerCode;

    @JsonProperty("accessKey")
    private String accessKey;

    @JsonProperty("requestId")
    private String requestId;

    @JsonProperty("amount")
    private long amount;

    @JsonProperty("orderId")
    private String orderId;

    @JsonProperty("orderInfo")
    private String orderInfo;

    @JsonProperty("redirectUrl")
    private String redirectUrl;

    @JsonProperty("ipnUrl")
    private String ipnUrl;

    @JsonProperty("extraData")
    private String extraData;

    @JsonProperty("requestType")
    private String requestType;

    @JsonProperty("signature")
    private String signature;

    @JsonProperty("lang")
    private String lang;
}
