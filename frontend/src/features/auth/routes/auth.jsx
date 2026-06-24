import { Outlet } from 'react-router-dom'
import { AuthLayout } from '@/features/auth/components/layouts/authLayout'

export function Auth() {

    return (
        <AuthLayout>
            <Outlet />
        </AuthLayout>
    )
}