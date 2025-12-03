---
title: "dataplex-create-glossary"
type: docs
weight: 1
description: >
  A "dataplex-create-glossary" tool creates a Dataplex business glossary.
aliases:
- /resources/tools/dataplex-create-glossary
---

## About

A `dataplex-create-glossary` tool provisions a new business glossary in Dataplex
Universal Catalog. It wraps the Dataplex Business Glossary `CreateGlossary`
API and waits for the long-running operation to complete before returning the
created glossary.

It's compatible with the following sources:

- [dataplex](../../sources/dataplex.md)

`dataplex-create-glossary` requires:

- `location` - Google Cloud region where the glossary should live (for example
  `us-central1`).
- `glossaryId` - Identifier for the glossary; becomes part of the resource
  name.

It also supports the following optional parameters:

- `displayName` - Friendly name for the glossary. Defaults to the ID when not
  provided.
- `description` - Description text for the glossary.
- `labels` - Map of string key/value labels to set on the glossary.
- `validateOnly` - When true, validates the request without creating the
  glossary. Defaults to `false`.

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
  dataplex-create-glossary:
    kind: dataplex-create-glossary
    source: my-dataplex-source
    description: Create a Dataplex business glossary.
```

## Reference

| **field**   | **type** | **required** | **description**                                                                           |
|-------------|:--------:|:------------:|-------------------------------------------------------------------------------------------|
| kind        |  string  |     true     | Must be "dataplex-create-glossary".                                                       |
| source      |  string  |     true     | Name of the source the tool should execute on.                                            |
| description |  string  |     true     | Description of the tool that is passed to the LLM.                                        |
| parameters.location    |  string  |     true     | Location where the glossary is created.                                                   |
| parameters.glossaryId  |  string  |     true     | Identifier for the glossary.                                                              |
| parameters.displayName |  string  |     false    | User-friendly display name.                                                               |
| parameters.description |  string  |     false    | Description text for the glossary.                                                        |
| parameters.labels      |   map    |     false    | Map of string labels to set on the glossary.                                              |
| parameters.validateOnly| boolean |     false    | Validate the request only, do not create the glossary. Defaults to false.                 |
