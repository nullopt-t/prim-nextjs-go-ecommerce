import random
import json
import os

random.seed(42)

def hex_id(prefix, idx):
    return f"{prefix:08x}-0000-0000-0000-{idx:012x}"

def esc(val):
    if val is None:
        return "NULL"
    return "'" + str(val).replace("'", "''") + "'"

def json_esc(obj):
    s = json.dumps(obj)
    return "'" + s.replace("'", "''") + "'::jsonb"

# Brand logos
brand_logos = [
    ("brands/apple-logo.png", "image/png", 45000),
    ("brands/samsung-logo.png", "image/png", 38000),
    ("brands/nike-logo.png", "image/png", 32000),
    ("brands/sony-logo.png", "image/png", 29000),
    ("brands/adidas-logo.png", "image/png", 35000),
    ("brands/dell-logo.png", "image/png", 41000),
    ("brands/logitech-logo.png", "image/png", 28000),
    ("brands/bose-logo.png", "image/png", 31000),
]

# Gallery angles by category
gallery_angles_by_cat = {
    "Laptops": ["products/laptop-angle-keyboard.jpg", "products/laptop-angle-side.jpg", "products/laptop-lifestyle-desk.jpg"],
    "Smartphones": ["products/phone-angle-camera.jpg", "products/phone-angle-screen.jpg", "products/phone-lifestyle-hand.jpg"],
    "Audio & Headphones": ["products/audio-case-detail.jpg", "products/audio-lifestyle-wear.jpg", "products/audio-cushion-macro.jpg"],
    "Shoes": ["products/shoe-sole-traction.jpg", "products/shoe-heel-angle.jpg", "products/shoe-pair-overhead.jpg"],
    "Outerwear": ["products/apparel-fabric-texture.jpg", "products/apparel-zipper-detail.jpg", "products/apparel-model-back.jpg"],
    "Tablets & Wearables": ["products/tablet-pencil-drawing.jpg", "products/wearable-wrist-display.jpg"],
    "Computer Accessories": ["products/accessory-sensor-bottom.jpg", "products/accessory-desk-setup.jpg"],
    "Gaming Consoles": ["products/gaming-controller-macro.jpg", "products/gaming-console-stand.jpg"],
}

brand_logo_map = {}
storage_objects = []
image_to_obj_id = {}
obj_idx = 1

for key, ctype, size in brand_logos:
    obj_id = hex_id(0x80000000, obj_idx)
    storage_objects.append((obj_id, 'catalog', key, ctype, size))
    brand_name = key.split("/")[1].replace("-logo.png", "").capitalize()
    brand_logo_map[brand_name] = obj_id
    obj_idx += 1

# Register extra gallery angles
for cat, angles in gallery_angles_by_cat.items():
    for ang in angles:
        if ang not in image_to_obj_id:
            obj_id = hex_id(0x80000000, obj_idx)
            storage_objects.append((obj_id, 'catalog', ang, 'image/jpeg', 120000))
            image_to_obj_id[ang] = obj_id
            obj_idx += 1

