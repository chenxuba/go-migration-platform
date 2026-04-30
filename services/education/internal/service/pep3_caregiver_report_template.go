package service

import "go-migration-platform/services/education/internal/model"

func pep3CaregiverReportTemplate() model.PEP3CaregiverReportTemplate {
	return model.PEP3CaregiverReportTemplate{
		ReportName:   "PEP-3儿童照顾者报告",
		SourcePDF:    "照顾报告彩.pdf",
		SubmitMode:   "caregiver_self_report",
		Instructions: "由家长或主要照顾者根据儿童近期日常表现填写。每题选择最能代表儿童情况的一项；信息类和备注类不计分，问题行为、个人自理、适应行为按选项自动换算为原始分。",
		ScoreRules: []model.PEP3CaregiverScoreRule{
			{ScaleCode: "PB", ScaleName: "问题行为", SectionCode: "problem_behavior", MaxRawScore: 20, Description: "第1-10题计分：没有出现=2，轻微/中度=1，严重=0；第11题为文字备注，不计分。"},
			{ScaleCode: "PSC", ScaleName: "个人自理", SectionCode: "personal_self_care", MaxRawScore: 26, Description: "第1-13题计分：第一项=2，第二项=1，第三项=0；第14题为文字备注，不计分。"},
			{ScaleCode: "AB", ScaleName: "适应行为", SectionCode: "adaptive_behavior", MaxRawScore: 30, Description: "第1-15题计分：第一项=2，第二项=1，第三项=0；第16题为文字备注，不计分。"},
		},
		Sections: []model.PEP3CaregiverReportSection{
			pep3CaregiverCurrentDevelopmentSection(),
			pep3CaregiverDiagnosisSection(),
			pep3CaregiverProblemBehaviorSection(),
			pep3CaregiverPersonalSelfCareSection(),
			pep3CaregiverAdaptiveBehaviorSection(),
		},
	}
}

func pep3CaregiverCurrentDevelopmentSection() model.PEP3CaregiverReportSection {
	return model.PEP3CaregiverReportSection{
		SectionCode: "current_development",
		Title:       "儿童现时发展程度",
		Description: "与其他没有特殊需要的同龄儿童比较，请照顾者估计儿童各方面表现大约相当于多少岁。",
		InputType:   "age_estimate",
		Scored:      false,
		Items: []model.PEP3CaregiverReportItem{
			pep3CaregiverAgeItem(1, "communication_age", "沟通表现（例如：模仿声音、牙牙学语、跟从指令、与人说话及明白他人说话内容）约为"),
			pep3CaregiverAgeItem(2, "motor_age", "大小肌肉能力（例如：使用肌肉来活动及使用手部操控物件）约为"),
			pep3CaregiverAgeItem(3, "social_age", "社交能力（例如：享受拥抱及谈话、与人交往、合作地玩耍、交朋友及跟随游戏规则）约为"),
			pep3CaregiverAgeItem(4, "self_care_age", "自理能力（例如：进食/饮、穿衣、洗澡和如厕）约为"),
			pep3CaregiverAgeItem(5, "thinking_age", "思考能力（例如：完成砌图、找寻隐藏物件及解决问题）约为"),
			pep3CaregiverAgeItem(6, "overall_age", "整体能力（所有技巧）约为"),
		},
	}
}

func pep3CaregiverDiagnosisSection() model.PEP3CaregiverReportSection {
	return model.PEP3CaregiverReportSection{
		SectionCode: "diagnosis",
		Title:       "诊断类别及程度",
		Description: "请用“是/不是/不知道”标记诊断类别，并标记该类别对儿童发展的影响程度。",
		InputType:   "diagnosis_matrix",
		Scored:      false,
		DiagnosisCategories: []model.PEP3CaregiverDiagnosisCategory{
			{Key: "language_disorder", Label: "语言障碍"},
			{Key: "autism", Label: "自闭症"},
			{Key: "emotional_disturbance", Label: "情绪困扰"},
			{Key: "learning_disability", Label: "学习障碍"},
			{Key: "intellectual_delay", Label: "智力迟缓"},
			{Key: "adhd", Label: "专注力不足/过度活跃症"},
			{Key: "asperger", Label: "亚氏保加症"},
			{Key: "schizophrenia", Label: "精神分裂症"},
			{Key: "pervasive_developmental_disorder", Label: "广泛发展障碍"},
			{Key: "rett_syndrome", Label: "蕾特氏症"},
			{Key: "other", Label: "其他"},
		},
	}
}

