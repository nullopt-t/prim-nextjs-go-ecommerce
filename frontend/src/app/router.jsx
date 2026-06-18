import { Routes, Route, Navigate } from 'react-router-dom'
import { Home } from '@/features/home/routes/home'

function Router() {

    return (
        <Routes>
            <Route path='/' element={<Navigate to='/home' />} />

            <Route path='/home' element={<Home />} />
        </Routes>
    )
}

export default Router;