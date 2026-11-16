BEGIN;

CREATE TABLE IF NOT EXISTS public.go_demo_jobs (
    id integer PRIMARY KEY,
    title text NOT NULL,
    status text NOT NULL
);

INSERT INTO public.go_demo_jobs (id, title, status) VALUES
    (1, 'Inspect demo equipment', 'scheduled'),
    (2, 'Replace a sample filter', 'in progress'),
    (3, 'Test the workshop checklist', 'done')
ON CONFLICT (id) DO NOTHING;

COMMIT;
