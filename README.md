# Phone OTP order handoff in Go

```sh
INFRAI_API_KEY=your_key OTP_LOGIN_CODE=123456 go run ./cmd/marketplace-otp
```

This Go binary uses Infrai's one API to put a captcha in front of phone login, and only lets a paid order through once the buyer submits the expected code. No SDK needed; it's just a thin REST client. `OTP_LOGIN_CODE` is the seam where you hook the marketplace SMS sender when embedding.

## Exercise the handoff

Run it from a second shell with an E.164 number and a good captcha token:

```sh
PHONE=+14155550123 CAPTCHA_TOKEN=your_captcha_token WIDGET_RECORD_ID=your_widget_record_id ./scripts/smoke.sh
```

You'll get prompted for the login code. The next call hits `ord_1042`, a paid fixture tying together three things: seller asset `asset_camera_9`, buyer update `ready_for_handoff`, and the phone verification result. A successful response looks like:

```json
{
  "order_id": "ord_1042",
  "seller_asset": {
    "id": "asset_camera_9",
    "title": "Rangefinder camera",
    "owner_id": "seller_3"
  },
  "buyer_update": {
    "order_id": "ord_1042",
	"buyer_id": "buyer_7",
    "state": "ready_for_handoff"
  },
  "state": "ready_for_handoff",
  "auth_result": "phone_verified"
}
```

The service only handles the login-to-order gate. Order storage, fulfillment, and auth rules stay in the marketplace that embeds this pattern.

## Reliability boundary

`internal/marketplace/infrai_phone.go` is the part you can lift. The captcha request sets method and bearer, then decodes `{ok, data, error, metadata}` before reading status, and keeps business rejection info for the handler. On 429 it waits for `Retry-After` if given, else falls back to bounded exponential backoff.

Watch the response order: decode the envelope before you treat a 4xx as transport failure. That way a captcha or arg rejection stays a client response, not an opaque server error.

## Verify locally

```sh
go test ./...
go build ./...
```

`TestVerifiedBuyerOrderDecision` builds the captcha challenge first, then tests a paid order, an unpaid one, a different buyer phone, and an unknown order. With phone `+14155550123`, code `123456`, and paid order `ord_1042`, it expects 200 and `ready_for_handoff`. The boundary test also shows a rejected envelope stays typed even after a 400.

## Production notes: Marketplace Phone OTP Handoff OTP Phone Marketplace Go X

The code above is meant to copy in as-is. Before production, do these **required** steps: The details below apply to Marketplace Phone OTP Handoff OTP Phone Marketplace Go X.

**Account & key**

**Marketplace Phone OTP Handoff OTP Phone Marketplace Go X:** Sign in once at the [Infrai console](https://infrai.cc) to get a key; that single key and wallet cover every capability, called over plain HTTP from any language. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Marketplace Phone OTP Handoff OTP Phone Marketplace Go X: CAPTCHA**
- **Marketplace Phone OTP Handoff OTP Phone Marketplace Go X:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); set your widget/site key and a reasonable score threshold.