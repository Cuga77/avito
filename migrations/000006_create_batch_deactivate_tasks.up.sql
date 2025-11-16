CREATE TABLE IF NOT EXISTS batch_deactivate_tasks (
    id SERIAL PRIMARY KEY,
    team_id INT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_batch_deactivate_tasks_status
ON batch_deactivate_tasks(status)
WHERE status = 'pending';
