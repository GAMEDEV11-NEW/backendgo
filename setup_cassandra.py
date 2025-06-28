#!/usr/bin/env python3
"""
GOSOCKET Cassandra Database Setup Script
This script sets up the Cassandra keyspace and tables for the GOSOCKET application.
"""

from cassandra.cluster import Cluster
from cassandra.auth import PlainTextAuthProvider
from cassandra.query import SimpleStatement
import sys
import time

# Configuration - Update these values as needed
CASSANDRA_HOST = 'localhost'
CASSANDRA_PORT = 9042
CASSANDRA_USERNAME = 'cassandra'
CASSANDRA_PASSWORD = 'cassandra'
KEYSPACE = 'myapp'

def print_status(message, status="INFO"):
    """Print formatted status message"""
    timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] [{status}] {message}")

def connect_to_cassandra():
    """Connect to Cassandra cluster"""
    try:
        print_status("Connecting to Cassandra cluster...")
        
        # Set up authentication
        auth_provider = PlainTextAuthProvider(
            username=CASSANDRA_USERNAME, 
            password=CASSANDRA_PASSWORD
        )
        
        # Connect to cluster
        cluster = Cluster(
            [CASSANDRA_HOST], 
            port=CASSANDRA_PORT,
            auth_provider=auth_provider,
            connect_timeout=10
        )
        
        session = cluster.connect()
        print_status("✅ Successfully connected to Cassandra", "SUCCESS")
        return cluster, session
        
    except Exception as e:
        print_status(f"❌ Failed to connect to Cassandra: {e}", "ERROR")
        sys.exit(1)

def create_keyspace(session):
    """Create the keyspace if it doesn't exist"""
    try:
        print_status("Creating keyspace...")
        
        keyspace_query = f"""
        CREATE KEYSPACE IF NOT EXISTS {KEYSPACE}
        WITH REPLICATION = {{
            'class': 'SimpleStrategy',
            'replication_factor': 1
        }}
        AND DURABLE_WRITES = true
        """
        
        session.execute(keyspace_query)
        print_status(f"✅ Keyspace '{KEYSPACE}' created successfully", "SUCCESS")
        
    except Exception as e:
        print_status(f"❌ Failed to create keyspace: {e}", "ERROR")
        sys.exit(1)

def create_tables(session):
    """Create all required tables"""
    try:
        print_status("Creating tables...")
        
        # Use the keyspace
        session.set_keyspace(KEYSPACE)
        
        # Table 1: Users table
        users_table = """
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
        """
        
        # Table 2: Sessions table with composite primary key
        sessions_table = """
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
        """
        
        # Table 3: Games table
        games_table = """
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
        """
        
        # Table 4: Contests table
        contests_table = """
        CREATE TABLE IF NOT EXISTS contests (
            contest_id TEXT PRIMARY KEY,
            contest_name TEXT,
            contest_win_price TEXT,
            contest_entryfee TEXT,
            contest_joinuser INT,
            contest_activeuser INT,
            contest_starttime TEXT,
            contest_endtime TEXT,
            created_at TIMESTAMP,
            updated_at TIMESTAMP
        )
        """
        
        # Table 5: Server announcements table
        announcements_table = """
        CREATE TABLE IF NOT EXISTS server_announcements (
            id TEXT PRIMARY KEY,
            title TEXT,
            content TEXT,
            type TEXT,
            priority TEXT,
            is_active BOOLEAN,
            created_at TEXT,
            updated_at TIMESTAMP
        )
        """
        
        # Table 6: Game updates table
        game_updates_table = """
        CREATE TABLE IF NOT EXISTS game_updates (
            id TEXT PRIMARY KEY,
            game_id TEXT,
            version TEXT,
            title TEXT,
            description TEXT,
            features LIST<TEXT>,
            bug_fixes LIST<TEXT>,
            is_required BOOLEAN,
            created_at TEXT,
            updated_at TIMESTAMP
        )
        """
        
        # Table 7: User sessions by session token (for quick lookups)
        sessions_by_token_table = """
        CREATE TABLE IF NOT EXISTS sessions_by_token (
            session_token TEXT PRIMARY KEY,
            mobile_no TEXT,
            device_id TEXT,
            user_id TEXT,
            jwt_token TEXT,
            fcm_token TEXT,
            created_at TIMESTAMP,
            expires_at TIMESTAMP,
            is_active BOOLEAN
        )
        """
        
        # Execute all table creation queries
        tables = [
            ("users", users_table),
            ("sessions", sessions_table),
            ("games", games_table),
            ("contests", contests_table),
            ("server_announcements", announcements_table),
            ("game_updates", game_updates_table),
            ("sessions_by_token", sessions_by_token_table)
        ]
        
        for table_name, query in tables:
            try:
                session.execute(query)
                print_status(f"✅ Table '{table_name}' created successfully", "SUCCESS")
            except Exception as e:
                print_status(f"❌ Failed to create table '{table_name}': {e}", "ERROR")
        
    except Exception as e:
        print_status(f"❌ Failed to create tables: {e}", "ERROR")
        sys.exit(1)

