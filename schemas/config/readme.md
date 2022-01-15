# Config file schemas

These schemas describe RoadRunner configuration file and used by:

- <https://github.com/SchemaStore/schemastore>

Schemas naming agreement: `<version_major>.<version_minor>.schema.json`.

## Contributing guide

If you want to modify the existing schema - your changes **MUST** be backward compatible. If your changes break backward compatibility - you **MUST** create a new schema file with a fresh version and "register" it in a [schemas catalog][schemas_catalog].

[schemas_catalog]:https://github.com/SchemaStore/schemastore/blob/master/src/api/json/catalog.json
