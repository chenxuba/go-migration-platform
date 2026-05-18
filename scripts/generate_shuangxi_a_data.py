#!/usr/bin/env python3
"""Generate Shuangxi course evaluation A static scale data from extracted text."""

from __future__ import annotations

import argparse
import json
import re
from collections import OrderedDict
from pathlib import Path


SCALE_CODE = "SHUANGXI_A"
SCALE_NAME = "双溪课程评量表A"
SCALE_VERSION = "A-2012-doc"
DATA_STATUS = "draft"
REVISION = "2026-05-18"

DOMAIN_DEFS = OrderedDict(
    [
        (
            "1",
            {
                "scale_code": "SENSORY",
                "scale_name": "感官知觉",
                "sort_no": 1,
            },
        ),
        (
            "2",
            {
                "scale_code": "GROSS_MOTOR",
                "scale_name": "粗大动作",
                "sort_no": 2,
            },
        ),
        (
            "3",
            {
                "scale_code": "FINE_MOTOR",
                "scale_name": "精细动作",
                "sort_no": 3,
            },
        ),
        (
            "4",
            {
                "scale_code": "SELF_CARE",
                "scale_name": "生活自理",
                "sort_no": 4,
            },
        ),
        (
            "5",
            {
                "scale_code": "COMMUNICATION",
                "scale_name": "沟通",
                "sort_no": 5,
            },
        ),
        (
            "6",
            {
                "scale_code": "COGNITION",
                "scale_name": "认知",
                "sort_no": 6,
            },
        ),
        (
            "7",
            {
                "scale_code": "SOCIAL_SKILLS",
                "scale_name": "社会技能",
                "sort_no": 7,
            },
        ),
    ]
)

DOMAIN_TITLE_FALLBACK = {
    "1": "感官知觉",
    "2": "粗大动作",
    "3": "精细动作",
    "4": "生活自理",
    "5": "沟通",
    "6": "认知",
    "7": "社会技能",
}

KNOWN_ITEM_CODE_FIXES = {
    ("2.1.2", "头部控制"): "2.1.1",
}

SOURCES = [
    "双溪课程评量表.doc",
    "双溪心智障碍个别化教育课程.doc",
    "双溪18-61742597955.pdf",
    "revision:" + REVISION,
]

TEXT_FIXES = {
    "盲蚁": "盲或",
    "迫视": "追视",
    "仪能": "仅能",
    "四—卜分贝": "四十分贝",
    "山记忆": "出记忆",
    "站一去儿": "站一会儿",
    "才飞巨蹲着": "才能蹲着",
    "抓握拘体": "抓握物体",
    "各科大小": "各种大小",
    "正碗之用途": "正确之用途",
    "多砷文具": "多种文具",
    "配介将": "配合将",
    "在揭示下": "在提示下",
    "适心新环境": "适应新环境",
    "新外境": "新环境",
    "筒单": "简单",
    "其池": "其他",
    "慨念": "概念",
    "消洁": "清洁",
    "使刚洗衣机": "使用洗衣机",
    "公安单位": "公家单位",
    "暂乐": "音乐",
    "门己": "自己",
    "井拒绝": "并拒绝",
    "双、打": "双、单",
}


def clean_line(line: str) -> str:
    line = line.replace("\x07", " ")
    line = line.replace("\u3000", " ")
    line = line.replace("口口口口", " ")
    line = line.replace("．", ".")
    line = line.replace("—一", "——")
    for old, new in TEXT_FIXES.items():
        line = line.replace(old, new)
    line = re.sub(r"\s+", " ", line)
    return line.strip()


def clean_title(value: str) -> str:
    value = clean_line(value)
    value = re.sub(r"\s*口+\s*$", "", value)
    return value.strip()


def normalize_score_token(token: str) -> str:
    token = token.strip()
    if token in {"l", "L", "I", "丨"}:
        return "1"
    return token


def normalize_item_code(code: str, title: str) -> str:
    return KNOWN_ITEM_CODE_FIXES.get((code, title), code)


def is_domain_heading(domain_no: str, title: str) -> bool:
    expected = DOMAIN_TITLE_FALLBACK.get(domain_no, "")
    return expected != "" and expected in clean_title(title)


def split_code(code: str) -> tuple[str, str]:
    parts = code.split(".")
    if len(parts) < 2:
        return parts[0], ""
    return parts[0], ".".join(parts[:2])


def standard_text(score_options: list[dict[str, object]]) -> str:
    return "\n".join(
        f"{item['value']}-{item['description']}" for item in score_options
    )


