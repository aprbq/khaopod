import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/features/auth/AuthContext'
import { LanguageProvider } from '@/i18n/LanguageContext'
import { queryClient } from '@/lib/queryClient'
import { App } from './App'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <LanguageProvider>
        <AuthProvider>
          <App />
        </AuthProvider>
      </LanguageProvider>
    </QueryClientProvider>
  </StrictMode>,
)
