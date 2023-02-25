-- Write your migrate up statements here
CREATE TABLE node_images (
    mac_address text PRIMARY KEY CHECK (mac_address != '') NOT NULL,
    image_tag text NOT NULL CHECK (image_tag != ''),
    image_type text NOT NULL CHECK (image_type != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: node_images_history table to keep track of each change here.
);
INSERT into node_images (mac_address, image_tag, image_type)
    VALUES ('test_mac', 'latest', 'display');
INSERT into node_images (mac_address, image_tag, image_type)
    VALUES ('test_mac_develop', 'develop', 'nodisplay-ofed');

CREATE TABLE images (
    image_bucket text NOT NULL CHECK (image_bucket != ''),
    image_name text NOT NULL CHECK (image_name != ''),
    image_tag text NOT NULL CHECK (image_tag != ''),
    image_type text NOT NULL CHECK (image_type != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (image_tag, image_type)
    -- TODO: images_history table to keep track of each change here.
);

INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-1.23.0-geforce.img', '1.23.0', 'base');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-develop-display.20230228-0243.img', 'develop', 'display');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-develop-nodisplay.20230228-0243.img', 'develop', 'nodisplay');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-develop-nodisplay-ofed.20230228-0243.img', 'develop', 'nodisplay-ofed');
-- TODO: CI will introduce new images with versioned tag and updates the entries with 'latest'
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-1.24.0-display.img', '1.24.0', 'display');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-1.24.0-nodisplay.img', '1.24.0', 'nodisplay');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-1.24.0-nodisplay-ofed.img', '1.24.0', 'nodisplay-ofed');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-1.24.0-display.img', 'latest', 'display');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-1.24.0-nodisplay.img', 'latest', 'nodisplay');
INSERT into images (image_bucket, image_name, image_tag, image_type)
    VALUES ('coreweave-ncore-images', 'ncore-1.24.0-nodisplay-ofed.img', 'latest', 'nodisplay-ofed');

---- create above / drop below ----
DROP TABLE node_images;
DROP TABLE images;
