# Импорт переводов ресурсов

!!! important
    После импорта переводов вам необходимо перезапустить сервер, чтобы изменения вступили в силу.
    Переводы хранятся в памяти процесса, запускающего сервер.
    Мы рассматриваем улучшение этого в будущих версиях.


!!! caution
    Перед импортом больших наборов данных (особенно в продакшен) рекомендуется сделать резервную копию базы данных и протестировать конфигурацию локально или на staging-сервере.


.Для импорта переводов ресурсов вам необходимо:
1. Подготовить файлы-источники,
1. выполнить команду.

## Файлы-источники

Подготовьте каталог, в котором будут находиться все ваши файлы-источники.

.`/import`:
```
```
/import
  /lowcode::compose:module.yaml
  /lowcode::compose:namespace.yaml
  /resource-translation.yaml

Файл `resource-translation.yaml` содержит переводы ресурсов, которые вы хотите импортировать, а два других файла (`lowcode::compose:module.yaml` и `lowcode::compose:namespace.yaml`) содержат определения ресурсов, для которых применяются переводы.

!!! important
    Текущий импортёр переводов ресурсов требует предоставить определения ресурсов, для которых вы применяете переводы.
    Мы рассматриваем улучшение этого в будущих версиях


## Запуск импорта

Используйте [команду CLI import](modules/devops-guide/pages/examples/cli/references/cli-reference.md#import) для импорта ваших переводов ресурсов.

Аргументом команды CLI должен быть путь к файлам-источникам; в нашем случае выше это `/import`.

```
```
!!! note
    По умолчанию команда CLI import пропускает уже существующие ресурсы.
    
    .Вы можете управлять поведением с помощью следующих флагов:
    --merge-left-existing             Update any existing values; existing data takes priority. Default skips.
    --merge-right-existing            Update any existing values; new data takes priority. Default skips.
    --replace-existing                Replace any existing values. Default skips.
    ----
    


.Пример выполнения команды CLI для нашего случая:
```bash
```
lowcode-server import /import
