# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.4] - 2026-08-27

This repository was extracted from a private monorepo with squashed history, so this entry
summarizes the public surface as of this tag rather than a granular diff from v0.1.0:

- Full `apify.epayco.co` surface: `Charges`, `Tokens`, `Customers`, `PSE`, `Cash`, `Daviplata`,
  `Safetypay`, `Standard`.
- `api.secure.payco.co` surface for recurring billing: `Plans`, `Subscriptions`.
- Webhook payload parsing and signature verification (`ParseWebhookPayload`, `VerifySignature`).
- Structured `*epayco.Error` on every failed call, including apify's HTTP-200-with-`success:false`
  responses.
- Runnable example under `examples/charge`.

## [0.1.0] - 2026-07-15

- Initial release: core client, apify resources, webhooks — live-tested against the ePayco
  sandbox.
