CREATE TABLE IF NOT EXISTS households (
    id SERIAL PRIMARY KEY,
    title TEXT
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    first_name TEXT,
    last_name TEXT,
    email TEXT UNIQUE,
    users_households SERIAL REFERENCES households(id)
);

CREATE TABLE IF NOT EXISTS goals (
    id SERIAL PRIMARY KEY,
    purpose TEXT,
    amount MONEY,
    goals_households SERIAL REFERENCES households(id)
);

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    title TEXT,
    institution INT,
    accounts INT,
    transit INT,
    accounts_owner SERIAL REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    about TEXT,
    amount MONEY,
    source TEXT,
    destination TEXT,
    transactions_accounts SERIAL REFERENCES accounts(id)
);

CREATE TABLE IF NOT EXISTS plans (
    id SERIAL PRIMARY KEY,
    title TEXT,
    about TEXT,
    due_date DATE,
    end_goals MONEY,
    plans_households SERIAL REFERENCES households(id)
)