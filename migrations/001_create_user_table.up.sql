CREATE TABLE IF NOT EXISTS "user" (
    id serial primary key,
    login varchar(255) not null unique,
    password varchar(255) not null,
    balance bigint not null default 0
)