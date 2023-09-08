# lacking-cli

Command Line Interface for the [lacking](https://github.com/mokiat/lacking) engine.

## Build Distributions

The CLI allows you to build various distributions.

A project needs to include an `app.yml` descriptor file which is used by this tool to bundle the project.

```yaml
id: example
long_id: com.example.acme
name: Example
version: 1.3.0
description: Example application
contact: Example <john.doe@example.com>
main: ./cmd/example
icon: resources/images/icon.png
copyright: Copyright © 2023, Example. All rights reserved.
```

Generated content will be placed inside the `dist` folder of the project.

### Linux / amd64

```sh
lacking dist linux <project_dir>
```

### MacOS / amd64

```sh
lacking dist macos <project_dir>
```
