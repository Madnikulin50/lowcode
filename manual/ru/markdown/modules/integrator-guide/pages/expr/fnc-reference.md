# Expression Function Reference

:leveloffset: +1

# Type functions

## `coalesce(...Any)`

Функция `coalesce` возвращает первое не-`null` значение.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = coalesce(null, 0, 1, 2)
|
```json
```
{
  "out": 0
}

|===

## `isEmpty(Any)`

Функция `isEmpty` возвращает `true`, если значение пусто.

Что считается пустым, зависит от типа, см. таблицу ниже.

[cols="1,1,1"]
|===
|Type |Value |Result

.2+|***Array***|[]|true
|~undefined~|true
|***Array***|~undefined~|true
|***Any***|~undefined~|false
|***Any***|[]|true
|***Vars***|~undefined~|true
|***Boolean***|~undefined~|true
|***DateTime***|~undefined~|true
|***Float***|~undefined~|true
|***Integer***|~undefined~|true
|***UnsignedInteger***|~undefined~|true
|***String***|~undefined~|true
|***User***|~undefined~|false
|===

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": ""
}
|
```
```
out = isEmpty(in)
|
```json
```
{
  "in": "",
  "out": true
}


|
```json
```
{
  "in": 0
}
|
```
```
out = isEmpty(in)
|
```json
```
{
  "in": 0,
  "out": true
}

|===

## `isNil(Any)`

Функция `isNil` возвращает `true`, если заданное значение равно `null`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": null
}
|
```
```
out = isNil(in)
|
```json
```
{
  "in": null,
  "out": true
}


|
```json
```
{
  "in": 10
}
|
```
```
out = isNil(in)
|
```json
```
{
  "in": 10,
  "out": false
}

|===


# Array functions

## `push(array, ...elements)`

Функция `push` добавляет указанные элементы в конец массива и возвращает новый массив.

**Исходный массив остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": []
}
|
```
```
new = push(arr, 1)
|
```json
```
{
  "arr": [],
  "new": [1]
}


|
```json
```
{
  "arr": []
}
|
```
```
new = push(arr, 1, 2, 3)
|
```json
```
{
  "arr": [],
  "new": [1, 2, 3]
}


|
```json
```
{
  "stuff": [1, 2, 3]
}
|
```
```
new = push([], 1, 2, 3)
|
```json
```
{
  "new": [1, 2, 3]
}

|===


## `pop(array)`

Функция `pop` возвращает последний элемент массива.

**Исходный массив остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": [1, 2, 3]
}
|
```
```
last = pop(arr)
|
```json
```
{
  "arr": [1, 2, 3],
  "last": 3
}


|
```json
```
{
  "arr": []
}
|
```
```
last = pop(arr)
|
```json
```
{
  "arr": [],
  "last": null
}

!!! note
    При использовании типа `Any` значением будет `null`.
    При использовании другого типа значением будет нулевое значение этого типа.


|
```json
```
{

}
|
```
```
last = pop([])
|
```json
```
{
  "last": null
}

!!! note
    При использовании типа `Any` значением будет `null`.
    При использовании другого типа значением будет нулевое значение этого типа.


|===

## `shift(array)`

Функция `shift` возвращает первый элемент массива.

**Исходный массив остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": [1, 2, 3]
}
|
```
```
first = shift(arr)
|
```json
```
{
  "arr": [1, 2, 3],
  "first": 1
}


|
```json
```
{
  "arr": []
}
|
```
```
first = shift(arr)
|
```json
```
{
  "arr": [],
  "first": null
}

!!! note
    При использовании типа `Any` значением будет `null`.
    При использовании другого типа значением будет нулевое значение этого типа.


|
```json
```
{

}
|
```
```
first = shift([])
|
```json
```
{
  "first": null
}

!!! note
    При использовании типа `Any` значением будет `null`.
    При использовании другого типа значением будет нулевое значение этого типа.


|===

## `count(array, ...elements)`

Функция `count` возвращает количество вхождений заданных элементов.

Функция `count` возвращает длину массива, если элемент не указан.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
nm = count(arr, "a")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "nm": 1
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
nm = count(arr, "a", "b")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "nm": 2
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
nm = count(arr)
|
```json
```
{
  "arr": ["a", "b", "c"],
  "nm": 3
}
|===

