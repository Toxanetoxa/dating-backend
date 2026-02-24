create table if not exists message
(
    id           uuid primary key,
    match_id     uuid references match (id) on delete cascade,
    user_id      uuid references users (id) on delete restrict,
    text         varchar(512)                not null,
    is_read      bool                        not null default false,
    is_delivered bool                        not null default false,
    created_at   timestamp without time zone not null default now()
)