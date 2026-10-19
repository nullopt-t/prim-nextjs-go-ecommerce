#!/usr/bin/env python3
"""Build one distinct, real logo per catalogue brand.

Most logos come from the Simple Icons set (official monochrome marks). Brands
missing there are pulled from Wikimedia Commons. House/obscure brands with no
freely available mark get a clean, distinct wordmark rendered locally so no two
brands ever share the same placeholder image.

Outputs a normalised 512x512 PNG per brand under media/catalog/brands/ and
seed/data/brand_logos.json for generate_seed.py.

Run from the backend/ directory:

    python3 seed/fetch_brand_logos.py
"""
import json
import os
import subprocess
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request

from PIL import Image, ImageChops, ImageDraw, ImageFont

HERE = os.path.dirname(os.path.abspath(__file__))
MEDIA = os.path.join(HERE, "media", "catalog", "brands")
DATA = os.path.join(HERE, "data")
OUT = os.path.join(DATA, "brand_logos.json")

UA = "prim-seed/1.0 (dev seed build)"
SIMPLE_BASE = "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/{slug}.svg"
FONT = "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"

# Brand order matches the seeded brand list (core brands first) so brand IDs
# stay stable across regenerations.
BRANDS = [
    ("Apple", "apple", None, "https://apple.com"),
    ("Samsung", "samsung", None, "https://samsung.com"),
    ("Nike", "nike", None, "https://nike.com"),
    ("Sony", "sony", None, "https://sony.com"),
    ("Adidas", "adidas", None, "https://adidas.com"),
    ("Dell", "dell", None, "https://dell.com"),
    ("Logitech", None, "File:Logitech Logo 2012.svg", "https://logitech.com"),
    ("Bose", "bose", None, "https://bose.com"),
    ("Asus", "asus", None, "https://asus.com"),
    ("Calvin Klein", None, "File:Calvin klein logo.svg", "https://calvinklein.com"),
    ("Heshe", None, None, "https://heshe.com"),
    ("Huawei", "huawei", None, "https://huawei.com"),
    ("IWC", None, "File:IWC Schaffhausen Logo.jpg", "https://iwc.com"),
    ("Lenovo", "lenovo", None, "https://lenovo.com"),
    ("Longines", None, "File:Longines wordmark logo.svg", "https://longines.com"),
    ("Off White", None, None, "https://off---white.com"),
    ("Oppo", "oppo", None, "https://oppo.com"),
    ("PRIM Select", None, None, None),
    ("Prada", None, "File:Prada-Logo.svg", "https://prada.com"),
    ("Puma", "puma", None, "https://puma.com"),
    ("Rolex", None, "File:Rolex wordmark logo.svg", "https://rolex.com"),
    ("Vivo", "vivo", None, "https://vivo.com"),
]


def slugify(name):
    out = "".join(c.lower() if c.isalnum() else "-" for c in name)
    while "--" in out:
        out = out.replace("--", "-")
    return out.strip("-")


def get(url, tries=4):
    for attempt in range(tries):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": UA})
            with urllib.request.urlopen(req, timeout=30) as r:
                return r.read()
        except urllib.error.HTTPError as e:
            if e.code == 429 and attempt < tries - 1:
                time.sleep(5 * (attempt + 1))
                continue
            raise


def commons_thumb(title):
    url = (
        "https://commons.wikimedia.org/w/api.php?action=query&titles="
        + urllib.parse.quote(title)
        + "&prop=imageinfo&iiprop=url&iiurlwidth=900&format=json"
    )
    data = json.loads(get(url))
    for page in data["query"]["pages"].values():
        info = (page.get("imageinfo") or [{}])[0]
        return info.get("thumburl") or info.get("url")
    raise RuntimeError(f"no image for {title}")


def render_svg(svg):
    with tempfile.NamedTemporaryFile(suffix=".svg", delete=False) as f:
        f.write(svg)
        src = f.name
    dst = src + ".png"
    try:
        subprocess.run(
            ["rsvg-convert", "-w", "700", "-h", "700", "-b", "white", src, "-o", dst],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        return Image.open(dst).convert("RGB")
    finally:
        for p in (src, dst):
            try:
                os.remove(p)
            except OSError:
                pass


def normalise(img, box=420):
    if img.mode in ("RGBA", "LA"):
        bg = Image.new("RGB", img.size, (255, 255, 255))
        bg.paste(img, mask=img.split()[-1])
        img = bg
    else:
        img = img.convert("RGB")
    # Trim the uniform white margin so the mark fills the tile evenly.
    diff = ImageChops.difference(img, Image.new("RGB", img.size, (255, 255, 255)))
    bbox = diff.getbbox()
    if bbox:
        img = img.crop(bbox)
    w, h = img.size
    scale = min(box / w, box / h)
    img = img.resize((max(1, round(w * scale)), max(1, round(h * scale))), Image.LANCZOS)
    canvas = Image.new("RGB", (512, 512), (255, 255, 255))
    canvas.paste(img, ((512 - img.width) // 2, (512 - img.height) // 2))
    return canvas


def wordmark(name):
    canvas = Image.new("RGB", (512, 512), (255, 255, 255))
    draw = ImageDraw.Draw(canvas)
    text = name.upper()
    size = 92
    font = None
    while size >= 24:
        font = ImageFont.truetype(FONT, size)
        bb = draw.textbbox((0, 0), text, font=font)
        if bb[2] - bb[0] <= 420:
            break
        size -= 4
    bb = draw.textbbox((0, 0), text, font=font)
    tw, th = bb[2] - bb[0], bb[3] - bb[1]
    draw.text(((512 - tw) / 2 - bb[0], (512 - th) / 2 - bb[1]), text, font=font, fill=(17, 24, 39))
    return canvas


def main():
    os.makedirs(MEDIA, exist_ok=True)
    manifest = []

    for name, simple_slug, commons_title, link in BRANDS:
        slug = slugify(name)
        key = f"brands/{slug}-logo.png"
        dest = os.path.join(MEDIA, f"{slug}-logo.png")

        img = None
        try:
            if simple_slug:
                img = render_svg(get(SIMPLE_BASE.format(slug=simple_slug)))
            elif commons_title:
                raw = get(commons_thumb(commons_title))
                from io import BytesIO

                img = Image.open(BytesIO(raw))
        except Exception as exc:  # network/format problem -> distinct wordmark
            print(f"  ! {name}: {exc}")

        img = normalise(img) if img else wordmark(name)
        img.save(dest, "PNG", optimize=True)
        size = os.path.getsize(dest)
        manifest.append({"brand": name, "key": key, "size": size, "link": link})
        print(f"  {name:14} {key:32} {size:>7} bytes")
        time.sleep(0.2)

    json.dump(manifest, open(OUT, "w"), indent=2, ensure_ascii=False)
    print(f"\nWrote {len(manifest)} brand logos")


if __name__ == "__main__":
    main()
