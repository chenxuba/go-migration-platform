#!/usr/bin/env python3
"""Extract VB-MAPP seed data from the local source documents.

The script intentionally keeps extraction deterministic and conservative:
it only reads the source files provided in ~/Downloads and writes structured
JSON files under docs/vbmapp.
"""

from __future__ import annotations

import json
import re
from collections import Counter, defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from docx import Document
from pypdf import PdfReader


REPO_ROOT = Path(__file__).resolve().parents[1]
DOWNLOADS = Path("/Users/chenrui/Downloads")
OUTPUT_DIR = REPO_ROOT / "docs" / "vbmapp"

SCALE_CODE = "VBMAPP"
SCALE_VERSION = "VBMAPP_CN_2ND_DRAFT_2026_05"

MILESTONE_DOCS = [
    {
        "level": 1,
        "ageBand": "0-18个月",
        "source": DOWNLOADS
        / "VB-MAPP第一阶段孤独症儿童测评量表社会行为和社会性游戏-922f359d0a4e767f5acfa1c7aa00b52acfc79cbc.docx",
    },
    {
        "level": 2,
        "ageBand": "18-30个月",
        "source": DOWNLOADS
        / "VB-MAPP第二阶段(1)孤独症儿童测评量表社会行为和社会性游戏-6bce59e30a4c2e3f5727a5e9856a561253d321f6.docx",
    },
    {
        "level": 3,
        "ageBand": "30-48个月",
        "source": DOWNLOADS
        / "VB-MAPP第三阶段孤独症儿童测评量表社会行为和社会性游戏-adba7191fe00bed5b9f3f90f76c66137ee064fbd.docx",
    },
]

MILESTONE_RULES_PDF = DOWNLOADS / "VB-MAPP第一部分------里程碑评估.pdf"
BARRIERS_PDF = DOWNLOADS / "VB-MAPP第二部分------障碍评估.pdf"
TRANSITION_PDF = DOWNLOADS / "VB-MAPP第三部分------转衔评估.pdf"

DOMAIN_ALIASES = {
    "提要求": ("MAND", "提要求"),
    "命名": ("TACT", "命名"),
    "听者反应": ("LR", "听者反应"),
    "听者技能": ("LR", "听者反应"),
    "视觉/配对": ("VP_MTS", "视觉感知和样本配对"),
    "视觉感知和样本配对": ("VP_MTS", "视觉感知和样本配对"),
    "VP-MTS": ("VP_MTS", "视觉感知和样本配对"),
    "独立游戏": ("INDEPENDENT_PLAY", "独立游戏"),
    "游戏": ("INDEPENDENT_PLAY", "独立游戏"),
    "社会游戏": ("SOCIAL", "社会行为和社会游戏"),
    "社会行为和社会游戏": ("SOCIAL", "社会行为和社会游戏"),
    "社会行为和社交游戏": ("SOCIAL", "社会行为和社会游戏"),
    "社交行为和社交游戏": ("SOCIAL", "社会行为和社会游戏"),
    "社会行为与社会游戏": ("SOCIAL", "社会行为和社会游戏"),
    "社交": ("SOCIAL", "社会行为和社会游戏"),
    "模仿": ("MOTOR_IMITATION", "动作模仿"),
    "动作模仿": ("MOTOR_IMITATION", "动作模仿"),
    "仿说": ("ECHOIC", "仿说"),
    "仿说(EESA)": ("ECHOIC", "仿说"),
    "语音": ("SPONT_VOCAL", "自发性语音行为"),
    "自发性的语音行为": ("SPONT_VOCAL", "自发性语音行为"),
    "自发性发生行为": ("SPONT_VOCAL", "自发性语音行为"),
    "LRFFC": ("LRFFC", "功能、特性和类别的听者反应"),
    "功能，特性和类别的听者反应": ("LRFFC", "功能、特性和类别的听者反应"),
    "功能、特性和类别的听者反应": ("LRFFC", "功能、特性和类别的听者反应"),
    "功能、特性、类别的听者反应": ("LRFFC", "功能、特性和类别的听者反应"),
    "对功能、特性、类别的听者反应": ("LRFFC", "功能、特性和类别的听者反应"),
    "对话": ("INTRAVERBAL", "对话"),
    "集体能力": ("GROUP", "教室常规和集体能力"),
    "教室常规和集体能力": ("GROUP", "教室常规和集体能力"),
    "教室常规和集体技能": ("GROUP", "教室常规和集体能力"),
    "教室常规和集体指令": ("GROUP", "教室常规和集体能力"),
    "语言": ("LINGUISTIC_STRUCTURE", "语言结构"),
    "语言结构": ("LINGUISTIC_STRUCTURE", "语言结构"),
    "阅读": ("READING", "阅读"),
    "书写": ("WRITING", "书写"),
    "算术": ("MATH", "算术"),
    "算数": ("MATH", "算术"),
}

