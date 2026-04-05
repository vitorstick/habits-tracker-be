import requests
import json
import os

BASE_URL = "http://localhost:8080/api"
PRODUCTION_URL = "https://habits-tracker-be.onrender.com/api"


class HabitTrackerAPI:
    def __init__(self, base_url=BASE_URL):
        self.base_url = base_url
        self.token = None

    def register(self, email, name, password):
        url = f"{self.base_url}/auth/register"
        payload = {
            "email": email,
            "name": name,
            "password": password
        }
        response = requests.post(url, json=payload)
        if response.status_code == 201:
            data = response.json()
            self.token = data.get("token")
            return data
        else:
            print(f"Registration failed: {response.text}")
            return None

    def login(self, email, password):
        url = f"{self.base_url}/auth/login"
        payload = {
            "email": email,
            "password": password
        }
        response = requests.post(url, json=payload)
        if response.status_code == 200:
            data = response.json()
            self.token = data.get("token")
            return data
        else:
            print(f"Login failed: {response.text}")
            return None

    def get_habits(self):
        if not self.token:
            print("Error: No token. Login or Register first.")
            return None
        
        url = f"{self.base_url}/habits"
        headers = {"Authorization": f"Bearer {self.token}"}
        response = requests.get(url, headers=headers)
        if response.status_code == 200:
            return response.json()
        return None

    def create_habit(self, name, description=""):
        if not self.token:
            return None
        
        url = f"{self.base_url}/habits"
        headers = {"Authorization": f"Bearer {self.token}"}
        payload = {"name": name, "description": description}
        response = requests.post(url, json=payload, headers=headers)
        return response.json() if response.status_code == 201 else None

    def log_habit(self, habit_id):
        if not self.token:
            return None
        
        url = f"{self.base_url}/habits/{habit_id}/log"
        headers = {"Authorization": f"Bearer {self.token}"}
        response = requests.post(url, headers=headers)
        return response.json() if response.status_code == 200 else None

    def logout(self):
        """Logging out is client-side: just discard the token."""
        self.token = None
        print("Logged out (token cleared locally).")

if __name__ == "__main__":
    # Example usage:
    # To use production: api = HabitTrackerAPI(base_url=PRODUCTION_URL)
    api = HabitTrackerAPI()

    
    # login
    email = "test@example.com"
    password = "password123"
    
    print(f"Logging in as {email}...")
    auth_data = api.login(email, password)
    
    if auth_data:
        print("Login successful!")
        habits = api.get_habits()
        print(f"Found {len(habits)} habits.")
        for h in habits:
            print(f"- {h['name']} (ID: {h['id']})")
    else:
        print("Could not login. Trying to register...")
        reg_data = api.register(email, "Test User", password)
        if reg_data:
            print("Registration successful!")
