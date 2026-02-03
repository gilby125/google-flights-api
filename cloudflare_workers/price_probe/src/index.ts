type Env = {
	GOOGLE_ABUSE_EXEMPTION?: string;
};

type CfSummary = {
	colo?: string;
	country?: string;
	region?: string;
	city?: string;
	asn?: number;
	asOrganization?: string;
	timezone?: string;
};

type ProbeResult = {
	ok: boolean;
	status: number;
	elapsed_ms: number;
	request_cf?: unknown;
	query: Record<string, unknown>;
	cache?: { hit: boolean; ttl_sec?: number; age_sec?: number };
	google?: {
		status: number;
		ok: boolean;
		error?: string;
		retried?: boolean;
	};
	prices?: {
		currency: string;
		min?: number;
		max?: number;
		price_range?: { low?: number; high?: number };
		sample: number[];
		total_extracted: number;
		lines_scanned: number;
	};
	debug?: {
		google_response_truncated?: string;
		cache_key?: string;
	};
};

const DEFAULT_UA =
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36";

const MAX_IATA_CODES_PER_SIDE = 4;
const MAX_CARRIERS = 12;
const MAX_EXTRACTED_PRICES = 250;
const DEFAULT_CACHE_TTL_SEC = 300;
const MAX_CACHE_TTL_SEC = 3600;

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify(data, null, 2), {
		status,
		headers: {
			"content-type": "application/json; charset=utf-8",
			"cache-control": "no-store",
			"access-control-allow-origin": "*",
		},
	});
}

function badRequest(message: string, extra?: Record<string, unknown>): Response {
	return jsonResponse({ ok: false, error: message, ...(extra ?? {}) }, 400);
}

function parseIntParam(value: string | null, fallback: number): number {
	if (!value) return fallback;
	const n = Number.parseInt(value, 10);
	return Number.isFinite(n) ? n : fallback;
}

function clampInt(value: number, min: number, max: number): number {
	if (value < min) return min;
	if (value > max) return max;
	return value;
}

function parseDateParam(value: string | null): string | null {
	if (!value) return null;
	const trimmed = value.trim();
	if (!/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) return null;
	return trimmed;
}

function parseIataList(value: string | null): string[] | null {
	if (!value) return null;
	const parts = value
		.split(",")
		.map((p) => p.trim().toUpperCase())
		.filter(Boolean);
	if (parts.length === 0) return null;
	if (parts.length > MAX_IATA_CODES_PER_SIDE) return null;
	for (const code of parts) {
		if (!/^[A-Z]{3}$/.test(code)) return null;
	}
	return parts;
}

function parseGoogleHost(value: string | null): string {
	const raw = (value ?? "").trim().toLowerCase();
	if (raw === "") return "www.google.com";
	// Prevent SSRF: allow only www.google.<tld> with no slashes/ports.
	if (!/^www\.google\.[a-z.]{2,15}$/.test(raw)) return "www.google.com";
	if (raw.includes("/") || raw.includes(":")) return "www.google.com";
	return raw;
}

function serializeLocationsAirports(airports: string[]): string {
	// Produces: [\"SFO\",0],[\"LAX\",0]
	return airports.map((code) => `[\\\"${code}\\\",0]`).join(",");
}

function serializeStops(stops: string): string {
	// Matches flights.serializeFlightStop (API enum, not URL enum).
	switch (stops) {
		case "nonstop":
			return "1";
		case "one_stop":
			return "2";
		case "two_stops":
			return "3";
		default:
			return "0"; // any
	}
}

function serializeCarriers(raw: string | null): string {
	if (!raw) return "[]";
	const parts = raw
		.split(",")
		.map((p) => p.trim().toUpperCase())
		.filter(Boolean);
	if (parts.length === 0) return "[]";

	const unique: string[] = [];
	const seen = new Set<string>();
	for (const token of parts) {
		if (!/^[A-Z0-9_]+$/.test(token)) continue;
		if (seen.has(token)) continue;
		if (unique.length >= MAX_CARRIERS) break;
		seen.add(token);
		unique.push(token);
	}
	if (unique.length === 0) return "[]";
	return `[${unique.map((t) => `\\\"${t}\\\"`).join(",")}]`;
}

function buildRawData(args: {
	tripType: "one_way" | "round_trip";
	class: number;
	adult: number;
	children: number;
	infantsLap: number;
	infantsSeat: number;
	stops: string;
	carriers: string;
	srcAirports: string[];
	dstAirports: string[];
	date: string;
	returnDate?: string;
}): string {
	// Mirrors flights.Session.getRawData for non-multicity (airport-only).
	const travelers = `[${args.adult},${args.children},${args.infantsLap},${args.infantsSeat}]`;
	const stops = serializeStops(args.stops);
	const src = serializeLocationsAirports(args.srcAirports);
	const dst = serializeLocationsAirports(args.dstAirports);

	const tripTypeEnum = args.tripType === "round_trip" ? 1 : 2; // RoundTrip=1, OneWay=2 in flights.TripType

	let raw = "";
	raw += `[null,null,${tripTypeEnum},null,[],${args.class},${travelers},null,null,null,null,null,null,[`;
	raw += `[[[${src}]],[[${dst}]],null,${stops},${args.carriers},[],\\\"${args.date}\\\",null,[],[],[],null,null,[],3]`;
	if (args.tripType === "round_trip") {
		if (!args.returnDate) {
			throw new Error("return date required for round_trip");
		}
		raw += `,[[[${dst}]],[[${src}]],null,${stops},${args.carriers},[],\\\"${args.returnDate}\\\",null,[],[],[],null,null,[],3]`;
	}
	return raw;
}

