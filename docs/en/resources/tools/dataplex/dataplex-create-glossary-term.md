---
title: "dataplex-create-glossary-term"
type: docs
weight: 1
description: >
  A "dataplex-create-glossary-term" tool creates a term inside a Dataplex business glossary.
aliases:
- /resources/tools/dataplex-create-glossary-term
---

## About

A `dataplex-create-glossary-term` tool creates a new term within an existing
Dataplex business glossary using the Business Glossary `CreateGlossaryTerm`
API.

It's compatible with the following sources:

- [dataplex](../../sources/dataplex.md)

`dataplex-create-glossary-term` requires:

- `glossaryProject` - Project that owns the target glossary.
- `glossaryLocation` - Location of the glossary (often `global`).
- `glossaryId` - ID of the glossary.
- `termId` - ID to assign to the new term.

It also supports the following optional parameters:

- `displayName` - Friendly name for the term.
- `description` - Description of the term.
- `labels` - Map of string key/value labels to set on the term.

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
  dataplex-create-glossary-term:
    kind: dataplex-create-glossary-term
    source: my-dataplex-source
    description: Create a term within a Dataplex glossary.
```

## Reference

| **field**   | **type** | **required** | **description**                                               |
|-------------|:--------:|:------------:|---------------------------------------------------------------|
| kind        |  string  |     true     | Must be "dataplex-create-glossary-term".                      |
| source      |  string  |     true     | Name of the source the tool should execute on.                |
| description |  string  |     true     | Description of the tool that is passed to the LLM.            |
| parameters.glossaryProject | string | true  | Project that owns the glossary.                               |
| parameters.glossaryLocation| string | true  | Location of the glossary (often global).                      |
| parameters.glossaryId      | string | true  | Glossary ID.                                                  |
| parameters.termId          | string | true  | Term ID to assign.                                            |
| parameters.displayName     | string | false | Friendly name for the term.                                   |
| parameters.description     | string | false | Description text.                                             |
| parameters.labels          |   map  | false | Map of string labels.                                         |
