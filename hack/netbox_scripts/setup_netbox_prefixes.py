#!/usr/bin/env python3
"""
NetBox Prefix Containers and Custom Fields Setup Script
This script sets up NetBox with three prefix containers and custom fields
"""

import urllib.request
import urllib.parse
import urllib.error
import json
import sys
import ssl

# NetBox configuration
NETBOX_URL = "http://localhost:8000"
ADMIN_TOKEN = "0123456789abcdef0123456789abcdef01234567"  # Default admin token from docker-compose


def make_api_request(method, endpoint, data=None):
    """Make an API request to NetBox"""
    headers = {
        "Authorization": f"Token {ADMIN_TOKEN}",
        "Content-Type": "application/json",
        "Accept": "application/json",
    }

    url = f"{NETBOX_URL}/api{endpoint}"

    try:
        # Create SSL context that doesn't verify certificates (for local development)
        ssl_context = ssl.create_default_context()
        ssl_context.check_hostname = False
        ssl_context.verify_mode = ssl.CERT_NONE

        # Prepare request data
        request_data = None
        if data is not None:
            request_data = json.dumps(data).encode("utf-8")

        # Create request object
        req = urllib.request.Request(
            url, data=request_data, headers=headers, method=method.upper()
        )

        # Make the request
        with urllib.request.urlopen(req, context=ssl_context) as response:
            response_data = response.read().decode("utf-8")

            # Create a response-like object for compatibility
            class Response:
                def __init__(self, status_code, text, data):
                    self.status_code = status_code
                    self.text = text
                    self._json_data = data

                def json(self):
                    return self._json_data

            # Parse JSON response if possible
            try:
                json_data = json.loads(response_data) if response_data else {}
            except json.JSONDecodeError:
                json_data = {}

            return Response(response.status, response_data, json_data)

    except urllib.error.HTTPError as e:
        # Handle HTTP errors (4xx, 5xx)
        error_data = e.read().decode("utf-8") if e.fp else ""
        try:
            json_data = json.loads(error_data) if error_data else {}
        except json.JSONDecodeError:
            json_data = {}

        class ErrorResponse:
            def __init__(self, status_code, text, data):
                self.status_code = status_code
                self.text = text
                self._json_data = data

            def json(self):
                return self._json_data

        return ErrorResponse(e.code, error_data, json_data)

    except urllib.error.URLError as e:
        print(
            "❌ Cannot connect to NetBox. Make sure it's running on http://localhost:8000"
        )
        return None
    except Exception as e:
        print(f"❌ API request failed: {str(e)}")
        return None


