/**
 * SMB2/3 inspection for DSPM. Lists disk shares and a few file names.
 * Does not return file contents.
 */
import net from 'node:net'
import { createHmac, randomBytes } from 'node:crypto'

const NT = {
  0x00000000: 'ok',
  0xC0000016: 'more',
  0xC000006D: 'неверный логин или пароль',
  0xC000006A: 'неверный пароль',
  0xC000006E: 'учётная запись ограничена',
  0xC0000072: 'учётная запись отключена',
  0xC0000234: 'учётная запись заблокирована',
  0xC0000022: 'доступ запрещён',
  0xC00000CC: 'общий ресурс не найден',
}

const CMD = { negotiate: 0, setup: 1, tree: 3, create: 5, read: 8, write: 9, queryDir: 14 }
const FILE_ID_BOTH = 37
const SRVSVC = '4b324fc8-1670-01d3-1278-5a47bf6ee188'
const NDR = '8a885d04-1ceb-11c9-9fe8-08002b104860'

function md4 (data) {
  const s = Buffer.alloc(((data.length + 8) >>> 6 << 6) + 64)
  data.copy(s)
  s[data.length] = 0x80
  s.writeUInt32LE(data.length << 3, s.length - 8)
  let a = 0x67452301
  let b = 0xefcdab89
  let c = 0x98badcfe
  let d = 0x10325476
  const rot = (x, n) => (x << n) | (x >>> (32 - n))
  const f = (x, y, z) => (x & y) | (~x & z)
  const g = (x, y, z) => (x & y) | (x & z) | (y & z)
  const h = (x, y, z) => x ^ y ^ z
  const step = (fn, a0, b0, c0, d0, xk, sh, t) => rot((a0 + fn(b0, c0, d0) + xk + t) >>> 0, sh)
  for (let i = 0; i < s.length; i += 64) {
    const x = Array.from({ length: 16 }, (_, j) => s.readUInt32LE(i + j * 4))
    const aa = a
    const bb = b
    const cc = c
    const dd = d
    const rounds = [
      [f, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15], [3, 7, 11, 19], 0],
      [g, [0, 4, 8, 12, 1, 5, 9, 13, 2, 6, 10, 14, 3, 7, 11, 15], [3, 5, 9, 13], 0x5a827999],
      [h, [0, 8, 4, 12, 2, 10, 6, 14, 1, 9, 5, 13, 3, 11, 7, 15], [3, 9, 11, 15], 0x6ed9eba1],
    ]
    for (const [fn, idx, sh, t] of rounds) {
      for (let j = 0; j < 16; j++) {
        const v = step(fn, a, b, c, d, x[idx[j]], sh[j % 4], t)
        ;[a, b, c, d] = [d, v, b, c]
      }
    }
    a = (a + aa) >>> 0
    b = (b + bb) >>> 0
    c = (c + cc) >>> 0
    d = (d + dd) >>> 0
  }
  const out = Buffer.alloc(16)
  out.writeUInt32LE(a, 0)
  out.writeUInt32LE(b, 4)
  out.writeUInt32LE(c, 8)
  out.writeUInt32LE(d, 12)
  return out
}

function utf16 (s) {
  return Buffer.from(String(s), 'utf16le')
}

function u16str (buf, off, chars) {
  return buf.subarray(off, off + chars * 2).toString('utf16le').replace(/\0+$/, '')
}

function hmacMd5 (key, data) {
  return createHmac('md5', key).update(data).digest()
}

function kdf (key, label, context) {
  const fixed = Buffer.concat([
    Buffer.from(label, 'utf8'),
    Buffer.from([0]),
    Buffer.from(context),
    Buffer.from([0, 0, 0, 128]),
  ])
  const i = Buffer.from([0, 0, 0, 1])
  return createHmac('sha256', key).update(Buffer.concat([i, fixed])).digest().subarray(0, 16)
}

