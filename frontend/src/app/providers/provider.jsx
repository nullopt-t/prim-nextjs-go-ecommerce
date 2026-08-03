import { ThemeProvider } from '@/context/theme'
import { BrowserRouter } from 'react-router-dom'

function Provider({ children }) {

    return (
        <BrowserRouter>
            <ThemeProvider>
                {children}
            </ThemeProvider>
        </BrowserRouter>
    )
}

export default Provider;