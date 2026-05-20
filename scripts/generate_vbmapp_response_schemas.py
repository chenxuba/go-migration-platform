#!/usr/bin/env python3
"""Generate VB-MAPP smart assessment evidence schemas.

The scoring tables tell us what score can be awarded. They do not tell the
Pad app what a teacher must record while assessing. This script adds that
missing product layer: one schema per VB-MAPP item, with evidence fields,
material strategy, related skills, and auto-completion rules.
"""

from __future__ import annotations

import json
import re
from collections import Counter
from pathlib import Path
from typing import Any


REPO_ROOT = Path(__file__).resolve().parents[1]
DATA_DIR = REPO_ROOT / "docs" / "vbmapp"

MILESTONE_ITEMS = DATA_DIR / "milestone-items.json"
SCORING_RULES = DATA_DIR / "milestone-scoring-rules.json"
BARRIERS = DATA_DIR / "barriers.json"
TRANSITIONS = DATA_DIR / "transition.json"

FIELD_TEMPLATES_OUT = DATA_DIR / "response-field-templates.json"
MATERIAL_PROFILES_OUT = DATA_DIR / "response-material-profiles.json"
MILESTONE_SCHEMAS_OUT = DATA_DIR / "milestone-response-schemas.json"
BARRIER_SCHEMAS_OUT = DATA_DIR / "barrier-response-schemas.json"
TRANSITION_SCHEMAS_OUT = DATA_DIR / "transition-response-schemas.json"
SUMMARY_OUT = DATA_DIR / "response-schema-summary.json"


