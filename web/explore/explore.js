/* global maplibregl, Globe */

const state = {
  mode: "globe", // "globe" | "map"
  globe: null,
  map: null,
  origin: null,
  edges: [],
};

function $(id) {
  return document.getElementById(id);
}

function showToast(message, variant = "danger") {
  const wrap = $("toastWrap");
  const id = `t_${Math.random().toString(16).slice(2)}`;
  const el = document.createElement("div");
  el.className = `toast align-items-center text-bg-${variant} border-0 show`;
  el.id = id;
  el.setAttribute("role", "alert");
  el.setAttribute("aria-live", "assertive");
  el.setAttribute("aria-atomic", "true");
  el.style.borderRadius = "12px";
  el.style.marginTop = "10px";
  el.innerHTML = `
    <div class="d-flex">
      <div class="toast-body">${escapeHtml(message)}</div>
      <button type="button" class="btn-close btn-close-white me-2 m-auto" aria-label="Close"></button>
    </div>`;
  wrap.appendChild(el);
  el.querySelector("button").addEventListener("click", () => el.remove());
  setTimeout(() => el.remove(), 6500);
}

function escapeHtml(s) {
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

async function fetchJSON(url) {
  const res = await fetch(url, { headers: { Accept: "application/json" } });
  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`;
    try {
      const body = await res.json();
      if (body && body.error) msg = body.error;
    } catch (_) {}
    throw new Error(msg);
  }
  return res.json();
}

function normalizeCode(code) {
  return String(code || "").trim().toUpperCase();
}

function priceColor(price, maxPrice) {
  const p = Math.max(0, Math.min(1, price / Math.max(1, maxPrice)));
  // green -> yellow -> red
  const r = Math.round(64 + 191 * p);
  const g = Math.round(220 - 140 * p);
  const b = Math.round(110 - 70 * p);
  return `rgb(${r},${g},${b})`;
}

function formatUSD(x) {
  if (!Number.isFinite(x)) return "—";
  return `$${Math.round(x).toLocaleString()}`;
}

function initGlobe() {
  const container = $("globeContainer");
  container.innerHTML = "";

  state.globe = Globe()(container)
    .globeImageUrl("https://unpkg.com/three-globe/example/img/earth-dark.jpg")
    .backgroundImageUrl("https://unpkg.com/three-globe/example/img/night-sky.png")
    .atmosphereColor("#7c3aed")
    .atmosphereAltitude(0.16)
    .pointAltitude(0.01)
    .pointRadius(0.18)
    .pointColor((d) => d.color)
    .pointsMerge(true)
    .arcStroke(0.9)
    .arcColor((d) => d.color)
    .arcDashLength(0.6)
    .arcDashGap(1.5)
    .arcDashAnimateTime(1400)
    .arcAltitude((d) => d.altitude)
    .arcLabel((d) => d.label)
    .onPointClick((d) => onSelectDestination(d.destCode))
    .onArcClick((d) => onSelectDestination(d.destCode));

  state.globe.controls().autoRotate = true;
  state.globe.controls().autoRotateSpeed = 0.35;
  state.globe.controls().enableDamping = true;
  state.globe.controls().dampingFactor = 0.08;
}

function initMap() {
  const container = $("mapContainer");
  container.innerHTML = "";

  const osmStyle = {
    version: 8,
    sources: {
      osm: {
        type: "raster",
        tiles: ["https://tile.openstreetmap.org/{z}/{x}/{y}.png"],
        tileSize: 256,
        attribution: "© OpenStreetMap contributors",
      },
    },
    layers: [{ id: "osm", type: "raster", source: "osm" }],
  };

  state.map = new maplibregl.Map({
    container,
    style: osmStyle,
    center: [0, 20],
    zoom: 1.4,
    pitch: 0,
    bearing: 0,
  });

  state.map.addControl(new maplibregl.NavigationControl({ visualizePitch: true }), "top-right");
  state.map.addControl(new maplibregl.FullscreenControl(), "top-right");

  state.map.on("load", () => {
    state.map.addSource("explore-lines", {
      type: "geojson",
      data: { type: "FeatureCollection", features: [] },
    });
    state.map.addLayer({
      id: "explore-lines",
      type: "line",
      source: "explore-lines",
      paint: {
        "line-width": 1.2,
        "line-opacity": 0.7,
        "line-color": ["get", "color"],
      },
    });

    state.map.addSource("explore-points", {
      type: "geojson",
      data: { type: "FeatureCollection", features: [] },
    });
    state.map.addLayer({
      id: "explore-points",
      type: "circle",
      source: "explore-points",
      paint: {
        "circle-radius": 4,
        "circle-color": ["get", "color"],
        "circle-opacity": 0.9,
        "circle-stroke-color": "rgba(255,255,255,0.65)",
        "circle-stroke-width": 0.6,
      },
    });

    state.map.on("click", "explore-points", (e) => {
      const f = e.features && e.features[0];
      if (!f || !f.properties || !f.properties.destCode) return;
      onSelectDestination(f.properties.destCode);
    });

    state.map.on("mouseenter", "explore-points", () => (state.map.getCanvas().style.cursor = "pointer"));
    state.map.on("mouseleave", "explore-points", () => (state.map.getCanvas().style.cursor = ""));
  });
}

function setMode(mode) {
  state.mode = mode;
  $("modePill").textContent = mode === "globe" ? "Globe" : "Map";
  $("globeContainer").style.display = mode === "globe" ? "block" : "none";
  $("mapContainer").style.display = mode === "map" ? "block" : "none";

  if (mode === "globe" && !state.globe) initGlobe();
  if (mode === "map" && !state.map) initMap();
  render();
}

function buildExploreUrl() {
  const origin = normalizeCode($("origin").value);
  const maxHops = Number($("maxHops").value || 1);
  const maxPrice = Number($("maxPrice").value || 800);
  const limit = Number($("limit").value || 500);
  const dateFrom = $("dateFrom").value || "";
  const dateTo = $("dateTo").value || "";
  const airlines = $("airlines").value.trim();

  const qs = new URLSearchParams();
  qs.set("origin", origin);
  qs.set("maxHops", String(maxHops));
  qs.set("maxPrice", String(maxPrice));
  qs.set("limit", String(limit));
  if (dateFrom) qs.set("dateFrom", dateFrom);
  if (dateTo) qs.set("dateTo", dateTo);
  if (airlines) qs.set("airlines", airlines);

  return { origin, maxPrice, url: `/api/v1/graph/explore?${qs.toString()}` };
}

async function loadTopAirports() {
  try {
    const list = await fetchJSON("/api/v1/airports/top");
    const dl = $("topAirports");
    dl.innerHTML = "";
    for (const a of list || []) {
      const opt = document.createElement("option");
      opt.value = a.Code || a.code || "";
      dl.appendChild(opt);
    }
  } catch (e) {
    showToast(`Failed to load top airports: ${e.message}`);
  }
}

async function explore() {
  const { origin, maxPrice, url } = buildExploreUrl();
  if (!origin || origin.length !== 3) {
    showToast("Enter a 3-letter IATA origin (e.g. ORD).", "warning");
    return;
  }

  $("exploreBtn").disabled = true;
  $("resultLabel").textContent = "Loading…";
  $("routeDetails").textContent = "Click a destination to load route stats.";

  try {
    const data = await fetchJSON(url);
    state.origin = origin;
    state.edges = (data.edges || [])
      .filter((e) => Number.isFinite(e.dest_lat) && Number.isFinite(e.dest_lon))
      .filter((e) => !(e.dest_lat === 0 && e.dest_lon === 0))
      .map((e) => ({
        originCode: e.origin_code,
        originLat: e.origin_lat,
        originLon: e.origin_lon,
        destCode: e.dest_code,
        destName: e.dest_name,
        destCity: e.dest_city,
        destCountry: e.dest_country,
        destLat: e.dest_lat,
        destLon: e.dest_lon,
        cheapestPrice: e.cheapest_price,
        hops: e.hops,
        color: priceColor(e.cheapest_price, maxPrice),
      }));

    const cheapest = state.edges.length ? Math.min(...state.edges.map((e) => e.cheapestPrice || Infinity)) : null;
    $("destCount").textContent = state.edges.length.toLocaleString();
    $("cheapestShown").textContent = cheapest == null || !Number.isFinite(cheapest) ? "—" : formatUSD(cheapest);
    $("resultLabel").textContent = `${state.edges.length.toLocaleString()} destinations`;

    if (!state.edges.length) showToast("No routes matched the filters.", "warning");
    render();
  } catch (e) {
    showToast(`Explore failed: ${e.message}`);
    $("resultLabel").textContent = "Error";
  } finally {
    $("exploreBtn").disabled = false;
  }
}

function render() {
  if (!state.origin || !state.edges) return;

  const maxPrice = Number($("maxPrice").value || 800);
  const originEdge = state.edges.find((e) => Number.isFinite(e.originLat) && Number.isFinite(e.originLon));
  const originLat = originEdge ? originEdge.originLat : 20;
  const originLon = originEdge ? originEdge.originLon : 0;

  if (state.mode === "globe" && state.globe) {
    const arcs = state.edges.map((e) => {
      const dist = Math.abs(e.destLat - originLat) + Math.abs(e.destLon - originLon);
      const altitude = Math.min(0.55, Math.max(0.12, dist / 250));
      return {
        startLat: originLat,
        startLng: originLon,
        endLat: e.destLat,
        endLng: e.destLon,
        color: e.color,
        altitude,
        destCode: e.destCode,
        label: `${state.origin} → ${e.destCode} • ${formatUSD(e.cheapestPrice)} • hops ${e.hops}`,
      };
    });

    const points = state.edges.map((e) => ({
      lat: e.destLat,
      lng: e.destLon,
      color: e.color,
      destCode: e.destCode,
    }));

    state.globe.arcsData(arcs);
    state.globe.pointsData(points);
    state.globe.pointLabel((d) => `${state.origin} → ${d.destCode}`);
    state.globe
      .controls()
      .autoRotate = state.edges.length > 0;

    // recentre gently
    state.globe.pointOfView({ lat: originLat, lng: originLon, altitude: 2.0 }, 900);
  }

  if (state.mode === "map" && state.map && state.map.isStyleLoaded()) {
    const lineFeatures = [];
    const pointFeatures = [];

    for (const e of state.edges) {
      lineFeatures.push({
        type: "Feature",
        geometry: {
          type: "LineString",
          coordinates: [
            [originLon, originLat],
            [e.destLon, e.destLat],
          ],
        },
        properties: {
          color: priceColor(e.cheapestPrice, maxPrice),
        },
      });
      pointFeatures.push({
        type: "Feature",
        geometry: {
          type: "Point",
          coordinates: [e.destLon, e.destLat],
        },
        properties: {
          color: priceColor(e.cheapestPrice, maxPrice),
          destCode: e.destCode,
        },
      });
    }

    const lines = state.map.getSource("explore-lines");
    const points = state.map.getSource("explore-points");
    if (lines) lines.setData({ type: "FeatureCollection", features: lineFeatures });
    if (points) points.setData({ type: "FeatureCollection", features: pointFeatures });

    state.map.easeTo({ center: [originLon, originLat], zoom: 2.1, duration: 800 });
  }
}

async function onSelectDestination(destCode) {
  const origin = state.origin;
  destCode = normalizeCode(destCode);
  if (!origin || !destCode) return;

  $("routeDetails").textContent = `Loading stats for ${origin} → ${destCode}…`;

  try {
    const stats = await fetchJSON(`/api/v1/graph/route-stats?origin=${encodeURIComponent(origin)}&dest=${encodeURIComponent(destCode)}`);
    const airlines = Array.isArray(stats.airlines) ? stats.airlines.filter(Boolean).slice(0, 10).join(", ") : "";
    const html = `
      <div><strong>${escapeHtml(origin)} → ${escapeHtml(destCode)}</strong></div>
      <div class="mt-1">Min: <code>${escapeHtml(formatUSD(stats.min_price))}</code> • Avg: <code>${escapeHtml(formatUSD(stats.avg_price))}</code> • Max: <code>${escapeHtml(formatUSD(stats.max_price))}</code></div>
      <div class="mt-1">Price points: <code>${escapeHtml(String(stats.price_points ?? "—"))}</code></div>
      ${airlines ? `<div class="mt-1">Airlines: <code>${escapeHtml(airlines)}</code></div>` : "" }
    `;
    $("routeDetails").innerHTML = html;
  } catch (e) {
    $("routeDetails").textContent = `No stats for ${origin} → ${destCode} (${e.message}).`;
  }
}

function wireUI() {
  const maxPrice = $("maxPrice");
  const label = $("maxPriceLabel");
  label.textContent = maxPrice.value;
  maxPrice.addEventListener("input", () => {
    label.textContent = maxPrice.value;
    if (state.edges && state.edges.length) render();
  });

  $("exploreBtn").addEventListener("click", explore);
  $("toggleModeBtn").addEventListener("click", () => setMode(state.mode === "globe" ? "map" : "globe"));
}

async function main() {
  wireUI();
  await loadTopAirports();
  initGlobe();
  setMode("globe");
}

main().catch((e) => showToast(e.message));

