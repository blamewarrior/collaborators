CREATE TABLE repositories (
    id SERIAL primary key,
    full_name varchar(255)
    UNIQUE (full_name)
);

CREATE TABLE collaboration (
    repository_id integer NOT NULL REFERENCES repositories(id),
    account_id integer NOT NULL REFERENCES accounts(id),
    UNIQUE (repository_id, account_id)
);

CREATE TABLE accounts (
    id SERIAL primary key,
    uid varchar(255),
    login varchar(255),
    permissions jsonb,
    UNIQUE (login)
);
