import logging
import os
import json
import threading
from datetime import datetime

import requests

logger = logging.getLogger('projgen.telemetry')
handler = logging.StreamHandler()
formatter = logging.Formatter('[%(asctime)s] %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)
logger.setLevel(logging.INFO)

# Read endpoint from env; fall back to no‐op if unset
_TELEMETRY_ENDPOINT = os.getenv('PROJGEN_TELEMETRY_URL', '').rstrip('/')
_LOCK = threading.Lock()

def init_telemetry():
    logger.info('Telemetry initialized')
    # You could send a session‐start ping here if desired

def record_event(event_name: str, properties: dict = None):
    payload = {
        "event": event_name,
        "properties": properties or {},
        "timestamp": datetime.utcnow().isoformat() + "Z"
    }
    logger.info(f'Recording telemetry: {payload}')
    if not _TELEMETRY_ENDPOINT:
        return  # no endpoint configured
    # Send asynchronously so we don’t block CLI
    def _send():
        try:
            headers = {'Content-Type': 'application/json'}
            requests.post(_TELEMETRY_ENDPOINT + '/track', headers=headers, data=json.dumps(payload), timeout=2)
        except Exception as e:
            logger.debug(f"Telemetry send failed: {e}")
    threading.Thread(target=_send, daemon=True).start()
