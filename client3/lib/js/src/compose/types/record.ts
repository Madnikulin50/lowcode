/* eslint-disable @typescript-eslint/ban-ts-comment */
import { AreObjectsOf, IsOf } from '../../guards'
import { Apply, CortezaID, ISO8601Date, NoID } from '../../cast'
import { Module } from './module'
import { Namespace } from './namespace'

const fieldIndex = Symbol('fieldIndex')
const propModule = Symbol('module')
const cleanValues = Symbol('cleanValues')

const reservedFieldNames = [
  'toJSON',
]

interface FieldIndex {
  isMulti: boolean;
  kind: string;
  defaultValue: Array<{ value: string }>;
}

interface RawValue {
  name: string;
  value?: string;
}

interface PartialRecord extends Partial<Omit<Record, 'values' | 'createdAt' | 'updatedAt' | 'deletedAt'>> {
  values?: RawValue[];

  createdAt?: string|number|Date;
  updatedAt?: string|number|Date;
  deletedAt?: string|number|Date;
}

export interface Values {
  [name: string]: string|string[]|undefined;
}

/**
 * Combination of valid value types/structures
 */
type ValueCombo = RawValue[] | Values | Values[]

/**
 * Combination of valid types that  can be passed as Record ctor's 1st (and 2nd) parameter
 */
type RecordCtorCombo = Record | Module | PartialRecord | ValueCombo

/**
 * For something to be useful module (for a Record), it needs to contain fields.
 * Use property access rather than hasOwnProperty: Vue 3 reactive proxies
 * hide class fields from hasOwnProperty but still expose .fields.
 */
function isModule (m?: unknown): m is Module {
  if (!m || typeof m !== 'object') {
    return false
  }
  const fields = (m as Module).fields
  if (!Array.isArray(fields) || fields.length === 0) {
    return false
  }
  // A Record (or a spread of one) can leak `fields` through a Vue proxy.
  // Modules have name/handle; records have recordID without those.
  const rec = m as Partial<Record>
  if (rec.recordID && rec.recordID !== NoID && (m as Module).name === undefined && (m as Module).handle === undefined) {
    return false
  }
  return true
}

/**
 * Vue 3 reactive() wraps class instances in a proxy that is not frozen
 * even when the target is. Always copy from the raw target when present.
 */
function unwrapVue<T> (v: T): T {
  let cur: unknown = v
  const seen = new Set<unknown>()
  while (cur && typeof cur === 'object' && !seen.has(cur)) {
    seen.add(cur)
    const raw = (cur as { __v_raw?: unknown }).__v_raw
    if (!raw || raw === cur) {
      break
    }
    cur = raw
  }
  return cur as T
}

function isRawValue (v: unknown): v is RawValue {
  return IsOf<RawValue>(v, 'name')
}

/**
 * True when the value looks like a real Corteza ID (not empty / NoID / 0).
 */
function hasResourceID (id: unknown): boolean {
  if (id == null || id === '' || id === 0) {
    return false
  }
  try {
    return CortezaID(id) !== NoID
  } catch {
    return true
  }
}

/**
 * Compare Corteza IDs without treating string vs number as a module change.
 */
function sameResourceID (a: unknown, b: unknown): boolean {
  if (a === b) {
    return true
  }
  if (a == null || b == null) {
    return false
  }
  if (String(a) === String(b)) {
    return true
  }
  try {
    return CortezaID(a) === CortezaID(b)
  } catch {
    return false
  }
}

/**
 * Read a Corteza ID from a payload without hasOwnProperty.
 * Vue 3 proxies and some class copies hide fields from hasOwnProperty,
 * so Apply() would leave recordID at NoID and the list would use a
 * placeholder like "moduleID-0-0-0:0" in bulk delete.
 */
