<p align="center">
    <a href="https://warrant.dev/"><img src="https://warrant.dev/images/logo-primary-wide.png" width="300px" alt="Warrant" /></a>
</p>
<p align="center">
  <a href="https://warrant.dev/">Website</a> |
  <a href="https://docs.warrant.dev/">Docs</a> |
  <a href="https://docs.warrant.dev/objecttypes/get-all-object-types/">API Reference</a> |
  <a href="https://join.slack.com/t/warrantcommunity/shared_invite/zt-12g84updv-5l1pktJf2bI5WIKN4_~f4w">Slack</a> |
  <a href="https://twitter.com/warrant_dev">Twitter</a>
</p>

# Warrant - Open Source Access Control as a Service

Warrant is an application access control service built for developers and product teams. It's designed to abstract away the complexity of managing user access control from teams building software products so they (1) can offer best-in-class access control to customers from day one and (2) focus their efforts on building their core product.

## Features

- A centralized authorization service (inspired by Google Zanzibar) for defining, storing, and managing your product's authorization model and access rules (we call these warrants).
- Supports a wide variety of common access control models from coarser Role Based Access Control (RBAC) to fine grained Relationship Based Access Control (ReBAC) and Attribute Based Access Control (ABAC) (e.g. `[user:1] is an [editor] of [document:x]`)
- Real-time, low latency `check` API to perform access checks in your application at runtime (e.g. _is `user:A editor of tenant:X`?_)
- Real-time `query` API to query and audit access rules for a given subject or object (e.g. _`which users in tenant:1 have access to object:A?`_)
- Built-in support for roles &amp; permissions (RBAC)
  - API endpoints to create and manage custom roles and permissions for your users
- Built-in support for multi-tenant access control
  - Define roles, permissions, and other access rules _per tenant_
  - Support scenarios where users have varying levels of access to resources depending on which tenant (or role) they're currently logged in as.
- Built-in support for pricing tiers
  - Control access to your application&apos;s features in real time based on the pricing tiers offered by your product (e.g. free-tier, growth, business, enterprise, etc)
- Front-end components and embeddable pages to allow/deny access to specific pages/UI elements and enable self-service management of roles &amp; permissions
  - Pre-built components that help you build UIs that give your customers the ability to manage roles &amp; permissions for themselves and their teammates
- Easily integrates with in-house and third-party authn/identity providers like Auth0
- Maintains a global event log that tracks all updates to authorization models and rules to make auditing, alerting, and debugging simple
- SDK support for many of the most commonly used programming languages:
  - Go
  - Java
  - JS/TS
  - Python
  - PHP
  - Ruby

## Use Cases

Warrant is built specifically for application authorization and access control, particularly for product, security, and compliance use-cases. Examples of problems Warrant solves are:

- Add role based access control (RBAC) to your SaaS application with the ability for your customers to self-manage their roles and permissions via the Warrant self service dashboard or your own custom dashboard built using Warrant's component library.
- Allow your customers to define and manage their own roles &amp; permissions for their tenant (organization)
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
- Multi-region availability
- Improved access check latency &amp; throughput for large scale use cases.

Once you've created an account, refer to our [docs](https://docs.warrant.dev/) to get started.

### Self-hosting

Warrant requires a database to store authorization models and access rules. The open source version currently has support for the following databases:

- MySQL
- Postgres (coming soon)
- SQLite (coming soon)
- To request support for another database, please [open an issue](https://github.com/warrant-dev/warrant/issues/new/choose)!

If you'd like to self-host the open source version of Warrant, you have a few options:

#### Install and run the binary

#### Run with Docker

#### Run with Kubernetes

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

Visit our [docs](https://docs.warrant.dev/) to learn more about Warrant's key concepts &amp; architecture and view our [quickstarts](https://docs.warrant.dev/quickstart/role-based-access-control/) &amp; [API reference](https://docs.warrant.dev/objecttypes/get-all-object-types/).

## Support

Join our [Slack community](https://join.slack.com/t/warrantcommunity/shared_invite/zt-12g84updv-5l1pktJf2bI5WIKN4_~f4w) to ask questions and get support.

## Contributing

To report a bug you found or request a feature that you'd like, open an issue. If you'd like to contribute, submit a PR to resolve the issue.

Contributions from the community are welcome! Just be sure to follow some ground rules:

- Never submit a PR without an issue.
- Issues should mention whether the issue is a bug or a feature.
- Issues reporting a bug should describe (1) steps to reproduce the bug, (2) what the current behavior is, and (3) what the expected behavior should be.
- Issues requesting a feature should (1) provide a description of the feature and (2) explain the intended use case for the feature.
