import Link from 'next/link'

export default function Home() {
  return <main className="container py-12 space-y-6"><h1 className="text-4xl font-bold">SUP-прогулки с инструктором у Анапы</h1><p>Безопасные маршруты по спокойной реке, фото и закатные туры. Выберите инструктора и забронируйте онлайн за 2 минуты.</p><div className="flex gap-3"><Link href="/booking" className="px-4 py-2 bg-blue-600 text-white rounded">Забронировать</Link><Link href="/instructors" className="px-4 py-2 border rounded">Смотреть инструкторов</Link></div></main>
}
