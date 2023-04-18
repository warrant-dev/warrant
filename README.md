<p align="center">
    <a href="https://warrant.dev/"><img src="https://warrant.dev/images/logo-primary-wide.png" width="300px" alt="Warrant" /></a>
</p>
<p align="center">
  <a href="https://warrant.dev/">Website</a> |
  <a href="https://app.warrant.dev/signup">Warrant Cloud</a> |
  <a href="https://docs.warrant.dev/">Docs</a> |
  <a href="https://docs.warrant.dev/objecttypes/get-all-object-types/">API Reference</a>
</p>

<p align="center">
    <img alt="GitHub" src="https://img.shields.io/github/license/warrant-dev/warrant?color=4F0DCC">
    <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/warrant-dev/warrant?color=FF5E00">
    <img alt="GitHub Workflow Status (with branch)" src="https://img.shields.io/github/actions/workflow/status/warrant-dev/warrant/ci.yaml?branch=main">
    <a href="https://join.slack.com/t/warrantcommunity/shared_invite/zt-12g84updv-5l1pktJf2bI5WIKN4_~f4w"><img alt="Slack Community" src="https://img.shields.io/badge/Slack%20Community-4A154B?style=flat&logo=slack"></a>
    <a href="https://twitter.com/warrant_dev"><img alt="Twitter Follow" src="https://img.shields.io/badge/follow-%40warrant__dev-1DA1F2?logo=twitter"></a>
    <a href="https://www.ycombinator.com/companies/warrant"><img alt="Backed by Y Combinator" src="https://img.shields.io/badge/Backed%20by-Y%20Combinator-%23E16E38"/></a>
</p>

# Warrant - Open Source Access Control Service

Warrant is an application access control service built for developers and product teams. It is a centralized authorization service (inspired by Google Zanzibar) designed to abstract away the complexity of managing user access control from teams building software products, so they can (1) offer best-in-class access control to customers from day one and (2) focus efforts on building their core product.

Warrant allows you to define, store, and manage your product's authorization model and access rules (we call these warrants), then check or query against these rules from your application at runtime. This allows you to easily implement any access control model from coarser role based access control (RBAC) (e.g. `does [user:1] have [permission:view-billing-details]?`) to fine grained relationship based access control (ReBAC) (e.g. `is [user:1] an [editor] of [document:docA]?`) and attribute based access control (ABAC) (e.g. `is [user:1] in [department:accounting]?`).

## Features

- A real-time, low latency `check` API to perform access checks in your application at runtime (e.g. _is `user:A editor of tenant:X`?_)
- Built-in support for roles & permissions (RBAC) + API endpoints to create and manage custom roles & permissions from your application
- Built-in support for multi-tenant access control
  - Define roles, permissions, access rules, etc. _per tenant_
  - Support scenarios where a user's level of access in your application is dependent on which tenant (or role) they're currently logged in as.
- Built-in support for pricing tiers - control access to your application's features in real time based on your product's pricing plans (e.g. free-tier, growth, business, enterprise, etc)
- Permission-aware front-end components to allow/deny access to specific pages/UI elements
- Pre-built components & embeddable pages to build UIs that give customers the ability to manage roles, permissions, and other access rules for themselves and teammates
- Integrates with in-house and third-party authn/identity providers like Auth0
- A global event log that tracks all updates to authorization models and access rules, making auditing, alerting, and debugging simple
- [SDKs](#sdks) in the most popular languages and frameworks

## Use Cases

Warrant is built specifically for application authorization and access control, particularly for product, security, and compliance use-cases. Examples of problems Warrant solves are:

- Add role based access control (RBAC) to your SaaS application with the ability for your customers to self-manage their roles and permissions via the Warrant self service dashboard or your own custom dashboard built using Warrant's component library.
- Allow your customers to define and manage their own roles & permissions for their tenant (organization)
- Add 'fine grained RBAC' (role based access to specific resources)
- Implement fine grained, object/resource-level authorization specific to your application's data model (`[user:1] is an [editor] of [document:x]`)
- Add centralized and auditable access control around your internal applications.
- Implement 'approval flows' (i.e. request access to a resource from an admin -> admin approves access).
- Add Google Docs-like sharing and permissioning for your application's resources and objects.
- Gate access to SaaS features based on your product's pricing tiers and feature packages.
- Satisfy auditing and compliance requirements of frameworks and standards such as SOC2, HIPAA, GDPR and CCPA.

## Getting Started

### Warrant Cloud

The quickest and easiest way to get started with Warrant is by using the managed cloud service. You can sign-up for a free account [here](https://app.warrant.dev/signup).

Warrant Cloud is compatible with the same APIs as this open source version and provides additional functionality like:

- An admin dashboard for quickly managing your authorization model and access rules via an intuitive, easy-to-use UI
- A real-time `query` API to query and audit access rules for a given subject or object (e.g. _`which users in tenant:1 have access to object:A?`_)
- Multi-region availability
- Improved access check latency & throughput for large scale use cases.

Once you've created an account, refer to our [docs](https://docs.warrant.dev/) to get started.

### Self-hosting

To self-host or run Warrant locally, follow one of the guides below (select the guide for your database of choice).

- [MySQL](/migrations/datastore/mysql/README.md)
- [Postgres](/migrations/datastore/postgres/README.md)
- [SQLite](/migrations/datastore/sqlite/README.md)
- To request support for another database, please [open an issue](https://github.com/warrant-dev/warrant/issues/new/choose)!

## SDKs

Warrant's native SDKs are compatible with both the cloud and open-source versions of Warrant. We currently support SDKs for:

- [Node.js](https://github.com/warrant-dev/warrant-node)
- [Go](https://github.com/warrant-dev/warrant-go)
- [Python](https://github.com/warrant-dev/warrant-python)
- [Ruby](https://github.com/warrant-dev/warrant-ruby)
- [PHP](https://github.com/warrant-dev/warrant-php)
- [Java](https://github.com/warrant-dev/warrant-java)
- [React](https://github.com/warrant-dev/react-warrant-js)
- [Angular](https://github.com/warrant-dev/angular-warrant)
- [Vue](https://github.com/warrant-dev/vue-warrant)

## Documentation

Visit our [docs](https://docs.warrant.dev/) to learn more about Warrant's key concepts & architecture and view our [quickstarts](https://docs.warrant.dev/quickstart/role-based-access-control/) & [API reference](https://docs.warrant.dev/objecttypes/get-all-object-types/).

## Support

Join our [Slack community](https://join.slack.com/t/warrantcommunity/shared_invite/zt-12g84updv-5l1pktJf2bI5WIKN4_~f4w) to ask questions and get support.

## Contributing

To report a bug you found or request a feature that you'd like, open an issue. If you'd like to contribute, submit a PR to resolve the issue.

Contributions from the community are welcome! Just be sure to follow some ground rules:

- Never submit a PR without an associated issue.
- Issues should mention whether the issue is a bug or a feature.
- Issues reporting a bug should describe (1) steps to reproduce the bug, (2) what the current behavior is, and (3) what the expected behavior should be.
- Issues requesting a feature should (1) provide a description of the feature and (2) explain the intended use case for the feature.
