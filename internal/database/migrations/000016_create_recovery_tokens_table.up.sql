CREATE TABLE IF NOT EXISTS public.recovery_tokens
(
id serial NOT NULL,
user_id int NOT NULL,
token text NOT NULL,
expiration_date timestamp with time zone NOT NULL,
confirmation_code text NOT NULL,
created_at timestamp with time zone default now(),
PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.recovery_tokens
    ADD CONSTRAINT recovery_token_user_relation FOREIGN KEY (user_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
    NOT VALID;
CREATE INDEX IF NOT EXISTS fki_a
    ON public.recovery_tokens(user_id);

