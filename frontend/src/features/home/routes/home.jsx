import MainLayout from '@/components/layouts/mainLayout'
import { Outlet } from 'react-router-dom'

export function Home() {
    return (
        <MainLayout>
            <Outlet />
        </MainLayout>
    )
}