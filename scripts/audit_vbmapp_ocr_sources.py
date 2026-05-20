#!/usr/bin/env python3
"""Audit generated VB-MAPP data against OCR text from the upper/lower books.

The upper guide is the primary source for item-level scoring logic. The lower
overview is the child record booklet and is useful for item order, repeated
assessment slots, barriers, transition, and task-analysis tracking. OCR is not
perfect, so this script only auto-corrects conservative cases and leaves
source conflicts for manual review.
"""

from __future__ import annotations

import argparse
import difflib
import json
import re
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[1]
DATA_DIR = REPO_ROOT / "docs" / "vbmapp"
MILESTONE_ITEMS = DATA_DIR / "milestone-items.json"
LOWER_OCR = Path("/tmp/vbmapp_ocr/lower_overview_ocr.txt")
UPPER_OCR = Path("/tmp/vbmapp_ocr/upper_guide_ocr.txt")

AUDIT_JSON = DATA_DIR / "ocr-source-audit.json"
AUDIT_MD = DATA_DIR / "ocr-source-audit.md"
CORRECTIONS_JSON = DATA_DIR / "milestone-source-corrections.json"


DOMAIN_HEADING_PATTERNS: dict[str, list[str]] = {
    "MAND": [r"提\s*要求"],
    "TACT": [r"命\s*名"],
    "LR": [r"听者\s*反应", r"听者\s*技能", r"听者"],
    "VP_MTS": [r"VP\s*-\s*MTS", r"视觉\s*/\s*配对", r"视觉.*?配对"],
    "INDEPENDENT_PLAY": [r"独立\s*游戏", r"游戏"],
    "SOCIAL": [r"社会.*?游戏", r"社交"],
    "MOTOR_IMITATION": [r"动作\s*模仿", r"模仿"],
    "ECHOIC": [r"仿\s*说"],
    "SPONT_VOCAL": [r"语\s*音"],
    "LRFFC": [r"LRFFC"],
    "INTRAVERBAL": [r"对\s*话"],
    "GROUP": [r"集体\s*能力", r"教室.*?集体"],
    "LINGUISTIC_STRUCTURE": [r"语言\s*结构", r"语言"],
    "READING": [r"阅\s*读"],
    "WRITING": [r"书\s*写"],
    "MATH": [r"算\s*术", r"算\s*数"],
}


def load_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def dump_json(path: Path, value: Any) -> None:
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def clean_text(value: str) -> str:
    value = value.replace("\u3000", " ")
    value = value.replace("（0）", "（O）").replace("(0)", "(O)")
    value = re.sub(r"\s+", " ", value)
    return value.strip()


def normalize_mode(value: str) -> str:
    text = clean_text(value)
    if not text:
        return ""
    text = text.replace("０", "0").replace("O钟", "0钟").replace("O秒", "0秒")
    text = re.sub(r"(?<=\d)O(?=\d|分钟|秒|周)", "0", text)
    text = text.replace("TO ", "TO").replace("TO：", "TO:")
    text = re.sub(r"TO\s*:?\s*", "TO:", text)
    text = text.strip("()（）")
    text = text.replace("0", "O") if text in {"0"} else text
    if text in {"T", "O", "E"}:
        return text
    match = re.search(r"TO:\s*([0-9]+)\s*(分钟|秒|周)", text)
    if match:
        return f"TO：{match.group(1)}{match.group(2)}"
    if text.startswith("TO"):
        return text.replace(":", "：")
    return text


def extract_mode(text: str) -> str:
    candidates: list[str] = []
    for match in re.finditer(r"[（(]\s*(TO\s*[:：]?\s*[0-9O]+\s*(?:分钟|秒|周)|[TEO0])\s*[）)]", text):
        candidates.append(match.group(1))
    if candidates:
        return normalize_mode(candidates[-1])
    match = re.search(r"TO\s*[:：]?\s*[0-9O]+\s*(?:分钟|秒|周)", text)
    if match:
        return normalize_mode(match.group(0))
    return ""


