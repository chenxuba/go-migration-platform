#!/usr/bin/env python3
from __future__ import annotations

import json
import math
import sys
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"

EXPECTED_DOMAINS = {
    "GM": "大运动",
    "FM": "精细动作",
    "AD": "适应能力",
    "LANG": "语言",
    "SOC": "社会行为",
}

EXPECTED_AGE_MONTHS = [
    1,
    2,
    3,
    4,
    5,
    6,
    7,
    8,
    9,
    10,
    11,
    12,
    15,
    18,
    21,
    24,
    27,
    30,
    33,
    36,
    42,
    48,
    54,
    60,
    66,
    72,
    78,
    84,
]

EXPECTED_ATTENTION_ITEMS = {5, 8, 14, 16, 34, 59, 86}

EXPECTED_PARENT_REPORT_ITEMS = {
    7,
    17,
    19,
    27,
    35,
    36,
    39,
    46,
    48,
    49,
    57,
    65,
    67,
    74,
    75,
    83,
    84,
    85,
    91,
    99,
    100,
    101,
    108,
    110,
    111,
    118,
    119,
    125,
    126,
    127,
    128,
    135,
    136,
    137,
    145,
    154,
    191,
    201,
}

HEADER_FRAGMENTS = ("附录", "测查项目", "操作方法", "测查通过要求")


def load_json(name: str):
    path = DOCS / name
    with path.open(encoding="utf-8") as handle:
        return json.load(handle)


def fail(message: str) -> None:
    print(f"erxin data validation failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def almost_equal(left: float, right: float) -> bool:
    return math.isclose(left, right, rel_tol=0, abs_tol=0.000001)


def validate_metadata(metadata: dict) -> None:
    if metadata.get("scale_code") != "ERXIN2":
        fail("metadata.scale_code must be ERXIN2")
    if metadata.get("item_count") != 261:
        fail("metadata.item_count must be 261")
    if metadata.get("domain_count") != 5:
        fail("metadata.domain_count must be 5")


def validate_domains(domains: list[dict]) -> None:
    actual = {item.get("scale_code"): item.get("scale_name") for item in domains}
    if actual != EXPECTED_DOMAINS:
        fail(f"domain map mismatch: {actual}")
    sort_values = [item.get("sort_no") for item in domains]
    if sort_values != [1, 2, 3, 4, 5]:
        fail(f"domain sort order mismatch: {sort_values}")


def validate_age_bands(age_bands: list[dict]) -> dict[int, dict]:
    by_month = {item.get("age_month"): item for item in age_bands}
    if sorted(by_month) != EXPECTED_AGE_MONTHS:
        fail(f"age band months mismatch: {sorted(by_month)}")
    for month, item in by_month.items():
        expected_total = 1.0 if month <= 12 else 3.0 if month <= 36 else 6.0
        if not almost_equal(float(item.get("domain_total_score", -1)), expected_total):
            fail(f"age band {month} has invalid domain total score")
    return by_month


def validate_item_bank(items: list[dict], age_bands: dict[int, dict]) -> None:
    if len(items) != 261:
        fail(f"item bank must contain 261 items, got {len(items)}")

    item_numbers = [int(item.get("item_no", 0)) for item in items]
    if item_numbers != list(range(1, 262)):
        fail("item_no values must be sorted and cover 1..261")

    counts = Counter((item["age_month"], item["domain_code"]) for item in items)
    weighted_totals = defaultdict(float)
    attention_items = set()
    parent_report_items = set()

    for item in items:
        item_no = item["item_no"]
        title = str(item.get("item_title", "")).strip()
        if not title:
            fail(f"item {item_no} has empty title")
        if "*" in title:
            fail(f"item {item_no} title still contains attention marker")
        if title.endswith("R") or title.startswith("R "):
            fail(f"item {item_no} title still contains parent report marker")

        age_month = item.get("age_month")
        if age_month not in age_bands:
            fail(f"item {item_no} has invalid age_month {age_month}")
        band = age_bands[age_month]
        if item.get("age_segment") != band.get("segment"):
            fail(f"item {item_no} has invalid age_segment")
        if not almost_equal(float(item.get("domain_month_total_score", -1)), float(band["domain_total_score"])):
            fail(f"item {item_no} has invalid domain_month_total_score")

        domain_code = item.get("domain_code")
        if EXPECTED_DOMAINS.get(domain_code) != item.get("domain_name"):
            fail(f"item {item_no} has invalid domain {domain_code}/{item.get('domain_name')}")

        if not isinstance(item.get("parent_report_allowed"), bool):
            fail(f"item {item_no} parent_report_allowed must be boolean")
        if item["parent_report_allowed"]:
            parent_report_items.add(item_no)

        if not isinstance(item.get("attention_if_failed"), bool):
            fail(f"item {item_no} attention_if_failed must be boolean")
        if item["attention_if_failed"]:
            attention_items.add(item_no)

        source_pages = item.get("source_pages")
        if item.get("source_pdf") != "儿心.pdf" or not isinstance(source_pages, list) or not source_pages:
            fail(f"item {item_no} must include source_pdf/source_pages")

        method = str(item.get("method", "")).strip()
        criteria = str(item.get("pass_criteria", "")).strip()
        if not method:
            fail(f"item {item_no} has empty method")
        if not criteria:
            fail(f"item {item_no} has empty pass_criteria")
        if any(fragment in method for fragment in HEADER_FRAGMENTS):
            fail(f"item {item_no} method contains header text")
        if any(fragment in criteria for fragment in HEADER_FRAGMENTS):
            fail(f"item {item_no} pass_criteria contains header text")

        weighted_totals[(age_month, domain_code)] += float(item.get("item_weight", 0))

    bad_counts = {key: value for key, value in counts.items() if value not in (1, 2)}
    if bad_counts:
        fail(f"every age/domain must have 1 or 2 items, got {bad_counts}")

    for key, count in counts.items():
        age_month, _ = key
        expected_item_weight = float(age_bands[age_month]["domain_total_score"]) / count
        for item in items:
            if (item["age_month"], item["domain_code"]) == key and not almost_equal(float(item["item_weight"]), expected_item_weight):
                fail(f"item {item['item_no']} has invalid item_weight")
        if not almost_equal(weighted_totals[key], float(age_bands[age_month]["domain_total_score"])):
            fail(f"weighted total mismatch for {key}")

    if attention_items != EXPECTED_ATTENTION_ITEMS:
        fail(f"attention item set mismatch: {sorted(attention_items)}")
    if parent_report_items != EXPECTED_PARENT_REPORT_ITEMS:
        fail(f"parent report item set mismatch: {sorted(parent_report_items)}")


def main() -> None:
    metadata = load_json("erxin-scale-metadata.json")
    domains = load_json("erxin-domain-map.json")
    age_bands = load_json("erxin-age-bands.json")
    items = load_json("erxin-item-bank-draft.json")

    validate_metadata(metadata)
    validate_domains(domains)
    band_by_month = validate_age_bands(age_bands)
    validate_item_bank(items, band_by_month)

    print("erxin data validation ok")


if __name__ == "__main__":
    main()
