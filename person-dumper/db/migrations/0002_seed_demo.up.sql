insert into public.schools (bin, name, db_schema, device_ip, device_port, is_enabled)
values ('970540001234','Demo School','school_970540001234','127.0.0.1',4370,true)
on conflict (bin) do nothing;
