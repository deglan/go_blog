CREATE TABLE IF NOT EXISTS roles (
    id bigserial PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL DEFAULT 0,
    description TEXT
);

INSERT INTO 
    roles(name,description,level)
VALUES('user', 'user can create posts and comments', 1);

INSERT INTO 
    roles(name,description,level)
VALUES('moderator', 'moderator can update other users posts', 2);

INSERT INTO 
    roles(name,description,level)
VALUES('admin', 'moderator can update and delete other users posts', 3);