import redis

# Connect to the proxy with tenant credentials
redis_client = redis.Redis(
    host='localhost',
    port=6380,
    username='tenant1',  # This will be used to determine the tenant prefix
    password='your_password',
    decode_responses=True
)

# Use Redis normally - the proxy will add "tenant1:" prefix to all keys
redis_client.set('user:1:profile', 'John Doe')  # Actual key: "tenant1:user:1:profile"
profile = redis_client.get('user:1:profile')    # Fetches "tenant1:user:1:profile"
print(profile)  # "John Doe"

# Multi-key operations work too
redis_client.mset({
    'user:1:email': 'john@example.com',
    'user:1:age': '30'
})  # Prefixes all keys with "tenant1:"

values = redis_client.mget(['user:1:email', 'user:1:age'])
print(values)  # ['john@example.com', '30']
