import { gzipSync } from "node:zlib";
import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

const assets = join(process.cwd(), "dist", "assets");
const entry = readdirSync(assets).find((name) => /^index-[\w-]+\.js$/.test(name));
if (!entry) throw new Error("Production entry bundle was not found; run npm run build first.");
const bytes = gzipSync(readFileSync(join(assets, entry))).byteLength;
const limit = 120 * 1024;
console.log(`Entry JavaScript: ${(bytes / 1024).toFixed(2)} KiB gzip (limit 120 KiB)`);
if (bytes > limit) process.exit(1);
