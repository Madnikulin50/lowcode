/**
 * Domain-specific help for Metric blocks. Describes the numbers on that block,
 * not Compose UX. Lookup: slug/pageHandle/title, then slug/title.
 */

export const HELP_MARK = '<!-- compose-help:v2 -->'

function words (text) {
  return String(text || '').replace(/[#*_`>[\]()]/g, ' ').split(/\s+/).filter(w => /[0-9A-Za-zА-Яа-яЁё]/.test(w)).length
}

function norm (s) {
  return String(s || '').trim().toLowerCase().replace(/\s+/g, ' ')
}

function expandTo (text, min, extras) {
  let out = String(text || '').trim()
  let i = 0
  while (words(out) < min && i < extras.length) {
    out += '\n\n' + extras[i]
    i++
  }
  return out.trim()
}

function moduleName (id, modulesByID) {
  if (!id || id === '0') return ''
  return modulesByID.get(String(id))?.name || ''
}

function opLabel (m) {
  const op = String(m.operation || 'count').toLowerCase()
  const field = m.metricField && m.metricField !== 'count' ? m.metricField : ''
  if (op === 'sum' && field) return `сумма поля «${field}»`
  if (op === 'avg' && field) return `среднее поля «${field}»`
  if (op === 'max' && field) return `максимум поля «${field}»`
  if (op === 'min' && field) return `минимум поля «${field}»`
  return 'количество записей'
}

const BLOCK_ARTICLES = {
  'invest/dashboard/сводка': `# Сводка портфеля

Четыре числа утреннего экрана инвестора и PMO. Это не «украшение дашборда», а очередь решений: где горит пакет, какие RFC ещё могут сдвинуть EAC, какие риски без плана.

**Проекты** — стройки в статусе active, не архив и не черновик инициации. Ноль при живом списке «Портфель» значит фильтр статуса на списке, не пустой контур. Рост без новых карточек не бывает: сначала заводят проект, потом клонируют WBS.

**На согласовании** — документы со статусом in_review. Большое число при пустом «Документы на мне» — очередь у других ролей (банк, технадзор, ГИП). Это узкое горло фазы: пока пакет не утверждён, оплату и закрытие этапа не двигайте.

**Открытые RFC** — запросы на изменение ещё на согласовании. Каждый может сдвинуть бюджет и срок. Не смотрите только штуки: один RFC на коридор трассы важнее десяти мелких. После утверждения должен появиться журнал и новый прогноз — иначе изменение прошло мимо процесса.

**Открытые риски** — реестр в статусе open. Сверяйте с графиком «Риски по баллу»: много открытых при пустом high-score — заниженная вероятность. Закрытие «чтобы цифра похорошела» без плана митигации ломает отчёт банку.

Пустой блок при живых списках ниже — предфильтр или права на модуль, не «всё спокойно».`,

  'invest/project/по проекту': `# Показатели этой стройки

Три числа только по открытому проекту, не по всему портфелю. Фильтр \`project = текущая запись\` уже зашит: цифра не совпадёт с дашбордом, если на дашборде нет того же отбора.

**Документы** — все пакеты этой стройки, любой статус. Ноль при заполненном чек-листе фазы значит документы не привязали к проекту или смотрите чужую карточку.

**Утверждены** — только status = approved. Доля к «Документы» — готовность пакета. Мало утверждённых при живом маршруте — смотрите «На согласовании» на странице документов, не правьте статус руками.

**RFC** — запросы этой стройки ещё in_review. Пока цифра не ноль, baseline (бюджет, срок, объём) ещё может сдвинуться. Перед утверждением симулируйте EAC на карточке RFC.

После факта с площадки или клонирования WBS обновите экран. Пересчёт EVM справа не меняет эти счётчики — они читают статусы документов и RFC, не SPI.`,

  'invest/documents/реестр': `# Реестр документов

Разрез модуля документов по статусу. Это воронка пакета, не договоры и не версии файлов.

**Всего** — все записи модуля без предфильтра блока. Если список ниже уже сужен проектом или «на мне», сумма колонок канбана может быть меньше.

**Черновики** — пакет не запущен. Долго висящие черновики оставляют чек-лист фазы красным. Не путайте с черновиком версии внутри уже отправленного документа.

**На согласовании** — in_review, живой маршрут. Сверяйте со списком «на мне» и канбаном рядом: цифра — штуки, канбан — кто держит шаг.

**Утверждены** — approved. Рост без падения «на согласовании» значит завели уже утверждённые задним числом или есть другой статус вне этих четырёх (отклонён не показан).

Отклонённые в этот блок не входят — их смотрите в таблице и на графике «Документы по статусу».`,

  'invest/budget/финансы': `# Финансы портфеля

Четыре числа по статьям бюджета (budget_lines), не по договорам и не по кассовым платежам.

**Статей** — сколько строк сметы заведено. Ноль — бюджет не разложили по WBS/статьям, EAC на карточке проекта будет пустым или ручным.

**План** — сумма поля planned. Это исходная смета, не «сколько осталось». Правка плана после утверждённого RFC должна идти через журнал, иначе план и факт разъедутся с SPI/CPI.

**Факт** — сумма actual. Растёт от фактов прогресса и закрытых актов, не от желания «подтянуть» отчёт. Факт выше плана при живом резерве — сначала статья, потом RFC.

**Резерв** — сумма reserve. Обнуление при открытых рисках high-score — дыра в отчёте банку. Не распределяйте резерв «на глаз» без статьи.

Сверяйте с графиками «Приход / расход» и списком платежей: касса может отставать от освоенного объёма. Валюта — рубли, как в префиксе.`,

  'cmdb/dashboard/at a glance': `# At a glance

Гигиена парка за один взгляд. Цифры читают модули Devices и Vulnerabilities, не «настроение дашборда».

**Devices** — все узлы инвентаря, любой статус. Скачок после обхода LAN — норма, если агент впервые увидел сегмент. Резкий рост unknown смотрите на графике типов.

**Online** — status = online. Должен подтверждаться last_seen. Вечный online при старом last_seen — кандидат на «Find stale online devices».

**Offline** — выключены или не отвечают. Рост offline без скана — кто-то правит статус руками.

**Open findings** — открытые находки сканера. Пороги на блоке красят кнопку: уже одна находка — warning, от пяти — danger. Не удаляйте CVE, чтобы цифра упала.

**Critical / High** — открытые CRITICAL и HIGH. Это очередь ИБ на сегодня, не весь реестр. Ноль при красном графике severity — фильтр статуса (acknowledge/fixed не входят).

Сверяйте со списками «Open vulnerabilities» и «Recently seen devices» ниже.`,

  'cmdb/devices/inventory': `# Inventory

Срез живого инвентаря по статусу узла. Тип устройства (server, phone, tablet) на этом блоке не режется — его смотрите в таблице и на графике «Devices by type».

**Devices** — все записи модуля. Дубли по IP/MAC раздуют счётчик и размажут CVE.

**Online / Offline** — текущий статус. Stale online сюда как online, пока цепочка hygiene его не сняла.

**Unknown** — статус unknown, не тип unknown. Часто после сырого скана, пока оператор не классифицировал узел.

После обхода сети обновите страницу. Ручной узел без скана появится здесь же — не плодите копии «на всякий случай».`,

  'cmdb/device/this host': `# This host

Только этот узел: открытые находки, службы, тяжёлые CVE. Фильтр \`device = эта запись\` уже стоит.

**Open vulns** — открытые находки на хосте. Ноль при заполненном списке Vulnerabilities справа — все закрыты или false positive, либо список без фильтра статуса.

**Services** — строки модуля Services (порт/баннер). Не путайте с полем open_ports на карточке: оно сырое, счётчик — нормализованные службы.

**High / Critical** — открытые HIGH и CRITICAL этого узла. Если цифра > 0, кнопка «Flag insecure ports» и письмо из цепочки имеют смысл. Acknowledge снимает «несмотрено» с дашборда, но не уменьшает это число, пока статус open.

После патча дождитесь следующего скана: находка может вернуться с тем же CVE.`,

  'cmdb/network/history': `# History

**Scans** — сколько прогонов привязано к этой сети (network = карточка). Это не «все сканы компании» и не устройства в CIDR.

Ноль при живой кнопке «Scan this network» — обход ещё не стартовал или запись скана без ссылки на сеть. Долгий рост running без finished смотрите на карточке скана и агенте, не запускайте третий обход того же CIDR.

Список ниже — те же записи. Счётчик и таблица должны совпасть, если нет другого предфильтра.`,

  'cmdb/scans/activity': `# Activity

Очередь обходов, не инвентарь.

**Scans** — все записи журнала сканов. Растёт с каждым запуском, включая failed.

**Running** — сейчас в полёте. Полоска «In flight» рядом должна быть согласована с этой цифрой. Часами running — агент завис или webhook не пришёл.

**Failed** — обход не дописал результат. Смотрите error на карточке, не перезапускайте пачками один и тот же CIDR.

Кнопка Start scan создаёт новую запись, не «чинит» failed.`,

  'cmdb/scan/found': `# Found

**Hosts** — поле found этой записи скана (максимум по фильтру recordID). Это сколько узлов агент положил в результат прогона, не текущий инвентарь всей CMDB.

Ноль при статусе success — пустой сегмент или фильтр сканера. Ноль при running — прогон ещё не досчитал. Число может быть меньше, чем новых Devices: дедуп по IP/MAC не создаёт вторую карточку.

Сверяйте со списками устройств «recently seen» на дашборде после завершения.`,

  'faris/dashboard/at a glance': `# At a glance

Group procurement pulse for the eight subsidiaries. Counts and averages come from vendor and purchase-request modules, not SAP actuals.

**Active onboarding** — vendor packs that are no longer draft and not yet approved or rejected. This is the live CR/VAT queue. A jump without new cards usually means a status was moved back into review.

**Open purchase requests** — PRs not approved and not rejected. Includes holds. Compare with the stalled list at the bottom of the page.

**Avg cycle (days)** — average cycle_days on approved PRs only. It ignores open tickets, so a short average with a long open queue means old approvals, not a healthy pipeline.

**Stalled / overdue** — stalled = true. These are the tickets group finance should act on this week. Clearing the flag without a status change just hides the number.

Empty metrics with rows in the overdue list: check field rights and that stalled is stored as boolean, not text.`,

  'faris/vendors/pipeline': `# Pipeline

Vendor onboarding funnel. Statuses are process steps, not ERP vendor-master states.

**Incomplete packs** — status incomplete: CR, VAT or attachments missing. Do not send these to compliance.

**In review** — procurement_review, compliance_review or finance_verify. One legal name should not sit in two of these; if the count is high, look at the kanban columns.

**Approved** — ready to be treated as an approved vendor in the demo. Rejected packs are not in this block. Approving here must go through the card actions so the approval log stays complete.

Subsidiary codes must stay stable: a mistyped code splits the same vendor across charts.`,

  'faris/purchase-requests/queue': `# Queue

Finance-facing slice of purchase requests, not the full PR mix (that is the chart on the left).

**With finance** — status finance_approval. This is the group-finance inbox.

**On hold** — explicitly parked. Holds should have a reason on the card; a large number with no comments is process theatre.

**Approved value** — sum of estimated_value on approved PRs, shown in SAR. It is estimated demo spend, not posted journal entries. Compare with «Mock spend by category» on the group dashboard.

Do not approve from this block: open the card so budget checks and the approval log run.`,

  'loop/dashboard/показатели': `# Показатели сводки сети

Плитка на общем дашборде розницы. Числа из чеков, трафика или стока — смотрите подпись каждой цифры, не заголовок блока.

Сверяйте с «Основная статистика», MoM и YoY на этой же странице: дубли одной кассы в двух плитках должны совпасть. Ноль при живой карте магазинов — фильтр периода или дыра загрузки, не закрытая сеть.

Служебные и тестовые блоки с тем же именем не используйте для заказа и премий. Решения — на экранах магазинов, остатков и маржи.`,

  'loop/dashboard/основная статистика': `# Основная статистика сети

Крупные цифры сводки розницы: оборот, трафик, точки. Это касса и поток, не «оценка управляющего».

Сверяйте MoM/YoY блоки рядом: скачок без чеков — дыра загрузки или смена границы региона. Ноль при живой карте магазинов — фильтр периода или пустой модуль чеков, не закрытая сеть.

Не принимайте заказ поставщику и премию смены только по этой плитке: откройте «Выручка месяц-к-месяцу» и список магазинов. Служебные блоки с именем «показатели» ниже не используйте для решений.`,

  'loop/dashboard/выручка месяц-к-месяцу': `# Выручка месяц-к-месяцу

Сравнение текущего месяца с предыдущим по чекам (обычно receipt / sales_amount). Плюс — рост кассы, минус — провал среднего чека, трафика или закрытие якоря.

Дыра в начале месяца — календарь, не катастрофа. Аномалия одного дня смотрите на графике «по датам», не увольняйте по одной плитке. Сверяйте с YoY: MoM падает в сезонности, YoY может быть в норме.

Цифры из чеков, их нельзя поправить на дашборде. Нулевой месяц при живых инцидентах кассы — сначала ККТ и загрузка.`,

  'loop/dashboard/выручка год-к-году': `# Выручка год-к-году

Тот же контур кассы, что MoM, но к тому же месяцу прошлого года. Нужен, чтобы не путать сезон с трендом: декабрь к ноябрю всегда «растёт».

Резкий провал YoY при ровном MoM — закрытие точки, смена ассортимента или дыра в прошлогодней загрузке. Сверяйте состав сети (новые/закрытые магазины) на странице магазинов: открытие десяти точек раздует YoY без роста сопоставимых.

Не сравнивайте с маржой на соседнем экране один в один: выручка валовая, маржа — после себестоимости.`,

  'loop/region/показатели': `# Показатели региона

Сводка выбранного региона, не всей сети. Предфильтр региона уже в блоке: цифра не совпадёт со сводкой сети.

Смотрите вместе с картой точек и типами магазинов. Белое пятно на карте при ненулевой выручке — нет координат, не «нет продаж». Mix форматов (у дома / ТЦ) ломает сравнение, если тип записан неверно.

Ноль при живом списке магазинов региона — фильтр периода чеков или права на модуль продаж.`,

  'loop/store/показатели': `# Показатели магазина

Цифры этой точки: касса, трафик, эффективность. Не сеть и не регион.

Провал выручки при живом трафике — средний чек, цена, касса. Провал чеков при живом потоке — конверсия на входе. Нулевой день чаще дыра загрузки ККТ.

Сверяйте блоки «Выручка: периоды» и «Трафик: периоды» на этой же карточке. Инциденты кассы/холода рядом объясняют провал лучше, чем «настроение» директора.`,

  'loop/store/выручка: периоды': `# Выручка: периоды

Касса этой точки в разрезе периодов (день/неделя/месяц — как настроено). Нужна, чтобы отделить акцию и выходной от системного падения.

Дыры в ряду — загрузка или закрытие смены, не обязательно ноль продаж. Скачок в день промо сверяйте с акцией и чеками, не с премией. YoY/MoM на сводке сети сюда не копируйте: здесь одна точка.`,

  'loop/store/трафик: периоды': `# Трафик: периоды

Входной поток точки (счётчики, не чеки). Падение трафика при живой выручке — вырос средний чек или сменился mix. Падение обоих — якорь, погода, ремонт входа, конкурент.

Нули в часы работы — датчик или загрузка, не «никто не зашёл». Не считайте конверсию с этой плитки, пока ряд трафика и ряд чеков не за тот же период.`,

  'loop/store/эффективность': `# Эффективность точки

Отношение кассы к трафику, площади или смене — как задано в метриках блока. Это не «оценка персонала» сама по себе.

Рост эффективности при падении трафика — меньше людей, те же чеки. Падение при живом потоке — очередь, касса, дырки в полке. Сверяйте с инцидентами и DOC на остатках: пустая полка бьёт конверсию раньше, чем выручку за месяц.`,

  'loop/product/показатели': `# Показатели SKU

Цифры выбранного товара: продажи, остаток, маржа — не всего каталога. Предфильтр product уже стоит.

Ноль продаж при живом остатке — нет выкладки, цена, дыра в чеках. Ноль остатка при продажах вчера — автозаказ не успел или rec на карточке другой.

Сверяйте со страницей «Остатки и автозаказ» и маржинальностью: минусовая маржа при «хороших» продажах — цена/себестоимость, не трафик.`,

  'loop/margin/валовая маржа': `# Валовая маржа

Сумма или доля gross profit / margin_pct по чекам и себестоимости. Не выручка и не «наценка на ценнике».

Падение маржи при росте выручки — промо, списание, ошибка cost. Рост маржи при падении штук — уход дешёвого микса. Нулевая маржа при живых чеках — не подтянута себестоимость.

Не заказывайте объём по этой плитке: сначала полка и DOC, потом маржа категории. Служебные блоки с опечаткой в имени игнорируйте.`,

  'loop/stock-reorder/остатки и автозаказ': `# Остатки и автозаказ

Сводка стока: критичные остатки, дни покрытия (DOC), к заказу. Читает stock_reorder_fact / inventories, не чеки.

Рост «критичных» при живой выручке — полка кончается быстрее заказа. DOC = 0 при ненулевом stock — формула покрытия, не пустой склад. Автозаказ не запускается с этой плитки: смотрите таблицу reorder_qty и поставщика.

Белая плитка при заполненном списке ниже — другой фильтр health_level. Не закрывайте инцидент «нет товара» только потому, что сводка зелёная: сводка по сети, карточка — по точке.`,

  'loop/risks/сводка рисков': `# Сводка рисков сети

Открытые риски полки, кассы, холода, охраны — штуки или балл, как настроено. Не инциденты за сегодня и не CVE.

Рост при пустом канбане — фильтр статуса на канбане. Падение без выезда ревизора — закрыли «чтобы цифра похорошела». Сверяйте с инцидентами: повтор на той же точке — системная причина.

Высокий балл без владельца на карточке риска не закрывайте сводкой. Новый риск заводите от инцидента, не наоборот.`,

  'loop/turnover/сводка оборачиваемости': `# Сводка оборачиваемости

Сток / продажи: сколько дней или оборотов «лежит» товар. Читает inventories / turnover, не кассу за один день.

Высокая оборачиваемость при дырках на полке — мало страхового запаса, автозаказ опаздывает. Низкая при живой выручке — затоваривание, не «хорошие продажи». Ноль — нет себестоимости или остатка в срезе, не «мгновенный уход».

Сверяйте с «Остатки и автозаказ» и маржой: медленный сток с минусовой маржой — кандидат на вывод, не на дозаказ. Не режьте заказ по одной сети: смотрите точку и SKU.

`,

  'loop/hr-personnel/сводка по персоналу': `# Сводка по персоналу

Штат, смены, явки или вакансии — по кадровым срезам, не по чекам. Ноль при живых магазинах — не подтянут staff_actual / retail_staff.

Перекос «мало людей» при высокой выручке — переработки и риск кассы, не повод резать заказ. Перекос «много» при мёртвом трафике — график смен, не премия.

Не увольняйте по этой плитке: откройте список персонала и отчёт директора точки. Тестовые блоки на сводке сети с именем «показатели» сюда не относятся.`,

  'backup/dashboard/сводка': `# Сводка резервного копирования

Четыре числа контура MinIO: что включено, что бежит, что упало, что можно вернуть.

**Источники** — enabled источники (SMB, каталоги, PostgreSQL, MySQL, внешний S3). Выключенный источник не попадает в due, даже если политика жива.

**Идут сейчас** — джобы status = running. Долгий рост без progress на карточке — агент офлайн или очередь стоит. Не жмите «Запустить due» пачками.

**Ошибки** — failed джобы. Список «Последние ошибки» ниже должен быть того же порядка. Секрет в ошибке не светите в комментариях.

**Снапшоты** — restorable = 1, то что реально можно накатить. Доля success на графике джобов не равна этому числу: успешный прогон с истёкшим retention уже не restorable.

Due и prune справа меняют следующие два счётчика после обновления страницы.`,
}

const ITEM_SNIPPETS = {
  'invest/проекты': 'Стройки в работе (обычно status = active). Не архив и не пустой черновик без WBS.',
  'invest/на согласовании': 'Документы in_review — пакет держит фазу и оплату, пока маршрут не закрыт.',
  'invest/открытые rfc': 'RFC ещё на согласовании: каждый может сдвинуть бюджет, срок и EAC.',
  'invest/открытые риски': 'Риски в статусе open. Сверяйте с баллом вероятность × влияние, не только со штуками.',
  'invest/документы': 'Пакеты выбранного проекта, все статусы. Ноль — нет привязки project или чужая карточка.',
  'invest/утверждены': 'Только approved. Доля к «всего» — готовность пакета, не статус договора.',
  'invest/rfc': 'Открытые RFC этой стройки. Пока не ноль, baseline ещё может измениться.',
  'invest/всего': 'Все записи модуля без отбора блока. Канбан и «на мне» могут показать меньше.',
  'invest/черновики': 'Пакет не отправлен. Долгий черновик оставляет чек-лист фазы красным.',
  'invest/статей': 'Сколько строк сметы заведено. Без статей EAC и план-факт пустые.',
  'invest/план': 'Сумма planned — исходная смета, не остаток.',
  'invest/факт': 'Сумма actual — освоение, не касса и не «желание отчёта».',
  'invest/резерв': 'Сумма reserve. Обнуление при открытых high-рисках — дыра для банка.',
  'cmdb/devices': 'Узлы инвентаря. Дубли по IP/MAC раздувают счётчик и размазывают CVE.',
  'cmdb/online': 'status = online. Проверяйте last_seen: вечный online при старой дате — stale.',
  'cmdb/offline': 'Не отвечают или выключены. Скачок без обхода — ручная правка статуса.',
  'cmdb/open findings': 'Открытые находки сканера. Не удаляйте CVE, чтобы цифра упала.',
  'cmdb/critical / high': 'Открытые CRITICAL и HIGH — очередь ИБ на сегодня.',
  'cmdb/unknown': 'Статус unknown (не тип unknown). Часто сырой скан до классификации.',
  'cmdb/open vulns': 'Открытые находки этого узла. Acknowledge не снимает open.',
  'cmdb/services': 'Нормализованные службы (порт/баннер), не сырое поле open_ports.',
  'cmdb/high / critical': 'Тяжёлые открытые CVE этого хоста.',
  'cmdb/scans': 'Записи журнала обходов. Каждый запуск — новая строка, включая failed.',
  'cmdb/running': 'Скан в полёте. Часами running — агент или webhook, не «ещё один Start».',
  'cmdb/failed': 'Обход не дописал результат. Смотрите error на карточке.',
  'cmdb/hosts': 'Поле found этого прогона: сколько узлов агент вернул, не весь парк.',
  'faris/active onboarding': 'Vendor packs in flight: not draft, not approved, not rejected.',
  'faris/open purchase requests': 'PRs still open, including holds. Compare with the stalled list.',
  'faris/avg cycle (days)': 'Average cycle_days on approved PRs only — ignores the live queue.',
  'faris/stalled / overdue': 'stalled = true. Act on the card; clearing the flag is not a decision.',
  'faris/incomplete packs': 'Missing CR/VAT/attachments. Do not send to compliance yet.',
  'faris/in review': 'Procurement, compliance or finance verify — one pack, one column.',
  'faris/approved': 'Approved vendors in the demo. Use card actions so the log stays complete.',
  'faris/with finance': 'PRs in finance_approval — group-finance inbox.',
  'faris/on hold': 'Explicitly parked tickets. Expect a reason on the card.',
  'faris/approved value': 'Sum of estimated_value on approved PRs (SAR estimates, not SAP postings).',
  'backup/источники': 'Включённые источники. Выключенный не попадает в due.',
  'backup/идут сейчас': 'Джобы running. Долгий рост — агент или очередь, не ещё один due.',
  'backup/ошибки': 'failed джобы. Секреты из error не копируйте в комментарии.',
  'backup/снапшоты': 'restorable снимки. Success джоба ≠ можно восстановить после retention.',
}

const DOMAIN_EXTRA = {
  invest: 'Цифры читают те же модули, что списки и графики рядом. После факта с площадки, RFC или платежа обновите экран; SPI/EAC считаются отдельно кнопкой EVM и в этот блок не входят.',
  cmdb: 'Счётчики обновляются при открытии страницы и после скана. Не правьте статус «чтобы цифра похорошела»: нужен аудит last_seen и CVE.',
  faris: 'These are demo ticket counts and SAR estimates, not live SAP. Status changes belong on the record card so the approval log stays complete.',
  backup: 'Due и prune на дашборде меняют running/failed/restorable после обновления. Офлайн-агент оставляет running врать.',
  loop: 'Цифры сети из чеков, стока, трафика и инцидентов. Сверяйте соседний список и график. Ноль чаще фильтр или дыра загрузки, не «в сети всё спокойно».',
}

const PAGE_ALIAS = {
  сводка: 'dashboard',
  дашборд: 'dashboard',
  'selected region': 'region',
  'selected store': 'store',
  'selected product': 'product',
  маржинальность: 'margin',
  'остатки и автозаказ': 'stock-reorder',
  риски: 'risks',
  персонал: 'hr-personnel',
  оборачиваемость: 'turnover',
  проекты: 'projects',
  документы: 'documents',
  бюджет: 'budget',
  проект: 'project',
  devices: 'devices',
  device: 'device',
  networks: 'network',
  network: 'network',
  scans: 'scans',
  scan: 'scan',
  'vendor onboarding': 'vendors',
  'purchase requests': 'purchase-requests',
  'group dashboard': 'dashboard',
}

export function lookupMetricBlockArticle ({ slug, pageHandle, pageTitle, blockTitle }) {
  const alias = PAGE_ALIAS[norm(pageHandle)] || PAGE_ALIAS[norm(pageTitle)] || ''
  const keys = [
    `${slug}/${pageHandle}/${blockTitle}`,
    `${slug}/${pageTitle}/${blockTitle}`,
    alias && `${slug}/${alias}/${blockTitle}`,
    `${slug}/${blockTitle}`,
  ].filter(Boolean).map(k => norm(k))
  for (const k of keys) {
    if (BLOCK_ARTICLES[k]) return BLOCK_ARTICLES[k]
  }
  return null
}

function itemSnippet (slug, metric, modulesByID) {
  const label = metric.label || 'Показатель'
  const snippet = ITEM_SNIPPETS[`${slug}/${norm(label)}`]
  const mod = moduleName(metric.moduleID, modulesByID)
  const how = opLabel(metric)
  const filter = String(metric.filter || '').trim()
  const filterHint = filter
    ? (filter.length > 80 ? `Отбор задан в блоке.` : `Отбор: \`${filter}\`.`)
    : 'Без отбора блока — все записи модуля, на которые хватает прав.'
  const head = snippet || `${how}${mod ? ` модуля «${mod}»` : ''}.`
  return `${head} ${mod && snippet ? `Источник — «${mod}», ${how}.` : ''} ${filterHint}`.replace(/\s+/g, ' ').trim()
}

export function generateMetricBlockHelp (block, page, ns, modulesByID) {
  const slug = ns.slug || ''
  const title = String(block?.title || 'Показатели').trim()
  const pageHandle = page.handle || ''
  const pageTitle = page.title || pageHandle
  const metrics = Array.isArray(block?.options?.metrics) ? block.options.metrics : []
  const article = lookupMetricBlockArticle({
    slug,
    pageHandle,
    pageTitle,
    blockTitle: title,
  })

  const labels = metrics.map(m => m.label).filter(Boolean)
  const description = article
    ? `${title}: ${labels.slice(0, 4).join(', ') || 'показатели экрана'}.`
    : `${title}${labels.length ? ` — ${labels.slice(0, 4).join(', ')}` : ''}.`

  if (article) {
    const help = String(article).includes(HELP_MARK) ? article : `${HELP_MARK}\n${article}`
    return {
      description: description.replace(/\s+/g, ' ').trim(),
      help: expandTo(help, 80, [DOMAIN_EXTRA[slug] || DOMAIN_EXTRA.loop]),
    }
  }

  const items = metrics.length
    ? metrics.map(m => `**${m.label || 'Показатель'}** — ${itemSnippet(slug, m, modulesByID)}`).join('\n\n')
    : 'В блоке нет настроенных показателей: в конфигураторе добавьте метрики (модуль, операция, фильтр).'

  const help = expandTo(`${HELP_MARK}
# ${title}

Показатели экрана «${pageTitle}» пространства «${ns.name || slug}». Каждое число — свой запрос к модулю, не «настроение» страницы.

${items}

Пустая цифра чаще значит фильтр, период или нет записей под отбор, а не ноль в предметной области. Расхождение со списком рядом — разный предфильтр или права на поле. Нажимайте на число, только если включён drill-down.
`, 80, [DOMAIN_EXTRA[slug] || DOMAIN_EXTRA.loop])

  return { description: description.replace(/\s+/g, ' ').trim(), help }
}
