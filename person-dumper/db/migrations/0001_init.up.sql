create table if not exists public.schools (
  id           bigserial primary key,
  bin          text        not null unique,
  name         text        not null,
  db_schema    text        not null,
  device_ip    inet        not null,
  device_port  int         not null default 4370,
  is_enabled   boolean     not null default true,
  created_at   timestamptz not null default now(),
  updated_at   timestamptz not null default now()
);

create index if not exists schools_enabled_idx on public.schools (is_enabled);
create index if not exists schools_bin_idx     on public.schools (bin);

create table if not exists public.sync_runs (
  id           bigserial primary key,
  run_id       text        not null,
  school_id    bigint      not null references public.schools(id) on delete cascade,
  started_at   timestamptz not null default now(),
  finished_at  timestamptz,
  total        int         not null default 0,
  inserted     int         not null default 0,
  updated      int         not null default 0,
  failed       int         not null default 0,
  ok           boolean,
  details      jsonb
);

create index if not exists sync_runs_run_id_idx    on public.sync_runs (run_id);
create index if not exists sync_runs_school_id_idx on public.sync_runs (school_id);
