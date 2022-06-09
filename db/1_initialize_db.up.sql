CREATE DATABASE IF NOT EXISTS joint;

CREATE TYPE interval AS enum ('weekly', 'bi-weekly', 'monthly');

CREATE TABLE IF NOT EXISTS joint.household (
    id SERIAL PRIMARY KEY,
    title TEXT
);

CREATE TABLE IF NOT EXISTS joint.user (
    id SERIAL PRIMARY KEY,
    first_name TEXT,
    last_name TEXT,
    email TEXT UNIQUE,
    user_household INT REFERENCES joint.household(id)
);

CREATE TABLE IF NOT EXISTS joint.goal (
    id SERIAL PRIMARY KEY,
    purpose TEXT,
    amount MONEY,
    goal_household INT REFERENCES joint.household(id)
);

CREATE TABLE IF NOT EXISTS joint.account (
    id SERIAL PRIMARY KEY,
    title TEXT,
    institution INT,
    account INT,
    transit INT,
    account_owner INT REFERENCES joint.user(id)
);

CREATE TABLE IF NOT EXISTS joint.trans (
    id SERIAL PRIMARY KEY,
    about TEXT,
    amount MONEY,
    source TEXT,
    destination TEXT
    trans_account INT REFERENCES joint.account(id)
);

CREATE TABLE IF NOT EXISTS joint.plan (
    id SERIAL PRIMARY KEY,
    title TEXT,
    about TEXT,
    due_date DATE,
    end_goal MONEY,
    plan_household INT REFERENCES joint.household(id)
)