## `length(array)`

Функция `length` возвращает длину входной строки.

## `has(arr, ...elements)`

Функция `has` проверяет, содержит ли переданный массив хотя бы один из элементов.

Функция возвращает `true`, если элементы найдены, иначе возвращает `false`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = has(arr, "a")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": true
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = has(arr, "/", "b")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": true
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = has(arr, "a", "b", "c")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": true
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = has(arr, "/")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": false
}
|===

## `hasAll(arr, ...elements)`

Функция `hasAll` проверяет, содержит ли переданный массив **все** элементы.

Функция возвращает `true`, если элементы найдены, иначе возвращает `false`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = hasAll(arr, "a")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": true
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = hasAll(arr, "/", "b")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": false
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = hasAll(arr, "a", "b", "c")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": true
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
out = hasAll(arr, "/")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "out": false
}
|===

## `find(arr, elements)`

Функция `find` возвращает позицию заданного элемента (нумерация с нуля).
Если элемент не существует, функция возвращает `-1`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
index = find(arr, "a")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "index": 0
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
index = find(arr, "b")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "index": 1
}


|
```json
```
{
  "arr": ["a", "b", "c"]
}
|
```
```
index = find(arr, "/")
|
```json
```
{
  "arr": ["a", "b", "c"],
  "index": -1
}
|===

## `sort(array, descending)`

Функция `sort` возвращает отсортированный массив: по возрастанию, если второй параметр — `false`, или по убыванию, если второй параметр — `true`.

**Исходный массив остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": ["c", "a", "b"]
}
|
```
```
sorted = sort(arr, false)
|
```json
```
{
  "arr": ["c", "a", "b"],
  "sorted": ["a", "b", "c"]
}


|
```json
```
{
  "arr": ["c", "a", "b"]
}
|
```
```
sorted = sort(arr, true)
|
```json
```
{
  "arr": ["c", "a", "b"],
  "sorted": ["c", "b", "a"]
}
|===

## `splice(array, start, end)`

Функция `splice` возвращает новый массив с элементами от начального индекса до (не включая) конечного индекса.

**Исходный массив остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "arr": ["1", "2", "3"]
}
|
```
```
new = splice(arr, 0, 1)
|
```json
```
{
  "arr": ["1", "2", "3"],
  "new": ["1"]
}


|
```json
```
{
  "arr": [4, 5, 6, 7]
}
|
```
```
new = splice(arr, 1, -1)
|
```json
```
{
  "arr": [4, 5, 6, 7],
  "new": [5, 6, 7]
}

!!! note
    Если `end` отрицательный, будут возвращены все элементы от начального индекса до конца массива.

|===


# String functions

## `trim(string)`

Функция `trim` удаляет все начальные и конечные пробельные символы, определённые стандартом Unicode.

.Список пробельных символов Unicode:
- `U+0020`: пробел,
- `U+00A0`: неразрывный пробел,
- `U+1680`: огамический пробел,
- `U+180E`: монгольский разделитель гласных,
- `U+2000`: пробел en quad,
- `U+2001`: пробел em quad,
- `U+2002`: пробел en,
- `U+2003`: пробел em,
- `U+2004`: пробел three-per-em,
- `U+2005`: пробел four-per-em,
- `U+2006`: пробел six-per-em,
- `U+2007`: пробел figure,
- `U+2008`: пробел punctuation,
- `U+2009`: тонкий пробел,
- `U+200A`: волосяной пробел,
- `U+200B`: нулевой ширины пробел,
- `U+202F`: узкий неразрывный пробел,
- `U+205F`: средний математический пробел,
- `U+3000`: идеографический пробел,
- `U+FEFF`: нулевой ширины неразрывный пробел.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "\t abcd \t"
}
|
```
```
out = trim(in)
|
```json
```
{
  "in": "\t abcd \t",
  "out": "abcd"
}


|
```json
```
{
  "in": "\t ab cd \t"
}
|
```
```
out = trim(in)
|
```json
```
{
  "in": "\t abcd \t",
  "out": "ab cd"
}
|===

## `trimLeft(string, remove)`

