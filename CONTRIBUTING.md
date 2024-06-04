# Contributing guide

Warrant is under active development and welcomes contributions from the community via pull requests.

Prior to submitting any PRs, please review and familiarize yourself with the following guidelines and our [Code of Conduct](/CODE_OF_CONDUCT.md).

## Issues

- All outstanding bugs and feature requests are [tracked as GitHub issues](https://github.com/warrant-dev/warrant/issues).
- If you find a bug or have a feature request, please [open an issue](https://github.com/warrant-dev/warrant/issues/new/choose). In order to prevent duplicate reports, please first search through existing open issues for your request prior to creating a new one.
- If you discover a security issue or vulnerability, do not create an issue. Please email us with details at security@warrant.dev.
- If you find small mistakes or issues in docs/instructions etc., feel free to submit PR fixes without first creating issues.
- If you'd like to contribute a fix or implementation for an issue, please first consult on your approach with a member of the Warrant team directly on the GitHub issue.
- [Here](https://github.com/warrant-dev/warrant/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) is a list of 'good first issues' for new contributors.

## Making changes

- Refer to the [local development guide](/development.md) to get your development environment set up.
- Warrant uses the fork & pull request flow. Please make sure to fork your own copy of the repo and use feature branches for development.
- Make sure to [test your changes](/development.md#running-tests) and add tests where necessary.

## Submitting pull requests

- Unless it's a minor change, never submit a PR without an associated issue.
- Once you've implemented and tested your code changes, submit a pull request.
- Pull requests will trigger ci jobs that run linters, static analysis and tests. It is the submitter's responsibility to ensure that all ci checks are passing.
- A member of the Warrant team will review your PR. Once approved, you may merge your PR into main.
- New versions will be tagged and released automatically.
