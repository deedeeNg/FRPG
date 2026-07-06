import { createContext, useContext, useEffect, useState } from 'react'

// ---------------------------------------------------------------------------
// Language preference + string lookup. Add keys to `strings` as pages get
// translated; `t(key)` falls back to English, then to the key itself.
// ---------------------------------------------------------------------------
export const LANGUAGES = [
  { code: 'en', label: 'English' },
  { code: 'vi', label: 'Tiếng Việt' },
]

const strings = {
  en: {
    // Settings
    'settings.title': 'Settings',
    'settings.subtitle': 'Manage how FRPG looks and speaks.',
    'settings.appearance': 'Appearance',
    'settings.appearance.hint': 'Choose a light or dark background, or follow your system.',
    'settings.theme.light': 'Light',
    'settings.theme.dark': 'Dark',
    'settings.theme.system': 'System',
    'settings.language': 'Language',
    'settings.language.hint': 'Choose the language for the interface.',

    // Brand
    'brand.tagline': 'Learn French. Level up.',

    // Auth
    'auth.welcome': 'Welcome back',
    'auth.welcome.sub': 'Continue your quest to fluency.',
    'auth.create': 'Create your account',
    'auth.create.sub': 'Start your quest to fluency.',
    'auth.login': 'Log in',
    'auth.loggingIn': 'Logging in…',
    'auth.createAccount': 'Create account',
    'auth.creatingAccount': 'Creating account…',
    'auth.orContinue': 'or continue with',
    'auth.orSignup': 'or sign up with',
    'auth.newAdventurer': 'New adventurer?',
    'auth.createOne': 'Create an account',
    'auth.haveAccount': 'Already have an account?',
    'auth.tab.login': 'Login',
    'auth.tab.signup': 'Sign up',

    // Form fields
    'field.email': 'Email',
    'field.password': 'Password',
    'field.confirmPassword': 'Confirm password',

    // Providers
    'provider.google': 'Continue with Google',
    'provider.facebook': 'Continue with Facebook',

    // Errors
    'error.generic': 'Something went wrong',
    'error.passwordShort': 'Password must be at least 8 characters.',
    'error.passwordMismatch': 'Passwords do not match.',

    // Landing
    'landing.title': 'Master French, one quest at a time.',
    'landing.subtitle': 'Pick a skill to begin your daily adventure. Choose your path.',
    'landing.hero': 'hero illustration',

    // Skills
    'skill.speaking': 'Speaking',
    'skill.listening': 'Listening',
    'skill.reading': 'Reading',
    'skill.writing': 'Writing',
    'skill.grammar': 'Grammar',
    'skill.vocabulary': 'Vocabulary',

    // Character attribute hexagon
    'stats.title': 'Character attributes',
    'attr.listening': 'Listening',
    'attr.reading': 'Reading',
    'attr.writing': 'Writing',
    'attr.speaking': 'Speaking',
    'attr.vocabulary': 'Vocabulary',
    'attr.grammar': 'Grammar',

    // Sidebar / nav
    'nav.home': 'Home',
    'nav.map': 'Map',
    'nav.learning': 'Learning',
    'nav.settings': 'Settings',
    'nav.logout': 'Log Out',
    'nav.platform': 'Platform',
    'nav.mainMenu': 'Main menu',

    // Profile
    'profile.title': 'Profile',
    'profile.logout': 'Log out',

    // Logout dialog
    'logout.title': 'Log out?',
    'logout.desc': "You'll be signed out of FRPG and need to log in again to continue.",
    'logout.action': 'Log out',

    // Common
    'common.connecting': 'Connecting…',
    'common.loading': 'Loading…',
    'common.cancel': 'Cancel',
  },
  vi: {
    // Settings
    'settings.title': 'Cài đặt',
    'settings.subtitle': 'Quản lý giao diện và ngôn ngữ của FRPG.',
    'settings.appearance': 'Giao diện',
    'settings.appearance.hint': 'Chọn nền sáng hoặc tối, hoặc theo hệ thống.',
    'settings.theme.light': 'Sáng',
    'settings.theme.dark': 'Tối',
    'settings.theme.system': 'Hệ thống',
    'settings.language': 'Ngôn ngữ',
    'settings.language.hint': 'Chọn ngôn ngữ cho giao diện.',

    // Brand
    'brand.tagline': 'Học tiếng Pháp. Lên cấp.',

    // Auth
    'auth.welcome': 'Chào mừng trở lại',
    'auth.welcome.sub': 'Tiếp tục hành trình chinh phục tiếng Pháp.',
    'auth.create': 'Tạo tài khoản',
    'auth.create.sub': 'Bắt đầu hành trình chinh phục tiếng Pháp.',
    'auth.login': 'Đăng nhập',
    'auth.loggingIn': 'Đang đăng nhập…',
    'auth.createAccount': 'Tạo tài khoản',
    'auth.creatingAccount': 'Đang tạo tài khoản…',
    'auth.orContinue': 'hoặc tiếp tục với',
    'auth.orSignup': 'hoặc đăng ký với',
    'auth.newAdventurer': 'Người mới?',
    'auth.createOne': 'Tạo tài khoản',
    'auth.haveAccount': 'Đã có tài khoản?',
    'auth.tab.login': 'Đăng nhập',
    'auth.tab.signup': 'Đăng ký',

    // Form fields
    'field.email': 'Email',
    'field.password': 'Mật khẩu',
    'field.confirmPassword': 'Xác nhận mật khẩu',

    // Providers
    'provider.google': 'Tiếp tục với Google',
    'provider.facebook': 'Tiếp tục với Facebook',

    // Errors
    'error.generic': 'Đã xảy ra lỗi',
    'error.passwordShort': 'Mật khẩu phải có ít nhất 8 ký tự.',
    'error.passwordMismatch': 'Mật khẩu không khớp.',

    // Landing
    'landing.title': 'Chinh phục tiếng Pháp, từng nhiệm vụ một.',
    'landing.subtitle': 'Chọn một kỹ năng để bắt đầu cuộc phiêu lưu hằng ngày. Chọn con đường của bạn.',
    'landing.hero': 'hình minh họa',

    // Skills
    'skill.speaking': 'Nói',
    'skill.listening': 'Nghe',
    'skill.reading': 'Đọc',
    'skill.writing': 'Viết',
    'skill.grammar': 'Ngữ pháp',
    'skill.vocabulary': 'Từ vựng',

    // Character attribute hexagon
    'stats.title': 'Thuộc tính nhân vật',
    'attr.listening': 'Nghe',
    'attr.reading': 'Đọc',
    'attr.writing': 'Viết',
    'attr.speaking': 'Nói',
    'attr.vocabulary': 'Từ vựng',
    'attr.grammar': 'Ngữ pháp',

    // Sidebar / nav
    'nav.home': 'Trang chủ',
    'nav.map': 'Bản đồ',
    'nav.learning': 'Học tập',
    'nav.settings': 'Cài đặt',
    'nav.logout': 'Đăng xuất',
    'nav.platform': 'Nền tảng',
    'nav.mainMenu': 'Menu chính',

    // Profile
    'profile.title': 'Hồ sơ',
    'profile.logout': 'Đăng xuất',

    // Logout dialog
    'logout.title': 'Đăng xuất?',
    'logout.desc': 'Bạn sẽ bị đăng xuất khỏi FRPG và cần đăng nhập lại để tiếp tục.',
    'logout.action': 'Đăng xuất',

    // Common
    'common.connecting': 'Đang kết nối…',
    'common.loading': 'Đang tải…',
    'common.cancel': 'Hủy',
  },
}

const STORAGE_KEY = 'frpg-lang'

const LanguageContext = createContext({ lang: 'en', setLang: () => {}, t: (k) => k })

export function LanguageProvider({ children, defaultLang = 'en' }) {
  const [lang, setLang] = useState(() => {
    if (typeof window === 'undefined') return defaultLang
    const saved = window.localStorage.getItem(STORAGE_KEY)
    return LANGUAGES.some((l) => l.code === saved) ? saved : defaultLang
  })

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, lang)
  }, [lang])

  const t = (key) => (strings[lang] && strings[lang][key]) || strings.en[key] || key

  return <LanguageContext.Provider value={{ lang, setLang, t }}>{children}</LanguageContext.Provider>
}

export function useLanguage() {
  return useContext(LanguageContext)
}
