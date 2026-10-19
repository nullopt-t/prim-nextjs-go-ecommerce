#!/usr/bin/env python3
"""Fetch real product images for the extra catalogue products.

This is a one-off (network) build step. It reads the committed DummyJSON
snapshot in ``seed/data/dummyjson_products.json`` (which hosts official
product renders/photos on a public CDN), downloads the images for a curated
selection of products, normalises them to JPEG, writes them under
``seed/media/catalog/products/`` and emits ``seed/data/extra_products.json``
for ``generate_seed.py`` to ingest.

Run from the backend/ directory:

    python3 seed/fetch_extra_media.py
"""
import io
import json
import os
import re
import urllib.request

from PIL import Image

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.path.join(HERE, "data")
MEDIA_PRODUCTS = os.path.join(HERE, "media", "catalog", "products")
RAW = os.path.join(DATA, "dummyjson_products.json")
OUT = os.path.join(DATA, "extra_products.json")

UA = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/124 Safari/537.36"
)

# DummyJSON brand -> keep only genuine manufacturer/brand names; anything
# generic ("Fashion Trends", "Classic Wear", ...) becomes our house brand.
REAL_BRANDS = {
    "Apple", "Samsung", "Asus", "Huawei", "Lenovo", "Dell", "Oppo", "Realme",
    "Vivo", "Nike", "Puma", "Off White", "Calvin Klein", "Prada", "Rolex",
    "Longines", "IWC", "Heshe", "Beats", "Gigabyte", "Amazon",
}
HOUSE_BRAND = "PRIM Select"

# DummyJSON category -> our catalogue leaf category.
CATEGORY_MAP = {
    "laptops": "Laptops",
    "smartphones": "Smartphones",
    "tablets": "Tablets & Wearables",
    "mens-shoes": "Shoes",
    "womens-shoes": "Shoes",
    "mens-shirts": "Shirts",
    "tops": "Tops",
    "womens-dresses": "Dresses",
    "sunglasses": "Sunglasses",
    "mens-watches": "Watches",
    "womens-watches": "Watches",
    "mobile-accessories": "Mobile Accessories",
    "womens-bags": "Bags",
    "womens-jewellery": "Jewellery",
    "sports-accessories": "Sports & Outdoors",
}

# Curated selection (titles must match the DummyJSON dataset exactly).
SELECTION = [
    # Laptops
    "Asus Zenbook Pro Dual Screen Laptop",
    "Huawei Matebook X Pro",
    "Lenovo Yoga 920",
    "New DELL XPS 13 9300 Laptop",
    # Smartphones
    "iPhone 13 Pro",
    "iPhone X",
    "Samsung Galaxy S10",
    "Samsung Galaxy S8",
    "Oppo K1",
    "Vivo X21",
    # Tablets
    "iPad Mini 2021 Starlight",
    "Samsung Galaxy Tab S8 Plus Grey",
    "Samsung Galaxy Tab White",
    # Men's shoes
    "Nike Air Jordan 1 Red And Black",
    "Puma Future Rider Trainers",
    "Sports Sneakers Off White & Red",
    # Women's shoes
    "Calvin Klein Heel Shoes",
    "Golden Shoes Woman",
    "Red Shoes",
    # Shirts
    "Blue & Black Check Shirt",
    "Man Plaid Shirt",
    "Men Check Shirt",
    # Tops
    "Blue Frock",
    "Girl Summer Dress",
    "Gray Dress",
    # Dresses
    "Black Women's Gown",
    "Dress Pea",
    "Corset Leather With Skirt",
    # Sunglasses
    "Black Sun Glasses",
    "Classic Sun Glasses",
    "Green and Black Glasses",
    # Watches
    "Rolex Submariner Watch",
    "Rolex Datejust",
    "Longines Master Collection",
    "IWC Ingenieur Automatic Steel",
    "Women's Wrist Watch",
    # Mobile accessories
    "Apple Airpods",
    "Apple Watch Series 4 Gold",
    "iPhone 12 Silicone Case with MagSafe Plum",
    # Bags
    "Prada Women Bag",
    "Heshe Women's Leather Bag",
    "White Faux Leather Backpack",
    # Jewellery
    "Green Crystal Earring",
    "Tropical Earring",
    # Sports
    "Baseball Glove",
    "Cricket Helmet",
]

USD_TO_EGP = 50  # rough dev conversion


def slugify(text):
    text = text.lower()
    text = re.sub(r"[^a-z0-9]+", "-", text).strip("-")
    return re.sub(r"-+", "-", text)


def download(url):
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=60) as resp:
        return resp.read()


def normalise(raw_bytes, out_path, max_side=900):
    img = Image.open(io.BytesIO(raw_bytes)).convert("RGBA")
    # Flatten any transparency onto white so JPEG output looks clean.
    bg = Image.new("RGB", img.size, (255, 255, 255))
    bg.paste(img, mask=img.split()[3])
    w, h = bg.size
    scale = max_side / max(w, h)
    if scale < 1:
        bg = bg.resize((round(w * scale), round(h * scale)), Image.LANCZOS)
    bg.save(out_path, "JPEG", quality=86, optimize=True)


def main():
    products = json.load(open(RAW))["products"]
    by_title = {p["title"]: p for p in products}
    os.makedirs(MEDIA_PRODUCTS, exist_ok=True)

    extra, missing = [], []
    for title in SELECTION:
        p = by_title.get(title)
        if not p:
            missing.append(title)
            continue
        slug = slugify(title)
        cat = CATEGORY_MAP[p["category"]]
        brand = p.get("brand") or HOUSE_BRAND
        if brand not in REAL_BRANDS:
            brand = HOUSE_BRAND

        images = []
        for i, url in enumerate(p["images"]):
            key = f"products/{slug}.jpg" if i == 0 else f"products/{slug}-{i + 1}.jpg"
            dest = os.path.join(MEDIA_PRODUCTS, os.path.basename(key))
            normalise(download(url), dest)
            images.append(key)

        price = round(p["price"] * USD_TO_EGP)
        crossed = round(price * 1.25)
        highlights = [
            p.get("warrantyInformation", "").strip(),
            p.get("shippingInformation", "").strip(),
            p.get("returnPolicy", "").strip(),
        ]
        highlights = [h for h in highlights if h][:3]

        extra.append(
            {
                "slug": slug,
                "title": title,
                "brand": brand,
                "category": cat,
                "description": p.get("description", title),
                "highlights": highlights,
                "price": price,
                "crossed_price": crossed,
                "images": images,
            }
        )
        print(f"  {len(images)} imgs  {slug}")

    if missing:
        raise SystemExit(f"selected titles not found in dataset: {missing}")

    json.dump(extra, open(OUT, "w"), indent=2, ensure_ascii=False)
    print(f"\nWrote {len(extra)} products ({sum(len(p['images']) for p in extra)} images) -> {OUT}")


if __name__ == "__main__":
    main()
