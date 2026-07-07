
from ai_image_video_generator.config import AppConfig
from ai_image_video_generator.config import get_quality_profile
from ai_image_video_generator.config import load_config


def test_load_config_defaults() -> None:
    config = load_config({})
    assert config.backend == "auto"
    assert config.comfyui_base_url == "http://127.0.0.1:8188"
    assert config.default_variant_count == 3


def test_load_config_from_env_mapping() -> None:
    env = {
        "AIVG_COMFYUI_BASE_URL": "http://localhost:9999",
        "AIVG_OUTPUT_DIR": "custom-output",
    }
    config = load_config(env)
    assert isinstance(config, AppConfig)
    assert config.comfyui_base_url == "http://localhost:9999"
    assert config.output_dir == "custom-output"


def test_get_quality_profile_default_photorealism_settings() -> None:
    profile = get_quality_profile("photo_studio_v1")
    assert profile.sampler == "dpmpp_2m_sde"
    assert profile.steps >= 30
    assert profile.cfg_scale <= 7.0
