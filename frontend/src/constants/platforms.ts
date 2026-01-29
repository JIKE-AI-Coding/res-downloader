export interface PlatformInfo {
  name: string
  icon: string
  category: string
}

export const PLATFORM_MAPPING: Record<string, PlatformInfo> = {
  'qq.com': { name: '微信视频号', icon: '📱', category: 'wechat' },
  'weixin.qq.com': { name: '微信公众号', icon: '💬', category: 'wechat' },
  'wxapp.com': { name: '微信小程序', icon: '�小程序', category: 'wechat' },
  'douyin.com': { name: '抖音', icon: '🎵', category: 'douyin' },
  'iesdouyin.com': { name: '抖音', icon: '🎵', category: 'douyin' },
  'kuaishou.com': { name: '快手', icon: '🎥', category: 'kuaishou' },
  'chenzhongtech.com': { name: '快手', icon: '🎥', category: 'kuaishou' },
  'bilibili.com': { name: 'B站', icon: '📺', category: 'bilibili' },
  'biligame.com': { name: 'B站游戏', icon: '🎮', category: 'bilibili' },
  'xiaohongshu.com': { name: '小红书', icon: '📕', category: 'xiaohongshu' },
  'xhslink.com': { name: '小红书', icon: '📕', category: 'xiaohongshu' },
  'youtube.com': { name: 'YouTube', icon: '▶️', category: 'youtube' },
  'youtu.be': { name: 'YouTube', icon: '▶️', category: 'youtube' },
  'iqiyi.com': { name: '爱奇艺', icon: '🎬', category: 'video' },
  'youku.com': { name: '优酷', icon: '🎬', category: 'video' },
  'tudou.com': { name: '土豆', icon: '🎬', category: 'video' },
  'v.qq.com': { name: '腾讯视频', icon: '🎬', category: 'video' },
  'mgtv.com': { name: '芒果TV', icon: '🥭', category: 'video' },
  'music.163.com': { name: '网易云音乐', icon: '🎵', category: 'music' },
  'y.qq.com': { name: 'QQ音乐', icon: '🎵', category: 'music' },
  'kugou.com': { name: '酷狗音乐', icon: '🎵', category: 'music' },
  'kuwo.cn': { name: '酷我音乐', icon: '🎵', category: 'music' },
  'default': { name: '其他平台', icon: '🌐', category: 'other' }
}

export const CATEGORY_NAMES: Record<string, { zh: string; en: string }> = {
  'wechat': { zh: '微信生态', en: 'WeChat' },
  'douyin': { zh: '抖音', en: 'Douyin' },
  'kuaishou': { zh: '快手', en: 'Kuaishou' },
  'bilibili': { zh: 'B站', en: 'Bilibili' },
  'xiaohongshu': { zh: '小红书', en: 'Xiaohongshu' },
  'youtube': { zh: 'YouTube', en: 'YouTube' },
  'video': { zh: '视频平台', en: 'Video Platforms' },
  'music': { zh: '音乐平台', en: 'Music Platforms' },
  'other': { zh: '其他平台', en: 'Other Platforms' }
}

/**
 * Get platform info from domain
 * @param domain - Domain string (e.g., "qq.com", "video.qq.com")
 * @returns PlatformInfo object
 */
export function getPlatformInfo(domain: string): PlatformInfo {
  if (!domain) return PLATFORM_MAPPING['default']

  // Direct match
  if (PLATFORM_MAPPING[domain]) {
    return PLATFORM_MAPPING[domain]
  }

  // Subdomain match (e.g., "video.qq.com" -> matches "qq.com")
  for (const [key, value] of Object.entries(PLATFORM_MAPPING)) {
    if (key !== 'default' && domain.endsWith('.' + key)) {
      return value
    }
  }

  return PLATFORM_MAPPING['default']
}

/**
 * Get category name in current locale
 * @param category - Category key
 * @param locale - Current locale ('zh' or 'en')
 * @returns Localized category name
 */
export function getCategoryName(category: string, locale: string = 'zh'): string {
  const cat = CATEGORY_NAMES[category]
  if (cat && locale in cat) {
    return cat[locale as keyof typeof cat]
  }
  return cat?.zh || category
}
