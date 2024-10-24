# Schemas

This directory contains public schemas for the most important parts of application.

**Do not rename or remove this directory or any file or directory inside.**

- You can validate existing config file using the following command:

 ```bash
docker run --rm -v "$(pwd):/src" -w "/src" node:20-alpine sh -c \
     "npm install -g ajv-cli && \
     ajv validate --all-errors --verbose \
       -s ./schemas/config/3.0.schema.json \
       --spec=draft2019 \
       -d ./.rr*.y*ml"
 ```