Функция `trimLeft` удаляет указанные символы из начала строки.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = trimLeft(in, "ab")
|
```json
```
{
  "in": "abcd",
  "out": "cd"
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = trimLeft(in, "abcd")
|
```json
```
{
  "in": "abcd",
  "out": ""
}
|===

## `trimRight(string, remove)`

Функция `trimRight` удаляет указанные символы из конца строки.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = trimFight(in, "cd")
|
```json
```
{
  "in": "abcd",
  "out": "ab"
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = trimFight(in, "abcd")
|
```json
```
{
  "in": "abcd",
  "out": ""
}
|===

## `toLower(string)`

Функция `toLower` возвращает новую строку, в которой заглавные буквы преобразованы в строчные.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "Abcd"
}
|
```
```
out = toLower(in)
|
```json
```
{
  "in": "Abcd",
  "out": "abcd"
}


|
```json
```
{
  "in": "ABCD"
}
|
```
```
out = toLower(in)
|
```json
```
{
  "in": "ABCD",
  "out": "abcd"
}
|===

## `toUpper(string)`

Функция `toUpper` возвращает новую строку, в которой строчные буквы преобразованы в заглавные.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "Abcd"
}
|
```
```
out = toUpper(in)
|
```json
```
{
  "in": "Abcd",
  "out": "abcd"
}


|
```json
```
{
  "in": "ABCD"
}
|
```
```
out = toUpper(in)
|
```json
```
{
  "in": "ABCD",
  "out": "abcd"
}
|===

## `shortest(string1, ...strings)`

Функция `shortest` возвращает самую короткую строку из переданных аргументов.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{

}
|
```
```
out = shortest("a", "aa", "aaa")
|
```json
```
{
  "out": "a"
}


|
```json
```
{

}
|
```
```
out = shortest("a")
|
```json
```
{
  "out": "a"
}
|===

## `longest(arg1, arg2, ...a`

Функция `longest` возвращает самую длинную строку из переданных аргументов.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{

}
|
```
```
out = longest("a", "aa", "aaa")
|
```json
```
{
  "out": "aaa"
}


|
```json
```
{

}
|
```
```
out = longest("a")
|
```json
```
{
  "out": "a"
}
|===

## `format(format, ...arguments)`

Функция `format` возвращает новую строку, созданную из заданного шаблона и аргументов.
Обратитесь к [String Formatting](modules/integrator-guide/partials/expr/expr/string-formatting.md) за более подробной информацией.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{

}
|
```
```
out = format("string %s, and float %.2f", "hi", 12.345)
|
```json
```
{
  "out": "string hi, and float 12.35"
}
|===

## `title(string)`

Функция `title` преобразует первый символ в заглавный.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = title(in)
|
```json
```
{
  "in": "abcd",
  "out": "Abcd"
}


|
```json
```
{
  "in": "abcd efg"
}
|
```
```
out = title(in)
|
```json
```
{
  "in": "abcd efg",
  "out": "Abcd efg"
}
|===

## `untitle(string)`

Функция `untitle` делает обратное тому, что делает `title(string)`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "Abcd"
}
|
```
```
out = untitle(in)
|
```json
```
{
  "in": "Abcd",
  "out": "abcd"
}


|
```json
```
{
  "in": "Abcd efg"
}
|
```
```
out = untitle(in)
|
```json
```
{
  "in": "Abcd efg",
  "out": "abcd efg"
}
|===

## `repeat(string, count)`

Функция `repeat` возвращает новую строку, в которой исходная повторяется `count` раз.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = count(in, 1)
|
```json
```
{
  "in": "abcd",
  "out": "abcd"
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = count(in, 2)
|
```json
```
{
  "in": "abcd",
  "out": "abcdabcd"
}
|===

## `replace(string, old, new,`

Функция `replace` возвращает копию строки s, в которой первые n неперекрывающихся вхождений old заменены на new.
Если old пусто, совпадение происходит в начале строки и после каждой последовательности UTF-8, что даёт до k+1 замен для строки из k рун.
Если n < 0, ограничения на количество замен нет.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "foo foo foo"
}
|
```
```
out = replace(in, "foo", "bar", 1)
|
```json
```
{
  "in": "foo foo foo",
  "out": "bar foo foo"
}


