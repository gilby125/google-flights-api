/* global maplibregl, Globe */

const state = {
  mode: "globe", // "globe" | "map"
  globe: null,
  map: null,
  origins: [],
  edges: [],
  edgesAll: [],
  settings: {
    showRoutes: true,
    showMarkers: true,
    showLabels: true,
    colorByPrice: true,
    animateRoutes: false,
    autoRotate: false,
  },
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

function loadSettings() {
  try {
    const raw = localStorage.getItem("exploreSettings");
    if (!raw) return;
    const parsed = JSON.parse(raw);
    if (!parsed || typeof parsed !== "object") return;
    state.settings = { ...state.settings, ...parsed };
  } catch (_) {}
}

function saveSettings() {
  try {
    localStorage.setItem("exploreSettings", JSON.stringify(state.settings));
  } catch (_) {}
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

function parseOrigins(raw) {
  const parts = String(raw || "")
    .split(",")
    .map((p) => normalizeCode(p))
    .filter((p) => p.length === 3);
  const seen = new Set();
  const out = [];
  for (const p of parts) {
    if (seen.has(p)) continue;
    seen.add(p);
    out.push(p);
  }
  return out;
}

function priceColor(price, maxPrice) {
  const p = Math.max(0, Math.min(1, price / Math.max(1, maxPrice)));
  // green -> yellow -> red
  const r = Math.round(64 + 191 * p);
  const g = Math.round(220 - 140 * p);
  const b = Math.round(110 - 70 * p);
  return `rgb(${r},${g},${b})`;
}

function themeColor(price, maxPrice) {
  if (!state.settings.colorByPrice) return "rgba(124,58,237,0.92)";
  return priceColor(price, maxPrice);
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
    .atmosphereColor("rgba(124, 58, 237, 0.35)")
    .atmosphereAltitude(0.09)
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
    .onPointClick((d) => onSelectRoute(d.originCode, d.destCode))
    .onArcClick((d) => onSelectRoute(d.originCode, d.destCode));

  // Destination labels (airport codes) are populated in render().
  state.globe
    .labelsData([])
    .labelText((d) => d.text)
    .labelSize(0.9)
    .labelColor(() => "rgba(255,255,255,0.75)")
    .labelDotRadius(0.18)
    .labelAltitude(0.012)
    .onLabelClick((d) => onSelectRoute(d.originCode, d.destCode));

  state.globe.controls().autoRotate = false;
  state.globe.controls().autoRotateSpeed = 0.2;
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

    state.map.addSource("explore-labels", {
      type: "geojson",
      data: { type: "FeatureCollection", features: [] },
    });
    state.map.addLayer({
      id: "explore-labels",
      type: "symbol",
      source: "explore-labels",
      layout: {
        "text-field": ["get", "destCode"],
        "text-size": 12,
        "text-offset": [0, 1.0],
        "text-anchor": "top",
        "text-allow-overlap": false,
      },
      paint: {
        "text-color": "rgba(255,255,255,0.85)",
        "text-halo-color": "rgba(0,0,0,0.85)",
        "text-halo-width": 1.2,
      },
    });

    state.map.on("click", "explore-lines", (e) => {
      const f = e.features && e.features[0];
      if (!f || !f.properties || !f.properties.destCode) return;
      onSelectRoute(f.properties.originCode, f.properties.destCode);
    });

    state.map.on("click", "explore-labels", (e) => {
      const f = e.features && e.features[0];
      if (!f || !f.properties || !f.properties.destCode) return;
      onSelectRoute(f.properties.originCode, f.properties.destCode);
    });

    state.map.on("click", "explore-points", (e) => {
      const f = e.features && e.features[0];
      if (!f || !f.properties || !f.properties.destCode) return;
      onSelectRoute(f.properties.originCode, f.properties.destCode);
    });

    state.map.on("mouseenter", "explore-lines", () => (state.map.getCanvas().style.cursor = "pointer"));
    state.map.on("mouseleave", "explore-lines", () => (state.map.getCanvas().style.cursor = ""));

    state.map.on("mouseenter", "explore-labels", () => (state.map.getCanvas().style.cursor = "pointer"));
    state.map.on("mouseleave", "explore-labels", () => (state.map.getCanvas().style.cursor = ""));

    state.map.on("mouseenter", "explore-points", () => (state.map.getCanvas().style.cursor = "pointer"));
    state.map.on("mouseleave", "explore-points", () => (state.map.getCanvas().style.cursor = ""));
  });
}

function applySettings() {
  // Globe settings
  if (state.globe) {
    state.globe.controls().autoRotate = !!state.settings.autoRotate;
    state.globe.arcDashAnimateTime(state.settings.animateRoutes ? 1400 : 0);
  }

  // Map settings (layers may not exist yet)
  if (state.map && state.map.isStyleLoaded()) {
    const vis = (on) => (on ? "visible" : "none");
    if (state.map.getLayer("explore-lines")) state.map.setLayoutProperty("explore-lines", "visibility", vis(state.settings.showRoutes));
    if (state.map.getLayer("explore-points")) state.map.setLayoutProperty("explore-points", "visibility", vis(state.settings.showMarkers));
    if (state.map.getLayer("explore-labels")) state.map.setLayoutProperty("explore-labels", "visibility", vis(state.settings.showLabels));
  }
}