# Catalog Definitions with explicit multi-variants where applicable
# Format: (category, brand, title, slug, desc, highlights, [ (var_title, price, crossed_price, attrs, optional_image_file) ])
catalog_defs = [
    # Laptops (15)
    ("Laptops", "Apple", "MacBook Pro 16\" M3 Max", "macbook-pro-16-m3-max", "Extreme performance with 16-core CPU and 40-core GPU.", ["16.2\" Liquid Retina XDR", "Up to 128GB Unified Memory", "22 hours battery life"], [
        ("Space Black 36GB/1TB", 349900, 379900, {"color":"Space Black","ram":"36GB","storage":"1TB"}, "products/macbook-pro-16-m3-max.jpg"),
        ("Silver 48GB/1TB", 399900, 429900, {"color":"Silver","ram":"48GB","storage":"1TB"}, "products/macbook-pro-16-m3-max-silver.jpg")
    ]),
    ("Laptops", "Apple", "MacBook Air 15\" M3", "macbook-air-15-m3", "Impossibly thin and fast 15-inch laptop.", ["15.3\" Liquid Retina display", "18-hour battery life", "Fanless quiet operation"], [
        ("Midnight 16GB/512GB", 149900, 169900, {"color":"Midnight","ram":"16GB","storage":"512GB"}, "products/macbook-air-15-m3.jpg"),
        ("Starlight 8GB/256GB", 129900, 139900, {"color":"Starlight","ram":"8GB","storage":"256GB"}, "products/macbook-air-15-m3-starlight.jpg")
    ]),
    ("Laptops", "Apple", "MacBook Pro 14\" M3 Pro", "macbook-pro-14-m3-pro", "Compact powerhouse for developers and creatives.", ["14.2\" ProMotion 120Hz", "Thunderbolt 4 ports", "HDMI and SD card reader"], [
        ("Space Black 18GB/512GB", 199900, 219900, {"color":"Space Black","ram":"18GB","storage":"512GB"}, "products/macbook-pro-14-m3-pro.jpg"),
        ("Silver 36GB/1TB", 239900, 259900, {"color":"Silver","ram":"36GB","storage":"1TB"}, "products/macbook-pro-16-m3-max-silver.jpg")
    ]),
    ("Laptops", "Dell", "Dell XPS 15 OLED", "dell-xps-15-oled", "Masterfully crafted creator laptop with 3.5K OLED.", ["15.6\" 3.5K Touch OLED", "Intel Core i9-13900H", "NVIDIA RTX 4070"], [
        ("Platinum Silver 32GB/1TB", 229900, 249900, {"color":"Platinum Silver","ram":"32GB","storage":"1TB"}, "products/dell-xps-15-oled.jpg"),
        ("Platinum Silver 64GB/2TB", 279900, 299900, {"color":"Platinum Silver","ram":"64GB","storage":"2TB"}, "products/dell-xps-15-oled.jpg")
    ]),
    ("Laptops", "Dell", "Dell XPS 13 Plus", "dell-xps-13-plus", "Minimalist, futuristic ultraportable with seamless glass trackpad.", ["13.4\" InfinityEdge 4K Touch", "Intel Core i7-1360P", "Capacitive touch function row"], [
        ("Graphite 16GB/512GB", 139900, 159900, {"color":"Graphite","ram":"16GB","storage":"512GB"}, "products/dell-xps-13-plus.jpg"),
        ("Platinum 32GB/1TB", 169900, 189900, {"color":"Platinum","ram":"32GB","storage":"1TB"}, "products/dell-xps-13-plus.jpg")
    ]),
    ("Laptops", "Dell", "Dell Alienware m16 R2", "alienware-m16-r2", "Stealth mode gaming laptop engineered for pro gamers.", ["16\" QHD+ 240Hz 3ms display", "Intel Core Ultra 9", "Cryo-tech cooling technology"], [
        ("Dark Metallic Moon 32GB/1TB RTX 4070", 209900, 239900, {"color":"Dark Metallic Moon","ram":"32GB","storage":"1TB"}, "products/alienware-m16-r2.jpg"),
        ("Dark Metallic Moon 64GB/2TB RTX 4080", 269900, 299900, {"color":"Dark Metallic Moon","ram":"64GB","storage":"2TB"}, "products/alienware-m16-r2.jpg")
    ]),
    ("Laptops", "Samsung", "Galaxy Book4 Ultra", "galaxy-book4-ultra", "Dynamic AMOLED 2X touchscreen with Intel Core Ultra.", ["16\" 3K 120Hz Touch AMOLED", "NVIDIA RTX 4070 Studio", "AKG Quad speakers with Dolby Atmos"], [
        ("Moonstone Gray 32GB/1TB", 239900, 269900, {"color":"Moonstone Gray","ram":"32GB","storage":"1TB"}, "products/galaxy-book4-ultra.jpg")
    ]),
    ("Laptops", "Samsung", "Galaxy Book4 Pro 360", "galaxy-book4-pro-360", "Versatile 2-in-1 convertible laptop with included S Pen.", ["360-degree folding hinge", "16\" AMOLED 120Hz display", "Ultra-low latency S Pen included"], [
        ("Platinum Silver 16GB/512GB", 159900, 179900, {"color":"Platinum Silver","ram":"16GB","storage":"512GB"}, "products/galaxy-book4-pro-360.jpg"),
        ("Moonstone Gray 32GB/1TB", 189900, 209900, {"color":"Moonstone Gray","ram":"32GB","storage":"1TB"}, "products/galaxy-book4-pro-360.jpg")
    ]),
    ("Laptops", "Sony", "Sony VAIO SX14 Ultra", "vaio-sx14-ultra", "Premium Japanese crafted carbon fiber executive laptop.", ["14\" 4K Anti-glare display", "Under 1.05kg weight", "Full array of legacy and USB-C ports"], [
        ("All Black Edition 32GB/1TB", 189900, 209900, {"color":"All Black","ram":"32GB","storage":"1TB"}, "products/vaio-sx14-ultra.jpg")
    ]),
    ("Laptops", "Dell", "Dell Latitude 9440 2-in-1", "dell-latitude-9440", "World's most collaborative commercial PC.", ["Collaboration Touchpad with Zoom shortcuts", "QHD+ InfinityEdge", "Zero-lattice mini-LED keyboard"], [
        ("Titan Gray 16GB/512GB", 179900, 199900, {"color":"Titan Gray","ram":"16GB","storage":"512GB"}, "products/dell-latitude-9440.jpg")
    ]),
    ("Laptops", "Dell", "Dell Precision 5680 Workstation", "dell-precision-5680", "World's smallest 16-inch high-performance workstation.", ["NVIDIA RTX 5000 Ada Generation", "16\" OLED Touch 100% DCI-P3", "Patented Dual Opposite Outlet fans"], [
        ("Aluminum Gray 64GB/2TB", 389900, 429900, {"color":"Aluminum Gray","ram":"64GB","storage":"2TB"}, "products/dell-precision-5680.jpg")
    ]),
    ("Laptops", "Apple", "MacBook Pro 14\" M3 Base", "macbook-pro-14-m3-base", "Affordable entry into the professional MacBook Pro lineup.", ["M3 8-core CPU, 10-core GPU", "Liquid Retina XDR display", "512GB high speed SSD"], [
        ("Silver 8GB/512GB", 159900, 169900, {"color":"Silver","ram":"8GB","storage":"512GB"}, "products/macbook-pro-14-m3-base.jpg"),
        ("Space Gray 16GB/512GB", 179900, 189900, {"color":"Space Gray","ram":"16GB","storage":"512GB"}, "products/macbook-pro-14-m3-base.jpg")
    ]),
    ("Laptops", "Samsung", "Galaxy Book4 360 15\"", "galaxy-book4-360-15", "Lightweight everyday convertible laptop with Super AMOLED.", ["15.6\" FHD Super AMOLED", "Intel Core 7 Series 1", "Fast 65W USB-C charger"], [
        ("Gray 16GB/512GB", 119900, 134900, {"color":"Gray","ram":"16GB","storage":"512GB"}, "products/galaxy-book4-360-15.jpg")
    ]),
    ("Laptops", "Dell", "Dell Inspiron 16 Plus", "dell-inspiron-16-plus", "Spacious 16-inch laptop built for creativity and entertainment.", ["16\" 2.5K 16:10 display", "Intel Core Ultra 7", "ExpressCharge 80% in 60 minutes"], [
        ("Ice Blue 16GB/1TB", 109900, 124900, {"color":"Ice Blue","ram":"16GB","storage":"1TB"}, "products/dell-inspiron-16-plus.jpg")
    ]),
    ("Laptops", "Dell", "Dell G16 Gaming Laptop", "dell-g16-gaming", "Immersive 16-inch gaming with mechanical cherry keyboard options.", ["16\" QHD+ 165Hz display", "NVIDIA RTX 4060", "Game Shift technology macro key"], [
        ("Metallic Nightshade 16GB/512GB", 129900, 144900, {"color":"Metallic Nightshade","ram":"16GB","storage":"512GB"}, "products/dell-g16-gaming.jpg")
    ]),

    # Smartphones (14)
    ("Smartphones", "Samsung", "Galaxy S24 Ultra", "galaxy-s24-ultra", "The definitive Android flagship with Galaxy AI and titanium frame.", ["200MP Quad-Telephoto zoom", "Titanium armor frame", "Built-in S Pen"], [
        ("Titanium Gray 256GB", 129999, 139999, {"color":"Titanium Gray","storage":"256GB"}, "products/galaxy-s24-ultra.jpg"),
        ("Titanium Black 512GB", 141999, 151999, {"color":"Titanium Black","storage":"512GB"}, "products/galaxy-s24-ultra-black.jpg"),
        ("Titanium Violet 1TB", 165999, 179999, {"color":"Titanium Violet","storage":"1TB"}, "products/galaxy-s24-ultra-violet.jpg")
    ]),
    ("Smartphones", "Samsung", "Galaxy S24 Plus", "galaxy-s24-plus", "Flagship performance, large QHD+ display, and slim bezels.", ["6.7\" Dynamic AMOLED 2X QHD+", "4,900 mAh battery", "Snapdragon 8 Gen 3"], [
        ("Onyx Black 256GB", 99999, 109999, {"color":"Onyx Black","storage":"256GB"}, "products/galaxy-s24-plus.jpg"),
        ("Marble Gray 512GB", 111999, 121999, {"color":"Marble Gray","storage":"512GB"}, "products/galaxy-s24-plus.jpg")
    ]),
    ("Smartphones", "Samsung", "Galaxy S24", "galaxy-s24-base", "Compact powerhouse with all the Galaxy AI flagship smarts.", ["6.2\" FHD+ 120Hz AMOLED", "Armor Aluminum 2.0", "Triple camera system"], [
        ("Cobalt Violet 128GB", 79999, 85999, {"color":"Cobalt Violet","storage":"128GB"}, "products/galaxy-s24-base.jpg"),
        ("Amber Yellow 256GB", 85999, 91999, {"color":"Amber Yellow","storage":"256GB"}, "products/galaxy-s24-base.jpg")
    ]),
    ("Smartphones", "Samsung", "Galaxy Z Fold6", "galaxy-z-fold-6", "Ultra-premium foldable phone transforming into a 7.6-inch tablet.", ["7.6\" Dynamic AMOLED 2X main screen", "Enhanced zero-gap hinge", "Dual preview camera features"], [
        ("Silver Shadow 256GB", 189999, 199999, {"color":"Silver Shadow","storage":"256GB"}, "products/galaxy-z-fold-6.jpg"),
        ("Navy 512GB", 201999, 214999, {"color":"Navy","storage":"512GB"}, "products/galaxy-z-fold-6.jpg")
    ]),
    ("Smartphones", "Samsung", "Galaxy Z Flip6", "galaxy-z-flip-6", "Pocket-sized iconic flip phone with 50MP camera and FlexWindow.", ["3.4\" FlexWindow cover display", "50MP wide camera with Auto Zoom", "Vapor chamber cooling"], [
        ("Mint 256GB", 109999, 119999, {"color":"Mint","storage":"256GB"}, "products/galaxy-z-flip-6.jpg"),
        ("Silver Shadow 512GB", 121999, 131999, {"color":"Silver Shadow","storage":"512GB"}, "products/galaxy-z-flip-6.jpg")
    ]),
    ("Smartphones", "Apple", "iPhone 15 Pro Max", "iphone-15-pro-max", "Forged in titanium with 5x optical telephoto and A17 Pro chip.", ["Grade 5 Titanium design", "Action button customizable trigger", "USB-C with 10Gbps USB 3 speeds"], [
        ("Natural Titanium 256GB", 119900, 129900, {"color":"Natural Titanium","storage":"256GB"}, "products/iphone-15-pro-max.jpg"),
        ("Blue Titanium 512GB", 139900, 149900, {"color":"Blue Titanium","storage":"512GB"}, "products/iphone-15-pro-max-blue.jpg"),
        ("Black Titanium 1TB", 159900, 169900, {"color":"Black Titanium","storage":"1TB"}, "products/iphone-15-pro-max-black.jpg")
    ]),
    ("Smartphones", "Apple", "iPhone 15 Pro", "iphone-15-pro", "Compact 6.1-inch titanium pro phone with ProMotion display.", ["6.1\" Super Retina XDR with ProMotion", "A17 Pro graphics processor", "48MP Main camera with spatial video"], [
        ("White Titanium 128GB", 99900, 109900, {"color":"White Titanium","storage":"128GB"}, "products/iphone-15-pro.jpg"),
        ("Natural Titanium 256GB", 109900, 119900, {"color":"Natural Titanium","storage":"256GB"}, "products/iphone-15-pro.jpg")
    ]),
    ("Smartphones", "Apple", "iPhone 15 Plus", "iphone-15-plus", "Huge 6.7-inch display with industry-leading battery longevity.", ["Dynamic Island notifications", "48MP Main camera with 2x Telephoto", "Color-infused back glass"], [
        ("Pink 128GB", 89900, 94900, {"color":"Pink","storage":"128GB"}, "products/iphone-15-plus.jpg"),
        ("Blue 256GB", 99900, 104900, {"color":"Blue","storage":"256GB"}, "products/iphone-15-plus.jpg")
    ]),
    ("Smartphones", "Apple", "iPhone 15", "iphone-15-base", "Dynamic Island, 48MP high-resolution camera, and USB-C.", ["Dynamic Island", "A16 Bionic chip", "All-day battery life"], [
        ("Black 128GB", 79900, 82900, {"color":"Black","storage":"128GB"}, "products/iphone-15-base.jpg"),
        ("Green 256GB", 89900, 92900, {"color":"Green","storage":"256GB"}, "products/iphone-15-base.jpg")
    ]),
    ("Smartphones", "Sony", "Sony Xperia 1 VI", "sony-xperia-1-vi", "Pro photographer smartphone with 85-170mm true optical zoom.", ["BRAVIA powered OLED display", "Exmor T for mobile sensor", "Dedicated two-stage shutter button"], [
        ("Black 256GB", 139900, 149900, {"color":"Black","storage":"256GB"}, "products/sony-xperia-1-vi.jpg"),
        ("Platinum Silver 512GB", 154900, 164900, {"color":"Platinum Silver","storage":"512GB"}, "products/sony-xperia-1-vi.jpg")
    ]),
    ("Smartphones", "Sony", "Sony Xperia 5 V", "sony-xperia-5-v", "Compact cinematic smartphone with exceptional 2-day battery.", ["6.1\" 120Hz HDR OLED", "High performance front stereo speakers", "3.5mm hi-res headphone jack"], [
        ("Blue 128GB", 89900, 99900, {"color":"Blue","storage":"128GB"}, "products/sony-xperia-5-v.jpg"),
        ("Black 128GB", 89900, 99900, {"color":"Black","storage":"128GB"}, "products/sony-xperia-5-v.jpg")
    ]),
    ("Smartphones", "Samsung", "Galaxy A55 5G", "galaxy-a55-5g", "Premium metal frame smartphone with Knox Vault security.", ["6.6\" Super AMOLED 120Hz", "50MP OIS camera", "IP67 water and dust resistance"], [
        ("Awesome Navy 128GB", 44900, 49900, {"color":"Awesome Navy","storage":"128GB"}, "products/galaxy-a55-5g.jpg"),
        ("Awesome Iceblue 256GB", 49900, 54900, {"color":"Awesome Iceblue","storage":"256GB"}, "products/galaxy-a55-5g.jpg")
    ]),
    ("Smartphones", "Samsung", "Galaxy A35 5G", "galaxy-a35-5g", "Vibrant display and nightography camera for everyday brilliance.", ["6.6\" FHD+ 120Hz display", "5,000mAh 2-day battery", "Gorilla Glass Victus+ front"], [
        ("Awesome Lilac 128GB", 35900, 39900, {"color":"Awesome Lilac","storage":"128GB"}, "products/galaxy-a35-5g.jpg")
    ]),
    ("Smartphones", "Apple", "iPhone SE (3rd Gen)", "iphone-se-3rd-gen", "Classic design packed with blazing A15 Bionic performance.", ["A15 Bionic chip", "Touch ID Home button", "5G cellular connectivity"], [
        ("Midnight 64GB", 42900, 44900, {"color":"Midnight","storage":"64GB"}, "products/iphone-se-3rd-gen.jpg"),
        ("Starlight 128GB", 47900, 49900, {"color":"Starlight","storage":"128GB"}, "products/iphone-se-3rd-gen.jpg")
    ]),

    # Audio & Headphones (14)
    ("Audio & Headphones", "Sony", "Sony WH-1000XM5 Wireless Noise Canceling", "sony-wh-1000xm5", "Industry benchmark active noise canceling with 8 microphones.", ["Auto NC Optimizer", "30-hour battery life", "Precise Voice Pickup with 4 beamforming mics"], [
        ("Black", 39800, 42000, {"color":"Black"}, "products/sony-wh-1000xm5.jpg"),
        ("Silver", 39800, 42000, {"color":"Silver"}, "products/sony-wh-1000xm5-silver.jpg"),
        ("Midnight Blue", 39800, 42000, {"color":"Midnight Blue"}, "products/sony-wh-1000xm5-midnight-blue.jpg")
    ]),
    ("Audio & Headphones", "Bose", "Bose QuietComfort Ultra Headphones", "bose-qc-ultra-headphones", "Spatial audio with world-class noise cancellation tailored to your ears.", ["Bose Immersive Audio", "CustomTune sound calibration", "24-hour battery with USB-C fast charge"], [
        ("Black", 42900, 44900, {"color":"Black"}, "products/bose-qc-ultra-headphones.jpg"),
        ("White Smoke", 42900, 44900, {"color":"White Smoke"}, "products/bose-qc-ultra-white.jpg"),
        ("Sandstone", 42900, 44900, {"color":"Sandstone"}, "products/bose-qc-ultra-headphones.jpg")
    ]),
    ("Audio & Headphones", "Apple", "AirPods Max", "airpods-max", "Computational audio with custom acoustic design and spatial audio.", ["Apple H1 headphone chips", "Knit-mesh canopy headband", "Digital Crown for precision volume"], [
        ("Space Gray", 54900, 57900, {"color":"Space Gray"}, "products/airpods-max.jpg"),
        ("Silver", 54900, 57900, {"color":"Silver"}, "products/airpods-max-silver.jpg"),
        ("Sky Blue", 54900, 57900, {"color":"Sky Blue"}, "products/airpods-max-sky-blue.jpg")
    ]),
    ("Audio & Headphones", "Apple", "AirPods Pro (2nd Gen) USB-C", "airpods-pro-2-usb-c", "Up to 2x more Active Noise Cancellation with Adaptive Audio.", ["H2 chip for smarter noise cancellation", "Dust, sweat, and water resistant (IP54)", "MagSafe Case with speaker and lanyard loop"], [
        ("White", 24900, 26900, {"color":"White"}, "products/airpods-pro-2-usb-c.jpg")
    ]),
    ("Audio & Headphones", "Bose", "Bose QuietComfort Ultra Earbuds", "bose-qc-ultra-earbuds", "Next-level spatial earbuds with groundbreaking noise cancellation.", ["CustomTune technology", "Immersion Mode 3D sound", "Up to 6 hours listening (24h with case)"], [
        ("Black", 29900, 31900, {"color":"Black"}, "products/bose-qc-ultra-earbuds.jpg"),
        ("White Smoke", 29900, 31900, {"color":"White Smoke"}, "products/bose-qc-ultra-white.jpg")
    ]),
    ("Audio & Headphones", "Sony", "Sony WF-1000XM5 True Wireless Earbuds", "sony-wf-1000xm5", "Best noise canceling earbuds with Dynamic Driver X.", ["Integrated Processor V2 and QN2e chip", "AI-based noise reduction algorithm", "Multipoint connection up to two devices"], [
        ("Black", 29800, 31900, {"color":"Black"}, "products/sony-wf-1000xm5.jpg"),
        ("Silver", 29800, 31900, {"color":"Silver"}, "products/sony-wf-1000xm5.jpg")
    ]),
    ("Audio & Headphones", "Sony", "Sony ULT WEAR Bass Headphones", "sony-ult-wear", "Massive bass button with premium noise cancellation.", ["ULT button for dual deep bass modes", "Integrated Processor V1", "30-hour battery with quick charging"], [
        ("Black", 19900, 22900, {"color":"Black"}, "products/sony-ult-wear.jpg"),
        ("Off White", 19900, 22900, {"color":"Off White"}, "products/sony-ult-wear.jpg")
    ]),
    ("Audio & Headphones", "Bose", "Bose SoundLink Max Portable Speaker", "bose-soundlink-max", "Big stereo sound and deep bass that turns any room into a party.", ["Up to 20 hours battery life", "IP67 waterproof and dustproof", "Removable rope carry handle"], [
        ("Black", 39900, 42900, {"color":"Black"}, "products/bose-soundlink-max.jpg"),
        ("Blue Dusk", 39900, 42900, {"color":"Blue Dusk"}, "products/bose-soundlink-max.jpg")
    ]),
    ("Audio & Headphones", "Bose", "Bose QuietComfort Headphones", "bose-quietcomfort-45", "Legendary noise cancellation, comfortable lightweight fit.", ["Quiet and Aware listening modes", "TriPort acoustic architecture", "24 hours on a single charge"], [
        ("Triple Black", 34900, 37900, {"color":"Triple Black"}, "products/bose-quietcomfort-45.jpg"),
        ("White Smoke", 34900, 37900, {"color":"White Smoke"}, "products/bose-qc-ultra-white.jpg")
    ]),
    ("Audio & Headphones", "Sony", "Sony WH-CH720N Noise Canceling", "sony-wh-ch720n", "Lightweight wireless noise canceling headphones for everyday use.", ["Lightest overhead wireless noise canceling model", "Integrated Processor V1", "35 hours battery life with quick charge"], [
        ("Black", 14800, 16900, {"color":"Black"}, "products/sony-wh-ch720n.jpg"),
        ("Blue", 14800, 16900, {"color":"Blue"}, "products/sony-wh-ch720n.jpg")
    ]),
    ("Audio & Headphones", "Apple", "AirPods (3rd Gen)", "airpods-3rd-gen", "Personalized Spatial Audio with dynamic head tracking.", ["Sweat and water resistant", "Force sensor touch controls", "Up to 30 hours total listening time"], [
        ("White", 16900, 17900, {"color":"White"}, "products/airpods-3rd-gen.jpg")
    ]),
    ("Audio & Headphones", "Sony", "Sony SRS-XB100 Compact Speaker", "sony-srs-xb100", "Small speaker with big sound and Sound Diffusion Processor.", ["Extra Bass with passive radiator", "16-hour battery life", "Multiway strap for easy carrying"], [
        ("Black", 5900, 6900, {"color":"Black"}, "products/sony-srs-xb100.jpg"),
        ("Light Green", 5900, 6900, {"color":"Light Green"}, "products/sony-srs-xb100.jpg")
    ]),
    ("Audio & Headphones", "Bose", "Bose SoundLink Flex Bluetooth Speaker", "bose-soundlink-flex", "PositionIQ technology automatically optimizes sound orientation.", ["Clear sound with deep bass", "Rugged IP67 silicone waterproof body", "12 hours per charge"], [
        ("Black", 14900, 16900, {"color":"Black"}, "products/bose-soundlink-flex.jpg"),
        ("Carmine Red", 14900, 16900, {"color":"Carmine Red"}, "products/bose-soundlink-flex.jpg")
    ]),
    ("Audio & Headphones", "Sony", "Sony LinkBuds S Noise Canceling Earbuds", "sony-linkbuds-s", "Never off smart earbuds that connect your online and offline worlds.", ["Smallest and lightest Hi-Res noise canceling earbuds", "Auto Play automated soundtrack", "Speak-to-Chat pause function"], [
        ("Earth Blue", 19800, 21900, {"color":"Earth Blue"}, "products/sony-linkbuds-s.jpg"),
        ("Black", 19800, 21900, {"color":"Black"}, "products/sony-linkbuds-s.jpg")
    ]),

    # Shoes & Footwear (14)
    ("Shoes", "Nike", "Nike Air Force 1 '07", "nike-air-force-1-07", "The basketball legend that adds fresh style to classic comfort.", ["Stitched leather overlays", "Nike Air encapsulated cushioning", "Perforations on the toe"], [
        ("Triple White 10", 11500, 12500, {"color":"Triple White","size":"10"}, "products/nike-air-force-1-07.jpg"),
        ("Triple Black 10.5", 11500, 12500, {"color":"Triple Black","size":"10.5"}, "products/nike-air-force-1-black.jpg"),
        ("White/University Red 11", 11500, 12500, {"color":"White/Red","size":"11"}, "products/nike-air-force-1-07.jpg")
    ]),
    ("Shoes", "Nike", "Nike Dunk Low Retro", "nike-dunk-low-retro", "Created for the hardwood, adopted by skate and street culture.", ["Crisp leather upper", "Padded low-cut collar", "Rubber cupsole traction pattern"], [
        ("Panda White/Black 9.5", 11500, 13000, {"color":"White/Black","size":"9.5"}, "products/nike-dunk-low-retro.jpg"),
        ("Panda White/Black 10", 11500, 13000, {"color":"White/Black","size":"10"}, "products/nike-dunk-low-retro.jpg"),
        ("Panda White/Black 11", 11500, 13000, {"color":"White/Black","size":"11"}, "products/nike-dunk-low-retro.jpg")
    ]),
    ("Shoes", "Adidas", "Adidas Ultraboost Light", "adidas-ultraboost-light", "Epic energy return with the lightest BOOST midsole ever engineered.", ["Light BOOST midsole technology", "Primeknit+ foot-hugging upper", "Continental Better Rubber outsole"], [
        ("Core Black 10", 19000, 21000, {"color":"Core Black","size":"10"}, "products/adidas-ultraboost-light.jpg"),
        ("Cloud White 10.5", 19000, 21000, {"color":"Cloud White","size":"10.5"}, "products/adidas-ultraboost-light.jpg"),
        ("Lucid Lemon 11", 19000, 21000, {"color":"Lucid Lemon","size":"11"}, "products/adidas-ultraboost-light.jpg")
    ]),
    ("Shoes", "Adidas", "Adidas Samba Classic", "adidas-samba-classic", "Timeless indoor soccer icon turned worldwide streetwear staple.", ["Full grain leather upper with suede toe cap", "Die-cut EVA insole", "Non-marking gum rubber outsole"], [
        ("Core Black/White 9", 9000, 10000, {"color":"Black/White","size":"9"}, "products/adidas-samba-classic.jpg"),
        ("Cloud White/Black 10", 9000, 10000, {"color":"White/Black","size":"10"}, "products/adidas-samba-white-black.jpg"),
        ("Cloud White/Black 11", 9000, 10000, {"color":"White/Black","size":"11"}, "products/adidas-samba-white-black.jpg")
    ]),
    ("Shoes", "Adidas", "Adidas Gazelle Bold Shoes", "adidas-gazelle-bold", "Classic low-top retro sneaker elevated with a stacked triple platform.", ["Triple-stacked gum rubber platform", "Supple suede upper", "Signature contrast 3-Stripes"], [
        ("Core Black/White 8", 12000, 13500, {"color":"Black/White","size":"8"}, "products/adidas-gazelle-bold.jpg"),
        ("Wonder Clay 8.5", 12000, 13500, {"color":"Wonder Clay","size":"8.5"}, "products/adidas-gazelle-bold.jpg")
    ]),
    ("Shoes", "Nike", "Nike Air Max Plus (Tn)", "nike-air-max-plus", "Tuned Air cushioning with the iconic flaming cage upper.", ["Tuned Air units in heel and forefoot", "Whale tail shank midfoot support", "Gradient synthetic mesh upper"], [
        ("Sunset Orange 10", 18000, 19500, {"color":"Sunset Orange","size":"10"}, "products/nike-air-max-plus.jpg"),
        ("Triple Black 10.5", 18000, 19500, {"color":"Triple Black","size":"10.5"}, "products/nike-air-force-1-black.jpg")
    ]),
    ("Shoes", "Nike", "Nike Pegasus 41 Road Running", "nike-pegasus-41", "Responsive daily trainer featuring dual Air Zoom units and ReactX foam.", ["Upgraded ReactX foam with 13% more energy return", "Dual Air Zoom pods", "Engineered breathable mesh upper"], [
        ("Volt/Black 10", 14000, 15000, {"color":"Volt/Black","size":"10"}, "products/nike-pegasus-41.jpg"),
        ("Black/White 11", 14000, 15000, {"color":"Black/White","size":"11"}, "products/nike-pegasus-41.jpg")
    ]),
    ("Shoes", "Adidas", "Adidas Stan Smith", "adidas-stan-smith", "Clean, minimalist tennis silhouette that defines effortless casual style.", ["Smooth synthetic leather upper", "Perforated 3-Stripes detailing", "OrthoLite sockliner comfort"], [
        ("Cloud White/Green 9.5", 10000, 11000, {"color":"White/Green","size":"9.5"}, "products/adidas-stan-smith.jpg"),
        ("Cloud White/Navy 10.5", 10000, 11000, {"color":"White/Navy","size":"10.5"}, "products/adidas-stan-smith.jpg")
    ]),
    ("Shoes", "Nike", "Nike Invincible 3 Cushion Running", "nike-invincible-3", "Max-cushion road running shoe with thick ZoomX foam stack.", ["Full-length ZoomX foam for softest landings", "Evolved Flyknit upper", "Wider rocker sole for transitions"], [
        ("White/Cobalt 10", 18000, 19500, {"color":"White/Cobalt","size":"10"}, "products/nike-invincible-3.jpg")
    ]),
    ("Shoes", "Adidas", "Adidas Campus 00s", "adidas-campus-00s", "Chunky skate proportions inspired by the iconic early 2000s era.", ["Padded tongue and collar", "Suede upper with contrast stitching", "Fat laces and off-white midsole"], [
        ("Core Black 9", 11000, 12000, {"color":"Core Black","size":"9"}, "products/adidas-campus-00s.jpg"),
        ("Grey/White 10", 11000, 12000, {"color":"Grey/White","size":"10"}, "products/adidas-campus-00s.jpg")
    ]),
    ("Shoes", "Nike", "Nike Air Jordan 1 Retro High OG", "nike-air-jordan-1-high", "The one that started it all. Premium leather and timeless Jordan heritage.", ["Genuine leather and textile construction", "Air-Sole unit in heel", "Solid rubber outsole with deep flex grooves"], [
        ("Chicago Red/Black 10.5", 18000, 20000, {"color":"Chicago","size":"10.5"}, "products/nike-air-jordan-1-high.jpg"),
        ("Royal Blue 11", 18000, 20000, {"color":"Royal Blue","size":"11"}, "products/nike-air-jordan-1-high.jpg")
    ]),
    ("Shoes", "Nike", "Nike Metcon 9 Training Shoes", "nike-metcon-9", "Built for intense lifters with enlarged Hyperlift heel plates.", ["Larger Hyperlift plate for heavy squats", "Rope wrap extended rubber walls", "Lightweight breathable mesh"], [
        ("Black/Smoke Grey 10", 15000, 16500, {"color":"Black/Smoke Grey","size":"10"}, "products/nike-metcon-9.jpg")
    ]),
    ("Shoes", "Adidas", "Adidas Adizero Adios Pro 3", "adidas-adizero-adios-pro-3", "Marathon winning racing shoe with carbon-infused EnergyRods.", ["Lightstrike Pro dual-layer foam", "ENERGYRODS 2.0 carbon spine", "Ultra-lightweight upper with continental rubber"], [
        ("Solar Orange 10", 25000, 27500, {"color":"Solar Orange","size":"10"}, "products/adidas-adizero-adios-pro-3.jpg")
    ]),
    ("Shoes", "Nike", "Nike Vaporfly 3 Road Racing", "nike-vaporfly-3", "The all-around road racing shoe engineered to break personal bests.", ["Full-length carbon fiber Flyplate", "ZoomX foam cushioning", "Thin waffle outsole for traction"], [
        ("Electric Volt 10.5", 26000, 28000, {"color":"Electric Volt","size":"10.5"}, "products/nike-vaporfly-3.jpg")
    ]),

    # Outerwear & Apparel (14)
    ("Outerwear", "Nike", "Nike Tech Fleece Windrunner Full-Zip", "nike-tech-fleece-windrunner", "Engineered fleece providing lightweight warmth with smooth double-sided texture.", ["Chevron chest design lines", "Zippered sleeve stash pocket", "Four-panel ergonomic hood"], [
        ("Black M", 14500, 16000, {"color":"Black","size":"M"}, "products/nike-tech-fleece-windrunner.jpg"),
        ("Dark Grey Heather L", 14500, 16000, {"color":"Dark Grey Heather","size":"L"}, "products/nike-tech-fleece-windrunner.jpg"),
        ("Khaki XL", 14500, 16000, {"color":"Khaki","size":"XL"}, "products/nike-tech-fleece-windrunner.jpg")
    ]),
    ("Outerwear", "Adidas", "Adidas Originals Beckenbauer Track Jacket", "adidas-beckenbauer-track-jacket", "Archival 60s soccer heritage jacket with ribbed stand-up collar.", ["Heavyweight cotton-blend pique", "Two-way full zip with stand-up collar", "Classic embroidered Trefoil on chest"], [
        ("Night Indigo M", 9000, 10000, {"color":"Night Indigo","size":"M"}, "products/adidas-beckenbauer-track-jacket.jpg"),
        ("Black L", 9000, 10000, {"color":"Black","size":"L"}, "products/adidas-beckenbauer-track-jacket.jpg")
    ]),
    ("Outerwear", "Nike", "Nike ACG Storm-FIT Cascade Rains Jacket", "nike-acg-storm-fit-jacket", "Fully seam-sealed waterproof jacket built for outdoor expeditions.", ["Storm-FIT ADV waterproof technology", "Adjustable bungee cord hood", "Packable into side pocket"], [
        ("Cacao Wow L", 18500, 21000, {"color":"Cacao Wow","size":"L"}, "products/nike-acg-storm-fit-jacket.jpg"),
        ("Black XL", 18500, 21000, {"color":"Black","size":"XL"}, "products/nike-acg-storm-fit-jacket.jpg")
    ]),
    ("Outerwear", "Adidas", "Adidas Terrex Multi 2.5L Rain Jacket", "adidas-terrex-rain-jacket", "Lightweight, packable waterproof jacket with breathable RAIN.RDY technology.", ["RAIN.RDY advanced waterproof barrier", "Fitted elastic hood and cuffs", "100% recycled polyester plain weave"], [
        ("Preloved Ink M", 13000, 14500, {"color":"Preloved Ink","size":"M"}, "products/adidas-terrex-rain-jacket.jpg")
    ]),
    ("Outerwear", "Nike", "Nike Sportswear Club Fleece Pullover", "nike-club-fleece-pullover", "Cozy brushed fleece hoodie that provides consistent everyday comfort.", ["Brushed-back fleece for softness", "Ribbed hem and cuffs", "Kangaroo front pocket"], [
        ("Black L", 6500, 7500, {"color":"Black","size":"L"}, "products/nike-club-fleece-pullover.jpg"),
        ("Grey Heather M", 6500, 7500, {"color":"Grey Heather","size":"M"}, "products/nike-club-fleece-pullover.jpg")
    ]),
    ("Outerwear", "Adidas", "Adidas Originals Firebird Track Top", "adidas-firebird-track-top", "The definitive 90s retro track top in glossy tricot fabric.", ["Glossy tricot fabric construction", "Full zip with stand-up collar", "Zip side pockets for essentials"], [
        ("Green M", 8500, 9500, {"color":"Green","size":"M"}, "products/adidas-firebird-track-top.jpg"),
        ("Black L", 8500, 9500, {"color":"Black","size":"L"}, "products/adidas-firebird-track-top.jpg")
    ]),
    ("Outerwear", "Nike", "Nike Tech Fleece Jogger Pants", "nike-tech-fleece-joggers", "Tapered modern joggers with elongated zippered pocket.", ["Thermal insulation without bulk", "Ribbed ankle cuffs", "Streamlined tailored fit"], [
        ("Black M", 12500, 14000, {"color":"Black","size":"M"}, "products/nike-tech-fleece-joggers.jpg"),
        ("Dark Grey Heather L", 12500, 14000, {"color":"Dark Grey Heather","size":"L"}, "products/nike-tech-fleece-joggers.jpg")
    ]),
    ("Outerwear", "Adidas", "Adidas Essentials 3-Stripes Warm-Up Top", "adidas-essentials-3-stripes-top", "Casual heritage warm-up jacket featuring contrast sleeve stripes.", ["Recycled polyester tricot", "Full front zip with high collar", "Side hand pockets"], [
        ("Legend Ink L", 6000, 7000, {"color":"Legend Ink","size":"L"}, "products/adidas-essentials-3-stripes-top.jpg")
    ]),
    ("Outerwear", "Nike", "Nike Pro Dri-FIT Tight Long Sleeve", "nike-pro-dri-fit-ls", "Compression base layer that wicks away sweat during heavy training.", ["Dri-FIT moisture wicking tech", "Flatlock seams reduce chafing", "Ventilated mesh underarms"], [
        ("Black M", 4000, 4800, {"color":"Black","size":"M"}, "products/nike-pro-dri-fit-ls.jpg"),
        ("White L", 4000, 4800, {"color":"White","size":"L"}, "products/nike-pro-dri-fit-ls.jpg")
    ]),
    ("Outerwear", "Adidas", "Adidas Own The Run Hooded Wind Jacket", "adidas-own-the-run-jacket", "Wind-resistant and water-repellent running shell with reflective details.", ["WIND.RDY blocking fabric", "Chest zip pocket for phone", "360-degree reflectivity for night safety"], [
        ("Black M", 8000, 9000, {"color":"Black","size":"M"}, "products/adidas-own-the-run-jacket.jpg")
    ]),
    ("Outerwear", "Nike", "Nike Dri-FIT UV Miler Running Tee", "nike-dri-fit-miler-tee", "Lightweight running shirt with UVA and UVB sun protection.", ["UVA and UVB blocking knit fabric", "Seams rolled back for shoulder motion", "Ultra-breathable feel"], [
        ("Game Royal L", 3800, 4500, {"color":"Game Royal","size":"L"}, "products/nike-dri-fit-miler-tee.jpg"),
        ("Black XL", 3800, 4500, {"color":"Black","size":"XL"}, "products/nike-dri-fit-miler-tee.jpg")
    ]),
    ("Outerwear", "Adidas", "Adidas Terrex Agravic Windweave Pro", "adidas-terrex-windweave-pro", "Ultralight trail running shell with variable breathability zones.", ["Windweave zoned body mapping", "Packable into own chest pocket", "DWR water-repellent coating"], [
        ("Impact Orange M", 16000, 18000, {"color":"Impact Orange","size":"M"}, "products/adidas-terrex-windweave-pro.jpg")
    ]),
    ("Outerwear", "Nike", "Nike ACG 'Insects' Dri-FIT ADV Top", "nike-acg-insects-top", "Breathable knit trail top tested in the wilderness of Oregon.", ["Dri-FIT ADV engineered ventilation", "All-over printed topographic graphic", "Ribbed crewneck collar"], [
        ("Anthracite L", 7000, 8000, {"color":"Anthracite","size":"L"}, "products/nike-acg-insects-top.jpg")
    ]),
    ("Outerwear", "Adidas", "Adidas Tiro 23 League Training Pants", "adidas-tiro-23-pants", "The classic slim soccer training pants with ankle zips.", ["AEROREADY sweat-wicking fabric", "Ankle zips for easy on/off over cleats", "Zip pockets for securely holding items"], [
        ("Black/White M", 5500, 6500, {"color":"Black/White","size":"M"}, "products/adidas-tiro-23-pants.jpg")
    ]),

    # Tablets & Wearables (14)
    ("Tablets & Wearables", "Apple", "iPad Pro 13\" M4 OLED", "ipad-pro-13-m4-oled", "Breakthrough Ultra Retina XDR tandem OLED display in a 5.1mm body.", ["Tandem OLED Ultra Retina XDR", "Apple M4 10-core silicon", "Supports Apple Pencil Pro and Magic Keyboard"], [
        ("Space Black 256GB WiFi", 129900, 139900, {"color":"Space Black","storage":"256GB"}, "products/ipad-pro-13-m4-oled.jpg"),
        ("Silver 512GB WiFi+5G", 169900, 179900, {"color":"Silver","storage":"512GB"}, "products/ipad-air-13-m2.jpg"),
        ("Space Black 1TB Nano-Texture", 209900, 229900, {"color":"Space Black","storage":"1TB"}, "products/ipad-pro-13-m4-oled.jpg")
    ]),
    ("Tablets & Wearables", "Apple", "iPad Air 13\" M2", "ipad-air-13-m2", "Spacious 13-inch Liquid Retina display powered by the M2 chip.", ["13\" Liquid Retina display", "Apple M2 silicon chip", "Landscape 12MP front camera with Center Stage"], [
        ("Space Gray 128GB", 79900, 84900, {"color":"Space Gray","storage":"128GB"}, "products/ipad-air-13-m2.jpg"),
        ("Starlight 256GB", 89900, 94900, {"color":"Starlight","storage":"256GB"}, "products/ipad-air-13-m2.jpg")
    ]),
    ("Tablets & Wearables", "Apple", "iPad mini (6th Gen)", "ipad-mini-6", "Mega power in mini size with 8.3-inch Liquid Retina display.", ["8.3\" Liquid Retina screen", "A15 Bionic chip with Neural Engine", "Supports Apple Pencil (2nd gen)"], [
        ("Space Gray 64GB", 49900, 52900, {"color":"Space Gray","storage":"64GB"}, "products/ipad-mini-6.jpg"),
        ("Purple 256GB", 64900, 67900, {"color":"Purple","storage":"256GB"}, "products/ipad-mini-6.jpg")
    ]),
    ("Tablets & Wearables", "Apple", "Apple Watch Ultra 2", "apple-watch-ultra-2", "Rugged 49mm titanium smartwatch for extreme outdoor adventure.", ["3000 nits brightest display", "Precision dual-frequency GPS", "Up to 72 hours Low Power Mode"], [
        ("Titanium Orange Ocean Band", 79900, 84900, {"color":"Titanium","size":"49mm"}, "products/apple-watch-ultra-2.jpg"),
        ("Titanium Black Trail Loop", 79900, 84900, {"color":"Titanium","size":"49mm"}, "products/apple-watch-ultra-2.jpg")
    ]),
    ("Tablets & Wearables", "Apple", "Apple Watch Series 9", "apple-watch-series-9", "Powerful S9 SiP with magical Double Tap gesture control.", ["Double tap pinch gesture control", "Edge-to-edge Always-On Retina display", "ECG, Blood Oxygen, and Temperature sensing"], [
        ("Midnight Aluminum 45mm", 42900, 45900, {"color":"Midnight","size":"45mm"}, "products/apple-watch-series-9.jpg"),
        ("Silver Aluminum 41mm", 39900, 42900, {"color":"Silver","size":"41mm"}, "products/apple-watch-series-9.jpg")
    ]),
    ("Tablets & Wearables", "Samsung", "Galaxy Tab S9 Ultra", "galaxy-tab-s9-ultra", "Massive 14.6-inch Dynamic AMOLED 2X tablet with IP68 water resistance.", ["14.6\" Dynamic AMOLED 120Hz", "Included IP68 S Pen with low latency", "Snapdragon 8 Gen 2 for Galaxy"], [
        ("Graphite 256GB", 119999, 129999, {"color":"Graphite","storage":"256GB"}, "products/galaxy-tab-s9-ultra.jpg"),
        ("Beige 512GB", 131999, 141999, {"color":"Beige","storage":"512GB"}, "products/galaxy-tab-s9-ultra.jpg")
    ]),
    ("Tablets & Wearables", "Samsung", "Galaxy Tab S9 FE Plus", "galaxy-tab-s9-fe-plus", "Creative powerhouse tablet with large 12.4\" display and S Pen.", ["12.4\" 90Hz display", "IP68 water and dust resistant", "Expandable microSD up to 1TB"], [
        ("Gray 128GB", 59999, 64999, {"color":"Gray","storage":"128GB"}, "products/galaxy-tab-s9-fe-plus.jpg"),
        ("Mint 256GB", 69999, 74999, {"color":"Mint","storage":"256GB"}, "products/galaxy-tab-s9-fe-plus.jpg")
    ]),
    ("Tablets & Wearables", "Samsung", "Galaxy Watch Ultra", "galaxy-watch-ultra", "Titanium cushion smartwatch built for ultra-marathons and ocean swimming.", ["Titanium grade 4 cushion case", "Dual-frequency GPS (L1+L5)", "10ATM water resistance up to 100m"], [
        ("Titanium Gray 47mm", 64999, 69999, {"color":"Titanium Gray","size":"47mm"}, "products/galaxy-watch-ultra.jpg"),
        ("Titanium White 47mm", 64999, 69999, {"color":"Titanium White","size":"47mm"}, "products/galaxy-watch-ultra.jpg")
    ]),
    ("Tablets & Wearables", "Samsung", "Galaxy Watch 7", "galaxy-watch-7", "3nm processor smartwatch with advanced metabolic and sleep AI tracking.", ["3nm Exynos processor", "BioActive Sensor with AGEs monitoring", "Dual-frequency GPS precision"], [
        ("Green 44mm", 32999, 34999, {"color":"Green","size":"44mm"}, "products/galaxy-watch-7.jpg"),
        ("Cream 40mm", 29999, 31999, {"color":"Cream","size":"40mm"}, "products/galaxy-watch-7.jpg")
    ]),
    ("Tablets & Wearables", "Samsung", "Galaxy Watch FE", "galaxy-watch-fe", "Essential fitness and wellness tracking in an iconic sapphire crystal design.", ["Sapphire crystal glass screen", "Body Composition BIA sensor", "Heart rate zones coaching"], [
        ("Black 40mm", 19999, 21999, {"color":"Black","size":"40mm"}, "products/galaxy-watch-fe.jpg")
    ]),
    ("Tablets & Wearables", "Apple", "Apple Watch SE (2nd Gen)", "apple-watch-se-2", "Essential health and activity features at an unbeatable value.", ["Crash Detection and Fall Detection", "Retina display up to 1000 nits", "Swimproof up to 50 meters (WR50)"], [
        ("Midnight Aluminum 44mm", 27900, 29900, {"color":"Midnight","size":"44mm"}, "products/apple-watch-se-2.jpg"),
        ("Starlight Aluminum 40mm", 24900, 26900, {"color":"Starlight","size":"40mm"}, "products/apple-watch-se-2.jpg")
    ]),
    ("Tablets & Wearables", "Samsung", "Galaxy Tab A9 Plus", "galaxy-tab-a9-plus", "Family tablet with 11-inch screen, Quad speakers, and multi-active windows.", ["11\" 90Hz smooth display", "Quad speakers powered by Dolby Atmos", "Samsung Kids mode parental controls"], [
        ("Graphite 64GB", 21999, 23999, {"color":"Graphite","storage":"64GB"}, "products/galaxy-tab-a9-plus.jpg"),
        ("Silver 128GB", 26999, 28999, {"color":"Silver","storage":"128GB"}, "products/galaxy-tab-a9-plus.jpg")
    ]),
    ("Tablets & Wearables", "Apple", "Apple Pencil Pro", "apple-pencil-pro", "Advanced stylus with barrel roll, squeeze gesture, and haptic feedback.", ["Squeeze sensor palette trigger", "Barrel roll gyroscopic control", "Find My tracking support"], [
        ("White", 12900, 13900, {"color":"White"}, "products/apple-pencil-pro.jpg")
    ]),
    ("Tablets & Wearables", "Apple", "Magic Keyboard for iPad Pro 13\"", "magic-keyboard-ipad-pro-13", "Sleek floating cantilever design with glass haptic trackpad.", ["Aluminum palm rest", "Function row with 14 shortcut keys", "USB-C pass-through charging"], [
        ("Space Black", 34900, 37900, {"color":"Space Black"}, "products/magic-keyboard-ipad-pro-13.jpg"),
        ("White", 34900, 37900, {"color":"White"}, "products/magic-keyboard-ipad-pro-13.jpg")
    ]),

    # Computer Accessories & Peripherals (15)
    ("Computer Accessories", "Logitech", "Logitech MX Master 3S Wireless Mouse", "logitech-mx-master-3s", "Iconic ergonomic performance mouse with quiet clicks and 8K DPI sensor.", ["8,000 DPI Darkfield sensor (tracks on glass)", "MagSpeed electromagnetic scroll wheel", "Quiet Clicks reduce 90% click noise"], [
        ("Graphite", 9999, 10999, {"color":"Graphite"}, "products/logitech-mx-master-3s.jpg"),
        ("Pale Gray", 9999, 10999, {"color":"Pale Gray"}, "products/logitech-mx-master-3s.jpg")
    ]),
    ("Computer Accessories", "Logitech", "Logitech MX Keys S Wireless Keyboard", "logitech-mx-keys-s", "Low-profile fluid mechanical typing with smart backlighting.", ["Spherically dished keys match fingertips", "Smart proximity backlighting sensors", "Smart Actions macro automations"], [
        ("Graphite", 10999, 11999, {"color":"Graphite"}, "products/logitech-mx-keys-s.jpg"),
        ("Pale Gray", 10999, 11999, {"color":"Pale Gray"}, "products/logitech-mx-keys-s.jpg")
    ]),
    ("Computer Accessories", "Logitech", "Logitech G PRO X SUPERLIGHT 2 Gaming Mouse", "logitech-g-pro-x-superlight-2", "60g ultralight esports mouse with LIGHTFORCE hybrid optical switches.", ["LIGHTFORCE optical-mechanical switches", "HERO 2 32,000 DPI precision sensor", "95-hour continuous battery life"], [
        ("Black", 15999, 17999, {"color":"Black"}, "products/logitech-g-pro-x-superlight-2.jpg"),
        ("White", 15999, 17999, {"color":"White"}, "products/logitech-g-pro-x-superlight-2-white.jpg"),
        ("Magenta", 15999, 17999, {"color":"Magenta"}, "products/logitech-g-pro-x-superlight-2-magenta.jpg")
    ]),
    ("Computer Accessories", "Logitech", "Logitech G915 LIGHTSPEED Wireless RGB Keyboard", "logitech-g915-lightspeed", "Low profile aircraft-grade aluminum mechanical gaming keyboard.", ["GL Tactile mechanical switches", "LIGHTSPEED pro-grade 1ms wireless", "Per-key LIGHTSYNC RGB backlighting"], [
        ("Carbon Tactile", 24999, 26999, {"color":"Carbon"}, "products/logitech-g915-lightspeed.jpg"),
        ("White Clicky", 24999, 26999, {"color":"White"}, "products/logitech-g915-lightspeed.jpg")
    ]),
    ("Computer Accessories", "Logitech", "Logitech Brio 4K Ultra HD Webcam", "logitech-brio-4k", "Professional 4K streaming and videoconferencing webcam with HDR.", ["4K Ultra HD at 30 fps / 1080p at 60 fps", "RightLight 3 auto light correction", "Windows Hello facial recognition"], [
        ("Black", 19999, 21999, {"color":"Black"}, "products/logitech-brio-4k.jpg")
    ]),
    ("Computer Accessories", "Logitech", "Logitech Lift Vertical Ergonomic Mouse", "logitech-lift-vertical-mouse", "57-degree vertical grip that relaxes hand posture for all-day comfort.", ["57-degree upright posture angle", "Whisper-quiet clicks", "SmartWheel magnetic speed scroll"], [
        ("Graphite", 6999, 7999, {"color":"Graphite"}, "products/logitech-lift-vertical-mouse.jpg"),
        ("Off-White", 6999, 7999, {"color":"Off-White"}, "products/logitech-lift-vertical-mouse.jpg"),
        ("Rose", 6999, 7999, {"color":"Rose"}, "products/logitech-lift-vertical-mouse.jpg")
    ]),
    ("Computer Accessories", "Dell", "Dell UltraSharp 32 4K USB-C Hub Monitor (U3223QE)", "dell-ultrasharp-32-4k-u3223qe", "IPS Black technology with 2000:1 contrast ratio and 90W USB-C hub.", ["31.5\" 4K UHD IPS Black panel", "90W USB-C power delivery hub with RJ45", "Built-in KVM switch and Picture-by-Picture"], [
        ("Platinum Silver", 79900, 89900, {"color":"Platinum Silver"}, "products/dell-ultrasharp-32-4k-u3223qe.jpg")
    ]),
    ("Computer Accessories", "Dell", "Dell Premier Wireless ANC Headset (WL7024)", "dell-premier-wl7024", "AI-based noise cancellation headset for distraction-free enterprise calls.", ["AI-based active noise cancellation", "Smart sensors mute mic upon removal", "Up to 78 hours listening battery life"], [
        ("Black", 29900, 32900, {"color":"Black"}, "products/dell-premier-wl7024.jpg")
    ]),
    ("Computer Accessories", "Dell", "Dell Pro Wireless Keyboard and Mouse (KM5221W)", "dell-km5221w-combo", "Longest battery life combo with up to 36 months of operation.", ["Up to 36 months battery lifespan", "Programmable shortcut keys", "Quiet typing chiclet keys"], [
        ("Black", 4999, 5999, {"color":"Black"}, "products/dell-km5221w-combo.jpg"),
        ("White", 4999, 5999, {"color":"White"}, "products/dell-km5221w-combo.jpg")
    ]),
    ("Computer Accessories", "Dell", "Dell Dual Charge Dock (HD22Q)", "dell-hd22q-dock", "World's first laptop docking station with integrated Qi wireless charging pad.", ["Built-in Qi fast wireless charging stand", "Supports dual 4K monitor outputs", "90W laptop power delivery"], [
        ("Black", 18900, 20900, {"color":"Black"}, "products/dell-hd22q-dock.jpg")
    ]),
    ("Computer Accessories", "Logitech", "Logitech Wave Keys Ergonomic Wireless Keyboard", "logitech-wave-keys", "Cushioned palm rest keyboard designed to feel instantly familiar and natural.", ["Waved compact layout", "Memory foam cushioned wrist rest", "Connects up to 3 devices via Bluetooth/Bolt"], [
        ("Off-White", 5999, 6999, {"color":"Off-White"}, "products/logitech-wave-keys.jpg"),
        ("Graphite", 5999, 6999, {"color":"Graphite"}, "products/logitech-wave-keys.jpg")
    ]),
    ("Computer Accessories", "Logitech", "Logitech G PRO X 2 LIGHTSPEED Gaming Headset", "logitech-g-pro-x-2", "Groundbreaking 50mm Graphene drivers for revolutionary audio clarity.", ["50mm Graphene audio drivers", "Up to 50 hours battery on a single charge", "Rotating durable aluminum hinge"], [
        ("Black", 24999, 26999, {"color":"Black"}, "products/logitech-g-pro-x-2.jpg"),
        ("White", 24999, 26999, {"color":"White"}, "products/logitech-g-pro-x-2.jpg")
    ]),
    ("Computer Accessories", "Dell", "Dell UltraSharp 38 Curved USB-C Hub Monitor", "dell-ultrasharp-38-curved", "Ultrawide 3840x1600 WQHD+ curved productivity display.", ["37.5\" WQHD+ 21:9 curved screen", "Built-in 90W USB-C hub and 2.5GbE LAN", "Integrated 9W dual stereo speakers"], [
        ("Silver/Black", 119900, 134900, {"color":"Silver/Black"}, "products/dell-ultrasharp-38-curved.jpg")
    ]),
    ("Computer Accessories", "Apple", "Magic Trackpad", "apple-magic-trackpad", "Wireless rechargeable glass trackpad supporting full Multi-Touch gestures.", ["Edge-to-edge glass surface", "Force Touch pressure sensors", "Long-lasting internal rechargeable battery"], [
        ("White", 12900, 13900, {"color":"White"}, "products/apple-magic-trackpad.jpg"),
        ("Black", 14900, 15900, {"color":"Black"}, "products/apple-magic-trackpad.jpg")
    ]),
    ("Computer Accessories", "Apple", "Magic Keyboard with Touch ID and Numeric Keypad", "apple-magic-keyboard-touch-id", "Quick fingerprint authentication and extended layout for spreadsheet tasks.", ["Touch ID sensor for fast secure logins", "Numeric keypad and full cursor keys", "Woven Lightning/USB-C charging cable"], [
        ("White", 17900, 18900, {"color":"White"}, "products/apple-magic-keyboard-touch-id.jpg"),
        ("Black", 19900, 20900, {"color":"Black"}, "products/apple-magic-keyboard-touch-id.jpg")
    ]),

    # Gaming Consoles & Gaming Gear (14)
    ("Gaming Consoles", "Sony", "PlayStation 5 Slim Console", "playstation-5-slim-console", "Ultra-high speed 1TB SSD with ray tracing and 4K 120fps capability.", ["1TB NVMe SSD internal storage", "Tempest 3D AudioTech", "Includes DualSense wireless controller"], [
        ("Disc Edition White", 49999, 54999, {"edition":"Disc Edition"}, "products/playstation-5-slim-console.jpg"),
        ("Digital Edition White", 44999, 49999, {"edition":"Digital Edition"}, "products/playstation-5-slim-console.jpg")
    ]),
    ("Gaming Consoles", "Sony", "PlayStation Portal Remote Player", "playstation-portal-remote-player", "Stream your PS5 games over home Wi-Fi directly to your hands.", ["8\" Full HD 1080p 60fps LCD screen", "DualSense haptic feedback integration", "3.5mm audio jack for headphones"], [
        ("White/Black", 19999, 21999, {"color":"White/Black"}, "products/playstation-portal-remote-player.jpg")
    ]),
    ("Gaming Consoles", "Sony", "PlayStation VR2 Horizon Call of the Mountain Bundle", "playstation-vr2-bundle", "4K HDR OLED virtual reality headset with eye tracking and haptics.", ["Dual 2000x2040 OLED displays", "110-degree field of view", "Includes PS VR2 Sense controllers"], [
        ("White", 59999, 64999, {"color":"White"}, "products/playstation-vr2-bundle.jpg")
    ]),
    ("Gaming Consoles", "Sony", "DualSense Edge Wireless Controller", "dualsense-edge-controller", "High-performance customizable pro gamepad for PS5 with swappable modules.", ["Changeable stick caps and back paddles", "Adjustable trigger stops and dead zones", "Braided USB cable with lockable housing"], [
        ("White/Black", 19999, 21999, {"color":"White/Black"}, "products/dualsense-edge-controller.jpg")
    ]),
    ("Gaming Consoles", "Sony", "PULSE Elite Wireless Headset", "pulse-elite-headset", "Planar magnetic drivers delivering audiophile gaming soundscapes.", ["Studio-inspired planar magnetic drivers", "Retractable microphone with AI noise rejection", "PlayStation Link ultra-low latency connection"], [
        ("White/Black", 14999, 16999, {"color":"White/Black"}, "products/pulse-elite-headset.jpg")
    ]),
    ("Gaming Consoles", "Sony", "PULSE Explore Wireless Earbuds", "pulse-explore-earbuds", "First planar magnetic wireless gaming earbuds with charging case.", ["Planar magnetic earbud drivers", "Dual hidden microphones with noise filtering", "Bluetooth + PS Link simultaneous dual connectivity"], [
        ("White", 19999, 21999, {"color":"White"}, "products/pulse-explore-earbuds.jpg")
    ]),
    ("Gaming Consoles", "Sony", "DualSense Wireless Controller", "dualsense-wireless-controller", "Dynamic adaptive triggers and immersive haptic feedback for PC and PS5.", ["Haptic feedback tactile rumble", "Dynamic resistance adaptive triggers", "Built-in microphone and headset jack"], [
        ("Midnight Black", 7499, 7999, {"color":"Midnight Black"}, "products/dualsense-wireless-controller.jpg"),
        ("Cosmic Red", 7499, 7999, {"color":"Cosmic Red"}, "products/dualsense-cosmic-red.jpg"),
        ("Volcanic Red", 7999, 8499, {"color":"Volcanic Red"}, "products/dualsense-volcanic-red.jpg")
    ]),
    ("Gaming Consoles", "Logitech", "Logitech G PRO X TKL Wireless Gaming Keyboard", "logitech-g-pro-x-tkl", "Esports-proven tenkeyless mechanical keyboard with LIGHTSPEED wireless.", ["LIGHTSPEED wireless and Bluetooth", "Dual-shot PBT keycaps", "Media control roller and game mode switch"], [
        ("Black GX Brown", 19999, 21999, {"color":"Black"}, "products/logitech-g-pro-x-tkl.jpg"),
        ("White GX Red", 19999, 21999, {"color":"White"}, "products/logitech-g-pro-x-tkl.jpg")
    ]),
    ("Gaming Consoles", "Logitech", "Logitech G502 X PLUS Wireless RGB Mouse", "logitech-g502-x-plus", "The world's most popular gaming mouse redesigned with LIGHTFORCE switches.", ["LIGHTFORCE optical-mechanical switches", "HERO 25K sub-micron sensor", "8-zone active LIGHTSYNC RGB lighting"], [
        ("Black", 15999, 17999, {"color":"Black"}, "products/logitech-g502-x-plus.jpg"),
        ("White", 15999, 17999, {"color":"White"}, "products/logitech-g502-x-plus.jpg")
    ]),
    ("Gaming Consoles", "Logitech", "Logitech G29 Driving Force Racing Wheel", "logitech-g29-racing-wheel", "Dual-motor force feedback racing wheel with stainless steel paddle shifters.", ["Dual-motor force feedback realism", "900-degree lock-to-lock rotation", "Nonlinear responsive brake pedal unit"], [
        ("Black Leather", 29999, 34999, {"color":"Black"}, "products/logitech-g29-racing-wheel.jpg")
    ]),
    ("Gaming Consoles", "Sony", "PlayStation 5 Console Covers (Slim)", "ps5-slim-console-covers", "Customize your PS5 Slim with precision-engineered colored faceplates.", ["Easy click-on installation", "Deep Chroma color matching finish", "Scratch-resistant polycarbonate"], [
        ("Midnight Black", 5499, 5999, {"color":"Midnight Black"}, "products/ps5-slim-console-covers.jpg"),
        ("Volcanic Red", 5999, 6499, {"color":"Volcanic Red"}, "products/dualsense-volcanic-red.jpg")
    ]),
    ("Gaming Consoles", "Logitech", "Logitech G Fits True Wireless Gaming Earbuds", "logitech-g-fits-earbuds", "Custom moldable photopolymer ear tips that fit the exact shape of your ears.", ["LIGHTFORM 60-second custom molding", "LIGHTSPEED pro gaming wireless", "Warm rich audio with custom EQ app"], [
        ("Black", 22999, 24999, {"color":"Black"}, "products/logitech-g-fits-earbuds.jpg"),
        ("White", 22999, 24999, {"color":"White"}, "products/logitech-g-fits-earbuds.jpg")
    ]),
    ("Gaming Consoles", "Logitech", "Logitech G ASTRO A50 X Wireless Gaming Headset", "astro-a50-x-headset", "PLAYSYNC 3-system HDMI 2.1 switcher connecting Xbox, PS5, and PC.", ["PLAYSYNC one-touch 3-system switching", "PRO-G GRAPHENE 40mm audio drivers", "Simultaneous 24-bit wireless + Bluetooth"], [
        ("Black", 37999, 39999, {"color":"Black"}, "products/astro-a50-x-headset.jpg"),
        ("White", 37999, 39999, {"color":"White"}, "products/astro-a50-x-headset.jpg")
    ]),
    ("Gaming Consoles", "Sony", "DualSense Charging Station", "dualsense-charging-station", "Click-in dock that fast charges up to two PS5 DualSense controllers.", ["Frees up console USB ports", "Charges at same speed as direct cable", "Easy drop-in click alignment"], [
        ("White/Black", 2999, 3499, {"color":"White/Black"}, "products/dualsense-charging-station.jpg")
    ]),
]


