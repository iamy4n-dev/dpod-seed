/**
 * Post-build HTML assertions for the registry pages.
 * Run after `npm run build` — reads dist/ directly.
 */
import { readFileSync } from 'fs';

let failures = 0;

function assert(condition, message) {
  if (!condition) {
    console.error(`FAIL: ${message}`);
    failures++;
  } else {
    console.log(`PASS: ${message}`);
  }
}

// --- registry/index.html ---
const indexHtml = readFileSync('dist/registry/index.html', 'utf8');

// Card must be an <a> with a registry href
assert(
  indexHtml.includes('href="/dpod-seed/registry/devops-k8s/"'),
  'card links to distro detail page'
);

// Card must have cursor:pointer so it looks clickable
assert(
  indexHtml.includes('cursor:pointer') || indexHtml.includes('cursor: pointer'),
  'card has cursor:pointer'
);

// Package names must render as text, not [object Object]
assert(
  !indexHtml.includes('[object Object]'),
  'packages do not render as [object Object] on index'
);
assert(
  indexHtml.includes('shell-zsh'),
  'package name renders correctly on index'
);

// --- registry/devops-k8s/index.html ---
const detailHtml = readFileSync('dist/registry/devops-k8s/index.html', 'utf8');

assert(
  !detailHtml.includes('[object Object]'),
  'packages do not render as [object Object] on detail page'
);
assert(
  detailHtml.includes('shell-zsh'),
  'package name renders on detail page'
);
assert(
  detailHtml.includes('v1.3.0'),
  'package version renders on detail page'
);
assert(
  detailHtml.includes('← All distros') || detailHtml.includes('All distros'),
  'detail page has back link'
);

if (failures > 0) {
  console.error(`\n${failures} assertion(s) failed.`);
  process.exit(1);
}
console.log('\nAll assertions passed.');
