CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE instructors (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  photo_url TEXT,
  bio TEXT,
  rating DOUBLE PRECISION DEFAULT 5,
  reviews_count INT DEFAULT 0,
  experience_years INT DEFAULT 1,
  tags JSONB NOT NULL DEFAULT '[]',
  languages JSONB NOT NULL DEFAULT '[]',
  base_price INT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE routes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  duration_minutes INT NOT NULL,
  difficulty TEXT NOT NULL,
  base_price INT NOT NULL,
  description TEXT,
  location_lat DOUBLE PRECISION NOT NULL,
  location_lng DOUBLE PRECISION NOT NULL,
  location_title TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE time_slots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  instructor_id UUID NOT NULL REFERENCES instructors(id),
  route_id UUID NOT NULL REFERENCES routes(id),
  start_at TIMESTAMPTZ NOT NULL,
  end_at TIMESTAMPTZ NOT NULL,
  capacity INT NOT NULL,
  remaining INT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE bookings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  instructor_id UUID NOT NULL REFERENCES instructors(id),
  route_id UUID NOT NULL REFERENCES routes(id),
  slot_id UUID NOT NULL REFERENCES time_slots(id),
  customer_name TEXT NOT NULL,
  phone TEXT NOT NULL,
  messenger TEXT,
  participants INT NOT NULL,
  options JSONB NOT NULL DEFAULT '{}',
  price_total INT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE weather_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  location_lat DOUBLE PRECISION NOT NULL,
  location_lng DOUBLE PRECISION NOT NULL,
  time_from TIMESTAMPTZ NOT NULL,
  time_to TIMESTAMPTZ NOT NULL,
  temperature DOUBLE PRECISION NOT NULL,
  wind_speed DOUBLE PRECISION NOT NULL,
  precipitation DOUBLE PRECISION NOT NULL,
  cloud_cover INT NOT NULL,
  conditions_level TEXT NOT NULL,
  score INT NOT NULL,
  raw JSONB NOT NULL,
  fetched_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_timeslots_filters ON time_slots(start_at, instructor_id, route_id, status);
CREATE INDEX idx_weather_cache ON weather_snapshots(location_lat, location_lng, time_from, fetched_at DESC);
