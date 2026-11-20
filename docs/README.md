# Midaz Receivables Plugin Documentation

Welcome to the comprehensive technical and business documentation for the **Midaz Receivables & Acquirer Plugin (`midaz-plugin-receivables`)**.

---

## 📚 Documentation Index

1. [01 - Product Requirements Document (PRD)](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/01-PRODUCT-REQUIREMENTS.md)
   * Problem statement, objectives, personas, functional requirements (FR-01 to FR-06), and non-functional requirements.
2. [02 - Technical Architecture & Design](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/02-TECHNICAL-ARCHITECTURE.md)
   * Hexagonal architecture, CQRS, directory structure, PostgreSQL schemas, and resiliency patterns.
3. [03 - Accounting & Double-Entry Specification](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/03-ACCOUNTING-DOUBLE-ENTRY.md)
   * Double-entry rules in Midaz, `@external` clearing accounts, receivables lifecycle (D+0 booking, D+X settlement, cancellations), and $N:1$ batch settlements.
4. [04 - Acquirer Providers Specification](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/04-ACQUIRER-PROVIDERS-SPEC.md)
   * Integration specifications for SumUp and PagBank (APIs, webhooks, authentication, and normalization to domain contracts).
5. [05 - Matcher 6 Reconciliation Specification](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/05-MATCHER-RECONCILIATION-SPEC.md)
   * Integration with Matcher 6 (`:4018`), Context and Sources configuration, rule matrix, upload payloads, and webhook callback handling.
6. [06 - Inter-Plugin Protocol: Open Finance Choreography](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/06-INTER-PLUGIN-PROTOCOL-OPENFINANCE.md)
   * Avoiding balance duplication, the `reconciliation.enabled` flag in Open Finance, and the three-way settlement workflow.
7. [07 - API & Midaz Console Manifest Specification](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/07-API-AND-MANIFEST-SPEC.md)
   * REST API endpoints, OpenAPI schemas, and `/api/manifest` microfrontend definition for Midaz Console.
8. [08 - Implementation Plan & Roadmap](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/08-IMPLEMENTATION-PLAN.md)
   * 6-phase disciplined execution roadmap from domain scaffolding to container packaging.
9. [09 - Testing Strategy & Quality Assurance](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/09-TESTING-STRATEGY.md)
   * Test pyramid, unit tests, PostgreSQL integration, and edge-case simulation suites.
