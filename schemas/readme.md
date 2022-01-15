# Schemas

This directory contains public schemas for the most important parts of application.

**Do not rename or remove this directory or any file or directory inside.**

- You can validate existing config file using the following command:

 ```bash
 $ docker run --rm -v "$(pwd):/src" -w "/src" node:14-alpine sh -c \
     "npm install -g ajv-cli && \
     ajv validate --all-errors --verbose \
       -s ./schemas/config/2.0.schema.json \
       -d ./.rr*.y*ml"
 ```
