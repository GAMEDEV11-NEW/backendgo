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

    # # Step 3: Create the sessions table with composite primary key
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS sessions (
    #         mobile_no TEXT,
    #         device_id TEXT,
    #         session_token TEXT,
    #         user_id TEXT,
    #         jwt_token TEXT,
    #         fcm_token TEXT,
    #         created_at TIMESTAMP,
    #         expires_at TIMESTAMP,
    #         is_active BOOLEAN,
    #         updated_at TIMESTAMP,
    #         PRIMARY KEY ((mobile_no, device_id), created_at)
    #     ) WITH CLUSTERING ORDER BY (created_at DESC)
    # """)

    # # Step 4: Create the users table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS users (
    #         id TEXT PRIMARY KEY,
    #         mobile_no TEXT,
    #         email TEXT,
    #         full_name TEXT,
    #         state TEXT,
    #         referral_code TEXT,
    #         referred_by TEXT,
    #         profile_data TEXT,
    #         language_code TEXT,
    #         language_name TEXT,
    #         region_code TEXT,
    #         timezone TEXT,
    #         user_preferences TEXT,
    #         status TEXT,
    #         created_at TIMESTAMP,
    #         updated_at TIMESTAMP
    #     )
    # """)

    # # Step 5: Create the games table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS games (
    #         id TEXT PRIMARY KEY,
    #         name TEXT,
    #         description TEXT,
    #         category TEXT,
    #         icon TEXT,
    #         banner TEXT,
    #         min_players INT,
    #         max_players INT,
    #         difficulty TEXT,
    #         rating DOUBLE,
    #         is_active BOOLEAN,
    #         is_featured BOOLEAN,
    #         tags LIST<TEXT>,
    #         metadata MAP<TEXT, TEXT>,
    #         created_at TEXT,
    #         updated_at TEXT
    #     )
    # """)

    # # Step 6: Create the contests table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS contests (
    #         contest_id TEXT PRIMARY KEY,
    #         contest_name TEXT,
    #         contest_win_price TEXT,
    #         contest_entryfee TEXT,
    #         contest_joinuser INT,
    #         contest_activeuser INT,
    #         contest_starttime TEXT,
    #         contest_endtime TEXT
    #     )
    # """)

    # # Step 7: Create the server_announcements table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS server_announcements (
    #         id TEXT PRIMARY KEY,
    #         title TEXT,
    #         content TEXT,
    #         type TEXT,
    #         priority TEXT,
    #         is_active BOOLEAN,
    #         created_at TEXT
    #     )
    # """)

    # # Step 8: Create the game_updates table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS game_updates (
    #         id TEXT PRIMARY KEY,
    #         game_id TEXT,
    #         version TEXT,
    #         title TEXT,
    #         description TEXT,
    #         features LIST<TEXT>,
    #         bug_fixes LIST<TEXT>,
    #         is_required BOOLEAN,
    #         created_at TEXT
    #     )
    # """)

    # # Step 9: Create the otp_store table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS otp_store(
    #         phone_or_email text,      
    #         otp_code text,           
    #         created_at TEXT,     
    #         expires_at TEXT,     
    #         purpose text,            
    #         is_verified boolean,      
    #         attempt_count int,         
    #         PRIMARY KEY ((phone_or_email), purpose, created_at)
    #     ) WITH CLUSTERING ORDER BY (purpose ASC, created_at DESC);
    # """)

  

    # # Step 10: Create the league_joins table with new schema
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS league_joins (
    #         user_id TEXT,
    #         status_id TEXT,
    #         join_month TEXT,
    #         joined_at TIMESTAMP,
    #         league_id TEXT,
    #         status TEXT,
    #         extra_data TEXT,
    #         id UUID,
    #         invite_code TEXT,
    #         opponent_league_id TEXT,
    #         opponent_user_id TEXT,
    #         role TEXT,
    #         updated_at TIMESTAMP,
    #         PRIMARY KEY ((user_id, status_id, join_month), joined_at DESC)
    #     )
    # """)

    # # Step 11: Create the pending_league_joins table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS pending_league_joins (
    #         status_id TEXT,
    #         join_month TEXT,
    #         join_day TEXT,
    #         league_id TEXT,
    #         joined_at TIMESTAMP,
    #         user_id TEXT,
    #         id UUID,
    #         opponent_user_id TEXT,
    #         PRIMARY KEY ((status_id, join_month, join_day, league_id), joined_at)
    #     ) WITH CLUSTERING ORDER BY (joined_at ASC)
    # """)

    # # Step 12: Create the sessions_by_socket table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS sessions_by_socket (
    #         socket_id TEXT PRIMARY KEY,
    #         mobile_no TEXT,
    #         user_id TEXT,
    #         session_token TEXT,
    #         created_at TIMESTAMP
    #     )
    # """)
    # print("✅ Created sessions_by_socket table")

    # # Step 13: Create the match_pairs table
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS match_pairs (
    #         id UUID PRIMARY KEY,
    #         user1_id TEXT,
    #         user2_id TEXT,
    #         user1_data TEXT,
    #         user2_data TEXT,
    #         status TEXT,
    #         created_at TIMESTAMP,
    #         updated_at TIMESTAMP
    #     )
    # """)
    # print("✅ Created match_pairs table")

    # Step 14: Create the dice_rolls_lookup table for fast lookups by game_id and user_id
    # session.execute("""
    #     CREATE TABLE IF NOT EXISTS dice_rolls_lookup (
    #         game_id TEXT,
    #         user_id TEXT,
    #         dice_id UUID,
    #         created_at TIMESTAMP,
    #         PRIMARY KEY ((game_id, user_id), dice_id)
    #     ) WITH CLUSTERING ORDER BY (dice_id DESC)
    # """)
    # print("✅ Created dice_rolls_lookup table")

    # Drop old dice_rolls_data table if it exists
    try:
        session.execute("DROP TABLE IF EXISTS dice_rolls_data")
        print("🗑️ Dropped old dice_rolls_data table")
    except Exception as e:
        print(f"⚠️ Could not drop old dice_rolls_data table: {e}")

    # Step 15: Create the dice_rolls_data table for full dice roll data indexed by lookup_dice_id and roll_id
    session.execute("""
        CREATE TABLE IF NOT EXISTS dice_rolls_data (
            lookup_dice_id UUID,
            roll_id UUID,
            dice_number INT,
            roll_timestamp TIMESTAMP,
            session_token TEXT,
            device_id TEXT,
            contest_id TEXT,
            created_at TIMESTAMP,
            PRIMARY KEY ((lookup_dice_id), roll_id)
        ) WITH CLUSTERING ORDER BY (roll_id DESC)
    """)
    print("✅ Created dice_rolls_data table")
    
    print("✅ All keyspace and tables created successfully!")

except Exception as e:
    print(f"❌ Error creating database schema: {e}")

finally:
    # Clean up resources
    session.shutdown()
    cluster.shutdown()
    print("🔌 Database connection closed.")