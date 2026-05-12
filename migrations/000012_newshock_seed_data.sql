-- +goose Up
-- Newshock seed data (sample from news.moneych.top analysis)

-- Use a fixed tenant_id for all seed data
-- Change this to match your actual tenant
-- +goose StatementBegin
DO $$
DECLARE
    v_tenant VARCHAR(36) := 'default';

    -- Theme UUIDs
    t_mideast    VARCHAR(36) := 'a0000000-0000-0000-0000-000000000001';
    t_ai_chip    VARCHAR(36) := 'a0000000-0000-0000-0000-000000000002';
    t_central    VARCHAR(36) := 'a0000000-0000-0000-0000-000000000003';
    t_ev         VARCHAR(36) := 'a0000000-0000-0000-0000-000000000004';
    t_trade      VARCHAR(36) := 'a0000000-0000-0000-0000-000000000005';
    t_defense    VARCHAR(36) := 'a0000000-0000-0000-0000-000000000006';
    t_gold       VARCHAR(36) := 'a0000000-0000-0000-0000-000000000007';
    t_energy     VARCHAR(36) := 'a0000000-0000-0000-0000-000000000008';
    t_earnings   VARCHAR(36) := 'a0000000-0000-0000-0000-000000000009';
    t_consumer   VARCHAR(36) := 'a0000000-0000-0000-0000-000000000010';
    t_realestate VARCHAR(36) := 'a0000000-0000-0000-0000-000000000011';
    t_global     VARCHAR(36) := 'a0000000-0000-0000-0000-000000000012';
    t_mineral    VARCHAR(36) := 'a0000000-0000-0000-0000-000000000013';
    t_regulatory VARCHAR(36) := 'a0000000-0000-0000-0000-000000000014';
    t_capital    VARCHAR(36) := 'a0000000-0000-0000-0000-000000000015';
    t_pharma     VARCHAR(36) := 'a0000000-0000-0000-0000-000000000016';
    t_crypto     VARCHAR(36) := 'a0000000-0000-0000-0000-000000000017';
    t_soe        VARCHAR(36) := 'a0000000-0000-0000-0000-000000000018';
    t_ai_app     VARCHAR(36) := 'a0000000-0000-0000-0000-000000000019';
    t_cpo        VARCHAR(36) := 'a0000000-0000-0000-0000-000000000020';

    -- Ticker UUIDs
    tk_aapl  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000001';
    tk_googl VARCHAR(36) := 'b0000000-0000-0000-0000-000000000002';
    tk_nvda  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000003';
    tk_meta  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000004';
    tk_msft  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000005';
    tk_tsla  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000006';
    tk_amzn  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000007';
    tk_amd   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000008';
    tk_xom   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000009';
    tk_intc  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000010';
    tk_cvx   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000011';
    tk_coin  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000012';
    tk_baba  VARCHAR(36) := 'b0000000-0000-0000-0000-000000000013';
    tk_arm   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000014';
    tk_nvo   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000015';
    tk_tsm   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000016';
    tk_lly   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000017';
    tk_mu    VARCHAR(36) := 'b0000000-0000-0000-0000-000000000018';
    tk_sam   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000019';
    tk_mcd   VARCHAR(36) := 'b0000000-0000-0000-0000-000000000020';

    -- Event UUIDs
    e1 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000001';
    e2 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000002';
    e3 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000003';
    e4 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000004';
    e5 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000005';
    e6 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000006';
    e7 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000007';
    e8 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000008';
    e9 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000009';
    e10 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000010';
    e11 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000011';
    e12 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000012';
    e13 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000013';
    e14 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000014';
    e15 VARCHAR(36) := 'c0000000-0000-0000-0000-000000000015';
BEGIN

