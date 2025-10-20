CREATE TABLE IF NOT EXISTS user_invitation (
    token bytea PRIMARY KEY,
    user_id bigint NOT NULL
);

ALTER TABLE
user_invitation
ADD CONSTRAINT fk_user
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
