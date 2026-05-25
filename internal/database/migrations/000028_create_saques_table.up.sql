BEGIN;

CREATE TABLE IF NOT EXISTS public.saques
(
    id               serial         NOT NULL,
    shop_id          integer        NOT NULL,
    user_id          integer        NOT NULL,
    description      text           NOT NULL DEFAULT '',
    asked_value      numeric(12, 2) NOT NULL,
    bank_name        text           NOT NULL DEFAULT '',
    pix_cpf          text,
    pix_titular      text,
    transfer_details jsonb,
    status           text           NOT NULL DEFAULT 'pending',
    created_at       timestamp with time zone DEFAULT now(),
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.saques
    ADD CONSTRAINT saque_shop_relation FOREIGN KEY (shop_id)
    REFERENCES public.shops (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
    NOT VALID;

ALTER TABLE IF EXISTS public.saques
    ADD CONSTRAINT saque_user_relation FOREIGN KEY (user_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE CASCADE
    NOT VALID;

COMMIT;
