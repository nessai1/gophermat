BEGIN;
DROP INDEX IF EXISTS enrollment_order_user_id_idx;
DROP TABLE IF EXISTS enrollment_order;
COMMIT;