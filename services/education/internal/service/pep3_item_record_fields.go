package service

import "go-migration-platform/services/education/internal/model"

func pep3ItemRecordFields(itemNo int) []model.PEP3ItemRecordField {
	fields := pep3ItemRecordFieldDefinitions()[itemNo]
	out := make([]model.PEP3ItemRecordField, 0, len(fields))
	for _, field := range fields {
		copied := field
		copied.Options = append([]model.PEP3ItemRecordFieldOption(nil), field.Options...)
		out = append(out, copied)
	}
	return out
}

func pep3ItemRecordFieldDefinitions() map[int][]model.PEP3ItemRecordField {
	return map[int][]model.PEP3ItemRecordField{
		5: {
			pep3RadioRecordField("touch_block_reaction", "触觉块反应",
				pep3RecordFieldOption("no_interest", "无兴趣"),
				pep3RecordFieldOption("unusual_interest", "怪异兴趣"),
			),
		},
		6: {
			pep3MultiRecordField("kaleidoscope_action", "万花筒动作",
				pep3RecordFieldOption("watch", "观看"),
				pep3RecordFieldOption("turn", "扭动"),
				pep3RecordFieldOption("watch_and_turn", "观看+扭动"),
			),
		},
		7: {
			pep3DominanceRecordField("first_observation", "第1次观察",
				pep3RecordFieldOption("left_eye", "左眼"),
				pep3RecordFieldOption("right_eye", "右眼"),
			),
			pep3DominanceRecordField("second_observation", "第2次观察",
				pep3RecordFieldOption("left_eye", "左眼"),
				pep3RecordFieldOption("right_eye", "右眼"),
			),
		},
		9: {
			pep3MultiRecordField("bell_attempts", "响铃尝试",
				pep3RecordFieldOption("first_attempt", "第1次尝试"),
				pep3RecordFieldOption("second_attempt", "第2次尝试"),
			),
		},
		16: {
			pep3MultiRecordField("daily_actions", "模仿动作",
				pep3SameValueRecordFieldOption("喂食"),
				pep3SameValueRecordFieldOption("饮水"),
				pep3SameValueRecordFieldOption("刷牙"),
				pep3SameValueRecordFieldOption("抹鼻"),
			),
		},
		17: {
			pep3MultiRecordField("puppet_body_parts", "指出手偶身体部位",
				pep3SameValueRecordFieldOption("眼"),
				pep3SameValueRecordFieldOption("耳"),
				pep3SameValueRecordFieldOption("口"),
				pep3SameValueRecordFieldOption("鼻"),
			),
		},
		18: {
			pep3MultiRecordField("self_body_parts", "指出自己身体部位",
				pep3SameValueRecordFieldOption("眼"),
				pep3SameValueRecordFieldOption("耳"),
				pep3SameValueRecordFieldOption("口"),
				pep3SameValueRecordFieldOption("鼻"),
			),
		},
		21: {
			pep3MultiRecordField("shape_positions", "正确位置",
				pep3SameValueRecordFieldOption("三角形"),
				pep3SameValueRecordFieldOption("圆形"),
				pep3SameValueRecordFieldOption("正方形"),
			),
		},
		22: {
			pep3MultiRecordField("shape_board_completed", "完成形状拼板",
				pep3SameValueRecordFieldOption("三角形"),
				pep3SameValueRecordFieldOption("圆形"),
				pep3SameValueRecordFieldOption("正方形"),
			),
		},
		23: {
			pep3MultiRecordField("shape_names", "说出形状名称",
				pep3SameValueRecordFieldOption("三角形"),
				pep3SameValueRecordFieldOption("圆形"),
				pep3SameValueRecordFieldOption("正方形"),
			),
		},
		24: {
			pep3MultiRecordField("shape_selection", "挑选形状",
				pep3SameValueRecordFieldOption("三角形"),
				pep3SameValueRecordFieldOption("圆形"),
				pep3SameValueRecordFieldOption("正方形"),
			),
		},
		25: {
			pep3MultiRecordField("object_puzzle_completed", "完成物件拼板",
				pep3SameValueRecordFieldOption("小鸡"),
				pep3SameValueRecordFieldOption("雨伞"),
				pep3SameValueRecordFieldOption("蝴蝶"),
				pep3SameValueRecordFieldOption("雪梨"),
			),
		},
		27: {
			pep3MultiRecordField("mitten_position_sizes", "手套拼块正确位置",
				pep3SameValueRecordFieldOption("大"),
				pep3SameValueRecordFieldOption("中"),
				pep3SameValueRecordFieldOption("小"),
			),
		},
		28: {
			pep3MultiRecordField("mitten_completed_sizes", "完成手套拼板",
				pep3SameValueRecordFieldOption("大"),
				pep3SameValueRecordFieldOption("中"),
				pep3SameValueRecordFieldOption("小"),
			),
		},
		29: {
			pep3MultiRecordField("size_naming", "说出物件大小",
				pep3RecordFieldOption("first_big", "第1次：大"),
				pep3RecordFieldOption("first_small", "第1次：小"),
				pep3RecordFieldOption("second_big", "第2次：大"),
				pep3RecordFieldOption("second_small", "第2次：小"),
			),
		},
		30: {
			pep3MultiRecordField("size_selection", "挑选大小物件",
				pep3RecordFieldOption("first_big", "第1次：大"),
				pep3RecordFieldOption("first_small", "第1次：小"),
				pep3RecordFieldOption("second_big", "第2次：大"),
				pep3RecordFieldOption("second_small", "第2次：小"),
			),
		},
		31: {
			pep3RadioRecordField("cat_puzzle_prompt", "完成方式",
				pep3SameValueRecordFieldOption("自行"),
				pep3SameValueRecordFieldOption("需示范"),
			),
			pep3NumberRecordField("completed_piece_count", "完成块数", "填写块数"),
		},
		32: {
			pep3NumberRecordField("interlocked_piece_count", "紧扣块数", "填写块数"),
		},
		33: {
			pep3RadioRecordField("cow_puzzle_prompt", "完成方式",
				pep3SameValueRecordFieldOption("自行"),
				pep3SameValueRecordFieldOption("需示范"),
			),
			pep3NumberRecordField("completed_piece_count", "完成块数", "填写块数"),
		},
		34: {
			pep3MultiRecordField("boy_puzzle_parts", "男孩拼图部位",
				pep3SameValueRecordFieldOption("头"),
				pep3SameValueRecordFieldOption("头发"),
				pep3SameValueRecordFieldOption("双眼"),
				pep3SameValueRecordFieldOption("鼻"),
				pep3SameValueRecordFieldOption("口"),
				pep3SameValueRecordFieldOption("身"),
				pep3SameValueRecordFieldOption("脚"),
			),
		},
		35: {
			pep3MultiRecordField("sound_objects", "模仿使用发声物",
				pep3SameValueRecordFieldOption("响板"),
				pep3SameValueRecordFieldOption("手铃"),
				pep3SameValueRecordFieldOption("匙子"),
			),
		},
		36: {
			pep3TextRecordField("sock", "袜子", "记录儿童回答"),
			pep3TextRecordField("cup", "杯子", "记录儿童回答"),
			pep3TextRecordField("toothbrush", "牙刷", "记录儿童回答"),
			pep3TextRecordField("crayon", "蜡笔", "记录儿童回答"),
			pep3TextRecordField("scissors", "剪刀", "记录儿童回答"),
			pep3TextRecordField("comb", "梳子", "记录儿童回答"),
			pep3TextRecordField("pencil", "铅笔", "记录儿童回答"),
		},
		37: {
			pep3MultiRecordField("object_use", "正确使用物件",
				pep3SameValueRecordFieldOption("杯子"),
				pep3SameValueRecordFieldOption("匙子"),
				pep3SameValueRecordFieldOption("蜡笔"),
				pep3SameValueRecordFieldOption("梳子"),
				pep3SameValueRecordFieldOption("剪刀"),
			),
		},
		38: {
			pep3MultiRecordField("requested_objects", "按要求交出物件",
				pep3SameValueRecordFieldOption("杯子"),
				pep3SameValueRecordFieldOption("匙子"),
				pep3SameValueRecordFieldOption("蜡笔"),
				pep3SameValueRecordFieldOption("梳子"),
				pep3SameValueRecordFieldOption("剪刀"),
			),
		},
		39: {
			pep3MultiRecordField("matched_picture_objects", "实物配对图片",
				pep3SameValueRecordFieldOption("袜子"),
				pep3SameValueRecordFieldOption("杯子"),
				pep3SameValueRecordFieldOption("牙刷"),
				pep3SameValueRecordFieldOption("匙子"),
				pep3SameValueRecordFieldOption("剪刀"),
				pep3SameValueRecordFieldOption("梳子"),
				pep3SameValueRecordFieldOption("铅笔"),
			),
		},
		40: {
			pep3TextareaRecordField("pointed_objects", "指出物件", "记录3个物件或儿童反应"),
		},
		43: {
			pep3TextRecordField("first_attempt", "第1次", "记录表现"),
			pep3TextRecordField("second_attempt", "第2次", "记录表现"),
			pep3TextRecordField("third_attempt", "第3次", "记录表现"),
		},
		44: {
			pep3MultiRecordField("tactile_objects", "触觉辨别物件",
				pep3SameValueRecordFieldOption("球"),
				pep3SameValueRecordFieldOption("积木"),
				pep3SameValueRecordFieldOption("蜡笔"),
				pep3SameValueRecordFieldOption("硬币"),
				pep3SameValueRecordFieldOption("匙子"),
			),
		},
		45: {
			pep3RadioRecordField("material_inspection", "检视及运用测试材料",
				pep3SameValueRecordFieldOption("短暂"),
				pep3SameValueRecordFieldOption("没有"),
			),
		},
		46: {
			pep3RadioRecordField("visual_inspection", "视觉检视测试材料",
				pep3SameValueRecordFieldOption("过分兴趣"),
				pep3SameValueRecordFieldOption("过分抗拒"),
			),
		},
		49: {
			pep3RadioRecordField("body_contact", "身体接触反应",
				pep3SameValueRecordFieldOption("拒绝"),
				pep3SameValueRecordFieldOption("过分抗拒"),
			),
		},
		50: {
			pep3RadioRecordField("tickle_response", "被搔痒反应",
				pep3SameValueRecordFieldOption("拒绝"),
				pep3SameValueRecordFieldOption("过分反应"),
			),
		},
		52: {
			pep3RadioRecordField("social_communication", "社交沟通反应",
				pep3SameValueRecordFieldOption("被动"),
				pep3SameValueRecordFieldOption("完全无反应"),
			),
		},
		54: {
			pep3MultiRecordField("gross_motor_imitation", "模仿大肌肉动作",
				pep3SameValueRecordFieldOption("举手"),
				pep3SameValueRecordFieldOption("摸鼻"),
				pep3SameValueRecordFieldOption("举手+摸鼻"),
			),
		},
		58: {
			pep3TextRecordField("first_attempt", "第1次", "记录接球表现"),
			pep3TextRecordField("second_attempt", "第2次", "记录接球表现"),
			pep3TextRecordField("third_attempt", "第3次", "记录接球表现"),
		},
		59: {
			pep3TextRecordField("first_attempt", "第1次", "记录抛球表现"),
			pep3TextRecordField("second_attempt", "第2次", "记录抛球表现"),
			pep3TextRecordField("third_attempt", "第3次", "记录抛球表现"),
		},
		60: {
			pep3TextRecordField("first_attempt", "第1次", "记录踢球表现"),
			pep3TextRecordField("second_attempt", "第2次", "记录踢球表现"),
			pep3TextRecordField("third_attempt", "第3次", "记录踢球表现"),
		},
		61: {
			pep3DominanceRecordField("kick_ball", "踢球",
				pep3RecordFieldOption("left_foot", "左脚"),
				pep3RecordFieldOption("right_foot", "右脚"),
			),
			pep3DominanceRecordField("stairs", "上梯级",
				pep3RecordFieldOption("left_foot", "左脚"),
				pep3RecordFieldOption("right_foot", "右脚"),
			),
		},
		64: {
			pep3RadioRecordField("string_reaction", "绳子反应",
				pep3SameValueRecordFieldOption("无兴趣"),
				pep3SameValueRecordFieldOption("怪异反应"),
			),
		},
		65: {
			pep3NumberRecordField("completed_bead_count", "完成珠粒数", "填写完成几粒"),
		},
		67: {
			pep3NumberRecordField("completed_bead_count", "完成珠粒数", "填写完成几粒"),
		},
		71: {
			pep3DominanceRecordField("dominant_hand", "惯用手",
				pep3RecordFieldOption("left_hand", "左手"),
				pep3RecordFieldOption("right_hand", "右手"),
			),
		},
		72: {
			pep3MultiRecordField("traced_shapes", "沿线描画图形",
				pep3SameValueRecordFieldOption("圆形"),
				pep3SameValueRecordFieldOption("正方形"),
				pep3SameValueRecordFieldOption("三角形"),
				pep3SameValueRecordFieldOption("菱形"),
			),
		},
		84: {
			pep3MultiRecordField("pretend_picture_objects", "假装使用图画物件",
				pep3SameValueRecordFieldOption("哨子"),
				pep3SameValueRecordFieldOption("球"),
				pep3SameValueRecordFieldOption("鼓"),
				pep3SameValueRecordFieldOption("钥匙"),
				pep3SameValueRecordFieldOption("槌子"),
			),
		},
		85: {
			pep3PictureCheckRecordField("picture_identification", "辨认14幅图画", []string{
				"A 杯子", "B 洋娃娃", "C 锁匙", "D 飞机", "E 鸟笼",
				"F 雨伞", "G 煮食", "H 系鞋带", "I 门", "J 木偶",
				"K 公鸡", "L 接球", "M 建筑", "N 警察", "O 砌积木",
				"P 烧烤", "Q 洗澡", "R 水壶", "S 火车头", "T 火箭起飞",
			}),
		},
		86: {
			pep3PictureCheckRecordField("picture_naming", "说出14幅图画名称", []string{
				"A 牛", "B 皮球/波", "C 花", "D 婴儿车", "E 牙刷",
				"F 雪柜", "G 油油", "H 打秋千", "I 樽", "J 电风扇",
				"K 企鹅", "L 溜冰", "M 抱着狗", "N 医生", "O 跳水",
				"P 踏车", "Q 举起", "R 教堂", "S 炉", "T 整路",
			}),
		},
		87: {
			pep3TextareaRecordField("spoken_phrase", "4-5词片语", "记录儿童说出的片语"),
		},
		88: {
			pep3MultiRecordField("recognized_characters", "辨别字卡",
				pep3SameValueRecordFieldOption("人"),
				pep3SameValueRecordFieldOption("口"),
				pep3SameValueRecordFieldOption("上"),
				pep3SameValueRecordFieldOption("山"),
				pep3SameValueRecordFieldOption("刀"),
				pep3SameValueRecordFieldOption("天"),
				pep3SameValueRecordFieldOption("火"),
				pep3SameValueRecordFieldOption("手"),
				pep3SameValueRecordFieldOption("田"),
			),
		},
		89: {
			pep3MultiRecordField("read_characters", "读出字卡",
				pep3SameValueRecordFieldOption("人"),
				pep3SameValueRecordFieldOption("口"),
				pep3SameValueRecordFieldOption("上"),
				pep3SameValueRecordFieldOption("山"),
				pep3SameValueRecordFieldOption("刀"),
				pep3SameValueRecordFieldOption("天"),
				pep3SameValueRecordFieldOption("火"),
				pep3SameValueRecordFieldOption("手"),
				pep3SameValueRecordFieldOption("田"),
			),
		},
		90: {
			pep3MultiRecordField("matched_characters", "生字配对",
				pep3SameValueRecordFieldOption("人"),
				pep3SameValueRecordFieldOption("口"),
				pep3SameValueRecordFieldOption("上"),
				pep3SameValueRecordFieldOption("山"),
				pep3SameValueRecordFieldOption("刀"),
				pep3SameValueRecordFieldOption("天"),
				pep3SameValueRecordFieldOption("火"),
				pep3SameValueRecordFieldOption("手"),
				pep3SameValueRecordFieldOption("田"),
			),
		},
		92: {
			pep3MultiRecordField("read_words", "读出单字",
				pep3SameValueRecordFieldOption("球"),
				pep3SameValueRecordFieldOption("狗"),
				pep3SameValueRecordFieldOption("猫"),
				pep3SameValueRecordFieldOption("屋"),
			),
		},
		95: {
			pep3MultiRecordField("reading_comprehension_questions", "阅读理解问题",
				pep3SameValueRecordFieldOption("小明有哪些动物呀？"),
				pep3SameValueRecordFieldOption("小明在玩什么？"),
				pep3SameValueRecordFieldOption("什么跳过小明的皮球？"),
			),
		},
		96: {
			pep3MultiRecordField("sentence_commands", "句子及遵从指令",
				pep3SameValueRecordFieldOption("拿起皮球"),
				pep3SameValueRecordFieldOption("皮球放桌子上"),
			),
		},
		97: {
			pep3NumberRecordField("put_in_block_count", "放入块数", "填写块数"),
		},
		99: {
			pep3NumberRecordField("first_attempt", "第1次", "填写块数"),
			pep3NumberRecordField("second_attempt", "第2次", "填写块数"),
			pep3NumberRecordField("third_attempt", "第3次", "填写块数"),
		},
		101: {
			pep3NumberRecordField("two_blocks", "2块", "记录数出块数"),
			pep3NumberRecordField("six_blocks", "6块", "记录数出块数"),
		},
		102: {
			pep3NumberRecordField("two_blocks", "2块", "记录数出块数"),
			pep3NumberRecordField("seven_blocks", "7块", "记录数出块数"),
		},
		104: {
			pep3NumberRecordField("stack_completed", "完成积木", "填写块数"),
			pep3NumberRecordField("jar_completed", "完成筹码", "填写块数"),
		},
		105: {
			pep3MultiRecordField("matched_colors", "配对颜色",
				pep3SameValueRecordFieldOption("红"),
				pep3SameValueRecordFieldOption("黄"),
				pep3SameValueRecordFieldOption("蓝"),
				pep3SameValueRecordFieldOption("白"),
				pep3SameValueRecordFieldOption("绿"),
			),
		},
		106: {
			pep3MultiRecordField("named_colors", "说出颜色",
				pep3SameValueRecordFieldOption("红"),
				pep3SameValueRecordFieldOption("黄"),
				pep3SameValueRecordFieldOption("蓝"),
				pep3SameValueRecordFieldOption("白"),
				pep3SameValueRecordFieldOption("绿"),
			),
		},
		107: {
			pep3MultiRecordField("selected_colors", "按指示挑选颜色",
				pep3SameValueRecordFieldOption("红"),
				pep3SameValueRecordFieldOption("黄"),
				pep3SameValueRecordFieldOption("蓝"),
				pep3SameValueRecordFieldOption("白"),
				pep3SameValueRecordFieldOption("绿"),
			),
		},
		108: {
			pep3RadioRecordField("classification_prompt", "示范情况",
				pep3SameValueRecordFieldOption("自行完成"),
				pep3SameValueRecordFieldOption("部分示范"),
				pep3SameValueRecordFieldOption("全部示范"),
			),
			pep3RadioRecordField("classification_basis", "分类依据",
				pep3SameValueRecordFieldOption("颜色"),
				pep3SameValueRecordFieldOption("形状"),
			),
			pep3NumberRecordField("completed_card_count", "完成张数", "填写张数"),
		},
		111: {
			pep3MultiRecordField("imitated_sounds", "模仿声音",
				pep3SameValueRecordFieldOption("m-m-m"),
				pep3SameValueRecordFieldOption("ba-ba"),
				pep3SameValueRecordFieldOption("pa-ta"),
				pep3SameValueRecordFieldOption("la-la"),
			),
		},
		112: {
			pep3TextRecordField("digits_7_9", "7-9", "填写儿童复述内容"),
			pep3TextRecordField("digits_5_3", "5-3", "填写儿童复述内容"),
		},
		113: {
			pep3TextRecordField("digits_2_4_1", "2-4-1", "填写儿童复述内容"),
			pep3TextRecordField("digits_5_7_9", "5-7-9", "填写儿童复述内容"),
		},
		114: {
			pep3TextRecordField("word_street", "街街", "填写儿童复述内容"),
			pep3TextRecordField("word_car", "车车", "填写儿童复述内容"),
			pep3TextRecordField("word_bye", "拜拜", "填写儿童复述内容"),
		},
		115: {
			pep3ChoiceRecordField("repeated_sentences", "正确复述的短句", "checkbox_group",
				pep3RecordFieldOption("bb_looking", "BB望住"),
				pep3RecordFieldOption("want_biscuit", "我要饼干"),
				pep3RecordFieldOption("crying_loudly", "佢大声喊"),
			),
		},
		116: {
			pep3ChoiceRecordField("eye_contact", "望着测试员面孔", "radio",
				pep3RecordFieldOption("brief", "短暂"),
				pep3RecordFieldOption("none", "没有"),
			),
		},
		117: {
			pep3ChoiceRecordField("delayed_echolalia", "延迟性鹦鹉式讲话", "radio",
				pep3RecordFieldOption("not_applicable", "不适用"),
				pep3RecordFieldOption("too_much", "过多"),
			),
		},
		119: {
			pep3TextRecordField("pronoun_response", "俾我", "填写儿童回应"),
		},
		120: {
			pep3TextRecordField("spoken_words", "儿童使用的词语", "填写儿童使用的词语"),
		},
		121: {
			pep3TextRecordField("spoken_words", "儿童使用的词语", "填写儿童使用的词语"),
		},
		122: {
			pep3TextRecordField("spoken_phrase", "儿童的语句", "填写儿童的语句"),
		},
		123: {
			pep3MultiRecordField("oral_commands", "遵从口语指令",
				pep3SameValueRecordFieldOption("拍吓个盒"),
				pep3SameValueRecordFieldOption("摸吓只狗"),
				pep3SameValueRecordFieldOption("企起身跳"),
				pep3SameValueRecordFieldOption("攞个杯给我，然后坐下"),
				pep3SameValueRecordFieldOption("敲吓度门，然后摸吓度墙"),
			),
		},
		125: {
			pep3MultiRecordField("gesture_responses", "对手势的反应",
				pep3SameValueRecordFieldOption("叫名字+招手"),
				pep3SameValueRecordFieldOption("坐下+拿走积木"),
				pep3SameValueRecordFieldOption("交回颜色笔"),
				pep3SameValueRecordFieldOption("其他"),
			),
		},
		126: {
			pep3MultiRecordField("no_stop_commands", "回应指令",
				pep3SameValueRecordFieldOption("不要"),
				pep3SameValueRecordFieldOption("停止"),
			),
		},
		129: {
			pep3TextRecordField("child_answer", "儿童答案", "记录儿童答案"),
		},
		130: {
			pep3TextRecordField("child_answer", "儿童答案", "记录儿童答案"),
		},
		131: {
			pep3MultiRecordField("single_actions", "单项动作",
				pep3SameValueRecordFieldOption("跳"),
				pep3SameValueRecordFieldOption("坐下"),
				pep3SameValueRecordFieldOption("企起身"),
			),
		},
		133: {
			pep3MultiRecordField("wh_questions", "回答问句",
				pep3SameValueRecordFieldOption("何人"),
				pep3SameValueRecordFieldOption("何事"),
				pep3SameValueRecordFieldOption("何地"),
				pep3SameValueRecordFieldOption("何时"),
			),
		},
		134: {
			pep3MultiRecordField("simple_action_commands", "遵从简单动作指令",
				pep3SameValueRecordFieldOption("坐下"),
				pep3SameValueRecordFieldOption("起身"),
				pep3SameValueRecordFieldOption("过来"),
				pep3SameValueRecordFieldOption("伸我"),
				pep3SameValueRecordFieldOption("放低手"),
				pep3SameValueRecordFieldOption("开门"),
				pep3SameValueRecordFieldOption("其他"),
			),
		},
		135: {
			pep3RadioRecordField("visual_self_stimulation", "视觉自我刺激",
				pep3SameValueRecordFieldOption("正常"),
				pep3SameValueRecordFieldOption("过分刺激"),
			),
		},
		136: {
			pep3RadioRecordField("space_material_exploration", "探索空间及材料",
				pep3SameValueRecordFieldOption("无兴趣"),
				pep3SameValueRecordFieldOption("怪异行为"),
			),
		},
		137: {
			pep3RadioRecordField("environment_exploration", "探索测试房间环境",
				pep3SameValueRecordFieldOption("无兴趣"),
				pep3SameValueRecordFieldOption("怪异行为"),
			),
		},
		138: {
			pep3RadioRecordField("sound_response", "声音作出反应",
				pep3SameValueRecordFieldOption("无反应"),
				pep3SameValueRecordFieldOption("过度反应"),
			),
		},
		139: {
			pep3RadioRecordField("texture_exploration", "探索不同质感",
				pep3SameValueRecordFieldOption("抗拒"),
				pep3SameValueRecordFieldOption("怪异兴趣"),
			),
		},
		140: {
			pep3RadioRecordField("taste_use", "使用味觉",
				pep3SameValueRecordFieldOption("正常"),
				pep3SameValueRecordFieldOption("过分兴趣"),
			),
		},
		141: {
			pep3RadioRecordField("smell_interest", "嗅觉兴趣",
				pep3SameValueRecordFieldOption("正常"),
				pep3SameValueRecordFieldOption("过分兴趣"),
			),
		},
		142: {
			pep3RadioRecordField("completed_tests", "完成适龄测试项目",
				pep3SameValueRecordFieldOption("混乱"),
				pep3SameValueRecordFieldOption("怪异行为"),
			),
		},
		144: {
			pep3RadioRecordField("repeated_sentences_heard", "重复听到的字词或片语",
				pep3SameValueRecordFieldOption("不适用"),
				pep3SameValueRecordFieldOption("过多"),
			),
		},
		145: {
			pep3RadioRecordField("repeated_words_sounds", "重复某些字词或声音",
				pep3SameValueRecordFieldOption("不适用"),
				pep3SameValueRecordFieldOption("过多"),
			),
		},
		146: {
			pep3RadioRecordField("speech_tone_volume_speed", "说话语调、音量和速度",
				pep3SameValueRecordFieldOption("不适用"),
				pep3SameValueRecordFieldOption("怪异"),
			),
		},
		147: {
			pep3RadioRecordField("meaningless_sounds", "无意义声音",
				pep3SameValueRecordFieldOption("不适用"),
				pep3SameValueRecordFieldOption("过多"),
			),
		},
		148: {
			pep3RadioRecordField("age_appropriate_vocabulary", "词汇沟通",
				pep3SameValueRecordFieldOption("不适用"),
				pep3SameValueRecordFieldOption("不恰当"),
			),
		},
		149: {
			pep3RadioRecordField("self_talk", "说奇异话句与自创语",
				pep3SameValueRecordFieldOption("不适用"),
				pep3SameValueRecordFieldOption("过多"),
			),
		},
		150: {
			pep3RadioRecordField("age_appropriate_articulation", "发音能力",
				pep3SameValueRecordFieldOption("不适用"),
				pep3SameValueRecordFieldOption("无法明白"),
			),
		},
		151: {
			pep3RadioRecordField("spontaneous_communication", "自发性沟通",
				pep3SameValueRecordFieldOption("无"),
				pep3SameValueRecordFieldOption("怪异沉迷"),
			),
		},
		154: {
			pep3RadioRecordField("cooperation", "与测试员合作",
				pep3SameValueRecordFieldOption("反复"),
				pep3SameValueRecordFieldOption("过分反抗"),
			),
		},
		155: {
			pep3RadioRecordField("organization", "语句组织能力",
				pep3SameValueRecordFieldOption("间中"),
				pep3SameValueRecordFieldOption("经常混乱"),
			),
		},
		159: {
			pep3RadioRecordField("pleasant_emotion", "恰当的情感",
				pep3SameValueRecordFieldOption("无变化"),
				pep3SameValueRecordFieldOption("怪异反应"),
			),
		},
		160: {
			pep3RadioRecordField("fear_response", "恐惧的反应",
				pep3SameValueRecordFieldOption("无"),
				pep3SameValueRecordFieldOption("过多"),
			),
		},
		161: {
			pep3RadioRecordField("attention", "合乎年龄的专注力",
				pep3SameValueRecordFieldOption("短"),
				pep3SameValueRecordFieldOption("极端反应"),
			),
		},
		162: {
			pep3RadioRecordField("transition", "从容地转换活动",
				pep3SameValueRecordFieldOption("接受"),
				pep3SameValueRecordFieldOption("极端反应"),
			),
		},
		168: {
			pep3RadioRecordField("request_help", "向测试员求助",
				pep3SameValueRecordFieldOption("无"),
				pep3SameValueRecordFieldOption("过分要求"),
			),
		},
		169: {
			pep3RadioRecordField("movement", "动作及动静",
				pep3SameValueRecordFieldOption("正常"),
				pep3SameValueRecordFieldOption("怪异动作"),
			),
		},
		170: {
			pep3RadioRecordField("return_to_examiner", "恰当地回应测试员",
				pep3SameValueRecordFieldOption("不察觉"),
				pep3SameValueRecordFieldOption("被动"),
			),
		},
		171: {
			pep3RadioRecordField("reward_response", "对物质奖励有反应",
				pep3SameValueRecordFieldOption("不一致"),
				pep3SameValueRecordFieldOption("无"),
			),
		},
		172: {
			pep3RadioRecordField("social_reward_response", "对社交赞赏有反应",
				pep3SameValueRecordFieldOption("不一致"),
				pep3SameValueRecordFieldOption("无"),
			),
		},
	}
}

