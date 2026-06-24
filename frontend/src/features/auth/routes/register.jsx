import Form from '@/features/auth/components/ui/form'

export function Register() {
    return (
        <div className='overflow-y-auto max-h-full'>
            <Form formType='register' />
        </div>
    )
}