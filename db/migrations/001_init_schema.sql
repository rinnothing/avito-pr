-- +goose Up
CREATE TYPE pr_status AS ENUM ('open', 'merged');

CREATE TABLE teams (
    name TEXT PRIMARY KEY
);

CREATE TABLE users (
    id        TEXT    PRIMARY KEY,
    username  TEXT    NOT NULL UNIQUE,
    team_name TEXT    NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE pull_requests (
    id         TEXT        PRIMARY KEY,
    name       TEXT        NOT NULL,
    author_id  TEXT        NOT NULL,
    status     pr_status   NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at  TIMESTAMPTZ,

    CONSTRAINT constr_pr_author
        FOREIGN KEY (author_id) REFERENCES users(id)
);

CREATE TABLE pr_reviewers (
    pr_id   TEXT NOT NULL,
    user_id TEXT NOT NULL,

    PRIMARY KEY (pr_id, user_id),

    CONSTRAINT constr_pr
        FOREIGN KEY (pr_id) REFERENCES pull_requests(id),

    CONSTRAINT constr_reviewer
        FOREIGN KEY (user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE pr_reviewers;
DROP TABLE pull_requests;
DROP TABLE users;
DROP TABLE teams;
DROP TYPE pr_status;
