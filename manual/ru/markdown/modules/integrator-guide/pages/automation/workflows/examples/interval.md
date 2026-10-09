# Интервал
:page-noindex: true

В этом примере предположим, что мы хотим отправить письмо всем пользователям каждое Рождество в полночь.
Для этого мы используем триггер типа `onInterval` с интервалом `0 0 25 12 *`

![](automation/workflows/examples/workflow-example-interval.png)

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onInterval`
*** *enabled**: checked
*** *constraints**: 
**** *interval**: `0 0 25 12 *`
2. **(2) Iterate over Users**:
*** *type**: `Users`
*** *results**:
**** *user target**: `user`
3. **(3) Send Email**:
*** *type**: `Email`
*** *arguments**:
**** *subject**:
****** **value type**: constant
****** **value**: `Merry christmas`
**** *to**:
****** **value type**: expression
****** **value**: `user.email`
**** *plain**:
****** **value type**: constant
****** **value**: `Merry christmas`
4. **(4) Done**:
******
