from pydantic import BaseModel

#### OAuth2 functions and models
fake_users_db = {
    "johndoe": {
        "username": "johndoe",
        "hashed_password": "hashedsecret",
    },
    "janedoe": {
        "username": "janedoe",
        "hashed_password": "hashedsecret2",
    },
}

class User(BaseModel):
    username: str

def fakely_hash_password(password: str):
    return f"hashed{password}"

class UserInDB(User):
    hashed_password: str

def get_user(db, username: str):
    if username in db:
        user_dict = db[username]
        return UserInDB(**user_dict)
    
def fake_token_generator(user: UserInDB) -> str:
    # This doesn't provide any security at all
    return f"tokenized{user.username}"

def fake_token_resolver(token: str) -> UserInDB | None:
    if token.startswith("tokenized"):
        user_id = token.removeprefix("tokenized")
        user = get_user(fake_users_db, user_id)
        return user

def get_user_from_token(token: str) -> User | None:
    """Get user from token."""
    user_in_db = fake_token_resolver(token)
    if user_in_db:
        return User(username=user_in_db.username)
    return None