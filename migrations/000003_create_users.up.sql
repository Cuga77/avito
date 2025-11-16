CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    team_id INT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT
);

CREATE INDEX idx_users_team_active ON users(team_id, is_active);
