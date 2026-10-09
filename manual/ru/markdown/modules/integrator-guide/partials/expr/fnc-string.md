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
