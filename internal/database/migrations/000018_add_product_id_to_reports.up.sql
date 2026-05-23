BEGIN;

ALTER TABLE public.reports
    ALTER COLUMN consignment_id DROP NOT NULL;

ALTER TABLE public.reports
    ADD COLUMN product_id INTEGER,
    ADD CONSTRAINT report_product_relation FOREIGN KEY (product_id)
        REFERENCES public.products (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID;

COMMIT;