|
```json
```
{
  "in": "foo foo foo"
}
|
```
```
out = replace(in, "foo", "bar", 0)
|
```json
```
{
  "in": "foo foo foo",
  "out": "foo foo foo"
}


|
```json
```
{
  "in": "foo foo foo"
}
|
```
```
out = replace(in, "foo", "bar", -1)
|
```json
```
{
  "in": "foo foo foo",
  "out": "bar bar bar"
}
|===

## `isUrl(string)`

Функция `isUrl` проверяет, является ли заданная строка корректным URL-адресом.
Если строка является корректным URL, функция возвращает `true`, иначе возвращает `false`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "https://www.example.tld"
}
|
```
```
out = isUrl(in)
|
```json
```
{
  "in": "https://www.example.tld",
  "out": true
}


|
```json
```
{
  "in": "uhoh"
}
|
```
```
out = isUrl(in)
|
```json
```
{
  "in": "uhoh",
  "out": false
}
|===

## `isEmail(string)`

Функция `isEmail` проверяет, является ли заданная строка корректным email-адресом.
Если строка является корректным email, функция возвращает `true`, иначе возвращает `false`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "hi@email.tld"
}
|
```
```
out = isEmail(in)
|
```json
```
{
  "in": "hi@email.tld",
  "out": true
}


|
```json
```
{
  "in": "uhoh"
}
|
```
```
out = isEmail(in)
|
```json
```
{
  "in": "uhoh",
  "out": false
}
|===

## `split(string, separator)`

Функция `split` возвращает массив строк, полученный разделением исходной строки по разделителю.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "aa.bb.cc"
}
|
```
```
out = split(in, ".")
|
```json
```
{
  "in": "aa.bb.cc",
  "out": ["aa", "bb", "cc"]
}


|
```json
```
{
  "in": "aa.bb.cc"
}
|
```
```
out = split(in, "/")
|
```json
```
{
  "in": "aa.bb.cc",
  "out": ["aa.bb.cc"]
}
|===

## `join(strings, separator)`

Функция `join` объединяет строки из массива в одну строку, разделённую разделителем.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": ["aa", "bb", "cc"]
}
|
```
```
out = join(in, ".")
|
```json
```
{
  "in": ["aa", "bb", "cc"],
  "out": "aa.bb.cc"
}
|===

## `hasSubstring(string, substring, case)`

Функция `hasSubstring` проверяет, содержит ли заданная строка подстроку.

Когда третий аргумент равен `true`, функция учитывает регистр, иначе — не учитывает.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasSubstring(in, "bc", false)
|
```json
```
{
  "in": "abcd",
  "out": true
}


|
```json
```
{
  "in": "aBCd"
}
|
```
```
out = hasSubstring(in, "bc", false)
|
```json
```
{
  "in": "aBCd",
  "out": true
}


