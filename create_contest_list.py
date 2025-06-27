#!/usr/bin/env python3
"""
Contest List Data Generator for Redis
This script creates contest list data and stores it in Redis under 'listcontest:current'
"""

import redis
import json
import time
from datetime import datetime, timedelta
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class ContestListGenerator:
    def __init__(self, redis_host='localhost', redis_port=6379, redis_db=0):
        """Initialize Redis connection and contest list generator"""
        self.redis_client = redis.Redis(
            host=redis_host,
            port=redis_port,
            db=redis_db,
            decode_responses=True
        )
        
    def test_redis_connection(self):
        """Test Redis connection"""
        try:
            self.redis_client.ping()
            logger.info("✅ Redis connection successful")
            return True
        except redis.ConnectionError as e:
            logger.error(f"❌ Redis connection failed: {e}")
            return False
    
    def generate_contest_list_data(self):
        """Generate sample contest list data"""
        current_time = datetime.utcnow()
        
        contest_list = [
            {
                "contestId": 1,
                "contestName": "Weekly Algorithm Challenge",
                "description": "Solve complex algorithmic problems and compete with top programmers worldwide",
                "startTime": (current_time + timedelta(days=7)).isoformat() + "Z",
                "endTime": (current_time + timedelta(days=14)).isoformat() + "Z",
                "status": "upcoming",
                "maxParticipants": 1000,
                "currentParticipants": 847,
                "categories": ["algorithms", "data-structures", "dynamic-programming"],
                "difficulty": "medium",
                "duration": "7 days",
                "reward": {
                    "currency": "USD",
                    "amount": 5000,
                    "distribution": {
                        "1st": 2500,
                        "2nd": 1500,
                        "3rd": 1000
                    }
                },
                "rules": [
                    "Individual participation only",
                    "Multiple programming languages allowed",
                    "Plagiarism will result in disqualification"
                ],
                "tags": ["competitive", "weekly", "algorithm"],
                "sponsor": "TechCorp Inc.",
                "registrationDeadline": (current_time + timedelta(days=6)).isoformat() + "Z"
            },
            {
                "contestId": 2,
                "contestName": "Data Science Hackathon",
                "description": "Build innovative data science solutions using real-world datasets",
                "startTime": (current_time + timedelta(days=2)).isoformat() + "Z",
                "endTime": (current_time + timedelta(days=7)).isoformat() + "Z",
                "status": "active",
                "maxParticipants": 500,
                "currentParticipants": 423,
                "categories": ["machine-learning", "data-analysis", "python", "statistics"],
                "difficulty": "hard",
                "duration": "5 days",
                "reward": {
                    "currency": "USD",
                    "amount": 10000,
                    "distribution": {
                        "1st": 5000,
                        "2nd": 3000,
                        "3rd": 2000
                    }
                },
                "rules": [
                    "Team participation (2-4 members)",
                    "Open source libraries allowed",
                    "Final presentation required"
                ],
                "tags": ["hackathon", "data-science", "ml"],
                "sponsor": "DataTech Solutions",
                "registrationDeadline": (current_time + timedelta(days=1)).isoformat() + "Z",
                "liveStats": {
                    "submissions": 156,
                    "teams": 89,
                    "avgScore": 78.5
                }
            },
            {
                "contestId": 3,
                "contestName": "Frontend Development Sprint",
                "description": "Create responsive and modern web applications using latest technologies",
                "startTime": (current_time - timedelta(days=5)).isoformat() + "Z",
                "endTime": (current_time - timedelta(days=2)).isoformat() + "Z",
                "status": "completed",
                "maxParticipants": 300,
                "currentParticipants": 298,
                "categories": ["frontend", "react", "javascript", "css", "ui-ux"],
                "difficulty": "easy",
                "duration": "3 days",
                "reward": {
                    "currency": "USD",
                    "amount": 3000,
                    "distribution": {
                        "1st": 1500,
                        "2nd": 900,
                        "3rd": 600
                    }
                },
                "rules": [
                    "Individual or team (max 2)",
                    "React/Vue/Angular required",
                    "Mobile responsive design"
                ],
                "tags": ["frontend", "sprint", "web-development"],
                "sponsor": "WebDev Pro",
                "registrationDeadline": (current_time - timedelta(days=6)).isoformat() + "Z",
                "results": {
                    "winner": "Team CodeCrafters",
                    "runnerUp": "SoloDev Master",
                    "thirdPlace": "UI Wizards",
                    "totalSubmissions": 245
                }
            },
            {
                "contestId": 4,
                "contestName": "Mobile App Innovation",
                "description": "Develop innovative mobile applications that solve real-world problems",
                "startTime": (current_time + timedelta(days=20)).isoformat() + "Z",
                "endTime": (current_time + timedelta(days=30)).isoformat() + "Z",
                "status": "upcoming",
                "maxParticipants": 800,
                "currentParticipants": 156,
                "categories": ["mobile-development", "ios", "android", "flutter", "react-native"],
                "difficulty": "medium",
                "duration": "10 days",
                "reward": {
                    "currency": "USD",
                    "amount": 8000,
                    "distribution": {
                        "1st": 4000,
                        "2nd": 2400,
                        "3rd": 1600
                    }
                },
                "rules": [
                    "Cross-platform development preferred",
                    "App must be functional and deployable",
                    "Documentation and demo video required"
                ],
                "tags": ["mobile", "innovation", "app-development"],
                "sponsor": "MobileTech Ventures",
                "registrationDeadline": (current_time + timedelta(days=19)).isoformat() + "Z",
                "specialFeatures": [
                    "App Store deployment opportunity",
                    "VC pitch session for top 3",
                    "Industry mentorship program"
                ]
            },
            {
                "contestId": 5,
                "contestName": "Blockchain Smart Contract Challenge",
                "description": "Build secure and efficient smart contracts for decentralized applications",
                "startTime": (current_time + timedelta(days=10)).isoformat() + "Z",
                "endTime": (current_time + timedelta(days=15)).isoformat() + "Z",
                "status": "upcoming",
                "maxParticipants": 400,
                "currentParticipants": 89,
                "categories": ["blockchain", "smart-contracts", "solidity", "ethereum", "defi"],
                "difficulty": "hard",
                "duration": "5 days",
                "reward": {
                    "currency": "ETH",
                    "amount": 5.0,
                    "distribution": {
                        "1st": 2.5,
                        "2nd": 1.5,
                        "3rd": 1.0
                    }
                },
                "rules": [
                    "Solidity programming required",
                    "Security audit mandatory",
                    "Gas optimization important"
                ],
                "tags": ["blockchain", "defi", "smart-contracts"],
                "sponsor": "CryptoChain Labs",
                "registrationDeadline": (current_time + timedelta(days=9)).isoformat() + "Z",
                "blockchainInfo": {
                    "network": "Ethereum Testnet",
                    "gasLimit": "3000000",
                    "compilerVersion": "0.8.19"
                }
            },
            {
                "contestId": 6,
                "contestName": "AI Chatbot Competition",
                "description": "Create intelligent chatbots using natural language processing and AI",
                "startTime": (current_time - timedelta(days=1)).isoformat() + "Z",
                "endTime": (current_time + timedelta(days=2)).isoformat() + "Z",
                "status": "active",
                "maxParticipants": 600,
                "currentParticipants": 567,
                "categories": ["artificial-intelligence", "nlp", "chatbot", "python", "api"],
                "difficulty": "medium",
                "duration": "3 days",
                "reward": {
                    "currency": "USD",
                    "amount": 4000,
                    "distribution": {
                        "1st": 2000,
                        "2nd": 1200,
                        "3rd": 800
                    }
                },
                "rules": [
                    "Must use AI/ML libraries",
                    "API integration required",
                    "Multi-language support bonus"
                ],
                "tags": ["ai", "chatbot", "nlp"],
                "sponsor": "AI Innovations Corp",
                "registrationDeadline": (current_time - timedelta(days=2)).isoformat() + "Z",
                "liveStats": {
                    "submissions": 234,
                    "avgAccuracy": 87.3,
                    "languagesSupported": 12
                }
            }
        ]
        
        # Create the data structure that matches the Go service format
        contest_data = {
            "gamelist": contest_list
        }
        
        return contest_data
    
    def store_contest_list_in_redis(self, data, expiration_minutes=5):
        """Store contest list data in Redis"""
        try:
            key = "listcontest:current"
            expiration_seconds = expiration_minutes * 60
            
            # Convert data to JSON string
            json_data = json.dumps(data, indent=2)
            
            # Store in Redis with expiration
            self.redis_client.setex(key, expiration_seconds, json_data)
            
            logger.info(f"✅ Contest list data stored in Redis with key: {key}")
            logger.info(f"⏰ Expiration: {expiration_minutes} minutes")
            logger.info(f"📊 Contests stored: {len(data['gamelist'])}")
            
            return True
            
        except Exception as e:
            logger.error(f"❌ Failed to store contest list in Redis: {e}")
            return False
    
    def retrieve_contest_list_from_redis(self):
        """Retrieve contest list data from Redis"""
        try:
            key = "listcontest:current"
            data = self.redis_client.get(key)
            
            if data:
                contest_data = json.loads(data)
                logger.info(f"✅ Contest list data retrieved from Redis")
                logger.info(f"📊 Contests found: {len(contest_data['gamelist'])}")
                return contest_data
            else:
                logger.warning("⚠️ No contest list data found in Redis")
                return None
                
        except Exception as e:
            logger.error(f"❌ Failed to retrieve contest list from Redis: {e}")
            return None
    
    def display_contest_list(self, data):
        """Display contest list data in a formatted way"""
        if not data or 'gamelist' not in data:
            logger.warning("⚠️ No contest data to display")
            return
        
        contests = data['gamelist']
        logger.info(f"\n{'='*80}")
        logger.info(f"🏆 CONTEST LIST DATA ({len(contests)} contests)")
        logger.info(f"{'='*80}")
        
        for i, contest in enumerate(contests, 1):
            logger.info(f"\n{i}. {contest['contestName']}")
            logger.info(f"   ID: {contest['contestId']}")
            logger.info(f"   Status: {contest['status']}")
            logger.info(f"   Difficulty: {contest['difficulty']}")
            logger.info(f"   Duration: {contest['duration']}")
            logger.info(f"   Participants: {contest['currentParticipants']}/{contest['maxParticipants']}")
            logger.info(f"   Reward: {contest['reward']['currency']} {contest['reward']['amount']}")
            logger.info(f"   Start: {contest['startTime']}")
            logger.info(f"   End: {contest['endTime']}")
            logger.info(f"   Categories: {', '.join(contest['categories'])}")
            
            if 'liveStats' in contest:
                logger.info(f"   Live Stats: {contest['liveStats']}")
            
            if 'results' in contest:
                logger.info(f"   Results: {contest['results']}")
    
    def run(self):
        """Main execution method"""
        logger.info("🚀 Starting Contest List Generator")
        
        # Test Redis connection
        if not self.test_redis_connection():
            return False
        
        # Generate contest list data
        logger.info("📝 Generating contest list data...")
        contest_data = self.generate_contest_list_data()
        
        # Store in Redis
        logger.info("💾 Storing contest list data in Redis...")
        if not self.store_contest_list_in_redis(contest_data):
            return False
        
        # Retrieve and display for verification
        logger.info("📖 Retrieving contest list data from Redis for verification...")
        retrieved_data = self.retrieve_contest_list_from_redis()
        
        if retrieved_data:
            self.display_contest_list(retrieved_data)
            logger.info("✅ Contest list data successfully created and stored!")
            return True
        else:
            logger.error("❌ Failed to verify contest list data")
            return False

def main():
    """Main function"""
    print("🏆 Contest List Generator for Redis")
    print("=" * 50)
    
    # Create generator instance
    generator = ContestListGenerator()
    
    # Run the generator
    success = generator.run()
    
    if success:
        print("\n✅ Contest list data successfully created!")
        print("🔑 Redis Key: listcontest:current")
        print("⏰ Expiration: 5 minutes")
        print("📊 Data Structure: {'gamelist': [...]}")
    else:
        print("\n❌ Failed to create contest list data")
        return 1
    
    return 0

if __name__ == "__main__":
    exit(main()) 