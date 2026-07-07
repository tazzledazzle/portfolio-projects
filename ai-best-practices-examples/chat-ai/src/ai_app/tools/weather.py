import json
from urllib.parse import quote
from urllib.request import urlopen


def fetch_weather(city: str) -> str:
    safe_city = quote(city)
    geo_url = (
        "https://geocoding-api.open-meteo.com/v1/search"
        f"?name={safe_city}&count=1&language=en&format=json"
    )
    try:
        with urlopen(geo_url, timeout=4) as response:
            payload = json.loads(response.read().decode("utf-8"))
        results = payload.get("results") or []
        if not results:
            return f"I could not find weather data for {city}"
        location = results[0]
        latitude = location["latitude"]
        longitude = location["longitude"]
        name = location.get("name", city)

        forecast_url = (
            "https://api.open-meteo.com/v1/forecast"
            f"?latitude={latitude}&longitude={longitude}&current=temperature_2m,weather_code"
        )
        with urlopen(forecast_url, timeout=4) as response:
            forecast_payload = json.loads(response.read().decode("utf-8"))
        current = forecast_payload.get("current", {})
        temp_c = current.get("temperature_2m", "?")
        code = current.get("weather_code", "?")
        return f"Current weather in {name}: {temp_c}C (code {code})"
    except Exception:
        return f"Weather tool is unavailable for {city} right now"