function resourceIDFrom (src: unknown, ...keys: string[]): string | undefined {
  if (!src || typeof src !== 'object') {
    return undefined
  }
  const o = src as { [key: string]: unknown }
  for (const key of keys) {
    let v: unknown
    try {
      v = o[key]
    } catch {
      continue
    }
    if (!hasResourceID(v)) {
      continue
    }
    try {
      const id = CortezaID(v)
      if (id !== NoID) {
        return id
      }
    } catch {
      continue
    }
  }
  return undefined
}

/**
 * Record class will be used all over the place, user scripts, etc..
 *
 * Constructor (and apply fn) is as versatile as possible to handle
 * different use-cases.
 */
export class Record {
  public recordID = NoID;
  public moduleID = NoID;
  public namespaceID = NoID;
  public revision = 0;

  public values: Values = {}
  public valueErrors: object = {}
  public meta: object = {};

  public createdAt?: Date = undefined;
  public updatedAt?: Date = undefined;
  public deletedAt?: Date = undefined;

  public ownedBy = undefined;
  public createdBy = undefined;
  public updatedBy = undefined;
  public deletedBy = undefined;

  public canUpdateRecord = false;
  public canReadRecord = false;
  public canDeleteRecord = false;
  public canUndeleteRecord = false;
  public canManageOwnerOnRecord = false;
  public canSearchRevision = false;
  public canGrant = false;

  // @ts-ignore
  private [fieldIndex]: Map<string, FieldIndex>
  private [propModule]?: Module
  private [cleanValues]: Values = {}

  constructor (recModVal1: RecordCtorCombo, recModVal2?: RecordCtorCombo) {
    if (recModVal1 instanceof Record) {
      this.module = recModVal1.module
      this.apply(recModVal1)
      return
    }

    if (isModule(recModVal1)) {
      this.module = recModVal1
      this.apply(recModVal2)
      return
    }

    if (isModule(recModVal2)) {
      this.module = recModVal2
      this.apply(recModVal1)
      return
    }

    throw new Error('invalid module used to initialize a record')
  }

  clone (): Record {
    return new Record(this.module, JSON.parse(JSON.stringify(this)))
  }

