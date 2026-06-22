def build_conditioning_payload(brand_reference_paths: list[str] | None) -> dict:
    refs = brand_reference_paths or []
    return {
        "controlnet_refs": refs,
        "ip_adapter_refs": refs,
    }