def create_indexes(session):
    """Create secondary indexes for better query performance"""
    try:
        print_status("Creating secondary indexes...")
        
        # Index on mobile_no for users table
        session.execute("""
        CREATE INDEX IF NOT EXISTS ON users (mobile_no)
        """)
        
        # Index on session_token for sessions table
        session.execute("""
        CREATE INDEX IF NOT EXISTS ON sessions (session_token)
        """)
        
        # Index on user_id for sessions table
        session.execute("""
        CREATE INDEX IF NOT EXISTS ON sessions (user_id)
        """)
        
        # Index on category for games table
        session.execute("""
        CREATE INDEX IF NOT EXISTS ON games (category)
        """)
        
        # Index on is_active for games table
        session.execute("""
        CREATE INDEX IF NOT EXISTS ON games (is_active)
        """)
        
        print_status("✅ Secondary indexes created successfully", "SUCCESS")
        
    except Exception as e:
        print_status(f"⚠️ Warning: Some indexes may already exist: {e}", "WARNING")

def insert_sample_data(session):
    """Insert sample data for testing"""
    try:
        print_status("Inserting sample data...")
        
        # Sample game data
        sample_game = """
        INSERT INTO games (id, name, description, category, icon, banner, 
                          min_players, max_players, difficulty, rating, 
                          is_active, is_featured, tags, metadata, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """
        
        session.execute(sample_game, (
            "game_001",
            "Sample Game",
            "A sample game for testing",
            "Action",
            "icon_url",
            "banner_url",
            2,
            10,
            "Medium",
            4.5,
            True,
            True,
            ["action", "multiplayer"],
            {"version": "1.0", "platform": "mobile"},
            "2024-01-01",
            "2024-01-01"
        ))
        
        # Sample contest data
        sample_contest = """
        INSERT INTO contests (contest_id, contest_name, contest_win_price, 
                             contest_entryfee, contest_joinuser, contest_activeuser,
                             contest_starttime, contest_endtime, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """
        
        session.execute(sample_contest, (
            "contest_001",
            "Sample Contest",
            "1000",
            "50",
            0,
            0,
            "2024-01-01 10:00:00",
            "2024-01-01 18:00:00",
            "2024-01-01",
            "2024-01-01"
        ))
        
        print_status("✅ Sample data inserted successfully", "SUCCESS")
        
    except Exception as e:
        print_status(f"⚠️ Warning: Sample data insertion failed: {e}", "WARNING")

def verify_setup(session):
    """Verify that all tables were created successfully"""
    try:
        print_status("Verifying database setup...")
        
        # Check if keyspace exists
        keyspace_query = """
        SELECT keyspace_name FROM system_schema.keyspaces 
        WHERE keyspace_name = ?
        """
        result = session.execute(keyspace_query, [KEYSPACE])
        if not result:
            print_status(f"❌ Keyspace '{KEYSPACE}' not found", "ERROR")
            return False
        
        # Check if tables exist
        tables = [
            "users", "sessions", "games", "contests", 
            "server_announcements", "game_updates", "sessions_by_token"
        ]
        
        for table in tables:
            table_query = """
            SELECT table_name FROM system_schema.tables 
            WHERE keyspace_name = ? AND table_name = ?
            """
            result = session.execute(table_query, [KEYSPACE, table])
            if not result:
                print_status(f"❌ Table '{table}' not found", "ERROR")
                return False
            else:
                print_status(f"✅ Table '{table}' verified", "SUCCESS")
        
        print_status("✅ Database setup verification completed successfully", "SUCCESS")
        return True
        
    except Exception as e:
        print_status(f"❌ Verification failed: {e}", "ERROR")
        return False

def main():
    """Main function to set up the database"""
    print_status("🚀 Starting GOSOCKET Cassandra Database Setup", "INFO")
    print_status("=" * 60, "INFO")
    
    # Connect to Cassandra
    cluster, session = connect_to_cassandra()
    
    try:
        # Create keyspace
        create_keyspace(session)
        
        # Create tables
        create_tables(session)
        
        # Create indexes
        create_indexes(session)
        
        # Insert sample data
        insert_sample_data(session)
        
        # Verify setup
        if verify_setup(session):
            print_status("=" * 60, "INFO")
            print_status("🎉 GOSOCKET Cassandra Database Setup Completed Successfully!", "SUCCESS")
            print_status("=" * 60, "INFO")
            print_status("📊 Database Information:", "INFO")
            print_status(f"   Host: {CASSANDRA_HOST}:{CASSANDRA_PORT}", "INFO")
            print_status(f"   Keyspace: {KEYSPACE}", "INFO")
            print_status(f"   Username: {CASSANDRA_USERNAME}", "INFO")
            print_status("=" * 60, "INFO")
            print_status("🚀 You can now start your GOSOCKET application!", "SUCCESS")
        else:
            print_status("❌ Database setup verification failed", "ERROR")
            sys.exit(1)
            
    except Exception as e:
        print_status(f"❌ Setup failed: {e}", "ERROR")
        sys.exit(1)
    finally:
        # Close connections
        session.shutdown()
        cluster.shutdown()
        print_status("🔌 Database connections closed", "INFO")

if __name__ == "__main__":
    main() 