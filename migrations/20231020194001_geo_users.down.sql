ALTER TABLE users
    ADD COLUMN geo_lat numeric(10, 8) NULL;
ALTER TABLE users
    ADD COLUMN geo_long numeric(10, 8) NULL;
ALTER TABLE users
    DROP COLUMN geolocation;