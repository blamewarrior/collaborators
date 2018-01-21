CREATE TABLE repositories (
    full_name varchar(255)
)

CREATE TABLE collaboration (
    repository_id integer REFERENCES repositories(id),
    account_id integer REFERENCES accounts(id)
);

CREATE TABLE accounts (
    id SERIAL primary key,
    uid varchar(255),
    login varchar(255),
    permissions jsonb
);

ALTER TABLE repositories ADD CONSTRAINT repository_full_name_idx UNIQUE(full_name);
