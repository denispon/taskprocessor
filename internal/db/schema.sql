CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    balance NUMERIC(10,2) NOT NULL DEFAULT 0
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    data TEXT NOT NULL,
    timeout_ns BIGINT NOT NULL DEFAULT 0
);

INSERT INTO users (id, name, balance) VALUES
    (1, 'Alex Johnson', 100.00),
    (2, 'Sam Rivera', 250.00),
    (3, 'Taylor Kim', 75.50),
    (4, 'Jordan Lee', 400.00),
    (5, 'Morgan Chen', 500.00),
    (6, 'Casey Patel', 125.25),
    (7, 'Riley Brooks', 700.00),
    (8, 'Jamie Nguyen', 80.00),
    (9, 'Avery Smith', 900.00),
    (10, 'Quinn Davis', 1000.00)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    balance = EXCLUDED.balance;

INSERT INTO tasks (id, data, timeout_ns) VALUES
    (1, 'Process sample order', 0),
    (2, 'Generate daily report', 5000000000),
    (3, 'Send notification', 2000000000),
    (4, 'Validate customer details', 3000000000),
    (5, 'Update inventory', 4000000000),
    (6, 'Create invoice', 2500000000),
    (7, 'Export transaction data', 6000000000),
    (8, 'Archive completed job', 1500000000),
    (9, 'Reconcile account balance', 3500000000),
    (10, 'Send confirmation email', 2000000000)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users), true);
SELECT setval(pg_get_serial_sequence('tasks', 'id'), (SELECT MAX(id) FROM tasks), true);