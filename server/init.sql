
-- Users table to store user information
CREATE TABLE IF NOT EXISTS users (
	id              UUID    PRIMARY KEY,
	username        TEXT    NOT NULL UNIQUE,
	email           TEXT    NOT NULL UNIQUE,
	password_hash   TEXT    NOT NULL
);

-- Polls table to store poll information
CREATE TABLE IF NOT EXISTS polls (
	id 			UUID 		    PRIMARY KEY,
	user_id 	UUID 		    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	question	TEXT		    NOT NULL,
	created_at 	TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
	expires_at 	TIMESTAMPTZ     NOT NULL
);

-- Poll options table to store the choices for each poll
CREATE TABLE IF NOT EXISTS poll_options (
	id      	UUID		PRIMARY KEY,
	poll_id 	UUID    	NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
	text    	TEXT    	NOT NULL,
	position    SMALLINT	NOT NULL CHECK (position BETWEEN 0 AND 5),

	-- Add a unique constraint to prevent duplicate positions within a poll
	UNIQUE		(poll_id, position)
);

-- Votes table to store user votes
CREATE TABLE votes (
	id			UUID			PRIMARY KEY,
	poll_id		UUID			NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
	option_id	UUID			NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,
	voted_at	TIMESTAMPTZ		NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_polls_user_id ON polls(user_id);
CREATE INDEX idx_poll_options_poll_id ON poll_options(poll_id);
CREATE INDEX idx_votes_poll_id ON votes(poll_id);
CREATE INDEX idx_votes_option_id ON votes(option_id);