for cat_name, brand_name, title, slug, desc, highlights, var_list in catalog_defs:
    prod_img = f"products/{slug}.jpg"
    if prod_img not in image_to_obj_id:
        obj_id = hex_id(0x80000000, obj_idx)
        storage_objects.append((obj_id, 'catalog', prod_img, 'image/jpeg', 150000))
        image_to_obj_id[prod_img] = obj_id
        obj_idx += 1
    
    for v in var_list:
        v_img = v[4] if len(v) > 4 else prod_img
        if v_img not in image_to_obj_id:
            obj_id = hex_id(0x80000000, obj_idx)
            storage_objects.append((obj_id, 'catalog', v_img, 'image/jpeg', 150000))
            image_to_obj_id[v_img] = obj_id
            obj_idx += 1

users = [
    (hex_id(0x10000000, 1), 'admin@prim.com', 'Admin User', 'admin', True, 'active'),
    (hex_id(0x10000000, 2), 'john.doe@example.com', 'John Doe', 'customer', True, 'active'),
    (hex_id(0x10000000, 3), 'jane.smith@example.com', 'Jane Smith', 'customer', True, 'active'),
    (hex_id(0x10000000, 4), 'support@prim.com', 'Support Specialist', 'support', True, 'active'),
    (hex_id(0x10000000, 5), 'alex.turner@example.com', 'Alex Turner', 'customer', True, 'active'),
    (hex_id(0x10000000, 6), 'sarah.connor@example.com', 'Sarah Connor', 'customer', True, 'active'),
    (hex_id(0x10000000, 7), 'michael.scott@example.com', 'Michael Scott', 'customer', True, 'active'),
    (hex_id(0x10000000, 8), 'elena.rostova@example.com', 'Elena Rostova', 'customer', True, 'active'),
    (hex_id(0x10000000, 9), 'david.beck@example.com', 'David Beck', 'customer', True, 'active'),
    (hex_id(0x10000000, 10), 'emily.watson@example.com', 'Emily Watson', 'customer', True, 'active'),
    (hex_id(0x10000000, 11), 'carlos.mendez@example.com', 'Carlos Mendez', 'customer', True, 'active'),
    (hex_id(0x10000000, 12), 'hannah.abbott@example.com', 'Hannah Abbott', 'customer', False, 'active'),
]

