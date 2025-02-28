<p align="center">
    <a href="https://warrant.dev/"><img src="https://warrant.dev/images/og-image.png" alt="Warrant" /></a>
</p>
<p align="center">
  <a href="https://warrant.dev/">Website</a> |
  <a href="https://workos.com/fine-grained-authorization">WorkOS FGA</a> |
  <a href="https://workos.com/docs/fga">Docs</a> |
  <a href="https://workos.com/docs/reference/fga">API Reference</a>
</p>

<p align="center">
    <a href="https://github.com/warrant-dev/warrant/blob/main/LICENSE"><img alt="GitHub" src="https://img.shields.io/github/license/warrant-dev/warrant?color=4F0DCC">
    <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/warrant-dev/warrant?color=FF5E00">
    <a href="https://github.com/warrant-dev/warrant/actions"><img alt="GitHub Workflow Status (with branch)" src="https://img.shields.io/github/actions/workflow/status/warrant-dev/warrant/go.yaml?branch=main">
    <a href="https://twitter.com/warrant_dev"><img alt="Twitter Follow" src="https://img.shields.io/badge/follow-%40warrant__dev-1DA1F2?logo=twitter"></a>
</p>

# Warrant - Google Zanzibar-inspired, Fine-Grained Authorization Service

Warrant is a **highly scalable, centralized, fine-grained authorization service** for _defining_, _storing_, _querying_, _checking_, and _auditing_ application authorization models and access rules. At its core, Warrant is a [relationship based access control (ReBAC)](https://en.wikipedia.org/wiki/Relationship-based_access_control) engine (inspired by [Google Zanzibar](https://research.google/pubs/pub48190/)) capable of enforcing any authorization paradigm, including role based access control (RBAC) (e.g. `[user:1] has [permission:view-billing-details]`), attribute based access control (ABAC) (e.g. `[user:1] can [view] [department:accounting] if [geo == "us"]`), and relationship based access control (ReBAC) (e.g. `[user:1] is an [editor] of [document:docA]`). It is especially useful for implementing fine-grained access control (FGAC) in internal and/or customer-facing applications.

## Features

- HTTP APIs for managing your authorization model, access rules, and other Warrant resources (roles, permissions, features, tenants, users, etc.) from an application, a CLI tool, etc.
- Real-time, low-latency API for performing access checks in your application(s) at runtime (e.g. `is [user:A] an [editor] of [tenant:X]?`)
- Integrates with in-house and third-party authn/identity providers like Auth0, Firebase, and more
- [SDKs](#sdks) for popular languages and frameworks (backend and frontend)
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

Check out the [development guide](/development.md) to learn how to run Warrant locally and refer to the [deployment examples](/deployment.md) for examples of self-hosting Warrant using Docker or Kubernetes.

## SDKs

- [Node.js](https://github.com/warrant-dev/warrant-node)
- [Go](https://github.com/warrant-dev/warrant-go)
- [Python](https://github.com/warrant-dev/warrant-python)
- [Ruby](https://github.com/warrant-dev/warrant-ruby)
- [PHP](https://github.com/warrant-dev/warrant-php)
- [Java](https://github.com/warrant-dev/warrant-java)
- [React](https://github.com/warrant-dev/react-warrant-js)
- [Angular](https://github.com/warrant-dev/angular-warrant)
- [Vue](https://github.com/warrant-dev/vue-warrant)

## Limitations

Serving check and query requests with low latency at high throughput requires running Warrant as a distributed service with the use of [Warrant-Tokens](https://workos.com/docs/fga/warrant-tokens) (also referred to as [Zookies](https://workos.com/blog/google-zanzibar-authorization#global-scale-low-latency) in Google Zanzibar). As a result, this open source version of Warrant is only capable of handling low-to-moderate throughput and is best suited for POCs, development/test environments, and low throughput use-cases.

## Get <10ms Latency at Scale

### WorkOS FGA

The quickest and easiest way to get low-latency performance for high-throughput production usage is to use [WorkOS FGA](https://workos.com/fine-grained-authorization), a fully managed, serverless fine-grained authorization service. With WorkOS FGA, you don't need to worry about managing multiple instances of Warrant or its underlying datastore (e.g. Postgres, MySQL, etc). It can scale to millions of warrants and hundreds of millions of check and query requests while still providing <10ms latencies. You can sign up for a free account [here](https://signin.workos.com/sign-up).

WorkOS FGA also provides additional functionality like:

- An admin dashboard for quickly managing your authorization model and access rules via an intuitive, easy-to-use UI
- Batch endpoints
- Multi-region availability
- Improved access check latency & throughput for large scale use cases

Once you've created an account, refer to our [docs](https://workos.com/docs/fga) to get started.

### Enterprise Self-Hosted

Interested in self-hosting an enterprise version of Warrant or WorkOS FGA? Please [contact us](https://workos.com/contact) for more information.

## Contributing

Contributions are welcome. Please see our [contributing guide](/CONTRIBUTING.md) for more details.
