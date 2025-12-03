---
title: "dataplex-update-glossary"
type: docs
weight: 1
description: >
  A "dataplex-update-glossary" tool updates display properties on a Dataplex business glossary.
aliases:
- /resources/tools/dataplex-update-glossary
---

## About

A `dataplex-update-glossary` tool updates an existing Dataplex business glossary
using the Business Glossary `UpdateGlossary` API. It supports updating display
metadata and labels and waits for the long-running operation to complete before
returning the updated glossary.

It's compatible with the following sources:

- [dataplex](../../sources/dataplex.md)

`dataplex-update-glossary` requires:

- `name` - Full resource name of the glossary in the format
  `projects/{project}/locations/{location}/glossaries/{glossary}`.

It also supports the following optional parameters:

- `displayName` - New friendly name for the glossary.
- `description` - New description for the glossary.
- `labels` - Map of string key/value labels. Provide an empty object to clear
  existing labels.
- `etag` - Optional ETag for optimistic concurrency control.
- `validateOnly` - When true, validates the request without applying changes.
  Defaults to `false`.

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
  dataplex-update-glossary:
    kind: dataplex-update-glossary
    source: my-dataplex-source
    description: Update a Dataplex business glossary.
```

## Reference

| **field**   | **type** | **required** | **description**                                                                           |
|-------------|:--------:|:------------:|-------------------------------------------------------------------------------------------|
| kind        |  string  |     true     | Must be "dataplex-update-glossary".                                                       |
| source      |  string  |     true     | Name of the source the tool should execute on.                                            |
| description |  string  |     true     | Description of the tool that is passed to the LLM.                                        |
| parameters.name        |  string  |     true     | Full resource name of the glossary to update.                                             |
| parameters.displayName |  string  |     false    | New display name for the glossary.                                                        |
| parameters.description |  string  |     false    | New description for the glossary.                                                         |
| parameters.labels      |   map    |     false    | Map of string labels. Provide an empty map to clear labels.                               |
| parameters.etag        |  string  |     false    | Optional ETag used for optimistic concurrency control.                                    |
| parameters.validateOnly| boolean |     false    | Validate the request only, do not apply changes. Defaults to false.                       |
