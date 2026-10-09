<a id="datetime-formatting"></a>
# Date and time formatting

## `%Y`

Description::
    Возвращает год с веком в виде десятичного числа.
example::
    `strftime(dateField, "%Y")` результат: "1993"

## `%y`

Description::
    Возвращает год без века в виде десятичного числа (00-99).
example::
    `strftime(dateField, "%y")` результат: "93"

## `%C`

Description::
    Возвращает год / 100 в виде десятичного числа; однозначные числа дополняются нулём слева.
example::
    `strftime(dateField, "%C")` результат: "19"

## `%m`

Description::
    Возвращает месяц в виде десятичного числа (01-12).
example::
    `strftime(dateField, "%m")` результат: "02"

## `%B`

Description::
    Возвращает полное название месяца на национальном языке.
example::
    `strftime(dateField, "%B")` результат: "February"

## `%b`

Description::
    Возвращает сокращённое название месяца на национальном языке.
example::
    `strftime(dateField, "%b")` результат: "Feb"

## `%U`

Description::
    Возвращает номер недели в году (воскресенье — первый день недели) в виде десятичного числа (00-53).
example::
    `strftime(dateField, "%U")` результат: "05"

## `%V`

Description::
    Возвращает номер недели в году (понедельник — первый день недели) в виде десятичного числа (01-53).
example::
    `strftime(dateField, "%V")` результат: "05"

## `%W`

Description::
    Возвращает номер недели в году (понедельник — первый день недели) в виде десятичного числа (00-53).
example::
    `strftime(dateField, "%W")` результат: "05"

## `%A`

Description::
    Возвращает полное название дня недели на национальном языке.
example::
    `strftime(dateField, "%A")` результат: "Tuesday"

## `%a`

Description::
    Возвращает сокращённое название дня недели на национальном языке.
example::
    `strftime(dateField, "%a")` результат: "Tue"

## `%d`

Description::
    Возвращает день месяца в виде десятичного числа (01-31).
example::
    `strftime(dateField, "%d")` результат: "02"

## `%e`

Description::
    Возвращает день месяца в виде десятичного числа (1-31).
example::
    `strftime(dateField, "%e")` результат: " 2"

## `%j`

Description::
    Возвращает день года в виде десятичного числа (001-366).
example::
    `strftime(dateField, "%j")` результат: "033"

## `%u`

Description::
    Возвращает день недели (понедельник — первый день недели) в виде десятичного числа (1-7).
example::
    `strftime(dateField, "%u")` результат: "5"

## `%w`

Description::
    Возвращает день недели (воскресенье — первый день недели) в виде десятичного числа (0-6).
example::
    `strftime(dateField, "%w")` результат: "2"

## `%H`

Description::
    Возвращает час (24-часовой формат) в виде десятичного числа (00-23).
example::
    `strftime(dateField, "%H")` результат: "06"

## `%k`

Description::
    Возвращает час (24-часовой формат) в виде десятичного числа (0-23).
example::
    `strftime(dateField, "%k")` результат: " 6"

## `%I`

Description::
    Возвращает час (12-часовой формат) в виде десятичного числа (01-12).
example::
    `strftime(dateField, "%I")` результат: "06"

## `%l`

Description::
    Возвращает час (12-часовой формат) в виде десятичного числа (1-12).
example::
    `strftime(dateField, "%l")` результат: " 6"

## `%M`

Description::
    Возвращает минуты в виде десятичного числа (00-59).
example::
    `strftime(dateField, "%M")` результат: "00"

## `%S`

Description::
    Возвращает секунды в виде десятичного числа (00-60).
example::
    `strftime(dateField, "%S")` результат: "00"

## `%S`

Description::
    Возвращает миллисекунды в виде десятичного числа (000-999).
example::
    `strftime(dateField, "%S")` результат: "000"

## `%p`

Description::
    Возвращает национальное представление «ante meridiem» (до полудня) или «post meridiem» (после полудня).
example::
    `strftime(dateField, "%p")` результат: "AM"

## `%c`

Description::
    Возвращает национальное представление времени и даты.
example::
    `strftime(dateField, "%c")` результат: "Tue Feb  2 06:00:00 1993"

## `%X`

Description::
    Возвращает национальное представление времени.
example::
    `strftime(dateField, "%X")` результат: "06:00:00"

## `%x`

Description::
    Возвращает национальное представление даты.
example::
    `strftime(dateField, "%x")` результат: "02/02/93"

## `%Z`

Description::
    Возвращает название часового пояса.
example::
    `strftime(dateField, "%Z")` результат: "-0500"

## `%z`

Description::
    Возвращает смещение часового пояса от UTC.
example::
    `strftime(dateField, "%z")` результат: "-0500"

## `%n`

Description::
    Возвращает символ новой строки (\n).
example::
    `strftime(dateField, "%n")` результат: "\n"

## `%t`

Description::
    Возвращает символ табуляции.
example::
    `strftime(dateField, "%t")` результат: "\t"

## `%%`

Description::
    Возвращает символ %.
example::
    `strftime(dateField, "%%")` результат: "%"

## `%F`

Description::
    Эквивалентно %Y-%m-%d.
example::
    `strftime(dateField, "%F")` результат: "1993-02-02"

## `%D`

Description::
    Эквивалентно %m/%d/%y.
example::
    `strftime(dateField, "%D")` результат: "02/02/93"

## `%R`

Description::
    Эквивалентно %H:%M.
example::
    `strftime(dateField, "%R")` результат: "06:00"

## `%r`

Description::
    Эквивалентно %I:%M:%S %p.
example::
    `strftime(dateField, "%r")` результат: "06:00:00 AM"

## `%T`

Description::
    Эквивалентно %H:%M:%S.
example::
    `strftime(dateField, "%T")` результат: "06:00:00"

## `%v`

Description::
    Эквивалентно %e-%b-%Y.
example::
    `strftime(dateField, "%v")` результат: " 2-Feb-1993"
