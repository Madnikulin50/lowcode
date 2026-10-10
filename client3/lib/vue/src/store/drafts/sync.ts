// Pure helpers for keeping drafts on the server in step with the local copy.

export interface DraftRef {
  namespaceID: string
  moduleID: string
  recordID: string
}

// parseDraftResource reads "compose:record/<namespace>/<module>/<record|0>"
export function parseDraftResource (resource: string): DraftRef | null {
  const m = /^compose:record\/(\d+)\/(\d+)\/(\d+)$/.exec(resource || '')
  return m ? { namespaceID: m[1], moduleID: m[2], recordID: m[3] } : null
}

export interface DraftLike {
  changeID: string
  timestamp?: Date | string | number
}

const time = (d: DraftLike): number => {
  const t = d.timestamp ? new Date(d.timestamp).getTime() : 0
  return Number.isNaN(t) ? 0 : t
}

export interface MergeResult<L, R> {
  // what the store should hold afterwards, by change ID, tagged with where the
  // newest copy came from
  merged: Record<string, { draft: L | R; source: 'local' | 'backend' }>

  // local drafts the server does not have yet or has an older copy of
  toUpload: L[]
}

// mergeDrafts combines the local drafts with the ones from the server. The
// newer copy of a draft wins; a local draft that is unknown to the server (or
// newer than its copy there) is to be uploaded.
export function mergeDrafts<L extends DraftLike, R extends DraftLike> (local: L[], remote: R[]): MergeResult<L, R> {
  const merged: MergeResult<L, R>['merged'] = {}
  const toUpload: L[] = []
  const remoteByID = new Map(remote.map(r => [r.changeID, r]))

  for (const r of remote) {
    merged[r.changeID] = { draft: r, source: 'backend' }
  }

  for (const l of local) {
    const r = remoteByID.get(l.changeID)

    if (!r || time(l) > time(r)) {
      merged[l.changeID] = { draft: l, source: 'local' }
      toUpload.push(l)
    }
  }

  return { merged, toUpload }
}
