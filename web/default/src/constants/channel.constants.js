export const CHANNEL_OPTIONS = [
  { key: 1, text: 'OpenAI', value: 1, color: 'green' },
  { key: 14, text: 'Anthropic Claude', value: 14, color: 'black' }, // ok
  { key: 33, text: 'AWS', value: 33, color: 'black' }, // ok
  { key: 3, text: 'Azure OpenAI', value: 3, color: 'olive' }, //ok
  { key: 11, text: 'Google PaLM2', value: 11, color: 'orange' }, // ok
  { key: 24, text: 'Google Gemini', value: 24, color: 'orange' }, // ok
  { key: 28, text: 'Mistral AI', value: 28, color: 'orange' }, // ok
  { key: 41, text: 'Novita', value: 41, color: 'purple' }, //ok
  { key: 40, text: '字节跳动豆包', value: 40, color: 'blue' }, //ok
  { key: 15, text: '百度文心千帆', value: 15, color: 'blue' }, //ok baidu
  { key: 17, text: '阿里通义千问', value: 17, color: 'orange' }, // ali
  { key: 18, text: '讯飞星火认知', value: 18, color: 'blue' }, // xunfei ok
  {
    key: 61, // 50
    text: 'OpenAI 兼容',
    value: 61, // 50
    color: 'olive',
    description: 'OpenAI 兼容渠道，支持设置 Base URL',
  },
  //{key: 14, text: 'Anthropic', value: 14, color: 'black'},
  //{ key: 33, text: 'AWS', value: 33, color: 'black' },
  //{key: 3, text: 'Azure', value: 3, color: 'olive'},
  //{key: 11, text: 'PaLM2', value: 11, color: 'orange'},
  //{key: 24, text: 'Gemini', value: 24, color: 'orange'},
  {
    key: 62, // 51
    text: 'Gemini (OpenAI)',
    value: 62, // 51
    color: 'orange',
    description: 'Gemini OpenAI 兼容格式',
  },
  //{ key: 28, text: 'Mistral AI', value: 28, color: 'orange' },
  //{ key: 41, text: 'Novita', value: 41, color: 'purple' },
//   {
//     key: 40,
//     text: '字节火山引擎',
//     value: 40,
//     color: 'blue',
//     description: '原字节跳动豆包',
//   },
//   {
//     key: 15,
//     text: '百度文心千帆',
//     value: 15,
//     color: 'blue',
//     tip: '请前往<a href="https://console.bce.baidu.com/qianfan/ais/console/applicationConsole/application/v1" target="_blank">此处</a>获取 AK（API Key）以及 SK（Secret Key），注意，V2 版本接口请使用 <strong>百度文心千帆 V2 </strong>渠道类型',
//   },
  {
    key: 58, // 47
    text: '百度文心千帆 V2',
    value: 58, // 47
    color: 'blue',
    tip: '请前往<a href="https://console.bce.baidu.com/iam/#/iam/apikey/list" target="_blank">此处</a>获取 API Key，注意本渠道仅支持<a target="_blank" href="https://cloud.baidu.com/doc/WENXINWORKSHOP/s/em4tsqo3v">推理服务 V2</a>相关模型',
  },
  //{
  //  key: 17,
  //  text: '阿里通义千问',
  //  value: 17,
  //  color: 'orange',
  //  tip: '如需使用阿里云百炼，请使用<strong>阿里云百炼</strong>渠道',
  //},
  { key: 60, text: '阿里云百炼', value: 60, color: 'orange' }, // 49
  //{
  //  key: 18,
  //  text: '讯飞星火认知',
  //  value: 18,
  //  color: 'blue',
  //  tip: '本渠道基于讯飞 WebSocket 版本 API，如需 HTTP 版本，请使用<strong>讯飞星火认知 V2</strong>渠道',
  //},
  {
    key: 59, // 48
    text: '讯飞星火认知 V2',
    value: 59,
    color: 'blue',
    tip: 'HTTP 版本的讯飞接口，前往<a href="https://console.xfyun.cn/services/cbm" target="_blank">此处</a>获取 HTTP 服务接口认证密钥',
  },
  { key: 16, text: '智谱 ChatGLM', value: 16, color: 'violet' },
  { key: 19, text: '360 智脑', value: 19, color: 'blue' },
  { key: 25, text: 'Moonshot AI', value: 25, color: 'black' },
  { key: 23, text: '腾讯混元', value: 23, color: 'teal' },
  { key: 26, text: '百川大模型', value: 26, color: 'orange' },
  { key: 27, text: 'MiniMax', value: 27, color: 'red' },
  { key: 29, text: 'Groq', value: 29, color: 'orange' },
  { key: 30, text: 'Ollama', value: 30, color: 'black' },
  { key: 31, text: '零一万物', value: 31, color: 'green' },
  { key: 32, text: '阶跃星辰', value: 32, color: 'blue' },
  { key: 34, text: 'Coze', value: 34, color: 'blue' },
  { key: 35, text: 'Cohere', value: 35, color: 'blue' },
  { key: 36, text: 'DeepSeek', value: 36, color: 'black' },
  { key: 37, text: 'Cloudflare', value: 37, color: 'orange' },
  { key: 38, text: 'DeepL', value: 38, color: 'black' },
  { key: 39, text: 'together.ai', value: 39, color: 'blue' },
  { key: 42, text: 'VertexAI', value: 42, color: 'blue' },
  { key: 43, text: 'Proxy', value: 43, color: 'blue' },
  { key: 44, text: 'SiliconFlow', value: 44, color: 'blue' },
  { key: 45, text: 'ImaginePro', value: 45, color: 'blue' }, // ok
  { key: 46, text: 'Friday', value: 46, color: 'blue' }, // ok
  { key: 47, text: 'KlingAI', value: 47, color: 'blue' }, // ok
  { key: 48, text: 'BFL', value: 48, color: 'blue' }, // ok
  { key: 49, text: 'panda', value: 49, color: 'green' }, // ok
  { key: 50, text: 'newaliyn', value: 50, color: 'green' }, //ok
  { key: 51, text: 'ai302', value: 51, color: 'green' }, // ok
  { key: 52, text: 'midjourney', value: 52, color: 'blue' }, // ok
  { key: 53, text: 'ailab', value: 53, color: 'blue' }, //ok
  { key: 54, text: 'aiguoguo', value: 54, color: 'green' }, // ok
  { key: 55, text: 'baidu2', value: 55, color: 'blue' }, //ok
  { key: 8, text: '自定义渠道', value: 8, color: 'pink' }, //ok
  { key: 22, text: '知识库：FastGPT', value: 22, color: 'blue' }, // ok
  { key: 21, text: '知识库：AI Proxy', value: 21, color: 'purple' }, //ok
  { key: 20, text: '代理：OpenRouter', value: 20, color: 'black' }, //ok
  { key: 56, text: 'xAI', value: 56, color: 'blue' }, // 45
  { key: 57, text: 'Replicate', value: 57, color: 'blue' }, //46 
  //{
  //  key: 8,
  //  text: '自定义渠道',
  //  value: 8,
  //  color: 'pink',
  //  tip: '不推荐使用，请使用 <strong>OpenAI 兼容</strong>渠道类型。注意，这里所需要填入的代理地址仅会在实际请求时替换域名部分，如果你想填入 OpenAI SDK 中所要求的 Base URL，请使用 OpenAI 兼容渠道类型',
  //  description: '不推荐使用，请使用 OpenAI 兼容渠道类型',
  //},
  //{ key: 22, text: '知识库：FastGPT', value: 22, color: 'blue' },
  //{ key: 21, text: '知识库：AI Proxy', value: 21, color: 'purple' },
  //{ key: 20, text: 'OpenRouter', value: 20, color: 'black' },
  { key: 2, text: '代理：API2D', value: 2, color: 'blue' }, // ok
  { key: 5, text: '代理：OpenAI-SB', value: 5, color: 'brown' }, //ok
  { key: 7, text: '代理：OhMyGPT', value: 7, color: 'purple' }, //ok
  { key: 10, text: '代理：AI Proxy', value: 10, color: 'purple' }, //ok
  { key: 4, text: '代理：CloseAI', value: 4, color: 'teal' }, // ok
  { key: 6, text: '代理：OpenAI Max', value: 6, color: 'violet' }, // ok
  { key: 9, text: '代理：AI.LS', value: 9, color: 'yellow' }, // ok
  { key: 12, text: '代理：API2GPT', value: 12, color: 'blue' }, //ok
  { key: 13, text: '代理：AIGC2D', value: 13, color: 'purple' }, //ok
];