def parse_items(text: str) -> tuple[list[dict[str, object]], OrderedDict[str, dict[str, object]]]:
    domain_head_re = re.compile(r"^([1-7])\s+(.+)$")
    skill_head_re = re.compile(r"^([1-7]\.\d+)\s+(.+)$")
    item_head_re = re.compile(r"^([1-7]\.\d+\.\d+)\s+(.+)$")
    score_re = re.compile(r"^([0-3lLI丨])\s+(.+)$")

    current_domain_no = ""
    current_domain_name = ""
    current_skill_code = ""
    current_skill_name = ""
    current_item: dict[str, object] | None = None
    current_score: dict[str, object] | None = None
    items: list[dict[str, object]] = []
    skills: OrderedDict[str, dict[str, object]] = OrderedDict()

    def flush_item() -> None:
        nonlocal current_item, current_score
        if current_item is None:
            return
        score_options = current_item.get("score_options", [])
        if isinstance(score_options, list):
            score_options.sort(key=lambda item: int(item["value"]))
            current_item["standard"] = standard_text(score_options)
            current_item["score_min"] = min(int(item["value"]) for item in score_options)
            current_item["score_max"] = max(int(item["value"]) for item in score_options)
        items.append(current_item)
        current_item = None
        current_score = None

    for raw_line in text.splitlines():
        line = clean_line(raw_line)
        if not line:
            continue
        if line.startswith("PAGE") or line.startswith("双溪个别化教育课程"):
            continue
        if line in {"评量表", "（二）评 量 表"}:
            continue

        item_match = item_head_re.match(line)
        if item_match:
            flush_item()
            raw_code = item_match.group(1)
            raw_title = clean_title(item_match.group(2))
            code = normalize_item_code(raw_code, raw_title)
            domain_no, skill_code = split_code(code)
            domain_def = DOMAIN_DEFS.get(domain_no)
            if domain_def is None:
                raise ValueError(f"unknown domain for item {code}")
            if current_skill_code != skill_code:
                current_skill_code = skill_code
                current_skill_name = skills.get(skill_code, {}).get("skill_name", "")
            current_domain_no = domain_no
            current_domain_name = domain_def["scale_name"]
            current_item = {
                "item_no": len(items) + 1,
                "item_code": code,
                "item_title": f"{code} {raw_title}",
                "test_item": raw_title,
                "domain_code": domain_def["scale_code"],
                "domain": current_domain_name,
                "domain_name": current_domain_name,
                "domain_sort_no": domain_def["sort_no"],
                "skill_code": current_skill_code,
                "skill_name": current_skill_name,
                "source_pdf": "双溪课程评量表.doc",
                "source_pages": [],
                "ocr_status": "doc-extracted-draft",
                "materials": "",
                "method": "依据0-3级评量标准观察、访谈或结合日常课程表现评分。",
                "describes": f"评量项目编码：{code}",
                "guidance": "",
                "guidance_video": "",
                "material_images": [],
                "score_options": [],
                "score_type": "0-3",
            }
            current_score = None
            continue

        skill_match = skill_head_re.match(line)
        if skill_match and not item_head_re.match(line):
            flush_item()
            current_skill_code = skill_match.group(1)
            current_skill_name = clean_title(skill_match.group(2))
            domain_no, _ = split_code(current_skill_code)
            domain_def = DOMAIN_DEFS.get(domain_no)
            if domain_def is None:
                continue
            current_domain_no = domain_no
            current_domain_name = domain_def["scale_name"]
            skills[current_skill_code] = {
                "skill_code": current_skill_code,
                "skill_name": current_skill_name,
                "domain_code": domain_def["scale_code"],
                "domain_name": current_domain_name,
                "sort_no": len(skills) + 1,
                "item_count": 0,
                "item_numbers": [],
            }
            continue

        domain_match = domain_head_re.match(line)
        if domain_match and is_domain_heading(domain_match.group(1), domain_match.group(2)):
            flush_item()
            current_domain_no = domain_match.group(1)
            current_domain_name = DOMAIN_TITLE_FALLBACK[current_domain_no]
            current_skill_code = ""
            current_skill_name = ""
            continue

        if current_item is not None:
            score_match = score_re.match(line)
            if score_match:
                value = int(normalize_score_token(score_match.group(1)))
                existing_values = {
                    int(option["value"])
                    for option in current_item["score_options"]  # type: ignore[index]
                }
                if value in existing_values:
                    missing_values = [candidate for candidate in range(4) if candidate not in existing_values]
                    if missing_values:
                        value = missing_values[0]
                description = clean_title(score_match.group(2))
                score_option = {
                    "value": value,
                    "label": f"{value}分",
                    "description": description,
                }
                current_item["score_options"].append(score_option)  # type: ignore[index, union-attr]
                current_score = score_option
                continue
            if current_score is not None:
                current_score["description"] = (
                    str(current_score.get("description", "")).rstrip() + " " + line
                ).strip()
                continue

        domain_match = domain_head_re.match(line)
        if domain_match:
            flush_item()
            domain_no = domain_match.group(1)
            if domain_no in DOMAIN_DEFS:
                current_domain_no = domain_no
                current_domain_name = DOMAIN_TITLE_FALLBACK[domain_no]
            continue

    flush_item()

    skill_sort_by_code = {code: idx + 1 for idx, code in enumerate(skills)}
    for item in items:
        skill_code = str(item["skill_code"])
        if not item["skill_name"] and skill_code in skills:
            item["skill_name"] = skills[skill_code]["skill_name"]
        item["skill_sort_no"] = skill_sort_by_code.get(skill_code, 0)
        if skill_code in skills:
            skills[skill_code]["item_count"] += 1
            skills[skill_code]["item_numbers"].append(item["item_no"])

    return items, skills


