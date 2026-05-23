BEGIN;

ALTER TABLE public.reports
    DROP CONSTRAINT report_product_relation;

ALTER TABLE public.reports
    DROP COLUMN product_id;

ALTER TABLE public.reports
    ALTER COLUMN consignment_id SET NOT NULL;

COMMIT;
