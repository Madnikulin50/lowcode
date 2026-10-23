/**
 * MinIO inspection for DSPM. Lists buckets and a few object names.
 * Does not return object contents.
 */
import { createHash, createHmac } from 'node:crypto'

const EMPTY = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'

function sha256 (data) {
  return createHash('sha256').update(data).digest('hex')
}

function hmac (key, data) {
  return createHmac('sha256', key).update(data).digest()
}

function awsEncode (value) {
  return encodeURIComponent(value).replace(/[!'()*]/g, ch => `%${ch.charCodeAt(0).toString(16).toUpperCase()}`)
}

export function signV4 ({ method, uri, query, host, accessKey, secretKey, region, amzDate, payloadHash, extra = {} }) {
  const headers = {
    host,
    'x-amz-content-sha256': payloadHash,
    'x-amz-date': amzDate,
    ...extra,
  }
  const names = Object.keys(headers).sort()
  const canonical = [
    method,
    uri,
    query,
    names.map(name => `${name}:${headers[name]}\n`).join(''),
    names.join(';'),
    payloadHash,
  ].join('\n')
  const scope = `${amzDate.slice(0, 8)}/${region}/s3/aws4_request`
  const toSign = ['AWS4-HMAC-SHA256', amzDate, scope, sha256(canonical)].join('\n')
  const key = hmac(hmac(hmac(hmac(`AWS4${secretKey}`, amzDate.slice(0, 8)), region), 's3'), 'aws4_request')
  const signature = createHmac('sha256', key).update(toSign).digest('hex')
  return {
    ...headers,
    authorization: `AWS4-HMAC-SHA256 Credential=${accessKey}/${scope}, SignedHeaders=${names.join(';')}, Signature=${signature}`,
  }
}

function stamp (date) {
  return date.toISOString().replace(/[:-]|\.\d{3}/g, '')
}

async function s3 (host, port, accessKey, secretKey, uri, query) {
  const amzDate = stamp(new Date())
  const signed = signV4({
    method: 'GET',
    uri,
    query,
    host: `${host}:${port}`,
    accessKey,
    secretKey,
    region: 'us-east-1',
    amzDate,
    payloadHash: EMPTY,
  })
  const url = `http://${host}:${port}${uri}${query ? `?${query}` : ''}`
  const res = await fetch(url, {
    headers: {
      Host: signed.host,
      'x-amz-content-sha256': signed['x-amz-content-sha256'],
      'x-amz-date': signed['x-amz-date'],
      Authorization: signed.authorization,
    },
    signal: AbortSignal.timeout(5000),
  })
  const text = await res.text()
  if (res.status === 403) throw new Error('неверный ключ доступа или секрет')
  if (!res.ok) {
    const code = (text.match(/<Code>([^<]+)<\/Code>/) || [])[1] || `s3 ${res.status}`
    if (code === 'NoSuchBucket') throw new Error('бакет не найден')
    throw new Error(code)
  }
  return text
}

function bucketNames (xml) {
  return [...xml.matchAll(/<Bucket>[\s\S]*?<Name>([^<]+)<\/Name>/g)].map(m => m[1])
}

function objectKeys (xml) {
  return [...xml.matchAll(/<Key>([^<]*)<\/Key>/g)].map(m => m[1])
}

export async function inspectMinio (endpoint, accessKey, secretKey) {
  const host = endpoint.host
  const port = endpoint.port || '9000'
  let names = []
  if (endpoint.database) names = [endpoint.database]
  else names = bucketNames(await s3(host, port, accessKey, secretKey, '/', '')).slice(0, 8)
  const items = []
  for (const name of names) {
    const uri = `/${name.split('/').map(awsEncode).join('/')}`
    let keys = []
    try {
      keys = objectKeys(await s3(host, port, accessKey, secretKey, uri, 'list-type=2&max-keys=4'))
    } catch (err) {
      if (endpoint.database) throw err
      keys = []
    }
    items.push({
      name,
      path: `s3://${host}:${port}/${name}`,
      sample: keys.slice(0, 4).join(', '),
      purpose: 'production',
      status: 'shadow',
    })
  }
  return items
}

export function minioSelfTest () {
  const signed = signV4({
    method: 'GET',
    uri: '/test.txt',
    query: '',
    host: 'examplebucket.s3.amazonaws.com',
    accessKey: 'AKIAIOSFODNN7EXAMPLE',
    secretKey: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
    region: 'us-east-1',
    amzDate: '20130524T000000Z',
    payloadHash: EMPTY,
    extra: { range: 'bytes=0-9' },
  })
  const signature = signed.authorization.split('Signature=')[1]
  if (signature !== 'f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41') {
    throw new Error(`sigv4 ${signature}`)
  }
}