DOMAIN_SORT = [
    "MAND",
    "TACT",
    "LR",
    "VP_MTS",
    "INDEPENDENT_PLAY",
    "SOCIAL",
    "MOTOR_IMITATION",
    "ECHOIC",
    "SPONT_VOCAL",
    "LRFFC",
    "INTRAVERBAL",
    "GROUP",
    "LINGUISTIC_STRUCTURE",
    "READING",
    "WRITING",
    "MATH",
]

LEVEL_AGE_BANDS = {
    1: "0-18个月",
    2: "18-30个月",
    3: "30-48个月",
}


def clean_text(value: str) -> str:
    value = value.replace("\x00", " ")
    value = value.replace("\u3000", " ")
    value = re.sub(r"\s+", " ", value)
    return value.strip()


def normalize_domain(raw: str) -> tuple[str, str]:
    domain = clean_text(raw)
    if domain in DOMAIN_ALIASES:
        return DOMAIN_ALIASES[domain]
    raise ValueError(f"unknown VB-MAPP domain: {raw!r}")


def domain_sort_no(code: str) -> int:
    try:
        return DOMAIN_SORT.index(code) + 1
    except ValueError:
        return 999


def extract_assessment_mode(text: str) -> str:
    matches = re.findall(r"[（(]([^()（）]*?(?:TO|T|O|E)[^()（）]*?)[）)]", text)
    return clean_text(matches[-1]) if matches else ""


def milestone_key(domain_code: str, milestone_no: int) -> str:
    return f"{domain_code}_{milestone_no:02d}M"


def parse_milestone_items() -> list[dict[str, Any]]:
    items: list[dict[str, Any]] = []
    seq = 1
    pattern = re.compile(r"^(?P<domain>.+?)\s*(?P<no>\d+)\s*-\s*M$")

    for doc_info in MILESTONE_DOCS:
        source = doc_info["source"]
        if not source.exists():
            raise FileNotFoundError(source)
        doc = Document(str(source))
        for table_index, table in enumerate(doc.tables, start=1):
            if not table.rows or len(table.rows[0].cells) < 2:
                continue
            raw_id = clean_text(table.rows[0].cells[0].text)
            description = clean_text(table.rows[0].cells[1].text)
            match = pattern.match(raw_id)
            if not match:
                raise ValueError(f"cannot parse milestone id {raw_id!r} in {source.name}")
            domain_code, domain_name = normalize_domain(match.group("domain"))
            milestone_no = int(match.group("no"))
            level = int(doc_info["level"])
            expected_level = 1 if milestone_no <= 5 else 2 if milestone_no <= 10 else 3
            if expected_level != level:
                raise ValueError(f"level mismatch for {raw_id}: document={level}, item={expected_level}")
            items.append(
                {
                    "sequenceNo": seq,
                    "milestoneId": milestone_key(domain_code, milestone_no),
                    "label": f"{domain_name}{milestone_no}-M",
                    "domainCode": domain_code,
                    "domainName": domain_name,
                    "domainSortNo": domain_sort_no(domain_code),
                    "level": level,
                    "ageBand": doc_info["ageBand"],
                    "milestoneNo": milestone_no,
                    "title": description,
                    "assessmentMode": extract_assessment_mode(description),
                    "scoreType": "milestone_0_0_5_1",
                    "sourceFile": source.name,
                    "sourceTableIndex": table_index,
                }
            )
            seq += 1
    return items


def build_domains(items: list[dict[str, Any]]) -> list[dict[str, Any]]:
    grouped: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for item in items:
        grouped[item["domainCode"]].append(item)

    domains = []
    for code in sorted(grouped, key=domain_sort_no):
        rows = sorted(grouped[code], key=lambda x: (x["level"], x["milestoneNo"]))
        level_counts = Counter(row["level"] for row in rows)
        domains.append(
            {
                "domainCode": code,
                "domainName": rows[0]["domainName"],
                "sortNo": domain_sort_no(code),
                "itemCount": len(rows),
                "maxScore": len(rows),
                "levels": [
                    {
                        "level": level,
                        "ageBand": LEVEL_AGE_BANDS[level],
                        "itemCount": level_counts.get(level, 0),
                        "milestoneNos": [row["milestoneNo"] for row in rows if row["level"] == level],
                    }
                    for level in (1, 2, 3)
                    if level_counts.get(level, 0)
                ],
            }
        )
    return domains


