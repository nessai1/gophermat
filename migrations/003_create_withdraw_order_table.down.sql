BEGIN;
DROP INDEX IF EXISTS withdraw_order_user_id_idx;
DROP TABLE IF EXISTS withdraw_order;
COMMIT;