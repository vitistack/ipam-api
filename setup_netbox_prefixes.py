#!/usr/bin/env python3
"""
NetBox Prefix Containers and Custom Fields Setup Script
This script sets up NetBox with three prefix containers and custom fields
"""

import requests
import json
import sys
import urllib3

# Disable SSL warnings for local development
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# NetBox configuration
NETBOX_URL = "http://localhost:8000"
ADMIN_TOKEN = "0123456789abcdef0123456789abcdef01234567"  # Default admin token from docker-compose

def make_api_request(method, endpoint, data=None):
    """Make an API request to NetBox"""
    headers = {
        "Authorization": f"Token {ADMIN_TOKEN}",
        "Content-Type": "application/json",
        "Accept": "application/json"
    }
    
    url = f"{NETBOX_URL}/api{endpoint}"
    
    try:
        if method.upper() == "GET":
            response = requests.get(url, headers=headers, verify=False)
        elif method.upper() == "POST":
            response = requests.post(url, headers=headers, json=data, verify=False)
        elif method.upper() == "PUT":
            response = requests.put(url, headers=headers, json=data, verify=False)
        elif method.upper() == "PATCH":
            response = requests.patch(url, headers=headers, json=data, verify=False)
        else:
            raise ValueError(f"Unsupported HTTP method: {method}")
        
        return response
    except requests.exceptions.ConnectionError:
        print("❌ Cannot connect to NetBox. Make sure it's running on http://localhost:8000")
        return None
    except Exception as e:
        print(f"❌ API request failed: {str(e)}")
        return None

def create_custom_fields():
    """Create custom fields for prefix containers"""
    print("🔧 Creating custom fields...")
    
    # Custom field for k8s_uuid
    k8s_uuid_field = {
        "name": "k8s_uuid",
        "label": "Kubernetes UUID",
        "type": "text",
        "object_types": ["ipam.prefix"],
        "description": "Kubernetes UUID for the prefix",
        "required": False
    }
    
    # First create a choice set for k8s_zone
    k8s_zone_choice_set = {
        "name": "k8s_zones",
        "description": "Kubernetes zone choices",
        "extra_choices": [
            ["inet", "Internet"],
            ["hnet-private", "Helsenett Private"],
            ["hnet-public", "Helsenett Public"]
        ]
    }
    
    # Create choice set first
    print("  Creating k8s_zone choice set...")
    response = make_api_request("POST", "/extras/custom-field-choice-sets/", k8s_zone_choice_set)
    choice_set_id = None
    if response and response.status_code == 201:
        print("  ✅ k8s_zone choice set created successfully")
        choice_set_id = response.json()["id"]
    elif response and response.status_code == 400:
        print("  ℹ️  k8s_zone choice set might already exist")
        print(f"     Response: {response.text}")
        # Try to get existing choice set
        existing_response = make_api_request("GET", "/extras/custom-field-choice-sets/?name=k8s_zones")
        if existing_response and existing_response.status_code == 200:
            results = existing_response.json()["results"]
            if results:
                choice_set_id = results[0]["id"]
                print(f"     Using existing choice set ID: {choice_set_id}")
    else:
        print(f"  ❌ Failed to create k8s_zone choice set")
        if response:
            print(f"     Status: {response.status_code}, Response: {response.text}")
    
    # Custom field for k8s_zone with choice set
    k8s_zone_field = {
        "name": "k8s_zone",
        "label": "Kubernetes Zone", 
        "type": "select",
        "object_types": ["ipam.prefix"],
        "description": "Kubernetes zone classification",
        "required": False,
        "choice_set": choice_set_id
    }
    
    # Create k8s_uuid field
    print("  Creating k8s_uuid custom field...")
    response = make_api_request("POST", "/extras/custom-fields/", k8s_uuid_field)
    if response and response.status_code == 201:
        print("  ✅ k8s_uuid custom field created successfully")
    elif response and response.status_code == 400:
        # Field might already exist
        print("  ℹ️  k8s_uuid custom field might already exist")
        print(f"     Response: {response.text}")
    else:
        print(f"  ❌ Failed to create k8s_uuid custom field")
        if response:
            print(f"     Status: {response.status_code}, Response: {response.text}")
    
    # Create k8s_zone field (only if we have a choice set)
    if choice_set_id:
        print("  Creating k8s_zone custom field...")
        response = make_api_request("POST", "/extras/custom-fields/", k8s_zone_field)
        if response and response.status_code == 201:
            print("  ✅ k8s_zone custom field created successfully")
        elif response and response.status_code == 400:
            print("  ℹ️  k8s_zone custom field might already exist")
            print(f"     Response: {response.text}")
        else:
            print(f"  ❌ Failed to create k8s_zone custom field")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")
    else:
        print("  ⚠️  Skipping k8s_zone custom field creation (no choice set)")

