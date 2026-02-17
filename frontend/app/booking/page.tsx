'use client'
import { useEffect, useMemo, useState } from 'react'
import dynamic from 'next/dynamic'
import { api } from '@/lib/api'
import { Instructor, Route, Slot, Weather } from '@/lib/types'

const StartMap = dynamic(() => import('@/components/StartMap'), { ssr: false })

export default function BookingPage() {
  const [instructors, setInstructors] = useState<Instructor[]>([])
  const [routes, setRoutes] = useState<Route[]>([])
  const [slots, setSlots] = useState<Slot[]>([])
  const [weather, setWeather] = useState<Weather | null>(null)
  const [form, setForm] = useState({ instructor_id:'', route_id:'', slot_id:'', date:new Date().toISOString().slice(0,10), participants:1, customer_name:'', phone:'', messenger:'', photo:false, drybag:false, vest:true })
  const [bookingId, setBookingId] = useState('')

  useEffect(() => { Promise.all([api<Instructor[]>('/api/instructors'), api<Route[]>('/api/routes')]).then(([i,r])=>{setInstructors(i);setRoutes(r); if(i[0]) setForm(f=>({...f,instructor_id:i[0].id})); if(r[0]) setForm(f=>({...f,route_id:r[0].id}))}) }, [])
  useEffect(() => {
    if (!form.date) return
    api<Slot[]>(`/api/availability?date=${form.date}&route_id=${form.route_id}&instructor_id=${form.instructor_id}`).then(setSlots)
  }, [form.date, form.route_id, form.instructor_id])
  useEffect(() => {
    const route = routes.find(r => r.id === form.route_id)
    const slot = slots.find(s=>s.id===form.slot_id)
    if (!route || !slot) return
    api<Weather>(`/api/weather?lat=${route.location_lat}&lng=${route.location_lng}&datetime=${encodeURIComponent(slot.start_at)}&route_id=${route.id}&instructor_id=${form.instructor_id}`).then(setWeather)
  }, [form.slot_id, routes, slots, form.instructor_id, form.route_id])

  const total = useMemo(() => {
    const instructor = instructors.find(i => i.id === form.instructor_id)
    const route = routes.find(r => r.id === form.route_id)
    const opts = (form.photo ? 700 : 0) + (form.drybag ? 200 : 0) + (form.vest ? 0 : 0)
    return ((instructor?.base_price || 0) + (route?.base_price || 0) + opts) * form.participants
  }, [form, instructors, routes])

  const submit = async () => {
    if (!/^\+?[0-9\-\s]{10,15}$/.test(form.phone)) return alert('Введите корректный телефон')
    const res = await api<any>('/api/bookings', { method:'POST', body: JSON.stringify({
      instructor_id: form.instructor_id, route_id: form.route_id, slot_id: form.slot_id,
      customer_name: form.customer_name, phone: form.phone, messenger: form.messenger, participants: form.participants,
      options: { photo: form.photo, drybag: form.drybag, vest: form.vest }, price_total: total
    }) })
    setBookingId(res.id)
  }

  if (bookingId) return <main className="container py-12"><h1 className="text-3xl font-bold">Бронирование подтверждено</h1><p className="mt-2">Номер заказа: <b>{bookingId}</b></p><p className="text-slate-600">Точное место старта отправим в мессенджер после подтверждения.</p></main>

  const route = routes.find(r => r.id === form.route_id)

  return <main className="container py-6 grid lg:grid-cols-[1fr_320px] gap-6"><section className="space-y-4"><h1 className="text-3xl font-bold">Бронирование SUP-прогулки</h1><div className="bg-white p-4 rounded-xl border space-y-3"><h2 className="font-semibold">1) Выбор</h2><select className="w-full border rounded p-2" value={form.instructor_id} onChange={e=>setForm({...form,instructor_id:e.target.value})}>{instructors.map(i=><option key={i.id} value={i.id}>{i.name}</option>)}</select><select className="w-full border rounded p-2" value={form.route_id} onChange={e=>setForm({...form,route_id:e.target.value})}>{routes.map(r=><option key={r.id} value={r.id}>{r.title}</option>)}</select><input type="date" className="w-full border rounded p-2" value={form.date} onChange={e=>setForm({...form,date:e.target.value})}/><select className="w-full border rounded p-2" value={form.slot_id} onChange={e=>setForm({...form,slot_id:e.target.value})}><option value="">Выберите слот</option>{slots.map(s=><option key={s.id} value={s.id}>{new Date(s.start_at).toLocaleString('ru-RU')} · мест: {s.remaining}</option>)}</select></div>
  <div className="bg-white p-4 rounded-xl border"><h2 className="font-semibold mb-2">2) Погода и условия</h2>{weather ? <div className="space-y-1"><p>{weather.temperature}°C · ветер {weather.wind_speed} м/с · осадки {weather.precipitation} мм</p><p>Оценка: <b>{weather.conditions_level}</b> ({weather.score}/100)</p><p className="text-sm text-slate-600">{weather.explanation}</p>{weather.conditions_level === 'Плохие' && <div className="p-2 bg-amber-50 border border-amber-300 rounded"><p className="font-medium">Рекомендуем перенести время.</p>{weather.suggested_slots?.map(s=><button key={s.id} onClick={()=>setForm({...form,slot_id:s.id})} className="mr-2 mt-2 px-2 py-1 border rounded">{new Date(s.start_at).toLocaleTimeString('ru-RU',{hour:'2-digit',minute:'2-digit'})}</button>)}</div>}</div> : <p className="text-slate-500">Выберите слот для прогноза</p>}</div>
  <div className="bg-white p-4 rounded-xl border space-y-2"><h2 className="font-semibold">3) Данные клиента</h2><input placeholder="Имя" className="w-full border rounded p-2" value={form.customer_name} onChange={e=>setForm({...form,customer_name:e.target.value})}/><input placeholder="Телефон" className="w-full border rounded p-2" value={form.phone} onChange={e=>setForm({...form,phone:e.target.value})}/><input placeholder="Мессенджер" className="w-full border rounded p-2" value={form.messenger} onChange={e=>setForm({...form,messenger:e.target.value})}/><label className="block"><input type="checkbox" checked={form.photo} onChange={e=>setForm({...form,photo:e.target.checked})}/> Фото/видео (+700 ₽)</label><label className="block"><input type="checkbox" checked={form.drybag} onChange={e=>setForm({...form,drybag:e.target.checked})}/> Гидромешок (+200 ₽)</label></div>
  <div className="bg-white p-4 rounded-xl border space-y-2"><h2 className="font-semibold">4) Карта старта</h2>{route && <><StartMap lat={route.location_lat} lng={route.location_lng}/><p className="text-sm">{route.location_title}. Точная точка после брони.</p><div className="flex gap-2"><a className="px-3 py-1 border rounded" href={`https://maps.google.com/?q=${route.location_lat},${route.location_lng}`} target="_blank">Google Maps</a><a className="px-3 py-1 border rounded" href={`https://yandex.ru/maps/?pt=${route.location_lng},${route.location_lat}&z=12`} target="_blank">Яндекс Карты</a></div></>}</div></section>
  <aside className="lg:sticky lg:top-20 h-fit bg-white border rounded-xl p-4"><h3 className="font-semibold">Итого</h3><p className="text-2xl font-bold mt-2">{total} ₽</p><button onClick={submit} className="mt-3 w-full py-2 bg-blue-600 text-white rounded">Подтвердить бронь</button></aside></main>
}
