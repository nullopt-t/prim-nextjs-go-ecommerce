import fs from 'node:fs';

const API_URL = "http://localhost:8080/api/v1";
const TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZGVhNGQzNmYtZjdhOC00ODZiLTg0YmQtOGRhN2Q1MmQwMDE4IiwidHlwZSI6ImFjY2Vzc190b2tlbiIsInN1YiI6ImRlYTRkMzZmLWY3YTgtNDg2Yi04NGJkLThkYTdkNTJkMDAxOCIsImV4cCI6MTc4NTg1NjcxMSwiaWF0IjoxNzg1ODU1ODExfQ.5BVg9Y_PssYixG72qKS0ff5pYY0oyOhYMgN-iefVGfU";

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

const req = async (path, method = "POST", body) => {
  const res = await fetch(API_URL + path, {
    method,
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${TOKEN}`
    },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) {
    console.error(`Error on ${method} ${path}:`, JSON.stringify(data));
    return null;
  }
  return data.data;
};

async function seed() {
  console.log("Seeding Categories...");
  const catNames = ["Electronics", "Clothing", "Home & Garden", "Sports & Outdoors", "Toys & Games"];
  const categories = [];
  for (const name of catNames) {
    const data = await req("/admin/categories", "POST", { name });
    if (data && data.id) {
        categories.push(data.id);
        console.log(`Created category: ${name} (id: ${data.id})`);
    }
    await sleep(randomInt(50, 100));
  }

  console.log("Seeding Brands...");
  const brandNames = ["Apple", "Samsung", "Nike", "Adidas", "Sony", "LG", "IKEA"];
  const brands = [];
  for (const name of brandNames) {
    const data = await req("/admin/brands", "POST", { name, link: `https://${name.toLowerCase().replace(/\s/g, '')}.com` });
    if (data && data.id) {
        brands.push(data.id);
        console.log(`Created brand: ${name} (id: ${data.id})`);
    }
    await sleep(randomInt(50, 100));
  }

  console.log("Seeding Tags...");
  const tagNames = ["summer-sale", "bestseller", "new-arrival", "discount", "limited-edition"];
  const tags = [];
  for (const name of tagNames) {
    const data = await req("/admin/tags", "POST", { name });
    if (data && data.id) {
        tags.push(data.id);
        console.log(`Created tag: ${name} (id: ${data.id})`);
    }
    await sleep(randomInt(50, 100));
  }

  if (categories.length === 0 || brands.length === 0) {
    console.error("Need categories and brands. Aborting.");
    return;
  }

  console.log("Seeding Products & Variants...");
  for (let i = 1; i <= 5; i++) {
    const brand_id = brands[randomInt(0, brands.length - 1)];
    const category_id = categories[randomInt(0, categories.length - 1)];
    const body = {
      brand_id,
      category_id,
      title: `Awesome Product ${i}`,
      description: `Detailed description for product ${i}`,
      highlights: ["High quality", "Durable", "Eco-friendly"]
    };
    const prod = await req("/admin/products", "POST", body);
    if (prod && prod.id) {
        console.log(`Created draft product: ${body.title} (public_id: ${prod.id})`);

        // Get internal UUID via admin GET
        const adminProdsRes = await fetch(API_URL + "/admin/products", {
          headers: { "Authorization": `Bearer ${TOKEN}` }
        });
        const adminProdsData = await adminProdsRes.json();
        const fullProd = (adminProdsData.data || []).find(p => p.public_id === prod.id);

        if (fullProd && fullProd.id) {
            // Create a variant first
            const vPrice = randomInt(1000, 50000);
            await req(`/admin/products/${fullProd.id}/variants`, "POST", {
              title: "Standard Variant",
              price: vPrice,
              crossed_out_price: vPrice + 1000,
              currency: "USD",
              attributes: { color: "Blue" },
              is_default: true
            });
            console.log(`Created variant for product ${fullProd.id}`);

            // Now publish
            await req(`/admin/products/${fullProd.id}/publish`, "POST", {});
            console.log(`Published product ${fullProd.id}`);
        }
    }
    await sleep(randomInt(50, 100));
  }

  console.log("Seeding completed successfully.");
}

seed().catch(err => console.error("Unhandled error:", err));
