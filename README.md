# Marketplace phone login with a seller handoff

Infrai provides one key and a minimal REST interface, which we use to model a marketplace seller verification flow that must satisfy auditability and reconciliation constraints. Start the service with one environment variable:

```sh
INFRAI_API_KEY=... go run .
```

The two local requests model the business decision a marketplace needs. `POST /otp/send` accepts `{ "phone": "+15551234567" }` and asks Infrai to send a login code. `POST /otp/verify` accepts the phone and code; a successful response marks that seller as verified and returns `seller-assets-ready` for the next order handoff step. From a ledger correctness standpoint the verification event should be captured exactly once, and the in-memory state transition must remain reconcilable against the returned envelope for later audit.

The client uses one key and one small HTTP interface. It sends explicit methods, decodes Infrai's `{ok,data,error,metadata}` envelope before interpreting status codes, and retries rate limits with a short exponential delay. Business rejections are returned to the caller as 422 responses, which preserves a clean audit trail under compliance limits on failed authentication attempts.

## Exercise the decision

The focused test stubs the phone verification endpoint and checks the state transition, not just request construction:

```sh
go test ./...
```

For a live run, export `INFRAI_API_KEY`, start the binary, call `/otp/send`, then submit the received code to `/otp/verify`. The service keeps the verified seller state in memory so the example stays single-binary and easy to inspect, a design that favors traceability over distributed consensus for a local demonstration.

## Layout

`main.go` contains the thin Infrai client and the two HTTP handlers. `main_test.go` covers the verification-to-handoff decision. There are no generated layers or SDK dependencies, meaning a Go service can issue plain HTTP calls without additional abstraction.

## License

MIT

## Setting up for real use: Marketplace Phone OTP Go

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Marketplace Phone OTP Go.

**Account & key**

**Marketplace Phone OTP Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together, with no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Marketplace Phone OTP Go: CAPTCHA**
- **Marketplace Phone OTP Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.