-- ========== THEMES ==========
INSERT INTO newshock_themes (id, name, description, category, strength, strength_norm, classification_confidence, ticker_count, event_count, trend, tenant_id) VALUES
(t_mideast, '中东地缘与能源', '科威特遭伊朗导弹攻击、伊朗议长和外长暂被移出美以清除名单、阿联酋ADNOC CEO批评伊朗将霍尔木兹海峡武器化、印度石油公司自2018年以来首次购买伊朗液化石油气、特朗普希望尽快结束伊朗战争，中东冲突升级影响能源供应和全球经济。', 'geopolitical', 5524.3, 100, 0.85, 499, 6397, 'rising', v_tenant),
(t_ai_chip, 'AI算力与半导体芯片', '智谱AI产品涨价、华为开发通用AI芯片、Kimi K2模型发布、OpenAI收购Windsurf失败及谷歌收购其资产、谷歌发布GenAI处理器库、RealSense分拆融资等事件，反映AI算力与半导体领域的持续投入与竞争', 'ai_semi', 3604.6, 97, 0.85, 763, 2062, 'rising', v_tenant),
(t_central, '央行货币政策与通胀', '美国3月CPI符合预期但核心通胀温和，中国PPI时隔41个月首次转正，受中东局势影响大宗商品价格分化，分析师关注PPI与CPI剪刀差及对货币政策的潜在影响。', 'macro_monetary', 2198.5, 94, 0.85, 55, 841, 'rising', v_tenant),
(t_ev, '新能源汽车', '智能电动汽车高层论坛召开，行业聚焦智驾芯片自研、超充安全边界、电池标准化及出海竞争，多位车企高管发表关于技术迭代、成本管控及市场渗透率的观点。', 'supply_chain', 1584.6, 91, 0.85, 333, 963, 'rising', v_tenant),
(t_trade, '地缘贸易与制裁', '美英达成关税贸易协议，美国将英国产汽车关税降至10%，牛肉关税接近零；欧盟提议若谈判失败对950亿欧元美国商品加征关税；中俄签署深化战略协作联合声明。', 'geopolitical', 1229.5, 88, 0.85, 98, 981, 'rising', v_tenant),
(t_defense, '全球国防支出超级周期', '美国国防部要求Anthropic开放AI技术供军方使用，日本民众抗议扩军政策，土耳其F-16坠机，特朗普国情咨文强调安全与主导地位，显示全球国防支出与军事AI应用加速。', 'defense', 948.8, 85, 0.85, 229, 1772, 'rising', v_tenant),
(t_gold, '黄金与贵金属投资', '白银价格大幅上涨，纽约期银突破110美元/盎司，瑞士宝盛看涨白银至125-150美元，上期所调整白银期货交易限额和保证金比例。', 'macro_monetary', 883.9, 82, 0.85, 167, 915, 'rising', v_tenant),
(t_energy, '新能源与储能', '3月ICE英国天然气期货累计上涨62.85%，TTF基准荷兰天然气期货累涨59.51%，欧盟碳排放交易许可期货累涨3.58%', 'energy', 864.2, 79, 0.7, 471, 462, 'rising', v_tenant),
(t_earnings, '财报季与业绩', 'AudioCodes、Avista、Energizer、Precision BioSciences、Energy Transfer等多家公司发布或即将发布季度财报。', 'earnings_event', 792.19, 76, 0.85, 795, 521, 'rising', v_tenant),
(t_consumer, '消费电子与科技巨头', '苹果MacBook Neo生产目标提高至1000万台，使用A18 Pro芯片；京东方今年将向苹果供应3500万块OLED屏；华为MatePad Pro Max即将亮相。', 'exploratory', 759.2, 74, 0.85, 116, 219, 'rising', v_tenant),
(t_realestate, '中国房地产宽松政策', '中共中央、国务院批复《现代化首都都市圈空间协同规划（2023-2035年）》，涉及京津冀城市体系优化和住房保障政策，同时上海启动收购二手住房用于保障性租赁住房工作。', 'macro_monetary', 747.4, 71, 0.85, 197, 532, 'rising', v_tenant),
(t_global, '全球市场与宏观', '欧元兑英镑上涨、MSCI新兴市场货币指数下跌、德国零售销售数据超预期、英国房价指数超预期、A股三大指数集体下跌。', 'macro_monetary', 741.98, 68, 0.85, 475, 1167, 'rising', v_tenant),
(t_mineral, '关键矿产与供应链安全', '华联控股拟收购阿根廷锂矿公司获取Arizaro项目80%权益，永杰新材与奥科宁克签署战略合作协议并收购其中国资产，凸显关键矿产供应链布局。', 'geopolitical', 603.9, 65, 0.85, 362, 717, 'rising', v_tenant),
(t_regulatory, '监管与合规', 'SEC提议允许上市公司提交半年报替代季报，爱尔兰对Meta展开调查，中央网信办整治AI技术滥用，国家药监局支持高端医疗器械创新。', 'regulatory', 592.1, 62, 0.8, 233, 599, 'rising', v_tenant),
(t_capital, '科技巨头资本运作与业务调整', '埃利奥特管理公司在伦敦证交所集团建立头寸并推动改善业绩，英特尔盘前涨超4%因特朗普政府讨论国家持股，联合健康盘前涨超12%因巴菲特建仓。', 'exploratory', 550.29, 59, 0.85, 218, 362, 'rising', v_tenant),
(t_pharma, '创新药行业业绩爆发', '礼来口服减肥药Foundayo获FDA批准，加剧与诺和诺德的竞争，华尔街预测2030年销售额可达180亿美元，显示医药创新进入超级周期。', 'pharma', 546.5, 56, 0.85, 190, 250, 'rising', v_tenant),
(t_crypto, '加密货币监管与稳定币立法', '纽约州总检察长起诉Coinbase和Gemini违反非法赌博法律，同时预测平台Kalshi和PolyMarket推进加密衍生品交易，监管与机构采纳并存。', 'regulatory', 442.5, 53, 0.85, 71, 428, 'rising', v_tenant),
(t_soe, '中国国企改革与资本运作', '中国神华拟1335.98亿元收购国家能源集团多项资产，国资委主任张玉卓强调央企聚焦新能源、AI等战略领域。', 'regulatory', 364.3, 50, 0.85, 109, 110, 'rising', v_tenant),
(t_ai_app, 'AI应用与软件生态更新', 'DeepL因AI冲击计划裁员约250人，CEO称AI带来巨大结构性转变，反映AI对传统软件行业的颠覆加速。', 'ai_semi', 734, 73, 0.85, 145, 380, 'rising', v_tenant),
(t_cpo, 'CPO与光通信算力硬件轮动', '二季度以来融资客净买入光通信龙头，中际旭创、东山精密、天孚通信居前，800G光模块放量。', 'ai_semi', 612.5, 67, 0.85, 88, 245, 'rising', v_tenant);