def pdf_lines(path: Path) -> list[str]:
    reader = PdfReader(str(path))
    raw_lines = []
    for page in reader.pages:
        text = page.extract_text() or ""
        raw_lines.extend(text.splitlines())
    lines = []
    for line in raw_lines:
        value = clean_text(line)
        if not value:
            continue
        if value.startswith("更多资源请关注"):
            continue
        lines.append(value)
    return lines


def parse_milestone_scoring_rules() -> list[dict[str, Any]]:
    if not MILESTONE_RULES_PDF.exists():
        return []

    stage_pattern = re.compile(r"^第(?P<stage>[一二三])阶段[：:](?P<domain>.+)$")
    item_pattern = re.compile(r"^(?P<no>\d+)\s*M\s+(?P<desc>.+)$")
    stage_to_level = {"一": 1, "二": 2, "三": 3}

    current_level: int | None = None
    current_domain_code = ""
    current_domain_name = ""
    pending: dict[str, Any] | None = None
    rules: list[dict[str, Any]] = []

    def flush_without_rule() -> None:
        nonlocal pending
        if pending is not None:
            rules.append(pending)
            pending = None

    for line in pdf_lines(MILESTONE_RULES_PDF):
        stage_match = stage_pattern.match(line)
        if stage_match:
            flush_without_rule()
            current_level = stage_to_level[stage_match.group("stage")]
            current_domain_code, current_domain_name = normalize_domain(stage_match.group("domain"))
            continue

        item_match = item_pattern.match(line)
        if item_match and current_level and current_domain_code:
            flush_without_rule()
            milestone_no = int(item_match.group("no"))
            pending = {
                "milestoneId": milestone_key(current_domain_code, milestone_no),
                "domainCode": current_domain_code,
                "domainName": current_domain_name,
                "level": current_level,
                "milestoneNo": milestone_no,
                "description": clean_text(item_match.group("desc")),
                "onePointCriteria": "",
                "halfPointCriteria": "",
                "sourceFile": MILESTONE_RULES_PDF.name,
            }
            continue

        if pending is None:
            continue

        one_match = re.search(r"1\s*分[：:](.*?)(?=1/2\s*(?:分|个)[：:]|$)", line)
        half_match = re.search(r"1/2\s*(?:分|个)[：:](.*)$", line)
        if one_match:
            pending["onePointCriteria"] = clean_text(one_match.group(1))
            if half_match:
                pending["halfPointCriteria"] = clean_text(half_match.group(1))
            rules.append(pending)
            pending = None
        else:
            pending["description"] = clean_text(pending["description"] + " " + line)

    flush_without_rule()
    return rules


def parse_barriers() -> list[dict[str, Any]]:
    if not BARRIERS_PDF.exists():
        return []

    heading_pattern = re.compile(r"^(?P<name>.+?)的评估(?:记分|计分)?方法$")
    score_pattern = re.compile(r"^(?P<score>[1-4])\s*分(?P<desc>.+)$")
    barriers: list[dict[str, Any]] = []
    current: dict[str, Any] | None = None
    current_score: dict[str, Any] | None = None

    for line in pdf_lines(BARRIERS_PDF):
        if line == "VB-MAPP 障碍评估":
            continue
        heading_match = heading_pattern.match(line)
        if heading_match:
            name = clean_text(heading_match.group("name"))
            current = {
                "barrierNo": len(barriers) + 1,
                "barrierCode": f"B{len(barriers) + 1:02d}",
                "barrierName": name,
                "minScore": 0,
                "maxScore": 4,
                "scoreOptions": [
                    {"score": 0, "description": "未见明显相关障碍或不作为主要问题"},
                ],
                "sourceFile": BARRIERS_PDF.name,
            }
            barriers.append(current)
            current_score = None
            continue

        if current is None:
            continue
        score_match = score_pattern.match(line)
        if score_match:
            current_score = {
                "score": int(score_match.group("score")),
                "description": clean_text(score_match.group("desc")),
            }
            current["scoreOptions"].append(current_score)
        elif current_score is not None:
            current_score["description"] = clean_text(current_score["description"] + " " + line)

    return barriers


