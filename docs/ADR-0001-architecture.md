# ADR-0001: Clean Architecture layering

Decision: Use domain/usecase/adapters/infrastructure separation to keep core logic independent from frameworks and storage. Interfaces live on consumer side (usecase) per Go best practices.

Status: Accepted (MVP)

Consequences: Easier to swap storage or extend decoders; slightly more boilerplate.


