#!/usr/bin/env node
/**
 * Устарел: пространство «Проверка ИД (АОСР)» объединено с «Стройконтроль ПТО».
 * Модули ИД, страницы и rule chain теперь создаёт ../compose/apply.mjs
 * (определения модулей — ../compose/modules_id.mjs); данные перенесены без потерь
 * скриптом ../compose/migrate_id_into_main.sql.
 */
console.error('apply_id.mjs больше не нужен: запустите `node ../compose/apply.mjs` (пространство stroykontrol).')
process.exit(1)
