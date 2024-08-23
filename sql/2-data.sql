

INSERT INTO users (
    id, email, stripe_customer_id
) VALUES (
    123, 'awesome-user@email.com', 'sk_123'
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