function goQueryEscape(input: string): string {
	// Good enough for this payload (no spaces). If spaces ever appear, match Go QueryEscape semantics.
	return encodeURIComponent(input).replace(/%20/g, "+");
}

function buildFReq(rawData: string): string {
	const prefix = `[null,"[[],`;
	const suffix = `],null,null,null,1,null,null,null,null,null,[]],1,0,0]"]`;
	return goQueryEscape(prefix + rawData + suffix);
}

function buildShoppingUrl(host: string, hl: string): string {
	// Keep the same base as the Go library; only swap `hl`.
	const short = hl.includes("-") ? hl.split("-")[0] : hl;
	const safe = /^[a-z]{2}$/i.test(short) ? short.toLowerCase() : "en";
	return `https://${host}/_/FlightsFrontendUi/data/travel.frontend.flights.FlightsFrontendService/GetShoppingResults?f.sid=-1300922759171628473&bl=boq_travel-frontend-ui_20230627.02_p1&hl=${safe}&soc-app=162&soc-platform=1&soc-device=1&_reqid=52717&rt=c`;
}

function buildXGoogExtHeader(args: { hl: string; gl: string; curr: string; tzOffsetMin: number }): string {
	const hl = args.hl;
	const gl = args.gl;
	const curr = args.curr;
	const tz = Number.isFinite(args.tzOffsetMin) ? Math.trunc(args.tzOffsetMin) : -120;
	return `["${hl}","${gl}","${curr}",1,null,[${tz}],null,[[48676280,48710756,47907128,48764689,48627726,48480739,48593234,48707380]],1,[]]`;
}

function cookiePairFromSecret(value: string): string {
	const trimmed = value.trim();
	if (trimmed.includes("=")) return trimmed;
	return `GOOGLE_ABUSE_EXEMPTION=${trimmed}`;
}

function cfSummaryFromRequest(request: Request): CfSummary {
	const cf = (request as any).cf ?? {};
	return {
		colo: cf.colo,
		country: cf.country,
		region: cf.region,
		city: cf.city,
		asn: cf.asn,
		asOrganization: cf.asOrganization,
		timezone: cf.timezone,
	};
}

function sleep(ms: number): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

function withTimeout(timeoutMs: number): { signal: AbortSignal; cancel: () => void } {
	const controller = new AbortController();
	const timeout = setTimeout(() => controller.abort(), timeoutMs);
	return { signal: controller.signal, cancel: () => clearTimeout(timeout) };
}

async function fetchGoogleCookies(host: string, userAgent: string, cacheTtlSec: number): Promise<string[]> {
	const cacheKey = new Request(`https://cache.local/google-init-cookies?host=${encodeURIComponent(host)}`);
	const cache = await caches.open("price_probe");
	const cached = await cache.match(cacheKey);
	if (cached) {
		try {
			const json = (await cached.json()) as { cookies?: string[] };
			if (Array.isArray(json.cookies) && json.cookies.length > 0) return json.cookies;
		} catch {
			// ignore cache errors
		}
	}

	const { signal, cancel } = withTimeout(8000);
	let init: Response;
	try {
		init = await fetch(`https://${host}/`, {
			method: "GET",
			headers: { "user-agent": userAgent, accept: "text/html" },
			signal,
		});
	} finally {
		cancel();
	}

	// In Workers, `Headers` may expose `getSetCookie()` for multi-value Set-Cookie.
	const anyHeaders = init.headers as unknown as { getSetCookie?: () => string[] };
	const setCookies = anyHeaders.getSetCookie?.() ?? [];
	const fallback = init.headers.get("set-cookie");
	const all = setCookies.length > 0 ? setCookies : fallback ? [fallback] : [];

	const pairs: string[] = [];
	for (const header of all) {
		const pair = header.split(";")[0]?.trim();
		if (pair) pairs.push(pair);
	}

	const ttl = clampInt(cacheTtlSec, 30, MAX_CACHE_TTL_SEC);
	await cache.put(
		cacheKey,
		new Response(JSON.stringify({ cookies: pairs }), {
			headers: { "content-type": "application/json", "cache-control": `public, max-age=${ttl}` },
		}),
	);

	return pairs;
}

async function fetchWithRetry(req: Request, init: RequestInit, retryOnStatuses: Set<number>): Promise<{ resp: Response; retried: boolean }> {
	let retried = false;
	const attempts = 2;
	for (let attempt = 0; attempt < attempts; attempt++) {
		try {
			const resp = await fetch(req, init);
			if (!retryOnStatuses.has(resp.status) || attempt === attempts - 1) {
				return { resp, retried };
			}
			retried = true;
		} catch (e) {
			if (attempt === attempts - 1) throw e;
			retried = true;
		}
		const base = 250 * Math.pow(2, attempt);
		const jitter = Math.floor(Math.random() * 150);
		await sleep(base + jitter);
	}
	throw new Error("unreachable");
}

type ParsedShopping = {
	prices: number[];
	priceRange?: { low?: number; high?: number };
	linesScanned: number;
};

type CachedProbePayload = {
	fetched_at_ms: number;
	google_status: number;
	prices: number[];
	price_range?: { low?: number; high?: number };
	lines_scanned: number;
};

function makeCacheKeyFromCanonical(canonical: Record<string, string>): string {
	const params = new URLSearchParams();
	const keys = Object.keys(canonical).sort();
	for (const k of keys) params.set(k, canonical[k]);
	return `https://cache.local/probe?${params.toString()}`;
}

function parseIsoDateOrNull(value: unknown): string | null {
	if (typeof value !== "string") return null;
	const trimmed = value.trim();
	if (!/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) return null;
	return trimmed;
}