FIELD_TEMPLATES: dict[str, dict[str, Any]] = {
    "score": {
        "label": "本题评分",
        "control": "score_buttons",
        "required": True,
        "recordPurpose": "最终分数仍由老师确认，系统只根据证据给出建议。",
    },
    "materials": {
        "label": "本次使用素材/情境",
        "control": "material_picker",
        "required": True,
        "recordPurpose": "记录实物、图卡、绘本、视频、活动或自然情境，方便复核和下次复测。",
        "columns": ["material_id", "material_name", "material_type", "prepared", "used"],
    },
    "stimulus_set": {
        "label": "刺激组合",
        "control": "stimulus_set_builder",
        "required": False,
        "recordPurpose": "记录目标和干扰项数量，避免凭印象判断是否满足组合大小要求。",
        "columns": ["set_no", "target", "distractors", "set_size", "presentation"],
    },
    "timer": {
        "label": "计时观察",
        "control": "timer_with_event_counter",
        "required": True,
        "recordPurpose": "对TO或含持续时间要求的题记录观察时长、事件数和有效事件数。",
        "columns": ["start_time", "planned_minutes", "actual_minutes", "event_count", "qualified_event_count"],
    },
    "prompt_level": {
        "label": "辅助水平",
        "control": "prompt_selector",
        "required": True,
        "recordPurpose": "区分无辅助、口头辅助、手势辅助、示范/仿说辅助和肢体辅助。",
        "options": ["independent", "allowed_verbal_cue", "gesture", "model_or_echoic", "partial_physical", "full_physical"],
    },
    "quality_checks": {
        "label": "达标核查",
        "control": "checklist",
        "required": True,
        "recordPurpose": "把题干中的关键限制条件变成可勾选项，例如无肢体辅助、目标物是否在眼前、是否自发。",
    },
    "teacher_note": {
        "label": "教师备注",
        "control": "multiline_text",
        "required": False,
        "recordPurpose": "记录无法结构化的临床判断、异常状态或家长补充。",
    },
    "mand_event_log": {
        "label": "提要求事件记录",
        "control": "event_table",
        "required": True,
        "recordPurpose": "保留孩子实际提出的词语/手语/图片交换、动机情境和辅助水平。",
        "columns": [
            "time",
            "motivation_or_context",
            "requested_item_or_action",
            "child_exact_response",
            "response_mode",
            "communication_partner",
            "environment",
            "prompt_level",
            "qualified",
        ],
    },
    "generalization_matrix": {
        "label": "泛化矩阵",
        "control": "matrix_table",
        "required": True,
        "recordPurpose": "记录目标是否跨人物、环境、样本或形式泛化。",
        "columns": ["target", "person", "environment", "example", "child_response", "qualified"],
    },
    "trial_log": {
        "label": "试次记录",
        "control": "trial_table",
        "required": True,
        "recordPurpose": "记录每个测试试次的指令、目标、孩子反应、正确性和辅助。",
        "columns": [
            "trial_no",
            "instruction_or_stimulus",
            "target",
            "stimulus_set_size",
            "child_response",
            "correct",
            "prompt_level",
        ],
    },
    "language_sample_log": {
        "label": "语言样本",
        "control": "language_sample_table",
        "required": True,
        "recordPurpose": "记录孩子原话、语言成分、是否自发、是否功能性使用。",
        "columns": ["time", "context", "child_exact_utterance", "language_parts", "spontaneous", "qualified"],
    },
    "vocabulary_inventory": {
        "label": "词汇/技能清单",
        "control": "inventory_counter",
        "required": True,
        "recordPurpose": "用于大量词汇、已知技能或积累清单类题目，支持批量导入和抽样复核。",
        "columns": ["category", "known_count", "sample_items", "source", "last_verified_at"],
    },
    "play_event_log": {
        "label": "游戏/活动观察",
        "control": "event_table",
        "required": True,
        "recordPurpose": "记录游戏材料、动作、独立持续时间、变化性和成人介入。",
        "columns": ["time", "activity_or_object", "action", "duration_seconds", "independent", "qualified"],
    },
    "peer_interaction_log": {
        "label": "同伴互动记录",
        "control": "event_table",
        "required": True,
        "recordPurpose": "记录同伴、发起/回应类型、成人辅助和互动是否达标。",
        "columns": ["time", "peer", "interaction_type", "child_response", "adult_prompt", "duration_seconds", "qualified"],
    },
    "imitation_trial_log": {
        "label": "模仿试次记录",
        "control": "trial_table",
        "required": True,
        "recordPurpose": "记录示范动作、是否持物、动作步骤、孩子模仿质量和辅助。",
        "columns": ["trial_no", "modeled_action", "object_used", "child_imitation", "correct", "prompt_level"],
    },
    "eesa_score_sheet": {
        "label": "EESA分测试",
        "control": "standard_subtest_score_sheet",
        "required": True,
        "recordPurpose": "仿说题按EESA得分映射里程碑分，需记录目标音/词和得分。",
        "columns": ["group", "target", "best_response_score", "phoneme_note"],
    },
    "vocalization_log": {
        "label": "自发语音记录",
        "control": "event_counter",
        "required": True,
        "recordPurpose": "记录声音或词语样本、近似程度、语调和是否自发。",
        "columns": ["time", "sound_or_word", "approximation", "intonation", "spontaneous", "qualified"],
    },
    "lrffc_trial_log": {
        "label": "LRFFC试次记录",
        "control": "trial_table",
        "required": True,
        "recordPurpose": "记录功能、特性、类别等语言条件，确认组合中目标唯一。",
        "columns": [
            "trial_no",
            "question_or_instruction",
            "target",
            "condition_parts",
            "stimulus_set_size",
            "unique_target_confirmed",
            "child_selection",
            "child_spontaneous_tact",
            "correct",
            "prompt_level",
        ],
    },
    "intraverbal_response_log": {
        "label": "对话反应记录",
        "control": "response_table",
        "required": True,
        "recordPurpose": "保留问题/话题和孩子原话，用于区分真正对话、背诵、仿说和猜测。",
        "columns": ["trial_no", "question_or_topic", "child_exact_response", "response_type", "correct", "prompt_level"],
    },
    "group_routine_log": {
        "label": "集体/常规观察",
        "control": "routine_observation_table",
        "required": True,
        "recordPurpose": "记录集体人数、活动、持续时间、注意比例、指令反应和破坏行为。",
        "columns": [
            "activity",
            "peer_count",
            "duration_minutes",
            "instruction_or_routine",
            "child_response",
            "attention_percent",
            "prompt_level",
            "disruption",
        ],
    },
    "reading_trial_log": {
        "label": "阅读试次/注意记录",
        "control": "trial_table",
        "required": True,
        "recordPurpose": "记录书本注意、字母/词语试次、孩子选择或朗读结果。",
        "columns": ["trial_no", "letter_word_or_book", "instruction", "child_response", "correct", "duration_seconds"],
    },
    "writing_artifact": {
        "label": "书写作品",
        "control": "artifact_upload_with_trial_table",
        "required": True,
        "recordPurpose": "保留书写作品或照片，并记录示范目标、描摹/仿写/独立完成情况。",
        "columns": ["trial_no", "model_or_target", "artifact", "legible_or_within_tolerance", "prompt_level"],
    },
    "math_trial_log": {
        "label": "数学试次记录",
        "control": "trial_table",
        "required": True,
        "recordPurpose": "记录数字、数量、比较、点数或配对试次和孩子反应。",
        "columns": ["trial_no", "instruction", "target_number_or_quantity", "child_response", "correct", "prompt_level"],
    },
    "learning_history_log": {
        "label": "学习历史/未训练确认",
        "control": "history_checklist",
        "required": False,
        "recordPurpose": "用于新反应、未训练泛化或维持类题目，记录是否经过直接训练和最近教学情况。",
        "columns": ["target", "trained_before", "last_taught_at", "source", "note"],
    },
    "barrier_behavior_log": {
        "label": "障碍行为证据",
        "control": "behavior_rubric_evidence",
        "required": True,
        "recordPurpose": "障碍评分必须记录频率、强度、情境和例子，避免只凭主观印象打分。",
        "columns": ["setting", "example", "frequency", "intensity", "duration", "impact", "current_strategy"],
    },
    "transition_evidence_summary": {
        "label": "转衔证据汇总",
        "control": "computed_summary_with_override",
        "required": True,
        "recordPurpose": "转衔评分来自里程碑、障碍、学习速度、泛化、自发性和集体适应等证据汇总。",
        "columns": ["basis_type", "computed_value", "suggested_score", "teacher_override", "override_reason"],
    },
}


