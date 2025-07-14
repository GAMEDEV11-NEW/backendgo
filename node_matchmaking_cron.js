// node_matchmaking_cron.js
// Node.js matchmaking cron job for Cassandra

const cassandra = require('cassandra-driver');
const moment = require('moment');

const client = new cassandra.Client({
  contactPoints: ['172.31.4.229'], // Update with your Cassandra node(s)
  localDataCenter: 'datacenter1', // Update with your data center name
  keyspace: 'myapp', // Update with your keyspace
  authProvider: new cassandra.auth.PlainTextAuthProvider('cassandra', 'cassandra'), // Set your password
});

// Create game pieces for both users
async function createGamePieces(gameId, user1Id, user2Id) {
  const now = new Date();
  
  // Create four pieces for user1
  for (let i = 1; i <= 4; i++) {
    const pieceId = cassandra.types.Uuid.random();
    const pieceType = `piece_${i}`;
    
    await client.execute(
      `INSERT INTO game_pieces (
        game_id, user_id, move_number, piece_id, player_id,
        from_pos_last, to_pos_last, piece_type, captured_piece,
        created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        gameId, user1Id, 0, pieceId, user1Id,
        '', 'initial', pieceType, '', now, now
      ],
      { prepare: true }
    );
  }
  
  // Create four pieces for user2
  for (let i = 1; i <= 4; i++) {
    const pieceId = cassandra.types.Uuid.random();
    const pieceType = `piece_${i}`;
    
    await client.execute(
      `INSERT INTO game_pieces (
        game_id, user_id, move_number, piece_id, player_id,
        from_pos_last, to_pos_last, piece_type, captured_piece,
        created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        gameId, user2Id, 0, pieceId, user2Id,
        '', 'initial', pieceType, '', now, now
      ],
      { prepare: true }
    );
  }
  
  console.log(`✅ Created game pieces for game ${gameId}`);
}

// Create dice rolls for both users
async function createDiceRolls(gameId, user1Id, user2Id) {
  const now = new Date();
  
  // Create dice lookup for user1
  const user1DiceId = cassandra.types.Uuid.random();
  await client.execute(
    'INSERT INTO dice_rolls_lookup (game_id, user_id, dice_id, created_at) VALUES (?, ?, ?, ?)',
    [gameId, user1Id, user1DiceId, now],
    { prepare: true }
  );
  
  // Create dice lookup for user2
  const user2DiceId = cassandra.types.Uuid.random();
  await client.execute(
    'INSERT INTO dice_rolls_lookup (game_id, user_id, dice_id, created_at) VALUES (?, ?, ?, ?)',
    [gameId, user2Id, user2DiceId, now],
    { prepare: true }
  );
  
  console.log(`✅ Created dice rolls for game ${gameId}`);
}