function addDaysIso(date: string, days: number): string {
	const ms = Date.parse(`${date}T00:00:00Z`);
	if (!Number.isFinite(ms)) throw new Error(`invalid date: ${date}`);
	const out = new Date(ms + days * 24 * 60 * 60 * 1000);
	const yyyy = out.getUTCFullYear();
	const mm = String(out.getUTCMonth() + 1).padStart(2, "0");
	const dd = String(out.getUTCDate()).padStart(2, "0");
	return `${yyyy}-${mm}-${dd}`;
}

type MultiCitySegmentInput = { src: string; dst: string; date: string };

type MultiCityRequest = {
	segments: MultiCitySegmentInput[];
	class?: string;
	stops?: string;
	adults?: number;
	children?: number;
	infants_lap?: number;
	infants_seat?: number;
	hl?: string;
	gl?: string;
	curr?: string;
	tz_offset_min?: number;
	carriers?: string;
	google_host?: string;
	top?: number;
	cache?: 0 | 1;
	cache_ttl_sec?: number;
	debug?: 0 | 1;
};

async function cacheGetProbe(cacheUrl: string): Promise<CachedProbePayload | null> {
	const cache = await caches.open("price_probe");
	const cached = await cache.match(new Request(cacheUrl));
	if (!cached) return null;
	try {
		const json = (await cached.json()) as CachedProbePayload;
		if (!json || typeof json.fetched_at_ms !== "number" || typeof json.google_status !== "number") return null;
		if (!Array.isArray(json.prices)) return null;
		return json;
	} catch {
		return null;
	}
}

async function cachePutProbe(cacheUrl: string, payload: CachedProbePayload, ttlSec: number): Promise<void> {
	const cache = await caches.open("price_probe");
	const ttl = clampInt(ttlSec, 0, MAX_CACHE_TTL_SEC);
	if (ttl <= 0) return;
	await cache.put(
		new Request(cacheUrl),
		new Response(JSON.stringify(payload), {
			headers: { "content-type": "application/json", "cache-control": `public, max-age=${ttl}` },
		}),
	);
}

function parseShoppingResponse(text: string): ParsedShopping {
	const lines = text.split("\n");
	if (lines.length <= 2) return { prices: [], linesScanned: lines.length };

	const prices: number[] = [];
	let low: number | undefined;
	let high: number | undefined;

	for (const line of lines.slice(2)) {
		if (line.trim() === "") continue;
		let outer: unknown;
		try {
			outer = JSON.parse(line);
		} catch {
			continue;
		}
		const inner = (outer as any)?.[0]?.[2];
		if (typeof inner !== "string") continue;

		let innerJson: unknown;
		try {
			innerJson = JSON.parse(inner);
		} catch {
			continue;
		}
		if (!Array.isArray(innerJson)) continue;

		const offers1 = (innerJson as any)?.[2]?.[0];
		const offers2 = (innerJson as any)?.[3]?.[0];
		const blocks: unknown[] = [];
		if (Array.isArray(offers1)) blocks.push(...offers1);
		if (Array.isArray(offers2)) blocks.push(...offers2);

		for (const offer of blocks) {
			if (!Array.isArray(offer)) continue;
			const price = (offer as any)?.[1]?.[0]?.[1];
			if (typeof price === "number" && Number.isFinite(price)) {
				prices.push(price);
				if (prices.length >= MAX_EXTRACTED_PRICES) break;
			}
		}

		const maybeLow = (innerJson as any)?.[5]?.[4]?.[1];
		const maybeHigh = (innerJson as any)?.[5]?.[5]?.[1];
		if (typeof maybeLow === "number" && Number.isFinite(maybeLow)) low = maybeLow;
		if (typeof maybeHigh === "number" && Number.isFinite(maybeHigh)) high = maybeHigh;
	}

	const priceRange =
		low !== undefined || high !== undefined ? { low: low ?? undefined, high: high ?? undefined } : undefined;

	return { prices, priceRange, linesScanned: lines.length };
}

function buildPriceGraphFReq(args: {
	rawData: string;
	rangeStart: string;
	rangeEnd: string;
	tripLengthDays: number;
}): string {
	const prefix = `[null,"[null,`;
	const suffix = `],null,null,null,1,null,null,null,null,null,[]],[\\\"${args.rangeStart}\\\",\\\"${args.rangeEnd}\\\"],null,[${args.tripLengthDays},${args.tripLengthDays}]]"]`;
	return goQueryEscape(prefix + args.rawData + suffix);
}

function buildPriceGraphUrl(host: string, hl: string): string {
	const short = hl.includes("-") ? hl.split("-")[0] : hl;
	const safe = /^[a-z]{2}$/i.test(short) ? short.toLowerCase() : "en";
	return `https://${host}/_/FlightsFrontendUi/data/travel.frontend.flights.FlightsFrontendService/GetCalendarGraph?f.sid=-8920707734915550076&bl=boq_travel-frontend-ui_20230627.07_p1&hl=${safe}&soc-app=162&soc-platform=1&soc-device=1&_reqid=261464&rt=c`;
}

type PriceGraphOffer = { start_date: string; return_date: string; price: number };