def parse_transition() -> list[dict[str, Any]]:
    if not TRANSITION_PDF.exists():
        return []

    category_pattern = re.compile(r"^转衔评估第(?P<name>[一二三])类")
    item_pattern = re.compile(r"^(?P<no>\d+)、(?P<name>.+)$")
    option_pattern = re.compile(r"^(?P<score>[1-5])、(?P<desc>.+)$")
    category_name = ""
    items: list[dict[str, Any]] = []
    current: dict[str, Any] | None = None
    current_option: dict[str, Any] | None = None
    in_recommendation = False

    for line in pdf_lines(TRANSITION_PDF):
        if line == "VB-MAPP 转衔评估":
            continue
        category_match = category_pattern.match(line)
        if category_match:
            category_name = f"第{category_match.group('name')}类"
            current = None
            current_option = None
            in_recommendation = False
            continue

        if current is None:
            item_match = item_pattern.match(line)
            if item_match:
                number = int(item_match.group("no"))
                if 1 <= number <= 18:
                    current = {
                        "transitionNo": number,
                        "transitionCode": f"T{number:02d}",
                        "transitionName": clean_text(item_match.group("name")),
                        "category": category_name,
                        "minScore": 1,
                        "maxScore": 5,
                        "scoreOptions": [],
                        "placementRecommendations": [],
                        "sourceFile": TRANSITION_PDF.name,
                    }
                    items.append(current)
                    current_option = None
                    in_recommendation = False
            continue

        if line == "评分标准及安置建议":
            in_recommendation = True
            current_option = None
            continue

        option_match = option_pattern.match(line)
        if option_match and not in_recommendation:
            current_option = {
                "score": int(option_match.group("score")),
                "description": clean_text(option_match.group("desc")),
            }
            current["scoreOptions"].append(current_option)
            continue

        item_match = item_pattern.match(line)
        if item_match:
            number = int(item_match.group("no"))
            if 1 <= number <= 18:
                current = {
                    "transitionNo": number,
                    "transitionCode": f"T{number:02d}",
                    "transitionName": clean_text(item_match.group("name")),
                    "category": category_name,
                    "minScore": 1,
                    "maxScore": 5,
                    "scoreOptions": [],
                    "placementRecommendations": [],
                    "sourceFile": TRANSITION_PDF.name,
                }
                items.append(current)
                current_option = None
                in_recommendation = False
                continue

        if in_recommendation:
            current["placementRecommendations"].append(line)
        elif current_option is not None:
            current_option["description"] = clean_text(current_option["description"] + " " + line)

    return sorted(items, key=lambda x: x["transitionNo"])


def write_json(path: Path, payload: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def validate_payloads(
    milestones: list[dict[str, Any]],
    domains: list[dict[str, Any]],
    rules: list[dict[str, Any]],
    barriers: list[dict[str, Any]],
    transition: list[dict[str, Any]],
) -> dict[str, Any]:
    errors: list[str] = []
    if len(milestones) != 170:
        errors.append(f"milestone item count expected 170, got {len(milestones)}")
    if len(domains) != 16:
        errors.append(f"domain count expected 16, got {len(domains)}")
    if rules and len(rules) != 170:
        errors.append(f"milestone scoring rule count expected 170, got {len(rules)}")
    if barriers and len(barriers) != 24:
        errors.append(f"barrier count expected 24, got {len(barriers)}")
    if transition and len(transition) != 18:
        errors.append(f"transition count expected 18, got {len(transition)}")

    milestone_ids = [item["milestoneId"] for item in milestones]
    duplicates = [mid for mid, count in Counter(milestone_ids).items() if count > 1]
    if duplicates:
        errors.append(f"duplicate milestone ids: {duplicates}")

    rule_ids = {rule["milestoneId"] for rule in rules}
    missing_rule_ids = [mid for mid in milestone_ids if rule_ids and mid not in rule_ids]
    if missing_rule_ids:
        errors.append(f"missing scoring rules for {len(missing_rule_ids)} milestones")

    return {
        "scaleCode": SCALE_CODE,
        "scaleVersion": SCALE_VERSION,
        "milestoneItemCount": len(milestones),
        "domainCount": len(domains),
        "milestoneScoringRuleCount": len(rules),
        "barrierCount": len(barriers),
        "transitionCount": len(transition),
        "domainItemCounts": {domain["domainCode"]: domain["itemCount"] for domain in domains},
        "errors": errors,
        "status": "ok" if not errors else "needs_review",
    }


def main() -> None:
    milestones = parse_milestone_items()
    domains = build_domains(milestones)
    rules = parse_milestone_scoring_rules()
    barriers = parse_barriers()
    transition = parse_transition()
    summary = validate_payloads(milestones, domains, rules, barriers, transition)

    write_json(OUTPUT_DIR / "milestone-items.json", milestones)
    write_json(OUTPUT_DIR / "domains.json", domains)
    write_json(OUTPUT_DIR / "milestone-scoring-rules.json", rules)
    write_json(OUTPUT_DIR / "barriers.json", barriers)
    write_json(OUTPUT_DIR / "transition.json", transition)
    write_json(OUTPUT_DIR / "extraction-summary.json", summary)

    print(json.dumps(summary, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
