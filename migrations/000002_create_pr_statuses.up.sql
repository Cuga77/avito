CREATE TABLE pr_statuses (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

INSERT INTO pr_statuses (id, name) VALUES (1, 'OPEN'), (2, 'MERGED');
