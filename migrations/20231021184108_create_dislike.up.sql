create table if not exists user_dislike
(
    id           uuid primary key,
    from_user_id uuid references users (id)  not null,
    to_user_id   uuid references users (id)  not null,
    created_at   timestamp without time zone not null default now()
)