MATERIAL_PROFILES: dict[str, dict[str, Any]] = {
    "potential_reinforcer_set": {
        "label": "潜在强化物/活动",
        "sourceLogic": "提要求评估优先准备孩子想要的实物或活动，让孩子看得到但不能直接取得。",
        "suggestedTypes": ["实物玩具", "食物/饮料", "活动", "社交游戏", "缺失物品"],
        "recommendedMaterials": [
            {"id": "mand_food_cookie", "name": "饼干", "type": "食物/饮料"},
            {"id": "mand_book", "name": "书", "type": "实物/活动"},
            {"id": "mand_ball", "name": "球", "type": "实物玩具"},
            {"id": "mand_bubbles", "name": "泡泡", "type": "社交游戏"},
            {"id": "mand_music", "name": "音乐", "type": "活动"},
            {"id": "mand_car", "name": "车", "type": "实物玩具"},
            {"id": "mand_swing", "name": "秋千", "type": "活动"},
            {"id": "mand_blocks", "name": "积木", "type": "实物玩具"},
            {"id": "mand_slinky", "name": "彩虹弹簧", "type": "实物玩具"},
            {"id": "mand_open", "name": "打开", "type": "动作/帮助"},
        ],
        "preparationChecks": ["确认动机存在", "避免孩子直接拿到", "准备多个年龄段偏好材料", "记录是否允许目标物在眼前"],
    },
    "tact_object_picture_book_set": {
        "label": "命名实物/图卡/绘本素材",
        "sourceLogic": "命名可使用图卡，但不能只依赖图卡，应至少部分使用实物、绘本或自然情境。",
        "suggestedTypes": ["实物", "图卡", "绘本", "视频画面", "自然环境事件"],
        "preparationChecks": ["标记2D/3D", "标记是否为强化物", "准备同一物品多个样本", "避免只抽测单一材料形式"],
    },
    "listener_shared_material_set": {
        "label": "听者反应共享素材",
        "sourceLogic": "听者反应可复用命名材料，但要记录刺激组合大小和呈现方式。",
        "suggestedTypes": ["实物", "图卡", "身体部位", "动作指令", "场景图", "自然环境物品"],
        "preparationChecks": ["设置目标和干扰项", "记录组合大小", "确认无额外视觉辅助", "记录泛化样本"],
    },
    "matching_visual_set": {
        "label": "视觉配对/样本配对素材",
        "sourceLogic": "视觉配对需要拼图、实物、图片或3D-2D配对材料，并记录目标与干扰项。",
        "suggestedTypes": ["嵌入式拼图", "相同实物", "相同图片", "不同大小/背景图片", "3D-2D配对"],
        "preparationChecks": ["记录样本类型", "记录干扰项", "记录是否相同/相似/同类", "保留错误选择"],
    },
    "natural_play_context": {
        "label": "自然游戏/独立活动情境",
        "sourceLogic": "游戏类题依赖自然观察和持续时间，素材应支持孩子独立操作。",
        "suggestedTypes": ["玩具套组", "绘本", "积木", "假想游戏材料", "美工材料"],
        "preparationChecks": ["启动计时", "记录成人介入", "记录玩法变化", "记录持续时间"],
    },
    "peer_social_context": {
        "label": "同伴互动情境",
        "sourceLogic": "社会游戏题需要同伴、活动、成人辅助和互动质量的证据。",
        "suggestedTypes": ["同伴游戏", "轮流活动", "身体互动游戏", "集体音乐/运动", "共享玩具"],
        "preparationChecks": ["记录同伴人数", "记录发起/回应", "记录成人辅助", "记录互动是否自发"],
    },
    "imitation_action_set": {
        "label": "动作模仿清单",
        "sourceLogic": "动作模仿题使用动作列表、持物动作和多步骤动作，并记录示范和孩子模仿。",
        "suggestedTypes": ["粗大动作", "精细动作", "持物动作", "面部动作", "两步动作"],
        "preparationChecks": ["记录示范动作", "记录是否持物", "避免无意识辅助", "记录泛化人物"],
    },
    "eesa_form": {
        "label": "EESA评估表",
        "sourceLogic": "仿说项目根据EESA分测试得分映射里程碑分。",
        "suggestedTypes": ["EESA表", "目标音节/词表", "录音"],
        "preparationChecks": ["记录分组得分", "保留典型发音样本", "记录音素错误"],
    },
    "lrffc_shared_set": {
        "label": "LRFFC功能/特性/类别素材",
        "sourceLogic": "LRFFC可复用命名/听者材料；涉及食物时不建议只用真实食物，以免选择偏差。",
        "suggestedTypes": ["实物", "图卡", "绘本场景", "视频画面", "功能/类别题库"],
        "preparationChecks": ["确认目标唯一", "记录语言条件成分", "避免食物偏好干扰", "记录孩子是否自发命名"],
    },
    "conversation_topic_bank": {
        "label": "对话/填空/主题题库",
        "sourceLogic": "对话题需要记录题目、话题、孩子原话和是否属于真正对话反应。",
        "suggestedTypes": ["歌曲/社交游戏填空", "WH问题", "故事短文", "视频/事件", "主题追问"],
        "preparationChecks": ["避免刚刚仿说", "记录原话", "记录问题顺序", "记录是否独立回答"],
    },
    "classroom_group_context": {
        "label": "教室常规/集体场景",
        "sourceLogic": "教室常规题依赖真实集体环境、持续时间、注意比例和辅助水平。",
        "suggestedTypes": ["点心/午餐", "排队", "圆圈时间", "集体课", "独立工作"],
        "preparationChecks": ["记录小组人数", "记录持续时间", "记录注意比例", "记录破坏行为"],
    },
    "language_sample_context": {
        "label": "自然语言样本/词汇清单",
        "sourceLogic": "语言结构题关注词汇量、词组长度、词形变化和句法，应保留自然样本或清单。",
        "suggestedTypes": ["自然对话", "提要求样本", "命名样本", "家长/教学清单", "短语样本"],
        "preparationChecks": ["标记是否仿说", "记录功能", "统计词汇/短语", "保留代表性原话"],
    },
    "reading_material_set": {
        "label": "阅读材料",
        "sourceLogic": "阅读题从书本注意到字母、名字、词图配对，需要记录材料和试次。",
        "suggestedTypes": ["绘本", "大写字母", "名字卡", "文字卡", "图文配对卡"],
        "preparationChecks": ["记录阅读材料", "记录注意时长", "记录目标字母/词", "记录正确性"],
    },
    "writing_artifact_set": {
        "label": "书写材料和作品",
        "sourceLogic": "书写题必须保留作品照片或扫描件，才能复核可辨读程度和距离标准。",
        "suggestedTypes": ["纸", "白板", "蜡笔/铅笔", "描摹模板", "字母/数字样本"],
        "preparationChecks": ["拍照留档", "记录示范/描摹/仿写", "记录可辨读程度", "记录容差标准"],
    },
    "math_material_set": {
        "label": "数学/数量材料",
        "sourceLogic": "数学题需要数字、数量、比较、点数或配对材料，并记录孩子选择或表达。",
        "suggestedTypes": ["数字卡", "实物数量", "比较物", "图片数量", "动作计数"],
        "preparationChecks": ["记录目标数", "记录指令", "记录数量材料", "记录正确性"],
    },
}