-- ========== TICKERS ==========
INSERT INTO newshock_tickers (id, symbol, name, market, hot_score, mention_count, tenant_id) VALUES
(tk_aapl, 'AAPL', '苹果', 'us', 473, 64, v_tenant),
(tk_googl, 'GOOGL', '谷歌', 'us', 378.5, 49, v_tenant),
(tk_nvda, 'NVDA', '英伟达', 'us', 329, 40, v_tenant),
(tk_meta, 'META', 'Meta', 'us', 300.5, 40, v_tenant),
(tk_msft, 'MSFT', '微软', 'us', 279.5, 37, v_tenant),
(tk_tsla, 'TSLA', '特斯拉', 'us', 231, 33, v_tenant),
(tk_amzn, 'AMZN', '亚马逊', 'us', 203, 25, v_tenant),
(tk_amd, 'AMD', '超威半导体', 'us', 169.5, 21, v_tenant),
(tk_xom, 'XOM', '埃克森美孚', 'us', 154, 17, v_tenant),
(tk_intc, 'INTC', '英特尔', 'us', 133, 17, v_tenant),
(tk_cvx, 'CVX', '雪佛龙', 'us', 127, 14, v_tenant),
(tk_coin, 'COIN', 'Coinbase Global', 'us', 103.5, 15, v_tenant),
(tk_baba, 'BABA', '阿里巴巴', 'us', 93.5, 13, v_tenant),
(tk_arm, 'ARM', 'ARM', 'us', 93, 12, v_tenant),
(tk_nvo, 'NVO', '诺和诺德', 'us', 82, 11, v_tenant),
(tk_tsm, 'TSM', '台积电', 'us', 76.5, 9, v_tenant),
(tk_lly, 'LLY', '礼来', 'us', 75.5, 10, v_tenant),
(tk_mu, 'MU', '美光科技', 'us', 72.5, 10, v_tenant),
(tk_sam, '005930.KS', '三星电子', 'kr', 64, 8, v_tenant),
(tk_mcd, 'MCD', '麦当劳', 'us', 55, 6, v_tenant);

