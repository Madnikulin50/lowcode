# Изменение встроенных скриптов

## Изменение встроенных скриптов

Если вы хотите изменить скрипт автоматизации, определённый другим расширением (например, скрипты CRM), у вас есть два варианта: создать изменённую копию или перезаписать скрипты.

### Создание изменённой копии

.Чтобы изменить расширение:
1. скопируйте исходный код расширения (**клонируйте** репозиторий или **скопируйте** файлы),
1. измените исходный код по своему усмотрению,
1. разверните свою версию вместо оригинальной.

### Перезапись скриптов

Corredor присваивает уникальные имена каждому скрипту автоматизации.
Имя генерируется из пути к файлу.

.Например:
```bash
```
# The CRM extension
/ server-scripts
  / crm
    / Lead
      / SetLabel.js

Скрипту `SetLabel.js` в качестве имени присваивается `/server-scripts/crm/Lead/SetLabel.js:default`.

Чтобы перезаписать скрипт `SetLabel.js`, необходимо определить скрипт, которому будет присвоено то же имя (по сути, имеющий тот же путь).

.Например:
```bash
```
# The CRM extension
/ package.json
/ node_modules
/ server-scripts
  / crm
    / Lead
      / SetLabel.js # 👈 We're targeting this one
      / AnotherScript.js

# Your extension
/ package.json
/ node_modules
/ server-scripts
  # To overwrite something in the CRM extension
  / crm
    / Lead
      / SetLabel.js # 👈 This version will replace the CRM version

  # The rest of your code goes here
  / extension
    / Lead
      / SomeScript.js

!!! important
    Чтобы это работало, необходимо убедиться, что ваше расширение *подключается после* расширения, которое вы хотите изменить.
    
    .Например:
    [source,.env]
    ----
    # This will NOT work; the CRM is included after
    CORREDOR_EXT_SEARCH_PATHS="/your-ext:/crm"
    
    # This will work; the CRM is included before
    CORREDOR_EXT_SEARCH_PATHS="/crm:/your-ext"
    ----
