import HomeLayout from '@/features/home/components/layouts/homeLayout'
import { Outlet } from 'react-router-dom'

export function Home() {
    return (
        <HomeLayout>
            <Outlet />
        </HomeLayout>
    )
}