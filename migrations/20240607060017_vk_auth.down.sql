alter table users
    rename column email to phone;

alter table users
    drop column vk_auth_token;

alter table users
    drop column auth_type;

drop type user_auth_type;