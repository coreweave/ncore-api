-- Write your migrate up statements here
CREATE TABLE node_images (
    mac_address text PRIMARY KEY CHECK (mac_address != '') NOT NULL,
    image_tag text NOT NULL CHECK (image_tag != ''),
    image_type text NOT NULL CHECK (image_type != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: node_images_history table to keep track of each change here.
);


CREATE TABLE images (
    image_bucket text NOT NULL CHECK (image_bucket != ''),
    image_name text NOT NULL CHECK (image_name != ''),
    image_cmdline text NOT NULL CHECK (image_cmdline != ''),
    image_tag text NOT NULL CHECK (image_tag != ''),
    image_type text NOT NULL CHECK (image_type != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (image_tag, image_type)
    -- TODO: images_history table to keep track of each change here.
);


---- create above / drop below ----
DROP TABLE node_images;
DROP TABLE images;
