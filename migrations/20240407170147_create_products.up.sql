create table if not exists product
(
    id    serial primary key,
    price numeric(12,2) not null,
    name  varchar(128)
);