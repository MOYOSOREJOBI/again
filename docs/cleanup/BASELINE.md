# Baseline Snapshot

- HEAD: `fe3caa2b2051c3a9da306e0422f41dd95526db53`
- Branch: `chore/cleanup-branch-hygiene`
- Timestamp: `2026-02-27T13:57:22Z`

## Commands Run

```bash
git status --porcelain
git log --oneline -n 20
git diff --stat origin/work..HEAD || true
rg -n "TODO\b|FIXME\b|HACK\b" -S .
rg -n "merge-conflict marker patterns" -S .
du -ah . | sort -rh | head -n 50
```

## Raw Output

    ## git status --porcelain

    ## git log --oneline -n 20
    fe3caa2 Merge pull request #60 from MOYOSOREJOBI/codex/replace-best-effort-checks-with-hard-evidence-markers-fv442l
    9b0ec1c Merge branch 'NEWMAIN' into codex/replace-best-effort-checks-with-hard-evidence-markers-fv442l
    4d3384f docs(cleanup): add final report and sanitize conflict-pattern docs
    0aa4581 Merge pull request #59 from MOYOSOREJOBI/codex/replace-best-effort-checks-with-hard-evidence-markers-if62mt
    f08ce14 Merge branch 'NEWMAIN' into codex/replace-best-effort-checks-with-hard-evidence-markers-if62mt
    eb57c23 Align gate and audit with strict marker-based runtime proof flow
    51057f8 Merge pull request #58 from MOYOSOREJOBI/codex/replace-best-effort-checks-with-hard-evidence-markers-rwb585
    b62a041 Merge branch 'NEWMAIN' into codex/replace-best-effort-checks-with-hard-evidence-markers-rwb585
    cf4af2c Refactor audit flow to score strictly from proof markers
    ea66446 Merge pull request #57 from MOYOSOREJOBI/codex/replace-best-effort-checks-with-hard-evidence-markers-elpsbu
    c051971 Add query seed-status handler with counted table fields
    b3691d4 Merge pull request #56 from MOYOSOREJOBI/codex/replace-best-effort-checks-with-hard-evidence-markers-57sibs
    71f3a4b Make Prometheus up-check derive jobs from config
    1a4bf5c Merge pull request #55 from MOYOSOREJOBI/codex/replace-best-effort-checks-with-hard-evidence-markers
    4eecf56 Replace runtime proof script with hard evidence markers
    eddb0e4 Merge pull request #54 from MOYOSOREJOBI/codex/improve-playwright-reliability-in-compose
    677d67f Run Playwright via compose e2e service
    593a0ac Merge pull request #53 from MOYOSOREJOBI/codex/implement-end-to-end-sentinel-audit-process-ixacgt
    15ad6d5 Merge branch 'NEWMAIN' into codex/implement-end-to-end-sentinel-audit-process-ixacgt
    a63324f audit: always emit scored result even when gate commands fail

    ## git diff --stat origin/work..HEAD
    fatal: ambiguous argument 'origin/work..HEAD': unknown revision or path not in the working tree.
    Use '--' to separate paths from revisions, like this:
    'git <command> [<revision>...] -- [<file>...]'

    ## TODO/FIXME/HACK
    ./docs/cleanup/baseline_commands.log:25:$ rg -n "TODO\\b|FIXME\\b|HACK\\b" -S .

    ## conflict markers

    ## size offenders
    453M	.
    451M	./web/node_modules
    451M	./web
    275M	./web/node_modules/@next
    150M	./web/node_modules/@next/swc-linux-x64-musl/next-swc.linux-x64-musl.node
    150M	./web/node_modules/@next/swc-linux-x64-musl
    126M	./web/node_modules/@next/swc-linux-x64-gnu/next-swc.linux-x64-gnu.node
    126M	./web/node_modules/@next/swc-linux-x64-gnu
    102M	./web/node_modules/next/dist
    102M	./web/node_modules/next
    65M	./web/node_modules/next/dist/compiled
    22M	./web/node_modules/typescript/lib
    22M	./web/node_modules/typescript
    18M	./web/node_modules/next/dist/compiled/next-server
    13M	./web/node_modules/next/dist/server
    13M	./web/node_modules/next/dist/esm
    11M	./web/node_modules/next/dist/compiled/react-dom-experimental/cjs
    11M	./web/node_modules/next/dist/compiled/react-dom-experimental
    8.5M	./web/node_modules/typescript/lib/typescript.js
    8.0M	./web/node_modules/next/dist/compiled/react-dom
    7.9M	./web/node_modules/next/dist/compiled/react-dom/cjs
    6.7M	./web/node_modules/next/dist/server/lib
    6.4M	./web/node_modules/tailwindcss
    5.9M	./web/node_modules/next/dist/server/lib/squoosh
    5.8M	./web/node_modules/typescript/lib/tsc.js
    5.4M	./web/node_modules/recharts
    5.1M	./web/node_modules/next/dist/esm/server
    4.9M	./web/node_modules/lodash
    4.4M	./web/node_modules/react-dom
    4.3M	./web/node_modules/tailwindcss/peers/index.js
    4.3M	./web/node_modules/tailwindcss/peers
    4.2M	./web/node_modules/next/dist/build
    4.2M	./web/node_modules/caniuse-lite
    4.1M	./web/node_modules/next/dist/server/lib/squoosh/avif
    4.1M	./web/node_modules/caniuse-lite/data
    3.3M	./web/node_modules/next/dist/client
    3.2M	./web/node_modules/next/dist/compiled/@vercel
    3.2M	./web/node_modules/@types
    3.1M	./web/node_modules/next/dist/esm/build
    3.1M	./web/node_modules/next/dist/compiled/next-server/app-page-experimental.runtime.dev.js.map
    2.9M	./web/node_modules/next/dist/compiled/@vercel/og
    2.8M	./web/node_modules/next/dist/compiled/next-server/app-page.runtime.dev.js.map
    2.6M	./web/node_modules/recharts/umd
    2.6M	./web/node_modules/next/dist/server/lib/squoosh/avif/avif_node_enc.wasm
    2.6M	./web/node_modules/next/dist/compiled/webpack
    2.5M	./web/node_modules/react-dom/cjs
    2.5M	./web/node_modules/next/dist/compiled/webpack/bundle5.js
    2.5M	./web/node_modules/next/dist/client/components
    2.4M	./web/node_modules/next/dist/esm/client
    2.4M	./web/node_modules/caniuse-lite/data/features