|
```json
```
{
  "in": "aBCd"
}
|
```
```
out = hasSubstring(in, "bc", true)
|
```json
```
{
  "in": "aBCd",
  "out": false
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasSubstring(in, "xy", false)
|
```json
```
{
  "in": "abcd",
  "out": false
}
|===

## `substring(string, start, end)`

Функция `substring` возвращает подстроку заданной строки.

И `start`, и `end` включительно (`[start, end]`).

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = substring(in, 1, 2)
|
```json
```
{
  "in": "abcd",
  "out": "bc"
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = substring(in, 1, -1)
|
```json
```
{
  "in": "abcd",
  "out": "bcd"
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = substring(in, 4, -1)
|
```json
```
{
  "in": "abcd",
  "out": ""
}
|===

## `hasPrefix(string, prefix)`

Функция `hasPrefix` проверяет, содержит ли заданная строка префикс.
Если префикс существует, функция возвращает `true`, иначе возвращает `false`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasPrefix(in, "ab")
|
```json
```
{
  "in": "abcd",
  "out": true
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasPrefix(in, "cd")
|
```json
```
{
  "in": "abcd",
  "out": false
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasPrefix(in, "xy")
|
```json
```
{
  "in": "abcd",
  "out": false
}
|===

## `hasSuffix(string, prefix)`

Функция `hasSuffix` проверяет, содержит ли заданная строка суффикс.
Если суффикс существует, функция возвращает `true`, иначе возвращает `false`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasSuffix(in, "cd")
|
```json
```
{
  "in": "abcd",
  "out": true
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasSuffix(in, "ab")
|
```json
```
{
  "in": "abcd",
  "out": false
}


|
```json
```
{
  "in": "abcd"
}
|
```
```
out = hasSuffix(in, "xy")
|
```json
```
{
  "in": "abcd",
  "out": false
}
|===

## `shorten(string, type, count)`

Функция `shorten` обрезает заданную строку до `count` символов или слов, когда `type` установлен в `char`.

Строка дополняется многоточием после точки обрезания.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "This is a whole sentence"
}
|
```
```
out = shorten(in, "word", 4)
|
```json
```
{
  "in": "This is a whole sentence",
  "out": "This is a whole …"
}
|===

## `camelize(string)`

Функция `camelize` возвращает новую строку в форме camelCase.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "Foo bar"
}
|
```
```
out = camelize(in)
|
```json
```
{
  "in": "Foo bar",
  "out": "fooBar"
}
|===

## `snakify(string)`

Функция `snakify` возвращает новую строку в форме snake_case.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "Foo bar baz"
}
|
```
```
out = snakify(in)
|
```json
```
{
  "in": "Foo bar baz",
  "out": "foo*bar*baz"
}
|===

## `match(string, regex)`

Функция `match` проверяет, соответствует ли строка заданному регулярному выражению.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result
|===

## `base64encode(string)``

Функция `base64encode` возвращает строку, закодированную в base64.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result
|===

## `length(string)`

Функция `length` возвращает длину входной строки.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result
|===

## `count(string, ...characters)`

Функция `count` возвращает количество вхождений заданных символов.

Функция `count` возвращает длину строки, если символ не указан.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "foo"
}
|
```
```
out = count(in, "o")
|
```json
```
{
  "in": "foo",
  "out": 2
}
|===


# Numeric functions

## `min(...number)`

Функция `min` возвращает число с наименьшим значением.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = min(1, 2, 3, -1)
|
```json
```
{
  "out": -1
}

|===

## `max(...number)`

Функция `max` возвращает число с наибольшим значением.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = min(1, 2, 3, -1)
|
```json
```
{
  "out": 3
}

|===

## `round(number, places)`

Функция `round` округляет `number` до указанного количества `places`.
Функция возвращает число с плавающей точкой.

!!! tip
    Чтобы избавиться от плавающей точки, просто приведите к `Integer`.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 10.123
}
|
```
```
out = round(in, 2)
|
```json
```
{
  "in": 10.123,
  "out": 10.12
}


|
```json
```
{
  "in": 10.123
}
|
```
```
out = round(in, 0)
|
```json
```
{
  "in": 10.123,
  "out": 10
}
|===

## `floor(number)`

Функция `floor` округляет число вниз до ближайшего целого.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 10.123
}
|
```
```
out = floor(in)
|
```json
```
{
  "in": 10.123,
  "out": 10
}

|===


## `ceil(number)`

Функция `ceil` округляет число вверх до ближайшего целого.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 10.123
}
|
```
```
out = ceil(in)
|
```json
```
{
  "in": 10.123,
  "out": 11
}
|===


## `abs(number)`

Функция `abs` возвращает абсолютное значение переданного `number`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": -10
}
|
```
```
out = abs(in)
|
```json
```
{
  "in": -10
  "out": 10
}


|
```json
```
{
  "in": 10
}
|
```
```
out = abs(in)
|
```json
```
{
  "in": 10
  "out": 10
}
|===

## `log(number)`

Функция `log` возвращает десятичный логарифм (по основанию 10) заданного `number`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 100
}
|
```
```
out = log(in)
|
```json
```
{
  "in": 100
  "out": 2
}
|===

## `pow(number, exp)`

Функция `pow` возвращает `number`, возведённое в степень `exp`.

|**pow**   |`pow(number, number)`|Функция возвращает x**y, x в степени y, см. [math.Pow](https://golang.org/pkg/math/#Pow)|`pow(2, 3)` результат: 8

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 2,
  "exp": 3
}
|
```
```
out = pow(in, exp)
|
```json
```
{
  "in": 2,
  "exp": 3,
  "out": 8
}