def strip_mode(text: str) -> str:
    text = re.sub(r"[（(]\s*(?:TO\s*[:：]?\s*[0-9O]+\s*(?:分钟|秒|周)|[TEO0])\s*[）)]", "", text)
    return clean_text(text)


def page_range_text(ocr_text: str, start_page: int, end_page: int) -> str:
    out: list[str] = []
    for page_no, body in iter_pages(ocr_text):
        if start_page <= page_no <= end_page:
            out.append(f"\n--- page {page_no} ---\n{body}")
    return "\n".join(out)


def iter_pages(ocr_text: str) -> list[tuple[int, str]]:
    parts = re.split(r"\n--- page (\d+) ---\n", ocr_text)
    pages: list[tuple[int, str]] = []
    for index in range(1, len(parts), 2):
        pages.append((int(parts[index]), parts[index + 1]))
    return pages


def parse_lower_milestones(lower_text: str, expected_count: int) -> list[dict[str, Any]]:
    block = page_range_text(lower_text, 2, 20)
    raw_lines = [line.strip() for line in block.splitlines() if line.strip()]
    lines: list[str] = []
    for line in raw_lines:
        if line.startswith("--- page"):
            continue
        if line.startswith("第") and "次" in line:
            continue
        if line in {"评估", "评", "估", "总分：", "备注："}:
            continue
        if line.startswith(("里程碑评估", "VB-MAPP", "Copyright")):
            continue
        if line.startswith(("（T", "（O", "（E", "（TO", "(T", "(O", "(E", "(TO")):
            continue
        lines.append(line)
    clean = "\n".join(lines)
    starts = list(re.finditer(r"(?m)^(\d{1,2})\.\s*", clean))
    entries: list[dict[str, Any]] = []
    for index, match in enumerate(starts):
        end = starts[index + 1].start() if index + 1 < len(starts) else len(clean)
        chunk = clean[match.start() : end].strip()
        no = int(match.group(1))
        title = clean_text(re.sub(r"^\d{1,2}\.\s*", "", chunk))
        entries.append(
            {
                "sourceOrdinal": index + 1,
                "lowerMilestoneNo": no,
                "lowerTitle": title,
                "lowerTitleWithoutMode": strip_mode(title),
                "lowerMode": extract_mode(title),
            }
        )
    if len(entries) != expected_count:
        raise RuntimeError(f"expected {expected_count} lower milestone entries, got {len(entries)}")
    return entries


def scoring_chapter_text(upper_text: str) -> str:
    lines = upper_text.splitlines()
    start = None
    end = None
    for index, line in enumerate(lines):
        if start is None and index > 1000 and line.strip() == "第三章":
            start = index
            continue
        if start is not None and index > start and line.strip() == "第六章":
            end = index
            break
    if start is None or end is None:
        raise RuntimeError("cannot locate upper guide milestone scoring chapters")
    return "\n".join(lines[start:end])


def heading_regex() -> re.Pattern[str]:
    alternatives: list[str] = []
    for domain, patterns in DOMAIN_HEADING_PATTERNS.items():
        for pattern in patterns:
            alternatives.append(rf"(?P<{domain}_{len(alternatives)}>{pattern})")
    alias_pattern = "|".join(alternatives)
    return re.compile(rf"(?P<alias>{alias_pattern})\s*(?P<no>\d{{1,2}})\s*-\s*M")


def domain_from_heading(alias: str) -> str:
    compact = re.sub(r"\s+", "", alias)
    for domain, patterns in DOMAIN_HEADING_PATTERNS.items():
        for pattern in patterns:
            if re.fullmatch(pattern.replace(r"\s*", r"\s*"), alias) or re.search(pattern, alias):
                return domain
        if domain == "VP_MTS" and "视觉" in compact and "配对" in compact:
            return domain
    raise ValueError(f"unknown heading alias: {alias!r}")


def milestone_id(domain: str, no: int) -> str:
    return f"{domain}_{no:02d}M"