def create_tenant_groups():
    """Create tenant groups"""
    print("🔧 Creating tenant groups...")

    tenant_groups = [
        {
            "display": "DCN",
            "name": "DCN",
            "slug": "dcn",
            "parent": None,
            "description": "Datacenter Network",
            "tags": [],
            "custom_fields": {},
        }
    ]

    created_tenant_groups = []

    for tenant_group in tenant_groups:
        print(f"  Creating tenant group: {tenant_group['name']}")

        # Check if prefix already exists
        response = make_api_request(
            "GET", f"/tenancy/tenant-groups/?name={tenant_group['name']}"
        )
        if response and response.status_code == 200:
            existing = response.json()["results"]
            if existing:
                print(
                    f"  ℹ️  Tenant Group {tenant_group['name']} already exists, updating..."
                )
                # Update existing prefix
                tenant_group_id = existing[0]["id"]
                response = make_api_request(
                    "PATCH", f"/tenancy/tenant-groups/{tenant_group_id}/", tenant_group
                )
                if response and response.status_code == 200:
                    print(f"  ✅ Updated tenant group: {tenant_group['name']}")
                    created_tenant_groups.append(response.json())
                else:
                    print(f"  ❌ Failed to create tenant group {tenant_group['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
            else:
                # Create new tenant group
                response = make_api_request(
                    "POST", "/tenancy/tenant-groups/", tenant_group
                )
                if response and response.status_code == 201:
                    print(f"  ✅ Created tenant group : {tenant_group['name']}")
                    created_tenant_groups.append(response.json())
                else:
                    print(f"  ❌ Failed to create tenant group {tenant_group['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
        else:
            print(f"  ❌ Failed to check existing tenant groups")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")

    return created_tenant_groups


def create_tenants(tenant_groups):

    tenant_group_id = next(
        (tg["id"] for tg in tenant_groups if tg["name"] == "DCN"), None
    )

    """Create tenants"""
    print("🔧 Creating tenants...")

    tenants = [
        {
            "display": "NHN",
            "name": "NHN",
            "slug": "NHN",
            "group": tenant_group_id,
            "description": "",
            "comments": "",
            "tags": [],
            "custom_fields": {},
        }
    ]

    created_tenants = []

    for tenant in tenants:
        print(f"  Creating tenant: {tenant['name']}")

        # Check if prefix already exists
        response = make_api_request("GET", f"/tenancy/tenants/?name={tenant['name']}")
        if response and response.status_code == 200:
            existing = response.json()["results"]
            if existing:
                print(f"  ℹ️  Tenant {tenant['name']} already exists, updating...")
                # Update existing prefix
                tenant_group_id = existing[0]["id"]
                response = make_api_request(
                    "PATCH", f"/tenancy/tenants/{tenant_group_id}/", tenant
                )
                if response and response.status_code == 200:
                    print(f"  ✅ Updated tenant : {tenant['name']}")
                    created_tenants.append(response.json())
                else:
                    print(f"  ❌ Failed to create tenant {tenant['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
            else:
                # Create new tenant group
                response = make_api_request("POST", "/tenancy/tenants/", tenant)
                if response and response.status_code == 201:
                    print(f"  ✅ Created tenant : {tenant['name']}")
                    created_tenants.append(response.json())
                else:
                    print(f"  ❌ Failed to create tenant {tenant['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
        else:
            print(f"  ❌ Failed to check existing tenant")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")

    return created_tenants


def create_vrfs():
    """Create custom fields for prefix containers"""
    print("🔧 Creating custom fields...")

    vrfs = [
        {
            "display": "nhc (nhc)",
            "name": "nhc",
        }
    ]

    created_vrfs = []

    for vrf in vrfs:
        print(f"  Creating prefix container: {vrf['name']}")

        # Check if prefix already exists
        response = make_api_request("GET", f"/ipam/vrfs/?name={vrf['name']}")
        if response and response.status_code == 200:
            existing = response.json()["results"]
            if existing:
                print(f"  ℹ️  Vrf {vrf['name']} already exists, updating...")
                # Update existing prefix
                prefix_id = existing[0]["id"]
                response = make_api_request("PATCH", f"/ipam/vrfs/{prefix_id}/", vrf)
                if response and response.status_code == 200:
                    print(f"  ✅ Updated vrf: {vrf['name']}")
                    created_vrfs.append(response.json())
                else:
                    print(f"  ❌ Failed to vrf {vrf['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
            else:
                # Create new vrf
                response = make_api_request("POST", "/ipam/vrfs/", vrf)
                if response and response.status_code == 201:
                    print(f"  ✅ Created vrf : {vrf['name']}")
                    created_vrfs.append(response.json())
                else:
                    print(f"  ❌ Failed to create vrf {vrf['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
        else:
            print(f"  ❌ Failed to check existing vrfs")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")

    return created_vrfs


def create_roles():
    """Create roles"""
    print("🔧 Creating roles...")

    roles = [{"name": "datacenter", "slug": "datacenter"}]

    created_roles = []

    for role in roles:
        print(f"  Creating role: {role['name']}")

        # Check if role already exists
        response = make_api_request("GET", f"/ipam/roles/?name={role['name']}")
        if response and response.status_code == 200:
            existing = response.json().get("results", [])
            if existing:
                print(f"  ℹ️  Role {role['name']} already exists, updating...")
                role_id = existing[0]["id"]
                response = make_api_request("PATCH", f"/ipam/roles/{role_id}/", role)
                if response and response.status_code == 200:
                    print(f"  ✅ Updated role: {role['name']}")
                    created_roles.append(response.json())
                else:
                    print(f"  ❌ Failed to update role {role['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
            else:
                # Create new role
                response = make_api_request("POST", "/ipam/roles/", role)
                if response and response.status_code == 201:
                    print(f"  ✅ Created role: {role['name']}")
                    created_roles.append(response.json())
                else:
                    print(f"  ❌ Failed to create role {role['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
        else:
            print(f"  ❌ Failed to check existing roles")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")

    return created_roles


def create_custom_fields():
    """Create custom_fields"""
    print("🔧 Creating custom_fields...")

    # "name": "k8s_uuid",
    # "label": "Kubernetes UUID",
    # "type": "text",
    # "object_types": ["ipam.prefix"],
    # "description": "Kubernetes UUID for the prefix",
    # "required": False,

    custom_fields = [
        {
            "display": "Domain",
            "object_types": ["ipam.prefix"],
            "type": "text",
            "object_type": None,
            "data_type": "string",
            "name": "domain",
            "label": "Domain",
            "description": "",
            "required": True,
        },
        {
            "display": "Environment",
            "object_types": ["ipam.prefix"],
            "type": "text",
            "object_type": None,
            "data_type": "string",
            "name": "env",
            "label": "Environment",
            "description": "",
            "required": True,
        },
        {
            "display": "Infrastructure",
            "object_types": ["ipam.prefix"],
            "type": "text",
            "object_type": None,
            "data_type": "string",
            "name": "infra",
            "label": "Infrastructure",
            "description": "",
            "required": True,
        },
        {
            "display": "Purpose",
            "object_types": ["ipam.prefix"],
            "type": "text",
            "object_type": None,
            "data_type": "string",
            "name": "purpose",
            "label": "Purpose",
            "description": "",
            "required": True,
        },
    ]

    created_custom_fields = []

    for field in custom_fields:
        print(f"  Creating Custom field: {field['name']}")

        # Check if custom field already exists
        response = make_api_request(
            "GET", f"/extras/custom-fields/?name={field['name']}"
        )
        if response and response.status_code == 200:
            existing = response.json().get("results", [])
            if existing:
                print(f"  ℹ️  Custom field {field['name']} already exists, updating...")
                role_id = existing[0]["id"]
                response = make_api_request(
                    "PATCH", f"/extras/custom-fields/{role_id}/", field
                )
                if response and response.status_code == 200:
                    print(f"  ✅ Updated Custom field: {field['name']}")
                    created_custom_fields.append(response.json())
                else:
                    print(f"  ❌ Failed to update Custom field {field['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
            else:
                # Create new custom field
                response = make_api_request("POST", "/extras/custom-fields/", field)
                if response and response.status_code == 201:
                    print(f"  ✅ Created Custom field: {field['name']}")
                    created_custom_fields.append(response.json())
                else:
                    print(f"  ❌ Failed to create Custom field {field['name']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.json()}"
                        )
        else:
            print(f"  ❌ Failed to check existing custom fields")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")

    return created_custom_fields


def create_k8s_custom_fields():
    """Create k8s custom fields for prefix containers"""
    print("🔧 Creating k8s custom fields...")

    # Custom field for k8s_uuid
    k8s_uuid_field = {
        "name": "k8s_uuid",
        "label": "Kubernetes UUID",
        "type": "text",
        "object_types": ["ipam.prefix"],
        "description": "Kubernetes UUID for the prefix",
        "required": False,
    }

    # First create a choice set for k8s_zone
    k8s_zone_choice_set = {
        "name": "k8s_zone_choices",
        "description": "Kubernetes zone choices",
        "extra_choices": [
            ["inet", "Internet"],
            ["hnet-private", "Helsenett Private"],
            ["hnet-public", "Helsenett Public"],
        ],
    }

    # Create choice set first
    print("  Creating k8s_zone choice set...")
    response = make_api_request(
        "POST", "/extras/custom-field-choice-sets/", k8s_zone_choice_set
    )
    choice_set_id = None
    if response and response.status_code == 201:
        print("  ✅ k8s_zone choice set created successfully")
        choice_set_id = response.json()["id"]
    elif response and response.status_code == 400:
        print("  ℹ️  k8s_zone choice set might already exist")
        # print(f"     Response: {response.text}")
        # Try to get existing choice set
        existing_response = make_api_request(
            "GET", "/extras/custom-field-choice-sets/?name=k8s_zones"
        )
        if existing_response and existing_response.status_code == 200:
            results = existing_response.json()["results"]
            if results:
                choice_set_id = results[0]["id"]
                print(f"  ✅ Using existing choice set ID: {choice_set_id}")
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
        "choice_set": choice_set_id,
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


def create_prefix_containers(vrfs, roles, tenants):
    """Create the three prefix containers"""
    print("🏗️  Creating prefix containers...")

    # Define the three prefix containers

    role = next((role["id"] for role in roles if role["name"] == "datacenter"), None)
    vrf = next((vrf["id"] for vrf in vrfs if vrf["name"] == "nhc"), None)
    tenant = next((tenant["id"] for tenant in tenants if tenant["name"] == "NHN"), None)
    prefix_containers = [
        {
            "prefix": "10.0.0.0/8",
            "description": "Internet prefix container",
            "is_pool": False,
            "vrf": vrf,
            "role": role,
            "status": "container",
            "tenant": tenant,
            "custom_fields": {
                "k8s_zone": "inet",
                "domain": "na",
                "env": "na",
                "infra": "na",
                "purpose": "na",
            },
        },
        {
            "prefix": "172.16.0.0/12",
            "description": "Helsenett Private prefix container",
            "is_pool": False,
            "vrf": vrf,
            "role": role,
            "status": "container",
            "tenant": tenant,
            "custom_fields": {
                "k8s_zone": "hnet-private",
                "domain": "na",
                "env": "na",
                "infra": "na",
                "purpose": "na",
            },
        },
        {
            "prefix": "192.168.0.0/16",
            "description": "Helsenett Public prefix container",
            "is_pool": False,
            "vrf": vrf,
            "role": role,
            "status": "container",
            "tenant": tenant,
            "custom_fields": {
                "k8s_zone": "hnet-public",
                "domain": "na",
                "env": "na",
                "infra": "na",
                "purpose": "na",
            },
        },
    ]

    created_prefixes = []

    for container in prefix_containers:
        print(
            f"  Creating prefix container: {container['prefix']} ({container['description']})"
        )

        # Check if prefix already exists
        response = make_api_request(
            "GET", f"/ipam/prefixes/?prefix={container['prefix']}"
        )
        if response and response.status_code == 200:
            existing = response.json()["results"]
            if existing:
                print(f"  ℹ️  Prefix {container['prefix']} already exists, updating...")
                # Update existing prefix
                prefix_id = existing[0]["id"]
                response = make_api_request(
                    "PATCH", f"/ipam/prefixes/{prefix_id}/", container
                )
                if response and response.status_code == 200:
                    print(f"  ✅ Updated prefix container: {container['prefix']}")
                    created_prefixes.append(response.json())
                else:
                    print(f"  ❌ Failed to update prefix {container['prefix']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
            else:
                # Create new prefix
                response = make_api_request("POST", "/ipam/prefixes/", container)
                if response and response.status_code == 201:
                    print(f"  ✅ Created prefix container: {container['prefix']}")
                    created_prefixes.append(response.json())
                else:
                    print(f"  ❌ Failed to create prefix {container['prefix']}")
                    if response:
                        print(
                            f"     Status: {response.status_code}, Response: {response.text}"
                        )
        else:
            print(f"  ❌ Failed to check existing prefixes")
            if response:
                print(f"     Status: {response.status_code}, Response: {response.text}")

    return created_prefixes


# def update_config_file(prefixes):
#     """Update the configuration file with the new prefix containers"""
#     print("📝 Updating configuration file...")

#     try:
#         config_path = '/Users/rogerwesterbo/dev/github/viti/ipam-api/config-docker-compose.json'
#         with open(config_path, 'r') as f:
#             config = json.load(f)

#         # Ensure netbox section exists
#         if 'netbox' not in config:
#             config['netbox'] = {}

#         # Update prefix containers
#         config['netbox']['prefix_containers'] = {
#             "internet": "10.0.0.0/8",
#             "helsenett_private": "172.16.0.0/12",
#             "helsenett_public": "192.168.0.0/16"
#         }

#         with open(config_path, 'w') as f:
#             json.dump(config, f, indent=2)

#         print("✅ Configuration file updated successfully!")

#     except Exception as e:
#         print(f"❌ Failed to update configuration file: {str(e)}")


def verify_setup():
    """Verify that everything was set up correctly"""
    print("🔍 Verifying setup...")

    # Check tenant groups
    response = make_api_request("GET", "/tenancy/tenant-groups/")
    if response and response.status_code == 200:
        tenant_groups = response.json().get("results", [])
        dcn_groups = [tg for tg in tenant_groups if tg["name"] == "DCN"]
        print(f"  Tenant groups found: {len(dcn_groups)}/1")
        for tg in dcn_groups:
            print(f"    - {tg['name']}")

    # Check tenants
    response = make_api_request("GET", "/tenancy/tenants/")
    if response and response.status_code == 200:
        tenants = response.json().get("results", [])
        nhn_tenants = [t for t in tenants if t["name"] == "NHN"]
        print(f"  Tenants found: {len(nhn_tenants)}/1")
        for t in nhn_tenants:
            print(f"    - {t['name']}")

    # Check vrfs
    response = make_api_request("GET", "/ipam/vrfs/")
    if response and response.status_code == 200:
        vrfs = response.json().get("results", [])
        nhc_vrfs = [v for v in vrfs if v["name"] == "nhc"]
        print(f"  VRFs found: {len(nhc_vrfs)}/1")
        for vrf in nhc_vrfs:
            print(f"    - {vrf['name']}")

    # Check roles
    response = make_api_request("GET", "/ipam/roles/")
    if response and response.status_code == 200:
        roles = response.json().get("results", [])
        datacenter_roles = [r for r in roles if r["name"] == "datacenter"]
        print(f"  Roles found: {len(datacenter_roles)}/1")
        for r in datacenter_roles:
            print(f"    - {r['name']}")

    # Check custom fields
    response = make_api_request("GET", "/extras/custom-fields/")
    if response and response.status_code == 200:
        fields = response.json().get("results", [])
        expected_fields = ["k8s_uuid", "k8s_zone", "domain", "env", "infra", "purpose"]
        found_fields = [f for f in fields if f["name"] in expected_fields]
        print(f"  Custom fields found: {len(found_fields)}/{len(expected_fields)}")
        for field in found_fields:
            print(f"    - {field['name']}: {field.get('label', '')}")

    # Check custom field choice sets
    response = make_api_request("GET", "/extras/custom-field-choice-sets/")
    if response and response.status_code == 200:
        choice_sets = response.json().get("results", [])
        k8s_zone_sets = [cs for cs in choice_sets if cs["name"] == "k8s_zone_choices"]
        print(f"  Custom field choice sets found: {len(k8s_zone_sets)}/1")
        for cs in k8s_zone_sets:
            print(f"    - {cs['name']}")

    # Check prefixes
    response = make_api_request("GET", "/ipam/prefixes/")
    if response and response.status_code == 200:
        prefixes = response.json().get("results", [])
        container_prefixes = [
            p
            for p in prefixes
            if p["prefix"] in ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
        ]
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

    # Create tenant groups
    tenant_groups = create_tenant_groups()
    print()
    # Create tenants
    tenants = create_tenants(tenant_groups)
    print()

    # Create VRFs
    vrfs = create_vrfs()
    print()

    # Create VRFs
    roles = create_roles()
    print()

    # Create custom fields
    create_k8s_custom_fields()
    print()
    create_custom_fields()
    print()

    # Create prefix containers
    prefixes = create_prefix_containers(vrfs, roles, tenants)
    print()

    # Update config file
    # update_config_file(prefixes)
    print()

    # Verify setup
    verify_setup()
    print()

    print("=" * 60)
    print("🎉 Setup completed!")
    print("=" * 60)
    print("Your NetBox now has:")
    print("✅ Tenant group: DCN")
    print("✅ Tenant: NHN")
    print("✅ VRF: nhc")
    print("✅ Role: datacenter")
    print("✅ Custom fields:")
    print("   - k8s_uuid (text)")
    print("   - k8s_zone (select)")
    print("   - domain (text, required)")
    print("   - env (text, required)")
    print("   - infra (text, required)")
    print("   - purpose (text, required)")
    print("✅ Custom field choice set: k8s_zone_choices")
    print("✅ Three prefix containers:")
    print("   - 10.0.0.0/8 (inet)")
    print("   - 172.16.0.0/12 (hnet-private)")
    print("   - 192.168.0.0/16 (hnet-public)")
    # print("✅ Updated configuration file")
    print()
    print("💡 Access NetBox at: http://localhost:8000")
    print("   Username: admin")
    print("   Password: admin")


if __name__ == "__main__":
    main()