|===


## `sqrt(number)`

Функция `sqrt` возвращает квадратный корень заданного `number`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 4
}
|
```
```
out = sqrt(in)
|
```json
```
{
  "in": 4,
  "out": 2
}

|===


## `sum(...number)`

Функция `sum` возвращает сумму всех переданных аргументов.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = sum(1, 2, 3, -1)
|
```json
```
{
  "out": 5
}
|===


## `average(...number)`

Функция `average` возвращает среднее значение из переданных аргументов.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = average(1, 2)
|
```json
```
{
  "out": 1.5
}
|===


## `random(a, b?)`

Функция `random` возвращает случайное число.
При вызове с одним аргументом (`random(to)`) случайное число находится между `0` и `to`.
При вызове с двумя аргументами (`random(from, to)`) случайное число находится между `from` и `to`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "to": 10
}
|
```
```
out = random(to)
|
```json
```
{
  "to": 10,
  "out": 3.412
}


|
```json
```
{
  "from": 5,
  "to": 10
}
|
```
```
out = random(to)
|
```json
```
{
  "from": 5,
  "to": 10,
  "out": 5.9
}
|===

## `int(Any)`

Функция `int` приводит аргумент к `Integer`.
Если значение не может быть приведено, функция возвращает `0`.

!!! tip
    Когда вы присваиваете значение переменной, оно автоматически приводится к указанному типу.
    Явное приведение требуется только при передаче аргументов.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "10"
}
|
```
```
out = int(in)
|
```json
```
{
  "in": "10",
  "out": 10
}


|
```json
```
{
  "in": "NO"
}
|
```
```
out = int(in)
|
```json
```
{
  "in": "NO",
  "out": 0
}
|===


# Date and time functions

## `earliest(DateTime, ...DateTime)`

|**earliest**     |`earliest(arg1, arg2, ...argN)`|Функция возвращает самую раннюю дату и время.|`earliest(datefield1, datefield2)` результат: "1970-01-01T00:00:00"


Функция `earliest` возвращает самую раннюю `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = earliest(parseISOTime("2019-01-01T00:00:00Z"), parseISOTime("2020-01-01T00:00:00Z"))
|
```json
```
{
  "out": "2019-01-01T00:00:00Z"
}

|===

## `latest`

Функция `latest` возвращает самую позднюю `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = latest(parseISOTime("2019-01-01T00:00:00Z"), parseISOTime("2020-01-01T00:00:00Z"))
|
```json
```
{
  "out": "2020-01-01T00:00:00Z"
}

|===

## `parseISOTime`

Функция `parseISOTime` разбирает метку времени в формате ISO.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z"
}
|
```
```
out = parseISOTime("2020-01-01T00:00:00Z")
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "out": "2020-01-01T00:00:00Z"
}

|===

## `modTime`

Функция `modTime` возвращает новую `DateTime` с добавленной `duration`.
Функция `modTime` взаимодействует с временной частью `DateTime`.

Используйте `modDate`, `modWeek`, `modMonth` или `modYear`, если вы хотите изменить более крупные компоненты.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1h"
}
|
```
```
out = modTime(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1h",
  "out": "2020-01-01T01:00:00Z",
}

|===

## `modDate(datetime, days)`

Функция `modDate` возвращает новую `DateTime` с добавленными `days`.
Функция `modDate` взаимодействует с частью даты (дни) `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modDate(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2020-01-02T00:00:00Z",
}

|===

## `modWeek(datetime, weeks)`

Функция `modWeek` возвращает новую `DateTime` с добавленными `weeks`.
Функция `modWeek` взаимодействует с частью даты (дни) `DateTime`.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modWeek(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2020-01-08T00:00:00Z",
}

|===

## `modMonth(datetime, months)`

Функция `modMonth` возвращает новую `DateTime` с добавленными `months`.
Функция `modMonth` взаимодействует с частью месяца `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modMonth(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2020-02-01T00:00:00Z",
}

|===

## `modYear(datetime, years)`