-- ========== EVENTS ==========
INSERT INTO newshock_events (id, title, summary, channel, importance, theme_id, theme_name, event_time, tenant_id) VALUES
(e1, '瑞杰金融上调Arm目标价', '瑞杰金融将Arm目标价从166美元上调至244美元，反映AI芯片需求乐观。', '金融快报', 2, t_capital, '科技巨头资本运作与业务调整', '2026-05-07 22:57:01', v_tenant),
(e2, 'DeepL计划裁员25%', '翻译工具DeepL因AI冲击计划裁员约250人，CEO称AI带来巨大结构性转变。', '金融快报', 3, t_ai_app, 'AI应用与软件生态更新', '2026-05-07 22:57:01', v_tenant),
(e3, 'SK海力士一季度净利润暴增五倍', 'AI芯片需求带动SK海力士一季度净利润同比增五倍，人均分红43万美元，DRAM价格预计涨超70%。', '驻华外电', 5, t_ai_chip, 'AI算力与半导体芯片', '2026-05-07 22:57:01', v_tenant),
(e4, '美国制裁伊拉克石油部副部长', '美国财政部因伊拉克支持伊朗，制裁其石油部副部长及相关民兵组织，加剧中东能源紧张。', '金融快报', 4, t_mideast, '中东地缘与能源', '2026-05-07 22:57:01', v_tenant),
(e5, '深圳首条人形机器人中试产线投产', '深圳首条人形机器人中试产线投产，产业产值超2400亿元，产业链93%可在本地配齐，突破量产瓶颈。', '金融快报', 4, t_ai_chip, 'AI算力与半导体芯片', '2026-05-07 22:57:01', v_tenant),
(e6, '美联储哈玛克：基准预期是较长一段时间内维持利率不变', '克利夫兰联储主席哈玛克表示，美联储政策声明暗示降息可能误导，基准预期是利率将在相当长一段时间内维持不变。', '金融快报', 4, t_central, '央行货币政策与通胀', '2026-05-07 22:57:01', v_tenant),
(e7, '麦当劳第一季度营收超预期', '麦当劳第一季度营收65.2亿美元，每股收益2.83美元，均超预期，全球可比销售增长3.8%。', '金融快报', 3, t_earnings, '财报季与业绩', '2026-05-07 22:57:01', v_tenant),
(e8, '七部门发布《医药代表管理办法》', '国家药监局等七部门修订《医药代表管理办法》，严格防范商业贿赂，对违规行为联合惩戒。', '金融快报', 4, t_regulatory, '监管与合规', '2026-05-07 22:57:01', v_tenant),
(e9, '纳指创盘中历史新高', '纳斯达克指数盘中创下历史新高，科技股领涨，AI概念股持续走强。', '金融快报', 4, t_ai_chip, 'AI算力与半导体芯片', '2026-05-07 22:57:01', v_tenant),
(e10, 'A股AI产业链爆发', 'A股AI产业链全面爆发，算力、光模块、AI应用等板块涨幅居前，多只个股涨停。', '金融快报', 4, t_ai_chip, 'AI算力与半导体芯片', '2026-05-07 22:57:01', v_tenant),
(e11, '欧盟有望本月敲定与美国的贸易协定', '欧盟资深立法者朗格表示，欧盟与美国贸易协定有望在5月12日或19日敲定。', '金融快报', 4, t_trade, '地缘贸易与制裁', '2026-05-07 22:52:59', v_tenant),
(e12, '谷歌推出Fitbit Air可穿戴设备', '谷歌推出全新Fitbit Air可穿戴设备，主打健康监测与AI功能整合。', '金融快报', 3, t_consumer, '消费电子与科技巨头', '2026-05-07 22:57:01', v_tenant),
(e13, '融资客青睐光通信板块', '二季度以来融资客净买入光通信龙头，中际旭创、东山精密、天孚通信居前，800G光模块放量。', '金融快报', 3, t_cpo, 'CPO与光通信算力硬件轮动', '2026-05-07 22:57:01', v_tenant),
(e14, '礼来口服减肥药获FDA批准', '礼来口服减肥药Foundayo获FDA批准，加剧与诺和诺德的竞争，华尔街预测2030年销售额可达180亿美元。', '金融快报', 5, t_pharma, '创新药行业业绩爆发', '2026-05-07 22:57:01', v_tenant),
(e15, '美股AI应用软件股因超预期业绩盘前走强', '美股AI应用软件股因超预期业绩盘前走强，多家SaaS公司上调全年指引。', '金融快报', 4, t_ai_app, 'AI应用与软件生态更新', '2026-05-07 22:57:01', v_tenant);

