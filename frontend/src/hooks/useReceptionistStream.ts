import { fetchEventSource } from '@microsoft/fetch-event-source'

export function useReceptionistStream(sessionId: string, onEvent: (d:any)=>void){
 let retries=0
 const token = localStorage.getItem('token')
 fetchEventSource(`/api/receptionist/sessions/${sessionId}/events?token=${token}`, {
  headers: { Authorization: `Bearer ${token}` },
  onopen(res){
   if(res.status===401){ window.location.href='/login'; throw new Error('401') }
   retries=0
   return Promise.resolve()
  },
  onmessage(ev){
   if(ev.event==='ping') return
   try{ onEvent(JSON.parse(ev.data)) }catch{}
  },
  onerror(){
   const delay = Math.min(1000 * Math.pow(2, retries++), 30000)
   return delay
  }
 })
}
