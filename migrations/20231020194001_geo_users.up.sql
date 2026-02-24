ALTER TABLE users
    DROP COLUMN geo_lat;
ALTER TABLE users
    DROP COLUMN geo_long;
ALTER TABLE users
    ADD geolocation geography(point, 4326) NULL;