DOMAIN_DESIGNS: dict[str, dict[str, str]] = {
    "MAND": {
        "uiPattern": "mand_event_recorder",
        "materialProfile": "potential_reinforcer_set",
        "recordType": "mand_event_log",
        "why": "提要求题必须保留孩子实际提出的内容、沟通形式、动机情境和辅助水平，否则无法判断是否为功能性要求。",
    },
    "TACT": {
        "uiPattern": "tact_trial_recorder",
        "materialProfile": "tact_object_picture_book_set",
        "recordType": "tact_trial_log",
        "why": "命名题需要记录刺激物、孩子实际命名、是否自发或有辅助，以及是否泛化到不同样本。",
    },
    "LR": {
        "uiPattern": "listener_response_recorder",
        "materialProfile": "listener_shared_material_set",
        "recordType": "listener_trial_log",
        "why": "听者反应题需要记录指令、刺激组合大小、目标、孩子选择或动作、正确性和辅助。",
    },
    "VP_MTS": {
        "uiPattern": "matching_trial_recorder",
        "materialProfile": "matching_visual_set",
        "recordType": "matching_trial_log",
        "why": "视觉配对题需要记录样本、选择组合、孩子配对结果和干扰项。",
    },
    "INDEPENDENT_PLAY": {
        "uiPattern": "play_observation_recorder",
        "materialProfile": "natural_play_context",
        "recordType": "play_observation_log",
        "why": "独立游戏题依赖自然观察，需要记录活动、材料、独立持续时间、变化性和成人介入。",
    },
    "SOCIAL": {
        "uiPattern": "peer_social_recorder",
        "materialProfile": "peer_social_context",
        "recordType": "peer_interaction_log",
        "why": "社会游戏题要记录同伴、发起/回应类型、成人辅助和持续时间，单纯分数无法解释社交质量。",
    },
    "MOTOR_IMITATION": {
        "uiPattern": "imitation_trial_recorder",
        "materialProfile": "imitation_action_set",
        "recordType": "imitation_trial_log",
        "why": "动作模仿题需要记录示范动作、是否持物、步骤数、孩子模仿质量和辅助。",
    },
    "ECHOIC": {
        "uiPattern": "eesa_subtest_recorder",
        "materialProfile": "eesa_form",
        "recordType": "eesa_score_sheet",
        "why": "仿说题来自EESA标准分测试，应录入分测试得分和目标发音样本，而不是只填里程碑分。",
    },
    "SPONT_VOCAL": {
        "uiPattern": "timed_vocal_observation_recorder",
        "materialProfile": "language_sample_context",
        "recordType": "vocalization_log",
        "why": "自发语音题需要在计时观察中记录声音/词语样本、次数、种类、语调和自发性。",
    },
    "LRFFC": {
        "uiPattern": "lrffc_trial_recorder",
        "materialProfile": "lrffc_shared_set",
        "recordType": "lrffc_trial_log",
        "why": "LRFFC题需要记录功能、特性、类别等语言条件，并确认刺激组中目标唯一。",
    },
    "INTRAVERBAL": {
        "uiPattern": "intraverbal_response_recorder",
        "materialProfile": "conversation_topic_bank",
        "recordType": "intraverbal_response_log",
        "why": "对话题需要保留问题/话题和孩子原话，才能区分真正对话、背诵、仿说和猜测。",
    },
    "GROUP": {
        "uiPattern": "group_routine_recorder",
        "materialProfile": "classroom_group_context",
        "recordType": "group_routine_log",
        "why": "教室常规题必须记录活动场景、持续时间、小组人数、注意比例、辅助和破坏行为。",
    },
    "LINGUISTIC_STRUCTURE": {
        "uiPattern": "language_sample_recorder",
        "materialProfile": "language_sample_context",
        "recordType": "language_sample_or_inventory",
        "why": "语言结构题要记录自然语言样本或词汇清单，重点是词汇量、词组长度、词形变化和句法结构。",
    },
    "READING": {
        "uiPattern": "reading_trial_recorder",
        "materialProfile": "reading_material_set",
        "recordType": "reading_trial_log",
        "why": "阅读题需要记录书本注意、字母/词语试次、孩子选择或朗读结果。",
    },
    "WRITING": {
        "uiPattern": "writing_artifact_recorder",
        "materialProfile": "writing_artifact_set",
        "recordType": "writing_artifact",
        "why": "书写题必须保留书写作品或照片，并记录示范目标、描摹/仿写/独立完成情况。",
    },
    "MATH": {
        "uiPattern": "math_trial_recorder",
        "materialProfile": "math_material_set",
        "recordType": "math_trial_log",
        "why": "算术题需要记录数字、数量、比较或配对试次和孩子反应。",
    },
}


def load_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def dump_json(path: Path, value: Any) -> None:
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def dedupe(values: list[str]) -> list[str]:
    out: list[str] = []
    for value in values:
        if value and value not in out:
            out.append(value)
    return out


def has_timed_observation(item: dict[str, Any]) -> bool:
    text = f"{item.get('assessmentMode', '')} {item.get('title', '')}"
    return "TO" in text.upper() or "分钟" in text or "秒" in text


def is_inventory_item(title: str) -> bool:
    return any(key in title for key in ["积累", "累计", "清单", "词汇量", "1000", "1200", "300个", "250件", "200个"])


def criterion_metric(rule: dict[str, Any], item: dict[str, Any]) -> dict[str, Any]:
    one = rule.get("onePointCriteria", "")
    half = rule.get("halfPointCriteria", "")
    title = item.get("title", "")
    metric = "qualified_count"
    if "%" in one or "%" in half or "%" in title:
        metric = "percent_correct_or_attention"
    elif "+" in one or "*" in one:
        metric = "component_count"
    elif "分钟" in one or "秒" in one:
        metric = "duration"
    elif "EESA" in title or item.get("domainCode") == "ECHOIC":
        metric = "eesa_score"
    elif is_inventory_item(title):
        metric = "inventory_count"
    return {
        "primaryMetric": metric,
        "onePointCriteria": one,
        "halfPointCriteria": half,
        "rawCriteria": {
            "description": rule.get("description", ""),
            "onePoint": one,
            "halfPoint": half,
        },
    }


