import { Routes, Route, Navigate } from 'react-router-dom'
import { Home } from '@/features/home/routes/home'
import { Auth, Login, Register, Reset, Forgot } from '@/features/auth'

function Router() {

    return (
        <Routes>
            <Route path='/' element={<Navigate to='/home' />} />

            <Route path='/home' element={<Home />} />

            <Route path='/auth' element={<Auth />}>
                <Route path='login' element={<Login />} />
                <Route path='register' element={<Register />} />
                <Route path='forgot' element={<Forgot />} />
                <Route path='reset' element={<Reset />} />
            </Route>
        </Routes>
    )
}

export default Router;