CREATE TABLE pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE,

    author_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    status_id SMALLINT NOT NULL DEFAULT 1 REFERENCES pr_statuses(id) ON DELETE RESTRICT
);

CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status_id ON pull_requests(status_id);