def parse_upper_sections(upper_text: str) -> dict[str, dict[str, Any]]:
    chapter = scoring_chapter_text(upper_text)
    regex = heading_regex()
    matches = list(regex.finditer(chapter))
    sections: dict[str, dict[str, Any]] = {}
    for index, match in enumerate(matches):
        alias = match.group("alias")
        domain = domain_from_heading(alias)
        no = int(match.group("no"))
        item_id = milestone_id(domain, no)
        if item_id in sections:
            continue
        end = matches[index + 1].start() if index + 1 < len(matches) else len(chapter)
        text = clean_text(chapter[match.start() : end])
        # Ignore accidental cross-domain false positives with no scoring markers.
        if "得1分" not in text and "目标：" not in text:
            continue
        sections[item_id] = {
            "upperHeading": clean_text(match.group(0)),
            "upperMode": extract_mode(text[:280]),
            "upperSectionLength": len(text),
            "upperHasGoal": "目标：" in text or "自标：" in text,
            "upperHasMaterials": "材料：" in text,
            "upperHasExamples": "例子：" in text,
            "upperHasOnePoint": "得1分" in text,
            "upperHasHalfPoint": "得½分" in text or "得1/2分" in text or "给1/2分" in text,
            "upperPreview": text[:500],
        }
    return sections


def similarity(left: str, right: str) -> float:
    left = strip_mode(left)
    right = strip_mode(right)
    if not left and not right:
        return 1
    return round(difflib.SequenceMatcher(None, left, right).ratio(), 4)


def source_status(current_mode: str, lower_mode: str, upper_mode: str) -> tuple[str, str, str]:
    current = normalize_mode(current_mode)
    lower = normalize_mode(lower_mode)
    upper = normalize_mode(upper_mode)
    if not current and lower and (not upper or upper == lower):
        return "safe_fill_current_empty", lower, "当前测量方式为空，下册/上册未冲突。"
    if current and lower and current != lower:
        if upper and upper == current:
            return "lower_conflicts_upper_current", current, "下册OCR与当前不同，但上册支持当前值。"
        if upper and upper == lower:
            return "safe_replace_current", lower, "当前测量方式与上下册一致结果不同。"
        return "source_conflict", current, "当前、下册或上册存在冲突，需要人工复核。"
    if current and upper and current != upper:
        return "source_conflict", current, "当前测量方式与上册OCR不同，需要人工复核。"
    return "ok", current, ""


