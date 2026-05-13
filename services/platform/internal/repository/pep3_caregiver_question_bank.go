package repository

import (
	"fmt"

	"go-migration-platform/services/platform/internal/model"
)

type pep3CaregiverQuestion struct {
	DomainCode string
	DomainName string
	ItemNo     int
	Prompt     string
	Score2     string
	Score1     string
	Score0     string
}

func pep3CaregiverQuestionBankItems() []model.ScaleQuestionBankItem {
	questions := []pep3CaregiverQuestion{
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 1, Prompt: "缺乏或不适当的眼神接触、缺乏面部表情以及缺乏沟通上的身体语言", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 2, Prompt: "语言发展迟缓或完全不说话", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 3, Prompt: "经常投入在一个或多个重复的兴趣或活动，兴趣异常强烈，或者兴趣过分狭窄", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 4, Prompt: "不能如其他同龄孩子一样发展友谊", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 5, Prompt: "有适当的语言，但不能主动与人开展对话或与人保持对话", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 6, Prompt: "坚持某个或某些重复而没有实际功能的常规或仪式", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 7, Prompt: "不会自发地与人分享开心的活动、有趣的事物或成功的事情", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 8, Prompt: "使用重复或古怪的语言", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 9, Prompt: "出现重复的身体活动，例如拍打或摆动手/手指、扭曲身体或作出复杂的身体动作", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PB", DomainName: "问题行为", ItemNo: 10, Prompt: "与人沟通时，不表达自己感受或不回应别人所表达的情绪", Score2: "此问题没有出现", Score1: "此问题有出现，程度是轻微/中度", Score0: "此问题有出现，程度是严重"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 1, Prompt: "在饭餐或小食时间，儿童会否拿起细小食物、咀嚼然后吞咽？", Score2: "能自行把食物放进口中、咀嚼及吞咽而没有出现问题", Score1: "在放食物进口中、咀嚼及吞咽过程中会出现问题", Score0: "不会自行放食物进口中、咀嚼及吞咽"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 2, Prompt: "在饭餐或小食时间，儿童能否以一只手拿起杯及饮用时不弄泻？", Score2: "能以一只手拿起杯及饮用时不弄泻", Score1: "未能灵巧地以杯进饮及饮用时会弄泻", Score0: "需要协助才能以杯进饮"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 3, Prompt: "在饭餐时，儿童能否以匙羹和叉自行进食？", Score2: "能以匙羹和叉自行进食，或许进食时会有一点食物跌在桌子上", Score1: "能以匙羹或叉自行进食一点食物", Score0: "未能使用任何一种餐具自行进食"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 4, Prompt: "在饭餐或小食时间，儿童能否拿起一个盛着饮品的小容器，然后把饮品倒进一个杯内而只有少许弄泻？", Score2: "能把饮品倒进杯内而只有少许弄泻", Score1: "尝试把饮品倒进杯内时会有一些弄泻", Score0: "从未尝试把饮品倒进杯内，或把饮品倒进杯内时会大量弄泻"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 5, Prompt: "在饭餐或小食时间后，儿童能否自行洗手和洗脸？", Score2: "不需要成人协助下，能自行以水和肥皂洗手和洗脸", Score1: "尝试自己洗手和洗脸，但需要成人协助", Score0: "未尝试自己洗手和洗脸"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 6, Prompt: "儿童能否自行洗澡？", Score2: "不需要成人协助，能自行洗澡", Score1: "尝试自行洗澡，但需要成人协助", Score0: "未尝试自行洗澡"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 7, Prompt: "在饭餐或小食时间后，儿童能否自行刷牙？", Score2: "能自行刷牙而不需要成人协助", Score1: "尝试刷牙，但需要成人协助", Score0: "未尝试自行刷牙"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 8, Prompt: "儿童能否自行脱衣服？", Score2: "能自行脱衣服而不需要成人协助", Score1: "尝试自行脱衣服，但需要成人协助", Score0: "需要很多协助才能脱衣服"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 9, Prompt: "儿童能否自行完成穿衣服过程，包括绑鞋带？", Score2: "能自行完成穿衣服过程，包括绑鞋带，而不需要成人协助", Score1: "尝试自行完成穿衣服过程，但需要成人协助", Score0: "需要很多协助才能穿衣服"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 10, Prompt: "儿童能否自行挑选当日要穿的衣服，包括能选择适合当日天气及活动的衣服？", Score2: "能独立选择适合当日天气及活动的衣服", Score1: "在成人协助下，能选择适合当日天气及活动的衣服", Score0: "未能独立选择合适衣服"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 11, Prompt: "儿童能否使用厕所如厕而没有意外（遗尿/遗便）？", Score2: "无需成人协助，能使用厕所如厕，而每星期不多过两次意外", Score1: "需成人协助才能使用厕所如厕，而每星期超过两次意外", Score0: "未能使用厕所如厕"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 12, Prompt: "儿童能否记得他/她的玩具或其他日常物件的摆放位置？", Score2: "能记得他/她的玩具或其他日常物件的摆放位置", Score1: "如经过练习，能记得一些物件的摆放位置", Score0: "未能记得物件的摆放位置"},
		{DomainCode: "PSC", DomainName: "个人自理", ItemNo: 13, Prompt: "儿童是否整晚都能安睡？", Score2: "通常整晚都能安睡", Score1: "有时整晚都能安睡", Score0: "很少可以整晚安睡"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 1, Prompt: "儿童会否在一天之中自发转换活动？能否无须你的帮助，从一项活动转至另一项活动？", Score2: "在一天之中，会自发转换活动", Score1: "偶然会尝试新活动或转换活动", Score0: "喜欢不断重复同一项活动"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 2, Prompt: "儿童会否望着周围的人？当对人说话或有人对他/她说话时，会否望着对方的脸？", Score2: "经常与人有眼神接触", Score1: "有时与人有眼神接触", Score0: "避开与人有眼神接触"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 3, Prompt: "如果你拥抱儿童，他/她会否正面地回应你？他/她会否自发拥抱别人？", Score2: "被熟识的成人拥抱时，他/她经常会有正面回应", Score1: "被熟识的成人拥抱时，他/她有时会有正面回应", Score0: "被熟识的成人拥抱时，他/她有负面反应或表现被动"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 4, Prompt: "儿童会否对身边发生的事显得有兴趣？会否主动走去观看有趣的事情以及尝试加入其活动？", Score2: "常常四处观看以及回应身边所发生的事", Score1: "对身边所发生的事显得有少许兴趣", Score0: "对身边所发生的事显得没有兴趣"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 5, Prompt: "儿童会否望着周围的孩子？会否主动地接近其他同龄孩子以及尝试与他们一起玩耍？", Score2: "常常会主动接近其他同龄孩子以及尝试与他们一起玩耍", Score1: "有时会主动接近其他同龄孩子以及与他们一起玩耍", Score0: "通常独自玩耍"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 6, Prompt: "当有新活动时，儿童会否参与？会否毫无犹疑地尝试不同的新经验？", Score2: "毫无犹疑地尝试不同的新活动", Score1: "在尝试新活动前，会表现犹疑不决", Score0: "常常自愿自行自己的活动，被打扰时会显得不快"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 7, Prompt: "儿童是否可以快速地学习新技巧？你有否发现他/她常常用新学的技巧？", Score2: "能经常学习新技巧", Score1: "有时会学习新技巧", Score0: "很少学习新技巧，通常投入于仪式化行为"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 8, Prompt: "儿童是否投入于每天所发生的事情？是否表现明白现实世界，而无显得混沌或脱离现实？", Score2: "身体上和心智上都经常投入于环境中的事和物", Score1: "有时投入于环境中的事和物，但有时脱离现实", Score0: "经常自我刺激，不投入环境中的事和物"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 9, Prompt: "儿童在玩玩具时有否出现假想的情节？会否以人的特性加入成为玩具的性情，赋予玩具生命力？", Score2: "经常玩玩具时有假想的情节", Score1: "有时玩玩具时有假想的情节", Score0: "没有假想的情节，似是对玩具或物件的部分表现过分着迷"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 10, Prompt: "儿童能否常常在日常活动中表现已有的常识，以及避免发生意外或受伤？他/她会否在感到挫败时克制自己而不会咬或拍打自己？", Score2: "玩耍时经常表现很小心和能避免发生意外或受伤", Score1: "玩耍时有时表现笨拙和可能会使自己受伤", Score0: "经常使自己受伤，或在感到挫败时咬或拍打自己"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 11, Prompt: "儿童能否安定地坐或站，而没有如摇晃、拍打手或其他怪异动作？", Score2: "站立时竖直和能安坐，并没有怪异的动作", Score1: "有时站立竖直和能安坐，但当感到疲倦时，有时会做出怪异的动作", Score0: "在坐或站时经常做出怪异的动作"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 12, Prompt: "儿童会否主动地请你留意他/她的成功表现？", Score2: "经常主动地请你留意他/她的成功表现", Score1: "很少或不恒常地对自己的成功表现满足", Score0: "不会因自己的成功而表现满足"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 13, Prompt: "儿童会否与其他孩子一起玩耍？", Score2: "会回应及与其他孩子一起玩耍", Score1: "会对在场的孩子有反应", Score0: "对在场的孩子没有兴趣"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 14, Prompt: "儿童能否不需你的协助而从一个房间走到另一个房间？", Score2: "能无需成人协助，从一个房间走到另一个房间", Score1: "不常常自行从一个房间走到另一个房间，或需成人协助才能做到", Score0: "不察觉环境，需要协助才能在转换活动时到另一个房间"},
		{DomainCode: "AB", DomainName: "适应行为", ItemNo: 15, Prompt: "儿童会否在行人路行走时和过马路前在街角停顿？", Score2: "会在行人路行走、在街角停顿和等待你与他/她一起过马路", Score1: "会在行人路行走，但需要你协助才会在街角停顿", Score0: "不察觉在街上行走时所需要注意的事情"},
	}

	items := make([]model.ScaleQuestionBankItem, 0, len(questions))
	for _, question := range questions {
		itemNo := pep3CaregiverQuestionItemNo(question.DomainCode, question.ItemNo)
		title := fmt.Sprintf("（%s%d）%s", question.DomainCode, question.ItemNo, question.Prompt)
		items = append(items, model.ScaleQuestionBankItem{
			ItemNo:     itemNo,
			ItemTitle:  title,
			TestItem:   title,
			DomainCode: question.DomainCode,
			DomainName: question.DomainName,
			Standard:   fmt.Sprintf("2分：%s\n1分：%s\n0分：%s", question.Score2, question.Score1, question.Score0),
			ScoreOptions: []model.ScaleQuestionBankScoreOption{
				{Value: 2, Label: "2分", Description: question.Score2},
				{Value: 1, Label: "1分", Description: question.Score1},
				{Value: 0, Label: "0分", Description: question.Score0},
			},
		})
	}
	return items
}

func pep3CaregiverQuestionItemNo(domainCode string, itemNo int) int {
	switch domainCode {
	case "PB":
		return 2000 + itemNo
	case "PSC":
		return 2100 + itemNo
	case "AB":
		return 2200 + itemNo
	default:
		return 2900 + itemNo
	}
}
