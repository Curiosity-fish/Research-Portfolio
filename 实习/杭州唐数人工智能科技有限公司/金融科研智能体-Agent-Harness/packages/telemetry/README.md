# @earendil-works/tfa-telemetry

Vendor-neutral telemetry contracts and typed schema utilities for tfa packages.

This package contains no exporter and depends on no telemetry backend. Applications provide a `TelemetryContext` adapter; tfa packages pass contexts explicitly and define their own domain schemas.