def build_audit(apply: bool) -> dict[str, Any]:
    if not LOWER_OCR.exists() or not UPPER_OCR.exists():
        raise FileNotFoundError("OCR files not found under /tmp/vbmapp_ocr; run scripts/ocr_pdf_vision.swift first")
    items: list[dict[str, Any]] = load_json(MILESTONE_ITEMS)
    lower_entries = parse_lower_milestones(LOWER_OCR.read_text(encoding="utf-8"), len(items))
    upper_sections = parse_upper_sections(UPPER_OCR.read_text(encoding="utf-8"))

    rows: list[dict[str, Any]] = []
    corrections: list[dict[str, Any]] = []
    for item, lower in zip(items, lower_entries):
        upper = upper_sections.get(item["milestoneId"], {})
        status, recommended_mode, note = source_status(
            item.get("assessmentMode", ""),
            lower.get("lowerMode", ""),
            upper.get("upperMode", ""),
        )
        row = {
            "milestoneId": item["milestoneId"],
            "label": item["label"],
            "domainCode": item["domainCode"],
            "milestoneNo": item["milestoneNo"],
            "currentMode": normalize_mode(item.get("assessmentMode", "")),
            "lowerMode": lower.get("lowerMode", ""),
            "upperMode": upper.get("upperMode", ""),
            "status": status,
            "recommendedMode": recommended_mode,
            "note": note,
            "titleSimilarityWithLower": similarity(item["title"], lower["lowerTitle"]),
            "upperSectionFound": bool(upper),
            "upperHasMaterials": bool(upper.get("upperHasMaterials")),
            "upperHasExamples": bool(upper.get("upperHasExamples")),
            "upperHasOnePoint": bool(upper.get("upperHasOnePoint")),
            "upperHasHalfPoint": bool(upper.get("upperHasHalfPoint")),
        }
        if status in {"safe_fill_current_empty", "safe_replace_current"} and recommended_mode != row["currentMode"]:
            corrections.append(
                {
                    "milestoneId": item["milestoneId"],
                    "field": "assessmentMode",
                    "from": row["currentMode"],
                    "to": recommended_mode,
                    "reason": note,
                }
            )
        rows.append(row)

    if apply and corrections:
        by_id = {correction["milestoneId"]: correction for correction in corrections}
        for item in items:
            correction = by_id.get(item["milestoneId"])
            if correction:
                item["assessmentMode"] = correction["to"]
        dump_json(MILESTONE_ITEMS, items)

    summary = {
        "milestoneCount": len(items),
        "lowerParsedCount": len(lower_entries),
        "upperSectionCount": len(upper_sections),
        "statusCounts": {},
        "safeCorrectionCount": len(corrections),
        "applied": apply,
    }
    for row in rows:
        summary["statusCounts"][row["status"]] = summary["statusCounts"].get(row["status"], 0) + 1

    audit = {
        "summary": summary,
        "corrections": corrections,
        "rows": rows,
        "upperMissingMilestoneIds": [
            item["milestoneId"] for item in items if item["milestoneId"] not in upper_sections
        ],
    }
    dump_json(AUDIT_JSON, audit)
    dump_json(CORRECTIONS_JSON, corrections)
    AUDIT_MD.write_text(render_markdown(audit), encoding="utf-8")
    return audit


def render_markdown(audit: dict[str, Any]) -> str:
    lines = [
        "# VB-MAPP OCR 来源核对报告",
        "",
        "## 汇总",
        "",
        f"- 里程碑题项：{audit['summary']['milestoneCount']}",
        f"- 下册解析题项：{audit['summary']['lowerParsedCount']}",
        f"- 上册逐项说明匹配：{audit['summary']['upperSectionCount']}",
        f"- 可安全修正：{audit['summary']['safeCorrectionCount']}",
        f"- 已应用修正：{'是' if audit['summary']['applied'] else '否'}",
        "",
        "状态统计：",
        "",
    ]
    for status, count in sorted(audit["summary"]["statusCounts"].items()):
        lines.append(f"- `{status}`：{count}")
    lines.extend(["", "## 可安全修正", ""])
    if audit["corrections"]:
        lines.append("| 项目 | 字段 | 原值 | 修正值 | 原因 |")
        lines.append("|---|---|---|---|---|")
        for correction in audit["corrections"]:
            lines.append(
                f"| `{correction['milestoneId']}` | `{correction['field']}` | "
                f"{correction['from'] or '-'} | {correction['to']} | {correction['reason']} |"
            )
    else:
        lines.append("无。")
    conflicts = [row for row in audit["rows"] if "conflict" in row["status"]]
    lines.extend(["", "## 需要人工复核", ""])
    if conflicts:
        lines.append("| 项目 | 当前 | 下册 | 上册 | 说明 |")
        lines.append("|---|---:|---:|---:|---|")
        for row in conflicts:
            lines.append(
                f"| `{row['milestoneId']}` | {row['currentMode'] or '-'} | "
                f"{row['lowerMode'] or '-'} | {row['upperMode'] or '-'} | {row['note']} |"
            )
    else:
        lines.append("无。")
    missing = audit["upperMissingMilestoneIds"]
    lines.extend(["", "## 上册未匹配项", ""])
    lines.append("、".join(f"`{item}`" for item in missing) if missing else "无。")
    return "\n".join(lines) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--apply", action="store_true", help="apply safe corrections to milestone-items.json")
    args = parser.parse_args()
    audit = build_audit(apply=args.apply)
    print(json.dumps(audit["summary"], ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
