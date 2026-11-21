# Выполнение скриптов автоматизации

Рабочий процесс позволяет выполнять скрипты автоматизации, зарегистрированные на сервере Corredor.
Возможность выполнения скриптов автоматизации может пригодиться, когда нужно реализовать сложную бизнес-логику, но при этом сохранить визуальное описание рабочего процесса.

## Скрипт автоматизации

!!! note
    Подробности о внутреннем устройстве скриптов автоматизации и о том, как развернуть их на сервере Corredor, см. в [Automation Scripts](modules/integrator-guide/pages/automation/workflows/automation/automation-scripts/index.md).


Скрипт автоматизации должен определять **ручной триггер для системного ресурса**.
В качестве отправной точки можете использовать следующий пример.

./server-scripts/invoked-from-the-workflow.js
```js
```
export default {
  label: 'Called From Workflow',
  description: 'This script will be called from workflow function',
  triggers ({ on }) {
    /**
     - Due to how the Corredor scripting system is designed right now,
     - triggers still need to be defined (even if the script is
     - executed explicitly from the workflow).
     */
    return on('manual').for('system')
  },

  exec (args, { logger }) {
    logger.info('success')
  }
}

## Рабочий процесс

Чтобы выполнить скрипт автоматизации, добавьте шаг функции {ICON*WORKFLOW*FUNCTION}, который вызывает функцию `Corredor automation script executor`.
Вставьте **ссылку на скрипт** в аргумент script.

.Ссылку на скрипт можно получить следующим образом:
- Перейдите в меню:LowCoooode Admin[Automation,Corredor scripts] и найдите скрипт автоматизации, который хотите выполнить, затем скопируйте его ссылку.
- Составьте её вручную по шаблону `/server-scripts/path-to/the-script/script-name.js:default`.
Для примера выше ссылка выглядит так: `/server-scripts/invoked-from-the-workflow.js:default`.

!!! note
    Если вы не видите свои скрипты автоматизации, обратитесь к документации [Automation Scripts](modules/integrator-guide/pages/automation/workflows/automation/automation-scripts/index.md).


Разместите триггер {ICON*WORKFLOW*TRIGGER} и соедините его с определённым ранее шагом функции.

Чтобы проверить, всё ли работает правильно, нажмите на значок запуска в правом верхнем углу триггера.
Если всё работает правильно, после выполнения теста сервер Corredor залогирует строку `"success"`.

## Передача пользовательских аргументов

Добавьте шаг выражения {ICON*WORKFLOW*EXPRESSIONS} между триггером и шагом функции.

.Определите следующие выражения:
```
```
scriptArgs          (Vars) <1>
scriptArgs.from     (Boolean) = false <2>
scriptArgs.workflow (String)  = "some string" <3>
<1> Определите новый набор переменных, который будет передан в скрипт автоматизации.
<2> Определите переменную `from` со значением `false`.
<3> Определите переменную `workflow` со значением `"some string"`.

Обновите шаг функции и задайте аргументу `args` значение `scriptArgs` (переменные, определённые на шаге выражения).

.Скрипт автоматизации может получить доступ к подготовленным аргументам следующим образом:
```js
```
export default {
  label: 'Called From Workflow',
  description: 'This script will be called from workflow function',
  triggers ({ on }) {
    /**
     - Due to how the Corredor scripting system is currently designed,
     - triggers still need to be defined (even if the script is
     - executed explicitly from the workflow).
     */
    return on('manual').for('system')
  },

  exec (args, { logger }) {
    const { from, workflow } = args

    logger.info('these are the special arguments we received', { from, workflow })
  }
}

Чтобы проверить, всё ли работает правильно, нажмите на значок запуска в правом верхнем углу триггера.
Если всё работает правильно, сервер Corredor залогирует строку `'these are the special arguments we received' { from: false, workflow: 'some string' }`.

## Чтение результата выполнения

В шаге функции, выполняющем скрипт автоматизации, укажите целевую переменную для сбора выходных данных скрипта.
Скрипт автоматизации должен возвращать объект.

!!! caution
    Действуют стандартные правила скриптов автоматизации для возвращаемых значений.
    Если скрипт автоматизации возвращает `false`, мы считаем это сигналом `Abort`, и рабочий процесс завершается с ошибкой.


.Пример скрипта автоматизации, возвращающего результат:
```js
```
export default {

  exec (args, { logger }) {
    logger.info('returning some values from the script')
    return {
      a: 123,
      b: true
    }
  }
}
