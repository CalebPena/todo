-- Fails on purpose if two accounts differ only by case or whitespace; merge them by hand first.
UPDATE users SET email = lower(trim(email)) WHERE email <> lower(trim(email));
ALTER TABLE users ADD CONSTRAINT users_email_normalized_check CHECK (email = lower(trim(email)));
