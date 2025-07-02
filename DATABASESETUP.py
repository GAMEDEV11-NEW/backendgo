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

try:
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

    # Step 3: Create the sessions table with composite primary key
    session.execute("""
        CREATE TABLE IF NOT EXISTS sessions (
            mobile_no TEXT,
            device_id TEXT,
            session_token TEXT,
            user_id TEXT,
            jwt_token TEXT,
            fcm_token TEXT,
            created_at TIMESTAMP,
            expires_at TIMESTAMP,
            is_active BOOLEAN,
            updated_at TIMESTAMP,
            PRIMARY KEY ((mobile_no, device_id), created_at)
        ) WITH CLUSTERING ORDER BY (created_at DESC)
    """)

    # Step 4: Create the users table
    session.execute("""
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            mobile_no TEXT,
            email TEXT,
            full_name TEXT,
            state TEXT,
            referral_code TEXT,
            referred_by TEXT,
            profile_data TEXT,
            language_code TEXT,
            language_name TEXT,
            region_code TEXT,
            timezone TEXT,
            user_preferences TEXT,
            status TEXT,
            created_at TIMESTAMP,
            updated_at TIMESTAMP
        )
    """)

    # Step 5: Create the games table
    session.execute("""
        CREATE TABLE IF NOT EXISTS games (
            id TEXT PRIMARY KEY,
            name TEXT,
            description TEXT,
            category TEXT,
            icon TEXT,
            banner TEXT,
            min_players INT,
            max_players INT,
            difficulty TEXT,
            rating DOUBLE,
            is_active BOOLEAN,
            is_featured BOOLEAN,
            tags LIST<TEXT>,
            metadata MAP<TEXT, TEXT>,
            created_at TEXT,
            updated_at TEXT
        )
    """)

    # Step 6: Create the contests table
    session.execute("""
        CREATE TABLE IF NOT EXISTS contests (
            contest_id TEXT PRIMARY KEY,
            contest_name TEXT,
            contest_win_price TEXT,
            contest_entryfee TEXT,
            contest_joinuser INT,
            contest_activeuser INT,
            contest_starttime TEXT,
            contest_endtime TEXT
        )
    """)

    # Step 7: Create the server_announcements table
    session.execute("""
        CREATE TABLE IF NOT EXISTS server_announcements (
            id TEXT PRIMARY KEY,
            title TEXT,
            content TEXT,
            type TEXT,
            priority TEXT,
            is_active BOOLEAN,
            created_at TEXT
        )
    """)

    # Step 8: Create the game_updates table
    session.execute("""
        CREATE TABLE IF NOT EXISTS game_updates (
            id TEXT PRIMARY KEY,
            game_id TEXT,
            version TEXT,
            title TEXT,
            description TEXT,
            features LIST<TEXT>,
            bug_fixes LIST<TEXT>,
            is_required BOOLEAN,
            created_at TEXT
        )
    """)

    # Step 9: Create the otp_store table
    session.execute("""
        CREATE TABLE IF NOT EXISTS otp_store(
            phone_or_email text,      
            otp_code text,           
            created_at TEXT,     
            expires_at TEXT,     
            purpose text,            
            is_verified boolean,      
            attempt_count int,         
            PRIMARY KEY ((phone_or_email), purpose, created_at)
        ) WITH CLUSTERING ORDER BY (purpose ASC, created_at DESC);
    """)

  

    # Step 10: Create the league_joins table with clustering columns and order for efficient queries
    session.execute("""
        CREATE TABLE IF NOT EXISTS league_joins (
            league_id TEXT,
            status TEXT,
            user_id TEXT,
            id UUID,
            joined_at TEXT,
            updated_at TEXT,
            invite_code TEXT,
            role TEXT,
            extra_data TEXT,
            PRIMARY KEY ((league_id, status), user_id, joined_at)
        ) WITH CLUSTERING ORDER BY (user_id ASC, joined_at DESC)
    """)

    # Add status_id column to existing table
    try:
        session.execute("ALTER TABLE league_joins ADD status_id TEXT")
        print("✅ Added status_id column to league_joins table")
    except Exception as e:
        print(f"ℹ️ status_id column already exists or error: {e}")
    
    print("✅ All keyspace and tables created successfully!")

except Exception as e:
    print(f"❌ Error creating database schema: {e}")

finally:
    # Clean up resources
    session.shutdown()
    cluster.shutdown()
    print("🔌 Database connection closed.")