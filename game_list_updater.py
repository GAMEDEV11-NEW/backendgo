#!/usr/bin/env python3
"""
Game List Updater Script
Updates game list in Redis and sends real-time notifications to connected Socket.IO clients
"""

import json
import time
import random
import requests
from datetime import datetime
import socketio
import threading

class GameListUpdater:
    def __init__(self, redis_host='127.0.0.1', redis_port=6379, socketio_url='http://localhost:8088'):
        self.redis_host = redis_host
        self.redis_port = redis_port
        self.socketio_url = socketio_url
        
        # Initialize Redis client
        try:
            import redis
            self.redis_client = redis.Redis(
                host=redis_host,
                port=redis_port,
                db=0,
                decode_responses=True,
                socket_connect_timeout=5,
                socket_timeout=5
            )
        except ImportError:
            print("❌ Redis module not found. Please run: pip install redis==5.0.1")
            self.redis_client = None
        except Exception as e:
            print(f"❌ Redis connection failed: {e}")
            self.redis_client = None
        
        # Initialize Socket.IO client for sending notifications
        self.sio = socketio.Client()
        self.setup_socketio_handlers()
        
        # Game list template
        self.game_templates = [
            {"name": "Poker", "type": "card"},
            {"name": "Rummy", "type": "card"},
            {"name": "Teen Patti", "type": "card"},
            {"name": "Carrom", "type": "board"},
            {"name": "Ludo", "type": "board"},
            {"name": "Snake & Ladder", "type": "board"},
            {"name": "Chess", "type": "strategy"},
            {"name": "Checkers", "type": "strategy"},
            {"name": "Tic Tac Toe", "type": "puzzle"},
            {"name": "Memory Game", "type": "puzzle"}
        ]
    
    def setup_socketio_handlers(self):
        """Setup Socket.IO event handlers"""
        @self.sio.event
        def connect():
            print("✅ Connected to Socket.IO server")
        
        @self.sio.event
        def disconnect():
            print("❌ Disconnected from Socket.IO server")
        
        @self.sio.event
        def connect_error(data):
            print(f"❌ Socket.IO connection error: {data}")
    
    def connect_to_socketio(self):
        """Connect to Socket.IO server"""
        try:
            self.sio.connect(self.socketio_url)
            return True
        except Exception as e:
            print(f"❌ Failed to connect to Socket.IO server: {e}")
            return False
    
    def disconnect_from_socketio(self):
        """Disconnect from Socket.IO server"""
        try:
            self.sio.disconnect()
        except:
            pass
    
    def generate_game_list(self, num_games=None):
        """Generate a random game list"""
        if num_games is None:
            num_games = random.randint(3, 8)
        
        games = []
        selected_templates = random.sample(self.game_templates, min(num_games, len(self.game_templates)))
        
        for i, template in enumerate(selected_templates):
            game = {
                "game_id": f"game_{i+1}_{int(time.time())}",
                "game_name": template["name"],
                "game_type": template["type"],
                "active_gamepalye": random.randint(100, 50000),
                "livegameplaye": random.randint(50, 20000),
                "min_players": random.randint(2, 4),
                "max_players": random.randint(4, 8),
                "status": random.choice(["active", "maintenance", "coming_soon"]),
                "created_at": datetime.now().isoformat()
            }
            games.append(game)
        
        return games
    
    def update_game_list_in_redis(self, game_list=None, cache_duration=300):
        """Update game list in Redis cache"""
        try:
            if self.redis_client is None:
                print("❌ Redis client is not available")
                return None
                
            if game_list is None:
                game_list = self.generate_game_list()
            
            game_list_data = {
                "gamelist": game_list,
                "updated_at": datetime.now().isoformat(),
                "cache_duration": f"{cache_duration} seconds",
                "total_games": len(game_list),
                "active_games": len([g for g in game_list if g.get("status") == "active"]),
                "total_players": sum(g.get("active_gamepalye", 0) for g in game_list)
            }
            
            # Cache in Redis
            self.redis_client.setex(
                "gamelist:current", 
                cache_duration, 
                json.dumps(game_list_data, indent=2)
            )
            
            print(f"✅ Game list updated in Redis:")
            print(f"   📊 Total games: {len(game_list)}")
            print(f"   🎮 Active games: {game_list_data['active_games']}")
            print(f"   👥 Total players: {game_list_data['total_players']}")
            print(f"   ⏰ Cache duration: {cache_duration} seconds")
            
            return game_list_data
            
        except Exception as e:
            print(f"❌ Failed to update game list in Redis: {e}")
            return None
    
    def send_game_list_update_notification(self, game_list_data):
        """Send real-time notification to all connected Socket.IO clients"""
        try:
            if not self.sio.connected:
                print("⚠️ Socket.IO not connected, attempting to reconnect...")
                if not self.connect_to_socketio():
                    return False
            
            # Send a simple trigger event to the Go server
            # The Go server will fetch the latest data from Redis and broadcast it
            self.sio.emit("trigger_game_list_update", {
                "message": "Update game list from Redis",
                "timestamp": datetime.now().isoformat()
            })
            
            # Also send to gameplay namespace
            self.sio.emit("trigger_game_list_update", {
                "message": "Update game list from Redis",
                "timestamp": datetime.now().isoformat()
            }, namespace="/gameplay")
            
            print("📡 Trigger event sent to Go server - it will fetch from Redis and broadcast to all clients")
            return True
            
        except Exception as e:
            print(f"❌ Failed to send trigger event: {e}")
            return False
    
    def update_and_notify(self, num_games=None, cache_duration=300):
        """Update game list and send notification to all clients"""
        print("\n🔄 Updating game list and notifying clients...")
        print("=" * 60)
        
        # Update game list in Redis
        game_list_data = self.update_game_list_in_redis(num_games, cache_duration)
        
        if game_list_data:
            # Send notification to connected clients
            self.send_game_list_update_notification(game_list_data)
            return True
        else:
            print("❌ Failed to update game list")
            return False
    
    def continuous_updates(self, interval_seconds=30, max_updates=None):
        """Continuously update game list at specified intervals"""
        print(f"\n🔄 Starting continuous game list updates every {interval_seconds} seconds...")
        print("=" * 60)
        
        update_count = 0
        
        try:
            while True:
                if max_updates and update_count >= max_updates:
                    print(f"✅ Completed {max_updates} updates")
                    break
                
                update_count += 1
                print(f"\n📅 Update #{update_count} - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
                
                # Update and notify
                self.update_and_notify()
                
                # Wait for next update
                if max_updates is None or update_count < max_updates:
                    print(f"⏰ Waiting {interval_seconds} seconds for next update...")
                    time.sleep(interval_seconds)
                
        except KeyboardInterrupt:
            print("\n⏹️ Continuous updates stopped by user")
        except Exception as e:
            print(f"\n❌ Continuous updates failed: {e}")
        finally:
            self.disconnect_from_socketio()

def main():
    """Main function to run the game list updater"""
    print("🎮 Game List Updater")
    print("Updates Redis cache and sends real-time notifications to Socket.IO clients")
    print("=" * 60)
    
    # Initialize updater
    updater = GameListUpdater()
    
    # Test Redis connection
    if updater.redis_client is None:
        print("❌ Redis is not available")
        print("💡 Please install Redis: pip install redis==5.0.1")
        print("💡 Make sure Redis server is running on 127.0.0.1:6379")
        return
    
    try:
        updater.redis_client.ping()
        print("✅ Redis connection successful")
    except Exception as e:
        print(f"❌ Redis connection failed: {e}")
        print("💡 Make sure Redis is running on 127.0.0.1:6379")
        return
    
    # Connect to Socket.IO
    if not updater.connect_to_socketio():
        print("⚠️ Socket.IO connection failed, but Redis updates will still work")
    
    # Menu for different operations
    while True:
        print("\n📋 Choose an operation:")
        print("1. Single update (update once)")
        print("2. Continuous updates (update every 30 seconds)")
        print("3. Custom continuous updates")
        print("4. Test Redis connection")
        print("5. Exit")
        
        choice = input("\nEnter your choice (1-5): ").strip()
        
        if choice == "1":
            # Single update
            num_games = input("Number of games to generate (press Enter for random): ").strip()
            if num_games:
                try:
                    num_games = int(num_games)
                    updater.update_and_notify(num_games)
                except ValueError:
                    print("❌ Invalid number, using random count")
                    updater.update_and_notify()
            else:
                updater.update_and_notify()
                
        elif choice == "2":
            # Continuous updates every 30 seconds
            max_updates = input("Maximum number of updates (press Enter for unlimited): ").strip()
            if max_updates:
                try:
                    max_updates = int(max_updates)
                    updater.continuous_updates(30, max_updates)
                except ValueError:
                    print("❌ Invalid number, using unlimited updates")
                    updater.continuous_updates(30)
            else:
                updater.continuous_updates(30)
                
        elif choice == "3":
            # Custom continuous updates
            try:
                interval = int(input("Update interval in seconds: "))
                max_updates = input("Maximum number of updates (press Enter for unlimited): ").strip()
                if max_updates:
                    max_updates = int(max_updates)
                    updater.continuous_updates(interval, max_updates)
                else:
                    updater.continuous_updates(interval)
            except ValueError:
                print("❌ Invalid input")
                
        elif choice == "4":
            # Test Redis connection
            try:
                updater.redis_client.ping()
                print("✅ Redis connection is working")
                
                # Test getting current game list
                current_data = updater.redis_client.get("gamelist:current")
                if current_data:
                    parsed_data = json.loads(current_data)
                    print(f"📊 Current game list in Redis: {len(parsed_data.get('gamelist', []))} games")
                else:
                    print("📝 No game list found in Redis")
                    
            except Exception as e:
                print(f"❌ Redis test failed: {e}")
                
        elif choice == "5":
            # Exit
            print("👋 Goodbye!")
            break
            
        else:
            print("❌ Invalid choice, please try again")
    
    # Cleanup
    updater.disconnect_from_socketio()

if __name__ == "__main__":
    main() 