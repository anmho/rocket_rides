

INSERT INTO users (
    id, email, stripe_customer_id
) VALUES (
    123, 'awesome-user@email.com', 'sk_123'
);

INSERT INTO users (
    id, email, stripe_customer_id
) VALUES (
    456, 'cool-user@email.com', 'sk_456'
);

-- Started request
INSERT INTO idempotency_keys (
    id, idempotency_key,
    request_method, request_params, request_path,
    recovery_point, user_id
) VALUES (
     736, 'testKeyStarted',
     'POST', '{}', '/rides',
     'started', 123
);

-- Ride created
INSERT INTO idempotency_keys (
    id, idempotency_key,
    request_method, request_params, request_path,
    recovery_point, user_id
) VALUES (
     737, 'testKeyRideCreated',
     'POST', '{}', '/rides',
     'ride_created', 123
 );

-- Charge created
INSERT INTO idempotency_keys (
    id, idempotency_key,
    request_method, request_params, request_path,
    recovery_point, user_id
) VALUES (
     738, 'testKeyChargeCreated',
     'POST', '{}', '/rides',
     'charge_created', 123
 );

-- Finished request
INSERT INTO idempotency_keys (
    id, idempotency_key,
    request_method, request_params, request_path,
    response_code, response_body,
    recovery_point, user_id
) VALUES (
    739, 'testKeyFinished',
    'POST', '{}', '/rides',
    201, '{}',
    'finished', 123
);


-- Ride where the charge hasn't been created yet
INSERT INTO rides (
    id, idempotency_key_id,
    origin_lat, origin_lon,
    target_lat, target_lon,
    user_id
) VALUES (
    1337, 738,
    1, 2,
    3, 4,
    123
 );

-- Ride where the charge has been created.
INSERT INTO rides (
    id, idempotency_key_id,
    origin_lat, origin_lon,
    target_lat, target_lon,
    stripe_charge_id,
    user_id
) VALUES (
    1442, 738,
    72, 72,
    72, 72,
    'ch_456',
    456
);

-- Test Audit Record
INSERT INTO audit_records (
    id, action, data, origin_ip,
    resource_id, resource_type,
    user_id
) VALUES (
    4321, 'created', '{}', '127.0.0.1',
    1441, 'ride',
    123
)
;