function parsePriceGraphOffers(text: string): { offers: PriceGraphOffer[]; linesScanned: number } {
	const lines = text.split("\n");
	if (lines.length <= 2) return { offers: [], linesScanned: lines.length };

	// Each frame: JSON array where [0][2] is a JSON string; inside, schema starts with: [null, rawOffers]
	const offers: PriceGraphOffer[] = [];

	for (const line of lines.slice(2)) {
		if (line.trim() === "") continue;
		let outer: unknown;
		try {
			outer = JSON.parse(line);
		} catch {
			continue;
		}
		const inner = (outer as any)?.[0]?.[2];
		if (typeof inner !== "string") continue;

		let innerJson: unknown;
		try {
			innerJson = JSON.parse(inner);
		} catch {
			continue;
		}

		const rawOffers = (innerJson as any)?.[1];
		if (!Array.isArray(rawOffers)) continue;

		for (const raw of rawOffers) {
			if (!Array.isArray(raw)) continue;
			const start = raw[0];
			const ret = raw[1];
			const price = raw?.[2]?.[0]?.[1];
			if (typeof start !== "string" || typeof ret !== "string") continue;
			if (typeof price !== "number" || !Number.isFinite(price)) continue;
			offers.push({ start_date: start, return_date: ret, price });
			if (offers.length >= MAX_EXTRACTED_PRICES) break;
		}
	}

	return { offers, linesScanned: lines.length };
}

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		if (request.method === "OPTIONS") {
			return new Response(null, {
				status: 204,
				headers: {
					"access-control-allow-origin": "*",
					"access-control-allow-methods": "GET, OPTIONS",
					"access-control-allow-headers": "content-type",
				},
			});
		}

		const url = new URL(request.url);
		if (url.pathname === "/healthz") {
			return jsonResponse({ ok: true }, 200);
		}

		if (url.pathname === "/multi-city") {
			if (request.method !== "POST") return jsonResponse({ ok: false, error: "POST required" }, 405);

			let body: MultiCityRequest;
			try {
				body = (await request.json()) as MultiCityRequest;
			} catch {
				return badRequest("invalid JSON body");
			}

			const debug = body.debug === 1;
			const cacheEnabled = body.cache !== 0 && !debug;
			const cacheTtlSec = clampInt(
				Number.isFinite(body.cache_ttl_sec as any) ? (body.cache_ttl_sec as number) : DEFAULT_CACHE_TTL_SEC,
				0,
				MAX_CACHE_TTL_SEC,
			);

			if (!Array.isArray(body.segments) || body.segments.length < 2) {
				return badRequest("segments must be an array with at least 2 items");
			}
			if (body.segments.length > 4) {
				return badRequest("segments max is 4");
			}

			const segments: MultiCitySegmentInput[] = [];
			for (let i = 0; i < body.segments.length; i++) {
				const seg = body.segments[i] as any;
				const src = typeof seg?.src === "string" ? seg.src.trim().toUpperCase() : "";
				const dst = typeof seg?.dst === "string" ? seg.dst.trim().toUpperCase() : "";
				const date = parseIsoDateOrNull(seg?.date) ?? "";
				if (!/^[A-Z]{3}$/.test(src) || !/^[A-Z]{3}$/.test(dst) || date === "") {
					return badRequest(`invalid segment at index ${i}`);
				}
				segments.push({ src, dst, date });
			}

			const stops = (body.stops || "any").toLowerCase();
			const hl = body.hl || "en-US";
			const gl = (body.gl || "US").toUpperCase();
			const curr = (body.curr || "USD").toUpperCase();
			const tzOffsetMin = Number.isFinite(body.tz_offset_min as any) ? Math.trunc(body.tz_offset_min as number) : -120;
			const googleHost = parseGoogleHost(body.google_host ?? null);
			const carriers = serializeCarriers(body.carriers ?? null);

			const adults = clampInt(Number.isFinite(body.adults as any) ? (body.adults as number) : 1, 1, 9);
			const children = clampInt(Number.isFinite(body.children as any) ? (body.children as number) : 0, 0, 9);
			const infantsLap = clampInt(Number.isFinite(body.infants_lap as any) ? (body.infants_lap as number) : 0, 0, 9);
			const infantsSeat = clampInt(Number.isFinite(body.infants_seat as any) ? (body.infants_seat as number) : 0, 0, 9);

			const topN = clampInt(Number.isFinite(body.top as any) ? (body.top as number) : 10, 0, 50);

			const classRaw = (body.class || "economy").toLowerCase();
			const classEnum = (() => {
				switch (classRaw) {
					case "premium_economy":
						return 2;
					case "business":
						return 3;
					case "first":
						return 4;
					default:
						return 1;
				}
			})();

			const started = Date.now();

			const canonicalForCache: Record<string, string> = {
				v: "1",
				mode: "multi_city",
				segments: segments.map((s) => `${s.src}-${s.dst}-${s.date}`).join("|"),
				stops,
				class: classRaw,
				adults: String(adults),
				children: String(children),
				infants_lap: String(infantsLap),
				infants_seat: String(infantsSeat),
				hl,
				gl,
				curr,
				tz_offset_min: String(tzOffsetMin),
				carriers: body.carriers ?? "",
				google_host: googleHost,
			};
			const cacheKeyUrl = makeCacheKeyFromCanonical(canonicalForCache);

			if (cacheEnabled) {
				const cached = await cacheGetProbe(cacheKeyUrl);
				if (cached && cached.google_status === 200) {
					const elapsed = Date.now() - started;
					const sorted = cached.prices.slice().sort((a, b) => a - b);
					const sample = topN > 0 ? sorted.slice(0, topN) : [];
					const ageSec = Math.max(0, Math.floor((Date.now() - cached.fetched_at_ms) / 1000));
					return jsonResponse(
						{
							ok: true,
							status: 200,
							elapsed_ms: elapsed,
							request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
							cache: { hit: true, ttl_sec: cacheTtlSec, age_sec: ageSec },
							query: canonicalForCache,
							google: { status: 200, ok: true, retried: false },
							prices: {
								currency: curr,
								min: sorted.length > 0 ? sorted[0] : undefined,
								max: sorted.length > 0 ? sorted[sorted.length - 1] : undefined,
								price_range: cached.price_range,
								sample,
								total_extracted: cached.prices.length,
								lines_scanned: cached.lines_scanned,
							},
						},
						200,
					);
				}
			}

			let googleResp: Response;
			let googleText = "";
			let upstreamRetried = false;
			try {
				const cookies = await fetchGoogleCookies(googleHost, DEFAULT_UA, 3600);
				if (env.GOOGLE_ABUSE_EXEMPTION) cookies.push(cookiePairFromSecret(env.GOOGLE_ABUSE_EXEMPTION));
				const cookieHeader = cookies.length > 0 ? cookies.join("; ") : undefined;

				// Build raw data in the same shape as flights.getRawData for MultiCity (airport-only).
				const travelers = `[${adults},${children},${infantsLap},${infantsSeat}]`;
				const stopsSer = serializeStops(stops);
				const rawSegments = segments
					.map((s) => {
						const src = serializeLocationsAirports([s.src]);
						const dst = serializeLocationsAirports([s.dst]);
						return `[[[${src}]],[[${dst}]],null,${stopsSer},${carriers},[],\\\"${s.date}\\\",null,[],[],[],null,null,[],3]`;
					})
					.join(",");

				const tripTypeEnum = 3; // MultiCity
				// Keep the trailing `[` section open; buildFReq's suffix closes the frame (mirrors flights.Session.getRawData).
				const rawData = `[null,null,${tripTypeEnum},null,[],${classEnum},${travelers},null,null,null,null,null,null,[${rawSegments}`;

				const fReq = buildFReq(rawData);
				const unix = Math.floor(Date.now() / 1000);
				const reqBody = `f.req=${fReq}&at=AAuQa1qjMakasqKYcQeoFJjN7RZ3%3A${unix}&`;

				const { signal, cancel } = withTimeout(25000);
				try {
					const req = new Request(buildShoppingUrl(googleHost, hl), { method: "POST" });
					const retryStatuses = new Set([429, 500, 502, 503, 504]);
					const fetched = await fetchWithRetry(
						req,
						{
							headers: {
								accept: "*/*",
								"accept-language": `${hl},en;q=0.9`,
								"cache-control": "no-cache",
								pragma: "no-cache",
								"content-type": "application/x-www-form-urlencoded;charset=UTF-8",
								"user-agent": DEFAULT_UA,
								...(cookieHeader ? { cookie: cookieHeader } : {}),
								"x-goog-ext-259736195-jspb": buildXGoogExtHeader({ hl, gl, curr, tzOffsetMin }),
							},
							body: reqBody,
							signal,
						},
						retryStatuses,
					);
					googleResp = fetched.resp;
					upstreamRetried = fetched.retried;
				} finally {
					cancel();
				}
				googleText = await googleResp.text();
			} catch (e) {
				const elapsed = Date.now() - started;
				return jsonResponse(
					{
						ok: false,
						status: 502,
						elapsed_ms: elapsed,
						request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
						cache: cacheEnabled ? { hit: false, ttl_sec: cacheTtlSec } : undefined,
						query: canonicalForCache,
						google: { status: 0, ok: false, retried: upstreamRetried, error: e instanceof Error ? e.message : String(e) },
					},
					502,
				);
			}

			const elapsed = Date.now() - started;
			const parsed = parseShoppingResponse(googleText);
			const sorted = parsed.prices.slice().sort((a, b) => a - b);
			const sample = topN > 0 ? sorted.slice(0, topN) : [];

			if (cacheEnabled && googleResp.status === 200) {
				await cachePutProbe(
					cacheKeyUrl,
					{
						fetched_at_ms: Date.now(),
						google_status: googleResp.status,
						prices: parsed.prices,
						price_range: parsed.priceRange,
						lines_scanned: parsed.linesScanned,
					},
					cacheTtlSec,
				);
			}

			const result: ProbeResult = {
				ok: googleResp.ok,
				status: googleResp.status,
				elapsed_ms: elapsed,
				request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
				cache: cacheEnabled ? { hit: false, ttl_sec: cacheTtlSec } : undefined,
				query: canonicalForCache,
				google: googleResp.ok
					? { status: googleResp.status, ok: true, retried: upstreamRetried }
					: { status: googleResp.status, ok: false, retried: upstreamRetried },
				prices: {
					currency: curr,
					min: sorted.length > 0 ? sorted[0] : undefined,
					max: sorted.length > 0 ? sorted[sorted.length - 1] : undefined,
					price_range: parsed.priceRange,
					sample,
					total_extracted: parsed.prices.length,
					lines_scanned: parsed.linesScanned,
				},
			};

			if (debug) result.debug = { google_response_truncated: googleText.slice(0, 4000), cache_key: cacheKeyUrl };
			return jsonResponse(result, googleResp.ok ? 200 : 502);
		}

		if (url.pathname !== "/probe" && url.pathname !== "/price-graph") {
			return jsonResponse(
				{
					ok: true,
					message:
						"Use /probe, /price-graph, or POST /multi-city. See cloudflare_workers/price_probe/README.md",
				},
				200,
			);
		}

		if (url.pathname === "/price-graph") {
			if (request.method !== "GET") return jsonResponse({ ok: false, error: "GET required" }, 405);

			const srcAirports = parseIataList(url.searchParams.get("src"));
			const dstAirports = parseIataList(url.searchParams.get("dst"));
			const from = parseDateParam(url.searchParams.get("from"));
			const to = parseDateParam(url.searchParams.get("to"));
			const tripLength = clampInt(parseIntParam(url.searchParams.get("trip_length_days"), 7), 1, 30);
			const tripTypeRaw = (url.searchParams.get("trip_type") || "round_trip").toLowerCase();
			if (tripTypeRaw !== "round_trip") return badRequest("price-graph currently supports trip_type=round_trip only");

			if (!srcAirports) return badRequest("invalid or missing src (IATA codes, comma-separated)");
			if (!dstAirports) return badRequest("invalid or missing dst (IATA codes, comma-separated)");
			if (!from) return badRequest("invalid or missing from (YYYY-MM-DD)");
			if (!to) return badRequest("invalid or missing to (YYYY-MM-DD)");

			const stops = (url.searchParams.get("stops") || "any").toLowerCase();
			const hl = url.searchParams.get("hl") || "en-US";
			const gl = (url.searchParams.get("gl") || "US").toUpperCase();
			const curr = (url.searchParams.get("curr") || "USD").toUpperCase();
			const tzOffsetMin = parseIntParam(url.searchParams.get("tz_offset_min"), -120);
			const googleHost = parseGoogleHost(url.searchParams.get("google_host"));

			const adults = clampInt(parseIntParam(url.searchParams.get("adults"), 1), 1, 9);
			const children = clampInt(parseIntParam(url.searchParams.get("children"), 0), 0, 9);
			const infantsLap = clampInt(parseIntParam(url.searchParams.get("infants_lap"), 0), 0, 9);
			const infantsSeat = clampInt(parseIntParam(url.searchParams.get("infants_seat"), 0), 0, 9);

			const topN = clampInt(parseIntParam(url.searchParams.get("top"), 10), 0, 50);
			const debug = url.searchParams.get("debug") === "1";
			const cacheEnabled = url.searchParams.get("cache") !== "0" && !debug;
			const cacheTtlSec = clampInt(parseIntParam(url.searchParams.get("cache_ttl_sec"), DEFAULT_CACHE_TTL_SEC), 0, MAX_CACHE_TTL_SEC);
			const carriers = serializeCarriers(url.searchParams.get("carriers"));

			const classRaw = (url.searchParams.get("class") || "economy").toLowerCase();
			const classEnum = (() => {
				switch (classRaw) {
					case "premium_economy":
						return 2;
					case "business":
						return 3;
					case "first":
						return 4;
					default:
						return 1;
				}
			})();

			const started = Date.now();
			const canonicalForCache: Record<string, string> = {
				v: "1",
				mode: "price_graph",
				src: srcAirports.join(","),
				dst: dstAirports.join(","),
				from,
				to,
				trip_length_days: String(tripLength),
				trip_type: "round_trip",
				stops,
				class: classRaw,
				adults: String(adults),
				children: String(children),
				infants_lap: String(infantsLap),
				infants_seat: String(infantsSeat),
				hl,
				gl,
				curr,
				tz_offset_min: String(Math.trunc(tzOffsetMin)),
				carriers: url.searchParams.get("carriers") ?? "",
				google_host: googleHost,
			};
			const cacheKeyUrl = makeCacheKeyFromCanonical(canonicalForCache);

			if (cacheEnabled) {
				const cached = await cacheGetProbe(cacheKeyUrl);
				if (cached && cached.google_status === 200) {
					const elapsed = Date.now() - started;
					const sorted = cached.prices.slice().sort((a, b) => a - b);
					const sample = topN > 0 ? sorted.slice(0, topN) : [];
					const ageSec = Math.max(0, Math.floor((Date.now() - cached.fetched_at_ms) / 1000));
					return jsonResponse(
						{
							ok: true,
							status: 200,
							elapsed_ms: elapsed,
							request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
							cache: { hit: true, ttl_sec: cacheTtlSec, age_sec: ageSec },
							query: canonicalForCache,
							prices: { currency: curr, min: sorted[0], max: sorted[sorted.length - 1], sample, total_extracted: cached.prices.length },
						},
						200,
					);
				}
			}

			let googleResp: Response;
			let googleText = "";
			let upstreamRetried = false;
				try {
					const cookies = await fetchGoogleCookies(googleHost, DEFAULT_UA, 3600);
					if (env.GOOGLE_ABUSE_EXEMPTION) cookies.push(cookiePairFromSecret(env.GOOGLE_ABUSE_EXEMPTION));
					const cookieHeader = cookies.length > 0 ? cookies.join("; ") : undefined;

					const returnDate = addDaysIso(from, tripLength);
					const rawData = buildRawData({
						tripType: "round_trip",
						class: classEnum,
						adult: adults,
						children,
						infantsLap,
						infantsSeat,
						stops,
						carriers,
						srcAirports,
						dstAirports,
						date: from,
						returnDate,
					});

					const fReq = buildPriceGraphFReq({ rawData, rangeStart: from, rangeEnd: to, tripLengthDays: tripLength });
					const unix = Math.floor(Date.now() / 1000);
					const reqBody = `f.req=${fReq}&at=AAuQa1oq5qIkgkQ2nG9vQZFTgSME%3A${unix}&`;

				const { signal, cancel } = withTimeout(25000);
				try {
					const req = new Request(buildPriceGraphUrl(googleHost, hl), { method: "POST" });
					const retryStatuses = new Set([429, 500, 502, 503, 504]);
					const fetched = await fetchWithRetry(
						req,
						{
							headers: {
								accept: "*/*",
								"accept-language": `${hl},en;q=0.9`,
								"cache-control": "no-cache",
								pragma: "no-cache",
								"content-type": "application/x-www-form-urlencoded;charset=UTF-8",
								"user-agent": DEFAULT_UA,
								...(cookieHeader ? { cookie: cookieHeader } : {}),
								"x-goog-ext-259736195-jspb": buildXGoogExtHeader({ hl, gl, curr, tzOffsetMin }),
							},
							body: reqBody,
							signal,
						},
						retryStatuses,
					);
					googleResp = fetched.resp;
					upstreamRetried = fetched.retried;
				} finally {
					cancel();
				}
				googleText = await googleResp.text();
			} catch (e) {
				const elapsed = Date.now() - started;
				return jsonResponse(
					{
						ok: false,
						status: 502,
						elapsed_ms: elapsed,
						request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
						cache: cacheEnabled ? { hit: false, ttl_sec: cacheTtlSec } : undefined,
						query: canonicalForCache,
						google: { status: 0, ok: false, retried: upstreamRetried, error: e instanceof Error ? e.message : String(e) },
					},
					502,
				);
			}

			const elapsed = Date.now() - started;
			const parsed = parsePriceGraphOffers(googleText);
			const prices = parsed.offers.map((o) => o.price).sort((a, b) => a - b);
			const sample = topN > 0 ? prices.slice(0, topN) : [];

			if (cacheEnabled && googleResp.status === 200) {
				await cachePutProbe(
					cacheKeyUrl,
					{
						fetched_at_ms: Date.now(),
						google_status: googleResp.status,
						prices,
						price_range: undefined,
						lines_scanned: parsed.linesScanned,
					},
					cacheTtlSec,
				);
			}

			return jsonResponse(
				{
					ok: googleResp.ok,
					status: googleResp.status,
					elapsed_ms: elapsed,
					request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
					cache: cacheEnabled ? { hit: false, ttl_sec: cacheTtlSec } : undefined,
					query: canonicalForCache,
					google: googleResp.ok
						? { status: googleResp.status, ok: true, retried: upstreamRetried }
						: { status: googleResp.status, ok: false, retried: upstreamRetried },
					offers: parsed.offers.slice(0, 200),
					prices: { currency: curr, min: prices[0], max: prices[prices.length - 1], sample, total_extracted: prices.length, lines_scanned: parsed.linesScanned },
					...(debug ? { debug: { google_response_truncated: googleText.slice(0, 4000), cache_key: cacheKeyUrl } } : {}),
				},
				googleResp.ok ? 200 : 502,
			);
		}

		const srcAirports = parseIataList(url.searchParams.get("src"));
		const dstAirports = parseIataList(url.searchParams.get("dst"));
		const date = parseDateParam(url.searchParams.get("date"));
		const returnDate = parseDateParam(url.searchParams.get("return"));
		const tripTypeRaw = (url.searchParams.get("trip_type") || "one_way").toLowerCase();
		const tripType = tripTypeRaw === "round_trip" ? "round_trip" : "one_way";

		if (!srcAirports) return badRequest("invalid or missing src (IATA codes, comma-separated)");
		if (!dstAirports) return badRequest("invalid or missing dst (IATA codes, comma-separated)");
		if (!date) return badRequest("invalid or missing date (YYYY-MM-DD)");
		if (tripType === "round_trip" && !returnDate) return badRequest("return is required for trip_type=round_trip");

		const stops = (url.searchParams.get("stops") || "any").toLowerCase();
		const hl = url.searchParams.get("hl") || "en-US";
		const gl = (url.searchParams.get("gl") || "US").toUpperCase();
		const curr = (url.searchParams.get("curr") || "USD").toUpperCase();
		const tzOffsetMin = parseIntParam(url.searchParams.get("tz_offset_min"), -120);
		const googleHost = parseGoogleHost(url.searchParams.get("google_host"));

		const adults = clampInt(parseIntParam(url.searchParams.get("adults"), 1), 1, 9);
		const children = clampInt(parseIntParam(url.searchParams.get("children"), 0), 0, 9);
		const infantsLap = clampInt(parseIntParam(url.searchParams.get("infants_lap"), 0), 0, 9);
		const infantsSeat = clampInt(parseIntParam(url.searchParams.get("infants_seat"), 0), 0, 9);

		const topN = clampInt(parseIntParam(url.searchParams.get("top"), 10), 0, 50);
		const debug = url.searchParams.get("debug") === "1";
		const cacheEnabled = url.searchParams.get("cache") !== "0" && !debug;
		const cacheTtlSec = clampInt(parseIntParam(url.searchParams.get("cache_ttl_sec"), DEFAULT_CACHE_TTL_SEC), 0, MAX_CACHE_TTL_SEC);
		const carriers = serializeCarriers(url.searchParams.get("carriers"));

		// Class enum matches flights.Class: Economy=1, PremiumEconomy=2, Business=3, First=4.
		const classRaw = (url.searchParams.get("class") || "economy").toLowerCase();
		const classEnum = (() => {
			switch (classRaw) {
				case "premium_economy":
					return 2;
				case "business":
					return 3;
				case "first":
					return 4;
				default:
					return 1;
			}
		})();

		const started = Date.now();

		const canonicalForCache: Record<string, string> = {
			v: "1",
			src: srcAirports.join(","),
			dst: dstAirports.join(","),
			date,
			return: returnDate ?? "",
			trip_type: tripType,
			stops,
			class: classRaw,
			adults: String(adults),
			children: String(children),
			infants_lap: String(infantsLap),
			infants_seat: String(infantsSeat),
			hl,
			gl,
			curr,
			tz_offset_min: String(Math.trunc(tzOffsetMin)),
			carriers: url.searchParams.get("carriers") ?? "",
			google_host: googleHost,
		};
		const cacheKeyUrl = makeCacheKeyFromCanonical(canonicalForCache);

		if (cacheEnabled) {
			const cached = await cacheGetProbe(cacheKeyUrl);
			if (cached && cached.google_status === 200) {
				const elapsed = Date.now() - started;
				const sorted = cached.prices.slice().sort((a, b) => a - b);
				const sample = topN > 0 ? sorted.slice(0, topN) : [];
				const ageSec = Math.max(0, Math.floor((Date.now() - cached.fetched_at_ms) / 1000));

				const result: ProbeResult = {
					ok: true,
					status: 200,
					elapsed_ms: elapsed,
					request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
					cache: { hit: true, ttl_sec: cacheTtlSec, age_sec: ageSec },
					query: {
						src: srcAirports,
						dst: dstAirports,
						date,
						return: returnDate,
						trip_type: tripType,
						stops,
						class: classRaw,
						adults,
						children,
						infants_lap: infantsLap,
						infants_seat: infantsSeat,
						hl,
						gl,
						curr,
						tz_offset_min: tzOffsetMin,
						carriers: url.searchParams.get("carriers") ?? undefined,
						google_host: googleHost,
					},
					google: { status: 200, ok: true, retried: false },
					prices: {
						currency: curr,
						min: sorted.length > 0 ? sorted[0] : undefined,
						max: sorted.length > 0 ? sorted[sorted.length - 1] : undefined,
						price_range: cached.price_range,
						sample,
						total_extracted: cached.prices.length,
						lines_scanned: cached.lines_scanned,
					},
				};

				if (debug) result.debug = { cache_key: cacheKeyUrl };
				return jsonResponse(result, 200);
			}
		}

		let googleResp: Response;
		let googleText = "";
		let upstreamRetried = false;
		try {
			const cookies = await fetchGoogleCookies(googleHost, DEFAULT_UA, 3600);
			if (env.GOOGLE_ABUSE_EXEMPTION) {
				cookies.push(cookiePairFromSecret(env.GOOGLE_ABUSE_EXEMPTION));
			}
			const cookieHeader = cookies.length > 0 ? cookies.join("; ") : undefined;

			const rawData = buildRawData({
				tripType,
				class: classEnum,
				adult: adults,
				children,
				infantsLap,
				infantsSeat,
				stops,
				carriers,
				srcAirports,
				dstAirports,
				date,
				returnDate: returnDate ?? undefined,
			});
			const fReq = buildFReq(rawData);
			const unix = Math.floor(Date.now() / 1000);
			const body = `f.req=${fReq}&at=AAuQa1qjMakasqKYcQeoFJjN7RZ3%3A${unix}&`;

			const { signal, cancel } = withTimeout(25000);
			try {
				const req = new Request(buildShoppingUrl(googleHost, hl), { method: "POST" });
				const retryStatuses = new Set([429, 500, 502, 503, 504]);
				const fetched = await fetchWithRetry(
					req,
					{
						headers: {
							accept: "*/*",
							"accept-language": `${hl},en;q=0.9`,
							"cache-control": "no-cache",
							pragma: "no-cache",
							"content-type": "application/x-www-form-urlencoded;charset=UTF-8",
							"user-agent": DEFAULT_UA,
							...(cookieHeader ? { cookie: cookieHeader } : {}),
							"x-goog-ext-259736195-jspb": buildXGoogExtHeader({ hl, gl, curr, tzOffsetMin }),
						},
						body,
						signal,
					},
					retryStatuses,
				);
				googleResp = fetched.resp;
				upstreamRetried = fetched.retried;
			} finally {
				cancel();
			}
			googleText = await googleResp.text();
		} catch (e) {
			const elapsed = Date.now() - started;
			const result: ProbeResult = {
				ok: false,
				status: 502,
				elapsed_ms: elapsed,
				request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
				cache: cacheEnabled ? { hit: false, ttl_sec: cacheTtlSec } : undefined,
				query: {
					src: srcAirports,
					dst: dstAirports,
					date,
					return: returnDate,
					trip_type: tripType,
					stops,
					class: classRaw,
					adults,
					children,
					infants_lap: infantsLap,
					infants_seat: infantsSeat,
					hl,
					gl,
					curr,
					tz_offset_min: tzOffsetMin,
					carriers: url.searchParams.get("carriers") ?? undefined,
					google_host: googleHost,
				},
				google: { status: 0, ok: false, retried: upstreamRetried, error: e instanceof Error ? e.message : String(e) },
			};
			return jsonResponse(result, 502);
		}

		const elapsed = Date.now() - started;
		const parsed = parseShoppingResponse(googleText);
		const sorted = parsed.prices.slice().sort((a, b) => a - b);
		const sample = topN > 0 ? sorted.slice(0, topN) : [];

		if (cacheEnabled && googleResp.status === 200) {
			await cachePutProbe(
				cacheKeyUrl,
				{
					fetched_at_ms: Date.now(),
					google_status: googleResp.status,
					prices: parsed.prices,
					price_range: parsed.priceRange,
					lines_scanned: parsed.linesScanned,
				},
				cacheTtlSec,
			);
		}

		const result: ProbeResult = {
			ok: googleResp.ok,
			status: googleResp.status,
			elapsed_ms: elapsed,
			request_cf: debug ? (request as any).cf : cfSummaryFromRequest(request),
			cache: cacheEnabled ? { hit: false, ttl_sec: cacheTtlSec } : undefined,
			query: {
				src: srcAirports,
				dst: dstAirports,
				date,
				return: returnDate,
				trip_type: tripType,
				stops,
				class: classRaw,
				adults,
				children,
				infants_lap: infantsLap,
				infants_seat: infantsSeat,
				hl,
				gl,
				curr,
				tz_offset_min: tzOffsetMin,
				carriers: url.searchParams.get("carriers") ?? undefined,
				google_host: googleHost,
			},
			google: googleResp.ok
				? { status: googleResp.status, ok: true, retried: upstreamRetried }
				: { status: googleResp.status, ok: false, retried: upstreamRetried },
			prices: {
				currency: curr,
				min: sorted.length > 0 ? sorted[0] : undefined,
				max: sorted.length > 0 ? sorted[sorted.length - 1] : undefined,
				price_range: parsed.priceRange,
				sample,
				total_extracted: parsed.prices.length,
				lines_scanned: parsed.linesScanned,
			},
		};

		if (debug) {
			result.debug = {
				google_response_truncated: googleText.slice(0, 4000),
				cache_key: cacheKeyUrl,
			};
		}

		return jsonResponse(result, googleResp.ok ? 200 : 502);
	},
};
