import concurrent.futures
import requests
import json

def register_user(i):
    user = {
        "username": f"user{i}",
        "password": "password"
    }
    user_json = json.dumps(user)
    response = requests.post('http://8.130.54.212:12345/api/register', data=user_json)
    return response.status_code

with concurrent.futures.ThreadPoolExecutor() as executor:
    results = list(executor.map(register_user, range(100)))
    
def login_user(username, password):
    user = {
        "username": username,
        "password": password
    }
    user_json = json.dumps(user)
    response = requests.post('http://8.130.54.212:12345/api/login', data=user_json)
    return response.cookies if response.status_code == 200 else None
def add_comment(i, cookies):
    comment = {
        "nid": 1,
        "content": f"This is a comment from user{i}."
    }
    comment_json = json.dumps(comment)
    response = requests.post('http://8.130.54.212:12345/api/add_comment', data=comment_json, cookies=cookies)
    return response.status_code

cookies = [login_user(f'user{i}', 'password') for i in range(100)]

with concurrent.futures.ThreadPoolExecutor() as executor:
    results = list(executor.map(add_comment, range(100), cookies))