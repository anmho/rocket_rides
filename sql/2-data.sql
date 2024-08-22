

INSERT INTO users (
    id, email, stripe_customer_id
) VALUES (
    123, 'awesome-user@email.com', 'sk_123'
);

INSERT INTO idempotency_keys (
    id, idempotency_key,
    request_method, request_params, request_path,
    response_code, response_body,
    recovery_point, user_id) VALUES (
    738, 'testKey',
    'POST', '{}', '/charges',
    200, '{}',
    'finished', 123
);