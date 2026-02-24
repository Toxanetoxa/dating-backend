create table if not exists user_photo
(
    id         uuid primary key            not null,
    user_id    uuid references users (uid) not null,
    url        varchar(256)                not null,
    is_main    boolean                              default false,
    created_at timestamp without time zone not null default now()
);