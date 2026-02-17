export type Instructor = { id:string; name:string; photo_url:string; bio:string; rating:number; reviews_count:number; experience_years:number; tags:string[]; languages:string[]; base_price:number; is_active:boolean };
export type Route = { id:string; title:string; duration_minutes:number; difficulty:string; base_price:number; description:string; location_lat:number; location_lng:number; location_title:string };
export type Slot = { id:string; instructor_id:string; route_id:string; start_at:string; end_at:string; capacity:number; remaining:number; status:string };
export type Weather = { temperature:number; wind_speed:number; precipitation:number; cloud_cover:number; conditions_level:string; explanation:string; score:number; suggested_slots?: Slot[] };
