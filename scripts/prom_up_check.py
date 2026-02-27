#!/usr/bin/env python3
import json, sys, urllib.parse, urllib.request

PROM = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:9090"
expected = sys.argv[2:]
if not expected:
    expected = ["gateway", "query", "alerts", "governance", "aggregator", "features", "inference", "simulator"]

q = urllib.parse.urlencode({"query": "up"}).encode()
url = f"{PROM}/api/v1/query?{q.decode()}"
with urllib.request.urlopen(url, timeout=5) as r:
    data = json.loads(r.read().decode())

if data.get("status") != "success":
    print("FAIL: prom query failed:", data)
    sys.exit(1)

series = data["data"]["result"]
job_vals = {}
for s in series:
    m = s.get("metric", {})
    job = m.get("job", "")
    val = float(s.get("value", [None, "0"])[1])
    if job:
        job_vals[job] = max(job_vals.get(job, 0.0), val)

missing = [j for j in expected if j not in job_vals]
down = [j for j in expected if job_vals.get(j, 0.0) < 1.0]

if missing or down:
    print("FAIL: Prometheus up check")
    if missing: print("  Missing jobs:", missing)
    if down:    print("  Down jobs:", down, "vals:", {j: job_vals.get(j) for j in down})
    print("  Known jobs:", sorted(job_vals.keys()))
    sys.exit(1)

print("PASS: Prometheus up==1 for:", expected)
