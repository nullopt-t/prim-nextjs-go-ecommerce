#!/usr/bin/env python3
"""Fetch real, product-matched images for the original catalogue products.

Direct manufacturer scraping is blocked in this environment, so this uses a
public image-search endpoint (Bing's image ``async`` fragment) to find real
product photos for each original product, then downloads and normalises the
best candidates to JPEG. One-off (network) build step.

Candidates are taken in order from *trusted* hosts first (manufacturers and
large retailers); secondary hosts are only accepted when the result title
mentions the brand and a model token. Each image is fetched from the original
URL (with a retry) and, failing that, from Bing's cached thumbnail so a
relevant image is almost always available.

Inputs : seed/data/original_products.json
Outputs: seed/media/catalog/products/<slug>.jpg (+ -2/-3/-4)
         seed/data/original_images.json       (slug -> [object keys])
         seed/data/original_images_meta.json  (audit trail)

Run from the backend/ directory:

    python3 seed/fetch_original_media.py [--limit N] [--only slug]
"""
import argparse
import hashlib
import html
import io
import json
import os
import re
import time
import urllib.parse
import urllib.request

from PIL import Image

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.path.join(HERE, "data")
MEDIA_PRODUCTS = os.path.join(HERE, "media", "catalog", "products")
RAW = os.path.join(DATA, "original_products.json")
OUT = os.path.join(DATA, "original_images.json")
META = os.path.join(DATA, "original_images_meta.json")

UA = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
)
BING = "https://www.bing.com/images/async?q={q}&first={first}&count=35&mmasync=1"

TIER1 = (
    "apple.com", "samsung.com", "sony.com", "bose.com", "dell.com",
    "logitech.com", "nike.com", "adidas.com", "lenovo.com", "asus.com",
    "hp.com", "microsoft.com", "playstation.com", "razer.com", "vaio.com",
    "garmin.com", "google.com", "m.media-amazon.com", "walmartimages.com",
    "bestbuy.com", "bbystatic.com", "target.com",
)
BLOCK = (
    "bhphotovideo.com", "pinterest.", "lookaside.", "fbsbx.com", "ebayimg.com",
    "sammobile.com", "androidauthority.com", "digitaltrends.com",
    "futurecdn.net", "anpoimages.com", "idealo.com", "pricerunner.com",
    "gsmarena.com", "tuttoandroid.net", "01net.com", "techbyte.sk",
    "gizchina.it", "androidheadlines.com", "tomsguide",
)
TARGET = 4
MAX_DL = 6
MIN_SIDE = 320


def fetch(url, referer=None):
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    if referer:
        req.add_header("Referer", referer)
    with urllib.request.urlopen(req, timeout=35) as r:
        return r.read(), r.headers.get("Content-Type", "")


def download(url):
    for attempt in range(2):
        try:
            return fetch(url, referer="https://www.bing.com/")
        except Exception:
            time.sleep(0.6 + attempt)
    return None


def model_tokens(title):
    return {t for t in re.findall(r"[a-z0-9]+", title.lower()) if any(c.isdigit() for c in t) and len(t) >= 2}


def relevant(title, brand, t):
    low = (t or "").lower()
    if brand and brand.lower() not in low:
        return False
    toks = model_tokens(title)
    return (not toks) or any(tok in low for tok in toks)


def search(query, brand, title):
    out = []
    for first in (0, 35):
        url = BING.format(q=urllib.parse.quote(query), first=first)
        try:
            body, _ = fetch(url)
        except Exception:
            continue
        text = body.decode("utf-8", "ignore")
        for raw in re.findall(r'm="(\{[^"]+?\})"', text):
            try:
                d = json.loads(html.unescape(raw))
            except Exception:
                continue
            murl, turl = d.get("murl"), d.get("turl")
            if not murl or not murl.lower().split("?")[0].endswith((".jpg", ".jpeg", ".png", ".webp")):
                continue
            host = urllib.parse.urlparse(murl).netloc.lower()
            if any(b in host for b in BLOCK):
                continue
            t = d.get("t", "")
            if any(p in host for p in TIER1):
                tier = 0
            elif relevant(title, brand, t):
                tier = 1
            else:
                continue
            out.append((tier, murl, turl, host, t))
        time.sleep(0.3)
    seen, uniq = set(), []
    for tier, murl, turl, host, t in sorted(out, key=lambda x: x[0]):
        if murl in seen:
            continue
        seen.add(murl)
        uniq.append((tier, murl, turl, host, t))
    return uniq


