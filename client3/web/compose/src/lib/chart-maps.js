import { registerMap } from 'echarts/core'

/**
 * Registers GeoJSON maps (world, china) with ECharts so the 'map' chart
 * type can render them. Maps are fetched once from the public assets and
 * cached; failures are non-fatal (chart renders empty until reload).
 */
const registered = new Set()

export async function ensureMapRegistered (name) {
  if (registered.has(name)) return true

  try {
    const res = await fetch(`${import.meta.env.BASE_URL || '/'}maps/${name}.json`)
    if (!res.ok) return false
    const geo = await res.json()
    registerMap(name, geo)
    registered.add(name)
    return true
  } catch (e) {
    console.error(`Failed to register map ${name}`, e)
    return false
  }
}

export default { ensureMapRegistered }
