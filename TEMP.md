Login Request → Generate Session Token → Store in Cassandra (sessions table)
     ↓
OTP Verification → Update User Status → Store JWT in Cassandra (sessions table)
     ↓
Profile Setup → Update User Data → Store in Cassandra (users table)
     ↓
Language Setup → Update Preferences → Store in Cassandra (users table)