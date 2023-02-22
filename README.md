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

Warrant is an application authorization & access control platform built for developers and product teams. It's designed to abstract and handle the complexity of managing authorization so teams can focus on building their core products.

Key features include:
- A centralized service (inspired by Google Zanzibar) for storing and managing access models and their associated rules (warrants). Supports a variety of custom and common access control schemes from role based access control (RBAC) to more complex relationship based access control (ReBAC) and fine grained access control schemes (ex. `[user:1] is an [editor] of [document:x]`)
- Built-in multi-tenant support - define access rules by tenant
- Built-in pricing tiers & feature support - define access rules based on SaaS pricing tiers and feature packages
- Real-time `check` and `query` APIs to check access for any given subject and to figure out who has access to what
- Front-end UI components and embeddable apps to enable self-service management, showing/hiding UIs etc. 
- Connectors to sync access rules from other sources (ex. IdPs, DBs)
- Events for all operations to enable audit logging and alerting

## Getting Started

### Warrant Cloud

The fastest and easiest way to get started with Warrant is through the managed cloud service. You can sign-up for a free account [here](https://app.warrant.dev/signup).

Warrant Cloud is compatible with the same APIs as Warrant OSS with added functionality such as the admin dashboard and custom improvements for multi-region availability and large-scale/high-throughput use cases.

Once you've created an account, use one of our [SDKs](/#SDKs) and reference the [docs](https://docs.warrant.dev/) to get started.

### Self-hosted

```shell
[Steps to install Warrant via Docker/binary]
```

## Use Cases

Warrant is built specifically for application authorization and access control use cases, particularly those related to security and compliance. Some examples of what you can do and build with Warrant:

- Add basic role based access control (RBAC) to a SaaS application including the ability for customers to self-service their roles and permissions via the embeddable management app.
- Support more advanced RBAC scenarios including custom roles and permissions per tenant or 'fine grained RBAC' (access to specific resources by role).
- Implement fine grained, object/resource-level authorization for your custom data model (`[user:1] is an [editor] of [document:x]`)
- Add centralized and auditable access control around your internal applications.
- Implement 'approval flows' (i.e. request access to a resource).
- Add a Google Docs-like sharing model for your own resources and objects.
- Gate access to SaaS features based on custom pricing tiers and feature packages.
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
