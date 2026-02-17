import Link from 'next/link'
import { api } from '@/lib/api'
import { Instructor } from '@/lib/types'

export default async function InstructorsPage({ searchParams }: { searchParams: { tag?: string } }) {
  const qs = searchParams.tag ? `?tag=${encodeURIComponent(searchParams.tag)}` : ''
  const instructors = await api<Instructor[]>(`/api/instructors${qs}`)
  return <main className="container py-8"><h1 className="text-3xl font-bold mb-4">Инструкторы</h1><div className="mb-4 flex gap-2"><Link href="/instructors?tag=новички" className="px-3 py-1 border rounded">новички</Link><Link href="/instructors?tag=закат" className="px-3 py-1 border rounded">закат</Link><Link href="/instructors?tag=спорт" className="px-3 py-1 border rounded">спорт</Link></div><div className="grid md:grid-cols-3 gap-4">{instructors.map(i=><Link href={`/instructors/${i.id}`} key={i.id} className="bg-white rounded-xl p-4 border"><img src={i.photo_url} alt={i.name} className="w-full h-48 object-cover rounded"/><h3 className="font-semibold mt-2">{i.name}</h3><p className="text-sm text-slate-600">Рейтинг {i.rating} ({i.reviews_count})</p><p className="text-sm">от {i.base_price} ₽</p></Link>)}</div></main>
}