function ntlmv2 (user, password, domain, challenge, targetInfo, clientChallenge) {
  const ntHash = md4(utf16(password))
  const v2 = hmacMd5(ntHash, utf16(`${user.toUpperCase()}${domain}`))
  const time = Buffer.alloc(8)
  const ft = BigInt(Date.now() + 11644473600000) * 10000n
  time.writeBigUInt64LE(ft)
  const blob = Buffer.concat([
    Buffer.from([0x01, 0x01, 0, 0, 0, 0, 0, 0]),
    time,
    clientChallenge,
    Buffer.alloc(4),
    targetInfo.length ? targetInfo : Buffer.from([0, 0, 0, 0]),
    Buffer.alloc(4),
  ])
  const proof = hmacMd5(v2, Buffer.concat([challenge, blob]))
  const ntResp = Buffer.concat([proof, blob])
  const lmResp = Buffer.concat([hmacMd5(v2, Buffer.concat([challenge, clientChallenge])), clientChallenge])
  const sessionKey = hmacMd5(v2, proof)
  return { ntResp, lmResp, sessionKey }
}

function ntlmType1 () {
  const flags = 0x00088207
  const body = Buffer.alloc(32)
  body.write('NTLMSSP\0', 0, 'ascii')
  body.writeUInt32LE(1, 8)
  body.writeUInt32LE(flags, 12)
  return body
}

function ntlmBlob (buf) {
  if (!buf || !buf.length) return null
  const at = buf.indexOf(Buffer.from('NTLMSSP\0', 'ascii'))
  if (at < 0) return null
  return buf.subarray(at)
}

function sessionSecurity (pdu) {
  const off = pdu.readUInt16LE(64 + 4)
  const len = pdu.readUInt16LE(64 + 6)
  if (!len || off + len > pdu.length) return Buffer.alloc(0)
  return pdu.subarray(off, off + len)
}

function parseType2 (buf) {
  const token = ntlmBlob(buf)
  if (!token || token.length < 32) throw new Error('сервер не прислал NTLM')
  const flags = token.readUInt32LE(20)
  const challenge = token.subarray(24, 32)
  let targetInfo = Buffer.alloc(0)
  if (flags & 0x00800000 && token.length >= 48) {
    const len = token.readUInt16LE(40)
    const off = token.readUInt32LE(44)
    if (off + len <= token.length) targetInfo = token.subarray(off, off + len)
  }
  return { challenge, targetInfo }
}

function ntlmType3 (user, password, domain, type2) {
  const clientChallenge = randomBytes(8)
  const { ntResp, lmResp, sessionKey } = ntlmv2(user, password, domain, type2.challenge, type2.targetInfo, clientChallenge)
  const dom = utf16(domain)
  const usr = utf16(user)
  const base = 64
  const body = Buffer.alloc(base + lmResp.length + ntResp.length + dom.length + usr.length)
  body.write('NTLMSSP\0', 0, 'ascii')
  body.writeUInt32LE(3, 8)
  const fields = [lmResp, ntResp, dom, usr, Buffer.alloc(0), Buffer.alloc(0)]
  let cursor = 12
  let off = base
  for (const val of fields) {
    body.writeUInt16LE(val.length, cursor)
    body.writeUInt16LE(val.length, cursor + 2)
    body.writeUInt32LE(off, cursor + 4)
    val.copy(body, off)
    off += val.length
    cursor += 8
  }
  body.writeUInt32LE(0x00088207, 60)
  return { body, sessionKey }
}

const NTLMSSP_NEGOTIATE_ANONYMOUS = 0x00000800

// ntlmType3Anonymous builds a proper NTLM "null session" response: zero-length
// LM/NT challenge responses and the ANONYMOUS negotiate flag set, rather than
// a real (but garbage) NTLMv2 response computed from an empty user/password.
// Servers configured for guest access (e.g. Samba's "guest ok = yes") widely
// expect this explicit anonymous form and reject the latter with
// STATUS_ACCESS_DENIED, which is what inspectSmb was hitting before this.
function ntlmType3Anonymous () {
  const base = 64
  const empty = Buffer.alloc(0)
  const body = Buffer.alloc(base)
  body.write('NTLMSSP\0', 0, 'ascii')
  body.writeUInt32LE(3, 8)
  const fields = [empty, empty, empty, empty, empty, empty]
  let cursor = 12
  const off = base
  for (const val of fields) {
    body.writeUInt16LE(val.length, cursor)
    body.writeUInt16LE(val.length, cursor + 2)
    body.writeUInt32LE(off, cursor + 4)
    cursor += 8
  }
  body.writeUInt32LE(0x00088207 | NTLMSSP_NEGOTIATE_ANONYMOUS, 60)
  return { body, sessionKey: Buffer.alloc(16) }
}

