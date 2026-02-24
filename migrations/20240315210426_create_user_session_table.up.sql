CREATE TABLE IF NOT EXISTS user_session
(
    id         uuid primary key,
    user_id    uuid                        not null references users (id) on delete cascade,
    device_id  varchar(256),
    ip         varchar(16),
    created_at timestamp without time zone not null default now()
);