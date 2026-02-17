INSERT INTO instructors (id,name,photo_url,bio,rating,reviews_count,experience_years,tags,languages,base_price,is_active) VALUES
('11111111-1111-1111-1111-111111111111','Алексей Морев','https://images.unsplash.com/photo-1500648767791-00dcc994a43e','Спокойные прогулки для новичков и семей.',4.9,132,7,'["новички","дети","закат"]','["RU","EN"]',3000,true),
('22222222-2222-2222-2222-222222222222','Мария Волна','https://images.unsplash.com/photo-1494790108377-be9c29b29330','Тренировки и SUP-фитнес на реке.',4.8,96,5,'["спорт","новички"]','["RU"]',3200,true),
('33333333-3333-3333-3333-333333333333','Илья Бриз','https://images.unsplash.com/photo-1506794778202-cad84cf45f1d','Фото-тур на закате, уверенный темп.',4.7,87,6,'["закат","спорт"]','["RU","EN"]',3500,true);

INSERT INTO routes (id,title,duration_minutes,difficulty,base_price,description,location_lat,location_lng,location_title) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','Река у Анапы — спокойная вода',90,'easy',2500,'Идеально для первого SUP: тихая вода, короткие остановки.',45.092,37.268,'Старт: река у Анапы'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','Река у Анапы — закатный маршрут',120,'medium',3200,'Маршрут к золотому часу с фотопаузами.',45.092,37.268,'Старт: река у Анапы');

DO $$
DECLARE d INT;
BEGIN
  FOR d IN 0..6 LOOP
    INSERT INTO time_slots (instructor_id, route_id, start_at, end_at, capacity, remaining, status)
    VALUES
    ('11111111-1111-1111-1111-111111111111','aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', now()::date + d + interval '9 hour', now()::date + d + interval '10 hour 30 minute', 6, 6, 'open'),
    ('22222222-2222-2222-2222-222222222222','aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', now()::date + d + interval '12 hour', now()::date + d + interval '13 hour 30 minute', 8, 8, 'open'),
    ('33333333-3333-3333-3333-333333333333','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', now()::date + d + interval '17 hour', now()::date + d + interval '19 hour', 5, 5, 'open');
  END LOOP;
END $$;
