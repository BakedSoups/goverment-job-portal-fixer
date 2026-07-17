import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const options = new Map();
for (let index = 2; index < process.argv.length; index += 2) {
  options.set(process.argv[index], process.argv[index + 1]);
}

const indexPath = options.get("--index");
const snapshotPath = options.get("--snapshot");
const outputPath = options.get("--output");
const slug = options.get("--slug") || "civicjobfinder";
if (!indexPath || !snapshotPath || !outputPath) {
  throw new Error("usage: node build-payload.mjs --index FILE --snapshot FILE --output FILE [--slug SLUG]");
}

const root = resolve(import.meta.dirname, "../..");
const assetPaths = [
  ["/shell.html", resolve(indexPath)],
  ["/data/jobs.json", resolve(snapshotPath)],
  ["/static/app.css", resolve(root, "static/app.css")],
  ["/static/app.js", resolve(root, "static/app.js")],
  ["/static/social-preview.png", resolve(root, "static/social-preview.png")],
  ["/static/data/bay-area-regions.geojson", resolve(root, "static/data/bay-area-regions.geojson")],
  ["/static/vendor/leaflet/leaflet.css", resolve(root, "static/vendor/leaflet/leaflet.css")],
  ["/static/vendor/leaflet/leaflet.js", resolve(root, "static/vendor/leaflet/leaflet.js")],
];

const files = await Promise.all(assetPaths.map(async ([path, file]) => ({
  path,
  contentBase64: (await readFile(file)).toString("base64"),
})));
const worker = await readFile(resolve(root, "deploy/zero/worker.js"), "utf8");
const payload = {
  source: "prebuilt",
  slug,
  mainModule: "worker.js",
  modules: [{ name: "worker.js", type: "esm", content: worker }],
  assets: { spa: false, files },
  bindings: { kv: false, d1: false, r2: false },
  description: "Live Bay Area government job search with explainable matching and map filters",
};

await writeFile(resolve(outputPath), JSON.stringify(payload));
const bytes = files.reduce((total, file) => total + Buffer.byteLength(file.contentBase64, "base64"), Buffer.byteLength(worker));
console.log(`built ${files.length} assets (${bytes} bytes before base64) for ${slug}`);
