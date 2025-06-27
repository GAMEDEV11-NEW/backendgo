from cassandra.cluster import Cluster
from cassandra.auth import PlainTextAuthProvider

# Use credentials and host from Go code
CASSANDRA_HOST = '172.31.4.229'
CASSANDRA_USERNAME = 'cassandra'
CASSANDRA_PASSWORD = 'cassandra'

# Set up authentication
auth_provider = PlainTextAuthProvider(username=CASSANDRA_USERNAME, password=CASSANDRA_PASSWORD)

# Connect to the Cassandra cluster
cluster = Cluster([CASSANDRA_HOST], auth_provider=auth_provider)
session = cluster.connect()

# Step 1: Create Keyspace
KEYSPACE = "myapp"
session.execute(f"""
    CREATE KEYSPACE IF NOT EXISTS {KEYSPACE}
    WITH REPLICATION = {{
        'class': 'SimpleStrategy',
        'replication_factor': 1
    }}
""")

# Step 2: Use the keyspace
session.set_keyspace(KEYSPACE)

# Step 3: Create the table
session.execute("""
    CREATE TABLE IF NOT EXISTS user_sessions (
        user_id TEXT,
        session_token TEXT,
        created_at TIMESTAMP,
        jwt_token TEXT,
        device_id TEXT,
        fcm_token TEXT,
        mobile_no TEXT,
        is_active BOOLEAN,
        updated_at TIMESTAMP,
        PRIMARY KEY ((user_id), created_at)
    ) WITH CLUSTERING ORDER BY (created_at DESC)
      AND compaction = {'class': 'SizeTieredCompactionStrategy'}
""")

print("✅ Keyspace and table created successfully!")
