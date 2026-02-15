#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const ticketingRoot = path.resolve(scriptDir, "..");

const defaults = {
  router: path.join(ticketingRoot, "frontend", "src", "router.ts"),
  frontendSrc: path.join(ticketingRoot, "frontend", "src"),
  source: path.join(
    ticketingRoot,
    "backend",
    "e2e",
    "contracts",
    "frontend_contract.source.json",
  ),
  output: path.join(
    ticketingRoot,
    "backend",
    "e2e",
    "contracts",
    "frontend_contract.json",
  ),
};

const options = parseArgs(process.argv.slice(2), defaults);
run(options);

function run(opts) {
  const routerContent = readText(opts.router);
  const source = readJSON(opts.source);

  const routesByName = parseRoutes(routerContent);
  const testIDs = collectTestIDs(opts.frontendSrc);

  const errors = [];
  const routes = buildRoutes(source.routeKeys || {}, routesByName, errors);
  const selectors = buildSelectors(source.selectorKeys || {}, testIDs, errors);

  if (errors.length > 0) {
    for (const err of errors) {
      console.error(`ERROR: ${err}`);
    }
    process.exit(1);
  }

  const contract = {
    schemaVersion: 1,
    sourceFile: path.relative(ticketingRoot, opts.source),
    routerFile: path.relative(ticketingRoot, opts.router),
    routes,
    selectors,
    flows: source.flows || {},
  };

  fs.mkdirSync(path.dirname(opts.output), { recursive: true });
  fs.writeFileSync(opts.output, `${JSON.stringify(contract, null, 2)}\n`);

  console.log(
    `Generated ${path.relative(ticketingRoot, opts.output)} (${Object.keys(routes).length} routes, ${Object.keys(selectors).length} selectors)`,
  );
}

function parseArgs(argv, defaultsMap) {
  const out = { ...defaultsMap };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--router") {
      out.router = resolvePath(argv[++i], defaultsMap.router);
    } else if (arg === "--frontend-src") {
      out.frontendSrc = resolvePath(argv[++i], defaultsMap.frontendSrc);
    } else if (arg === "--source") {
      out.source = resolvePath(argv[++i], defaultsMap.source);
    } else if (arg === "--output") {
      out.output = resolvePath(argv[++i], defaultsMap.output);
    } else {
      throw new Error(`unknown argument: ${arg}`);
    }
  }
  return out;
}

function resolvePath(input, fallback) {
  if (!input) return fallback;
  return path.isAbsolute(input) ? input : path.resolve(process.cwd(), input);
}

function readText(file) {
  return fs.readFileSync(file, "utf8");
}

function readJSON(file) {
  return JSON.parse(readText(file));
}

function parseRoutes(routerText) {
  const routes = {};
  const routeRegex =
    /{[\s\S]*?path:\s*["'`]([^"'`]+)["'`][\s\S]*?name:\s*["'`]([^"'`]+)["'`][\s\S]*?}/g;
  for (const match of routerText.matchAll(routeRegex)) {
    const routePath = match[1];
    const routeName = match[2];
    routes[routeName] = routePath;
  }
  return routes;
}

function buildRoutes(routeKeys, routesByName, errors) {
  const result = {};
  for (const key of Object.keys(routeKeys).sort()) {
    const routeDef = routeKeys[key] || {};
    const routeName = routeDef.routeName;
    if (!routeName) {
      errors.push(`route key "${key}" is missing routeName`);
      continue;
    }
    const routePath = routesByName[routeName];
    if (!routePath) {
      errors.push(
        `route key "${key}" points to missing route name "${routeName}"`,
      );
      continue;
    }
    const params = extractParams(routePath);
    if (Array.isArray(routeDef.params)) {
      const expected = routeDef.params;
      if (expected.join(",") !== params.join(",")) {
        errors.push(
          `route key "${key}" params mismatch (expected "${expected.join(",")}", got "${params.join(",")}")`,
        );
        continue;
      }
    }
    result[key] = {
      name: routeName,
      path: routePath,
      params,
    };
  }
  return result;
}

function buildSelectors(selectorKeys, existingTestIDs, errors) {
  const result = {};
  for (const key of Object.keys(selectorKeys).sort()) {
    const testID = selectorKeys[key];
    if (!testID || typeof testID !== "string") {
      errors.push(`selector key "${key}" has invalid test id`);
      continue;
    }
    if (!existingTestIDs.has(testID)) {
      errors.push(
        `selector key "${key}" references missing data-testid "${testID}"`,
      );
      continue;
    }
    result[key] = `[data-testid="${testID}"]`;
  }
  return result;
}

function collectTestIDs(rootDir) {
  const out = new Set();
  for (const file of walkFiles(rootDir)) {
    if (!isFrontendSource(file)) continue;
    const content = readText(file);
    const testIDRegex = /data-testid\s*=\s*["']([^"']+)["']/g;
    for (const match of content.matchAll(testIDRegex)) {
      out.add(match[1]);
    }
  }
  return out;
}

function* walkFiles(dir) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      yield* walkFiles(fullPath);
      continue;
    }
    if (entry.isFile()) {
      yield fullPath;
    }
  }
}

function isFrontendSource(file) {
  return /\.(vue|ts|tsx|js|jsx|html)$/.test(file);
}

function extractParams(routePath) {
  const matches = routePath.matchAll(/:([a-zA-Z0-9_]+)/g);
  const params = [];
  for (const match of matches) {
    params.push(match[1]);
  }
  return params;
}
