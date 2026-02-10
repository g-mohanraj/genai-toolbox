---
title: "dataplex-list-glossaries"
type: docs
weight: 1
description: >
  A "dataplex-list-glossaries" tool lists business glossaries in a Dataplex location.
aliases:
- /resources/tools/dataplex-list-glossaries
---

## About

A `dataplex-list-glossaries` tool returns the business glossaries that exist in a
given Dataplex location. It wraps the Business Glossary `ListGlossaries` API.

It's compatible with the following sources:

- [dataplex](../../sources/dataplex.md)

`dataplex-list-glossaries` requires:

- `location` - Google Cloud region to list glossaries from (for example
  `us-central1`).

It also supports the following optional parameters:

- `pageSize` - Maximum number of glossaries to return (default `50`, max `1000`).
- `pageToken` - Token from a previous response to continue listing.
- `filter` - Filter on glossary fields, e.g. `display_name="sales"`.
- `orderBy` - Ordering for results, e.g. `name` or `create_time`.

## Requirements

### IAM Permissions

Dataplex uses [Identity and Access Management (IAM)][iam-overview] to control
user and group access to Dataplex resources. Toolbox will use your
[Application Default Credentials (ADC)][adc] to authorize and authenticate when
interacting with [Dataplex][dataplex-docs].

In addition to [setting the ADC for your server][set-adc], you need to ensure
the IAM identity has been given the correct IAM permissions for the tasks you
intend to perform. See [Dataplex Universal Catalog IAM permissions][iam-permissions]
and [Dataplex Universal Catalog IAM roles][iam-roles] for more information on
applying IAM permissions and roles to an identity.

[iam-overview]: https://cloud.google.com/dataplex/docs/iam-and-access-control
[adc]: https://cloud.google.com/docs/authentication#adc
[set-adc]: https://cloud.google.com/docs/authentication/provide-credentials-adc
[iam-permissions]: https://cloud.google.com/dataplex/docs/iam-permissions
[iam-roles]: https://cloud.google.com/dataplex/docs/iam-roles
[dataplex-docs]: https://cloud.google.com/dataplex

## Example

```yaml
tools:
  dataplex-list-glossaries:
    kind: dataplex-list-glossaries
    source: my-dataplex-source
    description: List Dataplex business glossaries in a location.
```

## Reference

| **field**   | **type** | **required** | **description**                                                      |
|-------------|:--------:|:------------:|----------------------------------------------------------------------|
| kind        |  string  |     true     | Must be "dataplex-list-glossaries".                                  |
| source      |  string  |     true     | Name of the source the tool should execute on.                       |
| description |  string  |     true     | Description of the tool that is passed to the LLM.                   |
| parameters.location | string | true  | Location to list glossaries from.                                    |
| parameters.pageSize | integer | false | Maximum results to return (default 50, max 1000).                    |
| parameters.pageToken| string | false | Token from a previous response to continue listing.                  |
| parameters.filter   | string | false | Filter on glossary fields.                                           |
| parameters.orderBy  | string | false | Ordering for results (e.g. name or create_time).                     |
