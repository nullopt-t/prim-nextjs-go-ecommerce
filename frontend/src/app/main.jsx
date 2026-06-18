import '@/app/styles/index.css'
import App from '@/app/App'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import Provider from '@/app/provider'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <Provider>
      <App />
    </Provider>
  </StrictMode>,
)