Функция `modYear` возвращает новую `DateTime` с добавленными `years`.
Функция `modYear` взаимодействует с частью года `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modYear(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2021-00-01T00:00:00Z",
}

|===

## `parseDuration`

Функция `parseDuration` возвращает разобранную длительность из заданной строки.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2h"
}
|
```
```
out = parseDuration(in)
|
```json
```
{
  "in": "2h",
  "out": "2h0m0s"
}

|===

## `strftime(datetime, pattern)`

Функция `strftime` возвращает отформатированную `DateTime` на основе заданного `pattern`.
Обратитесь к [Datetime Formatting](modules/integrator-guide/partials/expr/expr/datetime-formatting.md) за более подробной информацией.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z"
}
|
```
```
out = strftime(in, "%Y-%m-%d")
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "out": "2020-01-01"
}

|===

## `isLeapYear(datetime)`

Функция `isLeapYear` возвращает `true`, если заданный `DateTime` является високосным годом.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z"
}
|
```
```
out = isLeapYear(in)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "out": true
}


|
```json
```
{
  "in": "2019-01-01T00:00:00Z"
}
|
```
```
out = isLeapYear(in)
|
```json
```
{
  "in": "2019-01-01T00:00:00Z",
  "out": false
}

|===

## `now`

Функция `now` возвращает текущую `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = now()
|
```json
```
{
  "out": "2022-02-24T18:00:00Z"
}

|===

## `isWeekDay`

|**isWeekDay**    |`isWeekDay(datetime)`|Функция возвращает true, если указанный день является будним днём.|`isWeekDay(datefield)` результат: `true`


Функция `isWeekDay` возвращает `true`, если заданный `DateTime` является будним днём.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2022-02-24T00:00:00Z"
}
|
```
```
out = isWeekDay(in)
|
```json
```
{
  "in": "2022-02-24T00:00:00Z",
  "out": true
}


|
```json
```
{
  "in": "2022-02-26T00:00:00Z"
}
|
```
```
out = isWeekDay(in)
|
```json
```
{
  "in": "2022-02-26T00:00:00Z",
  "out": false
}

|===

## `sub(from, to)`

Функция `sub` возвращает разницу между двумя `DateTime` в миллисекундах.

`from` должен быть больше `to`; в противном случае функция вызовет ошибку.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "from": "2022-02-02T00:00:00Z",
  "to": "2022-02-01T00:00:00Z"
}
|
```
```
out = sub(from, to)
|
```json
```
{
  "from": "2022-02-02T00:00:00Z",
  "to": "2022-02-01T00:00:00Z"
  "out": 86400000
}

|===


# KV functions

!!! caution
    Результирующий тип функции KV зависит от первого аргумента.
    Нельзя передавать несколько различных типов KV (KV, KVV, Vars) в одну и ту же функцию.
    
    Например; `merge(KV, KVV, Vars)` не допускается.


## `set(kv, k, v)`

Функция `set` присваивает значение заданной переменной типа `KV`.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": {}
}
|
```
```
out = set(in, "foo", "bar")
|
```json
```
{
  "in": {},
  "out": {"foo": "bar"}
}

|===

## `merge(kv, ...kv)`

|**merge**  |`merge(KV, arg1, ...argN)`|Функция объединяет все заданные типы KV в один тип KV.|`merge(&KVV{"foo": ["foo"]}, &KVV{"bar": ["bar"]})` результат: `&KVV{"foo": ["foo"], "bar": ["bar"]}`, То же самое для KV и Vars.

Функция `merge` объединяет все переменные типа `KV` в один `KV`.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "a": {"foo": "foo"},
  "b": {"bar": "baz"}
}
|
```
```
out = merge(a, b)
|
```json
```
{
  "a": {"foo": "foo"},
  "b": {"bar": "baz"}
  "out": {"foo": "foo", "bar": "baz"}
}

|===

## `filter(kv, ...include)`

Функция `filter` возвращает новый `KV`, содержащий только указанные ключи.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"}
}
|
```
```
out = filter(in, "foo")
|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"},
  "out": {"foo": "foo"}
}

|===

## `omit(kv, ...exclude)`

Функция `omit` возвращает новый `KV` без указанных ключей.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"}
}
|
```
```
out = filter(in, "foo")
|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"},
  "out": {"bar": "bar"}
}

|===


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


:leveloffset: -1
