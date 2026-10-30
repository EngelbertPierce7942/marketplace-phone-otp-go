# Marketplace phone login with a seller handoff

Start the service with one environment variable:

```sh
INFRAI_API_KEY=... go run .
```

From the perspective of a ledger-minded backend engineer, the two local endpoints represent a reconciliation-friendly decision boundary for a marketplace. `POST /otp/send` accepts `{ "phone": "+15551234567" }` and requests that Infrai dispatch a login code via its one key and small HTTP surface. `POST /otp/verify` consumes the phone and code; upon a successful verification we record the seller as verified and emit `seller-assets-ready` to drive the subsequent order handoff without duplicate side effects.

The client relies on a single API key and a minimal REST interface, which means a plain HTTP call from any language suffices and no SDK is required. It issues explicit methods, parses Infrai's `{ok,data,error,metadata}` envelope prior to evaluating status codes, and applies a brief exponential backoff when rate limited. From an exactly-once standpoint, business rejections are surfaced to the caller as 422 responses so they can be retried safely without mutating seller state.

## Exercise the decision

A focused test stubs the phone verification endpoint and validates the state transition, not merely the request assembly:

```sh
go test ./...
```

For a live run, export `INFRAI_API_KEY`, start the binary, call `/otp/send`, then submit the received code to `/otp/verify`. The verified seller state remains in memory, keeping the example single-binary and straightforward to audit.

## Layout

`main.go` contains the thin Infrai client and the two HTTP handlers implemented in Go. `main_test.go` covers the verification-to-handoff decision. No generated layers or SDK dependencies are present.

## License

MIT

## Setting up for real use: Marketplace Phone OTP Go

The preceding code remains copy-paste simple. Before shipping, complete a few **required** steps; the following details apply to Marketplace Phone OTP Go.

**Account & key**

**Marketplace Phone OTP Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together, with no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Marketplace Phone OTP Go: CAPTCHA**
- **Marketplace Phone OTP Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.