def base_fields_for(item: dict[str, Any]) -> list[str]:
    domain = item["domainCode"]
    no = int(item["milestoneNo"])
    title = item["title"]
    timed = has_timed_observation(item)
    fields = ["score", "materials", "prompt_level", "quality_checks", "teacher_note"]

    if domain == "MAND":
        fields = ["score", "materials", "mand_event_log", "prompt_level", "quality_checks", "teacher_note"]
        if no == 3:
            fields.insert(3, "generalization_matrix")
        if timed:
            fields.insert(2, "timer")
        if no in {5, 6, 10, 14, 15}:
            fields.insert(-1, "learning_history_log")
    elif domain == "TACT":
        fields = ["score", "materials", "stimulus_set", "trial_log", "prompt_level", "quality_checks", "teacher_note"]
        if timed or "自发" in title:
            fields.insert(2, "timer")
            fields.insert(3, "language_sample_log")
        if is_inventory_item(title) or no in {7, 10, 15}:
            fields.insert(2, "vocabulary_inventory")
        if no in {7, 12, 13}:
            fields.insert(3, "generalization_matrix")
    elif domain == "LR":
        fields = ["score", "materials", "stimulus_set", "trial_log", "prompt_level", "quality_checks", "teacher_note"]
        if timed:
            fields.insert(2, "timer")
        if is_inventory_item(title) or no in {10, 15}:
            fields.insert(2, "vocabulary_inventory")
        if no in {7, 12, 13}:
            fields.insert(3, "generalization_matrix")
    elif domain == "VP_MTS":
        fields = ["score", "materials", "stimulus_set", "trial_log", "prompt_level", "quality_checks", "teacher_note"]
        if timed or item.get("assessmentMode") == "O":
            fields.insert(2, "timer")
    elif domain == "INDEPENDENT_PLAY":
        fields = ["score", "materials", "timer", "play_event_log", "prompt_level", "quality_checks", "teacher_note"]
    elif domain == "SOCIAL":
        fields = ["score", "materials", "timer", "peer_interaction_log", "prompt_level", "quality_checks", "teacher_note"]
    elif domain == "MOTOR_IMITATION":
        fields = ["score", "materials", "imitation_trial_log", "prompt_level", "quality_checks", "teacher_note"]
        if timed or item.get("assessmentMode") == "O":
            fields.insert(2, "timer")
        if no in {3, 5}:
            fields.insert(3, "generalization_matrix")
    elif domain == "ECHOIC":
        fields = ["score", "eesa_score_sheet", "quality_checks", "teacher_note"]
    elif domain == "SPONT_VOCAL":
        fields = ["score", "timer", "vocalization_log", "quality_checks", "teacher_note"]
    elif domain == "LRFFC":
        fields = ["score", "materials", "stimulus_set", "lrffc_trial_log", "prompt_level", "quality_checks", "teacher_note"]
        if no == 10:
            fields.insert(3, "language_sample_log")
        if is_inventory_item(title) or no == 15:
            fields.insert(2, "vocabulary_inventory")
    elif domain == "INTRAVERBAL":
        fields = ["score", "materials", "intraverbal_response_log", "prompt_level", "quality_checks", "teacher_note"]
        if item.get("assessmentMode") == "O" or "自发" in title:
            fields.insert(2, "timer")
            fields.insert(3, "language_sample_log")
        if is_inventory_item(title) or no == 12:
            fields.insert(2, "vocabulary_inventory")
    elif domain == "GROUP":
        fields = ["score", "materials", "group_routine_log", "prompt_level", "quality_checks", "teacher_note"]
        if timed or "分钟" in title:
            fields.insert(2, "timer")
        if no == 14:
            fields.insert(-1, "learning_history_log")
    elif domain == "LINGUISTIC_STRUCTURE":
        fields = ["score", "language_sample_log", "vocabulary_inventory", "quality_checks", "teacher_note"]
    elif domain == "READING":
        fields = ["score", "materials", "reading_trial_log", "prompt_level", "quality_checks", "teacher_note"]
        if timed:
            fields.insert(2, "timer")
    elif domain == "WRITING":
        fields = ["score", "materials", "writing_artifact", "prompt_level", "quality_checks", "teacher_note"]
    elif domain == "MATH":
        fields = ["score", "materials", "math_trial_log", "prompt_level", "quality_checks", "teacher_note"]
    return dedupe(fields)


def record_depth_for(item: dict[str, Any], fields: list[str]) -> str:
    if "writing_artifact" in fields:
        return "artifact_required"
    if "vocabulary_inventory" in fields and "trial_log" not in fields and "language_sample_log" not in fields:
        return "inventory_required"
    if "timer" in fields:
        return "timed_observation_required"
    if item.get("assessmentMode") == "O":
        return "observation_log_required"
    return "trial_or_event_log_required"


