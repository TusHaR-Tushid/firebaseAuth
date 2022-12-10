CREATE TABLE IF NOT EXISTS users(
                                    id serial primary key not null ,
                                    name TEXT NOT NULL ,
                                    email TEXT UNIQUE CHECK (email <> '') NOT NULL ,
                                    password TEXT NOT NULL ,
                                    phone_no TEXT NOT NULL ,
                                    age INTEGER  NOT NULL ,
                                    gender TEXT NOT NULL ,
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                    archived_at TIMESTAMP WITH TIME ZONE
);

create type status_type as enum('pending', 'accepted');

CREATE TABLE IF NOT EXISTS friend_request(
                                    id serial primary key not null ,
                                    request_from INTEGER REFERENCES users(id),
                                    request_to INTEGER REFERENCES users(id),
                                    status status_type NOT NULL DEFAULT 'pending',
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL ,
                                    archived_at TIMESTAMP WITH TIME ZONE
);