func pep3TextRecordField(key, label, placeholder string) model.PEP3ItemRecordField {
	return model.PEP3ItemRecordField{
		Key:         key,
		Label:       label,
		FieldType:   "text",
		DisplayType: "填空",
		Placeholder: placeholder,
	}
}

func pep3NumberRecordField(key, label, placeholder string) model.PEP3ItemRecordField {
	return model.PEP3ItemRecordField{
		Key:         key,
		Label:       label,
		FieldType:   "number",
		DisplayType: "数字",
		Placeholder: placeholder,
	}
}

func pep3TextareaRecordField(key, label, placeholder string) model.PEP3ItemRecordField {
	return model.PEP3ItemRecordField{
		Key:         key,
		Label:       label,
		FieldType:   "textarea",
		DisplayType: "填空",
		Placeholder: placeholder,
	}
}

func pep3RadioRecordField(key, label string, options ...model.PEP3ItemRecordFieldOption) model.PEP3ItemRecordField {
	return pep3ChoiceRecordField(key, label, "radio", options...)
}

func pep3DominanceRecordField(key, label string, options ...model.PEP3ItemRecordFieldOption) model.PEP3ItemRecordField {
	field := pep3ChoiceRecordField(key, label, "radio", options...)
	field.DisplayType = "选择"
	return field
}

