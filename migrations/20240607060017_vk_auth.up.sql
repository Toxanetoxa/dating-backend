create type user_auth_type as ENUM ('email', 'vk');

alter table users
    add column auth_type user_auth_type not null default 'email';

alter table users
    add column vk_auth_token varchar(512);

alter table users
    rename column phone to email;