function smbHeader (command, messageId, sessionId, treeId, extra = 0) {
  const h = Buffer.alloc(64 + extra)
  h.writeUInt32LE(0x424d53fe, 0)
  h.writeUInt16LE(64, 4)
  h.writeUInt16LE(command, 12)
  h.writeUInt16LE(1, 14)
  h.writeBigUInt64LE(BigInt(messageId), 24)
  h.writeUInt32LE(treeId >>> 0, 36)
  h.writeBigUInt64LE(BigInt(sessionId), 40)
  return h
}

function sign (packet, key) {
  if (!key) return packet
  packet.writeUInt32LE(packet.readUInt32LE(16) | 0x00000008, 16)
  packet.fill(0, 48, 64)
  const mac = createHmac('sha256', key).update(packet).digest().subarray(0, 16)
  mac.copy(packet, 48)
  return packet
}

function uuidBytes (text) {
  const hex = text.replace(/-/g, '')
  const b = Buffer.from(hex, 'hex')
  const out = Buffer.alloc(16)
  out.writeUInt32LE(b.readUInt32BE(0), 0)
  out.writeUInt16LE(b.readUInt16BE(4), 4)
  out.writeUInt16LE(b.readUInt16BE(6), 6)
  b.copy(out, 8, 8)
  return out
}

function rpcBind () {
  const iface = uuidBytes(SRVSVC)
  const ndr = uuidBytes(NDR)
  const ctx = Buffer.alloc(44)
  ctx.writeUInt16LE(0, 0)
  ctx.writeUInt8(1, 2)
  iface.copy(ctx, 4)
  ctx.writeUInt16LE(3, 20)
  ctx.writeUInt16LE(0, 22)
  ndr.copy(ctx, 24)
  ctx.writeUInt16LE(2, 40)
  const hdr = Buffer.alloc(16 + 12)
  hdr[0] = 5
  hdr[2] = 11
  hdr[3] = 0x03
  hdr[4] = 0x10
  hdr.writeUInt16LE(hdr.length + ctx.length, 8)
  hdr.writeUInt32LE(1, 12)
  hdr.writeUInt16LE(4280, 16)
  hdr.writeUInt16LE(4280, 18)
  hdr.writeUInt8(1, 24)
  return Buffer.concat([hdr, ctx])
}

function ndrString (text, referent) {
  const chars = utf16(`${text}\0`)
  const n = chars.length / 2
  const pad = (4 - (chars.length % 4)) % 4
  const head = Buffer.alloc(16)
  head.writeUInt32LE(referent, 0)
  head.writeUInt32LE(n, 4)
  head.writeUInt32LE(0, 8)
  head.writeUInt32LE(n, 12)
  return Buffer.concat([head, chars, Buffer.alloc(pad)])
}

function rpcShareEnum () {
  const stub = Buffer.concat([
    Buffer.from([0, 0, 0, 0]),
    Buffer.from([1, 0, 0, 0]),
    Buffer.from([0, 0, 2, 0]),
    Buffer.from([0, 0, 0, 0]),
    Buffer.from([0, 0, 0, 0]),
    Buffer.from([0xff, 0xff, 0xff, 0xff]),
    Buffer.from([4, 0, 2, 0]),
    Buffer.from([0, 0, 0, 0]),
  ])
  const hdr = Buffer.alloc(24)
  hdr[0] = 5
  hdr[2] = 0
  hdr[3] = 0x03
  hdr[4] = 0x10
  hdr.writeUInt16LE(hdr.length + stub.length, 8)
  hdr.writeUInt32LE(2, 12)
  hdr.writeUInt32LE(stub.length, 16)
  hdr.writeUInt16LE(15, 22)
  return Buffer.concat([hdr, stub])
}