addresses = [
    (hex_id(0x20000000, 1), hex_id(0x10000000, 2), 'Home', 'John Doe', '+1234567890', '123 Main St', 'Apt 4B', 'New York', 'NY', '10001', 'USA', True),
    (hex_id(0x20000000, 2), hex_id(0x10000000, 3), 'Work', 'Jane Smith', '+1987654321', '456 Market St', 'Suite 100', 'San Francisco', 'CA', '94103', 'USA', True),
    (hex_id(0x20000000, 3), hex_id(0x10000000, 5), 'Home', 'Alex Turner', '+14155551234', '789 Sunset Blvd', 'Apt 12', 'Los Angeles', 'CA', '90028', 'USA', True),
    (hex_id(0x20000000, 4), hex_id(0x10000000, 6), 'Home', 'Sarah Connor', '+15125556789', '101 Cyberdyne Way', None, 'Austin', 'TX', '78701', 'USA', True),
    (hex_id(0x20000000, 5), hex_id(0x10000000, 7), 'Office', 'Michael Scott', '+15705559876', '1725 Slough Ave', 'Suite 200', 'Scranton', 'PA', '18503', 'USA', True),
    (hex_id(0x20000000, 6), hex_id(0x10000000, 8), 'Home', 'Elena Rostova', '+442071838750', '221B Baker St', 'Flat 2', 'London', 'Greater London', 'NW1 6XE', 'GBR', True),
    (hex_id(0x20000000, 7), hex_id(0x10000000, 9), 'Home', 'David Beck', '+4930123456', 'Friedrichstraße 43', None, 'Berlin', 'Berlin', '10117', 'DEU', True),
    (hex_id(0x20000000, 8), hex_id(0x10000000, 10), 'Home', 'Emily Watson', '+12065554321', '500 Pine St', 'Floor 4', 'Seattle', 'WA', '98101', 'USA', True),
    (hex_id(0x20000000, 9), hex_id(0x10000000, 11), 'Apartment', 'Carlos Mendez', '+13055557890', '800 Brickell Ave', 'PH 3', 'Miami', 'FL', '33131', 'USA', True),
    (hex_id(0x20000000, 10), hex_id(0x10000000, 2), 'Beach House', 'John Doe', '+1234567890', '50 Ocean Drive', None, 'Miami Beach', 'FL', '33139', 'USA', False),
]

