# Access control functons

<a id="isDescendantOf"></a>
## `isDescendantOf(userID, resourceOwnerID, paths ...string)`

Функция `isDescendantOf` проверяет, находится ли `userID` строго выше `resourceOwnerID` в дереве организации.
Если `userID` и `resourceOwnerID` находятся в одной группе пользователей, результатом оценки будет `false`.

!!! important
    Поскольку результат этой функции зависит от состояния системы, следующие примеры являются входными и выходными значениями.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{}
|
```
```
is = isDescendantOf(currentUser, resource.ownerID, "path 1")
|
```json
```
{
  "is": false
}


|
```json
```
{}
|
```
```
is = isDescendantOf(userA, userB)
|
```json
```
{
  "is": true
}

|===


<a id="isDescendantOfC"></a>
## `isDescendantOfC(userID, resourceOwnerID, paths ...string)`

Функция `isDescendantOfC` является сокращением для стандартного пути ``.
Функция возвращает `true`, если `userID` находится строго выше `resourceOwnerID` по стандартному пути ``.

!!! important
    Поскольку результат этой функции зависит от состояния системы, следующие примеры являются входными и выходными значениями.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{}
|
```
```
is = isDescendantOfC(currentUser, resource.ownerID)
|
```json
```
{
  "is": false
}


|
```json
```
{}
|
```
```
is = isDescendantOfC(userA, userB)
|
```json
```
{
  "is": true
}

|===
<a id="isDescendantOfR"></a>
## `isDescendantOfR(userID, resourceOwnerID, paths ...string)`

Функция `isDescendantOfR` является сокращением для стандартного пути ``.
Функция возвращает `true`, если `userID` находится строго выше `resourceOwnerID` по стандартному пути ``.

!!! important
    Поскольку результат этой функции зависит от состояния системы, следующие примеры являются входными и выходными значениями.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{}
|
```
```
is = isDescendantOfR(currentUser, resource.ownerID)
|
```json
```
{
  "is": false
}


|
```json
```
{}
|
```
```
is = isDescendantOfR(userA, userB)
|
```json
```
{
  "is": true
}

|===
<a id="isDescendantOfU"></a>
## `isDescendantOfU(userID, resourceOwnerID, paths ...string)`

Функция `isDescendantOfU` является сокращением для стандартного пути ``.
Функция возвращает `true`, если `userID` находится строго выше `resourceOwnerID` по стандартному пути ``.

!!! important
    Поскольку результат этой функции зависит от состояния системы, следующие примеры являются входными и выходными значениями.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{}
|
```
```
is = isDescendantOfU(currentUser, resource.ownerID)
|
```json
```
{
  "is": false
}


|
```json
```
{}
|
```
```
is = isDescendantOfU(userA, userB)
|
```json
```
{
  "is": true
}

|===
<a id="isDescendantOfD"></a>
## `isDescendantOfD(userID, resourceOwnerID, paths ...string)`

Функция `isDescendantOfD` является сокращением для стандартного пути ``.
Функция возвращает `true`, если `userID` находится строго выше `resourceOwnerID` по стандартному пути ``.

!!! important
    Поскольку результат этой функции зависит от состояния системы, следующие примеры являются входными и выходными значениями.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{}
|
```
```
is = isDescendantOfD(currentUser, resource.ownerID)
|
```json
```
{
  "is": false
}


|
```json
```
{}
|
```
```
is = isDescendantOfD(userA, userB)
|
```json
```
{
  "is": true
}

|===
