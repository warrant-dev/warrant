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

Warrant is an application authorization & access control platform built for developers and product teams. It's designed to abstract away the complexity of managing authorization from teams so they can focus their efforts on building their core product.

Key features include:
- A centralized service (inspired by Google Zanzibar) for storing and managing authorization models and their associated access rules (we call these warrants). The service supports a wide variety of common access control patterns from coarser Role Based Access Control (RBAC) to fine grained Relationship Based Access Control (ReBAC) and Attribute Based Access Control (ABAC) schemes (ex. `[user:1] is an [editor] of [document:x]`).
- Real-time, low latency `check` API to check for specific access rules (i.e. *is user:A editor of tenant:X*)
- Real-time `query` API to query and audit access rules for a given subject or object
- Built-in support for multi-tenant access control - define access rules by tenant
- Built-in support for pricing tiers & features - define access rules based on your SaaS pricing tiers and feature packages
- Front-end components and embeddable pages to allow/deny access to certain pages/UI elements, enable self-service management of permissions, etc.
- Connectors to sync tenants, users, and access rules from other sources (i.e. IdPs, DBs, etc).
- A global event log of all operations for audit logging, alerting, and debugging authorization models

## Getting Started

### Warrant Cloud

The fastest and easiest way to get started with Warrant is through the managed cloud service. You can sign-up for a free account [here](https://app.warrant.dev/signup).

Warrant Cloud is compatible with the same APIs as this open source version with additional functionality such as the admin dashboard, multi-region availability, and improved latency &amp; throughput for large scale use cases.

Once you've created an account, use one of our [SDKs](/#SDKs) and reference the [docs](https://docs.warrant.dev/) to get started.

### Self-hosted

```shell
[Steps to install Warrant via Docker/binary]
```

## Use Cases

Warrant is built specifically for application authorization and access control use cases, particularly those related to security and compliance. Examples of problems Warrant solves are:

- Add role based access control (RBAC) to your SaaS application with the ability for your customers to self-manage their roles and permissions via the Warrant self service dashboard or your own custom dashboard built using Warrant's component library.
- Allow your customers to define and manage their own roles &amp; permissions for their tenant (organization)
- Add 'fine grained RBAC' (role based access to specific resources)
- Implement fine grained, object/resource-level authorization specific to your application's data model (`[user:1] is an [editor] of [document:x]`)
- Add centralized and auditable access control around your internal applications.
- Implement 'approval flows' (i.e. request access to a resource from an admin -> admin approves access).
- Add Google Docs-like sharing and permissioning for your application's resources and objects.
- Gate access to SaaS features based on your product's pricing tiers and feature packages.
- Satisfy auditing and compliance requirements of frameworks and standards such as SOC2, HIPAA, GDPR and CCPA.

## SDKs

Warrant's native SDKs are compatible with both the cloud and open-source version. We currently support:
- [Node.js](https://github.com/warrant-dev/warrant-node)
- [Go](https://github.com/warrant-dev/warrant-go)
- [Python](https://github.com/warrant-dev/warrant-python)
- [Ruby](https://github.com/warrant-dev/warrant-ruby)
- [PHP](https://github.com/warrant-dev/warrant-php)
- [Java](https://github.com/warrant-dev/warrant-java)
- [React](https://github.com/warrant-dev/react-warrant-js)
- [Angular](https://github.com/warrant-dev/angular-warrant)
- [Vue](https://github.com/warrant-dev/vue-warrant)

## Documentation & Support

Check out our [docs](https://docs.warrant.dev/) for deep-dives into key concepts and architecture as well as [quickstarts](https://docs.warrant.dev/quickstart/role-based-access-control/) and the [API reference](https://docs.warrant.dev/objecttypes/get-all-object-types/).

Join our [Slack community](https://join.slack.com/t/warrantcommunity/shared_invite/zt-12g84updv-5l1pktJf2bI5WIKN4_~f4w) to ask questions and get support.

## Contributing

TBD

## License

TBD
