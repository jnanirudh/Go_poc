-- Seed fund_entries

INSERT INTO fund_entries (
    trx_id,
    user_id,
    user_name,
    user_pan_num,
    date_of_purchase,
    nav,
    no_of_units
) VALUES
('TRX1001', 'USR001', 'Rahul Sharma', 'ABCDE1234F', '2025-01-15', 1250, 100),
('TRX1002', 'USR002', 'Priya Verma', 'PQRSV5678K', '2025-02-10', 890, 250),
('TRX1003', 'USR003', 'Amit Gupta', 'LMNOP9876Q', '2025-03-05', 1525, 75),
('TRX1004', 'USR004', 'Sneha Iyer', 'ZXCVB4321M', '2025-04-20', 2100, 50),
('TRX1005', 'USR005', 'Arjun Mehta', 'GHJKL2468P', '2025-05-12', 980, 180);

-- Seed mf_details

INSERT INTO mf_details (
    trx_id,
    fund_name,
    amc_name,
    type
) VALUES
('TRX1001', 'Parag Parikh Flexi Cap Fund', 'PPFAS Mutual Fund', 'Equity'),
('TRX1002', 'SBI Bluechip Fund', 'SBI Mutual Fund', 'Equity'),
('TRX1003', 'HDFC Balanced Advantage Fund', 'HDFC Mutual Fund', 'Hybrid'),
('TRX1004', 'ICICI Prudential Liquid Fund', 'ICICI Prudential Mutual Fund', 'Debt'),
('TRX1005', 'Nippon India Small Cap Fund', 'Nippon India Mutual Fund', 'Equity');