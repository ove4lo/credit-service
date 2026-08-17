CREATE TABLE IF NOT EXISTS applications (
    id SERIAL PRIMARY KEY,
    client TEXT NOT NULL,
    amount INTEGER NOT NULL,
    term INTEGER NOT NULL,
    status TEXT NOT NULL
);