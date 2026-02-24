create table if not exists admin
(
    id            uuid primary key,
    login         varchar(256) unique         not null,
    password_hash text                        not null,
    created_at    timestamp without time zone not null default now()
);

create table if not exists admin_token
(
    id         uuid primary key,
    admin_id    uuid references admin (id)  not null,
    secret     varchar(512)                not null,
    expire     timestamp without time zone not null,
    created_at timestamp without time zone not null default now()
);