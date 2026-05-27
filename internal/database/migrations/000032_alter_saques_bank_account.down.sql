ALTER TABLE IF EXISTS public.saques
    DROP COLUMN IF EXISTS bank_account_id;

ALTER TABLE IF EXISTS public.saques
    ADD COLUMN pix_key          text,
    ADD COLUMN pix_titular      text,
    ADD COLUMN transfer_details jsonb,
    ADD COLUMN bank_name        text NOT NULL DEFAULT '';