-- ========== THEME-TICKER RELATIONSHIPS ==========
INSERT INTO newshock_theme_tickers (theme_id, ticker_id) VALUES
-- AI算力 → NVDA, AMD, TSM, MU, ARM, INTC
(t_ai_chip, tk_nvda), (t_ai_chip, tk_amd), (t_ai_chip, tk_tsm), (t_ai_chip, tk_mu), (t_ai_chip, tk_arm), (t_ai_chip, tk_intc),
-- 消费电子 → AAPL, GOOGL, META, MSFT
(t_consumer, tk_aapl), (t_consumer, tk_googl), (t_consumer, tk_meta), (t_consumer, tk_msft),
-- 中东能源 → XOM, CVX
(t_mideast, tk_xom), (t_mideast, tk_cvx),
-- 新能源 → TSLA, BABA
(t_ev, tk_tsla), (t_ev, tk_baba),
-- 财报 → AAPL, GOOGL, MSFT, AMZN, MCD
(t_earnings, tk_aapl), (t_earnings, tk_googl), (t_earnings, tk_msft), (t_earnings, tk_amzn), (t_earnings, tk_mcd),
-- 加密 → COIN
(t_crypto, tk_coin),
-- 医药 → LLY, NVO
(t_pharma, tk_lly), (t_pharma, tk_nvo),
-- 三星 → 存储/AI
(t_ai_chip, tk_sam),
-- 科技资本 → INTC, ARM
(t_capital, tk_intc), (t_capital, tk_arm),
-- CPO → 光通信 (用 NVDA 代表算力需求)
(t_cpo, tk_nvda);

-- ========== EVENT-TICKER RELATIONSHIPS ==========
INSERT INTO newshock_event_tickers (event_id, ticker_id) VALUES
(e1, tk_arm),       -- 瑞杰金融上调Arm目标价
(e3, tk_sam),       -- SK海力士 (三星同赛道)
(e5, tk_nvda),      -- 人形机器人 → AI算力
(e7, tk_mcd),       -- 麦当劳财报
(e9, tk_nvda),      -- 纳指新高 → NVDA
(e9, tk_aapl),      -- 纳指新高 → AAPL
(e10, tk_nvda),     -- A股AI爆发 → NVDA
(e10, tk_amd),      -- A股AI爆发 → AMD
(e12, tk_aapl),     -- Fitbit → 苹果竞品
(e14, tk_lly),      -- 礼来减肥药
(e14, tk_nvo),      -- 诺和诺德竞品
(e15, tk_msft),     -- AI软件股 → MSFT
(e15, tk_meta);     -- AI软件股 → META

-- ========== REGIME ==========
INSERT INTO newshock_regime (id, regime_type, confidence, summary, tenant_id) VALUES
('d0000000-0000-0000-0000-000000000001', 'risk_on', 0.78,
 'AI产业链持续领涨，半导体与算力需求强劲。中东地缘风险虽高但市场已部分定价，资金继续涌入科技成长股。美联储维持利率不变的预期明确，流动性环境友好。',
 v_tenant);

END $$;
-- +goose StatementEnd

-- +goose Down
DELETE FROM newshock_event_tickers;
DELETE FROM newshock_theme_tickers;
DELETE FROM newshock_events WHERE tenant_id = 'default';
DELETE FROM newshock_tickers WHERE tenant_id = 'default';
DELETE FROM newshock_themes WHERE tenant_id = 'default';
DELETE FROM newshock_regime WHERE tenant_id = 'default';
