import fs from 'node:fs'

const API_URL = "http://localhost:8080/api/v1";

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

async function seedTags() {
  for (let i = 577; i <= 100000; i++) {
      const data = await fetch(API_URL + "/admin/tags", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: `#${i}`,
        }),
      });
      console.log(await data.json())
      await sleep(randomInt(10, 50))
  }
}

await seedTags();