brands_list = [
    ("Apple", "https://apple.com", brand_logo_map["Apple"]),
    ("Samsung", "https://samsung.com", brand_logo_map["Samsung"]),
    ("Nike", "https://nike.com", brand_logo_map["Nike"]),
    ("Sony", "https://sony.com", brand_logo_map["Sony"]),
    ("Adidas", "https://adidas.com", brand_logo_map["Adidas"]),
    ("Dell", "https://dell.com", brand_logo_map["Dell"]),
    ("Logitech", "https://logitech.com", brand_logo_map["Logitech"]),
    ("Bose", "https://bose.com", brand_logo_map["Bose"]),
]

brands = []
brand_id_by_name = {}
for i, (name, link, logo_id) in enumerate(brands_list, 1):
    b_id = hex_id(0x30000000, i*2 - 1)
    b_pub = hex_id(0x30000000, i*2)
    brands.append((b_id, b_pub, name, link, logo_id))
    brand_id_by_name[name] = b_id

tag_names = ['Featured', 'Sale', 'New Arrival', 'Best Seller', 'Trending', 'Limited Edition', 'Eco-Friendly', "Editor's Choice"]
tags = []
tag_id_by_name = {}
for i, name in enumerate(tag_names, 1):
    t_id = hex_id(0x50000000, i)
    tags.append((t_id, name))
    tag_id_by_name[name] = t_id

