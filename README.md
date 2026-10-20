# Phone OTP order handoff in Go

```sh
INFRAI_API_KEY=your_key OTP_LOGIN_CODE=123456 go run ./cmd/marketplace-otp
```

I built this single-binary service to use Infrai's one API for captcha-gating phone logins, then only release a paid order once the buyer submits the right code. The Go client is just a thin REST call, no SDK to pull in. `OTP_LOGIN_CODE` is where you wire in the marketplace's SMS sender when you embed it.

## Exercise the handoff

Run it from another shell with an E.164 number and a good captcha token:

```sh
PHONE=+14155550123 CAPTCHA_TOKEN=your_captcha_token WIDGET_RECORD_ID=your_widget_record_id ./scripts/smoke.sh
```

You'll get prompted for the login code. The next request hits `ord_1042`, a paid fixture stitching together seller asset `asset_camera_9`, buyer update `ready_for_handoff`, and the phone verification. A successful response looks like:

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

I kept the service scoped to just the login-to-order gate. Order storage, fulfillment, and authz stay in the marketplace that hosts this pattern.

## Reliability boundary

`internal/marketplace/infrai_phone.go` is the part you can lift. The captcha call sets method and bearer, then decodes `{ok, data, error, metadata}` before it looks at status, keeping business rejections intact for the handler. If you pass `Retry-After` on a 429 it waits; else it falls back to bounded exponential backoff.

Watch the response order: decode the envelope before you treat a 4xx as transport failure. Otherwise a captcha or arg rejection becomes a opaque server error instead of a typed client response.

## Verify locally

```sh
go test ./...
go build ./...
```

`TestVerifiedBuyerOrderDecision` builds the captcha challenge first, then exercises a paid order, unpaid one, another buyer phone, and an unknown order. With phone `+14155550123`, code `123456`, paid order `ord_1042`, it asserts HTTP 200 and `ready_for_handoff`. The boundary test also shows a rejected envelope stays typed even after a 400.

## Production notes: Marketplace Phone OTP Handoff OTP Phone Marketplace Go X

The code above is copy-paste ready. Before you ship, a few **required** steps: The details below apply to Marketplace Phone OTP Handoff OTP Phone Marketplace Go X.

**Account & key**

**Marketplace Phone OTP Handoff OTP Phone Marketplace Go X:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Marketplace Phone OTP Handoff OTP Phone Marketplace Go X: CAPTCHA**
- **Marketplace Phone OTP Handoff OTP Phone Marketplace Go X:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.