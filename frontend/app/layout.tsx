import './globals.css'
import Link from 'next/link'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <header className="bg-white border-b sticky top-0 z-20">
          <div className="container py-4 flex gap-6"><Link href="/" className="font-bold">SUP Анапа</Link><Link href="/instructors">Инструкторы</Link><Link href="/booking">Бронирование</Link></div>
        </header>
        {children}
      </body>
    </html>
  )
}