def build_domains(items: list[dict[str, object]], skills: OrderedDict[str, dict[str, object]]) -> list[dict[str, object]]:
    by_domain: OrderedDict[str, list[dict[str, object]]] = OrderedDict(
        (definition["scale_code"], []) for definition in DOMAIN_DEFS.values()
    )
    for item in items:
        by_domain[str(item["domain_code"])].append(item)

    domains: list[dict[str, object]] = []
    for domain_no, definition in DOMAIN_DEFS.items():
        domain_code = definition["scale_code"]
        domain_items = by_domain[domain_code]
        domain_skills = [
            skill
            for skill in skills.values()
            if skill["domain_code"] == domain_code
        ]
        item_numbers = [int(item["item_no"]) for item in domain_items]
        domains.append(
            {
                **definition,
                "category": "双溪课程评量表A",
                "domain_no": int(domain_no),
                "item_count": len(domain_items),
                "max_raw_score": len(domain_items) * 3,
                "item_numbers": item_numbers,
                "skills": domain_skills,
            }
        )
    return domains


def validate(items: list[dict[str, object]], domains: list[dict[str, object]]) -> None:
    if len(items) != 209:
        raise ValueError(f"expected 209 items, got {len(items)}")
    if len(domains) != 7:
        raise ValueError(f"expected 7 domains, got {len(domains)}")
    missing = []
    for item in items:
        values = sorted(int(option["value"]) for option in item["score_options"])  # type: ignore[index]
        if values != [0, 1, 2, 3]:
            missing.append((item["item_code"], values))
    if missing:
        details = ", ".join(f"{code}:{values}" for code, values in missing[:20])
        raise ValueError(f"items with incomplete score options: {details}")
    expected_counts = {
        "SENSORY": 21,
        "GROSS_MOTOR": 25,
        "FINE_MOTOR": 14,
        "SELF_CARE": 24,
        "COMMUNICATION": 56,
        "COGNITION": 21,
        "SOCIAL_SKILLS": 48,
    }
    for domain in domains:
        code = domain["scale_code"]
        if domain["item_count"] != expected_counts[code]:
            raise ValueError(f"domain {code} item_count={domain['item_count']}, want {expected_counts[code]}")


def write_json(path: Path, data: object) -> None:
    path.write_text(
        json.dumps(data, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("source_txt", type=Path)
    parser.add_argument("--out-dir", type=Path, default=Path("docs"))
    args = parser.parse_args()

    text = args.source_txt.read_text(encoding="utf-8", errors="ignore")
    items, skills = parse_items(text)
    domains = build_domains(items, skills)
    validate(items, domains)

    metadata = {
        "scale_code": SCALE_CODE,
        "scale_name": SCALE_NAME,
        "scale_version": SCALE_VERSION,
        "source_standard": "双溪心智障碍儿童个别化教育课程",
        "source_files": SOURCES[:-1],
        "item_count": len(items),
        "domain_count": len(domains),
        "skill_count": len(skills),
        "score_min": 0,
        "score_max": 3,
        "data_status": DATA_STATUS,
        "revision": REVISION,
        "scoring_note": "每题按0、1、2、3四级评分，领域原始分为该领域题目得分合计。",
    }

    args.out_dir.mkdir(parents=True, exist_ok=True)
    write_json(args.out_dir / "shuangxi-a-item-bank.json", items)
    write_json(args.out_dir / "shuangxi-a-domain-map.json", domains)
    write_json(args.out_dir / "shuangxi-a-scale-metadata.json", metadata)
    print(
        f"generated {len(items)} items, {len(domains)} domains, {len(skills)} skills into {args.out_dir}"
    )


if __name__ == "__main__":
    main()