categories = [
    (hex_id(0x40000000, 1), hex_id(0x40000000, 2), None, 'Electronics'),
    (hex_id(0x40000000, 3), hex_id(0x40000000, 4), None, 'Apparel'),
    (hex_id(0x40000000, 5), hex_id(0x40000000, 6), hex_id(0x40000000, 1), 'Laptops'),
    (hex_id(0x40000000, 7), hex_id(0x40000000, 8), hex_id(0x40000000, 1), 'Smartphones'),
    (hex_id(0x40000000, 9), hex_id(0x40000000, 10), hex_id(0x40000000, 3), 'Shoes'),
    (hex_id(0x40000000, 11), hex_id(0x40000000, 12), hex_id(0x40000000, 1), 'Audio & Headphones'),
    (hex_id(0x40000000, 13), hex_id(0x40000000, 14), hex_id(0x40000000, 1), 'Tablets & Wearables'),
    (hex_id(0x40000000, 15), hex_id(0x40000000, 16), hex_id(0x40000000, 3), 'Outerwear'),
    (hex_id(0x40000000, 17), hex_id(0x40000000, 18), hex_id(0x40000000, 1), 'Computer Accessories'),
    (hex_id(0x40000000, 19), hex_id(0x40000000, 20), hex_id(0x40000000, 1), 'Gaming Consoles'),
]

