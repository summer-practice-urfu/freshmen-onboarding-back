-- Table: public.Posts

-- DROP TABLE IF EXISTS public."Posts";

CREATE TABLE IF NOT EXISTS public."Posts"
(
    id uuid NOT NULL,
    title text COLLATE pg_catalog."default" NOT NULL,
    content text COLLATE pg_catalog."default",
    rating integer NOT NULL DEFAULT 0,
    imgl text COLLATE pg_catalog."default",
    CONSTRAINT "Posts_pkey" PRIMARY KEY (id)
    )