function parseShares (stub) {
  if (stub.length < 24) return []
  let o = 4
  const container = stub.readUInt32LE(o)
  o += 4
  if (!container) return []
  const count = stub.readUInt32LE(o)
  o += 4
  const arr = stub.readUInt32LE(o)
  o += 4
  if (!arr || !count) return []
  const max = stub.readUInt32LE(o)
  o += 4
  const n = Math.min(count, max, 64)
  const entries = []
  for (let i = 0; i < n; i++) {
    if (o + 12 > stub.length) break
    entries.push({
      namePtr: stub.readUInt32LE(o),
      type: stub.readUInt32LE(o + 4),
      remarkPtr: stub.readUInt32LE(o + 8),
    })
    o += 12
  }
  const readStr = () => {
    if (o + 12 > stub.length) return ''
    const maxCount = stub.readUInt32LE(o)
    const actual = stub.readUInt32LE(o + 8)
    o += 12
    const chars = Math.min(actual, maxCount, 256)
    const text = u16str(stub, o, chars)
    o += chars * 2
    o = (o + 3) & ~3
    return text
  }
  const strings = new Map()
  for (const entry of entries) {
    if (entry.namePtr && !strings.has(entry.namePtr)) strings.set(entry.namePtr, readStr())
    if (entry.remarkPtr && !strings.has(entry.remarkPtr)) strings.set(entry.remarkPtr, readStr())
  }
  return entries.map(entry => ({
    name: strings.get(entry.namePtr) || '',
    type: entry.type & 0x0fffffff,
  })).filter(s => s.name)
}

function parseDir (buf) {
  const names = []
  let o = 0
  while (o + 64 <= buf.length && names.length < 8) {
    const next = buf.readUInt32LE(o)
    const nameLen = buf.readUInt32LE(o + 60)
    const name = buf.subarray(o + 104, o + 104 + nameLen).toString('utf16le')
    if (name && name !== '.' && name !== '..') names.push(name)
    if (!next) break
    o += next
  }
  return names
}

function connect (host, port) {
  return new Promise((resolve, reject) => {
    const socket = net.connect({ host, port, timeout: 5000 })
    const chunks = []
    let buf = Buffer.alloc(0)
    const waiters = []
    const fail = (err) => {
      socket.destroy()
      while (waiters.length) waiters.shift().reject(err)
      reject(err)
    }
    socket.on('error', fail)
    socket.on('timeout', () => fail(new Error('smb: таймаут')))
    socket.on('data', (d) => {
      buf = Buffer.concat([buf, d])
      while (buf.length >= 4) {
        const len = buf.readUIntBE(1, 3)
        if (buf.length < 4 + len) break
        const pdu = buf.subarray(4, 4 + len)
        buf = buf.subarray(4 + len)
        const waiter = waiters.shift()
        if (waiter) waiter.resolve(pdu)
        else chunks.push(pdu)
      }
    })
    socket.once('connect', () => resolve({
      socket,
      recv () {
        if (chunks.length) return Promise.resolve(chunks.shift())
        return new Promise((res, rej) => waiters.push({ resolve: res, reject: rej }))
      },
      send (packet) {
        const frame = Buffer.alloc(4 + packet.length)
        frame.writeUIntBE(packet.length, 1, 3)
        packet.copy(frame, 4)
        socket.write(frame)
      },
    }))
  })
}

function statusOf (pdu) {
  return pdu.readUInt32LE(8)
}

function check (pdu, allow = []) {
  const st = statusOf(pdu)
  if (st === 0 || allow.includes(st)) return st
  throw new Error(NT[st] || `smb status 0x${st.toString(16)}`)
}

