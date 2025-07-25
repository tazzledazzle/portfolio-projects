import yaml
import os

CONFIG_DIR = os.path.join(os.path.dirname(__file__), '../../configs')

def load_profiles():
    profiles = []
    for filename in os.listdir(CONFIG_DIR):
        if filename.endswith(".yaml"):
            with open(os.path.join(CONFIG_DIR, filename)) as f:
                profiles.append(yaml.safe_load(f))
    return profiles
