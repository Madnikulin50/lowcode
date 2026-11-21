// Session stacktraces carry workflow variables in their typed JSON form,
// e.g. {"text": {"@value": "hi", "@type": "String"}}, nested for Vars. This
// flattens that to plain data.
export function unwrapTyped (v) {
  if (Array.isArray(v)) return v.map(unwrapTyped)
  if (v && typeof v === 'object') {
    if ('@value' in v && '@type' in v) return unwrapTyped(v['@value'])
    return Object.fromEntries(Object.entries(v).map(([k, x]) => [k, unwrapTyped(x)]))
  }
  return v
}

// The AI steps of a session: every frame whose results hold an AI trace
// (the trace result of the ai* workflow functions - server/pkg/aiagent/trace.go).
export function aiSteps (stacktrace) {
  return (Array.isArray(stacktrace) ? stacktrace : [])
    .map((frame) => ({ frame, results: unwrapTyped(frame.results || {}) }))
    .filter(({ results }) => results.trace && typeof results.trace === 'object' && 'agent' in results.trace)
    .map(({ frame, results }) => ({
      stepID: frame.stepID,
      stepTime: frame.stepTime,
      createdAt: frame.createdAt,
      error: frame.error,
      trace: results.trace,
      results: Object.fromEntries(Object.entries(results).filter(([k]) => k !== 'trace')),
    }))
}