def evidence_targets(item: dict[str, Any]) -> list[str]:
    domain = item["domainCode"]
    no = int(item["milestoneNo"])
    targets: list[str]
    if domain == "MAND":
        targets = ["孩子实际发出的要求", "沟通形式", "动机情境", "辅助水平", "是否功能性获得目标"]
        if no == 3:
            targets += ["人物泛化", "环境泛化", "样本泛化"]
        if no in {8, 11, 13, 14, 15}:
            targets += ["完整原话", "语言成分"]
    elif domain == "TACT":
        targets = ["刺激物", "孩子实际命名", "正确性", "辅助水平", "材料形式"]
        if no in {7, 10, 15}:
            targets += ["已知命名清单", "抽样复核结果"]
    elif domain == "LR":
        targets = ["口头指令", "目标/干扰项", "孩子选择或动作", "正确性", "辅助水平"]
    elif domain == "LRFFC":
        targets = ["语言条件", "目标唯一性", "目标/干扰项", "孩子选择", "是否自发命名目标"]
    elif domain == "INTRAVERBAL":
        targets = ["问题或话题", "孩子原话", "反应类型", "是否独立回答", "是否按顺序回答"]
    elif domain == "GROUP":
        targets = ["集体人数", "活动场景", "持续时间", "注意比例", "指令反应", "破坏行为"]
    elif domain == "LINGUISTIC_STRUCTURE":
        targets = ["自然语言样本", "词汇/短语数量", "语言成分", "是否仿说", "是否功能性使用"]
    else:
        targets = ["测试材料", "孩子反应", "正确性或持续时间", "辅助水平", "是否达标"]
    return dedupe(targets)


def quality_checks_for(item: dict[str, Any]) -> list[str]:
    title = item["title"]
    mode = item.get("assessmentMode", "")
    checks: list[str] = []
    if "无辅助" in title or "没有辅助" in title:
        checks.append("确认未使用题干不允许的辅助")
    if "自发" in title:
        checks.append("确认不是口头提示后的反应")
    if "不需要肢体辅助" in title:
        checks.append("确认没有肢体辅助")
    if "物件可在眼前" in title or "所要的物件可在眼前" in title:
        checks.append("记录目标物是否在眼前")
    if "泛化" in title:
        checks.append("确认泛化维度满足题干要求")
    if "新" in title or "没有经过特别的训练" in title:
        checks.append("确认不是刚训练出的反应")
    if "TO" in mode or "分钟" in title:
        checks.append("确认计时观察时长足够")
    if not checks:
        checks.append("确认记录的证据可支持本题评分")
    return checks


def auto_completion(item: dict[str, Any], fields: list[str], metric: dict[str, Any]) -> dict[str, Any]:
    domain = item["domainCode"]
    no = int(item["milestoneNo"])
    strategy = "teacher_confirmation_with_structured_evidence"
    computed = ["suggest_score_from_recorded_evidence"]
    if domain == "MAND":
        strategy = "count_qualified_unique_mand_events"
        computed = ["unique_mand_count", "prompt_filtered_count"]
        if no == 3:
            computed += ["person_generalization_count", "environment_generalization_count", "exemplar_generalization_count"]
    elif domain in {"TACT", "LR", "VP_MTS", "MOTOR_IMITATION", "READING", "MATH"}:
        strategy = "count_correct_trials"
        computed = ["correct_trial_count", "trial_total", "prompt_filtered_count"]
    elif domain == "ECHOIC":
        strategy = "map_eesa_total_to_milestone_score"
        computed = ["eesa_total", "eesa_group_score"]
    elif domain == "SPONT_VOCAL":
        strategy = "count_spontaneous_vocalizations"
        computed = ["qualified_vocal_count", "different_sound_count"]
    elif domain == "LRFFC":
        strategy = "count_correct_lrffc_trials"
        computed = ["correct_trial_count", "condition_part_count", "unique_target_count"]
        if no == 10:
            computed.append("spontaneous_tact_percent")
    elif domain == "INTRAVERBAL":
        strategy = "count_independent_intraverbal_responses"
        computed = ["correct_response_count", "topic_count", "question_sequence_count"]
    elif domain == "GROUP":
        strategy = "evaluate_group_routine_observation"
        computed = ["duration_minutes", "attention_percent", "instruction_response_count", "disruption_count"]
    elif domain == "LINGUISTIC_STRUCTURE":
        strategy = "count_language_sample_or_inventory"
        computed = ["inventory_count", "qualified_utterance_count", "language_part_count"]
    elif domain in {"INDEPENDENT_PLAY", "SOCIAL"}:
        strategy = "count_qualified_observation_events"
        computed = ["duration_minutes", "qualified_event_count", "independence_or_peer_prompt_level"]
    if "vocabulary_inventory" in fields:
        computed.append("inventory_known_count")
    if "timer" in fields:
        computed.append("actual_observation_minutes")
    return {
        "canAutoCompleteDraft": True,
        "canSuggestScore": True,
        "scoreStrategy": strategy,
        "primaryMetric": metric["primaryMetric"],
        "computedIndicators": dedupe(computed),
        "requiresTeacherConfirmation": True,
        "teacherCanOverride": True,
    }


