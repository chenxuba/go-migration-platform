export interface PEP3StaticQuestionBankItem {
  itemNo: number
  domainCode: string
  domainName: string
  title: string
  materials: string
  method: string
  scoring: string
  sourcePages: number[]
}

export interface PEP3StaticQuestionBankPage {
  key: 'P1' | 'P3'
  title: string
  subtitle: string
  layout: string
  itemRange: string
  items: PEP3StaticQuestionBankItem[]
}

export const pep3StaticQuestionBankPages: PEP3StaticQuestionBankPage[] = [
  {
    key: 'P1',
    title: 'P1 静态题库',
    subtitle: '题目 1-14，覆盖小肌肉、模仿、认知与非语言行为观察。',
    layout: '基础操作与早期反应观察布局',
    itemRange: '1-14',
    items: [
      { itemNo: 1, domainCode: 'FM', domainName: '小肌肉', title: '（1）旋开瓶盖', materials: '肥皂泡液', method: '将瓶放在桌上并交给儿童，以手势示意旋开瓶盖；必要时示范1次。', scoring: '2=能自行旋开；1=有旋动动作但未完成；0=未尝试或未作出所需动作。', sourcePages: [4] },
      { itemNo: 2, domainCode: 'FM', domainName: '小肌肉', title: '（2）吹肥皂泡', materials: '肥皂泡液、肥皂泡棒', method: '测试员示范吹泡泡，再把肥皂泡棒交给儿童并要求其吹泡泡。', scoring: '2=成功吹出；1=有吹泡动作但未成功；0=无成功也无所需动作。', sourcePages: [4] },
      { itemNo: 3, domainCode: 'FM', domainName: '小肌肉', title: '（3）目光追视', materials: '同上', method: '在第2项活动中观察儿童是否追视肥皂泡移动。', scoring: '2=明显追视；1=偶有短暂注视；0=未追视或未尝试。', sourcePages: [4] },
      { itemNo: 4, domainCode: 'FM', domainName: '小肌肉', title: '（4）目光追视跨越中线', materials: '肥皂泡棒、肥皂泡或儿童感兴趣物件', method: '将物件在儿童面前由左至右跨越中线移动，观察追视范围。', scoring: '2=持续追视动作；1=追视至中线或稍越中线；0=未追视至中线。', sourcePages: [4] },
      { itemNo: 5, domainCode: 'CMB', domainName: '行为特征-非语言', title: '（5）检视触觉块', materials: '不同触觉块3块', method: '将触觉块放在儿童面前并保持沉默，观察抓、尝、嗅、拒绝接触等反应。', scoring: '2=反应适当；1=兴趣不寻常或不感兴趣；0=出现不恰当或过度感官反应。', sourcePages: [5] },
      { itemNo: 6, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（6）使用万花筒', materials: '万花筒', method: '示范观看和转动万花筒底盘，然后交给儿童并示意使用。', scoring: '2=观看并转动；1=只完成其中一项；0=未尝试。', sourcePages: [5] },
      { itemNo: 7, domainCode: 'FM', domainName: '小肌肉', title: '（7）表现出能够使用惯用眼', materials: '同上', method: '在第6项中观察儿童是否稳定使用同一眼睛观看万花筒，可进行第2次观察。', scoring: '2=持续使用同一眼；1=大部分时间同一眼；0=双眼观看或两次均未观看。', sourcePages: [5] },
      { itemNo: 8, domainCode: 'CVP', domainName: '认知（语言/语前）', title: '（8）转向手摇铃声', materials: '手摇铃', method: '在儿童看不见处摇铃，观察其对声音的反应与转向。', scoring: '2=清楚反应并正确转向声源；1=有听到表现但方向不正确；0=无反应。', sourcePages: [5] },
      { itemNo: 9, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（9）模仿按动响铃2次', materials: '响铃', method: '示范连续按响2次并要求儿童模仿，必要时再次示范。', scoring: '2=按响2次；1=按响1次或多于2次；0=再次示范后仍未尝试。', sourcePages: [6] },
      { itemNo: 10, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（10）手指插入胶泥并做出凹位', materials: '胶泥', method: '示范将手指插入胶泥球做出凹位，并示意儿童模仿。', scoring: '2=能做出明显凹位；1=插入但凹位不明显；0=未尝试。', sourcePages: [6] },
      { itemNo: 11, domainCode: 'FM', domainName: '小肌肉', title: '（11）抓握竹棒', materials: '胶泥球1个、竹棒6枝', method: '把胶泥压成薄饼并插竹棒，示意儿童插入或拔出竹棒。', scoring: '2=以前二指/指侧/前三指操作至少1枝；1=用整掌或多于三指；0=未能操作。', sourcePages: [6] },
      { itemNo: 12, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（12）听生日歌假装吹蜡烛', materials: '同上', method: '竹棒插入后唱生日歌并假装吹蜡烛，观察儿童是否聆听和模仿。', scoring: '2=聆听并假装吹；1=聆听但未吹；0=没有留意唱歌。', sourcePages: [6] },
      { itemNo: 13, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（13）享受音乐', materials: '胶泥', method: '唱生日歌时观察儿童对音乐题材或活动的兴趣与反应。', scoring: '2=开心并留心或跟唱；1=有留意但未模仿；0=对音乐没有兴趣。', sourcePages: [6] },
      { itemNo: 14, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（14）搓胶泥条', materials: '胶泥', method: '将胶泥分成2份，示范搓成长条后要求儿童模仿。', scoring: '2=能搓成长条；1=尝试或有类似动作；0=未能做出搓的动作。', sourcePages: [6] },
    ],
  },
  {
    key: 'P3',
    title: 'P3 静态题库',
    subtitle: '题目 15-27，覆盖手偶互动、语言理解/表达、拼板与跨越中线动作。',
    layout: '手偶互动与拼板任务布局',
    itemRange: '15-27',
    items: [
      { itemNo: 15, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（15）操控手偶', materials: '猫或狗手偶', method: '测试员戴上手偶模拟互动，随后交给儿童并要求其操控手偶。', scoring: '2=能自行戴上并模仿操控；1=戴上但未操控；0=未能或未尝试戴上。', sourcePages: [7] },
      { itemNo: 16, domainCode: 'VMI', domainName: '模仿（视觉/动作）', title: '（16）利用手偶模仿日常动作', materials: '手偶、匙子、杯子、牙刷、纸巾', method: '测试员示范手偶使用物件，逐项测试喂食、饮水、刷牙、抹鼻。', scoring: '2=模仿至少3项；1=模仿至少1项；0=未能模仿任何动作。', sourcePages: [7] },
      { itemNo: 17, domainCode: 'RL', domainName: '语言理解', title: '（17）指出手偶的3个身体部位', materials: '猫或狗手偶', method: '要求儿童指出手偶的口、眼、鼻、耳等身体部位。', scoring: '2=指出或触摸至少3个；1=至少1个正确或概括指向；0=无正确部位。', sourcePages: [7] },
      { itemNo: 18, domainCode: 'RL', domainName: '语言理解', title: '（18）指出3个自己的身体部位', materials: '无', method: '在没有动作提示下要求儿童指出自己的口、眼、鼻、耳等部位。', scoring: '2=指出或触摸至少3个；1=至少1个正确或概括指向；0=无正确部位。', sourcePages: [8] },
      { itemNo: 19, domainCode: 'EL', domainName: '语言表达', title: '（19）使用两个手偶演出故事', materials: '猫及狗手偶', method: '儿童和测试员各戴一个手偶，鼓励儿童用手偶与测试员演出故事。', scoring: '2=有交往且演出有次序情节；1=只独自演出或交往不足；0=没有演出故事。', sourcePages: [8] },
      { itemNo: 20, domainCode: 'SR', domainName: '社交互动', title: '（20）与手偶作假想游戏', materials: '同上', method: '在第19项同时观察儿童是否表现假想力和创意玩法。', scoring: '2=有假想力和创意；1=能跟随测试员提议；0=未使用假想活动。', sourcePages: [8] },
      { itemNo: 21, domainCode: 'CVP', domainName: '认知（语言/语前）', title: '（21）指示出3块拼块的正确位置', materials: '形状拼板及圆形、正方形、三角形拼块', method: '要求儿童把形状拼块放入对应位置，困难时可示范圆形后再测。', scoring: '2=3个位置正确或相近；1=至少1个正确或相近；0=示范后仍未能放入。', sourcePages: [8] },
      { itemNo: 22, domainCode: 'FM', domainName: '小肌肉', title: '（22）完成形状拼板', materials: '同上', method: '在第21项同时观察儿童完成形状拼板的情况。', scoring: '2=无需示范完成3个；1=至少完成1个或需示范；0=示范后仍未完成。', sourcePages: [9] },
      { itemNo: 23, domainCode: 'EL', domainName: '语言表达', title: '（23）说出形状名称', materials: '圆形、正方形、三角形拼块', method: '逐一指着形状拼块，询问儿童形状名称。', scoring: '2=说出3种形状；1=至少1种正确或同称但有1个正确；0=无正确名称。', sourcePages: [9] },
      { itemNo: 24, domainCode: 'RL', domainName: '语言理解', title: '（24）挑选形状', materials: '同上', method: '要求儿童交出或指出圆形、正方形、三角形。', scoring: '2=全部正确；1=至少1种正确；0=无正确交出或指出。', sourcePages: [9] },
      { itemNo: 25, domainCode: 'CVP', domainName: '认知（语言/语前）', title: '（25）完成4块物件拼板', materials: '雨伞、小鸡、蝴蝶、雪梨物件拼板', method: '逐块给予拼块并要求放入拼板，困难时示范完成后再测。', scoring: '2=无需示范完成4块；1=至少放入1块；0=示范后仍未放入。', sourcePages: [9] },
      { itemNo: 26, domainCode: 'GM', domainName: '大肌肉', title: '（26）跨越中线取拼块', materials: '同上或蜡笔、积木等替代材料', method: '将拼块放在儿童左右两侧，观察其是否跨越中线拿取。', scoring: '2=多于1次跨越中线；1=跨越1次；0=未尝试或用转身代替。', sourcePages: [10] },
      { itemNo: 27, domainCode: 'CVP', domainName: '认知（语言/语前）', title: '（27）指示出3块手套拼块的正确位置', materials: '手套拼板', method: '把手套拼板和拼块放在桌上，要求儿童把拼块放回拼板。', scoring: '2=3块位置正确或相近；1=至少1块正确或相近；0=示范后仍未能放入。', sourcePages: [10] },
    ],
  },
]
