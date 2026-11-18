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
			if (Object.hasOwn(schema, key)) {
				stripIds(schema[key], false);
			}
		}
	}
}

// Load the main schema and all its referenced schemas
const dereferenced = await $RefParser.dereference('./config/3.0.schema.json');

// Remove $id and $schema from anything but the root
stripIds(dereferenced, true);

// Conditional rules require properties from the parent schema. See https://ajv.js.org/strict-mode.html#defined-required-properties
const ajv = new Ajv2019({strict: true, strictRequired: false, allErrors: true})
const validator = ajv.compile(dereferenced)

const data = fs.readFileSync('../.rr.yaml', 'utf-8');
const format = { format: '%time% %message%', time_format: '15:04:05' };
const fileOptions = { file_logger_options: { max_size: 10 } };
const brokers = ['127.0.0.1:9092'];
const cases = [
	{ name: 'no logs section', config: { version: '3' }, valid: true },
	{ name: 'global log format', config: { version: '3', logs: format }, valid: true },
	{ name: 'channel log format', config: { version: '3', logs: { channels: { http: format } } }, valid: true },
	{ name: 'removed global file options', config: { version: '3', logs: fileOptions }, valid: false },
	{ name: 'removed channel file options', config: { version: '3', logs: { channels: { http: fileOptions } } }, valid: false },
	{ name: 'flat AMQP address', config: { version: '3', amqp: { addr: 'amqp://localhost:5672/' } }, valid: false },
	{ name: 'Kafka TLS with a root CA', config: { version: '3', kafka: { brokers, tls: { root_ca: '/ssl/ca.crt' } } }, valid: true },
	{ name: 'Kafka TLS with only a key', config: { version: '3', kafka: { brokers, tls: { key: '/ssl/key.pem' } } }, valid: false },
	{ name: '.rr.yaml', config: yaml.load(data), valid: true },
];

for (const { name, config, valid } of cases) {
	if (validator(config) !== valid) {
		throw new Error(`${name}: expected valid=${valid}, errors: ${JSON.stringify(validator.errors, null, 2)}`);
	}
}

console.log('No errors found in schemas.');
