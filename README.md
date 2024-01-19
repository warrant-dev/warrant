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
    <img alt="GitHub Workflow Status (with branch)" src="https://img.shields.io/github/actions/workflow/status/warrant-dev/warrant/go.yaml?branch=main">
    <a href="https://join.slack.com/t/warrantcommunity/shared_invite/zt-12g84updv-5l1pktJf2bI5WIKN4_~f4w"><img alt="Slack Community" src="https://img.shields.io/badge/Slack%20Community-4A154B?style=flat&logo=slack"></a>
    <a href="https://twitter.com/warrant_dev"><img alt="Twitter Follow" src="https://img.shields.io/badge/follow-%40warrant__dev-1DA1F2?logo=twitter"></a>
    <a href="https://www.ycombinator.com/companies/warrant"><img alt="Backed by Y Combinator" src="https://img.shields.io/badge/Backed%20by-Y%20Combinator-%23E16E38"/></a>
</p>

# Warrant - Google Zanzibar-inspired, Fine-Grained Authorization Service

Warrant is a **highly scalable, centralized, fine-grained authorization service** for _defining_, _storing_, _querying_, _checking_, and _auditing_ application authorization models and access rules. At its core, Warrant is a [relationship based access control (ReBAC)](https://en.wikipedia.org/wiki/Relationship-based_access_control) engine (inspired by [Google Zanzibar](https://research.google/pubs/pub48190/)) capable of enforcing any authorization paradigm, including role based access control (RBAC) (e.g. `[user:1] has [permission:view-billing-details]`), attribute based access control (ABAC) (e.g. `[user:1] can [view] [department:accounting] if [geo == "us"]`), and relationship based access control (ReBAC) (e.g. `[user:1] is an [editor] of [document:docA]`). It is especially useful for implementing fine-grained access control (FGAC) in internal and/or customer-facing applications.

## Features

- HTTP APIs for managing your authorization model, access rules, and other Warrant resources (roles, permissions, features, tenants, users, etc.) from an application, a CLI tool, etc.
- Real-time, low latency API for performing access checks in your application(s) at runtime (e.g. `is [user:A] an [editor] of [tenant:X]?`)
- Integrates with in-house and third-party authn/identity providers like Auth0 and Firebase
- Officially supported [SDKs](#sdks) for most popular languages and frameworks
- Support for a number of databases, including: MySQL, Postgres, and SQLite (in-memory or file)

## Use Cases

Warrant is built specifically for application authorization and access control, particularly for product, security, and compliance use-cases. Examples of problems Warrant solves are:

- Add role based access control (RBAC) to your SaaS application with the ability for your customers to self-manage their roles and permissions via the Warrant self-service dashboard or your own custom dashboard built using Warrant's component library.
- Allow customers to define and manage their own roles & permissions for their tenant (organization)
- Add 'fine-grained role-based access control' (role based access to specific resources)
- Implement fine-grained, object/resource-level authorization specific to your application's data model (`[user:1] is an [editor] of [document:x]`)
- Add centralized and auditable access control around your internal applications and tools.
- Implement 'approval flows' (i.e. request access to a resource from an admin -> admin approves access).
- Add Google Docs-like sharing and permissioning for your application's resources and objects.
- Gate access to SaaS features based on your product's pricing tiers and feature packages.
- Satisfy auditing and compliance requirements of frameworks and standards such as SOC2, HIPAA, GDPR and CCPA.

## Getting Started

### Warrant Cloud

The quickest and easiest way to get started with Warrant is using the managed cloud service. You can sign-up for a free account [here](https://app.warrant.dev/signup).

Warrant Cloud is compatible with the same APIs as this open source version and provides additional functionality like:

- An admin dashboard for quickly managing your authorization model and access rules via an intuitive, easy-to-use UI
- A real-time `query` API to query and audit access rules for a given subject or object (e.g. _`which users in tenant:1 have access to object:A?`_)
- Multi-region availability
- Improved access check latency & throughput for large scale use cases.

Once you've created an account, refer to our [docs](https://docs.warrant.dev/) to get started.

### Local Development &amp; Self-Hosting

Check out the [development guide](/development.md) to learn how to run Warrant and refer to the [deployment examples](/deployment.md) for examples of self-hosting Warrant using Docker or Kubernetes.

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

Contributions welcome. Please see our [contributing guide](/CONTRIBUTING.md) for more details.