func pep3CaregiverProblemBehaviorSection() model.PEP3CaregiverReportSection {
	max := 20
	return model.PEP3CaregiverReportSection{
		SectionCode: "problem_behavior",
		Title:       "问题行为",
		Description: "评估儿童行为问题的严重程度。第1-10题单选并计分，第11题为补充说明。",
		InputType:   "single_choice",
		ScaleCode:   "PB",
		ScaleName:   "问题行为",
		Scored:      true,
		MaxRawScore: &max,
		Items: []model.PEP3CaregiverReportItem{
			pep3CaregiverProblemItem(1, "eye_contact_body_language", "缺乏或不适当的眼神接触、缺乏面部表情以及缺乏沟通上的身体语言"),
			pep3CaregiverProblemItem(2, "speech_delay_or_absent", "语言发展迟缓或完全不说话"),
			pep3CaregiverProblemItem(3, "restricted_repetitive_interest", "经常投入在一个或多个重复的兴趣或活动，兴趣异常强烈，或者兴趣过分狭窄"),
			pep3CaregiverProblemItem(4, "peer_friendship", "不能如其他同龄孩子一样发展友谊"),
			pep3CaregiverProblemItem(5, "conversation_initiation", "有适当的语言，但不能主动与人开展对话或与人保持对话"),
			pep3CaregiverProblemItem(6, "nonfunctional_routines", "坚持某个或某些重复而没有实际功能的常规或仪式"),
			pep3CaregiverProblemItem(7, "sharing_enjoyment", "不会自发地与人分享开心的活动、有趣的事物或成功的事情"),
			pep3CaregiverProblemItem(8, "repetitive_or_odd_language", "使用重复或古怪的语言"),
			pep3CaregiverProblemItem(9, "repetitive_body_movements", "出现重复的身体活动，例如拍打或摆动手/手指、扭曲身体或作出复杂的身体动作"),
			pep3CaregiverProblemItem(10, "emotion_response", "与人沟通时，不表达自己感受或不回应别人所表达的情绪"),
			pep3CaregiverTextItem(11, "other_problem_behavior", "其他问题行为（请列出）"),
		},
	}
}

func pep3CaregiverPersonalSelfCareSection() model.PEP3CaregiverReportSection {
	max := 26
	return model.PEP3CaregiverReportSection{
		SectionCode: "personal_self_care",
		Title:       "个人自理",
		Description: "选择最能表达儿童个人自理情况的一项。第1-13题单选并计分，第14题为补充说明。",
		InputType:   "single_choice",
		ScaleCode:   "PSC",
		ScaleName:   "个人自理",
		Scored:      true,
		MaxRawScore: &max,
		Items: []model.PEP3CaregiverReportItem{
			pep3CaregiverScoredItem(1, "chew_swallow", "在饭餐或小食时间，儿童会否拿起细小食物、咀嚼然后吞咽？", "能自行把食物放进口中、咀嚼及吞咽而没有出现问题", "在放食物进口中、咀嚼及吞咽过程中会出现问题", "不会自行放食物进口中、咀嚼及吞咽"),
			pep3CaregiverScoredItem(2, "drink_from_cup", "在饭餐或小食时间，儿童能否以一只手拿起杯及饮用时不弄泻？", "能以一只手拿起杯及饮用时不弄泻", "未能灵巧地以杯进饮及饮用时会弄泻", "需要协助才能以杯进饮"),
			pep3CaregiverScoredItem(3, "use_spoon_fork", "在饭餐时，儿童能否以匙羹和叉自行进食？", "能以匙羹和叉自行进食，或许进食时会有一点食物跌在桌子上", "能以匙羹或叉自行进食一点食物", "未能使用任何一种餐具自行进食"),
			pep3CaregiverScoredItem(4, "pour_drink", "在饭餐或小食时间，儿童能否拿起一个盛着饮品的小容器，然后把饮品倒进一个杯内而只有少许弄泻？", "能把饮品倒进杯内而只有少许弄泻", "尝试把饮品倒进杯内时会有一些弄泻", "从未尝试把饮品倒进杯内，或把饮品倒进杯内时会大量弄泻"),
			pep3CaregiverScoredItem(5, "wash_hands_face", "在饭餐或小食时间后，儿童能否自行洗手和洗脸？", "不需要成人协助下，能自行以水和肥皂洗手和洗脸", "尝试自己洗手和洗脸，但需要成人协助", "未尝试自己洗手和洗脸"),
			pep3CaregiverScoredItem(6, "bathe", "儿童能否自行洗澡？", "不需要成人协助，能自行洗澡", "尝试自行洗澡，但需要成人协助", "未尝试自行洗澡"),
			pep3CaregiverScoredItem(7, "brush_teeth", "在饭餐或小食时间后，儿童能否自行刷牙？", "能自行刷牙而不需要成人协助", "尝试刷牙，但需要成人协助", "未尝试自行刷牙"),
			pep3CaregiverScoredItem(8, "undress", "儿童能否自行脱衣服？", "能自行脱衣服而不需要成人协助", "尝试自行脱衣服，但需要成人协助", "需要很多协助才能脱衣服"),
			pep3CaregiverScoredItem(9, "dress_and_shoes", "儿童能否自行完成穿衣服过程，包括绑鞋带？", "能自行完成穿衣服过程，包括绑鞋带，而不需要成人协助", "尝试自行完成穿衣服过程，但需要成人协助", "需要很多协助才能穿衣服"),
			pep3CaregiverScoredItem(10, "choose_clothes", "儿童能否自行挑选当日要穿的衣服，包括能选择适合当日天气及活动的衣服？", "能独立选择适合当日天气及活动的衣服", "在成人协助下，能选择适合当日天气及活动的衣服", "未能独立选择合适衣服"),
			pep3CaregiverScoredItem(11, "toileting", "儿童能否使用厕所如厕而没有意外（遗尿/遗便）？", "无需成人协助，能使用厕所如厕，而每星期不多过两次意外", "需成人协助才能使用厕所如厕，而每星期超过两次意外", "未能使用厕所如厕"),
			pep3CaregiverScoredItem(12, "remember_object_location", "儿童能否记得他/她的玩具或其他日常物件的摆放位置？", "能记得他/她的玩具或其他日常物件的摆放位置", "如经过练习，能记得一些物件的摆放位置", "未能记得物件的摆放位置"),
			pep3CaregiverScoredItem(13, "sleep_through_night", "儿童是否整晚都能安睡？", "通常整晚都能安睡", "有时整晚都能安睡", "很少可以整晚安睡"),
			pep3CaregiverTextItem(14, "other_self_care_problem", "其他自理问题（请列出）"),
		},
	}
}

