# Checkout flow status — where it completes vs. fails

Confirmed live, 2026-08-01, against all three configured UCP merchants using the exact
request shapes `internal/ucp.Client.CreateCheckout` / `CompleteCheckout` send
(`meta.ucp-agent.profile`, `checkout.line_items`, `checkout.payment.instruments` with
`handler_id: "shopify.card"`).

## Result: the wall is systemic, not merchant-specific

`complete_checkout` fails identically — same code, same message, same data string — on
all three merchants tested. This is a Shopify platform-level auth gate on the
`complete_checkout` tool, not something specific to one store's account/app config.

## Per-merchant results

| Merchant | Product | `create_checkout` | `complete_checkout` |
|---|---|---|---|
| Headphone Zone | Sennheiser HD 600 (`PRODUCT_CHOICE=headphones`) | ✅ succeeds — returns real checkout `id`, status `requires_escalation` | ❌ `-32000 AuthenticationRequired` |
| Bombay Shaving Company | Sensi Smart 3 Razor (`PRODUCT_CHOICE=shaving`) | ✅ succeeds — returns real checkout `id`, status `requires_escalation` | ❌ `-32000 AuthenticationRequired` |
| Mokobara | Kaleido Backpack - 28L (spot-checked, not wired into `cmd/pay`) | ✅ succeeds — returns real checkout `id`, status `requires_escalation` | ❌ `-32000 AuthenticationRequired` |

## The `complete_checkout` error — identical across all three

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32000,
    "message": "AuthenticationRequired",
    "data": "Unauthorized: A valid JWT is required to call complete_checkout. See https://shopify.dev/docs/agents/get-started/authentication for instructions on generating a token and authenticating your requests."
  }
}
```

Byte-for-byte identical `code`, `message`, and `data` on Headphone Zone, Bombay Shaving,
and Mokobara. No merchant returned a different error shape, a different reason, or a
partial success. This matches Shopify's own documented split: `search_catalog`,
`get_product`, and `create_checkout` only require the UCP agent-profile handshake
(`meta.ucp-agent.profile`); `complete_checkout` requires a separate Shopify Dev
Dashboard JWT (client-id/secret, `POST /auth/access_token`) scoped per developer/app —
see [docs/... authentication reference](https://shopify.dev/docs/agents/get-started/authentication),
confirmed earlier in this build.

## `create_checkout` — succeeds on all three, with one real per-merchant difference

Every merchant returns a valid checkout object (`status: "requires_escalation"`,
`isError:true` alongside it — expected, not fatal, since a missing extension
interaction/buyer input is still a normal checkout state). One difference worth noting:
Bombay Shaving's checkout also advertises a `com.google.pay` payment handler
(`gatewayMerchantId`, full Google Pay tokenization config) alongside `dev.shopify.card`;
Headphone Zone and Mokobara only advertised `dev.shopify.card`. Not relevant to the
`complete_checkout` auth wall (which fires before any handler-specific logic runs), but
a real per-merchant config difference if this project revisits payment handler selection
later.

## What this means for the project

Both `PRODUCT_CHOICE` paths (`headphones`, `shaving`) in `cmd/pay`/`internal/purchase`
behave identically once they reach step 3 (merchant checkout): `create_checkout`
succeeds, `complete_checkout` 401s, the failure is caught and reported to Prava as
`DECLINED` with the real reason — never hardcoded. This is expected, current, and
confirmed platform-wide behavior, not a bug in this codebase and not something fixable
by trying a different merchant. Unblocking it requires either:

1. Shopify Dev Dashboard credentials scoped to each merchant (unlikely to be obtainable
   for live third-party stores we don't operate), or
2. Pointing the flow at a Shopify dev/test store under our own account, where we could
   generate that JWT ourselves.
