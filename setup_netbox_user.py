#!/usr/bin/env python3
"""
NetBox User and Token Setup Script
This script helps create a new user and API token in NetBox
"""

import requests
import json
import sys
from urllib3.packages.urllib3.exceptions import InsecureRequestWarning

# Disable SSL warnings for local development
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

# NetBox configuration
NETBOX_URL = "http://localhost:8000"
ADMIN_TOKEN = "0123456789abcdef0123456789abcdef01234567"  # Default admin token from docker-compose

def create_user_and_token(username, email, password, first_name="", last_name=""):
    """Create a new user and generate an API token"""
    
    headers = {
        "Authorization": f"Token {ADMIN_TOKEN}",
        "Content-Type": "application/json",
        "Accept": "application/json"
    }
    
    try:
        # First, create the user
        user_data = {
            "username": username,
            "email": email,
            "first_name": first_name,
            "last_name": last_name,
            "is_active": True,
            "is_staff": False,
            "is_superuser": False
        }
        
        # Set password if provided
        if password:
            user_data["password"] = password
        
        print(f"Creating user '{username}'...")
        user_response = requests.post(
            f"{NETBOX_URL}/api/users/users/",
            headers=headers,
            json=user_data,
            verify=False
        )
        
        if user_response.status_code == 201:
            user = user_response.json()
            user_id = user["id"]
            print(f"✅ User '{username}' created successfully (ID: {user_id})")
        elif user_response.status_code == 400:
            # User might already exist, try to get existing user
            print(f"User '{username}' might already exist, trying to find existing user...")
            users_response = requests.get(
                f"{NETBOX_URL}/api/users/users/?username={username}",
                headers=headers,
                verify=False
            )
            if users_response.status_code == 200:
                users = users_response.json()["results"]
                if users:
                    user_id = users[0]["id"]
                    print(f"✅ Found existing user '{username}' (ID: {user_id})")
                else:
                    print(f"❌ Failed to create or find user '{username}'")
                    print(f"Response: {user_response.text}")
                    return None
            else:
                print(f"❌ Failed to create user '{username}'")
                print(f"Response: {user_response.text}")
                return None
        else:
            print(f"❌ Failed to create user '{username}'")
            print(f"Status: {user_response.status_code}")
            print(f"Response: {user_response.text}")
            return None
        
        # Create API token for the user
        print(f"Creating API token for user '{username}'...")
        token_data = {
            "user": user_id,
            "description": f"API token for {username}",
            "write_enabled": True
        }
        
        token_response = requests.post(
            f"{NETBOX_URL}/api/users/tokens/",
            headers=headers,
            json=token_data,
            verify=False
        )
        
        if token_response.status_code == 201:
            token = token_response.json()
            print(f"✅ API token created successfully!")
            print(f"🔑 Token: {token['key']}")
            print(f"📝 Description: {token['description']}")
            return {
                "username": username,
                "user_id": user_id,
                "token": token['key'],
                "token_id": token['id']
            }
        else:
            print(f"❌ Failed to create API token")
            print(f"Status: {token_response.status_code}")
            print(f"Response: {token_response.text}")
            return None
            
    except requests.exceptions.ConnectionError:
        print("❌ Cannot connect to NetBox. Make sure it's running on http://localhost:8000")
        return None
    except Exception as e:
        print(f"❌ An error occurred: {str(e)}")
        return None

def test_token(token):
    """Test if the token works by making a simple API call"""
    headers = {
        "Authorization": f"Token {token}",
        "Accept": "application/json"
    }
    
    try:
        response = requests.get(f"{NETBOX_URL}/api/", headers=headers, verify=False)
        if response.status_code == 200:
            print("✅ Token test successful! The token is working correctly.")
            return True
        else:
            print(f"❌ Token test failed. Status: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Token test failed: {str(e)}")
        return False

def main():
    print("🚀 NetBox User and Token Setup")
    print("=" * 40)
    
    # Get user input
    username = input("Enter username: ").strip()
    if not username:
        print("❌ Username is required")
        return
    
    email = input("Enter email: ").strip()
    if not email:
        print("❌ Email is required")
        return
    
    first_name = input("Enter first name (optional): ").strip()
    last_name = input("Enter last name (optional): ").strip()
    password = input("Enter password (optional, will be auto-generated if empty): ").strip()
    
    # Create user and token
    result = create_user_and_token(username, email, password, first_name, last_name)
    
    if result:
        print("\n" + "=" * 40)
        print("🎉 Setup completed successfully!")
        print("=" * 40)
        print(f"Username: {result['username']}")
        print(f"User ID: {result['user_id']}")
        print(f"API Token: {result['token']}")
        print("\n📋 Save this token - you won't be able to see it again!")
        
        # Test the token
        print("\n🧪 Testing the token...")
        test_token(result['token'])
        
        # Show how to use it
        print(f"\n💡 Usage examples:")
        print(f"curl -H 'Authorization: Token {result['token']}' {NETBOX_URL}/api/")
        print(f"\nOr update your config-docker-compose.json:")
        print(f'  "token": "{result['token']}"')
        
        # Ask if they want to update the config file
        update_config = input("\nWould you like to update config-docker-compose.json with this token? (y/N): ").strip().lower()
        if update_config == 'y':
            try:
                with open('/Users/rogerwesterbo/dev/github/viti/ipam-api/config-docker-compose.json', 'r') as f:
                    config = json.load(f)
                
                config['netbox']['token'] = result['token']
                
                with open('/Users/rogerwesterbo/dev/github/viti/ipam-api/config-docker-compose.json', 'w') as f:
                    json.dump(config, f, indent=2)
                
                print("✅ Config file updated successfully!")
            except Exception as e:
                print(f"❌ Failed to update config file: {str(e)}")
    else:
        print("❌ Setup failed")

if __name__ == "__main__":
    main()