cat_id_by_name = {
    'Laptops': hex_id(0x40000000, 5),
    'Smartphones': hex_id(0x40000000, 7),
    'Shoes': hex_id(0x40000000, 9),
    'Audio & Headphones': hex_id(0x40000000, 11),
    'Tablets & Wearables': hex_id(0x40000000, 13),
    'Outerwear': hex_id(0x40000000, 15),
    'Computer Accessories': hex_id(0x40000000, 17),
    'Gaming Consoles': hex_id(0x40000000, 19),
}

category_attributes = [
    (hex_id(0x41000000, 1), cat_id_by_name['Laptops'], 'ram', 'RAM Capacity', 'enum', ["8GB", "16GB", "18GB", "32GB", "36GB", "48GB", "64GB"], True, 1),
    (hex_id(0x41000000, 2), cat_id_by_name['Laptops'], 'storage', 'Storage Size', 'enum', ["256GB", "512GB", "1TB", "2TB", "4TB"], True, 2),
    (hex_id(0x41000000, 3), cat_id_by_name['Smartphones'], 'storage', 'Internal Storage', 'enum', ["128GB", "256GB", "512GB", "1TB"], True, 1),
    (hex_id(0x41000000, 4), cat_id_by_name['Smartphones'], 'color', 'Color Finish', 'enum', ["Titanium Gray", "Titanium Black", "Titanium Violet", "Titanium Yellow"], True, 2),
    (hex_id(0x41000000, 5), cat_id_by_name['Shoes'], 'size', 'Shoe Size (US)', 'enum', ["8", "8.5", "9", "9.5", "10", "10.5", "11", "11.5", "12"], True, 1),
    (hex_id(0x41000000, 6), cat_id_by_name['Shoes'], 'color', 'Colorway', 'enum', ["White", "Black", "Core Black", "Cloud White", "Blue"], True, 2),
    (hex_id(0x41000000, 7), cat_id_by_name['Audio & Headphones'], 'wireless', 'Wireless / Bluetooth', 'boolean', None, True, 1),
    (hex_id(0x41000000, 8), cat_id_by_name['Audio & Headphones'], 'anc', 'Active Noise Cancelling', 'boolean', None, True, 2),
    (hex_id(0x41000000, 9), cat_id_by_name['Outerwear'], 'size', 'Apparel Size', 'enum', ["XS", "S", "M", "L", "XL", "XXL"], True, 1),
    (hex_id(0x41000000, 10), cat_id_by_name['Computer Accessories'], 'connectivity', 'Connectivity Type', 'enum', ["Bluetooth", "2.4GHz Wireless", "Wired USB-C"], True, 1)
]

products = []
variants = []
tag_assignments = []
variant_media = []
inventory_ledgers = []

prod_counter = 1
var_counter = 1
vmedia_counter = 1
inv_counter = 1

for cat_name, brand_name, title, slug, desc, highlights, var_list in catalog_defs:
    p_id = hex_id(0x60000000, prod_counter)
    b_id = brand_id_by_name[brand_name]
    c_id = cat_id_by_name[cat_name]
    p_type = 'variable' if var_list and len(var_list) > 1 else 'simple'
    
    prod_main_img = f"products/{slug}.jpg"
    thumb_id = image_to_obj_id[prod_main_img]
    
    products.append((p_id, b_id, c_id, slug, title, desc, highlights, 'published', p_type, thumb_id))
    
    t_picks = random.sample(list(tag_id_by_name.values()), k=random.randint(1, 3))
    for t_id in t_picks:
        tag_assignments.append((p_id, t_id))
        
    if not var_list:
        var_list = [("Standard", 19999, 21999, {}, prod_main_img)]
        
    extra_angles = gallery_angles_by_cat.get(cat_name, [])
        
    for v_idx, v_item in enumerate(var_list):
        v_title = v_item[0]
        v_price = v_item[1]
        v_cross = v_item[2]
        v_attrs = v_item[3]
        v_img_key = v_item[4] if len(v_item) > 4 else prod_main_img
        v_thumb_id = image_to_obj_id[v_img_key]
        
        v_id = hex_id(0x70000000, var_counter)
        v_sku = f"{slug[:10].upper()}-{var_counter:03d}"
        is_def = (v_idx == 0)
        variants.append((v_id, v_sku, p_id, is_def, v_title, v_price, v_cross, 'USD', v_attrs, v_thumb_id))
        
        # 1. Primary Hero Image (sort_order = 0)
        vm_id = hex_id(0x71000000, vmedia_counter*2 - 1)
        vm_pub = hex_id(0x71000000, vmedia_counter*2)
        variant_media.append((vm_id, vm_pub, v_id, v_thumb_id, 'image', 0))
        vmedia_counter += 1
        
        # 2. Add extra gallery images for each variant (sort_order = 1, 2, ...)
        for order_idx, extra_img in enumerate(extra_angles, 1):
            extra_obj_id = image_to_obj_id[extra_img]
            vm_id_extra = hex_id(0x71000000, vmedia_counter*2 - 1)
            vm_pub_extra = hex_id(0x71000000, vmedia_counter*2)
            variant_media.append((vm_id_extra, vm_pub_extra, v_id, extra_obj_id, 'image', order_idx))
            vmedia_counter += 1
        
        inv_id = hex_id(0x81000000, inv_counter)
        qty = random.randint(20, 200)
        inventory_ledgers.append((inv_id, v_id, qty, 'restock'))
        inv_counter += 1
        
        var_counter += 1
    prod_counter += 1

# Generate SQL Output
lines = []
lines.append("-- =====================================================================")
lines.append("-- PRIM E-COMMERCE COMPREHENSIVE SEED DATA (114 Products with Variant Media)")
lines.append("-- PostgreSQL")
lines.append("-- =====================================================================")
lines.append("\nBEGIN;\n")

# 1. storage_objects
lines.append("-- =====================================================================")
lines.append("-- 1. STORAGE OBJECTS (Media & Images)")
lines.append("-- =====================================================================")
lines.append("INSERT INTO storage_objects (id, bucket, object_key, content_type, file_size, status, created_at, updated_at) VALUES")
val_rows = [f"({esc(o[0])}, {esc(o[1])}, {esc(o[2])}, {esc(o[3])}, {o[4]}, 'uploaded', now() - INTERVAL '60 days', now() - INTERVAL '60 days')" for o in storage_objects]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 2. users
lines.append("-- =====================================================================")
lines.append("-- 2. USERS (Admins, Support, Customers)")
lines.append("-- =====================================================================")
lines.append("INSERT INTO users (id, email, full_name, role, is_email_verified, status, created_at, updated_at) VALUES")
val_rows = [f"({esc(u[0])}, {esc(u[1])}, {esc(u[2])}, {esc(u[3])}, {'true' if u[4] else 'false'}, {esc(u[5])}, now() - INTERVAL '60 days', now() - INTERVAL '60 days')" for u in users]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 3. addresses
lines.append("-- =====================================================================")
lines.append("-- 3. ADDRESSES")
lines.append("-- =====================================================================")
lines.append("INSERT INTO addresses (id, user_id, label, full_name, phone, line1, line2, city, state, postal_code, country, is_default, created_at, updated_at) VALUES")
val_rows = [f"({esc(a[0])}, {esc(a[1])}, {esc(a[2])}, {esc(a[3])}, {esc(a[4])}, {esc(a[5])}, {esc(a[6])}, {esc(a[7])}, {esc(a[8])}, {esc(a[9])}, {esc(a[10])}, {'true' if a[11] else 'false'}, now() - INTERVAL '50 days', now() - INTERVAL '50 days')" for a in addresses]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 4. brands
lines.append("-- =====================================================================")
lines.append("-- 4. BRANDS")
lines.append("-- =====================================================================")
lines.append("INSERT INTO product_brands (id, public_id, name, link, logo_object_id, created_at, updated_at) VALUES")
val_rows = [f"({esc(b[0])}, {esc(b[1])}, {esc(b[2])}, {esc(b[3])}, {esc(b[4])}, now() - INTERVAL '50 days', now() - INTERVAL '50 days')" for b in brands]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 5. tags
lines.append("-- =====================================================================")
lines.append("-- 5. TAGS")
lines.append("-- =====================================================================")
lines.append("INSERT INTO product_tags (id, name, created_at, updated_at) VALUES")
val_rows = [f"({esc(t[0])}, {esc(t[1])}, now() - INTERVAL '50 days', now() - INTERVAL '50 days')" for t in tags]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 6. categories
lines.append("-- =====================================================================")
lines.append("-- 6. CATEGORIES")
lines.append("-- =====================================================================")
lines.append("INSERT INTO product_categories (id, public_id, parent_id, name, created_at, updated_at) VALUES")
val_rows = [f"({esc(c[0])}, {esc(c[1])}, {esc(c[2])}, {esc(c[3])}, now() - INTERVAL '50 days', now() - INTERVAL '50 days')" for c in categories]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 7. category_attributes
lines.append("-- =====================================================================")
lines.append("-- 7. CATEGORY ATTRIBUTES")
lines.append("-- =====================================================================")
lines.append("INSERT INTO category_attributes (id, category_id, key, label, value_type, allowed_values, is_filterable, sort_order, created_at, updated_at) VALUES")
val_rows = [f"({esc(ca[0])}, {esc(ca[1])}, {esc(ca[2])}, {esc(ca[3])}, {esc(ca[4])}, {json_esc(ca[5]) if ca[5] else 'NULL'}, {'true' if ca[6] else 'false'}, {ca[7]}, now() - INTERVAL '50 days', now() - INTERVAL '50 days')" for ca in category_attributes]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 8. products
lines.append("-- =====================================================================")
lines.append("-- 8. PRODUCTS (114 Unique Products)")
lines.append("-- =====================================================================")
lines.append("INSERT INTO products (id, brand_id, category_id, slug, title, description, highlights, status, product_type, thumbnail_object_id, created_at, updated_at) VALUES")
val_rows = [f"({esc(p[0])}, {esc(p[1])}, {esc(p[2])}, {esc(p[3])}, {esc(p[4])}, {esc(p[5])}, {json_esc(p[6])}, {esc(p[7])}, {esc(p[8])}, {esc(p[9])}, now() - INTERVAL '40 days', now() - INTERVAL '40 days')" for p in products]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 9. product_variants
lines.append("-- =====================================================================")
lines.append("-- 9. PRODUCT VARIANTS (With Variant-Specific Images)")
lines.append("-- =====================================================================")
lines.append("INSERT INTO product_variants (id, sku, product_id, is_default, title, price, crossed_out_price, currency, attributes, thumbnail_object_id, created_at, updated_at) VALUES")
val_rows = [f"({esc(v[0])}, {esc(v[1])}, {esc(v[2])}, {'true' if v[3] else 'false'}, {esc(v[4])}, {v[5]}, {v[6] if v[6] is not None else 'NULL'}, {esc(v[7])}, {json_esc(v[8])}, {esc(v[9])}, now() - INTERVAL '40 days', now() - INTERVAL '40 days')" for v in variants]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 10. product_tag_assignments
lines.append("-- =====================================================================")
lines.append("-- 10. PRODUCT TAG ASSIGNMENTS")
lines.append("-- =====================================================================")
lines.append("INSERT INTO product_tag_assignments (product_id, tag_id, created_at) VALUES")
val_rows = [f"({esc(ta[0])}, {esc(ta[1])}, now() - INTERVAL '40 days')" for ta in tag_assignments]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (product_id, tag_id) DO NOTHING;\n")

