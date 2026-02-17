import Link from 'next/link'
import { api } from '@/lib/api'
import { Instructor } from '@/lib/types'

export default async function InstructorPage({ params }: { params: { id: string } }) {
  const i = await api<Instructor>(`/api/instructors/${params.id}`)
  return <main className="container py-8"><div className="grid md:grid-cols-2 gap-6 bg-white border rounded-xl p-4"><img src={i.photo_url} alt={i.name} className="rounded-xl w-full h-80 object-cover"/><div><h1 className="text-3xl font-bold">{i.name}</h1><p className="text-slate-600 mt-2">{i.bio}</p><p className="mt-2">Опыт: {i.experience_years} лет</p><p>Языки: {Array.isArray(i.languages)? i.languages.join(', ') : ''}</p><p className="font-semibold mt-2">от {i.base_price} ₽</p><Link href={`/booking?instructor_id=${i.id}`} className="inline-block mt-4 px-4 py-2 bg-blue-600 text-white rounded">Выбрать этого инструктора</Link></div></div></main>
}
