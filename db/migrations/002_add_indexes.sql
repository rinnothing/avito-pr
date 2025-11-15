-- +goose Up
CREATE INDEX idx_pr_reviewers ON pr_reviewers(user_id);
CREATE INDEX idx_user_teams ON users(team_name);

-- +goose Down
DROP INDEX IF EXISTS idx_pr_reviewers;
DROP INDEX IF EXISTS idx_user_teams;
