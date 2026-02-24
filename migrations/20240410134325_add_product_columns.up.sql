ALTER TABLE product
    ADD old_price numeric(12, 2);
ALTER TABLE product
    ADD currency varchar(8) not null default 'RUB';
ALTER TABLE product
    ADD validity integer;
