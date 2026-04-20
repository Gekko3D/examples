# Feature Context: FEAT-1234 — Add Retry Logic for Pricing API Calls

## Metadata

| Field | Value |
|---|---|
| **Ticket** | [PROJ-1234](https://jira.company.com/browse/PROJ-1234) |
| **Status** | ACTIVE |
| **Author** | Jane Engineer |
| **Created** | 2026-02-15 |
| **Branch** | `feature/FEAT-1234-pricing-retry` |
| **PR** | — |

## Objective

Add circuit breaker and retry logic to Pricing Service calls to prevent cascading failures during Pricing Service degradation. Currently a single timeout causes a 500 error on quote creation.

## Scope

**In Scope:** Circuit breaker on Pricing gRPC calls, configurable retry with backoff, fallback to cached prices, health check endpoint, alerting on circuit open.

**Out of Scope:** REST migration (separate initiative), cache warming strategy (FEAT-1300).

## Design Summary

Resilience4j circuit breaker wrapping `PricingClient.getPrice()`. On 5 consecutive failures, circuit opens and falls back to Redis-cached price (flagged as stale). New actuator health indicator for pricing connectivity. See ADR-008.

---

## Change Log

| Date | Change | Files |
|---|---|---|
| 2026-02-15 | Added Resilience4j, configured circuit breaker | `pom.xml`, `application.yml` |
| 2026-02-16 | Wrapped PricingClient with CB, added fallback | `PricingClient.java`, `PricingFallbackService.java` |
| 2026-02-17 | Added health indicator and metrics | `PricingHealthIndicator.java` |
| 2026-02-18 | Unit and integration tests | `PricingClientTest.java`, `PricingFallbackIT.java` |

---

## Documentation Deltas

### Architecture Delta
> Target: `.ai/docs/ARCHITECTURE.md`

- **Type:** Modified
- **Section:** Core Components → Scalability & Resilience
- **Details:** Added circuit breaker between Quote Engine and Pricing Service. 5-failure threshold, 60s window, 30s half-open wait. Fallback: Redis cached prices flagged as stale.
- **Suggested Update:** Add circuit breaker annotation on architecture diagram. Update Scalability section with Resilience4j config.

### Dependency Delta
> Target: `.ai/docs/DEPENDENCIES.md`

**New Libraries:**

| Library | Version | Purpose |
|---|---|---|
| resilience4j-circuitbreaker | 2.1.0 | Circuit breaker for external calls |
| resilience4j-spring-boot3 | 2.1.0 | Spring Boot auto-configuration |

### API Delta

**Modified Endpoints:**

| Endpoint | Change | Breaking? |
|---|---|---|
| `POST /api/v3/quotes` | Response now includes `pricing_source` field (`live` or `cached`) | No (additive) |

### Configuration Delta

| Key | Default | Description | Env-Specific? |
|---|---|---|---|
| `resilience4j.circuitbreaker.instances.pricingService.failureRateThreshold` | 50 | % failure to open circuit | No |
| `resilience4j.circuitbreaker.instances.pricingService.slidingWindowSize` | 10 | Calls in sliding window | No |
| `resilience4j.circuitbreaker.instances.pricingService.waitDurationInOpenState` | 30s | Wait before half-open | No |
| `app.pricing.fallback.cache-ttl` | 15m | Max age for fallback cache | Yes (prod: 10m) |

### Deployment Delta
> Target: `.ai/docs/DEPLOYMENT.md`

- [x] New health check endpoint: `/actuator/health/pricing`
- [ ] New ConfigMap key: `app.pricing.fallback.cache-ttl` (add to all envs)

---

## ADRs Created

| ADR | Title |
|---|---|
| [ADR-008](./../adr/ADR-008-pricing-circuit-breaker.md) | Circuit breaker for Pricing Service resilience |

## Known Issues Resolved

| Issue ID | Resolution |
|---|---|
| ISSUE-001 | Pricing cache inconsistency — partially resolved with consistent fallback + TTL |

## New Tech Debt Introduced

| Description | Severity | Ticket |
|---|---|---|
| Stale price flag not yet shown in Portal UI | Low | PROJ-1400 |

---

## Testing

- [x] Unit tests: circuit breaker state transitions, fallback behavior
- [x] Integration tests: WireMock simulating Pricing Service failures
- [x] Edge cases: cached price expired + circuit open → returns 503

## Rebase Status

- **Status:** NOT REBASED
- **Rebased On:** —
- **Rebased By:** —
