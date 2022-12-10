CREATE TABLE IF NOT EXISTS sessions(
                                       id serial primary key  not null ,
                                       user_id INTEGER REFERENCES users(id) NOT NULL ,
                                       created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                       expires_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE users ALTER COLUMN name DROP NOT NULL ;
ALTER TABLE users ALTER COLUMN password DROP NOT NULL ;
ALTER TABLE users ALTER COLUMN phone_no DROP NOT NULL ;
ALTER TABLE users ALTER COLUMN age DROP NOT NULL ;
ALTER TABLE users ALTER COLUMN gender DROP NOT NULL ;