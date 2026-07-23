CREATE DATABASE fund_db;
DROP TABLE IF EXISTS mf_details;
DROP TABLE IF EXISTS fund_entries;

CREATE TABLE fund_entries (
    trx_id            VARCHAR(50) PRIMARY KEY,
    user_id           VARCHAR(50) NOT NULL,
    user_name         VARCHAR(255) NOT NULL,
    user_pan_num      VARCHAR(20) NOT NULL,
    date_of_purchase  DATE NOT NULL,
    nav 			  BIGINT NOT NULL,
    no_of_units       BIGINT NOT NULL
);

CREATE TABLE mf_details (
    id                SERIAL PRIMARY KEY,
    trx_id            VARCHAR(50) NOT NULL REFERENCES fund_entries(trx_id) ON DELETE CASCADE,
    fund_name         VARCHAR(255) NOT NULL,
    amc_name          VARCHAR(255) NOT NULL,
    type              VARCHAR(50) NOT NULL
);

CREATE INDEX idx_mf_details_trx_id ON mf_details(trx_id);