def related_items_for(item: dict[str, Any], item_ids: set[str]) -> list[dict[str, str]]:
    domain = item["domainCode"]
    no = int(item["milestoneNo"])
    relations: list[dict[str, str]] = []

    def add(relation: str, milestone_id: str, reason: str) -> None:
        if milestone_id in item_ids and milestone_id != item["milestoneId"]:
            relations.append({"relation": relation, "milestoneId": milestone_id, "reason": reason})

    if domain == "TACT":
        add("shared_materials", f"LR_{no:02d}M", "命名材料常可复用于听者反应，但需改变任务要求。")
        add("shared_materials", f"VP_MTS_{no:02d}M", "同一实物/图卡可用于视觉配对和样本配对。")
    elif domain == "LR":
        add("shared_materials", f"TACT_{no:02d}M", "听者反应材料常来自命名素材库。")
        add("shared_materials", f"VP_MTS_{no:02d}M", "听者反应中的目标/干扰项可复用配对材料。")
    elif domain == "VP_MTS":
        add("shared_materials", f"TACT_{no:02d}M", "配对材料可继续作为命名刺激物。")
        add("shared_materials", f"LR_{no:02d}M", "配对材料可继续作为听者反应刺激组合。")
    elif domain == "MAND":
        add("language_transfer", f"TACT_{no:02d}M", "已会命名的目标可能转化为新的提要求反应。")
        add("listener_context", f"LR_{no:02d}M", "听者理解和动机情境会影响提要求表现。")
    elif domain == "LRFFC":
        add("prerequisite", f"TACT_{no:02d}M", "LRFFC目标常需要孩子已有相关命名或物品经验。")
        add("prerequisite", f"LR_{no:02d}M", "LRFFC建立在听者辨别和刺激组合能力上。")
        add("transfer_check", f"INTRAVERBAL_{no:02d}M", "同一功能/类别关系可迁移到对话反应。")
    elif domain == "INTRAVERBAL":
        add("language_sample_source", f"MAND_{no:02d}M", "部分自发评论可能同时表现为提要求或对话反应。")
        add("language_sample_source", f"TACT_{no:02d}M", "命名词汇量影响对话回答。")
    elif domain == "GROUP":
        add("contextual_support", f"SOCIAL_{no:02d}M", "集体参与常受同伴互动能力影响。")
        add("contextual_support", f"INDEPENDENT_PLAY_{no:02d}M", "独立工作和游戏持续性会影响集体任务表现。")
    elif domain == "LINGUISTIC_STRUCTURE":
        add("sample_source", f"MAND_{no:02d}M", "提要求原话可纳入语言结构样本。")
        add("sample_source", f"TACT_{no:02d}M", "命名原话可纳入语言结构样本。")
        add("sample_source", f"INTRAVERBAL_{no:02d}M", "对话原话可纳入语言结构样本。")
    return relations


def item_specific_design(item: dict[str, Any]) -> dict[str, Any]:
    domain = item["domainCode"]
    no = int(item["milestoneNo"])
    title = item["title"]
    design: dict[str, Any] = {
        "recordingInstructions": [],
        "uiEmphasis": [],
    }
    if domain == "MAND" and no == 1:
        design["recordingInstructions"] = [
            "逐条记录孩子实际发出的词语、手语或图片交换内容。",
            "记录要求的目标物/活动和当时动机，确认不是单纯仿说。",
            "标记是否使用仿说、模仿或其他辅助；若使用肢体辅助则不得计为达标事件。",
        ]
        design["uiEmphasis"] = ["快速新增一次要求", "原话/手势/图片必填", "自动去重统计2个有效要求"]
    elif domain == "MAND" and no == 3:
        design["recordingInstructions"] = [
            "选择一个强化物，分别记录人物、环境、样本三个维度的泛化。",
            "系统按2个人、2个环境、2个例子自动生成矩阵格。",
        ]
        design["uiEmphasis"] = ["2+2+2泛化矩阵", "未覆盖维度高亮"]
    elif domain == "MAND" and no in {4, 8, 9, 11, 13}:
        design["recordingInstructions"] = ["开启计时观察，记录每次自发要求的原话和语言成分。"]
        design["uiEmphasis"] = ["计时器固定可见", "自发事件一键记录", "自动统计不同要求"]
    elif domain == "TACT" and no in {7, 10, 15}:
        design["recordingInstructions"] = ["支持导入已知命名清单，并抽样记录本次复核试次。"]
        design["uiEmphasis"] = ["清单计数", "抽样试次", "实物/图卡/绘本来源标记"]
    elif domain == "LR" and no in {5, 6, 7, 11}:
        design["recordingInstructions"] = ["先搭建刺激组合，再按试次记录目标、干扰项和孩子选择。"]
        design["uiEmphasis"] = ["刺激组合生成器", "目标唯一性检查", "正确率统计"]
    elif domain == "LRFFC":
        design["recordingInstructions"] = ["记录功能/特性/类别语言条件，确认组合中只有一个正确目标。"]
        design["uiEmphasis"] = ["条件成分标签", "目标唯一性", "自发命名记录"]
    elif domain == "GROUP":
        design["recordingInstructions"] = ["记录小组人数、持续时间、注意比例、指令反应和破坏行为。"]
        design["uiEmphasis"] = ["课堂观察计时", "注意比例滑杆", "指令反应计数"]
    elif domain == "WRITING":
        design["recordingInstructions"] = ["每个书写试次上传作品照片，并记录示范/描摹/仿写/独立完成。"]
        design["uiEmphasis"] = ["作品拍照", "可辨读/容差判定"]
    elif is_inventory_item(title):
        design["recordingInstructions"] = ["使用清单或历史教学记录批量录入，再保留本次抽样复核证据。"]
        design["uiEmphasis"] = ["清单导入", "抽样复核", "来源标记"]
    return design


def schema_for_milestone(
    item: dict[str, Any],
    rule: dict[str, Any],
    item_ids: set[str],
) -> dict[str, Any]:
    domain_design = DOMAIN_DESIGNS[item["domainCode"]]
    fields = base_fields_for(item)
    metric = criterion_metric(rule, item)
    schema = {
        "moduleCode": "milestones",
        "milestoneId": item["milestoneId"],
        "label": item["label"],
        "domainCode": item["domainCode"],
        "domainName": item["domainName"],
        "level": item["level"],
        "ageBand": item["ageBand"],
        "milestoneNo": item["milestoneNo"],
        "assessmentMode": item.get("assessmentMode", ""),
        "recordDepth": record_depth_for(item, fields),
        "recordType": domain_design["recordType"],
        "uiPattern": domain_design["uiPattern"],
        "materialProfileId": domain_design["materialProfile"],
        "whyRecord": domain_design["why"],
        "evidenceTargets": evidence_targets(item),
        "fieldTemplateIds": fields,
        "qualityChecks": quality_checks_for(item),
        "scoreEvidence": {
            **metric,
            "requiresTeacherConfirmation": True,
        },
        "autoCompletion": auto_completion(item, fields, metric),
        "relatedItems": related_items_for(item, item_ids),
        "itemDesign": item_specific_design(item),
        "sourceFiles": dedupe([item.get("sourceFile", ""), rule.get("sourceFile", "")]),
    }
    return schema


