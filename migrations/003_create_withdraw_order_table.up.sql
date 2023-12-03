BEGIN;
CREATE TABLE IF NOT EXISTS withdraw_order (
    id serial PRIMARY KEY,
    order_id varchar(100) not null,
    user_id int not null references "user" (id),
    processed_at timestamp not null default now(),
    sum bigint not null
);
CREATE INDEX withdraw_order_user_id_idx ON withdraw_order (user_id);
COMMIT;