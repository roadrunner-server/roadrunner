# Schemas

This directory contains public schemas for the most important parts of application.

**Do not rename or remove this directory or any file or directory inside.**

- You can validate existing config file using the following command from the project root.

 ```bash
docker run --rm -v "$(pwd):/src" -w "/src" node:22-alpine sh -c \
     "cd schemas && npm install && node test.js"
 ```