def schema_for_barrier(item: dict[str, Any]) -> dict[str, Any]:
    return {
        "moduleCode": "barriers",
        "barrierCode": item["barrierCode"],
        "barrierName": item["barrierName"],
        "barrierNo": item["barrierNo"],
        "recordDepth": "rating_with_behavior_evidence_required",
        "uiPattern": "barrier_rubric_with_behavior_log",
        "fieldTemplateIds": ["score", "barrier_behavior_log", "teacher_note"],
        "evidenceTargets": ["具体行为例子", "发生频率", "严重程度", "持续时间", "影响范围", "当前应对策略"],
        "scoreEvidence": {
            "scoreRange": [item["minScore"], item["maxScore"]],
            "scoreOptions": item.get("scoreOptions", []),
            "requiresTeacherConfirmation": True,
        },
        "autoCompletion": {
            "canAutoCompleteDraft": True,
            "canSuggestScore": False,
            "scoreStrategy": "teacher_rates_barrier_from_structured_behavior_evidence",
            "computedIndicators": ["frequency_summary", "intensity_summary", "high_risk_flag"],
            "requiresTeacherConfirmation": True,
            "teacherCanOverride": True,
        },
        "sourceFiles": dedupe([item.get("sourceFile", "")]),
    }


def schema_for_transition(item: dict[str, Any]) -> dict[str, Any]:
    auto_codes = {"T01", "T02", "T03", "T04", "T05"}
    can_suggest = item["transitionCode"] in auto_codes
    return {
        "moduleCode": "transition",
        "transitionCode": item["transitionCode"],
        "transitionName": item["transitionName"],
        "transitionNo": item["transitionNo"],
        "category": item.get("category", ""),
        "recordDepth": "computed_summary_with_teacher_override",
        "uiPattern": "transition_rating_with_auto_basis",
        "fieldTemplateIds": ["score", "transition_evidence_summary", "teacher_note"],
        "evidenceTargets": ["系统计算依据", "老师补充证据", "安置建议", "历史变化", "人工覆盖原因"],
        "scoreEvidence": {
            "scoreRange": [item["minScore"], item["maxScore"]],
            "scoreOptions": item.get("scoreOptions", []),
            "placementRecommendations": item.get("placementRecommendations", []),
            "requiresTeacherConfirmation": True,
        },
        "autoCompletion": {
            "canAutoCompleteDraft": can_suggest,
            "canSuggestScore": can_suggest,
            "scoreStrategy": "derive_from_milestone_barrier_and_context_summary" if can_suggest else "teacher_rates_transition_with_evidence",
            "computedIndicators": ["milestone_total", "barrier_total", "domain_scores", "history_delta", "teacher_override_reason"],
            "requiresTeacherConfirmation": True,
            "teacherCanOverride": True,
        },
        "sourceFiles": dedupe([item.get("sourceFile", "")]),
    }


def main() -> None:
    milestones = load_json(MILESTONE_ITEMS)
    rules = {row["milestoneId"]: row for row in load_json(SCORING_RULES)}
    barriers = load_json(BARRIERS)
    transitions = load_json(TRANSITIONS)
    item_ids = {item["milestoneId"] for item in milestones}

    milestone_schemas = [
        schema_for_milestone(item, rules.get(item["milestoneId"], {}), item_ids)
        for item in milestones
    ]
    barrier_schemas = [schema_for_barrier(item) for item in barriers]
    transition_schemas = [schema_for_transition(item) for item in transitions]

    all_schemas = milestone_schemas + barrier_schemas + transition_schemas
    summary = {
        "schemaVersion": "VBMAPP_SMART_RESPONSE_SCHEMA_DRAFT_2026_05_20",
        "itemCount": len(all_schemas),
        "milestoneItemCount": len(milestone_schemas),
        "barrierItemCount": len(barrier_schemas),
        "transitionItemCount": len(transition_schemas),
        "recordDepthCounts": dict(Counter(row["recordDepth"] for row in all_schemas)),
        "uiPatternCounts": dict(Counter(row["uiPattern"] for row in all_schemas)),
        "fieldTemplateCount": len(FIELD_TEMPLATES),
        "materialProfileCount": len(MATERIAL_PROFILES),
        "sources": [
            "docs/vbmapp/milestone-items.json",
            "docs/vbmapp/milestone-scoring-rules.json",
            "docs/vbmapp/barriers.json",
            "docs/vbmapp/transition.json",
            "里程碑评估材料提示手册2.0.pdf",
            "各领域DOC任务分析表",
            "VB-MAPP上册指南与下册概况",
        ],
    }

    dump_json(FIELD_TEMPLATES_OUT, FIELD_TEMPLATES)
    dump_json(MATERIAL_PROFILES_OUT, MATERIAL_PROFILES)
    dump_json(MILESTONE_SCHEMAS_OUT, milestone_schemas)
    dump_json(BARRIER_SCHEMAS_OUT, barrier_schemas)
    dump_json(TRANSITION_SCHEMAS_OUT, transition_schemas)
    dump_json(SUMMARY_OUT, summary)

    print(f"wrote {MILESTONE_SCHEMAS_OUT} ({len(milestone_schemas)} items)")
    print(f"wrote {BARRIER_SCHEMAS_OUT} ({len(barrier_schemas)} items)")
    print(f"wrote {TRANSITION_SCHEMAS_OUT} ({len(transition_schemas)} items)")
    print(f"wrote {FIELD_TEMPLATES_OUT} ({len(FIELD_TEMPLATES)} templates)")
    print(f"wrote {MATERIAL_PROFILES_OUT} ({len(MATERIAL_PROFILES)} profiles)")
    print(f"wrote {SUMMARY_OUT}")


if __name__ == "__main__":
    main()