def create_prefix_containers():
    """Create the three prefix containers"""
    print("🏗️  Creating prefix containers...")
    
    # Define the three prefix containers
    prefix_containers = [
        {
            "prefix": "10.0.0.0/8",
            "description": "Internet prefix container",
            "is_pool": True,
            "custom_fields": {
                "k8s_zone": "inet"
            }
        },
        {
            "prefix": "172.16.0.0/12",
            "description": "Helsenett Private prefix container", 
            "is_pool": True,
            "custom_fields": {
                "k8s_zone": "hnet-private"
            }
        },
        {
            "prefix": "192.168.0.0/16",
            "description": "Helsenett Public prefix container",
            "is_pool": True,
            "custom_fields": {
                "k8s_zone": "hnet-public"
            }
        }
    ]
    
    created_prefixes = []
    
    for container in prefix_containers:
        print(f"  Creating prefix container: {container['prefix']} ({container['description']})")
        
        # Check if prefix already exists
        response = make_api_request("GET", f"/ipam/prefixes/?prefix={container['prefix']}")
        if response and response.status_code == 200:
            existing = response.json()["results"]
            if existing:
                print(f"  ℹ️  Prefix {container['prefix']} already exists, updating...")
                # Update existing prefix
                prefix_id = existing[0]["id"]
                response = make_api_request("PATCH", f"/ipam/prefixes/{prefix_id}/", container)
                if response and response.status_code == 200:
                    print(f"  ✅ Updated prefix container: {container['prefix']}")
                    created_prefixes.append(response.json())
                else:
                    print(f"  ❌ Failed to update prefix {container['prefix']}")
                    if response:
                        print(f"     Status: {response.status_code}, Response: {response.text}")
            else:
                # Create new prefix
                response = make_api_request("POST", "/ipam/prefixes/", container)
                if response and response.status_code == 201:
                    print(f"  ✅ Created prefix container: {container['prefix']}")
                    created_prefixes.append(response.json())
                else:
                    print(f"  ❌ Failed to create prefix {container['prefix']}")
                    if response:
                        print(f"     Status: {response.status_code}, Response: {response.text}")
        else:
            print(f"  ❌ Failed to check existing prefixes")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")
    
    return created_prefixes

def update_config_file(prefixes):
    """Update the configuration file with the new prefix containers"""
    print("📝 Updating configuration file...")
    
    try:
        config_path = '/Users/rogerwesterbo/dev/github/viti/ipam-api/config-docker-compose.json'
        with open(config_path, 'r') as f:
            config = json.load(f)
        
        # Ensure netbox section exists
        if 'netbox' not in config:
            config['netbox'] = {}
        
        # Update prefix containers
        config['netbox']['prefix_containers'] = {
            "internet": "10.0.0.0/8",
            "helsenett_private": "172.16.0.0/12", 
            "helsenett_public": "192.168.0.0/16"
        }
        
        with open(config_path, 'w') as f:
            json.dump(config, f, indent=2)
        
        print("✅ Configuration file updated successfully!")
        
    except Exception as e:
        print(f"❌ Failed to update configuration file: {str(e)}")

def verify_setup():
    """Verify that everything was set up correctly"""
    print("🔍 Verifying setup...")
    
    # Check custom fields
    response = make_api_request("GET", "/extras/custom-fields/")
    if response and response.status_code == 200:
        fields = response.json()["results"]
        k8s_fields = [f for f in fields if f["name"] in ["k8s_uuid", "k8s_zone"]]
        print(f"  Custom fields found: {len(k8s_fields)}/2")
        for field in k8s_fields:
            print(f"    - {field['name']}: {field['label']}")
    
    # Check prefixes
    response = make_api_request("GET", "/ipam/prefixes/?is_pool=true")
    if response and response.status_code == 200:
        prefixes = response.json()["results"]
        container_prefixes = [p for p in prefixes if p["prefix"] in ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]]
        print(f"  Prefix containers found: {len(container_prefixes)}/3")
        for prefix in container_prefixes:
            zone = prefix.get("custom_fields", {}).get("k8s_zone", "not set")
            print(f"    - {prefix['prefix']}: k8s_zone={zone}")

def main():
    print("🚀 NetBox Prefix Containers and Custom Fields Setup")
    print("=" * 60)
    
    # Test connection first
    print("🧪 Testing NetBox connection...")
    response = make_api_request("GET", "/")
    if not response or response.status_code != 200:
        print("❌ Cannot connect to NetBox API. Please check:")
        print("  - NetBox is running (docker-compose up)")
        print("  - NetBox is accessible at http://localhost:8000")
        print("  - Admin token is correct")
        return
    
    print("✅ NetBox connection successful!")
    print()
    
    # Create custom fields
    create_custom_fields()
    print()
    
    # Create prefix containers
    prefixes = create_prefix_containers()
    print()
    
    # Update config file
    update_config_file(prefixes)
    print()
    
    # Verify setup
    verify_setup()
    print()
    
    print("=" * 60)
    print("🎉 Setup completed!")
    print("=" * 60)
    print("Your NetBox now has:")
    print("✅ Custom fields: k8s_uuid (text) and k8s_zone (select)")
    print("✅ Three prefix containers:")
    print("   - 10.0.0.0/8 (inet)")
    print("   - 172.16.0.0/12 (hnet-private)")
    print("   - 192.168.0.0/16 (hnet-public)")
    print("✅ Updated configuration file")
    print()
    print("💡 Access NetBox at: http://localhost:8000")
    print("   Username: admin")
    print("   Password: admin")

if __name__ == "__main__":
    main()
