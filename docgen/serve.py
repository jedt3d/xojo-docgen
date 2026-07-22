#!/usr/bin/env python3
"""Route requests to per-project doc sites by Host header for local testing.

lvh.me resolves to 127.0.0.1, so http://<project>.lvh.me:<port>/ reaches this
server, which serves docs/api-published/<project>/ based on the first label of
the Host header. e.g. http://ee_web.lvh.me:8910/ -> docs/api-published/eeweb/.

Usage:
    python3 tools/docgen/serve.py [--port 8910] [--root docs/api-published]
"""
import argparse
import http.server
import os
import socketserver
import sys
from pathlib import Path
from urllib.parse import urlparse


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--port", type=int, default=8910)
    ap.add_argument("--root", default="docs/api-published",
                    help="dir containing per-project site subdirs")
    args = ap.parse_args()

    root = Path(args.root).resolve()
    if not root.is_dir():
        sys.exit(f"error: {root} does not exist. Run 'make docs' first.")

    class Handler(http.server.SimpleHTTPRequestHandler):
        def translate_path(self, path):
            # Pick the project subdir from the Host header's first label.
            host = self.headers.get("Host", "").split(":")[0]
            first_label = host.split(".")[0].lower()
            project_dir = root / first_label
            if not project_dir.is_dir():
                # Fall back to serving the root (listing of all projects).
                project_dir = root
            rel = urlparse(path).path
            if rel.startswith("/"):
                rel = rel[1:]
            candidate = (project_dir / rel).resolve()
            # Directory requests -> index.html
            if candidate.is_dir():
                candidate = candidate / "index.html"
            return str(candidate)

        def end_headers(self):
            self.send_header("Cache-Control", "no-store")
            super().end_headers()

        def log_message(self, fmt, *a):
            sys.stderr.write("%s - %s\n" % (self.address_string(), fmt % a))

    with socketserver.TCPServer(("0.0.0.0", args.port), Handler) as httpd:
        sites = sorted(p.name for p in root.iterdir() if (p / "index.html").exists())
        print(f"Serving {len(sites)} project site(s) from {root}/")
        for s in sites:
            print(f"  http://{s}.lvh.me:{args.port}/   (also http://{s}.localhost:{args.port}/)")
        print(f"\nProject index: http://lvh.me:{args.port}/")
        print("Ctrl-C to stop.\n")
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\nStopping.")


if __name__ == "__main__":
    main()