async function runMatchmaking() {
  const today = moment().format('YYYY-MM-DD');
  const statusId = '1';

  for (let leagueIdNum = 1; leagueIdNum <= 10; leagueIdNum++) {
    const leagueId = leagueIdNum.toString();
    const result = await client.execute(
      'SELECT user_id, league_id, joined_at, id, status_id, join_day FROM pending_league_joins WHERE status_id = ? AND join_day = ? AND league_id = ? ORDER BY joined_at ASC LIMIT 100',
      [statusId, today, leagueId]
    );
    const users = result.rows;
    
    console.log(`League ${leagueId}: Found ${users.length} pending users`);

    for (let i = 0; i + 1 < users.length; i += 2) {
      const user1 = users[i];
      const user2 = users[i + 1];
      const matchPairId = cassandra.types.Uuid.random();
      const now = new Date();

      try {
        // Convert UUIDs properly
        const user1Id = user1.id; // Already a Uuid object
        const user2Id = user2.id; // Already a Uuid object
        
        // For debugging
        console.log('Types verification:');
        console.log('matchPairId:', matchPairId, 'type:', matchPairId.constructor.name);
        console.log('user1Id:', user1Id, 'type:', user1Id.constructor.name);
        console.log('user2Id:', user2Id, 'type:', user2Id.constructor.name);

        // Insert with proper parameter passing
        await client.execute(
          'INSERT INTO match_pairs (id, user1_id, user2_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)',
          [
            matchPairId,               // uuid
            user1Id.toString(),        // text (converted from Uuid)
            user2Id.toString(),        // text (converted from Uuid)
            'active',                  // text
            now,                       // timestamp
            now                        // timestamp
          ],
          { prepare: true } // CRITICAL: Use prepared statements
        );
        console.log('✅ Successfully inserted into match_pairs');

        // Create game pieces for both users
        await createGamePieces(matchPairId.toString(), user1.user_id, user2.user_id);
        
        // Create dice rolls for both users
        await createDiceRolls(matchPairId.toString(), user1.user_id, user2.user_id);

        // 4. Update league_joins with opponent info (requires join_month)
        const joinMonth1 = moment(user1.joined_at).format('YYYY-MM');
        const joinMonth2 = moment(user2.joined_at).format('YYYY-MM');
        
        await client.execute(
          'UPDATE league_joins SET opponent_user_id = ?, opponent_league_id = ? WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?',
          [user2.user_id, user2.league_id, user1.user_id, user1.status_id, joinMonth1, new Date(user1.joined_at)],
          { prepare: true }
        );
        await client.execute(
          'UPDATE league_joins SET opponent_user_id = ?, opponent_league_id = ? WHERE user_id = ? AND status_id = ? AND join_month = ? AND joined_at = ?',
          [user1.user_id, user1.league_id, user2.user_id, user2.status_id, joinMonth2, new Date(user2.joined_at)],
          { prepare: true }
        );
        console.log('✅ Updated league_joins with opponent info');

        // 5. Update pending_league_joins with opponent info
        await client.execute(
          'UPDATE pending_league_joins SET opponent_user_id = ? WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?',
          [user2.user_id, user1.status_id, user1.join_day, user1.league_id, new Date(user1.joined_at)],
          { prepare: true }
        );
        await client.execute(
          'UPDATE pending_league_joins SET opponent_user_id = ? WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?',
          [user1.user_id, user2.status_id, user2.join_day, user2.league_id, new Date(user2.joined_at)],
          { prepare: true }
        );
        console.log('✅ Updated pending_league_joins with opponent info');

        // 6. Delete both users from pending_league_joins
        await client.execute(
          'DELETE FROM pending_league_joins WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?',
          [user1.status_id, user1.join_day, user1.league_id, new Date(user1.joined_at)],
          { prepare: true }
        );
        await client.execute(
          'DELETE FROM pending_league_joins WHERE status_id = ? AND join_day = ? AND league_id = ? AND joined_at = ?',
          [user2.status_id, user2.join_day, user2.league_id, new Date(user2.joined_at)],
          { prepare: true }
        );
        console.log('✅ Deleted both users from pending_league_joins');

        console.log(`✅ Successfully matched: ${user1.user_id} (${user1.id}) vs ${user2.user_id} (${user2.id}) in league ${leagueId}`);
        
      } catch (err) {
        console.error('Detailed error:', {
          message: err.message,
          query: err.query,
          stack: err.stack,
          inner: err.inner
        });
        throw err;
      }
    }
  }
}

async function main() {
  try {
    await client.connect();
    console.log('Connected to Cassandra. Starting matchmaking cron...');
    
    // Do-while style loop: run matchmaking, then wait for interval
    while (true) {
      try {
        await runMatchmaking();
        console.log('Matchmaking run completed, waiting 5 seconds...');
        await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5 seconds
      } catch (err) {
        console.error('Error in matchmaking run:', err);
        await new Promise(resolve => setTimeout(resolve, 5000)); // Wait before retry
      }
    }
  } catch (err) {
    console.error('Error connecting to Cassandra:', err);
    process.exit(1);
  }
}

main(); 