export async function inspectSmb (endpoint, username, password) {
  const host = endpoint.host
  const port = Number(endpoint.port || 445)
  const link = await connect(host, port)
  let messageId = 0
  let sessionId = 0n
  let treeId = 0
  let signKey = null
  const dialectBox = { value: 0x0210 }
  const preauth = { hash: null }

  const call = async (command, body, tree = 0) => {
    const hdr = smbHeader(command, messageId++, sessionId, tree || treeId)
    body.copy(hdr, 64)
    const packet = Buffer.concat([hdr.subarray(0, 64), body])
    link.send(sign(packet, signKey))
    let pdu = await link.recv()
    if (statusOf(pdu) === 0x00000103) pdu = await link.recv()
    return pdu
  }

  try {
    const salt = randomBytes(32)
    const guid = randomBytes(16)
    const dialects = [0x0202, 0x0210, 0x0300, 0x0302, 0x0311]
    const fixed = Buffer.alloc(36)
    fixed.writeUInt16LE(36, 0)
    fixed.writeUInt16LE(dialects.length, 2)
    fixed.writeUInt16LE(1, 4)
    guid.copy(fixed, 12)
    const dial = Buffer.alloc(dialects.length * 2)
    dialects.forEach((d, i) => dial.writeUInt16LE(d, i * 2))
    let negotiate = Buffer.concat([fixed, dial])
    const ctxOff = (64 + negotiate.length + 7) & ~7
    fixed.writeUInt32LE(ctxOff, 28)
    fixed.writeUInt16LE(1, 32)
    const ctx = Buffer.alloc(8 + 4 + 32 + 6)
    ctx.writeUInt16LE(1, 0)
    ctx.writeUInt16LE(38, 2)
    ctx.writeUInt16LE(1, 8)
    ctx.writeUInt16LE(32, 10)
    ctx.writeUInt16LE(1, 12)
    salt.copy(ctx, 14)
    const pad = Buffer.alloc(ctxOff - (64 + 36 + dial.length))
    negotiate = Buffer.concat([fixed, dial, pad, ctx])
    const negReq = Buffer.concat([smbHeader(CMD.negotiate, messageId, 0n, 0).subarray(0, 64), negotiate])
    preauth.hash = (await import('node:crypto')).createHash('sha512').update(negReq).digest()
    link.send(negReq)
    messageId++
    const neg = await link.recv()
    check(neg)
    preauth.hash = (await import('node:crypto')).createHash('sha512').update(Buffer.concat([preauth.hash, neg])).digest()
    dialectBox.value = neg.readUInt16LE(64 + 4)

    const setup1 = Buffer.alloc(24)
    setup1.writeUInt16LE(25, 0)
    setup1.writeUInt8(1, 3)
    const t1 = ntlmType1()
    setup1.writeUInt16LE(88, 12)
    setup1.writeUInt16LE(t1.length, 14)
    let pdu = await call(CMD.setup, Buffer.concat([setup1, t1]))
    const setupStatus = check(pdu, [0xC0000016])
    sessionId = pdu.readBigUInt64LE(40)
    const challenge = ntlmBlob(sessionSecurity(pdu))
    if (challenge) {
      const auth = (!username && !password)
        ? ntlmType3Anonymous()
        : ntlmType3(username, password || '', '', parseType2(challenge))
      const setup2 = Buffer.alloc(24)
      setup2.writeUInt16LE(25, 0)
      setup2.writeUInt8(1, 3)
      setup2.writeUInt16LE(88, 12)
      setup2.writeUInt16LE(auth.body.length, 14)
      pdu = await call(CMD.setup, Buffer.concat([setup2, auth.body]))
      check(pdu)
      const dialect = dialectBox.value
      const realKey = auth.sessionKey.some(b => b !== 0)
      if (realKey && dialect >= 0x0311) signKey = kdf(auth.sessionKey, 'SMBSigningKey', preauth.hash)
      else if (realKey && dialect >= 0x0300) signKey = kdf(auth.sessionKey, 'SMB2AESCMAC', 'SmbSign')
      else if (realKey) signKey = auth.sessionKey
    } else if (setupStatus !== 0) {
      throw new Error('сервер не прислал NTLM')
    }

    const openShare = async (share) => {
      const path = utf16(`\\\\${host}\\${share}`)
      const body = Buffer.alloc(8)
      body.writeUInt16LE(9, 0)
      body.writeUInt16LE(64 + 8, 4)
      body.writeUInt16LE(path.length, 6)
      const resp = await call(CMD.tree, Buffer.concat([body, path]), 0)
      check(resp)
      return resp.readUInt32LE(36)
    }

    const listDir = async (share) => {
      const tid = await openShare(share)
      const name = utf16('')
      const create = Buffer.alloc(56)
      create.writeUInt16LE(57, 0)
      create.writeUInt32LE(2, 4)
      create.writeUInt32LE(0x00100081, 24)
      create.writeUInt32LE(7, 32)
      create.writeUInt32LE(1, 36)
      create.writeUInt32LE(1, 40)
      create.writeUInt16LE(64 + 56, 44)
      create.writeUInt16LE(name.length, 46)
      const created = await call(CMD.create, Buffer.concat([create, name]), tid)
      check(created)
      const fileId = created.subarray(64 + 64, 64 + 80)
      const pattern = utf16('*')
      const q = Buffer.alloc(32)
      q.writeUInt16LE(33, 0)
      q.writeUInt8(FILE_ID_BOTH, 2)
      q.writeUInt8(0x01, 3)
      fileId.copy(q, 8)
      q.writeUInt16LE(64 + 32, 24)
      q.writeUInt16LE(pattern.length, 26)
      q.writeUInt32LE(64 * 1024, 28)
      const listed = await call(CMD.queryDir, Buffer.concat([q, pattern]), tid)
      const st = statusOf(listed)
      if (st === 0x80000006 || st === 0xC000000F) return []
      check(listed)
      const outOff = listed.readUInt16LE(64 + 2)
      const outLen = listed.readUInt32LE(64 + 4)
      return parseDir(listed.subarray(outOff, outOff + outLen))
    }

    let shares = []
    if (endpoint.database) shares = [{ name: endpoint.database, type: 0 }]
    else try {
      const tid = await openShare('IPC$')
      const pipeName = utf16('srvsvc')
      const create = Buffer.alloc(56)
      create.writeUInt16LE(57, 0)
      create.writeUInt32LE(2, 4)
      create.writeUInt32LE(0x0012019f, 24)
      create.writeUInt32LE(7, 32)
      create.writeUInt32LE(1, 36)
      create.writeUInt16LE(64 + 56, 44)
      create.writeUInt16LE(pipeName.length, 46)
      const created = await call(CMD.create, Buffer.concat([create, pipeName]), tid)
      check(created)
      const fileId = created.subarray(64 + 64, 64 + 80)
      const writeRpc = async (payload) => {
        const w = Buffer.alloc(48)
        w.writeUInt16LE(49, 0)
        w.writeUInt16LE(64 + 48, 2)
        w.writeUInt32LE(payload.length, 4)
        fileId.copy(w, 16)
        const resp = await call(CMD.write, Buffer.concat([w, payload]), tid)
        check(resp)
        const r = Buffer.alloc(49)
        r.writeUInt16LE(49, 0)
        r.writeUInt32LE(64 * 1024, 4)
        fileId.copy(r, 16)
        r.writeUInt32LE(1, 32)
        const rd = await call(CMD.read, r, tid)
        check(rd)
        const off = rd.readUInt16LE(64 + 2)
        const len = rd.readUInt32LE(64 + 4)
        return rd.subarray(off, off + len)
      }
      const ack = await writeRpc(rpcBind())
      if (ack.length < 16 || ack[2] !== 12) throw new Error('smb: srvsvc недоступен')
      const enumResp = await writeRpc(rpcShareEnum())
      const frag = enumResp.readUInt16LE(8)
      shares = parseShares(enumResp.subarray(24, frag))
    } catch {
      // Anonymous sessions on many servers can't enumerate shares via
      // srvsvc (empty list or access denied) even though a named share
      // still accepts a direct tree connect - fall through to the
      // well-known-name probe below instead of failing the whole scan.
      shares = []
    }

    const disks = shares.filter(s => s.type === 0 && !s.name.endsWith('$')).slice(0, 8)

    if (!disks.length && !endpoint.database) {
      // Guest-only shares (e.g. Samba's "public") are frequently invisible
      // to NetShareEnumAll under an anonymous session even though they can
      // be tree-connected directly - probe the common guest share names.
      for (const name of ['public', 'share', 'shared']) {
        try {
          await openShare(name)
          disks.push({ name, type: 0 })
        } catch { /* not present / not accessible, try next */ }
      }
    }

    const items = []
    for (const share of disks) {
      let names = []
      try { names = await listDir(share.name) } catch { names = [] }
      items.push({
        name: share.name,
        path: `//${host}/${share.name}`,
        sample: names.slice(0, 4).join(', '),
        purpose: 'production',
        status: 'shadow',
      })
    }
    return items
  } finally {
    link.socket.destroy()
  }
}

export function smbSelfTest () {
  const nt = md4(utf16('password')).toString('hex')
  if (nt !== '8846f7eaee8fb117ad06bdd830b7586c') throw new Error(`md4 ${nt}`)
}
