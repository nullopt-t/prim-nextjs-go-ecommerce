import { Header, Footer } from '@/components/ui'

function MainLayout({ children }) {
    return (
        <>
            <Header />
            {children}
            <Footer />
        </>
    )
}

export default MainLayout;