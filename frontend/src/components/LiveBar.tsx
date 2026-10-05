export function LiveBar({ live, typing, pressure, target }: { live: boolean; typing: boolean; pressure: number; target: string }) {
  return (
    <div className="flex items-center gap-3 text-xs border-b p-2 bg-white sticky top-0">
      <span className={live ? 'text-green-600' : 'text-gray-400'}>● {live ? 'Live' : 'Offline'}</span>
      {typing && <span className="animate-pulse">Typing...</span>}
      <span className="ml-auto">Pressure: {pressure}/100 {pressure>70?'🔥':''}</span>
      <span>Target {target}</span>
      <span className="text-yellow-600">Incentive ★★★★</span>
    </div>
  )
}

// ReceptionistSection.tsx me use:
 // const { events, live, typing } = useReceptionistStream(session?.id ?? null)
 // <LiveBar live={live} typing={typing} pressure={session?.pressure_score ?? 0} target="3/5" />
 // events.map(ev => ev.type==='expert_wait' && <div>Experts se baat ({ev.payload.done}/{ev.payload.total}, {ev.payload.elapsed}s)</div>)