function setMode(mode) {
  state.mode = mode;
  $("modePill").textContent = mode === "globe" ? "Globe" : "Map";
  $("globeContainer").style.display = mode === "globe" ? "block" : "none";
  $("mapContainer").style.display = mode === "map" ? "block" : "none";

  if (mode === "globe" && !state.globe) initGlobe();
  if (mode === "map" && !state.map) initMap();
  applySettings();
  render();
}

function buildExploreUrl() {
  const origins = parseOrigins($("origin").value);
  const maxHops = Number($("maxHops").value || 1);
  const maxPrice = Number($("maxPrice").value || 800);
  const limit = Number($("limit").value || 500);
  const dateFrom = $("dateFrom").value || "";
  const dateTo = $("dateTo").value || "";
  const airlines = $("airlines").value.trim();
  const source = "price_point";

  const qs = new URLSearchParams();
  if (origins.length === 1) qs.set("origin", origins[0]);
  if (origins.length > 1) qs.set("origins", origins.join(","));
  qs.set("maxHops", String(maxHops));
  qs.set("maxPrice", String(maxPrice));
  qs.set("limit", String(limit));
  qs.set("source", source);
  if (dateFrom) qs.set("dateFrom", dateFrom);
  if (dateTo) qs.set("dateTo", dateTo);
  if (airlines) qs.set("airlines", airlines);

  return { origins, maxPrice, url: `/api/v1/graph/explore?${qs.toString()}`, qs };
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
  const { origins, maxPrice, url, qs } = buildExploreUrl();
  if (!origins.length) {
    showToast("Enter one or more 3-letter IATA origins (e.g. MKE or MKE,JFK).", "warning");
    return;
  }

  $("exploreBtn").disabled = true;
  $("resultLabel").textContent = "Loading…";
  $("routeDetails").textContent = "Click a destination to load route stats.";

  try {
    let data = await fetchJSON(url);
    if (!Array.isArray(data.edges) || data.edges.length === 0) {
      // Fallback: some datasets may only have ROUTE edges (avgPrice) but not PRICE_POINT edges yet.
      // Switch automatically so the UI doesn't look "broken".
      const fallbackQs = new URLSearchParams(qs);
      fallbackQs.set("source", "route");
      fallbackQs.delete("dateFrom");
      fallbackQs.delete("dateTo");
      data = await fetchJSON(`/api/v1/graph/explore?${fallbackQs.toString()}`);
      if (Array.isArray(data.edges) && data.edges.length > 0) {
        showToast("No PRICE_POINT data found; showing ROUTE data instead.", "warning");
      }
    }
    state.origin = origin;
    state.origins = origins;
    state.edgesAll = (data.edges || [])
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
      }));

    applyMaxPriceFilter();
    applySettings();

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

function applyMaxPriceFilter() {
  const maxPrice = Number($("maxPrice").value || 800);
  state.edges = (state.edgesAll || []).filter((e) => Number.isFinite(e.cheapestPrice) && e.cheapestPrice <= maxPrice);
  $("resultLabel").textContent = state.edges.length ? `${state.edges.length.toLocaleString()} destinations` : "—";
  $("destCount").textContent = state.edges.length ? state.edges.length.toLocaleString() : "—";
  const cheapest = state.edges.length ? Math.min(...state.edges.map((e) => e.cheapestPrice || Infinity)) : null;
  $("cheapestShown").textContent = cheapest == null || !Number.isFinite(cheapest) ? "—" : formatUSD(cheapest);
}

