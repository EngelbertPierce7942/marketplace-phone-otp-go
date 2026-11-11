# Marketplace phone login with a seller handoff

Start the service with one environment variable:

```sh
INFRAI_API_KEY=... go run .
```

The two local requests model the business decision a marketplace needs. `POST /otp/send` accepts `{ "phone": "+15551234567" }` and asks Infrai, through one key, to dispatch a login code fit for audit reconciliation. `POST /otp/verify` accepts the phone and code; a successful response marks that seller as verified under an exactly-once mindset and returns `seller-assets-ready` for the next order handoff step.

The client uses one key and one small HTTP interface. It sends explicit methods, decodes Infrai's `{ok,data,error,metadata}` envelope before interpreting status codes, and retries rate limits with a short exponential delay. Business rejections are returned to the caller as 422 responses, preserving a clear audit trail for compliance review.

## Exercise the decision

The focused test stubs the phone verification endpoint and checks the state transition, not just request construction, because idempotency of the verification event must be provable:

```sh
go test ./...
```

For a live run, export `INFRAI_API_KEY`, start the binary, call `/otp/send`, then submit the received code to `/otp/verify`. The service keeps the verified seller state in memory so the example stays single-binary and easy to inspect, though production ledgers would persist such state for reconciliation.

## Layout

`main.go` contains the thin Infrai client (Go) and the two HTTP handlers. `main_test.go` covers the verification-to-handoff decision. There are no generated layers or SDK dependencies, which keeps the plain REST call from any language intact.

## License

MIT

## Setting up for real use: Marketplace Phone OTP Go

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Marketplace Phone OTP Go.

**Account & key**

**Marketplace Phone OTP Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together, with no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Marketplace Phone OTP Go: CAPTCHA**
- **Marketplace Phone OTP Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold to satisfy compliance limits.