# 11. variant_media
lines.append("-- =====================================================================")
lines.append("-- 11. VARIANT MEDIA")
lines.append("-- =====================================================================")
lines.append("INSERT INTO variant_media (id, public_id, variant_id, object_id, media_type, sort_order) VALUES")
val_rows = [f"({esc(vm[0])}, {esc(vm[1])}, {esc(vm[2])}, {esc(vm[3])}, {esc(vm[4])}, {vm[5]})" for vm in variant_media]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 12. inventory_ledgers
lines.append("-- =====================================================================")
lines.append("-- 12. INVENTORY LEDGERS")
lines.append("-- =====================================================================")
lines.append("INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at) VALUES")
val_rows = [f"({esc(inv[0])}, {esc(inv[1])}, {inv[2]}, {esc(inv[3])}, now() - INTERVAL '35 days')" for inv in inventory_ledgers]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 13. coupons
lines.append("-- =====================================================================")
lines.append("-- 13. COUPONS")
lines.append("-- =====================================================================")
lines.append("""INSERT INTO coupons (id, code, discount_type, discount_value, currency, min_cart_amount, usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at) VALUES
('c0000000-0000-0000-0000-000000000001', 'WELCOME10', 'percentage', 10, 'USD', 5000, 1000, 1, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000002', 'SUMMER20', 'percentage', 20, 'USD', 10000, 500, 1, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000003', 'FLASH50', 'fixed_amount', 5000, 'USD', 20000, 200, 1, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000004', 'VIP100', 'fixed_amount', 10000, 'USD', 50000, 100, 2, now() - INTERVAL '40 days', now() + INTERVAL '60 days', 'active', now() - INTERVAL '40 days', now() - INTERVAL '40 days'),
('c0000000-0000-0000-0000-000000000005', 'EXPIRED30', 'percentage', 30, 'USD', 10000, 100, 1, now() - INTERVAL '40 days', now() - INTERVAL '5 days', 'archived', now() - INTERVAL '40 days', now() - INTERVAL '40 days')
ON CONFLICT (id) DO NOTHING;
""")

# 14. carts
lines.append("-- =====================================================================")
lines.append("-- 14. CARTS")
lines.append("-- =====================================================================")
lines.append("""INSERT INTO carts (id, user_id, session_id, created_at, updated_at) VALUES
('a0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000005', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000006', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000007', NULL, now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000006', NULL, 'sess_anon_987654321_guest_a', now() - INTERVAL '5 days', now() - INTERVAL '1 hour'),
('a0000000-0000-0000-0000-000000000007', NULL, 'sess_anon_123456789_guest_b', now() - INTERVAL '5 days', now() - INTERVAL '1 hour')
ON CONFLICT (id) DO NOTHING;
""")

# 15. cart_items
lines.append("-- =====================================================================")
lines.append("-- 15. CART ITEMS")
lines.append("-- =====================================================================")
lines.append(f"""INSERT INTO cart_items (id, cart_id, variant_id, quantity, price_at_purchase, currency, carted_at) VALUES
('a1000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', '{variants[0][0]}', 1, {variants[0][5]}, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', '{variants[1][0]}', 1, {variants[1][5]}, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000002', '{variants[2][0]}', 1, {variants[2][5]}, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000003', '{variants[3][0]}', 1, {variants[3][5]}, 'USD', now() - INTERVAL '2 days'),
('a1000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000004', '{variants[4][0]}', 2, {variants[4][5]}, 'USD', now() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;
""")

# 16. inventory_reservations
lines.append("-- =====================================================================")
lines.append("-- 16. INVENTORY RESERVATIONS")
lines.append("-- =====================================================================")
lines.append(f"""INSERT INTO inventory_reservations (id, variant_id, cart_id, quantity, expires_at, created_at, released_at) VALUES
('82000000-0000-0000-0000-000000000001', '{variants[0][0]}', 'a0000000-0000-0000-0000-000000000001', 1, now() + INTERVAL '1 hour', now() - INTERVAL '10 minutes', NULL),
('82000000-0000-0000-0000-000000000002', '{variants[1][0]}', 'a0000000-0000-0000-0000-000000000001', 1, now() + INTERVAL '1 hour', now() - INTERVAL '10 minutes', NULL),
('82000000-0000-0000-0000-000000000003', '{variants[2][0]}', 'a0000000-0000-0000-0000-000000000002', 1, now() + INTERVAL '45 minutes', now() - INTERVAL '15 minutes', NULL)
ON CONFLICT (id) DO NOTHING;
""")

# 17. orders & order items & reviews for all products
lines.append("-- =====================================================================")
lines.append("-- 17. ORDERS")
lines.append("-- =====================================================================")

customer_users = [u for u in users if u[3] == 'customer']
orders_data = []
order_items_data = []
reviews_data = []
payments_data = []

review_templates = {
    5: [
        ("Exceptional quality and performance!", "Exceeded all my expectations. The build quality and speed are truly second to none."),
        ("Best purchase this year, hands down", "Works flawlessly out of the box. Absolutely worth every single penny."),
        ("Outstanding craftsmanship and premium feel", "Everything feels so refined and durable. Daily use has been a complete pleasure."),
        ("Five stars without hesitation!", "Super fast delivery, gorgeous packaging, and top-tier performance."),
        ("Highly recommended for anyone looking for quality", "I did a ton of research before buying this and I am 100% satisfied with my choice.")
    ],
    4: [
        ("Great product with minor room for improvement", "Overall fantastic experience. Solid build quality, just wish the setup guide had a bit more detail."),
        ("Very satisfied with the purchase", "Performs reliably every day. Looks sleek and modern on my desk."),
        ("Solid 4 stars, great value for money", "Feature set is rich and handles heavy usage without breaking a sweat.")
    ],
    3: [
        ("Decent overall, meets expectations", "Does what it says on the tin. Good everyday reliability, though not mindblowing."),
        ("Average performance for the price point", "Standard quality. It works fine for my daily needs.")
    ]
}

o_counter = 1
oi_counter = 1
rev_counter = 1

for p_idx, prod in enumerate(products):
    p_id = prod[0]
    p_title = prod[4]
    
    # Get variants for this product
    prod_vars = [v for v in variants if v[2] == p_id]
    
    # Create 2 to 4 reviews per product
    num_revs = random.randint(2, 4)
    chosen_users = random.sample(customer_users, k=min(num_revs, len(customer_users)))
    
    for u_idx, u in enumerate(chosen_users):
        u_id = u[0]
        u_email = u[1]
        u_name = u[2]
        
        target_var = prod_vars[u_idx % len(prod_vars)]
        v_id = target_var[0]
        v_sku = target_var[1]
        v_price = target_var[5]
        
        o_id = hex_id(0x90000000, o_counter)
        oi_id = hex_id(0x91000000, oi_counter)
        rev_id = hex_id(0x92000000, rev_counter)
        pay_id = hex_id(0xf0000000, o_counter)
        
        # Order
        shipping_addr = {"name": u_name, "line1": "123 Market St", "city": "San Francisco", "postal_code": "94103", "country": "USA"}
        orders_data.append((o_id, u_id, u_email, shipping_addr, shipping_addr, 'delivered', None, 0, v_price, 'USD'))
        
        # Order Item
        p_snap = {"title": p_title, "sku": v_sku, "price": v_price}
        order_items_data.append((oi_id, o_id, v_id, 1, v_price, p_snap))
        
        # Payment
        payments_data.append((pay_id, o_id, v_price, 'USD', 'captured', 'stripe', f"ch_stripe_seed_{o_counter:04d}"))
        
        # Review
        rating_score = random.choices([5, 4, 3], weights=[70, 25, 5])[0]
        t_title, t_body = random.choice(review_templates[rating_score])
        reviews_data.append((rev_id, p_id, u_id, oi_id, rating_score, t_title, t_body, 'approved'))
        
        o_counter += 1
        oi_counter += 1
        rev_counter += 1

# Insert Orders
lines.append("INSERT INTO orders (id, customer_id, customer_email, shipping_address, billing_address, status, coupon_id, discount_amount, total_amount, currency, created_at, updated_at) VALUES")
val_rows = [f"({esc(o[0])}, {esc(o[1])}, {esc(o[2])}, {json_esc(o[3])}, {json_esc(o[4])}, {esc(o[5])}, {esc(o[6])}, {o[7]}, {o[8]}, {esc(o[9])}, now() - INTERVAL '25 days', now() - INTERVAL '20 days')" for o in orders_data]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 18. order_items
lines.append("-- =====================================================================")
lines.append("-- 18. ORDER ITEMS")
lines.append("-- =====================================================================")
lines.append("INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot) VALUES")
val_rows = [f"({esc(oi[0])}, {esc(oi[1])}, {esc(oi[2])}, {oi[3]}, {oi[4]}, {json_esc(oi[5])})" for oi in order_items_data]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 19. order_status_history
lines.append("-- =====================================================================")
lines.append("-- 19. ORDER STATUS HISTORY")
lines.append("-- =====================================================================")
lines.append("""INSERT INTO order_status_history (id, order_id, status, notes, created_at) VALUES
('93000000-0000-0000-0000-000000000001', '90000000-0000-0000-0000-000000000001', 'pending', 'Order placed by customer', now() - INTERVAL '20 days'),
('93000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000001', 'paid', 'Payment authorized and captured via Stripe', now() - INTERVAL '20 days' + INTERVAL '5 minutes'),
('93000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000001', 'delivered', 'Package delivered to front door', now() - INTERVAL '15 days')
ON CONFLICT (id) DO NOTHING;
""")

# 20. coupon_redemptions
lines.append("-- =====================================================================")
lines.append("-- 20. COUPON REDEMPTIONS")
lines.append("-- =====================================================================")
lines.append("""INSERT INTO coupon_redemptions (id, coupon_id, user_id, order_id, discount_amount, created_at) VALUES
('c1000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', '90000000-0000-0000-0000-000000000002', 13000, now() - INTERVAL '18 days')
ON CONFLICT (id) DO NOTHING;
""")

# 21. payments
lines.append("-- =====================================================================")
lines.append("-- 21. PAYMENTS")
lines.append("-- =====================================================================")
lines.append("INSERT INTO payments (id, order_id, amount, currency, status, provider, provider_transaction_id, error_message, created_at, updated_at) VALUES")
val_rows = [f"({esc(py[0])}, {esc(py[1])}, {py[2]}, {esc(py[3])}, {esc(py[4])}, {esc(py[5])}, {esc(py[6])}, NULL, now() - INTERVAL '20 days', now() - INTERVAL '20 days')" for py in payments_data]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 22. reviews
lines.append("-- =====================================================================")
lines.append("-- 22. REVIEWS (Seeded for all 114 products)")
lines.append("-- =====================================================================")
lines.append("INSERT INTO reviews (id, product_id, user_id, order_item_id, rating, title, body, status, created_at, updated_at) VALUES")
val_rows = [f"({esc(rv[0])}, {esc(rv[1])}, {esc(rv[2])}, {esc(rv[3])}, {rv[4]}, {esc(rv[5])}, {esc(rv[6])}, {esc(rv[7])}, now() - INTERVAL '{random.randint(2, 20)} days', now() - INTERVAL '1 day')" for rv in reviews_data]
lines.append(",\n".join(val_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

# 23. wishlist_items
lines.append("-- =====================================================================")
lines.append("-- 23. WISHLIST ITEMS")
lines.append("-- =====================================================================")
lines.append(f"""INSERT INTO wishlist_items (id, user_id, product_id, created_at) VALUES
('e0000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', '{products[0][0]}', now() - INTERVAL '15 days'),
('e0000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', '{products[1][0]}', now() - INTERVAL '15 days')
ON CONFLICT (id) DO NOTHING;
""")

lines.append("\nCOMMIT;\n")

sql_content = "\n".join(lines)
with open("migrations/000002_seed_data.up.sql", "w") as f:
    f.write(sql_content)

print(f"Generated 000002_seed_data.up.sql with {len(storage_objects)} storage_objects and {len(variants)} variants ({len(sql_content)} bytes).")
