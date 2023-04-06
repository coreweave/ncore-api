CREATE TABLE subnet_default_images (
    subnet cidr NOT NULL PRIMARY KEY,
    image_tag text NOT NULL CHECK (image_tag != ''),
    image_type text NOT NULL CHECK (image_type != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: add subnet_default_images table to keep track of each change here.
);

---- create above / drop below ----

DROP TABLE subnet_default_images;
