export const IsOf = <T>(v: unknown, ...props: (keyof T)[]): v is T => {
  if (!v || typeof v !== 'object') {
    return false
  }

  for (const prop of props) {
    try {
      if (Object.prototype.hasOwnProperty.call(v, prop)) {
        continue
      }
    } catch {
      // Vue 3 proxies may throw or lie about own keys
    }
    try {
      // Property access still works on reactive class instances.
      // Skip inherited functions (Array#values, Object#toString, …).
      const val = (v as T)[prop]
      if (val !== undefined && typeof val !== 'function') {
        continue
      }
    } catch {
      return false
    }
    return false
  }

  return true
}

// eslint-disable-next-line valid-typeof
const every = (a: unknown|unknown[], t: string): boolean => Array.isArray(a) && a.every(i => typeof i === t)

export const AreStrings = (a: unknown|unknown[]): a is string[] => every(a, 'string')
export const AreBooleans = (a: unknown|unknown[]): a is boolean[] => every(a, 'boolean')
export const AreNumbers = (a: unknown|unknown[]): a is number[] => every(a, 'number')
export const AreObjects = (a: unknown|unknown[]): a is object[] => every(a, 'object')

export function AreObjectsOf<T> (a: unknown|unknown[], ...props: (keyof T)[]): a is T[] {
  if (!a || !Array.isArray(a)) {
    return false
  }

  if (a.length === 0) {
    return true
  }

  return AreObjects(a) && a.every(i => IsOf<T>(i, ...props))
}
