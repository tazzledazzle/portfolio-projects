from pathlib import Path


def build_review_gallery(
    output_dir: Path,
    image_paths: list[str],
    video_paths: list[str],
) -> Path:
    output_dir.mkdir(parents=True, exist_ok=True)
    html_path = output_dir / "review_gallery.html"

    image_items = "\n".join(
        f'<li><img src="{path}" alt="{path}" style="max-width:240px" /><p>{path}</p></li>'
        for path in image_paths
    )
    video_items = "\n".join(
        f'<li><video src="{path}" controls style="max-width:320px"></video><p>{path}</p></li>'
        for path in video_paths
    )
    html = f"""<!doctype html>
<html>
  <head><meta charset="utf-8" /><title>Review Gallery</title></head>
  <body>
    <h1>Image Variants</h1>
    <ul>{image_items}</ul>
    <h1>Video Clips</h1>
    <ul>{video_items}</ul>
  </body>
</html>
"""
    html_path.write_text(html, encoding="utf-8")
    return html_path
