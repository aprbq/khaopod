import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { translations } from './translations'
import type { Lang, TranslationKey } from './translations'

const STORAGE_KEY = 'lang'

type TFunc = (key: TranslationKey, params?: Record<string, string | number>) => string

interface LanguageContextValue {
  lang: Lang
  setLang: (lang: Lang) => void
  t: TFunc
}

const LanguageContext = createContext<LanguageContextValue | null>(null)

function detectInitial(): Lang {
  const saved = localStorage.getItem(STORAGE_KEY)
  return saved === 'en' || saved === 'th' ? saved : 'th'
}

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLang] = useState<Lang>(detectInitial)

  useEffect(() => {
    document.documentElement.lang = lang
    localStorage.setItem(STORAGE_KEY, lang)
  }, [lang])

  const t = useCallback<TFunc>(
    (key, params) => {
      let text = translations[lang][key] ?? key
      if (params) {
        for (const [k, v] of Object.entries(params)) {
          text = text.replace(`{${k}}`, String(v))
        }
      }
      return text
    },
    [lang],
  )

  const value = useMemo<LanguageContextValue>(() => ({ lang, setLang, t }), [lang, t])

  return <LanguageContext value={value}>{children}</LanguageContext>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useLang(): LanguageContextValue {
  const ctx = useContext(LanguageContext)
  if (!ctx) throw new Error('useLang ต้องอยู่ภายใต้ <LanguageProvider>')
  return ctx
}
