BEGIN;

ALTER TABLE IF EXISTS public.saques
    DROP COLUMN IF EXISTS pix_key,
    DROP COLUMN IF EXISTS pix_titular,
    DROP COLUMN IF EXISTS transfer_details,
    DROP COLUMN IF EXISTS bank_name;

ALTER TABLE IF EXISTS public.saques
    ADD COLUMN bank_account_id integer;

ALTER TABLE IF EXISTS public.saques
    ADD CONSTRAINT saques_bank_account_relation FOREIGN KEY (bank_account_id)
    REFERENCES public.bank_accounts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE SET NULL
    NOT VALID;

COMMIT;