def amazon_bigger(url):
    if "media-amazon.com" not in url:
        return url
    return re.sub(r"\._[A-Z0-9,_]+_\.", "._SL1200_.", url)


def white_score(raw):
    img = Image.open(io.BytesIO(raw)).convert("RGB")
    img.thumbnail((220, 220))
    w, h = img.size
    px = img.load()
    border = [px[x, 0] for x in range(w)] + [px[x, h - 1] for x in range(w)]
    border += [px[0, y] for y in range(h)] + [px[w - 1, y] for y in range(h)]
    border = border[:: max(1, len(border) // 500)]
    good = sum(1 for r, g, b in border if r >= 225 and g >= 225 and b >= 225)
    return good / len(border)


def normalise(raw, dest, max_side=900):
    img = Image.open(io.BytesIO(raw)).convert("RGBA")
    if min(img.size) < MIN_SIDE:
        return False
    bg = Image.new("RGB", img.size, (255, 255, 255))
    bg.paste(img, mask=img.split()[3])
    w, h = bg.size
    scale = max_side / max(w, h)
    if scale < 1:
        bg = bg.resize((round(w * scale), round(h * scale)), Image.LANCZOS)
    bg.save(dest, "JPEG", quality=86, optimize=True)
    return True


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--limit", type=int, default=0)
    ap.add_argument("--only", default="")
    args = ap.parse_args()

    products = json.load(open(RAW))
    if args.only:
        products = [p for p in products if p["slug"] == args.only]
    if args.limit:
        products = products[: args.limit]

    os.makedirs(MEDIA_PRODUCTS, exist_ok=True)
    used_hashes = set()
    mapping, meta = {}, {}

    for i, p in enumerate(products, 1):
        slug, title, brand = p["slug"], p["title"], p["brand"]
        query = title if title.lower().startswith(brand.lower()) else f"{brand} {title}"

        got, srcs, local = [], [], set()
        for _tier, murl, turl, host, _t in search(query, brand, title):
            if len(got) >= TARGET:
                break
            for url in dict.fromkeys((amazon_bigger(murl), murl, turl)):
                if not url:
                    continue
                res = download(url)
                if not res:
                    continue
                raw, ctype = res
                if not ctype.startswith("image") and not url.lower().split("?")[0].endswith(
                    (".jpg", ".jpeg", ".png", ".webp")
                ):
                    continue
                h = hashlib.md5(raw).hexdigest()
                if h in local or h in used_hashes:
                    break
                dest = os.path.join(MEDIA_PRODUCTS, f"{slug}.jpg" if not got else f"{slug}-{len(got) + 1}.jpg")
                try:
                    if not normalise(raw, dest):
                        break
                except Exception:
                    continue
                local.add(h)
                used_hashes.add(h)
                got.append(f"products/{os.path.basename(dest)}")
                try:
                    w = round(white_score(raw), 2)
                except Exception:
                    w = -1
                srcs.append({"key": got[-1], "host": host, "white": w, "url": murl})
                break
            time.sleep(0.1)
        mapping[slug] = got
        meta[slug] = srcs
        print(f"[{i}/{len(products)}] {len(got)} imgs  {slug}")
        time.sleep(0.4)

    if os.path.exists(OUT) and not args.only:
        prev = json.load(open(OUT))
        prev.update(mapping)
        mapping = prev
    if os.path.exists(META) and not args.only:
        prevm = json.load(open(META))
        prevm.update(meta)
        meta = prevm
    json.dump(mapping, open(OUT, "w"), indent=2, ensure_ascii=False)
    json.dump(meta, open(META, "w"), indent=2, ensure_ascii=False)

    zero = [s for s, v in mapping.items() if not v]
    print(f"\nWrote {len(mapping)} products, total imgs {sum(len(v) for v in mapping.values())}")
    if zero:
        print(f"WARNING {len(zero)} products with no image: {zero}")


if __name__ == "__main__":
    main()