func pep3CaregiverAdaptiveBehaviorSection() model.PEP3CaregiverReportSection {
	max := 30
	return model.PEP3CaregiverReportSection{
		SectionCode: "adaptive_behavior",
		Title:       "适应行为",
		Description: "选择最能表达儿童现时适应行为的一项。第1-15题单选并计分，第16题为补充说明。",
		InputType:   "single_choice",
		ScaleCode:   "AB",
		ScaleName:   "适应行为",
		Scored:      true,
		MaxRawScore: &max,
		Items: []model.PEP3CaregiverReportItem{
			pep3CaregiverScoredItem(1, "activity_transition", "儿童会否在一天之中自发转换活动？能否无须你的帮助，从一项活动转至另一项活动？", "在一天之中，会自发转换活动", "偶然会尝试新活动或转换活动", "喜欢不断重复同一项活动"),
			pep3CaregiverScoredItem(2, "eye_contact_with_people", "儿童会否望着周围的人？当对人说话或有人对他/她说话时，会否望着对方的脸？", "经常与人有眼神接触", "有时与人有眼神接触", "避开与人有眼神接触"),
			pep3CaregiverScoredItem(3, "hug_response", "如果你拥抱儿童，他/她会否正面地回应你？他/她会否自发拥抱别人？", "被熟识的成人拥抱时，他/她经常会有正面回应", "被熟识的成人拥抱时，他/她有时会有正面回应", "被熟识的成人拥抱时，他/她有负面反应或表现被动"),
			pep3CaregiverScoredItem(4, "interest_in_surroundings", "儿童会否对身边发生的事显得有兴趣？会否主动走去观看有趣的事情以及尝试加入其活动？", "常常四处观看以及回应身边所发生的事", "对身边所发生的事显得有少许兴趣", "对身边所发生的事显得没有兴趣"),
			pep3CaregiverScoredItem(5, "approach_peers", "儿童会否望着周围的孩子？会否主动地接近其他同龄孩子以及尝试与他们一起玩耍？", "常常会主动接近其他同龄孩子以及尝试与他们一起玩耍", "有时会主动接近其他同龄孩子以及与他们一起玩耍", "通常独自玩耍"),
			pep3CaregiverScoredItem(6, "new_activities", "当有新活动时，儿童会否参与？会否毫无犹疑地尝试不同的新经验？", "毫无犹疑地尝试不同的新活动", "在尝试新活动前，会表现犹疑不决", "常常自愿自行自己的活动，被打扰时会显得不快"),
			pep3CaregiverScoredItem(7, "learn_new_skills", "儿童是否可以快速地学习新技巧？你有否发现他/她常常用新学的技巧？", "能经常学习新技巧", "有时会学习新技巧", "很少学习新技巧，通常投入于仪式化行为"),
			pep3CaregiverScoredItem(8, "reality_engagement", "儿童是否投入于每天所发生的事情？是否表现明白现实世界，而无显得混沌或脱离现实？", "身体上和心智上都经常投入于环境中的事和物", "有时投入于环境中的事和物，但有时脱离现实", "经常自我刺激，不投入环境中的事和物"),
			pep3CaregiverScoredItem(9, "pretend_play", "儿童在玩玩具时有否出现假想的情节？会否以人的特性加入成为玩具的性情，赋予玩具生命力？", "经常玩玩具时有假想的情节", "有时玩玩具时有假想的情节", "没有假想的情节，似是对玩具或物件的部分表现过分着迷"),
			pep3CaregiverScoredItem(10, "safety_awareness", "儿童能否常常在日常活动中表现已有的常识，以及避免发生意外或受伤？他/她会否在感到挫败时克制自己而不会咬或拍打自己？", "玩耍时经常表现很小心和能避免发生意外或受伤", "玩耍时有时表现笨拙和可能会使自己受伤", "经常使自己受伤，或在感到挫败时咬或拍打自己"),
			pep3CaregiverScoredItem(11, "sit_stand_calmly", "儿童能否安定地坐或站，而没有如摇晃、拍打手或其他怪异动作？", "站立时竖直和能安坐，并没有怪异的动作", "有时站立竖直和能安坐，但当感到疲倦时，有时会做出怪异的动作", "在坐或站时经常做出怪异的动作"),
			pep3CaregiverScoredItem(12, "success_response", "儿童会否主动地请你留意他/她的成功表现？", "经常主动地请你留意他/她的成功表现", "很少或不恒常地对自己的成功表现满足", "不会因自己的成功而表现满足"),
			pep3CaregiverScoredItem(13, "play_with_children", "儿童会否与其他孩子一起玩耍？", "会回应及与其他孩子一起玩耍", "会对在场的孩子有反应", "对在场的孩子没有兴趣"),
			pep3CaregiverScoredItem(14, "room_to_room", "儿童能否不需你的协助而从一个房间走到另一个房间？", "能无需成人协助，从一个房间走到另一个房间", "不常常自行从一个房间走到另一个房间，或需成人协助才能做到", "不察觉环境，需要协助才能在转换活动时到另一个房间"),
			pep3CaregiverScoredItem(15, "street_safety", "儿童会否在行人路行走时和过马路前在街角停顿？", "会在行人路行走、在街角停顿和等待你与他/她一起过马路", "会在行人路行走，但需要你协助才会在街角停顿", "不察觉在街上行走时所需要注意的事情"),
			pep3CaregiverTextItem(16, "other_relationship_abnormality", "在与人关系上，儿童有否出现其他异常的表现？"),
		},
	}
}

