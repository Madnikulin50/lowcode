# Метка времени
:page-noindex: true

В этом примере предположим, что мы хотим поздравить всех пользователей с Новым годом.
Все письма должны быть отправлены ровно в полночь.
Для этого мы используем триггер типа `onTimestamp` с меткой времени `2021-01-01T00:00:00Z`

![](automation/workflows/examples/workflow-example-timestamp.png)

******
.Детали шагов рабочего процесса:
[%collapsible.result]
1. **(1) Test trigger**:
*** *resource**: `System`
*** *event**: `onTimestamp`
*** *enabled**: checked
*** *constraints**: 
**** *timestamp**: `2021-01-01T00:00:00Z`
2. **(2) Iterate over Users**:
*** *type**: `Users`
*** *results**:
**** *user target**: `user`
3. **(3) Send Email**:
*** *type**: `Email`
*** *arguments**:
**** *subject**:
****** **value type**: constant
****** **value**: `Happy new year`
**** *to**:
****** **value type**: expression
****** **value**: `user.email`
**** *plain**:
****** **value type**: constant
****** **value**: `Happy new year`
4. **(4) Done**:
******
