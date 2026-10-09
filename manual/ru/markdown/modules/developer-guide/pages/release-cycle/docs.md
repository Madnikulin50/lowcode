# LowCoooode Документация

## При подготовке пререлиза

1. создайте новую ветку `yyyy-qq-x` (`$V`)
1. включите ветку в `antora-playbook.yml`
1. измените версию в `src/antora.yml`

```
```
version: '$V'
prerelease: -develop
display_version: '$V-develop'

## При выпуске релиза

1. переключитесь на ветку, которую вы хотите выпустить (`$V`)
1. измените версию в `src/antora.yml`
1. измените asciidoc-атрибут `page-latest` в `antora-playbook.yaml`

```
```
version: '$V'
