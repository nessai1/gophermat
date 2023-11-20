BEGIN;
CREATE TABLE enrollment_order (
    order_id char(16) not null PRIMARY KEY,
    user_id int not null references "user" (id),
    status varchar(30) not null,
    accrual real not null default 0.0,
    is_accrual_transferred bool not null default false
);
CREATE INDEX enrollment_order_user_id_idx ON enrollment_order (user_id);
COMMIT;