BEGIN;
CREATE TABLE enrollment_order (
    order_id varchar(100) not null PRIMARY KEY,
    user_id int not null references "user" (id),
    status varchar(30) not null,
    accrual bigint not null default 0,
    uploaded_at timestamp not null default now()
);
CREATE INDEX enrollment_order_user_id_idx ON enrollment_order (user_id);
COMMIT;