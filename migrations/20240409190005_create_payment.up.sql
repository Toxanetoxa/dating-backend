create type payment_status as ENUM ('new', 'pending', 'error', 'paid', 'canceled', 'refunded');

create table if not exists payment
(
    id              uuid primary key,
    external_id     varchar(256),

    status          payment_status              not null default 'new',
    external_status varchar(256),

    price           numeric(12, 2)              not null,
    currency        varchar(8),

    paid_price      numeric(12, 2),
    paid_currency   varchar(8),

    description     varchar(256),

    user_id         uuid                        references users (id) on delete set null,
    product_id      integer                     references product (id) on delete set null,

    created_at      timestamp without time zone not null default now(),
    updated_at      timestamp without time zone not null default now()
);