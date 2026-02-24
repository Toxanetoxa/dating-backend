create table if not exists match
(
    id         uuid primary key,
    user_1_id  uuid references users (id)  not null,
    user_2_id  uuid references users (id)  not null,
    chat_init  bool                        not null default false,
    created_at timestamp without time zone not null default now()
)