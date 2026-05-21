ALTER TABLE public.clients ADD COLUMN IF NOT EXISTS telephone TEXT DEFAULT '00000000000';
UPDATE public.clients SET telephone = '00000000000' WHERE telephone IS NULL;