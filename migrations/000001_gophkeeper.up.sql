CREATE EXTENSION pgcrypto;
CREATE TABLE users (
                        user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        login VARCHAR(255),
                        password VARCHAR(255)
);

CREATE TABLE users_data (
                       record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       user_id VARCHAR(255),
                       record_type VARCHAR(255),
                       metadata VARCHAR(255),
                       encoded_data VARCHAR(255)
);