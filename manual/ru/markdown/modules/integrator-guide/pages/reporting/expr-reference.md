# Справочник по выражениям

## Конъюнкция

[cols="1s,5a"]
|===
| [#expr-conjunction-and]#[expr-conjunction-and,and](#expr-conjunction-and,and)#
|
Возвращает `true`, если все элементы равны `true`, иначе `false`.

.[#expr-conjunction-and-sig]#[expr-conjunction-and-sig,Сигнатуры:](#expr-conjunction-and-sig,Сигнатуры:)#
- `<Boolean>[ && <Boolean>[...]]` => `<Boolean>`
- `<Boolean>[ and <Boolean>[...]]` => `<Boolean>`
- `and(<Boolean>[, <Boolean>, ...]])` => `<Boolean>`

.[#expr-conjunction-and-ex-a]#[expr-conjunction-and-ex-a,Примеры выражений:](#expr-conjunction-and-ex-a,Примеры выражений:)#
- `bool_field && true`
- `(field == 'value') && bool_field`
- `bool_field and true`
- `(field == 'value') and bool_field`

.[#expr-conjunction-and-ex-b]#[expr-conjunction-and-ex-b,Примеры функций:](#expr-conjunction-and-ex-b,Примеры функций:)#
- `and(bool_field, true)`
- `and((field == 'value'), bool_field)`

| [#expr-conjunction-or]#[expr-conjunction-or,or](#expr-conjunction-or,or)#
|
Возвращает `true`, если любой элемент равен `true`, иначе `false`.

.[#expr-conjunction-or-sig]#[expr-conjunction-or-sig,Сигнатуры:](#expr-conjunction-or-sig,Сигнатуры:)#
- `<Boolean>[ \|\| <Boolean>[...]]` => `<Boolean>`
- `<Boolean>[ or <Boolean>[...]]` => `<Boolean>`
- `or(<Boolean>[, <Boolean>, ...]])` => `<Boolean>`

.[#expr-conjunction-or-ex-a]#[expr-conjunction-or-ex-a,Примеры выражений:](#expr-conjunction-or-ex-a,Примеры выражений:)#
- `bool_field \|\| true`
- `(field == 'value') \|\| bool_field`
- `bool_field or true`
- `(field == 'value') or bool_field`

.[#expr-conjunction-or-ex-b]#[expr-conjunction-or-ex-b,Примеры функций:](#expr-conjunction-or-ex-b,Примеры функций:)#
- `or(bool_field, true)`
- `or((field == 'value'), bool_field)`

|===

## Сравнение

[cols="1s,5a"]
|===
| [#expr-comparison-eq]#[expr-comparison-eq,equals](#expr-comparison-eq,equals)#
|
Возвращает `true`, если типы сравнимы и значения совпадают.

.[#expr-comparison-eq-sig]#[expr-comparison-eq-sig,Сигнатуры:](#expr-comparison-eq-sig,Сигнатуры:)#
- `<Any> = <Any>` => `<Boolean>`
- `<Any> == <Any>` => `<Boolean>`
- `<Any> === <Any>` => `<Boolean>`
- `eq(<Any>, <Any>)` => `<Boolean>`

.[#expr-comparison-eq-ex-a]#[expr-comparison-eq-ex-a,Примеры выражений:](#expr-comparison-eq-ex-a,Примеры выражений:)#
- `field == 'value'`
- `number == 10`
- `year(dateField) == year(now())`

.[#expr-comparison-eq-ex-b]#[expr-comparison-eq-ex-b,Примеры функций:](#expr-comparison-eq-ex-b,Примеры функций:)#
- `eq(field, 'value')`
- `eq(number, 10)`
- `eq(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-comparison-ne]#[expr-comparison-ne,not equals](#expr-comparison-ne,not equals)#
|
Возвращает `true`, если типы не сравнимы или значения не совпадают.

.[#expr-comparison-ne-sig]#[expr-comparison-ne-sig,Сигнатуры:](#expr-comparison-ne-sig,Сигнатуры:)#
- `<Any> != <Any>` => `<Boolean>`
- `<Any> !== <Any>` => `<Boolean>`
- `ne(<Any>, <Any>)` => `<Boolean>`

.[#expr-comparison-ne-ex-a]#[expr-comparison-ne-ex-a,Примеры выражений:](#expr-comparison-ne-ex-a,Примеры выражений:)#
- `field != 'value'`
- `number != 10`
- `year(dateField) != year(now())`

.[#expr-comparison-ne-ex-b]#[expr-comparison-ne-ex-b,Примеры функций:](#expr-comparison-ne-ex-b,Примеры функций:)#
- `ne(field, 'value')`
- `ne(number, 10)`
- `ne(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-comparison-lt]#[expr-comparison-lt,less than](#expr-comparison-lt,less than)#
|
Возвращает `true`, если левое значение меньше правого.

.[#expr-comparison-lt-sig]#[expr-comparison-lt-sig,Сигнатуры:](#expr-comparison-lt-sig,Сигнатуры:)#
- `<Any> < <Any>` => `<Boolean>`
- `lt(<Any>, <Any>)` => `<Boolean>`

.[#expr-comparison-lt-ex-a]#[expr-comparison-lt-ex-a,Примеры выражений:](#expr-comparison-lt-ex-a,Примеры выражений:)#
- `field < 'value'`
- `number < 10`
- `year(dateField) < year(now())`

.[#expr-comparison-lt-ex-b]#[expr-comparison-lt-ex-b,Примеры функций:](#expr-comparison-lt-ex-b,Примеры функций:)#
- `lt(field, 'value')`
- `lt(number, 10)`
- `lt(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-comparison-le]#[expr-comparison-le,less than or equal to](#expr-comparison-le,less than or equal to)#
|
Возвращает `true`, если левое значение меньше или равно правому.

.[#expr-comparison-le-sig]#[expr-comparison-le-sig,Сигнатуры:](#expr-comparison-le-sig,Сигнатуры:)#
- `<Any> <= <Any>` => `<Boolean>`
- `le(<Any>, <Any>)` => `<Boolean>`

.[#expr-comparison-le-ex-a]#[expr-comparison-le-ex-a,Примеры выражений:](#expr-comparison-le-ex-a,Примеры выражений:)#
- `field <= 'value'`
- `number <= 10`
- `year(dateField) <= year(now())`

.[#expr-comparison-le-ex-b]#[expr-comparison-le-ex-b,Примеры функций:](#expr-comparison-le-ex-b,Примеры функций:)#
- `le(field, 'value')`
- `le(number, 10)`
- `le(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-comparison-gt]#[expr-comparison-gt,greater than](#expr-comparison-gt,greater than)#
|
Возвращает `true`, если левое значение больше правого.

.[#expr-comparison-gt-sig]#[expr-comparison-gt-sig,Сигнатуры:](#expr-comparison-gt-sig,Сигнатуры:)#
- `<Any> > <Any>` => `<Boolean>`
- `gt(<Any>, <Any>)` => `<Boolean>`

.[#expr-comparison-gt-ex-a]#[expr-comparison-gt-ex-a,Примеры выражений:](#expr-comparison-gt-ex-a,Примеры выражений:)#
- `field > 'value'`
- `number > 10`
- `year(dateField) > year(now())`

.[#expr-comparison-gt-ex-b]#[expr-comparison-gt-ex-b,Примеры функций:](#expr-comparison-gt-ex-b,Примеры функций:)#
- `gt(field, 'value')`
- `gt(number, 10)`
- `gt(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-comparison-ge]#[expr-comparison-ge,greater or equal than](#expr-comparison-ge,greater or equal than)#
|
Возвращает `true`, если левое значение больше или равно правому.

.[#expr-comparison-ge-sig]#[expr-comparison-ge-sig,Сигнатуры:](#expr-comparison-ge-sig,Сигнатуры:)#
- `<Any> >= <Any>` => `<Boolean>`
- `ge(<Any>, <Any>)` => `<Boolean>`

.[#expr-comparison-ge-ex-a]#[expr-comparison-ge-ex-a,Примеры выражений:](#expr-comparison-ge-ex-a,Примеры выражений:)#
- `field >= 'value'`
- `number >= 10`
- `year(dateField) >= year(now())`

.[#expr-comparison-ge-ex-b]#[expr-comparison-ge-ex-b,Примеры функций:](#expr-comparison-ge-ex-b,Примеры функций:)#
- `ge(field, 'value')`
- `ge(number, 10)`
- `ge(year(dateField), year(now()))`

|===

## Сравнение строк

[cols="1s,5a"]
|===
| [#expr-str-comparison-like]#[expr-str-comparison-like,like](#expr-str-comparison-like,like)#
|
Возвращает `true`, если левое значение соответствует шаблону справа.

.[#expr-str-comparison-like-sig]#[expr-str-comparison-like-sig,Сигнатуры:](#expr-str-comparison-like-sig,Сигнатуры:)#
- `<String> like <String>` => `<Boolean>`
- `like(<String>, <String>)` => `<Boolean>`

.[#expr-str-comparison-like-ex-a]#[expr-str-comparison-like-ex-a,Примеры выражений:](#expr-str-comparison-like-ex-a,Примеры выражений:)#
- `field like 'value'`
- `field like 'va_ue'`
- `name like 'test%'`

.[#expr-str-comparison-like-ex-b]#[expr-str-comparison-like-ex-b,Примеры функций:](#expr-str-comparison-like-ex-b,Примеры функций:)#
- `like(field, 'value')`
- `like(field, 'va_ue')`
- `like(name, 'test%')`

|===

[cols="1s,5a"]
|===
| [#expr-str-comparison-nlike]#[expr-str-comparison-nlike,not like](#expr-str-comparison-nlike,not like)#
|
Возвращает `true`, если левое значение не соответствует шаблону справа.

.[#expr-str-comparison-nlike-sig]#[expr-str-comparison-nlike-sig,Сигнатуры:](#expr-str-comparison-nlike-sig,Сигнатуры:)#
- `<String> not like <String>` => `<Boolean>`
- `nlike(<String>, <String>)` => `<Boolean>`

.[#expr-str-comparison-nlike-ex-a]#[expr-str-comparison-nlike-ex-a,Примеры выражений:](#expr-str-comparison-nlike-ex-a,Примеры выражений:)#
- `field not like 'value'`
- `field not like 'va_ue'`
- `name not like 'test%'`

.[#expr-str-comparison-nlike-ex-b]#[expr-str-comparison-nlike-ex-b,Примеры функций:](#expr-str-comparison-nlike-ex-b,Примеры функций:)#
- `nlike(field, 'value')`
- `nlike(field, 'va_ue')`
- `nlike(name, 'test%')`

|===

## Операции со строками

|===
| [#expr-string-concat]#[expr-string-concat,concat](#expr-string-concat,concat)#
|
.[#expr-str-concat-sig]#[expr-str-concat-sig,Сигнатуры:](#expr-str-concat-sig,Сигнатуры:)#
- `concat(<String>[, <String>, ...]])` => `<String>`

.[#expr-str-concat-ex-a]#[expr-str-concat-ex-a,Примеры выражений:](#expr-str-concat-ex-a,Примеры выражений:)#


.[#expr-str-concat-ex-b]#[expr-str-concat-ex-b,Примеры функций:](#expr-str-concat-ex-b,Примеры функций:)#
- `concat('a', 'b')`
- `concat(field, 'b')`
|===

## Арифметика

[cols="1s,5a"]
|===
| [#expr-arithmetic-add]#[expr-arithmetic-add,add](#expr-arithmetic-add,add)#
|
Складывает два числа.

.[#expr-arithmetic-add-sig]#[expr-arithmetic-add-sig,Сигнатуры:](#expr-arithmetic-add-sig,Сигнатуры:)#
- `<Number> + <Number>` => `<Number>`
- `add(<Number>, <Number>)` => `<Number>`

.[#expr-arithmetic-add-ex-a]#[expr-arithmetic-add-ex-a,Примеры выражений:](#expr-arithmetic-add-ex-a,Примеры выражений:)#
- `100 + 20`
- `number + 10`
- `year(dateField) + year(now())`

.[#expr-arithmetic-add-ex-b]#[expr-arithmetic-add-ex-b,Примеры функций:](#expr-arithmetic-add-ex-b,Примеры функций:)#
- `add(100, 20)`
- `add(number, 10)`
- `add(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-arithmetic-sub]#[expr-arithmetic-sub,sub](#expr-arithmetic-sub,sub)#
|
Вычитает два числа.

.[#expr-arithmetic-sub-sig]#[expr-arithmetic-sub-sig,Сигнатуры:](#expr-arithmetic-sub-sig,Сигнатуры:)#
- `<Number> - <Number>` => `<Number>`
- `sub(<Number>, <Number>)` => `<Number>`

.[#expr-arithmetic-sub-ex-a]#[expr-arithmetic-sub-ex-a,Примеры выражений:](#expr-arithmetic-sub-ex-a,Примеры выражений:)#
- `100 - 20`
- `number - 10`
- `year(dateField) - year(now())`

.[#expr-arithmetic-sub-ex-b]#[expr-arithmetic-sub-ex-b,Примеры функций:](#expr-arithmetic-sub-ex-b,Примеры функций:)#
- `sub(100, 20)`
- `sub(number, 10)`
- `sub(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-arithmetic-mult]#[expr-arithmetic-mult,mult](#expr-arithmetic-mult,mult)#
|
Умножает два числа.

.[#expr-arithmetic-mult-sig]#[expr-arithmetic-mult-sig,Сигнатуры:](#expr-arithmetic-mult-sig,Сигнатуры:)#
- `<Number> * <Number>` => `<Number>`
- `mult(<Number>, <Number>)` => `<Number>`

.[#expr-arithmetic-mult-ex-a]#[expr-arithmetic-mult-ex-a,Примеры выражений:](#expr-arithmetic-mult-ex-a,Примеры выражений:)#
- `100 * 20`
- `number * 10`
- `year(dateField) * year(now())`

.[#expr-arithmetic-mult-ex-b]#[expr-arithmetic-mult-ex-b,Примеры функций:](#expr-arithmetic-mult-ex-b,Примеры функций:)#
- `mult(100, 20)`
- `mult(number, 10)`
- `mult(year(dateField), year(now()))`

|===

[cols="1s,5a"]
|===
| [#expr-arithmetic-div]#[expr-arithmetic-div,div](#expr-arithmetic-div,div)#
|
Делит два числа друг на друга.

.[#expr-arithmetic-div-sig]#[expr-arithmetic-div-sig,Сигнатуры:](#expr-arithmetic-div-sig,Сигнатуры:)#
- `<Number> / <Number>` => `<Number>`
- `div(<Number>, <Number>)` => `<Number>`

.[#expr-arithmetic-div-ex-a]#[expr-arithmetic-div-ex-a,Примеры выражений:](#expr-arithmetic-div-ex-a,Примеры выражений:)#
- `100 / 20`
- `number / 10`
- `year(dateField) / year(now())`

.[#expr-arithmetic-div-ex-b]#[expr-arithmetic-div-ex-b,Примеры функций:](#expr-arithmetic-div-ex-b,Примеры функций:)#
- `div(100, 20)`
- `div(number, 10)`
- `div(year(dateField), year(now()))`

|===

## Агрегация

[cols="1s,5a"]
|===
| [#expr-aggregation-count]#[expr-aggregation-count,count](#expr-aggregation-count,count)#
|
Возвращает количество элементов.

.[#expr-aggregation-count-sig]#[expr-aggregation-count-sig,Сигнатуры:](#expr-aggregation-count-sig,Сигнатуры:)#
- `count()` => `<Number>`

.[#expr-aggregation-count-ex-a]#[expr-aggregation-count-ex-a,Примеры выражений:](#expr-aggregation-count-ex-a,Примеры выражений:)#
- `count()`

.[#expr-aggregation-count-ex-b]#[expr-aggregation-count-ex-b,Примеры функций:](#expr-aggregation-count-ex-b,Примеры функций:)#
- `count()`
|===


[cols="1s,5a"]
|===
| [#expr-aggregation-sum]#[expr-aggregation-sum,sum](#expr-aggregation-sum,sum)#
|
Возвращает сумму элементов.

.[#expr-aggregation-sum-sig]#[expr-aggregation-sum-sig,Сигнатуры:](#expr-aggregation-sum-sig,Сигнатуры:)#
- `sum(<Number>)` => `<Number>`

.[#expr-aggregation-sum-ex-a]#[expr-aggregation-sum-ex-a,Примеры выражений:](#expr-aggregation-sum-ex-a,Примеры выражений:)#
- `field + current_column`

.[#expr-aggregation-sum-ex-b]#[expr-aggregation-sum-ex-b,Примеры функций:](#expr-aggregation-sum-ex-b,Примеры функций:)#
- `sum(field)`
|===

[cols="1s,5a"]
|===
| [#expr-aggregation-max]#[expr-aggregation-max,max](#expr-aggregation-max,max)#
|
Возвращает максимум из двух значений.

.[#expr-aggregation-max-sig]#[expr-aggregation-max-sig,Сигнатуры:](#expr-aggregation-max-sig,Сигнатуры:)#
- `max(<Number>)` => `<Number>`

.[#expr-aggregation-max-ex-a]#[expr-aggregation-max-ex-a,Примеры выражений:](#expr-aggregation-max-ex-a,Примеры выражений:)#
- `max(field)`

.[#expr-aggregation-max-ex-b]#[expr-aggregation-max-ex-b,Примеры функций:](#expr-aggregation-max-ex-b,Примеры функций:)#
- `max(field)`
|===

[cols="1s,5a"]
|===
| [#expr-aggregation-min]#[expr-aggregation-min,min](#expr-aggregation-min,min)#
|
Возвращает минимум из двух значений.

.[#expr-aggregation-min-sig]#[expr-aggregation-min-sig,Сигнатуры:](#expr-aggregation-min-sig,Сигнатуры:)#
- `min(<Number>)` => `<Number>`

.[#expr-aggregation-min-ex-a]#[expr-aggregation-min-ex-a,Примеры выражений:](#expr-aggregation-min-ex-a,Примеры выражений:)#
- `min(field)`

.[#expr-aggregation-min-ex-b]#[expr-aggregation-min-ex-b,Примеры функций:](#expr-aggregation-min-ex-b,Примеры функций:)#
- `min(field)`
|===

[cols="1s,5a"]
|===
| [#expr-aggregation-avg]#[expr-aggregation-avg,avg](#expr-aggregation-avg,avg)#
|
Возвращает среднее двух значений.

.[#expr-aggregation-avg-sig]#[expr-aggregation-avg-sig,Сигнатуры:](#expr-aggregation-avg-sig,Сигнатуры:)#
- `avg(<Number>)` => `<Number>`

.[#expr-aggregation-avg-ex-a]#[expr-aggregation-avg-ex-a,Примеры выражений:](#expr-aggregation-avg-ex-a,Примеры выражений:)#
- `avg(field)`

.[#expr-aggregation-avg-ex-b]#[expr-aggregation-avg-ex-b,Примеры функций:](#expr-aggregation-avg-ex-b,Примеры функций:)#
- `avg(field)`
|===

## Манипуляции с датами

[cols="1s,5a"]
|===
| [#expr-date-now]#[expr-date-now,now](#expr-date-now,now)#
|
Возвращает текущую временную метку.

.[#expr-date-now-sig]#[expr-date-now-sig,Сигнатуры:](#expr-date-now-sig,Сигнатуры:)#
- `now()` => `<DateTime>`

.[#expr-date-now-ex-a]#[expr-date-now-ex-a,Примеры выражений:](#expr-date-now-ex-a,Примеры выражений:)#
- `now()`

|===

[cols="1s,5a"]
|===
| [#expr-date-quarter]#[expr-date-quarter,quarter](#expr-date-quarter,quarter)#
|
Возвращает квартал временной метки.

.[#expr-date-quarter-sig]#[expr-date-quarter-sig,Сигнатуры:](#expr-date-quarter-sig,Сигнатуры:)#
- `quarter(<DateTime>)` => `<Number>`

.[#expr-date-quarter-ex-a]#[expr-date-quarter-ex-a,Примеры выражений:](#expr-date-quarter-ex-a,Примеры выражений:)#
- `quarter(date_field)`
- `quarter(now())`
- `quarter(date('2021-01-01T01:00:00Z'))`

|===

[cols="1s,5a"]
|===
| [#expr-date-year]#[expr-date-year,year](#expr-date-year,year)#
|
Возвращает год временной метки.

.[#expr-date-year-sig]#[expr-date-year-sig,Сигнатуры:](#expr-date-year-sig,Сигнатуры:)#
- `year(<DateTime>)` => `<Number>`

.[#expr-date-year-ex-a]#[expr-date-year-ex-a,Примеры выражений:](#expr-date-year-ex-a,Примеры выражений:)#
- `year(date_field)`
- `year(now())`
- `year(date('2021-01-01T01:00:00Z'))`

|===

[cols="1s,5a"]
|===
| [#expr-date-month]#[expr-date-month,month](#expr-date-month,month)#
|
Возвращает месяц временной метки.

.[#expr-date-month-sig]#[expr-date-month-sig,Сигнатуры:](#expr-date-month-sig,Сигнатуры:)#
- `month(<DateTime>)` => `<Number>`

.[#expr-date-month-ex-a]#[expr-date-month-ex-a,Примеры выражений:](#expr-date-month-ex-a,Примеры выражений:)#
- `month(date_field)`
- `month(now())`
- `month(date('2021-01-01T01:00:00Z'))`

|===

## Приведение типов

[cols="1s,5a"]
|===
| [#expr-cast-float]#[expr-cast-float,float](#expr-cast-float,float)#
|
Приводит значение к числу с плавающей точкой.

.[#expr-cast-float-sig]#[expr-cast-float-sig,Сигнатуры:](#expr-cast-float-sig,Сигнатуры:)#
- `float(<Any>)` => `<Float>`

.[#expr-cast-float-ex-a]#[expr-cast-float-ex-a,Примеры выражений:](#expr-cast-float-ex-a,Примеры выражений:)#
- `float(some_field)`
- `float('10.9')`
- `avg(cost) > float(10.9)`

|===

[cols="1s,5a"]
|===
| [#expr-cast-string]#[expr-cast-string,string](#expr-cast-string,string)#
|
Приводит значение к строке.

!!! caution
    Длина строки для PostgreSQL и MySQL ограничена `8192` символами.


.[#expr-cast-string-sig]#[expr-cast-string-sig,Сигнатуры:](#expr-cast-string-sig,Сигнатуры:)#
- `string(<Any>)` => `<String>`

.[#expr-cast-string-ex-a]#[expr-cast-string-ex-a,Примеры выражений:](#expr-cast-string-ex-a,Примеры выражений:)#
- `string(some_field)`
- `string('10.9')`

|===

[cols="1s,5a"]
|===
| [#expr-cast-int]#[expr-cast-int,int](#expr-cast-int,int)#
|
Приводит значение к целому числу.

.[#expr-cast-int-sig]#[expr-cast-int-sig,Сигнатуры:](#expr-cast-int-sig,Сигнатуры:)#
- `int(<Any>)` => `<Int>`

.[#expr-cast-int-ex-a]#[expr-cast-int-ex-a,Примеры выражений:](#expr-cast-int-ex-a,Примеры выражений:)#
- `int(some_field)`
- `int('10.9')`
- `avg(cost) > int(10.9)`

|===

[cols="1s,5a"]
|===
| [#expr-date-date]#[expr-date-date,date](#expr-date-date,date)#
|
Приводит значение к временной метке.

.[#expr-date-date-sig]#[expr-date-date-sig,Сигнатуры:](#expr-date-date-sig,Сигнатуры:)#
- `date(<Ant>)` => `<DateTime>`

.[#expr-date-date-ex-a]#[expr-date-date-ex-a,Примеры выражений:](#expr-date-date-ex-a,Примеры выражений:)#
- `date(some_field)`
- `date('2021-01-01T01:00:00Z')`

|===
