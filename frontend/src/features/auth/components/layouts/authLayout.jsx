import AuthImage from '@/assets/auth.png'
import Form from '@/features/auth/components/ui/form'

export function AuthLayout({ children }) {

    return (
        <div className='grid grid-cols-1 md:grid-cols-2 min-h-screen max-h-screen overflow-hidden'>
            <div className='overflow-y-auto max-h-full'>
                <Form formType='register' />
            </div>
            <div className='aspect-square w-0 h-0 md:w-full md:h-full duration'>
                <img className='w-full h-full object-center object-cover' src={AuthImage} alt='Authentication image' />
            </div>

            {children}
        </div>
    )
}
