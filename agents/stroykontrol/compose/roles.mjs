/**
 * Roles of the «Стройконтроль ПТО» namespace (ТЗ 3.8): creates four platform roles and gives each
 * the access it needs to the modules of the namespace. The agent knows the roles by their handles
 * (cmp/roles.go) and, run with --roles=on, checks the user behind every change it makes on their behalf.
 * Run after apply.mjs (it reads applied.json for the namespace):
 *
 *   node compose/roles.mjs            # create the roles and the rules
 *   DRY_RUN=1 node compose/roles.mjs  # only print what would be granted
 *
 * Members are added to the roles by the administrator (Администрирование → Роли).
 *
 *   Наблюдатель (stroykontrol_view)    — reads everything, changes nothing;
 *   Проектировщик (stroykontrol_des)   — reads everything, loads documentation (sets, volume list) and works on remarks;
 *   Специалист КК (stroykontrol_qc)    — as the designer, plus decisions on findings, comparisons and correction packages;
 *   Руководитель проекта (stroykontrol_mgr) — writes everything of the namespace (objects, participants, notification recipients).
 */
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { mintToken, detectBase, apiFactory, ensureRole, grantModuleAccess, setOf } from './helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const DRY = process.env.DRY_RUN === '1'

// module handles each role may write; every role reads all modules of the namespace
const DOCS = ['document_sets', 'object_volumes', 'pd_rd_comparisons', 'cmp_documents']
const REMARKS = ['expertise_remarks', 'expertise_remark_history', 'expertise_imports', 'notifications', 'tasks', 'task_history']
const REVIEW = ['pd_rd_discrepancies', 'remark_cards', 'correction_packages', 'id_check_runs', 'findings']

export const ROLES = [
  { handle: 'stroykontrol_view', name: 'Стройконтроль: наблюдатель', write: [] },
  { handle: 'stroykontrol_des', name: 'Стройконтроль: проектировщик', write: [...DOCS, ...REMARKS] },
  { handle: 'stroykontrol_qc', name: 'Стройконтроль: специалист по контролю качества', write: [...DOCS, ...REMARKS, ...REVIEW] },
  { handle: 'stroykontrol_mgr', name: 'Стройконтроль: руководитель проекта', write: '*' },
]

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const nsID = applied.namespaceID
  const token = await mintToken()
  const composeBase = await detectBase(token)
  const api = apiFactory(composeBase, token)
  const sys = apiFactory(composeBase.replace(/\/compose$/, '/system'), token)
  // the modules as they are now (applied.json may predate the last apply)
  const modules = setOf(await api('GET', `/namespace/${nsID}/module/?limit=500`))
    .filter(m => !m.deletedAt).map(m => ({ id: String(m.moduleID), handle: m.handle }))
  console.log('namespace', nsID, '— modules:', modules.length)

  for (const role of ROLES) {
    const writes = new Set(role.write === '*' ? modules.map(m => m.handle) : role.write)
    const missing = role.write === '*' ? [] : role.write.filter(h => !modules.some(m => m.handle === h))
    if (missing.length) console.log('  (нет в пространстве, пропущены:', missing.join(', ') + ')')
    if (DRY) {
      console.log(role.handle, 'read: all;', 'write:', [...writes].join(', ') || '—')
      continue
    }
    const roleID = await ensureRole(sys, role)
    for (const m of modules) {
      await grantModuleAccess(sys, roleID, nsID, m.id, { write: writes.has(m.handle) })
    }
    console.log(role.handle, roleID, '— rules for', modules.length, 'modules, writes', writes.size)
  }
  console.log('Добавьте пользователей в роли (Администрирование → Роли) и запустите агент с --roles=on.')
}

main().catch(e => { console.error(e); process.exit(1) })
