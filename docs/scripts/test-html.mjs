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

// --- registry/y4n/backend-python-base/index.html ---
// Distros with slashes in their name must generate a detail page under a nested path.
const slashDistroHtml = readFileSync('dist/registry/y4n/backend-python-base/index.html', 'utf8');

assert(
  slashDistroHtml.includes('backend-python-base') || slashDistroHtml.includes('y4n/backend-python-base'),
  'slash-named distro detail page renders distro name'
);
assert(
  slashDistroHtml.includes('← All distros') || slashDistroHtml.includes('All distros'),
  'slash-named distro detail page has back link'
);

// Index must also link to the slash-named distro's detail page
assert(
  indexHtml.includes('href="/dpod-seed/registry/y4n/backend-python-base/"'),
  'index card links to slash-named distro detail page'
);

// --- README and source link on detail page ---
assert(
  detailHtml.includes('k8s-tools and zsh'),
  'detail page renders README body content'
);
assert(
  detailHtml.includes('href="https://github.com/iamy4n-dev/distros/tree/v0.2.0/distros/devops-k8s"'),
  'detail page has GitHub source link'
);

if (failures > 0) {
  console.error(`\n${failures} assertion(s) failed.`);
  process.exit(1);
}
console.log('\nAll assertions passed.');
