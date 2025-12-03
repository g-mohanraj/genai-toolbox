---
title: "dataplex-create-aspect-type"
type: docs
weight: 1
description: >
  A "dataplex-create-aspect-type" tool creates a custom Dataplex aspect type.
aliases:
- /resources/tools/dataplex-create-aspect-type
---

## About

A `dataplex-create-aspect-type` tool calls the Dataplex Catalog
`CreateAspectType` API to define a new aspect type in a project/region. Supply
the aspect type ID and a JSON payload that matches the Dataplex aspect type
schema (including `metadataTemplate`).

It is compatible with the following sources:

- [dataplex](../../sources/dataplex.md)

`dataplex-create-aspect-type` requires:

- `location` - Region where the aspect type will be created (e.g.
  `us-central1`).
- `aspectTypeId` - Identifier for the aspect type to create.
- `aspectType` - Aspect type definition payload (JSON object). Must include a
  `metadataTemplate`.
- `validateOnly` - When true, validates the request without creating the aspect
  type. Defaults to false.

## Example

```yaml
tools:
  dataplex-create-aspect-type:
    kind: dataplex-create-aspect-type
    source: my-dataplex-source
    description: Create a Dataplex aspect type.
```

## Reference

| **field**   | **type** | **required** | **description**                                                             |
|-------------|:--------:|:------------:|-----------------------------------------------------------------------------|
| kind        |  string  |     true     | Must be "dataplex-create-aspect-type".                                      |
| source      |  string  |     true     | Name of the source the tool should execute on.                              |
| description |  string  |     true     | Description of the tool that is passed to the LLM.                          |
| parameters.location | string | true | Dataplex location where the aspect type is created.                         |
| parameters.aspectTypeId | string | true | Aspect type identifier.                                                     |
| parameters.aspectType | map | true | Aspect type payload (must include `metadataTemplate`).                       |
| parameters.validateOnly | boolean | false | Validate without creating. Defaults to false.                               |
