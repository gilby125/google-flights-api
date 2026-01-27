#!/usr/bin/env node
/**
 * Decode Google Flights `f.req` payloads from a copied cURL command (or raw body).
 *
 * Usage:
 *   node scripts/decode_f_req.js path/to/curl.txt
 *   cat curl.txt | node scripts/decode_f_req.js
 *
 * It prints:
 *  - outer JSON (pretty) when possible
 *  - inner JSON (pretty) when possible (outer[1] is typically a JSON string)
 */
const fs = require("fs");

function readAllStdin() {
  return new Promise((resolve) => {
    let buf = "";
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => (buf += chunk));
    process.stdin.on("end", () => resolve(buf));
  });
}

function extractBodyFromCurl(text) {
  const candidates = [];

  // Common patterns from DevTools "Copy as cURL"
  for (const flag of ["--data-raw", "--data-binary", "--data", "-d"]) {
    const reSingle = new RegExp(`${flag}\\s+\\$?'([^']*)'`, "g");
    const reDouble = new RegExp(`${flag}\\s+\"([^\"]*)\"`, "g");

    for (const m of text.matchAll(reSingle)) candidates.push(m[1]);
    for (const m of text.matchAll(reDouble)) candidates.push(m[1]);
  }

  // If it's already just a raw body, accept it.
  if (!candidates.length && text.includes("f.req=")) {
    candidates.push(text.trim());
  }

  // Pick the body that contains f.req if possible.
  const withFreq = candidates.find((c) => c.includes("f.req="));
  return withFreq || candidates[0] || "";
}

function parseFormEncoded(body) {
  const params = new URLSearchParams(body);
  const value = params.get("f.req");
  if (!value) return "";
  // URLSearchParams already decodes percent-encoding. We still want the raw JSON string.
  return value;
}

function tryPrettyJson(label, value) {
  try {
    const parsed = JSON.parse(value);
    // eslint-disable-next-line no-console
    console.log(`\n== ${label} (parsed JSON) ==\n${JSON.stringify(parsed, null, 2)}\n`);
    return parsed;
  } catch {
    // eslint-disable-next-line no-console
    console.log(`\n== ${label} (raw) ==\n${value}\n`);
    return null;
  }
}

async function main() {
  const file = process.argv[2];
  const text = file ? fs.readFileSync(file, "utf8") : await readAllStdin();
  if (!text.trim()) {
    // eslint-disable-next-line no-console
    console.error("No input. Provide a file path or pipe text via stdin.");
    process.exit(2);
  }

  const body = extractBodyFromCurl(text);
  if (!body) {
    // eslint-disable-next-line no-console
    console.error("Could not find a request body in input.");
    process.exit(2);
  }

  const freq = parseFormEncoded(body);
  if (!freq) {
    // eslint-disable-next-line no-console
    console.error("Could not find `f.req` in the request body.");
    process.exit(2);
  }

  // `f.req` is a JSON string (often containing an inner JSON string at index 1).
  const outer = tryPrettyJson("outer f.req", freq);
  if (Array.isArray(outer) && typeof outer[1] === "string") {
    tryPrettyJson("inner f.req[1]", outer[1]);
  }
}

main().catch((err) => {
  // eslint-disable-next-line no-console
  console.error(err);
  process.exit(1);
});

