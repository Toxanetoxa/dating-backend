create type user_status as ENUM ('active', 'new', 'inactive', 'deleted');
create type user_sex as ENUM ('male', 'female', 'undefined');
create table if not exists users
(
    uid        uuid primary key,
    phone      varchar(16) unique          not null,
    status     user_status                          default 'new',

    first_name varchar(255),
    sex        user_sex,
    birthday   timestamp without time zone,
    city       varchar(128),
    about      varchar(512),

    geo_lat    numeric(10, 8),
    geo_long   numeric(10, 8),

    created_at timestamp without time zone not null default now(),
    updated_at timestamp without time zone not null default now()
);