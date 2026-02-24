create table if not exists user_like
(
    id           uuid primary key,
    from_user_id uuid references users (id)  not null,
    to_user_id   uuid references users (id)  not null,
    created_at   timestamp without time zone not null default now()
)