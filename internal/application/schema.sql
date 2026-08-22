CREATE TABLE IF NOT EXISTS applications (
    id SERIAL PRIMARY KEY,
    client TEXT NOT NULL,
    amount INTEGER NOT NULL,
    term INTEGER NOT NULL,
    status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS debts (
    id SERIAL PRIMARY KEY,
    client TEXT NOT NULL,
    amount INTEGER NOT NULL,
    closed BOOLEAN NOT NULL DEFAULT false
);

-- Test data to give the verification process something to calculate
INSERT INTO debts (client, amount, closed) VALUES
    ('Aleksandra', 300000, false),
    ('Aleksandra', 400000, false), -- She has 700,000 in outstanding debt → should be rejected
    ('Aleksandra', 200000, true), -- Settled debt; not included in the total amount
    ('Ivan', 100000, false), -- He has a debt of 100,000 → approved
    ('Al', 0, false);