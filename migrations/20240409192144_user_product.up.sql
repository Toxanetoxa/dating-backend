create table if not exists user_product
(
    user_id    uuid references users (id) on delete cascade      not null,
    product_id integer references product (id) on delete cascade not null,
    expire     timestamp without time zone
);