import $RefParser from "@apidevtools/json-schema-ref-parser";
import Ajv2019 from "ajv/dist/2019.js"
import fs from 'fs';
import yaml from 'js-yaml';

function stripIds(schema, first) {
	if (schema !== null && typeof schema === 'object') {
		if (!first) {
			// Every referenced schema we pull in should have its $id and $schema stripped, or ajv complains
			// Skip the root object, as that should retain the $schema and $id
			delete schema.$id;
			delete schema.$schema;
		}
		for (const key in schema) {
			if (schema.hasOwnProperty(key)) {
				stripIds(schema[key], false);
			}
		}
	}
}

// Load the main schema and all its referenced schemas
const dereferenced = await $RefParser.dereference('./config/3.0.schema.json');

// Remove $id and $schema from anything but the root
stripIds(dereferenced, true);

const ajv = new Ajv2019({strict: true})
const validator = ajv.compile(dereferenced)

const data = fs.readFileSync('../.rr.yaml', 'utf-8');
const schema = yaml.load(data);

// Validate the file
if (!validator(schema)) {
	throw new Error("Errors: " + JSON.stringify(validator.errors))
}
