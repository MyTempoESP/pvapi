-- add price column to cart_product
ALTER TABLE public.cart_product
ADD COLUMN price NUMERIC(12,2) NOT NULL DEFAULT 0;