function render() {
  if (!state.origins || !state.edges) return;

  const maxPrice = Number($("maxPrice").value || 800);
  const originByCode = new Map();
  for (const e of state.edges) {
    if (!originByCode.has(e.originCode) && Number.isFinite(e.originLat) && Number.isFinite(e.originLon) && !(e.originLat === 0 && e.originLon === 0)) {
      originByCode.set(e.originCode, { lat: e.originLat, lon: e.originLon });
    }
  }
  const firstOrigin = state.origins.find((o) => originByCode.has(o));
  const originLat = firstOrigin ? originByCode.get(firstOrigin).lat : 20;
  const originLon = firstOrigin ? originByCode.get(firstOrigin).lon : 0;

  if (state.mode === "globe" && state.globe) {
    const arcs = (state.settings.showRoutes ? state.edges : [])
      .map((e) => {
        const o = originByCode.get(e.originCode) || { lat: originLat, lon: originLon };
        const dist = Math.abs(e.destLat - o.lat) + Math.abs(e.destLon - o.lon);
        const altitude = Math.min(0.55, Math.max(0.12, dist / 250));
        return {
          startLat: o.lat,
          startLng: o.lon,
          endLat: e.destLat,
          endLng: e.destLon,
          color: themeColor(e.cheapestPrice, maxPrice),
          altitude,
          originCode: e.originCode,
          destCode: e.destCode,
          label: `${e.originCode} → ${e.destCode} • ${formatUSD(e.cheapestPrice)} • hops ${e.hops}`,
        };
      });

    const bestByDest = new Map();
    for (const e of state.edges) {
      const cur = bestByDest.get(e.destCode);
      if (!cur || e.cheapestPrice < cur.cheapestPrice) bestByDest.set(e.destCode, e);
    }
    const points = (state.settings.showMarkers ? Array.from(bestByDest.values()) : []).map((e) => ({
      lat: e.destLat,
      lng: e.destLon,
      color: themeColor(e.cheapestPrice, maxPrice),
      originCode: e.originCode,
      destCode: e.destCode,
    }));
    const labels = (state.settings.showLabels ? Array.from(bestByDest.values()) : []).map((e) => ({
      lat: e.destLat,
      lng: e.destLon,
      text: e.destCode,
      originCode: e.originCode,
      destCode: e.destCode,
    }));

    state.globe.arcsData(arcs);
    state.globe.pointsData(points);
    state.globe.labelsData(labels);
    state.globe.pointLabel((d) => `${d.originCode} → ${d.destCode}`);
    state.globe.controls().autoRotate = !!state.settings.autoRotate;

    // recentre gently
    state.globe.pointOfView({ lat: originLat, lng: originLon, altitude: 2.0 }, 900);
  }

  if (state.mode === "map" && state.map && state.map.isStyleLoaded()) {
    const lineFeatures = [];
    const pointFeatures = [];
    const labelFeatures = [];

    for (const e of state.edges) {
      const o = originByCode.get(e.originCode) || { lat: originLat, lon: originLon };
      if (state.settings.showRoutes) {
        lineFeatures.push({
          type: "Feature",
          geometry: {
            type: "LineString",
            coordinates: [
              [o.lon, o.lat],
              [e.destLon, e.destLat],
            ],
          },
          properties: {
            color: themeColor(e.cheapestPrice, maxPrice),
            originCode: e.originCode,
            destCode: e.destCode,
            cheapestPrice: e.cheapestPrice,
            hops: e.hops,
          },
        });
      }
    }

    const bestByDest = new Map();
    for (const e of state.edges) {
      const cur = bestByDest.get(e.destCode);
      if (!cur || e.cheapestPrice < cur.cheapestPrice) bestByDest.set(e.destCode, e);
    }
    for (const e of bestByDest.values()) {
      if (state.settings.showMarkers) {
        pointFeatures.push({
          type: "Feature",
          geometry: {
            type: "Point",
            coordinates: [e.destLon, e.destLat],
          },
          properties: {
            color: themeColor(e.cheapestPrice, maxPrice),
            originCode: e.originCode,
            destCode: e.destCode,
          },
        });
      }
      if (state.settings.showLabels) {
        labelFeatures.push({
          type: "Feature",
          geometry: {
            type: "Point",
            coordinates: [e.destLon, e.destLat],
          },
          properties: {
            originCode: e.originCode,
            destCode: e.destCode,
          },
        });
      }
    }

    const lines = state.map.getSource("explore-lines");
    const points = state.map.getSource("explore-points");
    const labels = state.map.getSource("explore-labels");
    if (lines) lines.setData({ type: "FeatureCollection", features: lineFeatures });
    if (points) points.setData({ type: "FeatureCollection", features: pointFeatures });
    if (labels) labels.setData({ type: "FeatureCollection", features: labelFeatures });

    state.map.easeTo({ center: [originLon, originLat], zoom: 2.1, duration: 800 });
  }
}

async function onSelectRoute(origin, destCode) {
  origin = normalizeCode(origin);
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
    if (state.edgesAll && state.edgesAll.length) {
      applyMaxPriceFilter();
      render();
    }
  });
  maxPrice.addEventListener("change", () => {
    // Re-query when the user stops dragging so raising the max price can pull in more routes.
    if (parseOrigins($("origin").value).length) explore();
  });

  $("exploreBtn").addEventListener("click", explore);
  $("toggleModeBtn").addEventListener("click", () => setMode(state.mode === "globe" ? "map" : "globe"));

  const bindSwitch = (id, key) => {
    const el = $(id);
    if (!el) return;
    el.checked = !!state.settings[key];
    el.addEventListener("change", () => {
      state.settings[key] = !!el.checked;
      saveSettings();
      applySettings();
      render();
    });
  };
  bindSwitch("swShowRoutes", "showRoutes");
  bindSwitch("swShowMarkers", "showMarkers");
  bindSwitch("swShowLabels", "showLabels");
  bindSwitch("swColorByPrice", "colorByPrice");
  bindSwitch("swAnimateRoutes", "animateRoutes");
  bindSwitch("swAutoRotate", "autoRotate");
}

async function main() {
  loadSettings();
  wireUI();
  await loadTopAirports();
  initGlobe();
  setMode("globe");
}

main().catch((e) => showToast(e.message));
