CREATE TABLE pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);
