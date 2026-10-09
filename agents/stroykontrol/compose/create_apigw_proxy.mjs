/**
 * Creates an API Gateway route + proxy filter forwarding
 * /gateway/agents/stroykontrol/* to the stroykontrol-web agent, so the
 * Compose page's IFrame block can reference a relative URL instead of an
 * absolute host:port that has to be hand-edited per environment (see
 * page-layout.ts / IFrameBase.vue's window.CortezaAPI-relative resolution).
 *
 *   AGENT_URL=http://localhost:8092 \
 *   COMPOSE_DSN=postgres://postgres:Zse45rdx@127.0.0.1:5432/test11?sslmode=disable \
 *   node create_apigw_proxy.mjs
 */
import { mintToken, detectBase, apiFactory } from './helpers.mjs'

const AGENT_URL = (process.env.AGENT_URL || 'http://localhost:8092').replace(/\/$/, '')
const ENDPOINT = '/agents/stroykontrol/*'

async function main () {
  const token = await mintToken()
  const composeBase = await detectBase(token)
  const systemBase = composeBase.replace(/\/compose$/, '/system')
  console.log('system base:', systemBase, '-> proxying', ENDPOINT, 'to', AGENT_URL)
  const api = apiFactory(systemBase, token)

  const route = await api('POST', '/apigw/route', {
    endpoint: ENDPOINT,
    method: 'GET',
    enabled: true,
    group: '0',
    meta: { debug: false, async: false, description: 'Proxy to stroykontrol-web agent' },
  })
  console.log('created route', route.routeID, route.endpoint)

  const filter = await api('PUT', '/apigw/filter', {
    routeID: route.routeID,
    weight: '0',
    kind: 'processer',
    ref: 'proxy',
    enabled: true,
    params: { location: AGENT_URL, auth: { type: 'noop', params: {} } },
  })
  console.log('created filter', filter.filterID, JSON.stringify(filter.params))
}

main().catch(e => { console.error(e); process.exit(1) })
