import { Header, Footer } from '@/components/ui'
import { Section } from '@/features/home/components/ui'

function HomeLayout({ children }) {
    return (
        <>
            <Header />
            {children}
            <Section />
            <Footer />
        </>
    )
}

export default HomeLayout;