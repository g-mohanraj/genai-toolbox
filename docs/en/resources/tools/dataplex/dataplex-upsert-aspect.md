---
title: "dataplex-upsert-aspect"
type: docs
weight: 1
description: >
  A "dataplex-upsert-aspect" tool creates or updates an aspect on a Dataplex entry.
aliases:
- /resources/tools/dataplex-upsert-aspect
---

## About

A `dataplex-upsert-aspect` tool attaches or updates an aspect on an entry in
Dataplex Universal Catalog. It uses the Entry `Update` API with an `aspects`
update mask so you can either add a brand new aspect or overwrite an existing
one in a single call.

It's compatible with the following sources:

- [dataplex](../../sources/dataplex.md)

`dataplex-upsert-aspect` requires:

- `entry` - Full resource name of the target entry in the format
  `projects/{project}/locations/{location}/entryGroups/{entryGroup}/entries/{entry}`.
- `aspectType` - Aspect type reference. Accepts either the full resource name
  (`projects/{project}/locations/{location}/aspectTypes/{aspectTypeId}`) or the
  dotted reference (`{project}.{location}.{aspectTypeId}`).
- `data` - JSON object that matches the schema of the aspect type.

It also supports the following optional parameters:

- `path` - Aspect path to bind to a nested resource (for example
  `Schema.column_name`). Leave empty to attach to the entry root.
- `deleteMissingAspects` - When true, removes other aspects in the same type
  and path range that are not provided in the request. Defaults to `false`.
- `allowMissingEntry` - Creates the entry if it does not already exist.
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
  dataplex-upsert-aspect:
    kind: dataplex-upsert-aspect
    source: my-dataplex-source
    description: Create or update an aspect on a Dataplex entry.
```

## Reference

| **field**            | **type** | **required** | **description**                                                                                                          |
|----------------------|:--------:|:------------:|--------------------------------------------------------------------------------------------------------------------------|
| kind                 |  string  |     true     | Must be "dataplex-upsert-aspect".                                                                                        |
| source               |  string  |     true     | Name of the source the tool should execute on.                                                                           |
| description          |  string  |     true     | Description of the tool that is passed to the LLM.                                                                       |
| parameters.entry     |  string  |     true     | Entry resource name to update.                                                                                           |
| parameters.aspectType|  string  |     true     | Aspect type reference (full resource name or dotted reference).                                                          |
| parameters.data      |   map    |     true     | JSON payload for the aspect.                                                                                             |
| parameters.path      |  string  |     false    | Path within the entry to bind the aspect to.                                                                             |
| parameters.deleteMissingAspects | boolean | false | Remove other aspects in the same range that are missing from the request. Defaults to false.                             |
| parameters.allowMissingEntry | boolean | false | Create the entry if it is missing. Defaults to false.                                                                    |
