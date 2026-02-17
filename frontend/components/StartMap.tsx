'use client'
import { MapContainer, Marker, Popup, TileLayer } from 'react-leaflet'
import 'leaflet/dist/leaflet.css'

export default function StartMap({ lat, lng }: { lat:number; lng:number }) {
  return <div className="h-72 rounded-xl overflow-hidden border"><MapContainer center={[lat,lng]} zoom={12} className="h-full w-full"><TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"/><Marker position={[lat,lng]}><Popup>Старт: река у Анапы (точка примерная)</Popup></Marker></MapContainer></div>
}
