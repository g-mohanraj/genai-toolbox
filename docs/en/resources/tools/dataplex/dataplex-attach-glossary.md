---
title: "dataplex-attach-glossary"
type: docs
weight: 1
description: >
  A "dataplex-attach-glossary" tool creates an entry link between a catalog entry and a glossary term.
aliases:
- /resources/tools/dataplex-attach-glossary
---

## About

A `dataplex-attach-glossary` tool links an existing catalog entry (for example a
BigQuery table) to a business glossary term by creating an EntryLink. For the
default `definition` link type, the data asset is the SOURCE and the glossary
term is the TARGET.

It's compatible with the following sources:

- [dataplex](../../sources/dataplex.md)

`dataplex-attach-glossary` requires:

- `parent` - Entry group resource that will own the link in the format
  `projects/{project}/locations/{location}/entryGroups/{entryGroup}`.
- `entryLinkId` - Identifier for the entry link.
- `sourceEntry` - Resource name of the entry to link from.
- `glossaryProject` - Project that contains the glossary term.
- `glossaryLocation` - Location for the glossary term (often `global`).
- `glossaryId` - Glossary ID that contains the term.
- `termId` - Glossary term ID to link to.

It also supports the following optional parameters:

- `entryLinkType` - Entry link type to use. Defaults to
  `projects/dataplex-types/locations/global/entryLinkTypes/definition`.

The tool builds the glossary term entry resource name in the required escaped
format and submits a `CreateEntryLink` request, waiting for completion before
returning.

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
  dataplex-attach-glossary:
    kind: dataplex-attach-glossary
    source: my-dataplex-source
    description: Link a catalog entry to a glossary term using an entry link.
```

## Reference

| **field**   | **type** | **required** | **description**                                                                 |
|-------------|:--------:|:------------:|---------------------------------------------------------------------------------|
| kind        |  string  |     true     | Must be "dataplex-attach-glossary".                                             |
| source      |  string  |     true     | Name of the source the tool should execute on.                                  |
| description |  string  |     true     | Description of the tool that is passed to the LLM.                              |
| parameters.parent        | string | true  | Entry group that will own the link.                                             |
| parameters.entryLinkId   | string | true  | Identifier for the entry link.                                                  |
| parameters.sourceEntry   | string | true  | Resource name of the source entry.                                              |
| parameters.glossaryProject | string | true | Project containing the glossary term.                                           |
| parameters.glossaryLocation | string | true | Location for the glossary term (often global).                                  |
| parameters.glossaryId    | string | true  | Glossary ID that contains the term.                                             |
| parameters.termId        | string | true  | Glossary term ID to link to.                                                    |
| parameters.entryLinkType | string | false | Entry link type. Defaults to definition link type.                              |
