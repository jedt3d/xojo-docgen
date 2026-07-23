# Build the Landmark reader payload from MkDocs-rendered page HTML.

from __future__ import annotations

import html
import json
import re
from pathlib import Path


_documents: list[dict[str, str]] = []
_heading_pattern = re.compile(
    r'<h([1-3])\s+id="([^"]+)"[^>]*>([\s\S]*?)</h\1>',
    re.IGNORECASE,
)
_tag_pattern = re.compile(r"<[^>]+>")


def _heading_text(markup: str) -> str:
    without_permalink = re.sub(
        r'<a[^>]*class="headerlink"[^>]*>[\s\S]*?</a>',
        "",
        markup,
        flags=re.IGNORECASE,
    )
    return html.unescape(_tag_pattern.sub("", without_permalink)).strip()


def on_config(config, **kwargs):
    del kwargs
    _documents.clear()
    return config


def on_page_content(output: str, page, **kwargs):
    del kwargs
    headings = list(_heading_pattern.finditer(output))
    if not headings:
        return output

    page_location = page.url
    root_heading = next((item for item in headings if item.group(1) == "1"), None)
    if root_heading is not None:
        next_section = next(
            (item for item in headings if item.start() > root_heading.end() and item.group(1) == "2"),
            None,
        )
        root_end = next_section.start() if next_section is not None else len(output)
        _documents.append(
            {
                "location": page_location,
                "title": _heading_text(root_heading.group(3)),
                "text": output[root_heading.end():root_end].strip(),
            }
        )

    section_headings = [item for item in headings if item.group(1) in {"2", "3"}]
    for index, heading in enumerate(section_headings):
        level = int(heading.group(1))
        end = len(output)
        for candidate in section_headings[index + 1:]:
            candidate_level = int(candidate.group(1))
            if level == 2 or candidate_level <= level:
                end = candidate.start()
                break
        _documents.append(
            {
                "location": f"{page_location}#{heading.group(2)}",
                "title": _heading_text(heading.group(3)) or page.title,
                "text": output[heading.end():end].strip(),
            }
        )

    return output


def on_post_build(config, **kwargs):
    del kwargs
    output = Path(config["site_dir"]) / "data" / "documents.json"
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(
        json.dumps({"docs": _documents}, ensure_ascii=False, separators=(",", ":")),
        encoding="utf-8",
    )