func pep3CaregiverAgeItem(itemNo int, key, prompt string) model.PEP3CaregiverReportItem {
	return model.PEP3CaregiverReportItem{
		ItemNo:    itemNo,
		Key:       key,
		Prompt:    prompt,
		FieldType: "number",
		Unit:      "岁",
		Scored:    false,
	}
}

func pep3CaregiverTextItem(itemNo int, key, prompt string) model.PEP3CaregiverReportItem {
	return model.PEP3CaregiverReportItem{
		ItemNo:    itemNo,
		Key:       key,
		Prompt:    prompt,
		FieldType: "textarea",
		Scored:    false,
	}
}

func pep3CaregiverProblemItem(itemNo int, key, prompt string) model.PEP3CaregiverReportItem {
	return model.PEP3CaregiverReportItem{
		ItemNo:    itemNo,
		Key:       key,
		Prompt:    prompt,
		FieldType: "radio",
		Scored:    true,
		Options: []model.PEP3CaregiverReportOption{
			pep3CaregiverOption("not_present", "此问题没有出现", 2),
			pep3CaregiverOption("mild_moderate", "此问题有出现，程度是轻微/中度", 1),
			pep3CaregiverOption("severe", "此问题有出现，程度是严重", 0),
		},
	}
}

func pep3CaregiverScoredItem(itemNo int, key, prompt, score2Label, score1Label, score0Label string) model.PEP3CaregiverReportItem {
	return model.PEP3CaregiverReportItem{
		ItemNo:    itemNo,
		Key:       key,
		Prompt:    prompt,
		FieldType: "radio",
		Scored:    true,
		Options: []model.PEP3CaregiverReportOption{
			pep3CaregiverOption("score_2", score2Label, 2),
			pep3CaregiverOption("score_1", score1Label, 1),
			pep3CaregiverOption("score_0", score0Label, 0),
		},
	}
}

func pep3CaregiverOption(value, label string, score int) model.PEP3CaregiverReportOption {
	return model.PEP3CaregiverReportOption{
		Value: value,
		Label: label,
		Score: pep3ScorePtr(score),
	}
}

func pep3ScorePtr(score int) *int {
	return &score
}