  /**
   * apply (partially) updates record and it's values
   *
   * @param p
   */
  apply (p?: unknown): void {
    if (p === undefined) {
      // This is a brand new record; set default values
      this.defaultValues()
      return
    }

    let r

    // Determine what kind of value we got.
    // Prefer `in` / property access: IsOf() uses hasOwnProperty, which Vue 3
    // proxies often fail, so a real record was treated as a values blob and
    // recordID stayed NoID.
    const recLike = Boolean(
      p &&
      typeof p === 'object' &&
      !Array.isArray(p) &&
      ('recordID' in (p as object) || 'values' in (p as object))
    )

    switch (true) {
      case recLike || (!Array.isArray(p) && (IsOf<Record>(p, 'recordID') || IsOf<Record>(p, 'values'))):
        // p1 is something that looks like a record object
        r = p as Record
        break

      case AreObjectsOf<RawValue>(p, 'name'):
        // assuming p1 is array of raw values
        r = ({ values: p as RawValue[] }) as PartialRecord
        break

      default:
        r = ({ values: p }) as Record
    }

    r = r as PartialRecord

    // Compare against this.moduleID (set via property access on the module),
    // not this.module.moduleID: `new Module(vueProxy)` can leave the copy at NoID
    // because Apply uses hasOwnProperty, which Vue 3 proxies often fail.
    if (hasResourceID(this.moduleID) && hasResourceID(r.moduleID) && !sameResourceID(r.moduleID, this.moduleID)) {
      throw new Error(`can not change module on a record (${String(this.moduleID)} [${typeof this.moduleID}] → ${String(r.moduleID)} [${typeof r.moduleID}])`)
    }

    if (hasResourceID(this.namespaceID) && hasResourceID(r.namespaceID) && !sameResourceID(r.namespaceID, this.namespaceID)) {
      throw new Error(`can not change namespace on a record (${String(this.namespaceID)} [${typeof this.namespaceID}] → ${String(r.namespaceID)} [${typeof r.namespaceID}])`)
    }

    const recID = resourceIDFrom(r, 'recordID', 'ID')
    if (recID) {
      this.recordID = recID
    } else {
      Apply(this, r, CortezaID, 'recordID')
    }
    Apply(this, r, CortezaID, 'moduleID', 'namespaceID')
    Apply(this, r, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt')
    Apply(this, r, CortezaID, 'ownedBy', 'createdBy', 'updatedBy', 'deletedBy')

    Apply(this, r, Number,
      'revision',
    )

    Apply(this, r, Boolean,
      'canUpdateRecord',
      'canReadRecord',
      'canDeleteRecord',
      'canUndeleteRecord',
      'canManageOwnerOnRecord',
      'canGrant',
    )

    // This is a brand-new record; set default values
    if (!r.recordID || r.recordID === NoID) {
      this.defaultValues()
    }

    if (r.values !== undefined) {
      this.updateValues(r.values)
    }

    if (!this[cleanValues]) {
      // When there are no clean values,
      // make copy of values so that we know if change occurred
      this[cleanValues] = Object.freeze({ ...this.values })
    }

    if (r.valueErrors) {
      this.valueErrors = r.valueErrors
    }

    if (IsOf(r, 'meta')) {
      this.meta = { ...r.meta }
    }
  }

  public get cleanValues (): Values {
    return this[cleanValues]
  }

  public get module (): Module {
    if (this[propModule] === undefined) {
      throw new Error('module not set')
    }

    return this[propModule] as Module
  }

  public set module (m: Module) {
    m = unwrapVue(m)

    if (this[propModule]) {
      if (!sameResourceID((this[propModule] as Module).moduleID, m.moduleID)) {
        throw new Error('module for this record already set')
      }
    }

    if (!m.fields || !Array.isArray(m.fields) || m.fields.length === 0) {
      throw new Error('module used to initialize a record does not contain any fields')
    }

    // Prefer CortezaID so snowflake string vs number does not look like a module change later
    try {
      this.moduleID = CortezaID(m.moduleID)
    } catch {
      this.moduleID = String(m.moduleID ?? NoID)
    }
    try {
      this.namespaceID = CortezaID(m.namespaceID)
    } catch {
      this.namespaceID = String(m.namespaceID ?? NoID)
    }

    this[fieldIndex] = new Map()

    if (Object.isFrozen(m)) {
      this[propModule] = m
    } else {
      // Making a copy and freezing it
      this[propModule] = Object.freeze(new Module(m))
    }

    (this[propModule] as Module).fields.forEach(f => {
      const {
        name,
        isMulti,
        kind,
        defaultValue,
      } = f

      if (reservedFieldNames.includes(name)) {
        throw new Error('can not use reserved field name ' + name)
      }

      this[fieldIndex].set(name, { isMulti, kind, defaultValue })
    })

    Object.freeze(this[fieldIndex])

    this.initValues()
  }

  public get namespace (): Namespace {
    return this.module.namespace
  }

  /**
   * Converts internal representation of values into array of RawValue objects
   */
  serializeValues (): RawValue[] {
    const vv: RawValue[] = []

    this[fieldIndex].forEach(({ isMulti }, name) => {
      if (this.values[name] === undefined) {
        return
      }

      const val = this.values[name] as string|string[]

      if (isMulti) {
        if (Array.isArray(this.values[name])) {
          for (let i = 0; i < val.length; i++) {
            if (val[i] !== undefined) {
              vv.push({ name, value: val[i].toString() })
            }
          }
        }
      } else {
        vv.push({ name, value: val.toString() })
      }
    })

    return vv
  }

  /**
   * Removes existing, resets default values and updates it with new ones
   */
  public setValues (...i: ValueCombo[]): void {
    this.initValues()
    this.defaultValues()
    this.updateValues(...i)
  }

  /**
   * Removes existing and resets default values
   */
  protected initValues (): void {
    const dst: Values = {}

    this[fieldIndex].forEach(({ isMulti }, name) => {
      if (isMulti) {
        dst[name] = []
      } else {
        dst[name] = undefined
      }
    })

    // TypeScript complains about incompatibility between
    // indexed object and toJSON function
    // @ts-ignore
    dst.toJSON = (): RawValue[] => this.serializeValues()

    this.values = dst
  }

  protected defaultValues (): void {
    this[fieldIndex].forEach(({ isMulti, defaultValue }, name) => {
      if (defaultValue && Array.isArray(defaultValue) && defaultValue.length > 0) {
        if (isMulti) {
          this.values[name] = defaultValue.map(({ value }) => value)
        } else {
          this.values[name] = defaultValue[0].value
        }
      }
    })
  }

  /**
   * Updates record's values object with provided input
   *
   * Accepted values:
   * 1. Array of RawValue objects:
   *    updateValues([{ name: ..., value: ...}, ...])
   *
   * 2. One or more Value object:
   *    updateValues({ foo: ..., bar: ... }, ...)
   */
  protected updateValues (...combo: ValueCombo[]): void {
    // If all values are formatted as raw value
    if (combo.length === 1 && AreObjectsOf<RawValue>(combo[0], 'name')) {
      (combo[0] as Array<RawValue>).forEach(({ name, value }) => this.setValue(name, value))
      return
    }

    (combo as Array<Values>).forEach(v => {
      if (Array.isArray(v)) {
        this.updateValues(...v)
        return
      }

      if (!v || typeof v !== 'object') {
        throw Error('expecting array of values or values object')
      }

      // Handle Values
      for (const name of Object.getOwnPropertyNames(v)) {
        this.setValue(name, v[name])
      }
    })
  }

  /**
   * Sets single value
   *
   * @param name
   * @param value
   */
  public setValue (name: string, value: undefined|string|string[], index = -1): void {
    // Skip reserved names
    if (reservedFieldNames.includes(name)) {
      return
    }

    // Skip unknown fields
    if (!this[fieldIndex].has(name)) {
      return
    }
    const { kind, isMulti } = this[fieldIndex].get(name) as FieldIndex

    if (value === undefined || value.length === 0) {
      // nothing given, nothing set
      this.values[name] = isMulti ? [] : (kind === 'Bool' ? '0' : undefined)
      return
    }

    if (isMulti) {
      if (Array.isArray(value)) {
        if (index < -1) {
          // assigning [] to [i]
          throw Error('can not set array of values to a single value')
        }

        this.values[name] = Array.isArray(value) ? value : [value]
        return
      }

      if (index === -1) {
        (this.values[name] as string[]).push(value)
        return
      }

      (this.values[name] as string[])[index] = value
      return
    }

    if (Array.isArray(value)) {
      value = value[0]
    }

    // Update with first item or set to undefined
    this.values[name] = value
  }

  public serialize (): Partial<Record> {
    const { toJSON, ...values } = this.values
    return { ...this, values }
  }

  /**
   * Returns resource ID
   */
  get resourceID (): string {
    return `${this.resourceType}:${this.recordID}`
  }

  /**
   * Resource type
   */
  get resourceType (): string {
    return 'compose:record'
  }

  /**
   * Proxy to Record's meta to maintain BC
   */
  get labels (): object {
    return this.meta
  }

  get properties (): string[] {
    return [
      'recordID',
      'moduleID',
      'namespaceID',
      'revision',
      'meta',
      'createdAt',
      'updatedAt',
      'deletedAt',
      'ownedBy',
      'createdBy',
      'updatedBy',
      'deletedBy',
      'canUpdateRecord',
      'canReadRecord',
      'canDeleteRecord',
      'canUndeleteRecord',
      'canManageOwnerOnRecord',
      'canSearchRevision',
      'canGrant',
    ]
  }
}
