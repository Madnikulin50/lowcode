# Генерация кода

Документация из YAML-определений (события, REST и т.п.) генерируется YAML-кодогенератором:

```bash
make codegen-legacy && ~/go/bin/lowcode-codegen -d ../path/to/docs/repo
```