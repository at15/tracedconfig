# tracedconfig

A Go library for managing configuration with traceable value origins and contextual error reporting.

## TODO

Interface

- [ ] How to define the config struct/interface to pass it around? I don't want to use struct tags. Maybe like `zod` in typescript?

Formats

- [ ] JSON need to update slowjson code base on interface
- [ ] YAML https://github.com/goccy/go-yaml
- [ ] TOML https://github.com/pelletier/go-toml?tab=readme-ov-file#contextualized-errors