func pep3MultiRecordField(key, label string, options ...model.PEP3ItemRecordFieldOption) model.PEP3ItemRecordField {
	return pep3ChoiceRecordField(key, label, "checkbox_group", options...)
}

func pep3ChoiceRecordField(key, label, fieldType string, options ...model.PEP3ItemRecordFieldOption) model.PEP3ItemRecordField {
	displayType := "填空"
	if fieldType == "radio" {
		displayType = "单选"
	}
	if fieldType == "checkbox_group" {
		displayType = "打勾"
	}
	return model.PEP3ItemRecordField{
		Key:         key,
		Label:       label,
		FieldType:   fieldType,
		DisplayType: displayType,
		Options:     append([]model.PEP3ItemRecordFieldOption(nil), options...),
	}
}

func pep3RecordFieldOption(value, label string) model.PEP3ItemRecordFieldOption {
	return model.PEP3ItemRecordFieldOption{Value: value, Label: label}
}

func pep3SameValueRecordFieldOption(label string) model.PEP3ItemRecordFieldOption {
	return model.PEP3ItemRecordFieldOption{Value: label, Label: label}
}

func pep3PictureAnswerFields(prefix string, labels []string) []model.PEP3ItemRecordField {
	fields := make([]model.PEP3ItemRecordField, 0, len(labels))
	for index, label := range labels {
		key := prefix + "_" + string(rune('a'+index)) + "_answer"
		fields = append(fields, pep3TextRecordField(key, label, "记录儿童答案"))
	}
	return fields
}

func pep3PictureCheckRecordField(key, label string, labels []string) model.PEP3ItemRecordField {
	options := make([]model.PEP3ItemRecordFieldOption, 0, len(labels))
	for _, optionLabel := range labels {
		options = append(options, pep3SameValueRecordFieldOption(optionLabel))
	}
	return pep3MultiRecordField(key, label, options...)
}
